package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

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

	pubsub := client.Subscribe(ctx, "market")
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
			log.Printf("New market event captured: %s", msg.Payload)
		}
	}
}
