package historical

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/mgordon34/gostonks/internal/config"
	"github.com/mgordon34/gostonks/market/internal/types"
)

type DataRequest struct {
	Market    string    `json:"market"`
	Symbol    string    `json:"symbol"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Timeframe string    `json:"timeframe"`
}

func HandleDataRequest(request DataRequest) {
	log.Printf(
		"Handling data request for %s, from %s to %s",
		request.Symbol,
		request.StartTime.Format("2006-01-02 15:04:05"),
		request.EndTime.Format("2006-01-02 15:04:05"),
	)

	candles := types.GetCandles(request.Market, request.Symbol, request.Timeframe, request.StartTime, request.EndTime)

	redisHost := config.Get("REDIS_HOST", "redis")
	redisPort := config.Get("REDIS_PORT", "6379")
	addr := fmt.Sprintf("%s:%s", redisHost, redisPort)

	client := redis.NewClient(&redis.Options{Addr: addr})
	defer client.Close()

	ctx := context.Background()
	for _, candle := range candles {
		payload, err := json.Marshal(candle)
		if err != nil {
			log.Printf("Failed to marshal candle: %v", err)
			continue
		}

		if err := client.RPush(ctx, "market", payload).Err(); err != nil {
			log.Printf("Failed to enqueue candle to redis: %v", err)
			return
		}
	}

	log.Printf("Enqueued %d candles to redis list 'market'", len(candles))
}
