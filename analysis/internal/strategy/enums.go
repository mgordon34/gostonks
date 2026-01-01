package strategy

type Direction string

const (
	Buyside Direction = "buyside"
	Sellside Direction = "sellside"
)

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
