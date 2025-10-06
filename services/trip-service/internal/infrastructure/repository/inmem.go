package repository

import (
	"context"
	"fmt"
	"sync"
	"ride-sharing/services/trip-service/internal/domain"
	pb "ride-sharing/shared/proto/trip"
)

type inmemRepository struct {
	mu        sync.RWMutex
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
	r.mu.Lock()
	defer r.mu.Unlock()
	
	r.trips[trip.ID.Hex()] = trip
	return trip, nil
}

func (r *inmemRepository) SaveRideFare(ctx context.Context, f *domain.RideFareModel) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	r.rideFares[f.ID.Hex()] = f
	return nil
}

func (r *inmemRepository) GetRideFareByID(ctx context.Context, id string) (*domain.RideFareModel, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	res, ok := r.rideFares[id]
	if !ok {
		return nil, fmt.Errorf("failed to get Fare by ID")
	}
	return res, nil
}

func (r *inmemRepository) GetTripByID(ctx context.Context, id string) (*domain.TripModel, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	res, ok := r.trips[id]
	if !ok {
		return nil, fmt.Errorf("failed to get Trip by ID")
	}
	return res, nil
}

func (r *inmemRepository) UpdateTripStatus(ctx context.Context, tripID string, status string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	trip, ok := r.trips[tripID]
	if !ok {
		return fmt.Errorf("trip not found")
	}
	
	trip.Status = status
	return nil
}

func (r *inmemRepository) AssignDriver(ctx context.Context, tripID string, driver *pb.TripDriver) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	trip, ok := r.trips[tripID]
	if !ok {
		return fmt.Errorf("trip not found")
	}
	
	trip.Driver = driver
	return nil
}
