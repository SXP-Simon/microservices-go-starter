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
	GetTripByID(ctx context.Context, id string) (*TripModel, error)
	UpdateTripStatus(ctx context.Context, tripID string, status string) error
	AssignDriver(ctx context.Context, tripID string, driver *pb.TripDriver) error
}

type TripService interface {
	CreateTrip(ctx context.Context, fare *RideFareModel) (*TripModel, error)
	GetRoute(ctx context.Context, pickup, destination *types.Coordinate) (*tripTypes.OsrmApiResponse, error)
	EstimatePackagesPriceWithRoute(route *tripTypes.OsrmApiResponse) []*RideFareModel
	GenerateTripFares(ctx context.Context, fares []*RideFareModel, userID string, route *tripTypes.OsrmApiResponse) ([]*RideFareModel, error)
	GetAndValidateFare(ctx context.Context, fareID, userID string) (*RideFareModel, error)
	AcceptTrip(ctx context.Context, tripID, driverID string) error
	DeclineTrip(ctx context.Context, tripID, driverID string) error
	UpdatePaymentStatus(ctx context.Context, tripID, status string) error
}

// TripEventPublisher 行程事件发布器接口
type TripEventPublisher interface {
	PublishTripCreated(ctx context.Context, trip *TripModel) error
	PublishDriverAssigned(ctx context.Context, trip *TripModel) error
	PublishNoDriversFound(ctx context.Context, tripID string) error
	PublishDriverNotInterested(ctx context.Context, tripID, driverID string) error
	Close() error
}

// ToProto 将TripModel转换为protobuf格式
func (t *TripModel) ToProto() *pb.Trip {
	// 转换行程费用
	var rideFare *pb.RideFare
	if t.RideFare != nil {
		rideFare = t.RideFare.ToProto()
	}
	
	// 转换路线
	var route *pb.Route
	if t.RideFare != nil && t.RideFare.Route != nil {
		route = t.RideFare.Route.ToProto()
	}
	
	return &pb.Trip{
		Id:           t.ID.Hex(),
		UserID:       t.UserID,
		Status:       t.Status,
		SelectedFare: rideFare,
		Route:        route,
		Driver:       t.Driver,
	}
}
