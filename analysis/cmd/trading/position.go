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
	Status 		PositionStatus
	Orders		[]Order
	Timestamp	time.Time
	Expiration	time.Time
}

func (p *Position) IsOpen() bool {
	return p.Status == PositionPending || p.Status == PositionOpen
}
