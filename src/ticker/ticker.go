package ticker

import (
	"context"
	"erb-backend/src/usecase"
	"log"
	"time"
)

type Ticker struct {
	ticker          *time.Ticker
	dispatchPlanned *usecase.DispatchPlannedUseCase
	advanceRoutes   *usecase.AdvanceRoutesUseCase
	unloadWagons    *usecase.UnloadWagonsUseCase
}

func NewTicker(
	interval time.Duration,
	dispatchPlanned *usecase.DispatchPlannedUseCase,
	advanceRoutes *usecase.AdvanceRoutesUseCase,
	unloadWagons *usecase.UnloadWagonsUseCase,
) *Ticker {
	return &Ticker{
		ticker:          time.NewTicker(interval),
		dispatchPlanned: dispatchPlanned,
		advanceRoutes:   advanceRoutes,
		unloadWagons:    unloadWagons,
	}
}

func (t *Ticker) Run(ctx context.Context) {
	for {
		select {
		case <-t.ticker.C:
			t.tick(ctx)
		case <-ctx.Done():
			t.ticker.Stop()
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
	err := t.unloadWagons.Execute(ctx)
	if err != nil {
		log.Println("ticker: failed to unload wagons: ", err)
		return
	}
	err = t.dispatchPlanned.Execute(ctx)
	if err != nil {
		log.Println("ticker: failed to dispatch planned: ", err)
		return
	}
	err = t.advanceRoutes.Execute(ctx)
	if err != nil {
		log.Println("ticker: failed to advance routes: ", err)
		return
	}
}
