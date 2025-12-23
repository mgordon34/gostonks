package strategy

import (
	"log"

	"github.com/mgordon34/gostonks/market/cmd/candle"
)

type GapStatus string

const (
	Open 			GapStatus = "open"
	PartiallyFilled GapStatus = "partially_filled"
	Filled 			GapStatus = "filled"
	Inversed 		GapStatus = "inversed"
)

type FairValueGap struct {
	Direction 			Direction
	StartPrice 			float64
	EndPrice 			float64
	Candle 				*candle.Candle
	State				GapStatus
	UnfilledPrice 		float64
	LastAffectedCandle 	*candle.Candle
}

type GapManager struct {
	gaps		[]FairValueGap
}

func (gm *GapManager) AddGap(fvg FairValueGap) {
}

func (gm *GapManager) UpdateGaps(candles []candle.Candle) {
	if len(candles) != 3 {
		log.Fatalf("UpdateGaps found invalid number of candles: %d", len(candles))
	}

	// check if gaps have been filled / inversed
	// lastCandle = candles[2]
	for i := len(gm.gaps) - 1; i >= 0; i-- {
		curGap := gm.gaps[i]
		switch curGap.State {
		case Open:
			if curGap.Direction == Buyside {

			}
		case PartiallyFilled:
		case Filled:
		case Inversed:
		}
	}
}
