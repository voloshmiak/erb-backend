package usecase

import (
	"context"

	"github.com/pkg/errors"
)

// MetricsOutput is the combined financial metrics returned by GET /api/metrics.
type MetricsOutput struct {
	TotalDeliveredAssignments int     `json:"totalDeliveredAssignments"`
	TotalEmptyRunKm           int     `json:"totalEmptyRunKm"`
	TotalCostEmptyRun         int     `json:"totalCostEmptyRun"`
	TotalLoadedRunKm          int     `json:"totalLoadedRunKm"`
	TotalRevenue              int     `json:"totalRevenue"`
	CostSavedVsNaive          float64 `json:"costSavedVsNaive"`
	NaiveCost                 float64 `json:"naiveCost"`
	OptimizedCost             float64 `json:"optimizedCost"`
	MatchRate                 float64 `json:"matchRate"`
	AvgEmptyRunKm             float64 `json:"avgEmptyRunKm"`
	OrdersMatched             int     `json:"ordersMatched"`
	OrdersUnmatched           int     `json:"ordersUnmatched"`
}

type GetMetricsUseCase struct {
	assignmentRepository AssignmentRepository
	metricsStore         *MetricsStore
}

func NewGetMetricsUseCase(repo AssignmentRepository, store *MetricsStore) *GetMetricsUseCase {
	return &GetMetricsUseCase{assignmentRepository: repo, metricsStore: store}
}

func (u *GetMetricsUseCase) Execute(ctx context.Context) (*MetricsOutput, error) {
	stats, err := u.assignmentRepository.GetStats(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get assignment stats")
	}

	snap := u.metricsStore.Snapshot()

	return &MetricsOutput{
		TotalDeliveredAssignments: stats.TotalDelivered,
		TotalEmptyRunKm:           stats.TotalEmptyRunKM,
		TotalCostEmptyRun:         stats.TotalCostEmptyRun,
		TotalLoadedRunKm:          stats.TotalLoadedRunKM,
		TotalRevenue:              stats.TotalLoadedRunKM * RevenuePerLoadedKM,
		CostSavedVsNaive:          snap.CostSaved,
		NaiveCost:                 snap.NaiveTotalCost,
		OptimizedCost:             snap.TotalCost,
		MatchRate:                 snap.MatchRate,
		AvgEmptyRunKm:             stats.AvgEmptyRunKM,
		OrdersMatched:             snap.OrdersMatched,
		OrdersUnmatched:           snap.OrdersUnmatched,
	}, nil
}
