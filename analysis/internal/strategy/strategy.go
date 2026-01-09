package strategy

import (
	"context"
	"fmt"
	"log"
	"math"
	"runtime/debug"
	"time"

	"github.com/mgordon34/gostonks/analysis/cmd/trading"
	"github.com/mgordon34/gostonks/market/cmd/candle"
)

type Strategy interface {
	ProcessCandle(c candle.Candle)
	GenerateSignal(c candle.Candle) *Signal
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
	Pools	 	LiquidityPoolManager
	Gaps   		GapManager
	trades 		int
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

func (b *BarStrategy) initializeDay(symbol string, timestamp time.Time) {
	b.Pools = LiquidityPoolManager{}
	b.Gaps = GapManager{}
	b.trades = 0

	prevDay := timestamp.AddDate(0, 0, -1)
	asiaOpen :=  time.Date(prevDay.Year(), prevDay.Month(), prevDay.Day(), 20, 0, 0, 0, b.Location)
	asiaClose :=  time.Date(timestamp.Year(), timestamp.Month(), timestamp.Day(), 3, 0, 0, 0, b.Location)
	londonOpen :=  time.Date(timestamp.Year(), timestamp.Month(), timestamp.Day(), 3, 0, 0, 0, b.Location)
	londonClose :=  time.Date(timestamp.Year(), timestamp.Month(), timestamp.Day(), 7, 0, 0, 0, b.Location)
	preMarketOpen :=  time.Date(timestamp.Year(), timestamp.Month(), timestamp.Day(), 7, 0, 0, 0, b.Location)

	asiaLow := b.getMinInRange(symbol, asiaOpen, asiaClose)
	asiaHigh := b.getMaxInRange(symbol, asiaOpen, asiaClose)
	londonLow := b.getMinInRange(symbol, londonOpen, londonClose)
	londonHigh := b.getMaxInRange(symbol, londonOpen, londonClose)
	preMarketLow := b.getMinInRange(symbol, preMarketOpen, timestamp)
	preMarketHigh := b.getMaxInRange(symbol, preMarketOpen, timestamp)

	b.Pools.AddLP(LiquidityPool{Price: asiaLow.Low, Direction: Sellside, Candle: &asiaLow, Name: "Asia Low"})
	b.Pools.AddLP(LiquidityPool{Price: asiaHigh.High, Direction: Buyside, Candle: &asiaHigh, Name: "Asia High"})
	b.Pools.AddLP(LiquidityPool{Price: londonLow.Low, Direction: Sellside, Candle: &londonLow, Name: "London Low"})
	b.Pools.AddLP(LiquidityPool{Price: londonHigh.High, Direction: Buyside, Candle: &londonHigh, Name: "London High"})
	b.Pools.AddLP(LiquidityPool{Price: preMarketLow.Low, Direction: Sellside, Candle: &preMarketLow, Name: "Pre Market Low"})
	b.Pools.AddLP(LiquidityPool{Price: preMarketHigh.High, Direction: Buyside, Candle: &preMarketHigh, Name: "Pre Market High"})

	// log.Printf("Active Pools: %v", b.Pools.GetPools(true))
	// log.Printf("Raided Pools: %v", b.Pools.GetPools(false))
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
				log.Printf("New day for %s: %s", c.Symbol, c.Timestamp.Format("2006-01-02 15:04:05"))
				b.initializeDay(c.Symbol, c.Timestamp)
			}

			b.Pools.UpdateLPs(c)
			b.Gaps.ProcessCandle(c)
		}
	}
}

func (b *BarStrategy) GenerateSignal(c candle.Candle) *Signal {
	for _, symbol := range b.Symbols {
		if c.Symbol != symbol || b.trades > 0 {
			continue
		}

		tsNY := c.Timestamp.In(b.Location)
		startTime := time.Date(tsNY.Year(), tsNY.Month(), tsNY.Day(), 9, 30, 0, 0, b.Location)
		endTime := time.Date(tsNY.Year(), tsNY.Month(), tsNY.Day(), 16, 00, 0, 0, b.Location)
		if tsNY.Before(startTime) || !tsNY.Before(endTime) {
			continue
		}
		// prevDay := tsNY.AddDate(0, 0, -1)
		// prevDayOpen := time.Date(prevDay.Year(), prevDay.Month(), prevDay.Day(), 9, 30, 0, 0, b.Location)
		// prevDayClose := time.Date(prevDay.Year(), prevDay.Month(), prevDay.Day(), 16, 00, 0, 0, b.Location)
		// prevDayHigh := b.getMaxInRange(c.Symbol, prevDayOpen, prevDayClose)
		// prevDayLow := b.getMinInRange(c.Symbol, prevDayOpen, prevDayClose)

		raids := b.Pools.GetPools(false)
		if len(raids) == 0 {
			continue
		}
		inverses, err := b.Gaps.GetInverses(&c, 0, 20)
		if err != nil {
			log.Fatalf("Error getting inverses: %v\n%s", err, debug.Stack())
		}

		for _, raid := range raids {
			raidAge, err := raid.RaidCandle.Age(&c)
			if err != nil {
				log.Fatalf("Error getting raid age: %v\n%s", err, debug.Stack())
			}
			raidWidth, err := raid.Candle.Age(raid.RaidCandle)
			if err != nil {
				log.Printf("Error getting raid width: %v\nLP candle: %+v\nRaid Candle: %+v\n%s", err, *raid.Candle, *raid.RaidCandle, debug.Stack())
			}

			if raidAge > 50 || raidWidth > math.MaxInt {
				continue
			}

			// In probable raid window, looking for inverses
			if len(inverses) == 0 {
				continue
			}
			for _, inverse := range inverses {
				if raid.Direction == Buyside && inverse.Direction == Buyside && c.Close < raid.Price /*&& c.Close < prevDayHigh.High*/ {
					// sl := b.getMaxInRange(c.Symbol, raid.RaidCandle.Timestamp, c.Timestamp).High
					sl := inverse.Candle.High + (inverse.Candle.High - c.Close) * 1
					if (sl - c.Close) > 100 || (sl - c.Close) < 0 {
						continue
					}
					tp := c.Close - (sl - c.Close) * 1
					signal := Signal{
						Action: trading.SellAction,
						EntryType: trading.MarketOrder,
						EntryPrice: &c.Close,
						TakeProfit: &tp,
						StopLoss: &sl,
						Timestamp: c.Timestamp,
						CancelTime: c.Timestamp.Add(120 * time.Minute),
					}
					b.trades++
					return &signal
				} else if raid.Direction == Sellside && inverse.Direction == Sellside  && c.Close > raid.Price /*&& c.Close > prevDayLow.Low*/ {
					// sl := b.getMinInRange(c.Symbol, raid.RaidCandle.Timestamp, c.Timestamp).Low
					sl := inverse.Candle.Low - (c.Close - inverse.Candle.Low) * 1
					if (c.Close - sl) > 100 || (c.Close - sl) < 0 {
						continue
					}
					tp := c.Close + (c.Close - sl) * 1
					signal := Signal{
						Action: trading.BuyAction,
						EntryType: trading.MarketOrder,
						EntryPrice: &c.Close,
						TakeProfit: &tp,
						StopLoss: &sl,
						Timestamp: c.Timestamp,
						CancelTime: c.Timestamp.Add(120 * time.Minute),
					}
					b.trades++
					return &signal
				}
			}
		}
	}
	return nil
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

func (b *BarStrategy) getMinInRange(symbol string, startTime time.Time, endTime time.Time) candle.Candle {
	var low candle.Candle

	if startTime.After(endTime) {
		log.Fatalf("startTime cannot be past endTime\n%s", debug.Stack())
	}
	for ts := startTime; !ts.After(endTime); ts = ts.Add(time.Minute) {
		ts = ts.UTC().Truncate(time.Minute)
		c := b.Bars[symbol][ts]

		if low.Low == 0.0 || c.Low < low.Low {
			low = c
		}
	}

	return low
}

func (b *BarStrategy) getMaxInRange(symbol string, startTime time.Time, endTime time.Time) candle.Candle {
	var high candle.Candle

	if startTime.After(endTime) {
		log.Fatalf("startTime cannot be past endTime\n%s", debug.Stack())
	}
	for ts := startTime; !ts.After(endTime); ts = ts.Add(time.Minute) {
		ts = ts.UTC().Truncate(time.Minute)
		c := b.Bars[symbol][ts]

		if high.High == 0.0 || c.High > high.High {
			high = c
		}
	}

	return high
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
