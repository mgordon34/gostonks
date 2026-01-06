package risk

import (
	"log"
	"math"
)

type RiskStrategy interface {
	GetQuantityForTrade(balance float64, entryPrice float64, stopLoss float64) int
}

type SimpleRiskStrategy struct {
	Market			string
	RiskPercent 	float64
}

func (rs *SimpleRiskStrategy) GetQuantityForTrade(balance float64, entryPrice float64, stopLoss float64) int {
	riskDollars := balance * rs.RiskPercent
	contractCost := math.Abs(entryPrice - stopLoss) * float64(GetRiskPerPoint(rs.Market))

	return int(riskDollars / contractCost)
}

func GetRiskPerPoint(market string) int {
	switch market {
	case "futures":
		return 2
	default:
		log.Fatalf("Unknown market: %s", market)
		return 0
	}
}
