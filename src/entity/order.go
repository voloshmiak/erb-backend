package entity

import (
	"time"

	"github.com/google/uuid"
)

type OrderStatus string

const (
	Pending   OrderStatus = "pending"
	Matched   OrderStatus = "matched"
	Fulfilled OrderStatus = "fulfilled"
	Cancelled OrderStatus = "cancelled"
)

type OrderType string

const (
	External OrderType = "external"
	Internal OrderType = "internal"
)

type Order struct {
	ID          uuid.UUID   `json:"id"`
	ClientName  string      `json:"clientName"`
	StationToID uuid.UUID   `json:"stationToId"`
	WagonType   WagonType   `json:"wagonType"`
	Quantity    int         `json:"quantity"`
	DesiredDate time.Time   `json:"desiredDate"`
	Status      OrderStatus `json:"status"`
	Type        OrderType   `json:"type"`
	CreatedAt   time.Time   `json:"createdAt"`
}

func NewOrder(clientName string, stationToID uuid.UUID, wagonType WagonType,
	quantity int, desiredDate time.Time, orderType OrderType) *Order {
	return &Order{
		ID:          uuid.New(),
		ClientName:  clientName,
		StationToID: stationToID,
		WagonType:   wagonType,
		Quantity:    quantity,
		DesiredDate: desiredDate,
		Status:      Pending,
		Type:        orderType,
		CreatedAt:   time.Now(),
	}
}
