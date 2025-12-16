package strategy

import (
	"context"
	"log"

	"github.com/mgordon34/gostonks/market/cmd/candle"
)

type Strategy interface {
	ProcessCandle(c candle.Candle)
}

type BarStrategy struct {
	ctx context.Context
	Name string
	Market string
	Symbols []string
	Lookback int
	Bars map[string][]candle.Candle
	repo candle.Repository
}

func NewBarStrategy(ctx context.Context, repo candle.Repository, name string, market string, symbols []string, lookback int) *BarStrategy {
	return &BarStrategy{
		ctx: ctx,
		repo: repo,
		Name: name,
		Market: market,
		Symbols: symbols,
		Lookback: lookback,
		Bars: make(map[string][]candle.Candle),
	}
}

func (b *BarStrategy) ProcessCandle(c candle.Candle) {
	for _, symbol := range b.Symbols {
		if c.Symbol == symbol {
			log.Printf("Processing %s candle for %s", c.Symbol, b.Name)

			if err := b.getNCandles(c); err != nil {
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

func (b *BarStrategy) getNCandles(c candle.Candle) error {
	return nil
}
