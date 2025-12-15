package candle

import "time"


type Candle struct {
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
