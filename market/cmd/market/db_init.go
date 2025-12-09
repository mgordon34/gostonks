package main

func GetCommands() []string {
	return []string{
		`CREATE TABLE IF NOT EXISTS candles (
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
}
