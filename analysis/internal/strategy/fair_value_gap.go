package strategy

import (
	"log"

	"github.com/mgordon34/gostonks/market/cmd/candle"
)

type GapStatus string

const (
	GapOpen 			GapStatus = "open"
	GapPartiallyFilled 	GapStatus = "partially_filled"
	GapFilled 			GapStatus = "filled"
	GapInversed 		GapStatus = "inversed"
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
	candles  	[]candle.Candle
	gaps		[]FairValueGap
}

func (gm *GapManager) ProcessCandle(candle candle.Candle) {
	gm.candles = append(gm.candles, candle)
	if len(gm.candles) > 3 {
		gm.candles = gm.candles[1:]
		gm.addGapIfExists()
	}

	gm.updateGaps()
}

func (gm *GapManager) addGapIfExists() {
	if len(gm.candles) < 3 {
		return
	}

	if gm.candles[0].High < gm.candles[2].Low {
		gap := FairValueGap{
			Direction: Buyside,
			StartPrice: gm.candles[0].High,
			EndPrice: gm.candles[2].Low,
			UnfilledPrice: gm.candles[2].Low,
			Candle: &gm.candles[1],
			LastAffectedCandle: &gm.candles[1],
			State: GapOpen,
		}
		gm.gaps = append(gm.gaps, gap)
	} else if gm.candles[0].Low > gm.candles[2].High {
		gap := FairValueGap{
			Direction: Sellside,
			StartPrice: gm.candles[0].Low,
			EndPrice: gm.candles[2].High,
			UnfilledPrice: gm.candles[2].High,
			Candle: &gm.candles[1],
			LastAffectedCandle: &gm.candles[1],
			State: GapOpen,
		}
		gm.gaps = append(gm.gaps, gap)
	}
}

func (gm *GapManager) updateGaps() {
	// check if gaps have been filled / inversed
	// lastCandle = candles[2]
	for i := len(gm.gaps) - 1; i >= 0; i-- {
		curGap := gm.gaps[i]
		switch curGap.State {
		case GapOpen:
			if curGap.Direction == Buyside {

			}
		case GapPartiallyFilled:
		case GapFilled:
		case GapInversed:
		}
	}
}
