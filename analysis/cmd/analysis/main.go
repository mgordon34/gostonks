package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/mgordon34/gostonks/analysis/cmd/portfolio"
	"github.com/mgordon34/gostonks/analysis/internal/risk"
	"github.com/mgordon34/gostonks/analysis/internal/strategy"
	"github.com/mgordon34/gostonks/internal/config"
	"github.com/mgordon34/gostonks/internal/storage"
	"github.com/mgordon34/gostonks/market/cmd/candle"
)

type SessionState struct {
    Name             string
    Portfolio        *portfolio.Portfolio
    CandlesProcessed int
    StartTime        time.Time
}


func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	redisHost := config.Get("REDIS_HOST", "redis")
	redisPort := config.Get("REDIS_PORT", "6379")
	addr := fmt.Sprintf("%s:%s", redisHost, redisPort)

	client := redis.NewClient(&redis.Options{Addr: addr})
	defer client.Close()

	db := storage.GetDB(config.Get("DB_URL", ""))
	candleRepository := candle.NewRepository(db)

	var strategies []strategy.Strategy
	strategies = append(strategies, strategy.NewBarStrategy(ctx, candleRepository, "iFVG Strat", "futures", []string{"NQ"}, 2880))
	port := portfolio.Portfolio{
		Name: "Backtest Portfolio",
		Strategies: strategies,
		Balance: 100000,
		RiskStrategy: &risk.SimpleRiskStrategy{Market: "futures", RiskPercent: .02},
	}

	log.Printf("Analysis service listening for candles on redis list 'market' at %s", addr)

	// TODO: Be able to handle multiple sessions
    var activeSession *SessionState

	for {
		values, err := client.BLPop(ctx, 0*time.Second, "market").Result()
		if err != nil {
			if errors.Is(err, context.Canceled) || ctx.Err() != nil {
				log.Printf("Strategy service shutting down: %v", ctx.Err())
				return
			}
			log.Printf("BLPOP error: %v", err)
			time.Sleep(time.Second)
			continue
		}

		var msg candle.QueueMessage
        json.Unmarshal([]byte(values[1]), &msg)

        switch msg.Type {
        case "event":
            if msg.Event.Name == "session_start" {
                activeSession = &SessionState{
                    Name:      msg.SessionName,
                    Portfolio: &port,
                    StartTime: time.Now(),
                }
                log.Printf("Session %s started with Portfolio %d",
                    msg.SessionName, msg.Event.PortfolioID)
            } else if msg.Event.Name == "session_complete" {
                if activeSession != nil && activeSession.Name == msg.SessionName {
                    log.Printf("Session %s complete: processed %d candles",
                        msg.SessionName, activeSession.CandlesProcessed)
                    portfolio.ReportPortfolioPerformance(activeSession.Portfolio.Positions)
                    activeSession = nil
                }
            }

        case "candle":
            if activeSession != nil && activeSession.Name == msg.SessionName {
                activeSession.Portfolio.ProcessCandle(*msg.Candle)
                activeSession.CandlesProcessed++
            }
        }
	}
}
