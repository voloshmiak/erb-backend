package repository

import (
	"context"
	"database/sql"
	"erb-backend/src/entity"

	"github.com/google/uuid"
)

type LocomotiveRepository struct {
	conn *sql.DB
}

func NewLocomotiveRepository(conn *sql.DB) *LocomotiveRepository {
	return &LocomotiveRepository{conn: conn}
}

func (r *LocomotiveRepository) List(ctx context.Context) ([]*entity.Locomotive, error) {
	rows, err := r.conn.QueryContext(ctx, `
		SELECT id, current_station_id, status, available_at, available_at_hour, train_id 
		FROM locomotives
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var locos []*entity.Locomotive
	for rows.Next() {
		var l entity.Locomotive
		var trainID sql.NullString
		if err := rows.Scan(&l.ID, &l.CurrentStationID, &l.Status, &l.AvailableAt, &l.AvailableAtHour, &trainID); err != nil {
			return nil, err
		}
		if trainID.Valid {
			id, _ := uuid.Parse(trainID.String)
			l.TrainID = &id
		}
		locos = append(locos, &l)
	}
	return locos, nil
}

func (r *LocomotiveRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Locomotive, error) {
	var l entity.Locomotive
	var trainID sql.NullString
	err := r.conn.QueryRowContext(ctx, `
		SELECT id, current_station_id, status, available_at, available_at_hour, train_id 
		FROM locomotives WHERE id = $1
	`, id).Scan(&l.ID, &l.CurrentStationID, &l.Status, &l.AvailableAt, &l.AvailableAtHour, &trainID)
	if err != nil {
		return nil, err
	}
	if trainID.Valid {
		id, _ := uuid.Parse(trainID.String)
		l.TrainID = &id
	}
	return &l, nil
}

func (r *LocomotiveRepository) ListByStatus(ctx context.Context, status entity.LocomotiveStatus) ([]*entity.Locomotive, error) {
	rows, err := r.conn.QueryContext(ctx, `
		SELECT id, current_station_id, status, available_at, available_at_hour, train_id 
		FROM locomotives WHERE status = $1
	`, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var locos []*entity.Locomotive
	for rows.Next() {
		var l entity.Locomotive
		var trainID sql.NullString
		if err := rows.Scan(&l.ID, &l.CurrentStationID, &l.Status, &l.AvailableAt, &l.AvailableAtHour, &trainID); err != nil {
			return nil, err
		}
		if trainID.Valid {
			id, _ := uuid.Parse(trainID.String)
			l.TrainID = &id
		}
		locos = append(locos, &l)
	}
	return locos, nil
}

func (r *LocomotiveRepository) GetAvailableAtStation(ctx context.Context, stationID uuid.UUID) (*entity.Locomotive, error) {
	var l entity.Locomotive
	var trainID sql.NullString
	err := r.conn.QueryRowContext(ctx, `
		SELECT id, current_station_id, status, available_at, available_at_hour, train_id 
		FROM locomotives 
		WHERE current_station_id = $1 AND status = 'idle'
		LIMIT 1
	`, stationID).Scan(&l.ID, &l.CurrentStationID, &l.Status, &l.AvailableAt, &l.AvailableAtHour, &trainID)
	if err != nil {
		return nil, err
	}
	if trainID.Valid {
		id, _ := uuid.Parse(trainID.String)
		l.TrainID = &id
	}
	return &l, nil
}

func (r *LocomotiveRepository) Update(ctx context.Context, l *entity.Locomotive) error {
	_, err := r.conn.ExecContext(ctx, `
		UPDATE locomotives 
		SET current_station_id = $1, status = $2, available_at = $3, available_at_hour = $4, train_id = $5 
		WHERE id = $6
	`, l.CurrentStationID, l.Status, l.AvailableAt, l.AvailableAtHour, l.TrainID, l.ID)
	return err
}

func (r *LocomotiveRepository) MarkIdleIfAvailable(ctx context.Context, simHour int64) error {
	_, err := r.conn.ExecContext(ctx, `
		UPDATE locomotives 
		SET status = 'idle', train_id = NULL 
		WHERE status = 'in_transit' AND available_at_hour <= $1
	`, simHour)
	return err
}
