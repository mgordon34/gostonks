package trading

import "time"

type Action string

const (
	BuyAction  Action = "buy"
	SellAction Action = "sell"
)

type OrderType string

const (
	MarketOrder OrderType = "market"
	LimitOrder  OrderType = "limit"
)

type OrderStatus string

const (
	OrderPending   OrderStatus = "pending"
	OrderSubmitted OrderStatus = "submitted"
	OrderFilled    OrderStatus = "filled"
	OrderCancelled OrderStatus = "cancelled"
)

type OrderRole string

const (
	EntryRole      OrderRole = "entry"
	StopLossRole   OrderRole = "stop_loss"
	TakeProfitRole OrderRole = "take_profit"
	ExitRole       OrderRole = "exit"
)

type Order interface {
	GetOrderType() OrderType
	GetPrice() *float64
	GetQuantity() int
	GetStatus() OrderStatus
	SetStatus(status OrderStatus)
	GetRole() OrderRole
}

// BaseOrder contains common fields for all order types
type BaseOrder struct {
	Status    	OrderStatus
	Price     	*float64
	Quantity	int
	Timestamp 	time.Time
}

func (o *BaseOrder) GetPrice() *float64 {
	return o.Price
}

func (o *BaseOrder) GetQuantity() int {
	return o.Quantity
}

func (o *BaseOrder) GetStatus() OrderStatus {
	return o.Status
}

func (o *BaseOrder) SetStatus(status OrderStatus) {
	o.Status = status
}

// MarketEntry executes at candle close
type MarketEntry struct {
	BaseOrder
	FillPrice *float64
}

func (o *MarketEntry) GetOrderType() OrderType {
	return MarketOrder
}

func (o *MarketEntry) GetRole() OrderRole {
	return EntryRole
}

func NewMarketEntry(price *float64, quantity int, timestamp time.Time) *MarketEntry {
	return &MarketEntry{
		BaseOrder: BaseOrder{
			Status:    OrderPending,
			Price:     price,
			Quantity:  quantity,
			Timestamp: timestamp,
		},
	}
}

// LimitEntry waits for price to return to entry level
type LimitEntry struct {
	BaseOrder
	CancelTime time.Time
}

func (o *LimitEntry) GetOrderType() OrderType {
	return LimitOrder
}

func (o *LimitEntry) GetRole() OrderRole {
	return EntryRole
}

func NewLimitEntry(price *float64, quantity int, timestamp time.Time, cancelTime time.Time) *LimitEntry {
	return &LimitEntry{
		BaseOrder: BaseOrder{
			Status:    OrderPending,
			Price:     price,
			Quantity:  quantity,
			Timestamp: timestamp,
		},
		CancelTime: cancelTime,
	}
}

// StopLoss triggers when price hits stop level
type StopLoss struct {
	BaseOrder
}

func (o *StopLoss) GetOrderType() OrderType {
	return MarketOrder
}

func (o *StopLoss) GetRole() OrderRole {
	return StopLossRole
}

func NewStopLoss(price *float64, quantity int, timestamp time.Time) *StopLoss {
	return &StopLoss{
		BaseOrder: BaseOrder{
			Status:    OrderPending,
			Price:     price,
			Quantity:  quantity,
			Timestamp: timestamp,
		},
	}
}

// TakeProfit triggers when price hits target level
type TakeProfit struct {
	BaseOrder
}

func (o *TakeProfit) GetOrderType() OrderType {
	return LimitOrder
}

func (o *TakeProfit) GetRole() OrderRole {
	return TakeProfitRole
}

func NewTakeProfit(price *float64, quantity int, timestamp time.Time) *TakeProfit {
	return &TakeProfit{
		BaseOrder: BaseOrder{
			Status:    OrderPending,
			Price:     price,
			Quantity:  quantity,
			Timestamp: timestamp,
		},
	}
}

// Exit represents a manual/forced position exit (market close, manual close, etc.)
type Exit struct {
	BaseOrder
	Reason string
}

func (o *Exit) GetOrderType() OrderType {
	return MarketOrder
}

func (o *Exit) GetRole() OrderRole {
	return ExitRole
}

func NewExit(price *float64, quantity int, timestamp time.Time, reason string) *Exit {
	return &Exit{
		BaseOrder: BaseOrder{
			Status:    OrderFilled,
			Price:     price,
			Quantity:  quantity,
			Timestamp: timestamp,
		},
		Reason: reason,
	}
}
