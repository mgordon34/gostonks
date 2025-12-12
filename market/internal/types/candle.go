package types

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/mgordon34/gostonks/internal/config"
	"github.com/mgordon34/gostonks/internal/storage"
)

type Candle struct {
	ID        int       `db:"id"`
	Market    string    `db:"market"`
	Symbol    string    `db:"symbol"`
	Timeframe string    `db:"timeframe"`
	Open      float64   `db:"open"`
	High      float64   `db:"high"`
	Low       float64   `db:"low"`
	Close     float64   `db:"close"`
	Volume    int       `db:"volume"`
	Timestamp time.Time `db:"timestamp"`
}

func GetCandles(market string, symbol string, timeframe string, startTime time.Time, endTime time.Time) []Candle {
	ctx := context.Background()
	db := storage.GetDB(config.Get("DB_URL", ""))
	sql := `SELECT id, market, symbol, timeframe, open, high, low, close, volume, timestamp
			FROM candles
			WHERE market = @market
			  AND symbol = @symbol
			  AND timeframe = @timeframe
			  AND timestamp >= @start_time
			  AND timestamp <= @end_time
			ORDER BY timestamp`

	rows, err := db.Query(
		ctx,
		sql,
		pgx.NamedArgs{
			"market":     market,
			"symbol":     symbol,
			"timeframe":  timeframe,
			"start_time": startTime,
			"end_time":   endTime,
		},
	)
	if err != nil {
		log.Fatalf("Failed to query candles: %v", err)
	}
	defer rows.Close()

	var candles []Candle
	for rows.Next() {
		var c Candle
		if err := rows.Scan(
			&c.ID,
			&c.Market,
			&c.Symbol,
			&c.Timeframe,
			&c.Open,
			&c.High,
			&c.Low,
			&c.Close,
			&c.Volume,
			&c.Timestamp,
		); err != nil {
			log.Fatalf("Failed to scan candle: %v", err)
		}
		candles = append(candles, c)
	}

	if err := rows.Err(); err != nil {
		log.Fatalf("Candle rows error: %v", err)
	}

	return candles
}

func AddCandle(candle Candle) int {
	ctx := context.Background()
	db := storage.GetDB(config.Get("DB_URL", ""))
	sql := `INSERT INTO candles (market, symbol, timeframe, open, high, low, close, volume, timestamp) VALUES (@market, @symbol, @timeframe, @open, @high, @low, @close, @volume, @timestamp) RETURNING id`

	defer db.Close()

	var id int
	err := db.QueryRow(
		ctx,
		sql,
		pgx.NamedArgs{
			"market":    candle.Market,
			"symbol":    candle.Symbol,
			"timeframe": candle.Timeframe,
			"open":      candle.Open,
			"high":      candle.High,
			"low":       candle.Low,
			"close":     candle.Close,
			"volume":    candle.Volume,
			"timestamp": candle.Timestamp,
		},
	).Scan(&id)

	if err != nil {
		log.Fatalf("Failed to add candle: %v", err)
	}

	return id
}
