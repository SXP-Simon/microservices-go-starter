package repository

import (
	"context"
	"fmt"
	"ride-sharing/services/trip-service/internal/domain"
)

// TODO: add mutexes to avoid concurrency problems
type inmemRepository struct {
	trips     map[string]*domain.TripModel
	rideFares map[string]*domain.RideFareModel
}

func NewInmemRepository() *inmemRepository {
	return &inmemRepository{
		trips:     make(map[string]*domain.TripModel),
		rideFares: make(map[string]*domain.RideFareModel),
	}
}

func (r *inmemRepository) CreateTrip(ctx context.Context, trip *domain.TripModel) (*domain.TripModel, error) {

	r.trips[trip.ID.Hex()] = trip
	return trip, nil
}
func (r *inmemRepository) SaveRideFare(ctx context.Context, f *domain.RideFareModel) error {
	r.rideFares[f.ID.Hex()] = f
	return nil
}

func (r *inmemRepository) GetRideFareByID(ctx context.Context, id string) (*domain.RideFareModel, error) {
	res, ok := r.rideFares[id]
	if ok == false {
		return nil, fmt.Errorf("failed to get Fare by ID")
	}
	return res, nil
}
