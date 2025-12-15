package strategy

import "log"

type Strategy interface {
	ExecuteStep()
}

type BarStrategy struct {
	Name string
	Symbols []string
}

func NewBarStrategy(name string, symbols []string) *BarStrategy {
	return &BarStrategy{
		Name: name,
		Symbols: symbols,
	}
}

func (b *BarStrategy) ExecuteStep() {
	log.Printf("Executing step for %s", b.Name)
}
