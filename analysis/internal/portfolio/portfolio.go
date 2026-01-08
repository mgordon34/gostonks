package portfolio

import (
	"log"
	"time"
	"golang.org/x/text/language"
    "golang.org/x/text/message"

	"github.com/mgordon34/gostonks/analysis/cmd/trading"
	"github.com/mgordon34/gostonks/analysis/internal/risk"
	"github.com/mgordon34/gostonks/analysis/internal/strategy"
	"github.com/mgordon34/gostonks/market/cmd/candle"
)

type Portfolio struct {
	Name       		string
	Strategies 		[]strategy.Strategy
	Balance    		float64
	Positions  		[]trading.Position
	RiskStrategy 	risk.RiskStrategy
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

		if p.isMarketCloseCandle(c) && pos.Status == trading.PositionOpen {
			p.closePositionAtPrice(pos, c.Close, "market_close", c.Timestamp)
			ReportPortfolioPerformance(p.Positions)
			continue
		}

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
	for _, strat := range p.Strategies {
		signal := strat.GenerateSignal(c)

		if signal != nil {
			log.Printf("Signal found: Action=%s Entry=%.2f TP=%.2f Sl=%.2f", signal.Action, *signal.EntryPrice, *signal.TakeProfit, *signal.StopLoss)
			for _, pos := range p.Positions {
				if pos.IsOpen() {
					log.Printf("Skipping signal as already in a position")
					return
				}
			}

			position := p.newPositionFromSignal(signal, c.Close)

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
			log.Printf("Position created: Action=%s, Quantity=%d, Status=%s", position.Action, position.Quantity, position.Status)
		}
	}
}

func (p *Portfolio) newPositionFromSignal(signal *strategy.Signal, fillPrice float64) *trading.Position {
	var orders []trading.Order
	var status trading.PositionStatus

	quantity := p.RiskStrategy.GetQuantityForTrade(p.Balance, *signal.EntryPrice, *signal.StopLoss)

	if signal.EntryType == trading.MarketOrder {
		// Market order fills immediately at candle close - position is OPEN
		entry := trading.NewMarketEntry(signal.EntryPrice, quantity, signal.Timestamp)
		entry.SetStatus(trading.OrderFilled)
		entry.FillPrice = &fillPrice
		orders = append(orders, entry)
		status = trading.PositionOpen
	} else {
		// Limit order waits for price - position is PENDING
		orders = append(orders, trading.NewLimitEntry(signal.EntryPrice, quantity, signal.Timestamp, signal.CancelTime))
		status = trading.PositionPending
	}

	// Create SL and TP (submitted immediately if market entry, pending if limit)
	sl := trading.NewStopLoss(signal.StopLoss, quantity, signal.Timestamp)
	tp := trading.NewTakeProfit(signal.TakeProfit, quantity, signal.Timestamp)
	if status == trading.PositionOpen {
		sl.SetStatus(trading.OrderSubmitted)
		tp.SetStatus(trading.OrderSubmitted)
	}
	orders = append(orders, sl, tp)

	return &trading.Position{
		Action:     signal.Action,
		Status:     status,
		Quantity: 	quantity,
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
					ReportPortfolioPerformance(p.Positions)
				}
			}

		case *trading.TakeProfit:
			if o.GetStatus() == trading.OrderSubmitted {
				if p.takeProfitTriggered(pos.Action, *o.GetPrice(), c) {
					o.SetStatus(trading.OrderFilled)
					p.closePosition(pos, trading.TakeProfitRole)
					log.Printf("Take profit triggered at %.2f", *o.GetPrice())
					ReportPortfolioPerformance(p.Positions)
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

func (p *Portfolio) closePositionAtPrice(pos *trading.Position, exitPrice float64, reason string, timestamp time.Time) {
	// Create and add exit order
	exit := trading.NewExit(&exitPrice, pos.Quantity, timestamp, reason)
	pos.Orders = append(pos.Orders, exit)

	// Cancel SL and TP orders
	for _, order := range pos.Orders {
		role := order.GetRole()
		if role == trading.StopLossRole || role == trading.TakeProfitRole {
			if order.GetStatus() == trading.OrderSubmitted {
				order.SetStatus(trading.OrderCancelled)
			}
		}
	}

	pos.Status = trading.PositionClosed
	log.Printf("Position closed at %.2f (%s)", exitPrice, reason)
}

func (p *Portfolio) isMarketCloseCandle(c candle.Candle) bool {
	loc, _ := time.LoadLocation("America/New_York")
	t := c.Timestamp.In(loc)
	return t.Hour() == 15 && t.Minute() == 59
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

func ReportPortfolioPerformance(positions []trading.Position) {
	var profit float64
	var fullTrades, fullWins int
	var exitTrades, exitWins int

	for _, pos := range positions {
		if pos.Status != trading.PositionClosed {
			continue
		}

		var entryPrice, exitPrice float64
		exitViaTP := false
		exitViaExit := false
		for _, order := range pos.Orders {
			if order.GetRole() == trading.EntryRole {
				entryPrice = *order.GetPrice()
			} else if order.GetRole() == trading.TakeProfitRole && order.GetStatus() == trading.OrderFilled {
				exitPrice = *order.GetPrice()
				exitViaTP = true
			} else if order.GetRole() == trading.StopLossRole && order.GetStatus() == trading.OrderFilled {
				exitPrice = *order.GetPrice()
			} else if order.GetRole() == trading.ExitRole && order.GetStatus() == trading.OrderFilled {
				exitPrice = *order.GetPrice()
				exitViaExit = true
			}
		}

		// Calculate P&L based on position direction
		var pnl float64
		if pos.Action == trading.BuyAction {
			pnl = exitPrice - entryPrice
		} else {
			pnl = entryPrice - exitPrice
		}
		profit += pnl * float64(pos.Quantity) * 2

		// Track wins separately for full closes vs exits
		isWin := exitViaTP || pnl > 0
		if exitViaExit {
			exitTrades++
			if isWin {
				exitWins++
			}
		} else {
			fullTrades++
			if isWin {
				fullWins++
			}
		}
	}

	fullWinrate := 0.0
	if fullTrades > 0 {
		fullWinrate = float64(fullWins) / float64(fullTrades)
	}
	exitWinrate := 0.0
	if exitTrades > 0 {
		exitWinrate = float64(exitWins) / float64(exitTrades)
	}
	totalTrades := fullTrades + exitTrades

	p := message.NewPrinter(language.English)
	str := p.Sprintf("Portfolio stats: %d trades, $%.2f profit", totalTrades, profit)
	log.Println(str)
	log.Printf("  Full closes (TP/SL): %d trades, %.2f winrate", fullTrades, fullWinrate)
	log.Printf("  Early exits: %d trades, %.2f winrate", exitTrades, exitWinrate)
}
