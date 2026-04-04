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
	ID               uuid.UUID
	OrderID          uuid.UUID
	WagonID          uuid.UUID
	EmptyRunKM       int
	CostEmptyRun     int
	EstimatedArrival time.Time
	Status           AssignmentStatus
}

type AssignmentResult struct {
	Assignment *Assignment
	Route      []uuid.UUID
}

type PlannedAssignment struct {
	ID          uuid.UUID
	OrderID     uuid.UUID
	WagonID     uuid.UUID
	FirstStepID uuid.UUID
}
