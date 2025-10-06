package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"ride-sharing/payment-service/internal/domain"
)

type service struct {
	repo     domain.PaymentRepository
	publisher domain.PaymentEventPublisher
}

// NewService 创建新的支付服务
func NewService(repo domain.PaymentRepository, publisher domain.PaymentEventPublisher) domain.PaymentService {
	return &service{
		repo:     repo,
		publisher: publisher,
	}
}

// CreatePaymentSession 创建支付会话
func (s *service) CreatePaymentSession(ctx context.Context, tripID, userID string, amount float64, currency string) (*domain.PaymentModel, error) {
	// 创建支付记录
	payment := &domain.PaymentModel{
		ID:        generatePaymentID(),
		TripID:    tripID,
		UserID:    userID,
		Amount:    amount,
		Currency:  currency,
		Status:    domain.PaymentStatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 保存支付记录
	savedPayment, err := s.repo.CreatePayment(ctx, payment)
	if err != nil {
		return nil, fmt.Errorf("创建支付记录失败: %w", err)
	}

	// TODO: 调用Stripe API创建支付会话
	// 这里暂时使用模拟数据
	stripeSessionID := fmt.Sprintf("cs_test_%s", savedPayment.ID)
	savedPayment.SessionID = stripeSessionID

	// 更新支付记录
	if err := s.repo.UpdatePaymentStatus(ctx, savedPayment.ID, domain.PaymentStatusPending); err != nil {
		log.Printf("更新支付状态失败: %v", err)
	}

	// 发布支付会话创建事件
	if s.publisher != nil {
		if err := s.publisher.PublishPaymentSessionCreated(ctx, savedPayment); err != nil {
			log.Printf("发布支付会话创建事件失败: %v", err)
		}
	}

	log.Printf("成功创建支付会话: 支付ID=%s, 行程ID=%s", savedPayment.ID, tripID)
	return savedPayment, nil
}

// ProcessPaymentWebhook 处理支付Webhook
func (s *service) ProcessPaymentWebhook(ctx context.Context, sessionID, status string) error {
	// 根据会话ID获取支付记录
	payment, err := s.repo.GetPaymentBySessionID(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("获取支付记录失败: %w", err)
	}

	if payment == nil {
		return fmt.Errorf("支付记录不存在")
	}

	// 更新支付状态
	var newStatus string
	switch status {
	case "payment_intent.succeeded":
		newStatus = domain.PaymentStatusSucceeded
	case "payment_intent.payment_failed":
		newStatus = domain.PaymentStatusFailed
	case "checkout.session.expired":
		newStatus = domain.PaymentStatusCancelled
	default:
		return fmt.Errorf("未知的支付状态: %s", status)
	}

	// 更新数据库中的状态
	if err := s.repo.UpdatePaymentStatus(ctx, payment.ID, newStatus); err != nil {
		return fmt.Errorf("更新支付状态失败: %w", err)
	}

	// 更新模型中的状态
	payment.Status = newStatus
	payment.UpdatedAt = time.Now()

	// 根据状态发布相应事件
	if s.publisher != nil {
		switch newStatus {
		case domain.PaymentStatusSucceeded:
			if err := s.publisher.PublishPaymentSuccess(ctx, payment); err != nil {
				log.Printf("发布支付成功事件失败: %v", err)
			}
		case domain.PaymentStatusFailed:
			if err := s.publisher.PublishPaymentFailed(ctx, payment); err != nil {
				log.Printf("发布支付失败事件失败: %v", err)
			}
		case domain.PaymentStatusCancelled:
			if err := s.publisher.PublishPaymentCancelled(ctx, payment); err != nil {
				log.Printf("发布支付取消事件失败: %v", err)
			}
		}
	}

	log.Printf("成功处理支付Webhook: 支付ID=%s, 状态=%s", payment.ID, newStatus)
	return nil
}

// GetPaymentByTripID 根据行程ID获取支付记录
func (s *service) GetPaymentByTripID(ctx context.Context, tripID string) (*domain.PaymentModel, error) {
	payment, err := s.repo.GetPaymentByTripID(ctx, tripID)
	if err != nil {
		return nil, fmt.Errorf("获取支付记录失败: %w", err)
	}

	return payment, nil
}

// generatePaymentID 生成支付ID
func generatePaymentID() string {
	// 简单实现，实际应用中可以使用更复杂的ID生成策略
	return fmt.Sprintf("pay_%d", time.Now().UnixNano())
}