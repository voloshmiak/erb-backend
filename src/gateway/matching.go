package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"erb-backend/src/entity"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"time"

	"github.com/google/uuid"
)

const (
	matchEndpoint = "/api/match"
)

type MatchingGateway struct {
	URL    string
	Client *http.Client
}

func NewMatchingGateway(url string) *MatchingGateway {
	return &MatchingGateway{
		URL:    url,
		Client: http.DefaultClient,
	}
}

type orderDTO struct {
	OrderID     string `json:"order_id"`
	StationToID string `json:"station_to_id"`
	WagonType   string `json:"wagon_type"`
	Quantity    int    `json:"quantity"`
	DesiredDate string `json:"desired_date"`
}

type wagonDTO struct {
	WagonID          string  `json:"wagon_id"`
	WagonNumber      string  `json:"wagon_number"`
	WagonType        string  `json:"wagon_type"`
	CurrentStationID string  `json:"current_station_id"`
	IdleDays         float64 `json:"idle_days"`
}

type stationDTO struct {
	StationID string  `json:"station_id"`
	Name      string  `json:"name"`
	Type      string  `json:"type"`
	Lat       float64 `json:"lat"`
	Lng       float64 `json:"lng"`
}

type edgeDTO struct {
	FromStationID string  `json:"from_station_id"`
	ToStationID   string  `json:"to_station_id"`
	DistanceKM    float64 `json:"distance_km"`
}

type locomotiveDTO struct {
	LocomotiveID     string `json:"locomotive_id"`
	CurrentStationID string `json:"current_station_id"`
}

type matchRequest struct {
	Orders      []orderDTO      `json:"orders"`
	Wagons      []wagonDTO      `json:"wagons"`
	Stations    []stationDTO    `json:"stations"`
	Edges       []edgeDTO       `json:"edges"`
	Locomotives []locomotiveDTO `json:"locomotives,omitempty"`
}

type matchedAssignment struct {
	OrderID        string   `json:"order_id"`
	WagonID        string   `json:"wagon_id"`
	WagonNumber    string   `json:"wagon_number"`
	Route          []string `json:"route"`
	EmptyRunKM     float64  `json:"empty_run_km"`
	CostEmptyRun   float64  `json:"cost_empty_run"`
	EstimatedHours float64  `json:"estimated_hours"`
}

type unmatchedOrder struct {
	OrderID string `json:"order_id"`
	Reason  string `json:"reason"`
}

type trainGroupDTO struct {
	TrainID      string   `json:"train_id"`
	WagonIDs     []string `json:"wagon_ids"`
	Source       string   `json:"source"`
	Dest         string   `json:"dest"`
	LocomotiveID *string  `json:"loco_id"`
	RepositionKM float64  `json:"reposition_km"`
	DistanceKM   float64  `json:"distance_km"`
}

type matchResponse struct {
	Assignments     []matchedAssignment    `json:"assignments"`
	UnmatchedOrders []unmatchedOrder       `json:"unmatched_orders"`
	Metrics         entity.MatchingMetrics `json:"metrics"`
	TrainGroups     []trainGroupDTO        `json:"train_groups,omitempty"`
}

// Conversion helpers

func toOrderDTOs(orders []*entity.Order) []orderDTO {
	out := make([]orderDTO, 0, len(orders))
	for _, o := range orders {
		out = append(out, orderDTO{
			OrderID:     o.ID.String(),
			StationToID: o.StationToID.String(),
			WagonType:   string(o.WagonType),
			Quantity:    o.Quantity,
			DesiredDate: o.DesiredDate.Format(time.DateOnly),
		})
	}
	return out
}

func toWagonDTOs(wagons []*entity.Wagon) []wagonDTO {
	out := make([]wagonDTO, 0, len(wagons))
	for _, w := range wagons {
		idleDays := time.Since(w.LastUnloadTime).Hours() / 24
		out = append(out, wagonDTO{
			WagonID:          w.ID.String(),
			WagonNumber:      w.Number,
			WagonType:        string(w.Type),
			CurrentStationID: w.CurrentStationID.String(),
			IdleDays:         idleDays,
		})
	}
	return out
}

func toStationDTOs(stations []*entity.Station) []stationDTO {
	out := make([]stationDTO, 0, len(stations))
	for _, s := range stations {
		out = append(out, stationDTO{
			StationID: s.ID.String(),
			Name:      s.Name,
			Type:      string(s.Type),
			Lat:       s.Location.Latitude,
			Lng:       s.Location.Longitude,
		})
	}
	return out
}

func toEdgeDTOs(edges []*entity.Edge) []edgeDTO {
	out := make([]edgeDTO, 0, len(edges))
	for _, e := range edges {
		if !e.IsActive {
			continue
		}
		out = append(out, edgeDTO{
			FromStationID: e.FromStationID.String(),
			ToStationID:   e.ToStationID.String(),
			DistanceKM:    e.DistanceKM,
		})
	}
	return out
}

func toLocomotiveDTOs(locomotives []*entity.Locomotive) []locomotiveDTO {
	out := make([]locomotiveDTO, 0, len(locomotives))
	for _, l := range locomotives {
		out = append(out, locomotiveDTO{
			LocomotiveID:     l.ID.String(),
			CurrentStationID: l.CurrentStationID.String(),
		})
	}
	return out
}

func (g *MatchingGateway) Match(ctx context.Context, orders []*entity.Order,
	wagons []*entity.Wagon, stations []*entity.Station, edges []*entity.Edge,
	locomotives []*entity.Locomotive) ([]*entity.AssignmentResult, []entity.TrainGroupResult, *entity.MatchingMetrics, error) {

	reqBody := matchRequest{
		Orders:      toOrderDTOs(orders),
		Wagons:      toWagonDTOs(wagons),
		Stations:    toStationDTOs(stations),
		Edges:       toEdgeDTOs(edges),
		Locomotives: toLocomotiveDTOs(locomotives),
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, g.URL+matchEndpoint,
		bytes.NewReader(body))
	if err != nil {
		return nil, nil, nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.Client.Do(req)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("do request: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {
			log.Println("failed to close response body: ", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, nil, nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	var result matchResponse
	if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, nil, nil, fmt.Errorf("decode response: %w", err)
	}

	for _, u := range result.UnmatchedOrders {
		log.Printf("matching: unmatched order %s: %s", u.OrderID, u.Reason)
	}

	m := result.Metrics
	log.Printf("matching metrics: matched=%d unmatched=%d match_rate=%.2f total_empty_km=%.1f total_cost=%.1f cost_saved=%.1f",
		m.OrdersMatched, m.OrdersUnmatched, m.MatchRate, m.TotalEmptyKM, m.TotalCost, m.CostSaved)

	out := make([]*entity.AssignmentResult, 0, len(result.Assignments))
	for _, a := range result.Assignments {
		orderID, err := uuid.Parse(a.OrderID)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("parse order_id %q: %w", a.OrderID, err)
		}
		wagonID, err := uuid.Parse(a.WagonID)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("parse wagon_id %q: %w", a.WagonID, err)
		}

		route := make([]uuid.UUID, 0, len(a.Route))
		for _, s := range a.Route {
			stationID, err := uuid.Parse(s)
			if err != nil {
				return nil, nil, nil, fmt.Errorf("parse station_id %q: %w", s, err)
			}
			route = append(route, stationID)
		}

		out = append(out, &entity.AssignmentResult{
			Assignment: &entity.Assignment{
				OrderID:          orderID,
				WagonID:          wagonID,
				EmptyRunKM:       int(math.Round(a.EmptyRunKM)),
				CostEmptyRun:     int(math.Round(a.CostEmptyRun)),
				EstimatedArrival: time.Now().Add(time.Duration(a.EstimatedHours * float64(time.Hour))),
			},
			Route: route,
		})
	}

	trainGroups := make([]entity.TrainGroupResult, 0, len(result.TrainGroups))
	for _, tg := range result.TrainGroups {
		trainID, err := uuid.Parse(tg.TrainID)
		if err != nil {
			log.Printf("match: failed to parse train_id %q: %v", tg.TrainID, err)
			continue
		}
		sourceID, err := uuid.Parse(tg.Source)
		if err != nil {
			log.Printf("match: failed to parse source station_id %q: %v", tg.Source, err)
			continue
		}
		destID, err := uuid.Parse(tg.Dest)
		if err != nil {
			log.Printf("match: failed to parse dest station_id %q: %v", tg.Dest, err)
			continue
		}

		var locoID *uuid.UUID
		if tg.LocomotiveID != nil {
			lid, err := uuid.Parse(*tg.LocomotiveID)
			if err != nil {
				log.Printf("match: failed to parse loco_id %q: %v", *tg.LocomotiveID, err)
				continue
			}
			locoID = &lid
		}

		wagonIDs := make([]uuid.UUID, 0, len(tg.WagonIDs))
		for _, widStr := range tg.WagonIDs {
			wid, err := uuid.Parse(widStr)
			if err != nil {
				log.Printf("match: failed to parse wagon_id %q in train %s: %v", widStr, trainID, err)
				continue
			}
			wagonIDs = append(wagonIDs, wid)
		}

		trainGroups = append(trainGroups, entity.TrainGroupResult{
			TrainID:              trainID,
			SourceStationID:      sourceID,
			DestinationStationID: destID,
			LocomotiveID:         locoID,
			WagonIDs:             wagonIDs,
			RepositionKM:         tg.RepositionKM,
			DistanceKM:           tg.DistanceKM,
		})
	}

	return out, trainGroups, &result.Metrics, nil
}
