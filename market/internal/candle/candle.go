package candle

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/mgordon34/gostonks/internal/config"
	"github.com/mgordon34/gostonks/internal/storage"
)

type Candle struct {
	ID			int 		`db:"id"`
	Market		string 		`db:"market"`
	Symbol		string 		`db:"symbol"`
	Timeframe	string 		`db:"timeframe"`
	Open		float64 	`db:"open"`
	High		float64 	`db:"high"`
	Low			float64 	`db:"low"`
	Close		float64 	`db:"close"`
	Volume		int 		`db:"volume"`
	Timestamp	time.Time	`db:"timestamp"`
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
			"market":       candle.Market,
			"symbol": 		candle.Symbol,
			"timeframe":	candle.Timeframe,
			"open":  		candle.Open,
			"high":      	candle.High,
			"low":      	candle.Low,
			"close":      	candle.Close,
			"volume":      	candle.Volume,
			"timestamp":    candle.Timestamp,
		},
	).Scan(&id)

	if err != nil {
		log.Fatalf("Failed to add candle: %v", err)
	}

	return id
}
