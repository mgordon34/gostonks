package ingest

import (
	"log"
	"time"

	"github.com/mgordon34/gostonks/market/internal/types"
)

// Request represents an ingest payload coming from the control queue.
type IngestRequest struct {
	FileName string `json:"file_name"`
}

// Handle processes an ingest request.
func HandleIngest(request IngestRequest) {
	log.Printf("Handing request to ingest data: %v", request)

	c := types.Candle{
		Market:    "futures",
		Symbol:    "NQ",
		Timeframe: "1m",
		Open:      100.0,
		High:      110.5,
		Low:       91.5,
		Close:     105.0,
		Volume:    1000,
		Timestamp: time.Now(),
	}
	id := types.AddCandle(c)
	log.Printf("Created candle with id %d", id)
}
