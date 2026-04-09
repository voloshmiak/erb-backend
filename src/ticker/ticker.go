package ticker

import (
	"context"
	"erb-backend/src/broadcaster"
	"erb-backend/src/repository"
	"erb-backend/src/simclock"
	"erb-backend/src/usecase"
	"log"
	"time"
)

type Ticker struct {
	simClock        *simclock.SimClock
	simStateRepo    *repository.SimStateRepository
	dispatchPlanned *usecase.DispatchPlannedUseCase
	advanceRoutes   *usecase.AdvanceRoutesUseCase
	unloadWagons    *usecase.UnloadWagonsUseCase
	b               *broadcaster.Broadcaster
	persistEvery    int
	tickCount       int
}

func NewTicker(
	clock *simclock.SimClock,
	simStateRepo *repository.SimStateRepository,
	dispatchPlanned *usecase.DispatchPlannedUseCase,
	advanceRoutes *usecase.AdvanceRoutesUseCase,
	unloadWagons *usecase.UnloadWagonsUseCase,
	b *broadcaster.Broadcaster,
) *Ticker {
	return &Ticker{
		simClock:        clock,
		simStateRepo:    simStateRepo,
		dispatchPlanned: dispatchPlanned,
		advanceRoutes:   advanceRoutes,
		unloadWagons:    unloadWagons,
		b:               b,
		persistEvery:    10,
	}
}

func (t *Ticker) Run(ctx context.Context) {
	for {
		select {
		case <-time.After(t.simClock.TickInterval()):
			t.tick(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func (t *Ticker) tick(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("ticker: panic recovered: ", r)
		}
	}()

	currentHour := t.simClock.Tick()

	t.tickCount++
	if t.tickCount%t.persistEvery == 0 {
		if err := t.simStateRepo.Save(ctx, currentHour); err != nil {
			log.Println("ticker: failed to persist sim state:", err)
		}
	}

	t.b.Publish(broadcaster.NewEvent(broadcaster.SimTick, map[string]any{
		"currentHour": currentHour,
		"displayTime": t.simClock.ToDisplayTime(currentHour),
		"speed":       t.simClock.Speed(),
	}))

	err := t.unloadWagons.Execute(ctx, currentHour)
	if err != nil {
		log.Println("ticker: failed to unload wagons: ", err)
		return
	}
	err = t.dispatchPlanned.Execute(ctx, currentHour)
	if err != nil {
		log.Println("ticker: failed to dispatch planned: ", err)
		return
	}
	err = t.advanceRoutes.Execute(ctx, currentHour)
	if err != nil {
		log.Println("ticker: failed to advance routes: ", err)
		return
	}
}
