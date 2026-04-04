package repository

import (
	"database/sql"
	"erb-backend/src/entity"
	"log"
	"time"

	"github.com/google/uuid"
)

type StationRepository struct {
	conn *sql.DB
}

func NewStationRepository(conn *sql.DB) *StationRepository {
	return &StationRepository{conn: conn}
}

func (r *StationRepository) Exists(id uuid.UUID) (bool, error) {
	var exists bool
	err := r.conn.QueryRow("SELECT EXISTS(SELECT 1 FROM stations WHERE id = $1)", id).Scan(&exists)
	return exists, err
}

func (r *StationRepository) List() ([]*entity.Station, []*entity.Edge, error) {
	rows, err := r.conn.Query(`
		SELECT
			s.id::text,
			s.name,
			s.type::text,
			ST_Y(s.location::geometry) AS lat,
			ST_X(s.location::geometry) AS lng,
			s.capacity,
			s.created_at,
			e.id::text,
			e.from_station_id::text,
			e.to_station_id::text,
			e.distance_km,
			e.created_at
		FROM stations s
		LEFT JOIN edges e ON e.from_station_id = s.id AND e.is_active = true
		ORDER BY s.id
	`)
	if err != nil {
		return nil, nil, err
	}
	defer func(rows *sql.Rows) {
		if err = rows.Close(); err != nil {
			log.Println("Failed to close rows:", err)
		}
	}(rows)

	return scanGraph(rows)
}

func scanGraph(rows *sql.Rows) ([]*entity.Station, []*entity.Edge, error) {
	stationOrder := make([]uuid.UUID, 0)
	stationMap := make(map[uuid.UUID]*entity.Station)
	var edges []*entity.Edge

	for rows.Next() {
		station, edge, err := scanRow(rows)
		if err != nil {
			return nil, nil, err
		}

		if _, seen := stationMap[station.ID]; !seen {
			stationOrder = append(stationOrder, station.ID)
			stationMap[station.ID] = station
		}

		if edge != nil {
			edges = append(edges, edge)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	stations := make([]*entity.Station, len(stationOrder))
	for i, id := range stationOrder {
		stations[i] = stationMap[id]
	}

	return stations, edges, nil
}

func scanRow(rows *sql.Rows) (*entity.Station, *entity.Edge, error) {
	var (
		sID        string
		sName      string
		sType      string
		lat, lng   float64
		capacity   int
		sCreatedAt time.Time

		eID        sql.NullString
		eFromID    sql.NullString
		eToID      sql.NullString
		eDistKM    sql.NullInt64
		eCreatedAt sql.NullTime
	)

	if err := rows.Scan(
		&sID, &sName, &sType, &lat, &lng, &capacity, &sCreatedAt,
		&eID, &eFromID, &eToID, &eDistKM, &eCreatedAt,
	); err != nil {
		return nil, nil, err
	}

	station, err := parseStation(sID, sName, sType, lat, lng, capacity, sCreatedAt)
	if err != nil {
		return nil, nil, err
	}

	edge, err := parseEdge(eID, eFromID, eToID, eDistKM, eCreatedAt)
	if err != nil {
		return nil, nil, err
	}

	return station, edge, nil
}

func parseStation(id, name, stationType string, lat, lng float64, capacity int, createdAt time.Time) (*entity.Station, error) {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}
	return &entity.Station{
		ID:        parsedID,
		Name:      name,
		Type:      entity.Type(stationType),
		Location:  entity.Location{Latitude: lat, Longitude: lng},
		Capacity:  capacity,
		CreatedAt: createdAt,
	}, nil
}

func parseEdge(id, fromID, toID sql.NullString, distKM sql.NullInt64, createdAt sql.NullTime) (*entity.Edge, error) {
	if !id.Valid {
		return nil, nil
	}

	parsedID, err := uuid.Parse(id.String)
	if err != nil {
		return nil, err
	}
	parsedFrom, err := uuid.Parse(fromID.String)
	if err != nil {
		return nil, err
	}
	parsedTo, err := uuid.Parse(toID.String)
	if err != nil {
		return nil, err
	}

	return &entity.Edge{
		ID:            parsedID,
		FromStationID: parsedFrom,
		ToStationID:   parsedTo,
		DistanceKM:    float64(distKM.Int64),
		IsActive:      true,
		CreatedAt:     createdAt.Time,
	}, nil
}
