package delivery

import "net/http"

type CityHTTPHandler interface {
	GetCityByCoordinates(w http.ResponseWriter, r *http.Request)
}

type CoordinatesRequest struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}
