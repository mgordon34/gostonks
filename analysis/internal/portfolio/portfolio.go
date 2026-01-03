package portfolio

import (
	"log"

	"github.com/mgordon34/gostonks/analysis/cmd/position"
	"github.com/mgordon34/gostonks/analysis/internal/strategy"
	"github.com/mgordon34/gostonks/market/cmd/candle"
)

type Portfolio struct {
	Name 		string
	Strategies 	[]strategy.Strategy
	Balance 	float64
	Positions	[]position.Position
}

func (p *Portfolio) ProcessCandle(c candle.Candle) {
	for _, strategy := range p.Strategies {
		strategy.ProcessCandle(c)
		signal := strategy.GenerateSignal(c)

		if signal != nil {
			log.Printf("Signal found: %+v", *signal)
		}
	}
}
