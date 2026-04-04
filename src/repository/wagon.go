package repository

import (
	"context"
	"database/sql"
	"erb-backend/src/entity"
	"log"
	"time"

	"github.com/google/uuid"
)

type WagonRepository struct {
	conn *sql.DB
}

func NewWagonRepository(conn *sql.DB) *WagonRepository {
	return &WagonRepository{conn: conn}
}

func (r *WagonRepository) UpdateStation(ctx context.Context, wagonID, stationID uuid.UUID) error {
	_, err := r.conn.ExecContext(ctx,
		"UPDATE wagons SET current_station_id = $1 WHERE id = $2",
		stationID, wagonID,
	)
	return err
}

func (r *WagonRepository) UpdateStatus(ctx context.Context,
	wagonID uuid.UUID, status entity.WagonStatus) error {
	_, err := r.conn.ExecContext(ctx,
		"UPDATE wagons SET status = $1 WHERE id = $2",
		status, wagonID,
	)
	return err
}

func (r *WagonRepository) FindLoadedReadyToUnload(ctx context.Context,
	olderThan time.Duration) ([]*entity.LoadedWagon, error) {
	rows, err := r.conn.QueryContext(ctx, `
		SELECT w.id::text, w.wagon_number
		FROM wagons w
		WHERE w.status = $1
		AND (
			SELECT MAX(rs.arrived_at)
			FROM route_steps rs
			JOIN assignments a ON a.id = rs.assignment_id
			WHERE a.wagon_id = w.id AND rs.arrived_at IS NOT NULL
		) < NOW() - ($2 * interval '1 second')
	`, entity.Loaded, olderThan.Seconds())
	if err != nil {
		return nil, err
	}
	defer func() {
		if err = rows.Close(); err != nil {
			log.Println("failed to close rows:", err)
		}
	}()

	var wagons []*entity.LoadedWagon
	for rows.Next() {
		var idStr string
		w := &entity.LoadedWagon{}
		if err = rows.Scan(&idStr, &w.WagonNumber); err != nil {
			return nil, err
		}
		if w.ID, err = uuid.Parse(idStr); err != nil {
			return nil, err
		}
		wagons = append(wagons, w)
	}
	return wagons, rows.Err()
}

func (r *WagonRepository) ListStatusCounts(ctx context.Context) ([]entity.WagonStatusCount, error) {
	rows, err := r.conn.QueryContext(ctx, `
		SELECT wagon_type, status, COUNT(*) AS count
		FROM wagons
		GROUP BY wagon_type, status
	`)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err = rows.Close(); err != nil {
			log.Println("failed to close rows:", err)
		}
	}()

	var counts []entity.WagonStatusCount
	for rows.Next() {
		var c entity.WagonStatusCount
		if err := rows.Scan(&c.Type, &c.Status, &c.Count); err != nil {
			return nil, err
		}
		counts = append(counts, c)
	}
	return counts, rows.Err()
}

func (r *WagonRepository) ListByStatus(ctx context.Context,
	status entity.WagonStatus) ([]*entity.Wagon, error) {
	rows, err := r.conn.QueryContext(ctx, `
		SELECT id::text, wagon_number, wagon_type, current_station_id::text
		FROM wagons
		WHERE status = $1
	`, status)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err = rows.Close(); err != nil {
			log.Println("failed to close rows:", err)
		}
	}()

	var wagons []*entity.Wagon
	for rows.Next() {
		var w entity.Wagon
		var stationIDStr string
		if err = rows.Scan(&w.ID, &w.Number, &w.Type, &stationIDStr); err != nil {
			return nil, err
		}
		if w.CurrentStationID, err = uuid.Parse(stationIDStr); err != nil {
			return nil, err
		}
		wagons = append(wagons, &w)
	}
	return wagons, rows.Err()
}
