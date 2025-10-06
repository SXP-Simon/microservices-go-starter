package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"ride-sharing/payment-service/internal/service"
	"ride-sharing/shared/events"
	"ride-sharing/shared/contracts"
	pb "ride-sharing/shared/proto/trip"
)

// PaymentEventSubscriber 支付服务事件订阅器
type PaymentEventSubscriber struct {
	subscriber events.Subscriber
	service    service.PaymentService
}

// NewPaymentEventSubscriber 创建支付事件订阅器
func NewPaymentEventSubscriber(subscriber events.Subscriber, service service.PaymentService) *PaymentEventSubscriber {
	return &PaymentEventSubscriber{
		subscriber: subscriber,
		service:    service,
	}
}

// SubscribeToTripEvents 订阅行程事件
func (s *PaymentEventSubscriber) SubscribeToTripEvents(ctx context.Context) error {
	// 订阅司机分配事件
	err := s.subscriber.Subscribe(
		"create_payment_session_queue",
		contracts.TripEventDriverAssigned,
		s.handleDriverAssigned,
	)
	if err != nil {
		return fmt.Errorf("订阅司机分配事件失败: %w", err)
	}

	log.Println("成功订阅行程事件")
	return nil
}

// handleDriverAssigned 处理司机分配事件
func (s *PaymentEventSubscriber) handleDriverAssigned(data []byte) error {
	var trip pb.Trip
	if err := json.Unmarshal(data, &trip); err != nil {
		return fmt.Errorf("解析司机分配事件失败: %w", err)
	}

	// 检查行程是否有费用信息
	if trip.SelectedFare == nil {
		log.Printf("行程缺少费用信息，跳过支付会话创建: 行程ID=%s", trip.Id)
		return nil
	}

	// 创建支付会话
	payment, err := s.service.CreatePaymentSession(
		context.Background(),
		trip.Id,
		trip.UserId,
		trip.SelectedFare.TotalPriceInCents,
		"usd", // 默认货币，实际应用中应该从配置获取
	)
	if err != nil {
		return fmt.Errorf("创建支付会话失败: %w", err)
	}

	log.Printf("成功创建支付会话: 行程ID=%s, 支付ID=%s, 金额=%.2f", 
		trip.Id, payment.ID, payment.Amount)
	
	return nil
}

// Close 关闭订阅器
func (s *PaymentEventSubscriber) Close() error {
	return s.subscriber.Close()
}