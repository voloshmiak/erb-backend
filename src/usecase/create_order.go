package usecase

import (
	"context"
	"erb-backend/src/broadcaster"
	"erb-backend/src/entity"
	"erb-backend/src/simclock"
	"log"
	"math"
	"time"

	"github.com/pkg/errors"

	"github.com/google/uuid"
)

var (
	ErrInvalidInput = errors.New("invalid input")
	ErrNotFound     = errors.New("not found")
)

type CreateOrderUseCase struct {
	orderRepository      OrderRepository
	stationRepository    StationRepository
	assignmentRepository AssignmentRepository
	routeStepRepository  RouteStepRepository
	wagonRepository      WagonRepository
	trainRepository      TrainRepository
	locomotiveRepository LocomotiveRepository
	broadcaster          Broadcaster
	matchingGateway      MatchingGateway
	simClock             *simclock.SimClock
	metricsStore         *MetricsStore
}

func NewCreateOrderUseCase(
	orderRepository OrderRepository,
	stationRepository StationRepository,
	assignmentRepository AssignmentRepository,
	routeStepRepository RouteStepRepository,
	wagonRepository WagonRepository,
	trainRepository TrainRepository,
	locomotiveRepository LocomotiveRepository,
	broadcaster Broadcaster,
	matchingGateway MatchingGateway,
	clock *simclock.SimClock,
	metricsStore *MetricsStore,
) *CreateOrderUseCase {
	return &CreateOrderUseCase{
		orderRepository:      orderRepository,
		stationRepository:    stationRepository,
		assignmentRepository: assignmentRepository,
		routeStepRepository:  routeStepRepository,
		wagonRepository:      wagonRepository,
		trainRepository:      trainRepository,
		locomotiveRepository: locomotiveRepository,
		broadcaster:          broadcaster,
		matchingGateway:      matchingGateway,
		simClock:             clock,
		metricsStore:         metricsStore,
	}
}

type CreateOrderInput struct {
	ClientName  string           `json:"clientName"`
	StationToID uuid.UUID        `json:"stationToId"`
	WagonType   entity.WagonType `json:"wagonType"`
	Quantity    int              `json:"quantity"`
	DesiredDate string           `json:"desiredDate"`
	Type        entity.OrderType `json:"type"`
}

func (i *CreateOrderInput) validate() error {
	if i.Type == "" {
		i.Type = entity.External
	}
	if i.Type == entity.Internal {
		if i.ClientName == "" {
			i.ClientName = "Internal"
		}
	} else {
		if i.ClientName == "" {
			return errors.Wrap(ErrInvalidInput, "client name is required")
		}
	}

	if i.StationToID.String() == "00000000-0000-0000-0000-000000000000" {
		return errors.Wrap(ErrInvalidInput, "station ID is required")
	}
	if i.Quantity <= 0 {
		return errors.Wrap(ErrInvalidInput, "quantity must be greater than zero")
	}
	return nil
}

func (u *CreateOrderUseCase) Execute(ctx context.Context, input CreateOrderInput) (*entity.Order, error) {
	if err := input.validate(); err != nil {
		return nil, err
	}

	parsedDesiredDate, err := time.Parse(time.DateOnly, input.DesiredDate)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse desired date")
	}

	if parsedDesiredDate.Before(time.Now()) {
		return nil, errors.Wrap(ErrInvalidInput, "desired date cannot be in the past")
	}

	exists, err := u.stationRepository.Exists(ctx, input.StationToID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to check station existence")
	}
	if !exists {
		return nil, errors.Wrap(ErrNotFound, "station")
	}

	order := entity.NewOrder(input.ClientName, input.StationToID, input.WagonType,
		input.Quantity, parsedDesiredDate, input.Type)

	if err = u.orderRepository.Create(ctx, order); err != nil {
		return nil, errors.Wrap(err, "failed to create order")
	}

	if order.Type == entity.External {
		u.broadcaster.Publish(broadcaster.NewEvent(broadcaster.OrderCreated, order))

		go func() {
			if err = u.match(context.WithoutCancel(ctx)); err != nil {
				log.Println("failed to match orderRepository after creation:", err)
			}
		}()
	}

	return order, nil
}

func (u *CreateOrderUseCase) match(ctx context.Context) error {
	orders, err := u.orderRepository.FindPending(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get pending orderRepository")
	}
	if len(orders) == 0 {
		return nil
	}

	wagons, err := u.wagonRepository.ListByStatus(ctx, entity.Idle)
	if err != nil {
		return errors.Wrap(err, "failed to get wagon counts")
	}

	stations, edges, err := u.stationRepository.List(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get stations and edges")
	}

	locomotives, err := u.locomotiveRepository.ListByStatus(ctx, entity.LocomotiveIdle)
	if err != nil {
		return errors.Wrap(err, "failed to get idle locomotives")
	}

	results, trainGroups, metrics, err := u.matchingGateway.Match(ctx, orders, wagons, stations, edges, locomotives)
	if err != nil {
		return errors.Wrap(err, "matching service failed")
	}

	u.metricsStore.Accumulate(ctx, metrics)

	edgeDistMap := buildEdgeDistanceMap(edges)
	stationTypeMap := buildStationTypeMap(stations)

	wagonsInGroups := make(map[uuid.UUID]struct{})
	for _, tg := range trainGroups {
		for _, wid := range tg.WagonIDs {
			wagonsInGroups[wid] = struct{}{}
		}
	}

	for _, r := range results {
		r.Assignment.ID = uuid.New()
		r.Assignment.Status = entity.AssignmentPlanned
		r.Assignment.LoadedRunKM = r.Assignment.EmptyRunKM

		// Compute step durations first to set ETA before inserting
		stepDurations := make([]float64, len(r.Route))
		var totalDurationHours float64
		for i, stationID := range r.Route {
			d := computeStepDuration(i, stationID, r.Route, edgeDistMap, stationTypeMap)
			stepDurations[i] = d
			totalDurationHours += d
		}

		r.Assignment.EstimatedArrival = u.simClock.ToDisplayTime(
			u.simClock.Now() + int64(math.Ceil(totalDurationHours)))

		// Assignment must exist before route steps (FK constraint)
		if err = u.assignmentRepository.Create(ctx, r.Assignment); err != nil {
			log.Printf("match: failed to create assignment for order %s: %v", r.Assignment.OrderID, err)
			continue
		}

		// Skip route steps if wagon is part of a train group (AdvanceTrains will move it)
		if _, inGroup := wagonsInGroups[r.Assignment.WagonID]; !inGroup {
			for i, stationID := range r.Route {
				if err = u.routeStepRepository.CreateForAssignment(ctx, r.Assignment.ID,
					i, stationID, stepDurations[i]); err != nil {
					log.Printf("match: failed to create route step %d for assignment %s: %v",
						i, r.Assignment.ID, err)
				}
			}
		}

		if err = u.orderRepository.UpdateStatus(ctx, r.Assignment.OrderID, entity.Matched); err != nil {
			log.Printf("match: failed to update order status %s: %v", r.Assignment.OrderID, err)
		}

		u.broadcaster.Publish(broadcaster.NewEvent(broadcaster.AssignmentCreated, r))
	}

	// Materialize train groups
	simHour := u.simClock.Now()
	for _, tg := range trainGroups {
		if len(tg.WagonIDs) == 0 {
			continue
		}

		repositionHours := tg.RepositionKM / 40.0
		distanceHours := tg.DistanceKM / 40.0
		// Train starts moving after loco finishes repositioning
		activatedAtHour := simHour + int64(math.Ceil(repositionHours))
		totalArrivalHour := simHour + int64(math.Ceil(repositionHours+distanceHours))

		train := &entity.Train{
			ID:              tg.TrainID, // reuse python-side UUID
			WagonIDs:        tg.WagonIDs,
			Route:           []uuid.UUID{tg.SourceStationID, tg.DestinationStationID},
			StepIndex:       0,
			SourceStationID: tg.SourceStationID,
			NextStationID:   tg.DestinationStationID,
			Status:          entity.TrainInTransit,
			LocomotiveID:    tg.LocomotiveID,
			CreatedAt:       u.simClock.ToDisplayTime(simHour),
			ActivatedAtHour: activatedAtHour,
			DurationHours:   distanceHours,
		}
		departedAt := u.simClock.ToDisplayTime(activatedAtHour)
		train.DepartedAt = &departedAt

		if err := u.trainRepository.Create(ctx, train); err != nil {
			log.Printf("match: failed to create train %s: %v", train.ID, err)
			continue
		}

		// Mark wagons as in_train
		for _, wid := range tg.WagonIDs {
			if err := u.wagonRepository.UpdateStatus(ctx, wid, entity.InTrain, nil); err != nil {
				log.Printf("match: failed to mark wagon %s in_train: %v", wid, err)
			}
		}

		// Reserve locomotive (if assigned by matcher)
		if tg.LocomotiveID != nil {
			loco, err := u.locomotiveRepository.GetByID(ctx, *tg.LocomotiveID)
			if err != nil {
				log.Printf("match: failed to fetch loco %s: %v", tg.LocomotiveID, err)
			} else {
				loco.Status = entity.LocomotiveInTransit
				loco.TrainID = &train.ID
				loco.AvailableAtHour = totalArrivalHour
				loco.AvailableAt = u.simClock.ToDisplayTime(totalArrivalHour)
				loco.CurrentStationID = tg.DestinationStationID
				if err := u.locomotiveRepository.Update(ctx, loco); err != nil {
					log.Printf("match: failed to reserve loco %s: %v", loco.ID, err)
				} else {
					u.broadcaster.Publish(broadcaster.NewEvent(broadcaster.LocomotiveDispatched, loco))
				}
			}
		}

		u.broadcaster.Publish(broadcaster.NewEvent(broadcaster.TrainCreated, train))
		u.broadcaster.Publish(broadcaster.NewEvent(broadcaster.TrainDispatched, train))
	}

	return nil
}

func computeStepDuration(index int, stationID uuid.UUID, route []uuid.UUID,
	edgeDistMap map[[2]uuid.UUID]float64, stationTypeMap map[uuid.UUID]entity.Type) float64 {
	if index == 0 {
		return 0
	}

	var durationHours float64

	prevStationID := route[index-1]
	key := [2]uuid.UUID{prevStationID, stationID}
	if dist, ok := edgeDistMap[key]; ok {
		durationHours = dist / 40.0
	} else {
		log.Printf("match: no edge found between %s and %s, using 0 travel time",
			prevStationID, stationID)
	}

	switch stationTypeMap[stationID] {
	case entity.Sorting:
		durationHours += 8.0
	case entity.Border:
		durationHours += 4.0
	}

	return durationHours
}

func buildEdgeDistanceMap(edges []*entity.Edge) map[[2]uuid.UUID]float64 {
	m := make(map[[2]uuid.UUID]float64, len(edges))
	for _, e := range edges {
		if !e.IsActive {
			continue
		}
		m[[2]uuid.UUID{e.FromStationID, e.ToStationID}] = e.DistanceKM
	}
	return m
}

func buildStationTypeMap(stations []*entity.Station) map[uuid.UUID]entity.Type {
	m := make(map[uuid.UUID]entity.Type, len(stations))
	for _, s := range stations {
		m[s.ID] = s.Type
	}
	return m
}
