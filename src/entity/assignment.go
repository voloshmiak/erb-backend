package entity

import (
	"time"

	"github.com/google/uuid"
)

type AssignmentStatus string

const (
	AssignmentPlanned   AssignmentStatus = "planned"
	AssignmentInTransit AssignmentStatus = "in_transit"
	AssignmentDelivered AssignmentStatus = "delivered"
)

type Assignment struct {
	ID               uuid.UUID        `json:"id"`
	OrderID          uuid.UUID        `json:"orderId"`
	WagonID          uuid.UUID        `json:"wagonId"`
	EmptyRunKM       int              `json:"emptyRunKm"`
	CostEmptyRun     int              `json:"costEmptyRun"`
	EstimatedArrival time.Time        `json:"estimatedArrival"`
	Status           AssignmentStatus `json:"status"`
}

type AssignmentResult struct {
	Assignment *Assignment `json:"assignment"`
	Route      []uuid.UUID `json:"route"`
}

type PlannedAssignment struct {
	ID          uuid.UUID
	OrderID     uuid.UUID
	WagonID     uuid.UUID
	FirstStepID uuid.UUID
}
