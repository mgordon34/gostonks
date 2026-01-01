package strategy

import (
	"log"
	"math"

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

func (gap *FairValueGap) Age(c *candle.Candle) (int, error) {
	return gap.Candle.Age(c)
}

func (gap *FairValueGap) Width() int {
	width, err := gap.Candle.Age(gap.LastAffectedCandle)
	if err != nil {
		log.Fatalf("Invalid width result %v", err)
	}

	return width
}

func (gap *FairValueGap) processCandle(c *candle.Candle) {
	if gap.State == GapInversed {
		return 
	}

	switch gap.Direction {
	case Buyside:
		if c.Low < gap.UnfilledPrice {
			gap.UnfilledPrice = math.Max(c.Low, gap.EndPrice)
			if c.Close < gap.StartPrice {
				gap.State = GapInversed
			} else {
				gap.State = GapPartiallyFilled
			}
			gap.LastAffectedCandle = c
		}
	case Sellside:
		if c.High > gap.UnfilledPrice {
			gap.UnfilledPrice = math.Min(c.High, gap.EndPrice)
			if c.Close > gap.StartPrice {
				gap.State = GapInversed
			} else {
				gap.State = GapPartiallyFilled
			}
			gap.LastAffectedCandle = c
		}
	}
}

type GapManager struct {
	candles  	[]candle.Candle
	gaps		[]FairValueGap
}

func (gm *GapManager) ProcessCandle(c candle.Candle) {
	gm.candles = append(gm.candles, c)
	if len(gm.candles) > 3 {
		gm.candles = gm.candles[1:]
		gm.addGapIfExists()
	}

	for i := range gm.gaps {
      gm.gaps[i].processCandle(&c)
	}
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
		// log.Printf("Adding FvG at %s: %+v", gap.Candle.Timestamp.Format(time.RFC3339), gap)
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
		// log.Printf("Adding FvG at %s: %+v", gap.Candle.Timestamp.Format(time.RFC3339), gap)
		gm.gaps = append(gm.gaps, gap)
	}
}

func (gm * GapManager) GetInverses(c *candle.Candle, maxAge int, maxWidth int) ([]FairValueGap, error) {
	if maxAge < 0 {
		maxAge = math.MaxInt
	}
	if maxWidth < 0 {
		maxWidth = math.MaxInt
	}

	var gaps []FairValueGap
	for _, gap := range gm.gaps {
		if gap.State != GapInversed {
			continue
		}
		age, err := gap.LastAffectedCandle.Age(c)
		if err != nil {
			return gaps, err
		}
		width := gap.Width()

		if age <= maxAge && width <= maxWidth {
			gaps = append(gaps, gap)
		}
	}

	return gaps, nil
}
