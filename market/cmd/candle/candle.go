package candle

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
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
func (c *Candle) Age(other *Candle) (int, error) {
	if other.Timestamp.Before(c.Timestamp) {
		return -1, fmt.Errorf("FairValueGap timestamp %s is after candle timestamp %s", c.Timestamp.Format(time.RFC3339), other.Timestamp.Format(time.RFC3339))
	}

	return int(other.Timestamp.Sub(c.Timestamp).Minutes()), nil
}


type Repository interface {
	GetCandles(ctx context.Context, market string, symbol string, timeframe string, startTime time.Time, endTime time.Time) []Candle
	GetPastCandles(ctx context.Context, market string, symbol string, timeframe string, startTime time.Time, count int) []Candle
	AddCandle(ctx context.Context, candle Candle) int
}

type CandleRepository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *CandleRepository {
	return &CandleRepository{db}
}

func (r *CandleRepository) GetPastCandles(ctx context.Context, market string, symbol string, timeframe string, startTime time.Time, count int) []Candle {
	sql := `SELECT id, market, symbol, timeframe, open, high, low, close, volume, timestamp
			FROM candles
			WHERE market = @market
			  AND symbol = @symbol
			  AND timeframe = @timeframe
			  AND timestamp <= @start_time
			ORDER BY timestamp DESC
			LIMIT @count`

	rows, err := r.db.Query(
		ctx,
		sql,
		pgx.NamedArgs{
			"market":     market,
			"symbol":     symbol,
			"timeframe":  timeframe,
			"start_time": startTime,
			"count":      count,
		},
	)
	if err != nil {
		log.Fatalf("Failed to query past candles: %v", err)
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
			log.Fatalf("Failed to scan past candle: %v", err)
		}
		candles = append(candles, c)
	}

	if err := rows.Err(); err != nil {
		log.Fatalf("Past candle rows error: %v", err)
	}

	return candles
}

func (r *CandleRepository) GetCandles(ctx context.Context, market string, symbol string, timeframe string, startTime time.Time, endTime time.Time) []Candle {
	sql := `SELECT id, market, symbol, timeframe, open, high, low, close, volume, timestamp
			FROM candles
			WHERE market = @market
			  AND symbol = @symbol
			  AND timeframe = @timeframe
			  AND timestamp >= @start_time
			  AND timestamp <= @end_time
			ORDER BY timestamp`

	rows, err := r.db.Query(
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

func (r *CandleRepository) AddCandle(ctx context.Context, candle Candle) int {
	sql := `INSERT INTO candles (market, symbol, timeframe, open, high, low, close, volume, timestamp) VALUES (@market, @symbol, @timeframe, @open, @high, @low, @close, @volume, @timestamp) RETURNING id`

	var id int
	err := r.db.QueryRow(
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
