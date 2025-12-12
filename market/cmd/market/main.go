package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/redis/go-redis/v9"

	"github.com/mgordon34/gostonks/internal/config"
	"github.com/mgordon34/gostonks/internal/storage"
	"github.com/mgordon34/gostonks/market/internal/historical"
	"github.com/mgordon34/gostonks/market/internal/ingest"
)

func main() {
	storage.InitTables(config.Get("DB_URL", ""), GetCommands())
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	redisHost := config.Get("REDIS_HOST", "redis")
	redisPort := config.Get("REDIS_PORT", "6379")
	addr := fmt.Sprintf("%s:%s", redisHost, redisPort)

	log.Printf("Starting market service; redis host=%s port=%s", redisHost, redisPort)

	client := redis.NewClient(&redis.Options{Addr: addr})
	defer client.Close()

	pubsub := client.Subscribe(ctx, "control")
	defer pubsub.Close()

	if _, err := pubsub.Receive(ctx); err != nil {
		log.Fatalf("Failed to subscribe to market channel: %v", err)
	}

	ch := pubsub.Channel()
	log.Printf("Listening for market events on %s channel 'market'", addr)

	for {
		select {
		case <-ctx.Done():
			log.Printf("Shutting down: %v", ctx.Err())
			return
		case msg, ok := <-ch:
			if !ok {
				log.Printf("Market subscription channel closed")
				return
			}
			log.Printf("New event captured on channel: %s", msg.Channel)
			if msg.Channel == "control" {
				handleControlMessage(msg.Payload)
			}
		}
	}
}

type ControlMessage struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

func handleControlMessage(payload string) {
	var controlMessage ControlMessage
	err := json.Unmarshal([]byte(payload), &controlMessage)
	if err != nil {
		log.Printf("Json unmarshalling failed: %d", err)
		return
	}

	switch controlMessage.Type {
	case "data_request":
		decodeAndHandle(controlMessage.Data, historical.HandleDataRequest)
	case "ingest_request":
		decodeAndHandle(controlMessage.Data, ingest.HandleIngest)
	default:
		log.Printf("Unknown control message type: %s", controlMessage.Type)
	}

}

func decodeAndHandle[T any](data json.RawMessage, handler func(T)) {
	var payload T
	if err := json.Unmarshal(data, &payload); err != nil {
		log.Printf("Json unmarshalling failed: %v", err)
		return
	}
	handler(payload)
}
