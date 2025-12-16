package historical

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/mgordon34/gostonks/market/cmd/candle"
)

type DataRequest struct {
	Market    string    `json:"market"`
	Symbol    string    `json:"symbol"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Timeframe string    `json:"timeframe"`
}

type Broker interface {
	RPush(ctx context.Context, key string, values ...interface{}) *redis.IntCmd
}

type Service struct {
	broker Broker
	repo candle.Repository
	queue  string
}

func NewService(broker Broker, repo candle.Repository) *Service {
	return &Service{
		broker: broker,
		repo: repo,
		queue:  "market",
	}
}

func (s *Service) HandleDataRequest(ctx context.Context, request DataRequest) {
	log.Printf(
		"Handling data request for %s, from %s to %s",
		request.Symbol,
		request.StartTime.Format("2006-01-02 15:04:05"),
		request.EndTime.Format("2006-01-02 15:04:05"),
	)

	candles := s.repo.GetCandles(ctx, request.Market, request.Symbol, request.Timeframe, request.StartTime, request.EndTime)

	for _, candle := range candles {
		payload, err := json.Marshal(candle)
		if err != nil {
			log.Printf("Failed to marshal candle: %v", err)
			continue
		}

		if err := s.broker.RPush(ctx, s.queue, payload).Err(); err != nil {
			log.Printf("Failed to enqueue candle to redis: %v", err)
			return
		}
	}

	log.Printf("Enqueued %d candles to redis list 'market'", len(candles))
}
