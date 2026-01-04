package strategy

import (
	"time"

	"github.com/mgordon34/gostonks/analysis/cmd/trading"
)

type Signal struct {
	Action		trading.Action
	Type 		trading.OrderType
	Price		float64
	TakeProfit 	float64
	StopLoss	float64
	Timestamp 	time.Time
	CancelTime	time.Time
}
