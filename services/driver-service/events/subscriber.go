package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/mmcloughlin/geohash"
	"ride-sharing/shared/events"
	"ride-sharing/shared/contracts"
	pb "ride-sharing/shared/proto/trip"
)

// DriverEventSubscriber Driver服务事件订阅器
type DriverEventSubscriber struct {
	subscriber events.Subscriber
	service    *Service
	publisher  *DriverEventPublisher
}

// NewDriverEventSubscriber 创建Driver事件订阅器
func NewDriverEventSubscriber(subscriber events.Subscriber, service *Service, publisher *DriverEventPublisher) *DriverEventSubscriber {
	return &DriverEventSubscriber{
		subscriber: subscriber,
		service:    service,
		publisher:  publisher,
	}
}

// SubscribeToTripEvents 订阅行程事件
func (s *DriverEventSubscriber) SubscribeToTripEvents(ctx context.Context) error {
	// 订阅行程创建事件
	err := s.subscriber.Subscribe(
		"find_available_drivers_queue",
		contracts.TripEventCreated,
		s.handleTripCreated,
	)
	if err != nil {
		return fmt.Errorf("订阅行程创建事件失败: %w", err)
	}

	log.Println("成功订阅行程事件")
	return nil
}

// handleTripCreated 处理行程创建事件
func (s *DriverEventSubscriber) handleTripCreated(data []byte) error {
	var trip pb.Trip
	if err := json.Unmarshal(data, &trip); err != nil {
		return fmt.Errorf("解析行程创建事件失败: %w", err)
	}

	// 获取行程起点坐标
	if trip.Route == nil || len(trip.Route.Geometry) == 0 || len(trip.Route.Geometry[0].Coordinates) == 0 {
		return fmt.Errorf("行程路线信息不完整")
	}

	startLocation := trip.Route.Geometry[0].Coordinates[0]
	pickupGeohash := geohash.Encode(startLocation.Latitude, startLocation.Longitude)

	// 查找附近的司机
	nearbyDrivers := s.service.FindNearbyDrivers(pickupGeohash, trip.SelectedFare.PackageSlug)
	
	if len(nearbyDrivers) == 0 {
		log.Printf("未找到可用司机，行程ID: %s", trip.Id)
		return nil
	}

	log.Printf("找到 %d 个可用司机，行程ID: %s", len(nearbyDrivers), trip.Id)

	// 向找到的司机发送行程请求
	for _, driver := range nearbyDrivers {
		tripRequest := DriverTripRequest{
			TripID:     trip.Id,
			DriverID:   driver.Driver.Id,
			RiderID:    trip.UserId,
			Pickup:     &Coordinate{Latitude: startLocation.Latitude, Longitude: startLocation.Longitude},
			Fare:       trip.SelectedFare.TotalPriceInCents,
			Package:    trip.SelectedFare.PackageSlug,
		}

		// 发布司机行程请求命令
		if err := s.publishDriverTripRequest(tripRequest); err != nil {
			log.Printf("发布司机行程请求失败: %v", err)
			continue
		}

		log.Printf("已向司机发送行程请求，司机ID: %s, 行程ID: %s", driver.Driver.Id, trip.Id)
	}

	return nil
}

// publishDriverTripRequest 发布司机行程请求命令
func (s *DriverEventSubscriber) publishDriverTripRequest(request DriverTripRequest) error {
	if s.publisher == nil {
		log.Printf("事件发布器未初始化，无法发布司机行程请求命令: %+v", request)
		return nil
	}
	
	return s.publisher.PublishDriverTripRequest(context.Background(), request)
}

// DriverTripRequest 司机行程请求
type DriverTripRequest struct {
	TripID   string      `json:"tripID"`
	DriverID string      `json:"driverID"`
	RiderID  string      `json:"riderID"`
	Pickup   *Coordinate `json:"pickup"`
	Fare     float64     `json:"fare"`
	Package  string      `json:"package"`
}

// Coordinate 坐标
type Coordinate struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// Close 关闭订阅器
func (s *DriverEventSubscriber) Close() error {
	return s.subscriber.Close()
}