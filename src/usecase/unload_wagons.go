package usecase

import (
	"context"
	"erb-backend/src/broadcaster"
	"erb-backend/src/entity"
	"log"
	"time"
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

func (uc *UnloadWagonsUseCase) Execute(ctx context.Context) {
	wagons, err := uc.wagons.GetLoadedReadyToUnload(ctx, uc.unloadAfter)
	if err != nil {
		log.Printf("unload_wagons: failed to get loaded wagons: %v", err)
		return
	}

	for _, w := range wagons {
		if err = uc.wagons.UpdateStatus(ctx, w.ID, entity.Idle); err != nil {
			log.Printf("unload_wagons: failed to update wagon status %s: %v", w.ID, err)
			continue
		}
		uc.broadcaster.Publish(broadcaster.NewEvent(broadcaster.WagonUnloaded, w))
	}
}
