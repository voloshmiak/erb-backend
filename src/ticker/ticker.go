package ticker

import (
	"context"
	"erb-backend/src/usecase"
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
	t.unloadWagons.Execute(ctx)
	t.dispatchPlanned.Execute(ctx)
	t.advanceRoutes.Execute(ctx)
}
