package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/opusdvs/DonWeather-ms-georesolve/internal/domain"
)

type PostgresRepoCity struct {
	db *sql.DB
}

func NewPostgresRepoCity(db *sql.DB) *PostgresRepoCity {
	return &PostgresRepoCity{db: db}
}

func (r *PostgresRepoCity) GetCityByCoordinates(ctx context.Context, coordinates domain.Coordinates) (domain.City, error) {
	var city domain.City
	// ST_Point ожидает (longitude, latitude), поэтому $2, $1
	query := `
		SELECT name
		FROM cities
		ORDER BY geom::geography <-> ST_SetSRID(ST_Point($2, $1), 4326)::geography
		LIMIT 1;
	`
	rows, err := r.db.QueryContext(ctx, query, coordinates.Latitude, coordinates.Longitude)
	if err != nil {
		return domain.City{}, err
	}
	defer rows.Close()

	if rows.Next() {
		if err := rows.Scan(&city.Name); err != nil {
			return domain.City{}, err
		}
		city.Coordinates = domain.Coordinates{
			Latitude:  coordinates.Latitude,
			Longitude: coordinates.Longitude,
		}
		return city, nil
	}

	return domain.City{}, errors.New("no city found")
}
