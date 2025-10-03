package main

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	pb "ride-sharing/shared/proto/driver"
)

type grpcHandler struct {
	pb.UnimplementedDriverServiceServer
	service *Service
}

func NewGrpcHandler(s *grpc.Server, service *Service) {
	handler := &grpcHandler{
		service: service,
	}
	pb.RegisterDriverServiceServer(s, handler)

}

func (h *grpcHandler) RegisterDriver(ctx context.Context, req *pb.RegisterDriverRequest) (*pb.RegisterDriverResponse, error) {
	driver, err := h.service.RegisterDriver(req.GetDriverID(), req.GetPackageSlug())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to register driver")
	}

	return &pb.RegisterDriverResponse{
		Driver: driver,
	}, nil
}
func (h *grpcHandler) UnRegisterDriver(ctx context.Context, req *pb.RegisterDriverRequest) (*pb.RegisterDriverResponse, error) {
	h.service.UnregisterDriver(req.GetDriverID())

	return &pb.RegisterDriverResponse{
		Driver: &pb.Driver{
			Id: req.GetDriverID(),
		},
	}, nil

}
