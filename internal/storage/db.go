package storage

import (
	"context"
	"log"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	pgInstance *pgxpool.Pool
	pgOnce     sync.Once
)

func Ping(ctx context.Context) error {
	return pgInstance.Ping(ctx)
}

func Close() {
	pgInstance.Close()
}

func GetDB(connURL string) *pgxpool.Pool {
	pgOnce.Do(func() {
		dbpool, err := pgxpool.New(context.Background(), connURL)
		if err != nil {
			log.Fatalf("Unable to connect to database: %v\n", err)
		}

		pgInstance = dbpool
	})

	return pgInstance
}

func InitTables(connURL string) {
	GetDB(connURL)

	commands := []string{
		`CREATE TABLE IF NOT EXISTS games (
            id SERIAL PRIMARY KEY,
            market VARCHAR(255) NOT NULL,
            symbol VARCHAR(255) NOT NULL,
            timeframe VARCHAR(255) NOT NULL,
            open REAL NOT NULL,
            high REAL NOT NULL,
            low REAL NOT NULL,
            close REAL NOT NULL,
            volume INT NOT NULL,
			timestamp TIMESTAMPTZ NOT NULL,
            CONSTRAINT uq_candles UNIQUE(market, symbol, timeframe, timestamp)
        )`,
	}

	for _, command := range commands {
		_, err := pgInstance.Exec(context.Background(), command)
		if err != nil {
			log.Fatal("Error initializing table: ", err)
		}
	}
}

