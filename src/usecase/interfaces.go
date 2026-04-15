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
	ActivateRouteStep(ctx context.Context, id uuid.UUID, simHour int64) error
	CreateForAssignment(ctx context.Context, assignmentID uuid.UUID,
		stepIndex int, stationID uuid.UUID, durationHours float64) error
}

type AssignmentRepository interface {
	Create(ctx context.Context, assignment *entity.Assignment) error
	FindOldestPlanned(ctx context.Context) (*entity.PlannedAssignment, error)
	FindActiveByWagonID(ctx context.Context, wagonID uuid.UUID) (*entity.Assignment, error)
	UpdateStatus(ctx context.Context, assignmentID uuid.UUID, status entity.AssignmentStatus) error
	GetStats(ctx context.Context) (*entity.AssignmentStats, error)
}

type OrderRepository interface {
	List(ctx context.Context) ([]*entity.Order, error)
	Create(ctx context.Context, order *entity.Order) error
	FindPending(ctx context.Context) ([]*entity.Order, error)
	UpdateIfFulfilled(ctx context.Context, orderID uuid.UUID) (*entity.Order, error)
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
	FindLoadedReadyToUnload(ctx context.Context, currentHour int64) ([]*entity.LoadedWagon, error)
	UpdateStation(ctx context.Context, wagonID, stationID uuid.UUID) error
	UpdateStatus(ctx context.Context, wagonID uuid.UUID, status entity.WagonStatus, stateUntilHour *int64) error
}

type TrainRepository interface {
	Create(ctx context.Context, train *entity.Train) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Train, error)
	ListActive(ctx context.Context) ([]*entity.Train, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status entity.TrainStatus, at time.Time) error
	UpdateProgress(ctx context.Context, train *entity.Train) error
}

type LocomotiveRepository interface {
	List(ctx context.Context) ([]*entity.Locomotive, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Locomotive, error)
	ListByStatus(ctx context.Context, status entity.LocomotiveStatus) ([]*entity.Locomotive, error)
	GetAvailableAtStation(ctx context.Context, stationID uuid.UUID) (*entity.Locomotive, error)
	Update(ctx context.Context, l *entity.Locomotive) error
	MarkIdleIfAvailable(ctx context.Context, simHour int64) error
}

type StationRepository interface {
	List(ctx context.Context) ([]*entity.Station, []*entity.Edge, error)
	Exists(ctx context.Context, id uuid.UUID) (bool, error)
}

type MatchingGateway interface {
	Match(ctx context.Context, orders []*entity.Order, wagons []*entity.Wagon,
		stations []*entity.Station, edges []*entity.Edge,
		locomotives []*entity.Locomotive,
	) ([]*entity.AssignmentResult, []entity.TrainGroupResult, *entity.MatchingMetrics, error)
}

type MatchingRunRepository interface {
	Save(ctx context.Context, m *entity.MatchingMetrics) error
	LoadAggregated(ctx context.Context) (*entity.MatchingRunAggregate, error)
}
