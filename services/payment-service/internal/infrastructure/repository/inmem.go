package repository

import (
	"context"
	"fmt"
	"sync"
	"time"

	"ride-sharing/payment-service/internal/domain"
)

// inmemRepository 内存存储库实现
type inmemRepository struct {
	payments   map[string]*domain.PaymentModel
	tripIndex  map[string]string // tripID -> paymentID
	sessionIndex map[string]string // sessionID -> paymentID
	mu         sync.RWMutex
}

// NewInmemRepository 创建新的内存存储库
func NewInmemRepository() *inmemRepository {
	return &inmemRepository{
		payments:     make(map[string]*domain.PaymentModel),
		tripIndex:    make(map[string]string),
		sessionIndex: make(map[string]string),
	}
}

// CreatePayment 创建支付记录
func (r *inmemRepository) CreatePayment(ctx context.Context, payment *domain.PaymentModel) (*domain.PaymentModel, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 检查是否已存在相同行程的支付记录
	if existingPaymentID, exists := r.tripIndex[payment.TripID]; exists {
		existingPayment := r.payments[existingPaymentID]
		if existingPayment.Status != domain.PaymentStatusFailed && 
		   existingPayment.Status != domain.PaymentStatusCancelled {
			return nil, fmt.Errorf("行程已存在未完成的支付记录")
		}
	}

	// 保存支付记录
	r.payments[payment.ID] = payment
	r.tripIndex[payment.TripID] = payment.ID
	
	if payment.SessionID != "" {
		r.sessionIndex[payment.SessionID] = payment.ID
	}

	return payment, nil
}

// GetPaymentByID 根据ID获取支付记录
func (r *inmemRepository) GetPaymentByID(ctx context.Context, id string) (*domain.PaymentModel, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	payment, exists := r.payments[id]
	if !exists {
		return nil, fmt.Errorf("支付记录不存在")
	}

	return payment, nil
}

// GetPaymentByTripID 根据行程ID获取支付记录
func (r *inmemRepository) GetPaymentByTripID(ctx context.Context, tripID string) (*domain.PaymentModel, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	paymentID, exists := r.tripIndex[tripID]
	if !exists {
		return nil, fmt.Errorf("行程支付记录不存在")
	}

	payment, exists := r.payments[paymentID]
	if !exists {
		return nil, fmt.Errorf("支付记录不存在")
	}

	return payment, nil
}

// GetPaymentBySessionID 根据会话ID获取支付记录
func (r *inmemRepository) GetPaymentBySessionID(ctx context.Context, sessionID string) (*domain.PaymentModel, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	paymentID, exists := r.sessionIndex[sessionID]
	if !exists {
		return nil, fmt.Errorf("会话支付记录不存在")
	}

	payment, exists := r.payments[paymentID]
	if !exists {
		return nil, fmt.Errorf("支付记录不存在")
	}

	return payment, nil
}

// UpdatePaymentStatus 更新支付状态
func (r *inmemRepository) UpdatePaymentStatus(ctx context.Context, id, status string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	payment, exists := r.payments[id]
	if !exists {
		return fmt.Errorf("支付记录不存在")
	}

	payment.Status = status
	payment.UpdatedAt = time.Now()

	return nil
}

// CleanupExpiredPayments 清理过期的支付记录（可选的维护方法）
func (r *inmemRepository) CleanupExpiredPayments(ctx context.Context, expiryDuration time.Duration) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	for id, payment := range r.payments {
		// 清理超过24小时的失败或取消的支付记录
		if (payment.Status == domain.PaymentStatusFailed || 
			payment.Status == domain.PaymentStatusCancelled) &&
			now.Sub(payment.UpdatedAt) > expiryDuration {
			
			delete(r.payments, id)
			delete(r.tripIndex, payment.TripID)
			if payment.SessionID != "" {
				delete(r.sessionIndex, payment.SessionID)
			}
		}
	}

	return nil
}