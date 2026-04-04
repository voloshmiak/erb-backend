package controller

import (
	"encoding/json"
	"erb-backend/src/entity"
	"erb-backend/src/usecase"
	"net/http"
)

type ListStationsController struct {
	usecase *usecase.ListStationsUseCase
}

func NewListStationsController(usecase *usecase.ListStationsUseCase) *ListStationsController {
	return &ListStationsController{usecase: usecase}
}

type stationResponse struct {
	StationID string  `json:"stationId"`
	Name      string  `json:"name"`
	Type      string  `json:"type"`
	Lat       float64 `json:"lat"`
	Lng       float64 `json:"lng"`
}

type edgeResponse struct {
	From       string `json:"from"`
	To         string `json:"to"`
	DistanceKM int    `json:"distanceKm"`
}

type listStationsResponse struct {
	Stations []stationResponse `json:"stations"`
	Edges    []edgeResponse    `json:"edges"`
}

// ServeHTTP godoc
// @Summary     List Stations
// @Description Retrieves all stations and active edges forming the railway graph
// @Tags        Station
// @Accept      json
// @Produce     json
// @Success     200 {object} listStationsResponse
// @Failure     500 {object} string "Failed to retrieve stations"
// @Router      /stations [get]
func (c *ListStationsController) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	stations, edges, err := c.usecase.Execute(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := listStationsResponse{
		Stations: mapStations(stations),
		Edges:    mapEdges(edges),
	}

	w.Header().Set("Content-Type", "application/json")

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func mapStations(stations []*entity.Station) []stationResponse {
	result := make([]stationResponse, len(stations))
	for i, s := range stations {
		result[i] = stationResponse{
			StationID: s.ID.String(),
			Name:      s.Name,
			Type:      string(s.Type),
			Lat:       s.Location.Latitude,
			Lng:       s.Location.Longitude,
		}
	}
	return result
}

func mapEdges(edges []*entity.Edge) []edgeResponse {
	result := make([]edgeResponse, len(edges))
	for i, e := range edges {
		result[i] = edgeResponse{
			From:       e.FromStationID.String(),
			To:         e.ToStationID.String(),
			DistanceKM: int(e.DistanceKM),
		}
	}
	return result
}
