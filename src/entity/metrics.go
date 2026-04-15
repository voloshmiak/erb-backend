package entity

// MatchingMetrics holds aggregate stats from a single matching service run.
// json tags must match the wire format from the matching service.
type MatchingMetrics struct {
	TotalEmptyKM    float64 `json:"total_empty_km"`
	AvgEmptyRunKM   float64 `json:"avg_empty_run_km"`
	TotalCost       float64 `json:"total_cost"`
	NaiveTotalCost  float64 `json:"naive_total_cost"`
	CostSaved       float64 `json:"cost_saved"`
	MatchRate       float64 `json:"match_rate"`
	WagonsMatched   int     `json:"wagons_matched"`
	OrdersMatched   int     `json:"orders_matched"`
	OrdersUnmatched int     `json:"orders_unmatched"`
}

// MatchingRunAggregate holds aggregated totals across all persisted matching runs.
type MatchingRunAggregate struct {
	TotalEmptyKM    float64
	TotalCost       float64
	NaiveTotalCost  float64
	CostSaved       float64
	AvgMatchRate    float64
	OrdersMatched   int
	OrdersUnmatched int
	RunCount        int
}
