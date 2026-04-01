package entity

import (
	"time"

	"github.com/google/uuid"
)

type Type string

const (
	Sorting Type = "sorting"
	Border  Type = "border"
	Port    Type = "port"
	Freight Type = "freight"
)

type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type Station struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Type      Type      `json:"type"`
	Location  Location  `json:"location"`
	Capacity  int       `json:"capacity"`
	CreatedAt time.Time `json:"createdAt"`
}

func NewStation(name string, stationType Type, location Location, capacity int) *Station {
	return &Station{
		ID:       uuid.New(),
		Name:     name,
		Type:     stationType,
		Location: location,
		Capacity: capacity,
	}
}
