package main

import (
	"context"
	"encoding/json"
	"log"

	"github.com/redis/go-redis/v9"
)

type ControlMessage struct {
	Type string      `json:"type"`
	Data DataRequest `json:"data"`
}

type DataRequest struct {
	Market    string `json:"market"`
	Symbol    string `json:"symbol"`
	Timeframe string `json:"timeframe"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
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

	msg := ControlMessage{
		Type: "data_request",
		Data: DataRequest{
			Market:    "futures",
			Symbol:    "NQ",
			Timeframe: "1m",
			StartTime: "2015-01-02T16:52:00-05:00",
			EndTime:   "2025-12-18T16:53:00-05:00",
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

	log.Println("Data request published successfully")
}
