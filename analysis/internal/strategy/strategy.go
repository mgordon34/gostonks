package strategy

import (
	"log"

	"github.com/mgordon34/gostonks/analysis/internal/candle"
)

type Strategy interface {
	ExecuteStep(c candle.Candle)
}

type BarStrategy struct {
	Name string
	Symbols []string
}

func NewBarStrategy(name string, symbols []string) *BarStrategy {
	return &BarStrategy{
		Name: name,
		Symbols: symbols,
	}
}

func (b *BarStrategy) ExecuteStep(c candle.Candle) {
	for _, symbol := range b.Symbols {
		if c.Symbol == symbol {
			log.Printf("Executing step for %s", b.Name)
		}
	}
}
