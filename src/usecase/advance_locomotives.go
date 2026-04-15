package usecase

import (
	"context"
	"erb-backend/src/broadcaster"
	"erb-backend/src/entity"
)

type AdvanceLocomotivesUseCase struct {
	locomotiveRepo LocomotiveRepository
	broadcaster    Broadcaster
}

func NewAdvanceLocomotivesUseCase(locomotiveRepo LocomotiveRepository, b Broadcaster) *AdvanceLocomotivesUseCase {
	return &AdvanceLocomotivesUseCase{locomotiveRepo: locomotiveRepo, broadcaster: b}
}

func (uc *AdvanceLocomotivesUseCase) Execute(ctx context.Context, simHour int64) error {
	// Mark locomotives as idle if their AvailableAtHour is reached
	// We can fetch them first to emit events, or just do a bulk update
	// To emit events, we'll fetch them
	locos, err := uc.locomotiveRepo.ListByStatus(ctx, entity.LocomotiveInTransit)
	if err != nil {
		return err
	}

	for _, l := range locos {
		if simHour >= l.AvailableAtHour {
			l.Status = entity.LocomotiveIdle
			l.TrainID = nil
			if err := uc.locomotiveRepo.Update(ctx, l); err != nil {
				return err
			}
			uc.broadcaster.Publish(broadcaster.NewEvent(broadcaster.LocomotiveIdle, l))
		}
	}

	return nil
}
