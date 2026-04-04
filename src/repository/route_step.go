package repository

import (
	"context"
	"database/sql"
	"erb-backend/src/entity"
	"errors"
	"log"

	"github.com/google/uuid"
)

type RouteStepRepository struct {
	conn *sql.DB
}

func NewRouteStepRepository(conn *sql.DB) *RouteStepRepository {
	return &RouteStepRepository{conn: conn}
}

func (r *RouteStepRepository) GetActiveRouteSteps(ctx context.Context) ([]*entity.ActiveRouteStep, error) {
	rows, err := r.conn.QueryContext(ctx, `
		SELECT
			rs.id::text,
			rs.assignment_id::text,
			a.order_id::text,
			a.wagon_id::text,
			w.wagon_number,
			rs.step_index,
			(SELECT COUNT(*) FROM route_steps WHERE assignment_id = rs.assignment_id) AS total_steps
		FROM route_steps rs
		JOIN assignments a ON a.id = rs.assignment_id AND a.status != $1
		JOIN wagons w ON w.id = a.wagon_id
		WHERE rs.status = $2
	`, entity.AssignmentDelivered, entity.RouteStepCurrent)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		if err = rows.Close(); err != nil {
			log.Println("failed to close rows:", err)
		}
	}(rows)

	var steps []*entity.ActiveRouteStep
	for rows.Next() {
		s, err := scanActiveRouteStep(rows)
		if err != nil {
			return nil, err
		}
		steps = append(steps, s)
	}
	return steps, rows.Err()
}

func (r *RouteStepRepository) CompleteRouteStep(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn.ExecContext(ctx,
		"UPDATE route_steps SET status = $2, arrived_at = NOW() WHERE id = $1",
		id, entity.RouteStepCompleted,
	)
	return err
}

func (r *RouteStepRepository) GetNextRouteStep(ctx context.Context, assignmentID uuid.UUID,
	stepIndex int) (*entity.RouteStep, error) {
	var (
		id          string
		stationID   string
		stationName string
		lat, lng    float64
	)
	err := r.conn.QueryRowContext(ctx, `
		SELECT
			rs.id::text,
			rs.station_id::text,
			s.name,
			ST_Y(s.location::geometry),
			ST_X(s.location::geometry)
		FROM route_steps rs
		JOIN stations s ON s.id = rs.station_id
		WHERE rs.assignment_id = $1 AND rs.step_index = $2
	`, assignmentID, stepIndex).Scan(&id, &stationID, &stationName, &lat, &lng)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	parsedID, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}
	parsedStationID, err := uuid.Parse(stationID)
	if err != nil {
		return nil, err
	}

	return &entity.RouteStep{
		ID:          parsedID,
		StationID:   parsedStationID,
		StepIndex:   stepIndex,
		StationName: stationName,
		Lat:         lat,
		Lng:         lng,
	}, nil
}

func (r *RouteStepRepository) CreateForAssignment(ctx context.Context, assignmentID uuid.UUID,
	stepIndex int, stationID uuid.UUID) error {
	_, err := r.conn.ExecContext(ctx, `
		INSERT INTO route_steps (id, assignment_id, station_id, step_index, status)
		VALUES ($1, $2, $3, $4, $5)
	`, uuid.New(), assignmentID, stationID, stepIndex, entity.RouteStepPending)
	return err
}

func (r *RouteStepRepository) ActivateRouteStep(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn.ExecContext(ctx,
		"UPDATE route_steps SET status = $2 WHERE id = $1",
		id, entity.RouteStepCurrent,
	)
	return err
}

func scanActiveRouteStep(rows *sql.Rows) (*entity.ActiveRouteStep, error) {
	var (
		id           string
		assignmentID string
		orderID      string
		wagonID      string
		wagonNumber  string
		stepIndex    int
		totalSteps   int
	)

	if err := rows.Scan(&id, &assignmentID, &orderID, &wagonID,
		&wagonNumber, &stepIndex, &totalSteps); err != nil {
		return nil, err
	}

	s := &entity.ActiveRouteStep{StepIndex: stepIndex, TotalSteps: totalSteps, WagonNumber: wagonNumber}

	for _, pair := range []struct {
		dst *uuid.UUID
		src string
	}{
		{&s.ID, id},
		{&s.AssignmentID, assignmentID},
		{&s.OrderID, orderID},
		{&s.WagonID, wagonID},
	} {
		parsed, err := uuid.Parse(pair.src)
		if err != nil {
			return nil, err
		}
		*pair.dst = parsed
	}

	return s, nil
}
