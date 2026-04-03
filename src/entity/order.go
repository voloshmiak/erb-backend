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

type Order struct {
	ID          uuid.UUID   `json:"id"`
	ClientName  string      `json:"clientName"`
	StationToID uuid.UUID   `json:"stationToId"`
	WagonType   WagonType   `json:"wagonType"`
	Quantity    int         `json:"quantity"`
	DesiredDate time.Time   `json:"desiredDate"`
	Status      OrderStatus `json:"status"`
	CreatedAt   time.Time   `json:"createdAt"`
}

func NewOrder(clientName string, stationToID uuid.UUID, wagonType WagonType,
	quantity int, desiredDate time.Time) *Order {
	return &Order{
		ID:          uuid.New(),
		ClientName:  clientName,
		StationToID: stationToID,
		WagonType:   wagonType,
		Quantity:    quantity,
		DesiredDate: desiredDate,
		Status:      Pending,
		CreatedAt:   time.Now(),
	}
}
