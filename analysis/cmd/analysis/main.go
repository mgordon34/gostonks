package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/mgordon34/gostonks/analysis/internal/strategy"
	"github.com/mgordon34/gostonks/internal/config"
	"github.com/mgordon34/gostonks/market/cmd/candle"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	redisHost := config.Get("REDIS_HOST", "redis")
	redisPort := config.Get("REDIS_PORT", "6379")
	addr := fmt.Sprintf("%s:%s", redisHost, redisPort)

	client := redis.NewClient(&redis.Options{Addr: addr})
	defer client.Close()

	var strategies []strategy.Strategy
	strategies = append(strategies, strategy.NewBarStrategy("iFVG Strat", []string{"NQ"}))

	log.Printf("Analysis service listening for candles on redis list 'market' at %s", addr)

	for {
		values, err := client.BLPop(ctx, 0*time.Second, "market").Result()
		if err != nil {
			if errors.Is(err, context.Canceled) || ctx.Err() != nil {
				log.Printf("Strategy service shutting down: %v", ctx.Err())
				return
			}
			log.Printf("BLPOP error: %v", err)
			time.Sleep(time.Second)
			continue
		}
		if len(values) == 2 {
			log.Printf("Received candle payload: %s", values[1])

			var c candle.Candle
			err := json.Unmarshal([]byte(values[1]), &c)
			if err != nil {
				log.Printf("Json unmarshalling failed: %d", err)
				continue
			}
			log.Printf("Candle: %v", c)

			for _, strategy := range strategies {
				strategy.ProcessCandle(c)
			}
			continue
		}
		log.Printf("Unexpected BLPOP response: %v", values)
	}
}
