package grpc

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"ride-sharing/services/trip-service/internal/domain"
	pb "ride-sharing/shared/proto/trip"
	"ride-sharing/shared/types"
)

type gRPCHandler struct {
	pb.UnimplementedTripServiceServer
	service domain.TripService
}

func NewGRPCHandler(server *grpc.Server, service domain.TripService) *gRPCHandler {
	handler := &gRPCHandler{
		service: service,
	}

	pb.RegisterTripServiceServer(server, handler)

	return handler
}
func (h *gRPCHandler) CreateTrip(ctx context.Context, req *pb.CreteTripRequest) (*pb.CreateTripResponse, error) {

	fareID := req.GetRideFareID()
	userID := req.GetUserID()

	rideFare, err := h.service.GetAndValidateFare(ctx, fareID, userID)

	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get and validate fare: %v", err)
	}

	trip, err := h.service.CreateTrip(ctx, rideFare)

	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create trip: %v", err)
	}

	return &pb.CreateTripResponse{
		TripID: trip.ID.Hex(),
	}, nil
}

func (h *gRPCHandler) PreviewTrip(ctx context.Context, req *pb.PreviewTripRequest) (*pb.PreviewTripResponse, error) {
	pickup := req.GetStartLocation()
	destination := req.GetEndLocation()

	pickupCords := &types.Coordinate{
		Latitude:  pickup.Latitude,
		Longitude: pickup.Longitude,
	}
	destinationCords := &types.Coordinate{
		Latitude:  destination.Latitude,
		Longitude: destination.Longitude,
	}

	userID := req.GetUserID()

	route, err := h.service.GetRoute(ctx, pickupCords, destinationCords)
	if err != nil {
		log.Println("Some error ", err)
		return nil, status.Errorf(codes.Internal, "failed to get route %v", err)
	}

	// Estimate the ride
	estimatedFares := h.service.EstimatePackagesPriceWithRoute(route)

	fares, err := h.service.GenerateTripFares(ctx, estimatedFares, userID, route)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate ride fares %v", err)

	}

	return &pb.PreviewTripResponse{
		Route:     route.ToProto(),
		RideFares: domain.ToRideFaresProto(fares),
	}, nil
}
