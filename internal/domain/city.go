package domain

import "context"

type Coordinates struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type City struct {
	Name        string      `json:"name"`
	Coordinates Coordinates `json:"coordinates"`
}

type CityRepository interface {
	GetCityByCoordinates(ctx context.Context, coordinates Coordinates) (City, error)
}
