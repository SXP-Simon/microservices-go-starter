package domain

import (
	"context"
	"time"
)

// PaymentModel 支付模型
type PaymentModel struct {
	ID          string    `json:"id"`
	TripID      string    `json:"tripID"`
	UserID      string    `json:"userID"`
	Amount      float64   `json:"amount"`
	Currency    string    `json:"currency"`
	Status      string    `json:"status"` // pending, succeeded, failed, cancelled
	SessionID   string    `json:"sessionID"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// PaymentRepository 支付存储库接口
type PaymentRepository interface {
	CreatePayment(ctx context.Context, payment *PaymentModel) (*PaymentModel, error)
	GetPaymentByID(ctx context.Context, id string) (*PaymentModel, error)
	GetPaymentByTripID(ctx context.Context, tripID string) (*PaymentModel, error)
	GetPaymentBySessionID(ctx context.Context, sessionID string) (*PaymentModel, error)
	UpdatePaymentStatus(ctx context.Context, id, status string) error
}

// PaymentService 支付服务接口
type PaymentService interface {
	CreatePaymentSession(ctx context.Context, tripID, userID string, amount float64, currency string) (*PaymentModel, error)
	ProcessPaymentWebhook(ctx context.Context, sessionID, status string) error
	GetPaymentByTripID(ctx context.Context, tripID string) (*PaymentModel, error)
}

// PaymentEventPublisher 支付事件发布器接口
type PaymentEventPublisher interface {
	PublishPaymentSessionCreated(ctx context.Context, payment *PaymentModel) error
	PublishPaymentSuccess(ctx context.Context, payment *PaymentModel) error
	PublishPaymentFailed(ctx context.Context, payment *PaymentModel) error
	PublishPaymentCancelled(ctx context.Context, payment *PaymentModel) error
	Close() error
}

// 支付状态常量
const (
	PaymentStatusPending   = "pending"
	PaymentStatusSucceeded = "succeeded"
	PaymentStatusFailed    = "failed"
	PaymentStatusCancelled = "cancelled"
)

// StripeSessionRequest Stripe会话请求
type StripeSessionRequest struct {
	TripID      string  `json:"tripID"`
	UserID      string  `json:"userID"`
	Amount      float64 `json:"amount"`
	Currency    string  `json:"currency"`
	SuccessURL  string  `json:"successURL"`
	CancelURL   string  `json:"cancelURL"`
}

// StripeSessionResponse Stripe会话响应
type StripeSessionResponse struct {
	SessionID string `json:"sessionID"`
	URL       string `json:"url"`
}