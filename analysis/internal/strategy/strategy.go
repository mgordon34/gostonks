package strategy

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/mgordon34/gostonks/market/cmd/candle"
)

type Strategy interface {
	ProcessCandle(c candle.Candle)
	GenerateSignal(c candle.Candle)
}

type BarStrategy struct {
	ctx      	context.Context
	Name     	string
	Market   	string
	Symbols  	[]string
	Lookback 	int
	Bars     	map[string]map[time.Time]candle.Candle
	repo   		candle.Repository

	Location 	*time.Location
	Pools	 	[]LiquidityPool
}

func NewBarStrategy(ctx context.Context, repo candle.Repository, name string, market string, symbols []string, lookback int) *BarStrategy {
	nyLocation, err := time.LoadLocation("America/New_York")
	if err != nil {
		log.Fatalf("failed to load America/New_York location: %v", err)
	}
	return &BarStrategy{
		ctx:      ctx,
		repo:     repo,
		Name:     name,
		Market:   market,
		Symbols:  symbols,
		Lookback: lookback,
		Bars:     make(map[string]map[time.Time]candle.Candle),

		Location: nyLocation,
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
				return
			}

			tsNY := c.Timestamp.In(b.Location)
			if tsNY.Hour() == 9 && tsNY.Minute() == 30 {
				log.Printf("Candle at 09:30 America/New_York for %s: %s", c.Symbol, c.Timestamp.Format("2006-01-02 15:04:05"))
				b.initializeDay(c.Symbol, c.Timestamp)
			}
		}
	}
}

func (b *BarStrategy) GenerateSignal(c candle.Candle) {
	for _, symbol := range b.Symbols {
		if c.Symbol == symbol {
			// log.Printf("Generating signals for %s", symbol)
		}
	}
}

func (b *BarStrategy) getNCandles(c candle.Candle) error {
	if len(b.Bars[c.Symbol]) >= b.Lookback {
		return nil
	}

	log.Println("Not enough bars in history, pulling from db...")

	candles := b.repo.GetPastCandles(b.ctx, c.Market, c.Symbol, c.Timeframe, c.Timestamp, b.Lookback)
	if len(candles) > 0 {
		if _, ok := b.Bars[c.Symbol]; !ok {
			b.Bars[c.Symbol] = make(map[time.Time]candle.Candle)
		}
		for _, bar := range candles {
			ts := bar.Timestamp.UTC().Truncate(time.Minute)
			b.Bars[c.Symbol][ts] = bar
		}
	}

	if len(b.Bars[c.Symbol]) < b.Lookback {
		return fmt.Errorf("could not find all lookback candles for %s", c.Symbol)
	}

	return nil
}

func (b *BarStrategy) initializeDay(symbol string, timestamp time.Time) {
	b.Pools = []LiquidityPool{}

	prevDay := timestamp.AddDate(0, 0, -1)
	asiaOpen :=  time.Date(prevDay.Year(), prevDay.Month(), prevDay.Day(), 20, 0, 0, 0, b.Location)
	asiaClose :=  time.Date(timestamp.Year(), timestamp.Month(), timestamp.Day(), 3, 0, 0, 0, b.Location)
	asiaLow, err := b.getMinInRange(symbol, asiaOpen, asiaClose)
	if err != nil {
		log.Fatal(err)
	}
	asiaHigh, err := b.getMaxInRange(symbol, asiaOpen, asiaClose)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Asia low: %f on %s", asiaLow.Low, asiaLow.Timestamp.In(b.Location).Format(time.RFC3339))
	log.Printf("Asia high: %f on %s", asiaHigh.High, asiaHigh.Timestamp.In(b.Location).Format(time.RFC3339))
}

func (b *BarStrategy) getMinInRange(symbol string, startTime time.Time, endTime time.Time) (candle.Candle, error) {
	var low candle.Candle

	if startTime.After(endTime) {
		return candle.Candle{}, errors.New("startTime cannot be past endTime")
	}
	for ts := startTime; !ts.After(endTime); ts = ts.Add(time.Minute) {
		ts = ts.UTC().Truncate(time.Minute)
		c := b.Bars[symbol][ts]

		if low.Low == 0.0 || c.Low < low.Low {
			low = c
		}
	}

	return low, nil
}

func (b *BarStrategy) getMaxInRange(symbol string, startTime time.Time, endTime time.Time) (candle.Candle, error) {
	var high candle.Candle

	if startTime.After(endTime) {
		return candle.Candle{}, errors.New("startTime cannot be past endTime")
	}
	for ts := startTime; !ts.After(endTime); ts = ts.Add(time.Minute) {
		ts = ts.UTC().Truncate(time.Minute)
		c := b.Bars[symbol][ts]

		if high.High == 0.0 || c.High < high.High {
			high = c
		}
	}

	return high, nil
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
