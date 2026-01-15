package backtest

import (
	"fmt"
	"log"
	"time"

	"github.com/mgordon34/gostonks/analysis/cmd/portfolio"
)

var sessionCount int

type BacktestStatus string

const (
    BacktestPending   BacktestStatus = "pending"
    BacktestRunning   BacktestStatus = "running"
    BacktestCompleted BacktestStatus = "completed"
    BacktestFailed    BacktestStatus = "failed"
)

type Backtest struct {
    ID          int
    SessionName string
    Status      BacktestStatus
    Market      string
    Symbols     []string
    StartTime   time.Time
    EndTime     time.Time
    Portfolio   *portfolio.Portfolio
    PortfolioID int
    CreatedAt   time.Time
    CompletedAt *time.Time
}

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

    sessionCount++
    return &Backtest{
        ID:          sessionCount,
        SessionName: fmt.Sprintf("Backtest-%d", sessionCount),
        Status:      BacktestPending,
		Market: "futures",
		Symbols: []string{"NQ"},
		StartTime: startTime,
		EndTime: endTime,
		PortfolioID: 1,
        CreatedAt:   time.Now(),
    }
}

func (b *Backtest) Start() {
    b.Status = BacktestRunning
}

func (b *Backtest) Complete() {
    b.Status = BacktestCompleted
    now := time.Now()
    b.CompletedAt = &now
}
