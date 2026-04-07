package usecase

import (
	"context"
	"erb-backend/src/broadcaster"
	"erb-backend/src/entity"
	"time"

	"github.com/google/uuid"
)

type RouteStepRepository interface {
	FindActiveRouteSteps(ctx context.Context) ([]*entity.ActiveRouteStep, error)
	CompleteRouteStep(ctx context.Context, id uuid.UUID) error
	FindNextRouteStep(ctx context.Context, assignmentID uuid.UUID,
		stepIndex int) (*entity.RouteStep, error)
	ActivateRouteStep(ctx context.Context, id uuid.UUID) error
	CreateForAssignment(ctx context.Context, assignmentID uuid.UUID,
		stepIndex int, stationID uuid.UUID) error
}

type AssignmentRepository interface {
	Create(ctx context.Context, assignment *entity.Assignment) error
	FindOldestPlanned(ctx context.Context) (*entity.PlannedAssignment, error)
	UpdateStatus(ctx context.Context, assignmentID uuid.UUID, status entity.AssignmentStatus) error
}

type OrderRepository interface {
	Create(ctx context.Context, order *entity.Order) error
	FindPending(ctx context.Context) ([]*entity.Order, error)
	UpdateIfFulfilled(ctx context.Context, orderID uuid.UUID) (bool, error)
	UpdateStatus(ctx context.Context, orderID uuid.UUID, status entity.OrderStatus) error
}

type Broadcaster interface {
	Subscribe() chan string
	Unsubscribe(chan string)
	Publish(broadcaster.Event)
}

type WagonRepository interface {
	List(ctx context.Context) ([]*entity.Wagon, error)
	ListStatusCounts(ctx context.Context) ([]entity.WagonStatusCount, error)
	ListByStatus(ctx context.Context, status entity.WagonStatus) ([]*entity.Wagon, error)
	FindLoadedReadyToUnload(ctx context.Context, olderThan time.Duration) ([]*entity.LoadedWagon, error)
	UpdateStation(ctx context.Context, wagonID, stationID uuid.UUID) error
	UpdateStatus(ctx context.Context, wagonID uuid.UUID, status entity.WagonStatus) error
}

type StationRepository interface {
	List(ctx context.Context) ([]*entity.Station, []*entity.Edge, error)
	Exists(ctx context.Context, id uuid.UUID) (bool, error)
}

type MatchingGateway interface {
	Match(ctx context.Context, orders []*entity.Order,
		wagons []*entity.Wagon) ([]*entity.AssignmentResult, error)
}
