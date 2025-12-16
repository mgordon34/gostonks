package strategy

import (
	"log"

	"github.com/mgordon34/gostonks/market/cmd/candle"
)

type Strategy interface {
	ProcessCandle(c candle.Candle)
}

type BarStrategy struct {
	Name string
	Symbols []string
	Lookback int
}

func NewBarStrategy(name string, symbols []string) *BarStrategy {
	b := &BarStrategy{
		Name: name,
		Symbols: symbols,
	}

	b.initStrategyData()
	return b
}

func (b *BarStrategy) ProcessCandle(c candle.Candle) {
	for _, symbol := range b.Symbols {
		if c.Symbol == symbol {
			log.Printf("Executing step for %s", b.Name)
		}
	}
}

func (b *BarStrategy) initStrategyData() {
	for _, symbol := range b.Symbols {
		log.Printf("Preloading data for %s", symbol)

	}
}
