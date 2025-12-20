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
	"github.com/mgordon34/gostonks/internal/storage"
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

	db := storage.GetDB(config.Get("DB_URL", ""))
	candleRepository := candle.NewRepository(db)

	var strategies []strategy.Strategy
	strategies = append(strategies, strategy.NewBarStrategy(ctx, candleRepository, "iFVG Strat", "futures", []string{"NQ"}, 2880))

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
			var c candle.Candle
			err := json.Unmarshal([]byte(values[1]), &c)
			if err != nil {
				log.Printf("Json unmarshalling failed: %d", err)
				continue
			}
			log.Printf("Received candle payload for %s on %s", c.Symbol, c.Timestamp.Format("2006-01-02 15:04:05"))

			for _, strategy := range strategies {
				strategy.ProcessCandle(c)
				strategy.GenerateSignal(c)
			}
			continue
		}
		log.Printf("Unexpected BLPOP response: %v", values)
	}
}
