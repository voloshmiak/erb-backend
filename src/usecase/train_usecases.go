package usecase

import (
	"context"
	"erb-backend/src/broadcaster"
	"erb-backend/src/entity"
	"erb-backend/src/simclock"
	"fmt"
	"math"

	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type CreateTrainUseCase struct {
	trainRepo   TrainRepository
	wagonRepo   WagonRepository
	broadcaster Broadcaster
	clock       *simclock.SimClock
}

func NewCreateTrainUseCase(trainRepo TrainRepository, wagonRepo WagonRepository, b Broadcaster, clock *simclock.SimClock) *CreateTrainUseCase {
	return &CreateTrainUseCase{trainRepo: trainRepo, wagonRepo: wagonRepo, broadcaster: b, clock: clock}
}

func (uc *CreateTrainUseCase) Execute(ctx context.Context, wagonIDs []uuid.UUID, route []uuid.UUID) (*entity.Train, error) {
	if len(route) < 2 {
		return nil, fmt.Errorf("route must have at least 2 stations")
	}

	sourceStationID := route[0]

	// Validate wagons
	wagons, err := uc.wagonRepo.List(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list wagons")
	}

	wagonMap := make(map[uuid.UUID]*entity.Wagon)
	for _, w := range wagons {
		wagonMap[w.ID] = w
	}

	for _, id := range wagonIDs {
		w, ok := wagonMap[id]
		if !ok {
			return nil, fmt.Errorf("wagon %s not found", id)
		}
		if w.CurrentStationID != sourceStationID {
			return nil, fmt.Errorf("wagon %s is at station %s, expected %s", id, w.CurrentStationID, sourceStationID)
		}
		if w.Status != entity.Idle {
			return nil, fmt.Errorf("wagon %s is not idle (status: %s)", id, w.Status)
		}
	}

	train := &entity.Train{
		ID:              uuid.New(),
		WagonIDs:        wagonIDs,
		Route:           route,
		StepIndex:       0,
		SourceStationID: sourceStationID,
		NextStationID:   route[1],
		Status:          entity.TrainForming,
		CreatedAt:       uc.clock.ToDisplayTime(uc.clock.Tick()),
	}

	if err := uc.trainRepo.Create(ctx, train); err != nil {
		return nil, errors.Wrap(err, "failed to create train")
	}

	// Update wagon statuses
	for _, id := range wagonIDs {
		if err := uc.wagonRepo.UpdateStatus(ctx, id, entity.InTrain, nil); err != nil {
			return nil, errors.Wrap(err, "failed to update wagon status")
		}
	}

	uc.broadcaster.Publish(broadcaster.NewEvent(broadcaster.TrainCreated, train))

	return train, nil
}

type DispatchTrainUseCase struct {
	trainRepo      TrainRepository
	stationRepo    StationRepository
	locomotiveRepo LocomotiveRepository
	broadcaster    Broadcaster
	clock          *simclock.SimClock
}

func NewDispatchTrainUseCase(trainRepo TrainRepository, stationRepo StationRepository, locomotiveRepo LocomotiveRepository, b Broadcaster, clock *simclock.SimClock) *DispatchTrainUseCase {
	return &DispatchTrainUseCase{trainRepo: trainRepo, stationRepo: stationRepo, locomotiveRepo: locomotiveRepo, broadcaster: b, clock: clock}
}

func (uc *DispatchTrainUseCase) Execute(ctx context.Context, trainID uuid.UUID, simHour int64) error {
	train, err := uc.trainRepo.GetByID(ctx, trainID)
	if err != nil {
		return err
	}

	if train.Status != entity.TrainForming {
		return fmt.Errorf("train %s is not in forming status", trainID)
	}

	// Assign locomotive
	loco, err := uc.locomotiveRepo.GetAvailableAtStation(ctx, train.SourceStationID)
	if err != nil {
		return fmt.Errorf("no locomotive available at station %s: %w", train.SourceStationID, err)
	}

	_, edges, err := uc.stationRepo.List(ctx)
	if err != nil {
		return err
	}

	// Calculate total route duration
	edgeDistMap := make(map[[2]uuid.UUID]float64)
	for _, e := range edges {
		if e.IsActive {
			edgeDistMap[[2]uuid.UUID{e.FromStationID, e.ToStationID}] = e.DistanceKM
		}
	}

	var totalDuration float64
	for i := 0; i < len(train.Route)-1; i++ {
		from := train.Route[i]
		to := train.Route[i+1]
		dist, ok := edgeDistMap[[2]uuid.UUID{from, to}]
		if !ok {
			return fmt.Errorf("no edge found between %s and %s", from, to)
		}
		totalDuration += dist / 60.0 // Assuming 60 km/h
	}

	// Set current segment duration
	var currentSegmentDuration float64
	dist, ok := edgeDistMap[[2]uuid.UUID{train.SourceStationID, train.NextStationID}]
	if !ok {
		return fmt.Errorf("no edge found between %s and %s", train.SourceStationID, train.NextStationID)
	}
	currentSegmentDuration = dist / 60.0

	train.Status = entity.TrainInTransit
	train.ActivatedAtHour = simHour
	train.DurationHours = currentSegmentDuration

	now := uc.clock.ToDisplayTime(simHour)
	train.DepartedAt = &now

	if err := uc.trainRepo.UpdateStatus(ctx, train.ID, entity.TrainInTransit, now); err != nil {
		return err
	}

	train.Status = entity.TrainInTransit // ensure it's set for progress update
	if err := uc.trainRepo.UpdateProgress(ctx, train); err != nil {
		return err
	}

	// Update locomotive
	loco.Status = entity.LocomotiveInTransit
	loco.TrainID = &train.ID
	loco.AvailableAtHour = simHour + int64(math.Ceil(totalDuration))
	loco.AvailableAt = uc.clock.ToDisplayTime(loco.AvailableAtHour)
	loco.CurrentStationID = train.Route[len(train.Route)-1] // Destination station

	if err := uc.locomotiveRepo.Update(ctx, loco); err != nil {
		return fmt.Errorf("failed to update locomotive: %w", err)
	}

	uc.broadcaster.Publish(broadcaster.NewEvent(broadcaster.TrainDispatched, train))

	return nil
}

type AdvanceTrainsUseCase struct {
	trainRepo   TrainRepository
	wagonRepo   WagonRepository
	stationRepo StationRepository
	broadcaster Broadcaster
	clock       *simclock.SimClock
}

func NewAdvanceTrainsUseCase(trainRepo TrainRepository, wagonRepo WagonRepository, stationRepo StationRepository, b Broadcaster, clock *simclock.SimClock) *AdvanceTrainsUseCase {
	return &AdvanceTrainsUseCase{trainRepo: trainRepo, wagonRepo: wagonRepo, stationRepo: stationRepo, broadcaster: b, clock: clock}
}

func (uc *AdvanceTrainsUseCase) Execute(ctx context.Context, simHour int64) error {
	trains, err := uc.trainRepo.ListActive(ctx)
	if err != nil {
		return err
	}

	for _, t := range trains {
		if t.Status != entity.TrainInTransit {
			continue
		}

		if simHour < t.ActivatedAtHour+int64(math.Ceil(t.DurationHours)) {
			continue
		}

		if err := uc.advanceTrain(ctx, t, simHour); err != nil {
			fmt.Printf("failed to advance train %s: %v\n", t.ID, err)
		}
	}

	return nil
}

func (uc *AdvanceTrainsUseCase) advanceTrain(ctx context.Context, t *entity.Train, simHour int64) error {
	// Update all wagons position
	stations, edges, err := uc.stationRepo.List(ctx)
	if err != nil {
		return err
	}

	stationMap := make(map[uuid.UUID]*entity.Station)
	for _, s := range stations {
		stationMap[s.ID] = s
	}

	nextStation := stationMap[t.NextStationID]

	for _, wagonID := range t.WagonIDs {
		if err := uc.wagonRepo.UpdateStation(ctx, wagonID, t.NextStationID); err != nil {
			return err
		}
		// Emit WagonMoved
		uc.broadcaster.Publish(broadcaster.NewEvent(broadcaster.WagonMoved, entity.WagonMovedPayload{
			WagonID:     wagonID,
			StationID:   t.NextStationID,
			StationName: nextStation.Name,
			Lat:         nextStation.Location.Latitude,
			Lng:         nextStation.Location.Longitude,
			StepIndex:   t.StepIndex,
			TotalSteps:  len(t.Route) - 1,
			ArrivedAt:   uc.clock.ToDisplayTime(simHour),
		}))
	}

	if t.StepIndex+1 < len(t.Route)-1 {
		// Next segment
		t.StepIndex++
		t.SourceStationID = t.Route[t.StepIndex]
		t.NextStationID = t.Route[t.StepIndex+1]

		var duration float64
		found := false
		for _, e := range edges {
			if e.FromStationID == t.SourceStationID && e.ToStationID == t.NextStationID {
				duration = e.DistanceKM / 60.0
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("no edge found for next segment of train %s", t.ID)
		}

		t.ActivatedAtHour = simHour
		t.DurationHours = duration
	} else {
		// Arrived at destination
		t.Status = entity.TrainArrived
		now := uc.clock.ToDisplayTime(simHour)
		t.ArrivedAt = &now

		for _, wagonID := range t.WagonIDs {
			if err := uc.wagonRepo.UpdateStatus(ctx, wagonID, entity.Idle, nil); err != nil {
				return err
			}
		}
		uc.broadcaster.Publish(broadcaster.NewEvent(broadcaster.TrainArrived, t))
	}

	return uc.trainRepo.UpdateProgress(ctx, t)
}

type ListTrainsUseCase struct {
	trainRepo TrainRepository
}

func NewListTrainsUseCase(trainRepo TrainRepository) *ListTrainsUseCase {
	return &ListTrainsUseCase{trainRepo: trainRepo}
}

func (uc *ListTrainsUseCase) Execute(ctx context.Context) ([]*entity.Train, error) {
	return uc.trainRepo.ListActive(ctx)
}

type GetTrainUseCase struct {
	trainRepo TrainRepository
}

func NewGetTrainUseCase(trainRepo TrainRepository) *GetTrainUseCase {
	return &GetTrainUseCase{trainRepo: trainRepo}
}

func (uc *GetTrainUseCase) Execute(ctx context.Context, id uuid.UUID) (*entity.Train, error) {
	return uc.trainRepo.GetByID(ctx, id)
}
