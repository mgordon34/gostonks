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
	RaidCandle	*candle.Candle
}

func (lp *LiquidityPool) BeenRaided() bool {
	return lp.timeRaided.IsZero()
}

func (lp *LiquidityPool) SetRaided(c candle.Candle) {
	lp.timeRaided = c.Timestamp
	lp.RaidCandle = &c
}

type LiquidityPoolManager struct {
	activePools []LiquidityPool
	raidedPools []LiquidityPool
}

func (lpm *LiquidityPoolManager) UpdateLPs(candle candle.Candle) {
	for i := len(lpm.activePools) - 1; i >= 0; i-- {
		curPool := lpm.activePools[i]
		if curPool.Direction == Buyside && candle.High >= curPool.Price {

			lpm.activePools = append(lpm.activePools[:i], lpm.activePools[i+1:]...)

			curPool.SetRaided(candle)
			lpm.raidedPools = append(lpm.raidedPools, curPool)

		} else if curPool.Direction == Sellside && candle.Low <= curPool.Price {

			lpm.activePools = append(lpm.activePools[:i], lpm.activePools[i+1:]...)

			curPool.SetRaided(candle)
			lpm.raidedPools = append(lpm.raidedPools, curPool)
		}
	}
}

func (lpm *LiquidityPoolManager) AddLP(lp LiquidityPool) {
	lpm.UpdateLPs(*lp.Candle)

	lpm.activePools = append(lpm.activePools, lp)
}

func (lpm *LiquidityPoolManager) GetPools(isActive bool) []LiquidityPool {
	if isActive {
		return lpm.activePools
	} else {
		return lpm.raidedPools
	}
}
