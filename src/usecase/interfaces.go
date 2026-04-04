package usecase

import (
	"context"
	"erb-backend/src/broadcaster"
	"erb-backend/src/entity"
	"time"

	"github.com/google/uuid"
)

type RouteStepRepository interface {
	GetActiveRouteSteps(ctx context.Context) ([]*entity.ActiveRouteStep, error)
	CompleteRouteStep(ctx context.Context, id uuid.UUID) error
	GetNextRouteStep(ctx context.Context, assignmentID uuid.UUID,
		stepIndex int) (*entity.RouteStep, error)
	ActivateRouteStep(ctx context.Context, id uuid.UUID) error
	CreateForAssignment(ctx context.Context, assignmentID uuid.UUID,
		stepIndex int, stationID uuid.UUID) error
}

type AssignmentRepository interface {
	Create(ctx context.Context, assignment *entity.Assignment) error
	GetOldestPlanned(ctx context.Context) (*entity.PlannedAssignment, error)
	UpdateStatus(ctx context.Context, assignmentID uuid.UUID, status entity.AssignmentStatus) error
}

type OrderRepository interface {
	Create(ctx context.Context, order *entity.Order) error
	GetPending(ctx context.Context) ([]*entity.Order, error)
	UpdateIfFulfilled(ctx context.Context, orderID uuid.UUID) (bool, error)
	UpdateStatus(ctx context.Context, orderID uuid.UUID, status entity.OrderStatus) error
}

type Broadcaster interface {
	Subscribe() chan string
	Unsubscribe(chan string)
	Publish(broadcaster.Event)
}

type WagonRepository interface {
	ListStatusCounts(ctx context.Context) ([]entity.WagonStatusCount, error)
	GetLoadedReadyToUnload(ctx context.Context, olderThan time.Duration) ([]*entity.LoadedWagon, error)
	UpdateStation(ctx context.Context, wagonID, stationID uuid.UUID) error
	UpdateStatus(ctx context.Context, wagonID uuid.UUID, status entity.WagonStatus) error
}

type StationRepository interface {
	List(ctx context.Context) ([]*entity.Station, []*entity.Edge, error)
	Exists(ctx context.Context, id uuid.UUID) (bool, error)
}

type MatchingGateway interface {
	Match(ctx context.Context, orders []*entity.Order,
		wagons []entity.WagonStatusCount) ([]*entity.AssignmentResult, error)
}
