package portfolio

import (
	"github.com/mgordon34/gostonks/analysis/internal/strategy"
)

type Portfolio struct {
	Name 		string
	Strategies 	[]strategy.Strategy
	Balance 	float64
	Positions	[]Position
}

