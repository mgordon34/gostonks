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

func InitTables(connURL string, commands []string) {
	GetDB(connURL)

	for _, command := range commands {
		_, err := pgInstance.Exec(context.Background(), command)
		if err != nil {
			log.Fatal("Error initializing table: ", err)
		}
	}
}

