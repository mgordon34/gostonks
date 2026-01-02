package portfolio

import (
	"time"

	"github.com/mgordon34/gostonks/analysis/internal/strategy"
)

type Portfolio struct {
	Name 		string
	Strategies 	[]strategy.Strategy
	Balance 	float64
	Positions	[]Position
}

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

type PositionStatus string

const (
	Pending PositionStatus = "pending"
	Open PositionStatus = "open"
	Closed PositionStatus = "closed"
	Cancelled PositionStatus = "cancelled"
)
