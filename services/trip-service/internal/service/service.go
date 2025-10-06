package service

import (
	"context"
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"io"
	"net/http"
	"ride-sharing/services/trip-service/internal/domain"
	tripTypes "ride-sharing/services/trip-service/pkg/types"
	"ride-sharing/shared/proto/trip"
	"ride-sharing/shared/types"
)

type service struct {
	repo     domain.TripRepository
	publisher domain.TripEventPublisher
}

func NewService(repo domain.TripRepository, publisher domain.TripEventPublisher) *service {
	return &service{
		repo:     repo,
		publisher: publisher,
	}
}
func (s *service) CreateTrip(ctx context.Context, fare *domain.RideFareModel) (*domain.TripModel, error) {
	t := &domain.TripModel{
		ID:       primitive.NewObjectID(),
		UserID:   fare.UserID,
		Status:   "pending",
		RideFare: fare,
		Driver:   &trip.TripDriver{},
	}
	
	// 保存行程到数据库
	trip, err := s.repo.CreateTrip(ctx, t)
	if err != nil {
		return nil, err
	}
	
	// 发布行程创建事件
	if s.publisher != nil {
		if err := s.publisher.PublishTripCreated(ctx, trip); err != nil {
			// 记录错误但不中断流程
			fmt.Printf("发布行程创建事件失败: %v\n", err)
		}
	}
	
	return trip, nil
}

func (s *service) GetRoute(ctx context.Context, pickup, destination *types.Coordinate) (*tripTypes.OsrmApiResponse, error) {
	url := fmt.Sprintf(
		"http://router.project-osrm.org/route/v1/driving/%f,%f;%f,%f?overview=full&geometries=geojson",
		pickup.Longitude,
		pickup.Latitude,
		destination.Longitude,
		destination.Latitude,
	)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch route from OSRM API: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read the response: %v", err)
	}

	var routeResp tripTypes.OsrmApiResponse
	if err := json.Unmarshal(body, &routeResp); err != nil {
		return nil, fmt.Errorf("failed to parse respnse: %v", err)
	}
	return &routeResp, nil

}

func (s *service) EstimatePackagesPriceWithRoute(route *tripTypes.OsrmApiResponse) []*domain.RideFareModel {
	baseFares := getBaseFares()
	estimatedFares := make([]*domain.RideFareModel, len(baseFares))

	for i, f := range baseFares {
		estimatedFares[i] = estimateFareRoute(f, route)
	}

	return estimatedFares
}

func (s *service) GenerateTripFares(ctx context.Context, rideFares []*domain.RideFareModel, userID string, route *tripTypes.OsrmApiResponse) ([]*domain.RideFareModel, error) {
	fares := make([]*domain.RideFareModel, len(rideFares))

	for i, f := range rideFares {
		id := primitive.NewObjectID()

		fare := &domain.RideFareModel{
			UserID:            userID,
			ID:                id,
			TotalPriceInCents: f.TotalPriceInCents,
			PackageSlug:       f.PackageSlug,
			Route:             route,
		}

		if err := s.repo.SaveRideFare(ctx, fare); err != nil {
			return nil, fmt.Errorf("failed to save trip fare: %w", err)
		}

		fares[i] = fare
	}

	return fares, nil
}

func (s *service) GetAndValidateFare(ctx context.Context, fareID, userID string) (*domain.RideFareModel, error) {
	fare, err := s.repo.GetRideFareByID(ctx, fareID)
	if err != nil {
		return nil, fmt.Errorf("failed to get trip fare: %w", err)
	}

	if fare == nil {
		return nil, fmt.Errorf("fare does not exist")
	}
	if userID != fare.UserID {
		return nil, fmt.Errorf("fare does not belog to the user")
	}

	return fare, nil
}

func estimateFareRoute(f *domain.RideFareModel, route *tripTypes.OsrmApiResponse) *domain.RideFareModel {
	pricingCfg := tripTypes.DefaultPricingConfig()
	carPackagePrice := f.TotalPriceInCents

	distanceKm := route.Routes[0].Distance
	durationInMinutes := route.Routes[0].Duration

	distanceFare := distanceKm * pricingCfg.PricePerUnitOfDistance
	timeFare := durationInMinutes * pricingCfg.PricingPerMinute
	totalPrice := carPackagePrice + distanceFare + timeFare

	return &domain.RideFareModel{
		TotalPriceInCents: totalPrice,
		PackageSlug:       f.PackageSlug,
	}
}

// AcceptTrip 司机接受行程
func (s *service) AcceptTrip(ctx context.Context, tripID, driverID string) error {
	// 获取行程信息
	trip, err := s.repo.GetTripByID(ctx, tripID)
	if err != nil {
		return fmt.Errorf("获取行程信息失败: %w", err)
	}
	
	if trip == nil {
		return fmt.Errorf("行程不存在")
	}
	
	// 创建司机信息
	driver := &trip.TripDriver{
		Id: driverID,
	}
	
	// 分配司机并更新行程状态
	if err := s.repo.AssignDriver(ctx, tripID, driver); err != nil {
		return fmt.Errorf("分配司机失败: %w", err)
	}
	
	if err := s.repo.UpdateTripStatus(ctx, tripID, "driver_assigned"); err != nil {
		return fmt.Errorf("更新行程状态失败: %w", err)
	}
	
	// 发布司机分配事件
	if s.publisher != nil {
		// 更新行程信息
		trip.Driver = driver
		trip.Status = "driver_assigned"
		
		if err := s.publisher.PublishDriverAssigned(ctx, trip); err != nil {
			// 记录错误但不中断流程
			fmt.Printf("发布司机分配事件失败: %v\n", err)
		}
	}
	
	return nil
}

// DeclineTrip 司机拒绝行程
func (s *service) DeclineTrip(ctx context.Context, tripID, driverID string) error {
	// 发布司机不感兴趣事件
	if s.publisher != nil {
		if err := s.publisher.PublishDriverNotInterested(ctx, tripID, driverID); err != nil {
			// 记录错误但不中断流程
			fmt.Printf("发布司机不感兴趣事件失败: %v\n", err)
		}
	}
	
	return nil
}

// UpdatePaymentStatus 更新支付状态
func (s *service) UpdatePaymentStatus(ctx context.Context, tripID, status string) error {
	// 更新行程状态
	if err := s.repo.UpdateTripStatus(ctx, tripID, status); err != nil {
		return fmt.Errorf("更新支付状态失败: %w", err)
	}
	
	return nil
}

func getBaseFares() []*domain.RideFareModel {
	return []*domain.RideFareModel{
		{
			PackageSlug:       "suv",
			TotalPriceInCents: 200,
		},
		{
			PackageSlug:       "sedan",
			TotalPriceInCents: 350,
		},
		{
			PackageSlug:       "van",
			TotalPriceInCents: 400,
		},
		{
			PackageSlug:       "luxury",
			TotalPriceInCents: 1000,
		},
	}
}
