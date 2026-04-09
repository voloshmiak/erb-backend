package usecase

import (
	"context"
	"erb-backend/src/broadcaster"
	"erb-backend/src/entity"

	"github.com/pkg/errors"
)

type UnloadWagonsUseCase struct {
	wagons      WagonRepository
	broadcaster Broadcaster
}

func NewUnloadWagonsUseCase(wagons WagonRepository, b Broadcaster) *UnloadWagonsUseCase {
	return &UnloadWagonsUseCase{
		wagons:      wagons,
		broadcaster: b,
	}
}

func (uc *UnloadWagonsUseCase) Execute(ctx context.Context, simHour int64) error {
	wagons, err := uc.wagons.FindLoadedReadyToUnload(ctx, simHour)
	if err != nil {
		return errors.Wrap(err, "failed to get loaded wagonRepository")
	}

	for _, w := range wagons {
		if err = uc.wagons.UpdateStatus(ctx, w.ID, entity.Idle, nil); err != nil {
			return errors.Wrap(err, "failed to update wagon status "+w.ID.String())
		}
		uc.broadcaster.Publish(broadcaster.NewEvent(broadcaster.WagonUnloaded, w))
	}

	return nil
}
