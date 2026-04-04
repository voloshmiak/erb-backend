package usecase

import (
	"context"
	"erb-backend/src/broadcaster"
	"erb-backend/src/entity"
	"time"

	"github.com/pkg/errors"
)

type UnloadWagonsUseCase struct {
	wagons      WagonRepository
	broadcaster Broadcaster
	unloadAfter time.Duration
}

func NewUnloadWagonsUseCase(wagons WagonRepository, b Broadcaster,
	unloadAfter time.Duration) *UnloadWagonsUseCase {
	return &UnloadWagonsUseCase{
		wagons:      wagons,
		broadcaster: b,
		unloadAfter: unloadAfter,
	}
}

func (uc *UnloadWagonsUseCase) Execute(ctx context.Context) error {
	wagons, err := uc.wagons.FindLoadedReadyToUnload(ctx, uc.unloadAfter)
	if err != nil {
		return errors.Wrap(err, "failed to get loaded wagonRepository")
	}

	for _, w := range wagons {
		if err = uc.wagons.UpdateStatus(ctx, w.ID, entity.Idle); err != nil {
			return errors.Wrap(err, "failed to update wagon status "+w.ID.String())
		}
		uc.broadcaster.Publish(broadcaster.NewEvent(broadcaster.WagonUnloaded, w))
	}

	return nil
}
