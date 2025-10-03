package domain

import (
	"context"
	"go.mongodb.org/mongo-driver/bson/primitive"
	tripTypes "ride-sharing/services/trip-service/pkg/types"
	pb "ride-sharing/shared/proto/trip"
	"ride-sharing/shared/types"
)

type TripModel struct {
	ID       primitive.ObjectID // To avoid conflicts with mongo
	UserID   string
	Status   string
	RideFare *RideFareModel
	Driver   *pb.TripDriver
}

type TripRepository interface {
	CreateTrip(ctx context.Context, trip *TripModel) (*TripModel, error)
	SaveRideFare(ctx context.Context, f *RideFareModel) error

	GetRideFareByID(ctx context.Context, id string) (*RideFareModel, error)
}

type TripService interface {
	CreateTrip(ctx context.Context, fare *RideFareModel) (*TripModel, error)
	GetRoute(ctx context.Context, pickup, destination *types.Coordinate) (*tripTypes.OsrmApiResponse, error)
	EstimatePackagesPriceWithRoute(route *tripTypes.OsrmApiResponse) []*RideFareModel
	GenerateTripFares(ctx context.Context, fares []*RideFareModel, userID string, route *tripTypes.OsrmApiResponse) ([]*RideFareModel, error)

	GetAndValidateFare(ctx context.Context, fareID, userID string) (*RideFareModel, error)
}
