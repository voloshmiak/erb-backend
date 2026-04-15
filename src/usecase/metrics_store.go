package usecase

import (
	"context"
	"log"
	"sync"

	"erb-backend/src/entity"
)

const (
	CostPerEmptyKM     = 20 // UAH/km
	RevenuePerLoadedKM = 30 // UAH/km
)

// MetricsStore accumulates matching metrics across all matching runs.
// In-memory state is backed by DB: persisted on each Accumulate, restored via Load on startup.
type MetricsStore struct {
	mu              sync.Mutex
	repo            MatchingRunRepository
	totalEmptyKM    float64
	totalCost       float64
	naiveTotalCost  float64
	costSaved       float64
	matchRate       float64
	ordersMatched   int
	ordersUnmatched int
	runCount        int
}

func NewMetricsStore(repo MatchingRunRepository) *MetricsStore {
	return &MetricsStore{repo: repo}
}

// Load restores accumulated state from DB. Call once on startup before serving requests.
func (s *MetricsStore) Load(ctx context.Context) error {
	agg, err := s.repo.LoadAggregated(ctx)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.totalEmptyKM = agg.TotalEmptyKM
	s.totalCost = agg.TotalCost
	s.naiveTotalCost = agg.NaiveTotalCost
	s.costSaved = agg.CostSaved
	s.matchRate = agg.AvgMatchRate
	s.ordersMatched = agg.OrdersMatched
	s.ordersUnmatched = agg.OrdersUnmatched
	s.runCount = agg.RunCount
	return nil
}

// Accumulate persists one matching run to DB and merges it into the in-memory state.
func (s *MetricsStore) Accumulate(ctx context.Context, m *entity.MatchingMetrics) {
	if m == nil {
		return
	}
	if err := s.repo.Save(ctx, m); err != nil {
		log.Printf("metrics_store: failed to persist matching run: %v", err)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.totalEmptyKM += m.TotalEmptyKM
	s.totalCost += m.TotalCost
	s.naiveTotalCost += m.NaiveTotalCost
	s.costSaved += m.CostSaved
	s.ordersMatched += m.OrdersMatched
	s.ordersUnmatched += m.OrdersUnmatched
	s.runCount++
	s.matchRate += (m.MatchRate - s.matchRate) / float64(s.runCount) // Welford running avg
}

// MetricsSnapshot is an immutable point-in-time copy of MetricsStore state.
type MetricsSnapshot struct {
	TotalEmptyKM    float64
	TotalCost       float64
	NaiveTotalCost  float64
	CostSaved       float64
	MatchRate       float64
	OrdersMatched   int
	OrdersUnmatched int
}

// Snapshot returns a consistent copy of accumulated metrics.
func (s *MetricsStore) Snapshot() MetricsSnapshot {
	s.mu.Lock()
	defer s.mu.Unlock()
	return MetricsSnapshot{
		TotalEmptyKM:    s.totalEmptyKM,
		TotalCost:       s.totalCost,
		NaiveTotalCost:  s.naiveTotalCost,
		CostSaved:       s.costSaved,
		MatchRate:       s.matchRate,
		OrdersMatched:   s.ordersMatched,
		OrdersUnmatched: s.ordersUnmatched,
	}
}
