package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"ride-sharing/services/trip-service/internal/domain"
	pb "ride-sharing/shared/proto/trip"
	"ride-sharing/shared/events"
	"ride-sharing/shared/contracts"
)

// TripEventPublisher Trip服务事件发布器
type TripEventPublisher struct {
	publisher events.Publisher
}

// NewTripEventPublisher 创建Trip事件发布器
func NewTripEventPublisher(publisher events.Publisher) *TripEventPublisher {
	return &TripEventPublisher{
		publisher: publisher,
	}
}

// PublishTripCreated 发布行程创建事件
func (p *TripEventPublisher) PublishTripCreated(ctx context.Context, trip *domain.TripModel) error {
	// 转换为protobuf格式
	tripProto := p.convertTripToProto(trip)
	
	// 发布事件
	err := p.publisher.PublishEvent(contracts.TripEventCreated, tripProto)
	if err != nil {
		return fmt.Errorf("发布行程创建事件失败: %w", err)
	}
	
	log.Printf("成功发布行程创建事件: %s", trip.ID.Hex())
	return nil
}

// PublishDriverAssigned 发布司机分配事件
func (p *TripEventPublisher) PublishDriverAssigned(ctx context.Context, trip *domain.TripModel) error {
	// 转换为protobuf格式
	tripProto := p.convertTripToProto(trip)
	
	// 发布事件
	err := p.publisher.PublishEvent(contracts.TripEventDriverAssigned, tripProto)
	if err != nil {
		return fmt.Errorf("发布司机分配事件失败: %w", err)
	}
	
	log.Printf("成功发布司机分配事件: %s", trip.ID.Hex())
	return nil
}

// PublishNoDriversFound 发布未找到司机事件
func (p *TripEventPublisher) PublishNoDriversFound(ctx context.Context, tripID string) error {
	// 创建事件数据
	eventData := map[string]string{
		"tripID": tripID,
	}
	
	// 发布事件
	err := p.publisher.PublishEvent(contracts.TripEventNoDriversFound, eventData)
	if err != nil {
		return fmt.Errorf("发布未找到司机事件失败: %w", err)
	}
	
	log.Printf("成功发布未找到司机事件: %s", tripID)
	return nil
}

// PublishDriverNotInterested 发布司机不感兴趣事件
func (p *TripEventPublisher) PublishDriverNotInterested(ctx context.Context, tripID, driverID string) error {
	// 创建事件数据
	eventData := map[string]string{
		"tripID":   tripID,
		"driverID": driverID,
	}
	
	// 发布事件
	err := p.publisher.PublishEvent(contracts.TripEventDriverNotInterested, eventData)
	if err != nil {
		return fmt.Errorf("发布司机不感兴趣事件失败: %w", err)
	}
	
	log.Printf("成功发布司机不感兴趣事件: 行程ID=%s, 司机ID=%s", tripID, driverID)
	return nil
}

// convertTripToProto 将TripModel转换为protobuf格式
func (p *TripEventPublisher) convertTripToProto(trip *domain.TripModel) *pb.Trip {
	// 转换行程费用
	var rideFare *pb.RideFare
	if trip.RideFare != nil {
		rideFare = trip.RideFare.ToProto()
	}
	
	// 转换路线
	var route *pb.Route
	if trip.RideFare != nil && trip.RideFare.Route != nil {
		route = trip.RideFare.Route.ToProto()
	}
	
	return &pb.Trip{
		Id:           trip.ID.Hex(),
		UserID:       trip.UserID,
		Status:       trip.Status,
		SelectedFare: rideFare,
		Route:        route,
		Driver:       trip.Driver,
	}
}

// Close 关闭发布器
func (p *TripEventPublisher) Close() error {
	return p.publisher.Close()
}