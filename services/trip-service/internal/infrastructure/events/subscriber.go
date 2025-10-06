package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"ride-sharing/services/trip-service/internal/domain"
	"ride-sharing/shared/events"
	"ride-sharing/shared/contracts"
)

// DriverTripResponse 司机行程响应
type DriverTripResponse struct {
	TripID   string `json:"tripID"`
	RiderID  string `json:"riderID"`
	DriverID string `json:"driverID"`
	Accept   bool   `json:"accept"`
}

// PaymentEvent 支付事件
type PaymentEvent struct {
	TripID string `json:"tripID"`
	Status string `json:"status"`
}

// TripEventSubscriber Trip服务事件订阅器
type TripEventSubscriber struct {
	subscriber events.Subscriber
	service    domain.TripService
	repo       domain.TripRepository
}

// NewTripEventSubscriber 创建Trip事件订阅器
func NewTripEventSubscriber(subscriber events.Subscriber, service domain.TripService, repo domain.TripRepository) *TripEventSubscriber {
	return &TripEventSubscriber{
		subscriber: subscriber,
		service:    service,
		repo:       repo,
	}
}

// SubscribeToDriverResponses 订阅司机响应
func (s *TripEventSubscriber) SubscribeToDriverResponses(ctx context.Context) error {
	// 订阅司机接受行程的响应
	err := s.subscriber.Subscribe(
		"driver_trip_response_queue",
		contracts.DriverCmdTripAccept,
		s.handleDriverAcceptTrip,
	)
	if err != nil {
		return fmt.Errorf("订阅司机接受行程事件失败: %w", err)
	}

	// 订阅司机拒绝行程的响应
	err = s.subscriber.Subscribe(
		"driver_trip_decline_queue",
		contracts.DriverCmdTripDecline,
		s.handleDriverDeclineTrip,
	)
	if err != nil {
		return fmt.Errorf("订阅司机拒绝行程事件失败: %w", err)
	}

	log.Println("成功订阅司机响应事件")
	return nil
}

// SubscribeToPaymentEvents 订阅支付事件
func (s *TripEventSubscriber) SubscribeToPaymentEvents(ctx context.Context) error {
	// 订阅支付成功事件
	err := s.subscriber.Subscribe(
		"payment_success_queue",
		contracts.PaymentEventSuccess,
		s.handlePaymentSuccess,
	)
	if err != nil {
		return fmt.Errorf("订阅支付成功事件失败: %w", err)
	}

	// 订阅支付失败事件
	err = s.subscriber.Subscribe(
		"payment_failed_queue",
		contracts.PaymentEventFailed,
		s.handlePaymentFailed,
	)
	if err != nil {
		return fmt.Errorf("订阅支付失败事件失败: %w", err)
	}

	log.Println("成功订阅支付事件")
	return nil
}

// handleDriverAcceptTrip 处理司机接受行程事件
func (s *TripEventSubscriber) handleDriverAcceptTrip(data []byte) error {
	var response DriverTripResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return fmt.Errorf("解析司机接受行程响应失败: %w", err)
	}

	// 处理司机接受行程的业务逻辑
	if err := s.service.AcceptTrip(context.Background(), response.TripID, response.DriverID); err != nil {
		return fmt.Errorf("处理司机接受行程失败: %w", err)
	}

	log.Printf("成功处理司机接受行程事件: 行程ID=%s, 司机ID=%s", response.TripID, response.DriverID)
	return nil
}

// handleDriverDeclineTrip 处理司机拒绝行程事件
func (s *TripEventSubscriber) handleDriverDeclineTrip(data []byte) error {
	var response DriverTripResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return fmt.Errorf("解析司机拒绝行程响应失败: %w", err)
	}

	// 处理司机拒绝行程的业务逻辑
	if err := s.service.DeclineTrip(context.Background(), response.TripID, response.DriverID); err != nil {
		return fmt.Errorf("处理司机拒绝行程失败: %w", err)
	}

	log.Printf("成功处理司机拒绝行程事件: 行程ID=%s, 司机ID=%s", response.TripID, response.DriverID)
	return nil
}

// handlePaymentSuccess 处理支付成功事件
func (s *TripEventSubscriber) handlePaymentSuccess(data []byte) error {
	var paymentEvent PaymentEvent
	if err := json.Unmarshal(data, &paymentEvent); err != nil {
		return fmt.Errorf("解析支付成功事件失败: %w", err)
	}

	// 处理支付成功的业务逻辑
	if err := s.service.UpdatePaymentStatus(context.Background(), paymentEvent.TripID, "paid"); err != nil {
		return fmt.Errorf("处理支付成功失败: %w", err)
	}

	log.Printf("成功处理支付成功事件: 行程ID=%s", paymentEvent.TripID)
	return nil
}

// handlePaymentFailed 处理支付失败事件
func (s *TripEventSubscriber) handlePaymentFailed(data []byte) error {
	var paymentEvent PaymentEvent
	if err := json.Unmarshal(data, &paymentEvent); err != nil {
		return fmt.Errorf("解析支付失败事件失败: %w", err)
	}

	// 处理支付失败的业务逻辑
	if err := s.service.UpdatePaymentStatus(context.Background(), paymentEvent.TripID, "payment_failed"); err != nil {
		return fmt.Errorf("处理支付失败失败: %w", err)
	}

	log.Printf("成功处理支付失败事件: 行程ID=%s", paymentEvent.TripID)
	return nil
}

// Close 关闭订阅器
func (s *TripEventSubscriber) Close() error {
	return s.subscriber.Close()
}