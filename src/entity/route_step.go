package entity

import (
	"time"

	"github.com/google/uuid"
)

type RouteStepStatus string

const (
	RouteStepPending   RouteStepStatus = "pending"
	RouteStepCurrent   RouteStepStatus = "current"
	RouteStepCompleted RouteStepStatus = "completed"
)

type ActiveRouteStep struct {
	ID           uuid.UUID
	AssignmentID uuid.UUID
	OrderID      uuid.UUID
	WagonID      uuid.UUID
	WagonNumber  string
	StepIndex    int
	TotalSteps   int
}

type RouteStep struct {
	ID          uuid.UUID
	StationID   uuid.UUID
	StepIndex   int
	StationName string
	Lat         float64
	Lng         float64
}

type WagonMovedPayload struct {
	WagonID     uuid.UUID `json:"wagonId"`
	WagonNumber string    `json:"wagonNumber"`
	StationID   uuid.UUID `json:"stationId"`
	StationName string    `json:"stationName"`
	Lat         float64   `json:"lat"`
	Lng         float64   `json:"lng"`
	StepIndex   int       `json:"stepIndex"`
	TotalSteps  int       `json:"totalSteps"`
	ArrivedAt   time.Time `json:"arrivedAt"`
}

type WagonArrivedPayload struct {
	WagonID      uuid.UUID `json:"wagonId"`
	OrderID      uuid.UUID `json:"orderId"`
	AssignmentID uuid.UUID `json:"assignmentId"`
}
