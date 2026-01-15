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
	SessionName string		`json:"session_name"`
	PortfolioID int			`json:"portfolio_id"`
	Market 		string 		`json:"market"`
	Symbol    	string    	`json:"symbol"`
	StartTime 	time.Time 	`json:"start_time"`
	EndTime   	time.Time 	`json:"end_time"`
	Timeframe 	string    	`json:"timeframe"`
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
	// Sending start message
    startMsg := candle.QueueMessage{
        Type:        "event",
        SessionName: request.SessionName,
        Event: &candle.Event{
            Name:        "session_start",
            PortfolioID: request.PortfolioID,  // Which portfolio config to use
        },
    }
    payload, _ := json.Marshal(startMsg)
    s.broker.RPush(ctx, s.queue, payload)


	log.Printf(
		"Handling data request for %s, from %s to %s",
		request.Symbol,
		request.StartTime.Format("2006-01-02 15:04:05"),
		request.EndTime.Format("2006-01-02 15:04:05"),
	)

	candles := s.repo.GetCandles(ctx, request.Market, request.Symbol, request.Timeframe, request.StartTime, request.EndTime)

	for _, c := range candles {
		msg := candle.QueueMessage{
			Type:        "candle",
			SessionName: request.SessionName,
			Candle:      &c,
		}
		payload, _ := json.Marshal(msg)
		if err := s.broker.RPush(ctx, s.queue, payload).Err(); err != nil {
			log.Printf("Failed to enqueue candle to redis: %v", err)
			return
		}
	}

	// Send completion event
    completeMsg := candle.QueueMessage{
        Type:        "event",
        SessionName: request.SessionName,
        Event: &candle.Event{
            Name:        "session_complete",
            CandleCount: len(candles),
        },
    }
    payload, _ = json.Marshal(completeMsg)
    s.broker.RPush(ctx, s.queue, payload)

	log.Printf("Enqueued %d candles to redis list 'market'", len(candles))
}
