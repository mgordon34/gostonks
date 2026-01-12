package backtest

import (
	"log"
	"time"

	"github.com/mgordon34/gostonks/analysis/cmd/portfolio"
	"github.com/mgordon34/gostonks/control/cmd/control"
)

type Backtest struct {
	Session 	*control.Session
	Market 		string
	Symbols 	[]string
	StartTime 	time.Time
	EndTime 	time.Time

	Portfolio	*portfolio.Portfolio
	PortfolioID	int
}

func NewBacktestSession() *Backtest {
	session := control.Session{
		ID: 1,
		Name: "Test Backtest",
	}

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
		Session: &session,
		Market: "futures",
		Symbols: []string{"NQ"},
		StartTime: startTime,
		EndTime: endTime,
		PortfolioID: 1,
	}
}
