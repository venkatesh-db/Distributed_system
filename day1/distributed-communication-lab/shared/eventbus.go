

package shared

type OrderEvent struct {
	OrderID string
	Status  string
}

var EventBus = make(chan OrderEvent, 100)