package portfolio

import (
	"log"

	"github.com/mgordon34/gostonks/analysis/cmd/trading"
	"github.com/mgordon34/gostonks/analysis/internal/strategy"
	"github.com/mgordon34/gostonks/market/cmd/candle"
)

type Portfolio struct {
	Name 		string
	Strategies 	[]strategy.Strategy
	Balance 	float64
	Positions	[]trading.Position
}

func (p *Portfolio) ProcessCandle(c candle.Candle) {
	p.updateStrategies(c)
	p.updatePositions(c)
	p.generateSignals(c)
}

func (p *Portfolio) updatePositions(c candle.Candle) {
}

func (p *Portfolio) updateStrategies(c candle.Candle) {
	for _, strategy := range p.Strategies {
		strategy.ProcessCandle(c)
	}
}

func (p *Portfolio) generateSignals(c candle.Candle) {
	for _, strategy := range p.Strategies {
		signal := strategy.GenerateSignal(c)

		if signal != nil {
			log.Printf("Signal found: %+v", *signal)
		}
	}
}
