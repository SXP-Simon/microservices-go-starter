package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"ride-sharing/payment-service/internal/domain"
	"ride-sharing/shared/events"
	"ride-sharing/shared/contracts"
)

// PaymentEventPublisher 支付事件发布器实现
type PaymentEventPublisher struct {
	publisher events.Publisher
}

// NewPaymentEventPublisher 创建支付事件发布器
func NewPaymentEventPublisher(publisher events.Publisher) *PaymentEventPublisher {
	return &PaymentEventPublisher{
		publisher: publisher,
	}
}

// PublishPaymentSessionCreated 发布支付会话创建事件
func (p *PaymentEventPublisher) PublishPaymentSessionCreated(ctx context.Context, payment *domain.PaymentModel) error {
	// 创建事件数据
	eventData := map[string]interface{}{
		"tripID":    payment.TripID,
		"sessionID": payment.SessionID,
		"amount":    payment.Amount,
		"currency":  payment.Currency,
		"userID":    payment.UserID,
	}
	
	// 发布事件
	err := p.publisher.PublishEvent(contracts.PaymentEventSessionCreated, eventData)
	if err != nil {
		return fmt.Errorf("发布支付会话创建事件失败: %w", err)
	}
	
	log.Printf("成功发布支付会话创建事件: 支付ID=%s, 会话ID=%s", payment.ID, payment.SessionID)
	return nil
}

// PublishPaymentSuccess 发布支付成功事件
func (p *PaymentEventPublisher) PublishPaymentSuccess(ctx context.Context, payment *domain.PaymentModel) error {
	// 创建事件数据
	eventData := map[string]interface{}{
		"tripID":    payment.TripID,
		"sessionID": payment.SessionID,
		"amount":    payment.Amount,
		"currency":  payment.Currency,
		"userID":    payment.UserID,
		"status":    payment.Status,
	}
	
	// 发布事件
	err := p.publisher.PublishEvent(contracts.PaymentEventSuccess, eventData)
	if err != nil {
		return fmt.Errorf("发布支付成功事件失败: %w", err)
	}
	
	log.Printf("成功发布支付成功事件: 支付ID=%s, 会话ID=%s", payment.ID, payment.SessionID)
	return nil
}

// PublishPaymentFailed 发布支付失败事件
func (p *PaymentEventPublisher) PublishPaymentFailed(ctx context.Context, payment *domain.PaymentModel) error {
	// 创建事件数据
	eventData := map[string]interface{}{
		"tripID":    payment.TripID,
		"sessionID": payment.SessionID,
		"amount":    payment.Amount,
		"currency":  payment.Currency,
		"userID":    payment.UserID,
		"status":    payment.Status,
	}
	
	// 发布事件
	err := p.publisher.PublishEvent(contracts.PaymentEventFailed, eventData)
	if err != nil {
		return fmt.Errorf("发布支付失败事件失败: %w", err)
	}
	
	log.Printf("成功发布支付失败事件: 支付ID=%s, 会话ID=%s", payment.ID, payment.SessionID)
	return nil
}

// PublishPaymentCancelled 发布支付取消事件
func (p *PaymentEventPublisher) PublishPaymentCancelled(ctx context.Context, payment *domain.PaymentModel) error {
	// 创建事件数据
	eventData := map[string]interface{}{
		"tripID":    payment.TripID,
		"sessionID": payment.SessionID,
		"amount":    payment.Amount,
		"currency":  payment.Currency,
		"userID":    payment.UserID,
		"status":    payment.Status,
	}
	
	// 发布事件
	err := p.publisher.PublishEvent(contracts.PaymentEventCancelled, eventData)
	if err != nil {
		return fmt.Errorf("发布支付取消事件失败: %w", err)
	}
	
	log.Printf("成功发布支付取消事件: 支付ID=%s, 会话ID=%s", payment.ID, payment.SessionID)
	return nil
}

// Close 关闭发布器
func (p *PaymentEventPublisher) Close() error {
	return p.publisher.Close()
}