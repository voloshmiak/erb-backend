package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"erb-backend/src/entity"
	"time"

	"github.com/google/uuid"
)

type TrainRepository struct {
	conn *sql.DB
}

func NewTrainRepository(conn *sql.DB) *TrainRepository {
	return &TrainRepository{conn: conn}
}

func (r *TrainRepository) Create(ctx context.Context, t *entity.Train) error {
	wagonIDsJSON, err := json.Marshal(t.WagonIDs)
	if err != nil {
		return err
	}
	routeJSON, err := json.Marshal(t.Route)
	if err != nil {
		return err
	}

	_, err = r.conn.ExecContext(ctx, `
		INSERT INTO trains (
			id, wagon_ids, route, step_index, source_station_id, next_station_id,
			status, created_at, activated_at_hour, duration_hours
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, t.ID, wagonIDsJSON, routeJSON, t.StepIndex, t.SourceStationID, t.NextStationID,
		t.Status, t.CreatedAt, t.ActivatedAtHour, t.DurationHours)
	return err
}

func (r *TrainRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Train, error) {
	t := &entity.Train{}
	var wagonIDsJSON, routeJSON []byte
	var departedAt, arrivedAt sql.NullTime

	err := r.conn.QueryRowContext(ctx, `
		SELECT id, wagon_ids, route, step_index, source_station_id, next_station_id,
			   status, created_at, departed_at, arrived_at, activated_at_hour, duration_hours
		FROM trains WHERE id = $1
	`, id).Scan(&t.ID, &wagonIDsJSON, &routeJSON, &t.StepIndex, &t.SourceStationID, &t.NextStationID,
		&t.Status, &t.CreatedAt, &departedAt, &arrivedAt, &t.ActivatedAtHour, &t.DurationHours)

	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(wagonIDsJSON, &t.WagonIDs); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(routeJSON, &t.Route); err != nil {
		return nil, err
	}

	if departedAt.Valid {
		t.DepartedAt = &departedAt.Time
	}
	if arrivedAt.Valid {
		t.ArrivedAt = &arrivedAt.Time
	}
	return t, nil
}

func (r *TrainRepository) ListActive(ctx context.Context) ([]*entity.Train, error) {
	rows, err := r.conn.QueryContext(ctx, `
		SELECT id, wagon_ids, route, step_index, source_station_id, next_station_id,
			   status, created_at, departed_at, arrived_at, activated_at_hour, duration_hours
		FROM trains
		WHERE status != $1
	`, entity.TrainArrived)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var trains []*entity.Train
	for rows.Next() {
		t := &entity.Train{}
		var wagonIDsJSON, routeJSON []byte
		var departedAt, arrivedAt sql.NullTime
		if err := rows.Scan(&t.ID, &wagonIDsJSON, &routeJSON, &t.StepIndex, &t.SourceStationID, &t.NextStationID,
			&t.Status, &t.CreatedAt, &departedAt, &arrivedAt, &t.ActivatedAtHour, &t.DurationHours); err != nil {
			return nil, err
		}

		if err := json.Unmarshal(wagonIDsJSON, &t.WagonIDs); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(routeJSON, &t.Route); err != nil {
			return nil, err
		}

		if departedAt.Valid {
			t.DepartedAt = &departedAt.Time
		}
		if arrivedAt.Valid {
			t.ArrivedAt = &arrivedAt.Time
		}
		trains = append(trains, t)
	}
	return trains, rows.Err()
}

func (r *TrainRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status entity.TrainStatus, at time.Time) error {
	var query string
	if status == entity.TrainInTransit {
		query = "UPDATE trains SET status = $1, departed_at = $2 WHERE id = $3"
	} else if status == entity.TrainArrived {
		query = "UPDATE trains SET status = $1, arrived_at = $2 WHERE id = $3"
	} else {
		query = "UPDATE trains SET status = $1 WHERE id = $3"
		_, err := r.conn.ExecContext(ctx, query, status, id)
		return err
	}
	_, err := r.conn.ExecContext(ctx, query, status, at, id)
	return err
}

func (r *TrainRepository) UpdateProgress(ctx context.Context, t *entity.Train) error {
	_, err := r.conn.ExecContext(ctx, `
		UPDATE trains SET step_index = $1, next_station_id = $2,
		activated_at_hour = $3, duration_hours = $4, status = $5, arrived_at = $6
		WHERE id = $7
	`, t.StepIndex, t.NextStationID, t.ActivatedAtHour, t.DurationHours, t.Status, t.ArrivedAt, t.ID)
	return err
}
