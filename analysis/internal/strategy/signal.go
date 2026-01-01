package strategy

import "time"

type Signal struct {
	Action		Action
	Type 		OrderType
	Price		float64
	TakeProfit 	float64
	StopLoss	float64
	Timestamp 	time.Time
	CancelTime	time.Time
}
