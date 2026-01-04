package trading

type Action string

const (
	BuyAction Action = "buy"
	SellAction Action = "sell"
)

type OrderType string

const (
	MarketOrder OrderType = "market"
	LimitOrder OrderType = "limit"
)
