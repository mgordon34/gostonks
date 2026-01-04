package portfolio

import (
	"log"

	"github.com/mgordon34/gostonks/analysis/cmd/trading"
	"github.com/mgordon34/gostonks/analysis/internal/strategy"
	"github.com/mgordon34/gostonks/market/cmd/candle"
)

type Portfolio struct {
	Name       string
	Strategies []strategy.Strategy
	Balance    float64
	Positions  []trading.Position
}

func (p *Portfolio) ProcessCandle(c candle.Candle) {
	p.updateStrategies(c)
	p.updatePositions(c)
	p.generateSignals(c)
}

func (p *Portfolio) updatePositions(c candle.Candle) {
	for i := range p.Positions {
		pos := &p.Positions[i]
		if !pos.IsOpen() {
			continue
		}

		// Check for position expiration (for pending limit entries)
		if pos.Status == trading.PositionPending && c.Timestamp.After(pos.Expiration) {
			p.cancelPosition(pos)
			continue
		}

		p.updateOrders(pos, c)
	}
}

func (p *Portfolio) updateStrategies(c candle.Candle) {
	for _, strategy := range p.Strategies {
		strategy.ProcessCandle(c)
	}
}

func (p *Portfolio) generateSignals(c candle.Candle) {
	// Skip if we already have an open position
	for _, pos := range p.Positions {
		if pos.IsOpen() {
			return
		}
	}

	for _, strat := range p.Strategies {
		signal := strat.GenerateSignal(c)

		if signal != nil {
			log.Printf("Signal found: %+v", *signal)

			position := newPositionFromSignal(signal, c.Close)

			// For limit orders, submit entry
			if signal.EntryType == trading.LimitOrder {
				for _, order := range position.Orders {
					if order.GetRole() == trading.EntryRole {
						order.SetStatus(trading.OrderSubmitted)
						break
					}
				}
			}

			p.Positions = append(p.Positions, *position)
			log.Printf("Position created: Action=%s, Status=%s", position.Action, position.Status)
		}
	}
}

func newPositionFromSignal(signal *strategy.Signal, fillPrice float64) *trading.Position {
	var orders []trading.Order
	var status trading.PositionStatus

	if signal.EntryType == trading.MarketOrder {
		// Market order fills immediately at candle close - position is OPEN
		entry := trading.NewMarketEntry(signal.EntryPrice, signal.Timestamp)
		entry.SetStatus(trading.OrderFilled)
		entry.FillPrice = &fillPrice
		orders = append(orders, entry)
		status = trading.PositionOpen
	} else {
		// Limit order waits for price - position is PENDING
		orders = append(orders, trading.NewLimitEntry(signal.EntryPrice, signal.Timestamp, signal.CancelTime))
		status = trading.PositionPending
	}

	// Create SL and TP (submitted immediately if market entry, pending if limit)
	sl := trading.NewStopLoss(signal.StopLoss, signal.Timestamp)
	tp := trading.NewTakeProfit(signal.TakeProfit, signal.Timestamp)
	if status == trading.PositionOpen {
		sl.SetStatus(trading.OrderSubmitted)
		tp.SetStatus(trading.OrderSubmitted)
	}
	orders = append(orders, sl, tp)

	return &trading.Position{
		Action:     signal.Action,
		Status:     status,
		Orders:     orders,
		Timestamp:  signal.Timestamp,
		Expiration: signal.CancelTime,
	}
}

func (p *Portfolio) updateOrders(pos *trading.Position, c candle.Candle) {
	for _, order := range pos.Orders {
		switch o := order.(type) {
		case *trading.LimitEntry:
			if o.GetStatus() == trading.OrderSubmitted {
				if p.limitPriceReached(pos.Action, *o.GetPrice(), c) {
					o.SetStatus(trading.OrderFilled)
					pos.Status = trading.PositionOpen
					p.activateExitOrders(pos)
					log.Printf("Limit entry filled at %.2f", *o.GetPrice())
				}
			}

		case *trading.StopLoss:
			if o.GetStatus() == trading.OrderSubmitted {
				if p.stopLossTriggered(pos.Action, *o.GetPrice(), c) {
					o.SetStatus(trading.OrderFilled)
					p.closePosition(pos, trading.StopLossRole)
					log.Printf("Stop loss triggered at %.2f", *o.GetPrice())
				}
			}

		case *trading.TakeProfit:
			if o.GetStatus() == trading.OrderSubmitted {
				if p.takeProfitTriggered(pos.Action, *o.GetPrice(), c) {
					o.SetStatus(trading.OrderFilled)
					p.closePosition(pos, trading.TakeProfitRole)
					log.Printf("Take profit triggered at %.2f", *o.GetPrice())
				}
			}
		}
	}
}

func (p *Portfolio) activateExitOrders(pos *trading.Position) {
	for _, order := range pos.Orders {
		if order.GetRole() == trading.StopLossRole || order.GetRole() == trading.TakeProfitRole {
			order.SetStatus(trading.OrderSubmitted)
		}
	}
}

func (p *Portfolio) closePosition(pos *trading.Position, reason trading.OrderRole) {
	pos.Status = trading.PositionClosed
	for _, order := range pos.Orders {
		if order.GetRole() != trading.EntryRole && order.GetRole() != reason {
			if order.GetStatus() == trading.OrderSubmitted {
				order.SetStatus(trading.OrderCancelled)
			}
		}
	}
}

func (p *Portfolio) cancelPosition(pos *trading.Position) {
	pos.Status = trading.PositionCancelled
	for _, order := range pos.Orders {
		if order.GetStatus() != trading.OrderFilled {
			order.SetStatus(trading.OrderCancelled)
		}
	}
}

func (p *Portfolio) limitPriceReached(action trading.Action, limitPrice float64, c candle.Candle) bool {
	if action == trading.BuyAction {
		return c.Low <= limitPrice
	}
	return c.High >= limitPrice
}

func (p *Portfolio) stopLossTriggered(action trading.Action, stopPrice float64, c candle.Candle) bool {
	if action == trading.BuyAction {
		return c.Low <= stopPrice
	}
	return c.High >= stopPrice
}

func (p *Portfolio) takeProfitTriggered(action trading.Action, targetPrice float64, c candle.Candle) bool {
	if action == trading.BuyAction {
		return c.High >= targetPrice
	}
	return c.Low <= targetPrice
}
