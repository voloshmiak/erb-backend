package repository

import (
	"context"
	"database/sql"
	"erb-backend/src/entity"
)

type MatchingRunRepository struct {
	conn *sql.DB
}

func NewMatchingRunRepository(conn *sql.DB) *MatchingRunRepository {
	return &MatchingRunRepository{conn: conn}
}

func (r *MatchingRunRepository) Save(ctx context.Context, m *entity.MatchingMetrics) error {
	_, err := r.conn.ExecContext(ctx, `
		INSERT INTO matching_runs (id, total_empty_km, total_cost, naive_total_cost, cost_saved, match_rate, orders_matched, orders_unmatched)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7)
	`, m.TotalEmptyKM, m.TotalCost, m.NaiveTotalCost, m.CostSaved, m.MatchRate, m.OrdersMatched, m.OrdersUnmatched)
	return err
}

func (r *MatchingRunRepository) LoadAggregated(ctx context.Context) (*entity.MatchingRunAggregate, error) {
	agg := &entity.MatchingRunAggregate{}
	err := r.conn.QueryRowContext(ctx, `
		SELECT
			COUNT(*),
			COALESCE(SUM(total_empty_km), 0),
			COALESCE(SUM(total_cost), 0),
			COALESCE(SUM(naive_total_cost), 0),
			COALESCE(SUM(cost_saved), 0),
			COALESCE(AVG(match_rate), 0),
			COALESCE(SUM(orders_matched), 0),
			COALESCE(SUM(orders_unmatched), 0)
		FROM matching_runs
	`).Scan(
		&agg.RunCount,
		&agg.TotalEmptyKM,
		&agg.TotalCost,
		&agg.NaiveTotalCost,
		&agg.CostSaved,
		&agg.AvgMatchRate,
		&agg.OrdersMatched,
		&agg.OrdersUnmatched,
	)
	return agg, err
}
