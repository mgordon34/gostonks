package portfolio

import (
	"time"

	"github.com/mgordon34/gostonks/analysis/internal/strategy"
)

type PositionStatus string

const (
	PositionPending PositionStatus = "pending"
	PositionOpen PositionStatus = "open"
	PositionClosed PositionStatus = "closed"
	PositionCancelled PositionStatus = "cancelled"
)

type Position struct {
	Action	 	strategy.Action
	Type 		strategy.OrderType
	EnterPrice	float64
	StopLoss	float64
	TakeProfit	float64
	ExitPrice	float64
	Status 		PositionStatus
	Timestamp	time.Time
	CancelTime	time.Time
}

func (p *Position) IsOpen() bool {
	return p.Status == PositionPending || p.Status == PositionOpen
}
