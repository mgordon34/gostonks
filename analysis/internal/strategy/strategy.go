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
	Bars     map[string][]candle.Candle
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
		Bars:     make(map[string][]candle.Candle),
	}
}

func (b *BarStrategy) ProcessCandle(c candle.Candle) {
	for _, symbol := range b.Symbols {
		if c.Symbol == symbol {
			log.Printf("Processing %s candle for %s", c.Symbol, b.Name)

			b.Bars[c.Symbol] = append(b.Bars[c.Symbol], c)

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
	endTime := c.Timestamp.Truncate(time.Minute)
	startTime := endTime.Add(-time.Duration(b.Lookback) * time.Minute)
	log.Printf("Getting lookback candles for %s from %s to %s", c.Symbol, startTime.Format("2006-01-02 15:04:05"), endTime.Format("2006-01-02 15:04:05"))

	if b.hasCandlesForRange(c.Symbol, startTime, endTime) {
		log.Printf("Verified %d candles for %s", b.Lookback, c.Symbol)
		b.trimBars(c.Symbol, startTime)
		return nil
	}
	log.Printf("Did not have all candles for %s in memory, pulling from DB...", c.Symbol)

	candles := b.repo.GetCandles(b.ctx, c.Market, c.Symbol, c.Timeframe, startTime, endTime)
	if len(candles) > 0 {
		b.Bars[c.Symbol] = append(b.Bars[c.Symbol], candles...)
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

	seen := make(map[string]struct{}, len(bars))
	for _, bar := range bars {
		ts := bar.Timestamp.UTC().Truncate(time.Minute)
		if ts.Before(start) || ts.After(end) {
			continue
		}
		seen[ts.Format(time.RFC3339)] = struct{}{}
	}

	for ts := start; !ts.After(end); ts = ts.Add(time.Minute) {
		key := ts.UTC().Truncate(time.Minute).Format(time.RFC3339)
		if _, ok := seen[key]; !ok {
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

	var trimmed []candle.Candle
	for _, bar := range bars {
		if bar.Timestamp.Before(windowStart) {
			continue
		}
		trimmed = append(trimmed, bar)
	}

	b.Bars[symbol] = trimmed
	log.Printf("Bars length after trim: %d", len(b.Bars[symbol]))
	log.Print(b.Bars[symbol])
}
