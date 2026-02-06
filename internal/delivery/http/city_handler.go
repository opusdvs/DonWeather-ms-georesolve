package delivery

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/opusdvs/DonWeather-ms-georesolve/internal/domain"
	"github.com/opusdvs/DonWeather-ms-georesolve/internal/usecase"
)

type CityHandler struct {
	service *usecase.CityService
}

func NewCityHandler(service *usecase.CityService) *CityHandler {
	return &CityHandler{service: service}
}

func (h *CityHandler) GetCityByCoordinates(w http.ResponseWriter, r *http.Request) {
	var coordinatesRequest CoordinatesRequest
	if err := json.NewDecoder(r.Body).Decode(&coordinatesRequest); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := ValidateCoordinates(coordinatesRequest); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	coordinates := domain.Coordinates{
		Latitude:  coordinatesRequest.Latitude,
		Longitude: coordinatesRequest.Longitude,
	}

	city, err := h.service.GetCityByCoordinates(r.Context(), coordinates)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(city)
	w.WriteHeader(http.StatusOK)
}

func ValidateCoordinates(coordinates CoordinatesRequest) error {
	if coordinates.Latitude == 0 || coordinates.Longitude == 0 {
		return errors.New("coordinates must not be 0")
	}
	if coordinates.Latitude < -90 || coordinates.Latitude > 90 {
		return errors.New("latitude must be between -90 and 90")
	}
	if coordinates.Longitude < -180 || coordinates.Longitude > 180 {
		return errors.New("longitude must be between -180 and 180")
	}
	return nil
}
