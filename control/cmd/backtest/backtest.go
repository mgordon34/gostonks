package backtest

import (
	"fmt"
	"log"
	"time"

	"github.com/mgordon34/gostonks/analysis/cmd/portfolio"
)

type Backtest struct {
	ID 			int
	SessionName string
	Market 		string
	Symbols 	[]string
	StartTime 	time.Time
	EndTime 	time.Time

	Portfolio	*portfolio.Portfolio
	PortfolioID	int
}

var sessionCount int

func NewBacktestSession() *Backtest {
	// TODO: make this an argument in
	startTime, err := time.Parse(time.RFC3339, "2015-01-02T16:52:00-05:00")
	if err != nil {
			log.Fatalf("Failed to parse start time: %v", err)
	}
	endTime, err := time.Parse(time.RFC3339, "2025-12-18T16:53:00-05:00")
	if err != nil {
			log.Fatalf("Failed to parse end time: %v", err)
	}

	return &Backtest{
		SessionName: fmt.Sprintf("Backtest-%d", sessionCount),
		Market: "futures",
		Symbols: []string{"NQ"},
		StartTime: startTime,
		EndTime: endTime,
		PortfolioID: 1,
	}
}
