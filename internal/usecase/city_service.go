package usecase

import (
	"context"

	"github.com/opusdvs/DonWeather-ms-georesolve/internal/domain"
)

type CityService struct {
	repository domain.CityRepository
}

func NewCityService(repository domain.CityRepository) *CityService {
	return &CityService{repository: repository}
}

func (s *CityService) GetCityByCoordinates(ctx context.Context, coordinates domain.Coordinates) (domain.City, error) {
	return s.repository.GetCityByCoordinates(ctx, coordinates)
}
