package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/mgordon34/gostonks/internal/config"
)

func main() {
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
				handle_control_message(msg.Payload)
			}
		}
	}
}

type ControlMessage struct {
	Type string				`json:"type"`
	Data json.RawMessage	`json:"data"`	
}

type DataRequest struct {
	Market 		string		`json:"market"`
	Symbol 		string		`json:"symbol"`
	StartTime 	time.Time	`json:"start_time"`
	EndTime 	time.Time	`json:"end_time"`
	Timeframe 	string		`json:"timeframe"`
}

type IngestRequest struct {
	FileName 	string	`json:"file_name"`
}

func handle_control_message(payload string) {
	log.Print("Handling control message")

	var controlMessage ControlMessage
	err := json.Unmarshal([]byte(payload), &controlMessage)
	if err != nil {
		log.Printf("Json unmarshalling failed: %d", err)
		return
	}

	log.Printf("Message type: %s", controlMessage.Type)

	if controlMessage.Type == "data_request" {
		var request DataRequest

		err := json.Unmarshal([]byte(controlMessage.Data), &request)
		if err != nil {
			log.Printf("Json unmarshalling failed: %d", err)
			return
		}

		handle_data_request(request)
	}
	if controlMessage.Type == "ingest_request" {
		var request IngestRequest

		err := json.Unmarshal([]byte(controlMessage.Data), &request)
		if err != nil {
			log.Printf("Json unmarshalling failed: %d", err)
			return
		}

		handle_ingest_request(request)
	}

}

func handle_data_request(request DataRequest) {
	log.Printf("Handling data request: %v", request)
}

func handle_ingest_request(request IngestRequest) {
	log.Printf("Handling ingest request: %v", request)
}
