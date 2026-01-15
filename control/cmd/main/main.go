package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/mgordon34/gostonks/control/cmd/backtest"
	"github.com/redis/go-redis/v9"
)

type ControlMessage struct {
	Type string      	`json:"type"`
	Data DataRequest 	`json:"data"`
}

type DataRequest struct {
	SessionName string 	`json:"session_name"`
	Market    string    `json:"market"`
	Symbol    string    `json:"symbol"`
	Timeframe string    `json:"timeframe"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

func main() {
	log.Println("Hello from the Control service")
	TriggerDataRequest()
}

func TriggerDataRequest() {
	ctx := context.Background()

	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	defer client.Close()

	// startTime, err := time.Parse(time.RFC3339, "2015-01-02T16:52:00-05:00")
	startTime, err := time.Parse(time.RFC3339, "2025-12-02T16:52:00-05:00")
	if err != nil {
		log.Fatalf("Failed to parse start time: %v", err)
	}
	endTime, err := time.Parse(time.RFC3339, "2025-12-18T16:53:00-05:00")
	if err != nil {
		log.Fatalf("Failed to parse end time: %v", err)
	}

	b := backtest.NewBacktestSession()
	log.Printf("Created new backtest session: %s with Portfolio ID %d", b.SessionName, b.PortfolioID)

	msg := ControlMessage{
		Type: "data_request",
		Data: DataRequest{
			SessionName: b.SessionName,
			Market:    "futures",
			Symbol:    "NQ",
			Timeframe: "1m",
			StartTime: startTime,
			EndTime:   endTime,
		},
	}

	payload, err := json.Marshal(msg)
	if err != nil {
		log.Fatalf("Failed to marshal message: %v", err)
	}

	err = client.Publish(ctx, "control", payload).Err()
	if err != nil {
		log.Fatalf("Failed to publish message: %v", err)
	}

	log.Println("Data request published successfully!")
}
