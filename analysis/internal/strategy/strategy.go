package strategy

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/mgordon34/gostonks/market/cmd/candle"
)

type Strategy interface {
	ProcessCandle(c candle.Candle)
}

type BarStrategy struct {
	ctx      context.Context
	Name     string
	Market   string
	Symbols  []string
	Lookback int
	Bars     map[string]map[time.Time]candle.Candle
	repo     candle.Repository
}

func NewBarStrategy(ctx context.Context, repo candle.Repository, name string, market string, symbols []string, lookback int) *BarStrategy {
	return &BarStrategy{
		ctx:      ctx,
		repo:     repo,
		Name:     name,
		Market:   market,
		Symbols:  symbols,
		Lookback: lookback,
		Bars:     make(map[string]map[time.Time]candle.Candle),
	}
}

func (b *BarStrategy) ProcessCandle(c candle.Candle) {
	for _, symbol := range b.Symbols {
		if c.Symbol == symbol {
			if _, ok := b.Bars[c.Symbol]; !ok {
				b.Bars[c.Symbol] = make(map[time.Time]candle.Candle)
			}

			ts := c.Timestamp.UTC().Truncate(time.Minute)
			b.Bars[c.Symbol][ts] = c

			if err := b.getNCandles(c); err != nil {
				log.Println(err)
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
	endTime := c.Timestamp.Truncate(time.Minute)
	startTime := endTime.Add(-time.Duration(b.Lookback) * time.Minute)

	if b.hasCandlesForRange(c.Symbol, startTime, endTime) {
		b.trimBars(c.Symbol, startTime)
		return nil
	}

	candles := b.repo.GetCandles(b.ctx, c.Market, c.Symbol, c.Timeframe, startTime, endTime)
	if len(candles) > 0 {
		if _, ok := b.Bars[c.Symbol]; !ok {
			b.Bars[c.Symbol] = make(map[time.Time]candle.Candle)
		}
		for _, bar := range candles {
			ts := bar.Timestamp.UTC().Truncate(time.Minute)
			b.Bars[c.Symbol][ts] = bar
		}
	}

	if !b.hasCandlesForRange(c.Symbol, startTime, endTime) {
		return fmt.Errorf("could not find all lookback candles for %s", c.Symbol)
	}

	b.trimBars(c.Symbol, startTime)
	return nil
}

func (b *BarStrategy) hasCandlesForRange(symbol string, start time.Time, end time.Time) bool {
	bars := b.Bars[symbol]
	if len(bars) == 0 {
		log.Print("No bars found from getCandles")
		return false
	}

	for ts := start; !ts.After(end); ts = ts.Add(time.Minute) {
		key := ts.UTC().Truncate(time.Minute)
		if _, ok := bars[key]; !ok {
			log.Printf("Missing candle at %s", ts.Format("2006-01-02 15:04:05"))
			return false
		}
	}

	return true
}

func (b *BarStrategy) trimBars(symbol string, windowStart time.Time) {
	bars := b.Bars[symbol]
	if len(bars) == 0 {
		return
	}

	cutoff := windowStart.UTC().Truncate(time.Minute)
	for ts := range bars {
		if ts.Before(cutoff) {
			delete(bars, ts)
		}
	}
}
