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
	Market string
	Symbols []string
	Lookback int
	Bars map[string][]candle.Candle
}

func NewBarStrategy(name string, market string, symbols []string, lookback int) *BarStrategy {
	b := &BarStrategy{
		Name: name,
		Market: market,
		Symbols: symbols,
		Lookback: lookback,
		Bars: make(map[string][]candle.Candle),
	}

	b.initStrategyData()
	return b
}

func (b *BarStrategy) ProcessCandle(c candle.Candle) {
	for _, symbol := range b.Symbols {
		if c.Symbol == symbol {
			log.Printf("Processing %s candle for %s", c.Symbol, b.Name)

			if err := b.getNCandles(c.Symbol, b.Lookback); err != nil {
				log.Printf("Failed to get lookback for %s on %x", c.Symbol, c.Timestamp.Format("2006-01-02 15:04:05"))
			}
		}
	}
}

func (b *BarStrategy) GenerateSignal(c candle.Candle) {
	for _, symbol := range b.Symbols {
		log.Printf("Generating signals for %s", symbol)
	}
}

func (b *BarStrategy) initStrategyData() {
	for _, symbol := range b.Symbols {
		log.Printf("Preloading data for %s", symbol)
		b.getNCandles(symbol, b.Lookback)
	}
}

func (b *BarStrategy) getNCandles(symbol string, n int) error {
	return nil
}
