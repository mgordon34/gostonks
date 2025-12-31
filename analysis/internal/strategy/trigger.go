package strategy

type Trigger struct {
	Price		float64
	Action		Action
	Direction 	Direction
	Age			int
	Expiration 	int
}

func (t *Trigger) isExpired() bool {
	return t.Age > t.Expiration
}

func (t *Trigger) isTriggered(price float64) bool {
	if t.Direction == Buyside {
		return  price > t.Price
	} else if t.Direction == Sellside {
		return price < t.Price
	}
	return false
}
