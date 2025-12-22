package strategy

import (
	"time"

	"github.com/mgordon34/gostonks/market/cmd/candle"
)

type LiquidityPool struct {
	Price 		float64
	Direction 	Direction
	Candle 		*candle.Candle

	timeRaided	time.Time
}

func (lp *LiquidityPool) BeenRaided() bool {
	return lp.timeRaided.IsZero()
}

func (lp *LiquidityPool) SetRaided(c candle.Candle) {
	lp.timeRaided = c.Timestamp
}

type LiquidityPoolManager struct {
	activePools []LiquidityPool
	raidedPools []LiquidityPool
}

func (lpm *LiquidityPoolManager) AddLP(lp LiquidityPool) {
	for i := len(lpm.activePools) - 1; i >= 0; i-- {
		curPool := lpm.activePools[i]
		if (curPool.Direction == Buyside && lp.Price >= curPool.Price) || (curPool.Direction == Sellside && lp.Price <= curPool.Price) {
			lpm.activePools = append(lpm.activePools[:i], lpm.activePools[i+1:]...)

			curPool.timeRaided = lp.Candle.Timestamp
			lpm.raidedPools = append(lpm.raidedPools, curPool)
		}
	}

	lpm.activePools = append(lpm.activePools, lp)
}

func (lpm *LiquidityPoolManager) GetPools(lp LiquidityPool, wantRaided bool) []LiquidityPool {
	if wantRaided {
		return lpm.raidedPools
	} else {
		return lpm.activePools
	}
}
