package strategy

import (
	"time"

	"github.com/mgordon34/gostonks/analysis/cmd/trading"
)

type Signal struct {
	Action		trading.Action
	Timestamp 	time.Time
	CancelTime	time.Time

	EntryType 	trading.OrderType
	EntryPrice	*float64

	TakeProfit 	*float64
	StopLoss	*float64
}
