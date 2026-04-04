package gateway

import (
	"context"
	"database/sql"
	"erb-backend/src/entity"
	"time"

	"github.com/google/uuid"
)

type MockMatchingGateway struct {
	conn *sql.DB
}

func NewMockMatchingGateway(conn *sql.DB) *MockMatchingGateway {
	return &MockMatchingGateway{conn: conn}
}

type idleWagon struct {
	id               uuid.UUID
	wagonType        entity.WagonType
	currentStationID uuid.UUID
}

func (g *MockMatchingGateway) Match(_ context.Context, orders []*entity.Order,
	_ []entity.WagonStatusCount) ([]*entity.AssignmentResult, error) {

	idle, err := g.queryIdleWagons()
	if err != nil {
		return nil, err
	}

	// index idle wagons by type for fast lookup
	byType := make(map[entity.WagonType][]idleWagon)
	for _, w := range idle {
		byType[w.wagonType] = append(byType[w.wagonType], w)
	}

	var results []*entity.AssignmentResult
	for _, order := range orders {
		pool := byType[order.WagonType]
		if len(pool) == 0 {
			continue
		}

		wagon := pool[0]
		byType[order.WagonType] = pool[1:] // mark as consumed

		results = append(results, &entity.AssignmentResult{
			Assignment: &entity.Assignment{
				OrderID:          order.ID,
				WagonID:          wagon.id,
				EstimatedArrival: time.Now().Add(2 * time.Hour),
			},
			Route: []uuid.UUID{wagon.currentStationID, order.StationToID},
		})
	}

	return results, nil
}

func (g *MockMatchingGateway) queryIdleWagons() ([]idleWagon, error) {
	rows, err := g.conn.Query(`
		SELECT id::text, wagon_type, current_station_id::text
		FROM wagons
		WHERE status = $1 AND current_station_id IS NOT NULL
	`, entity.Idle)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var wagons []idleWagon
	for rows.Next() {
		var idStr, stationStr string
		var wType entity.WagonType
		if err = rows.Scan(&idStr, &wType, &stationStr); err != nil {
			return nil, err
		}
		id, err := uuid.Parse(idStr)
		if err != nil {
			return nil, err
		}
		stationID, err := uuid.Parse(stationStr)
		if err != nil {
			return nil, err
		}
		wagons = append(wagons, idleWagon{id: id, wagonType: wType, currentStationID: stationID})
	}
	return wagons, rows.Err()
}
