package historical

import (
	"log"
	"time"

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

	for _, candle := range candles {
		log.Print(candle)
	}

}
