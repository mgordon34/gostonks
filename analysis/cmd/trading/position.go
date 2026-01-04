package trading

import (
	"time"
)

type PositionStatus string

const (
	PositionPending PositionStatus = "pending"
	PositionOpen PositionStatus = "open"
	PositionClosed PositionStatus = "closed"
	PositionCancelled PositionStatus = "cancelled"
)

type Position struct {
	Action	 	Action
	Type 		OrderType
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
