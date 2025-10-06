package grpc

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"ride-sharing/payment-service/internal/domain"
	"ride-sharing/payment-service/internal/service"
)

// gRPCHandler gRPC处理器
type gRPCHandler struct {
	service service.PaymentService
}

// NewGRPCHandler 创建新的gRPC处理器
func NewGRPCHandler(server *grpc.Server, paymentService service.PaymentService) *gRPCHandler {
	handler := &gRPCHandler{
		service: paymentService,
	}

	// 注册gRPC服务（这里需要定义proto文件）
	// pb.RegisterPaymentServiceServer(server, handler)

	return handler
}

// CreatePaymentSession 创建支付会话
func (h *gRPCHandler) CreatePaymentSession(ctx context.Context, req *CreatePaymentSessionRequest) (*CreatePaymentSessionResponse, error) {
	// 验证请求参数
	if req.TripId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "行程ID不能为空")
	}
	if req.UserId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "用户ID不能为空")
	}
	if req.Amount <= 0 {
		return nil, status.Errorf(codes.InvalidArgument, "金额必须大于0")
	}

	// 创建支付会话
	payment, err := h.service.CreatePaymentSession(ctx, req.TripId, req.UserId, req.Amount, req.Currency)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "创建支付会话失败: %v", err)
	}

	// 构建响应
	response := &CreatePaymentSessionResponse{
		PaymentId: payment.ID,
		SessionId: payment.SessionID,
		Amount:    payment.Amount,
		Currency:  payment.Currency,
		Status:    payment.Status,
	}

	log.Printf("成功创建支付会话: 支付ID=%s, 行程ID=%s", payment.ID, req.TripId)
	return response, nil
}

// GetPaymentByTripId 根据行程ID获取支付记录
func (h *gRPCHandler) GetPaymentByTripId(ctx context.Context, req *GetPaymentByTripIdRequest) (*GetPaymentByTripIdResponse, error) {
	// 验证请求参数
	if req.TripId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "行程ID不能为空")
	}

	// 获取支付记录
	payment, err := h.service.GetPaymentByTripID(ctx, req.TripId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "获取支付记录失败: %v", err)
	}

	// 构建响应
	response := &GetPaymentByTripIdResponse{
		PaymentId: payment.ID,
		TripId:    payment.TripID,
		UserId:    payment.UserID,
		Amount:    payment.Amount,
		Currency:  payment.Currency,
		Status:    payment.Status,
		SessionId: payment.SessionID,
		CreatedAt: payment.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: payment.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	log.Printf("成功获取支付记录: 支付ID=%s, 行程ID=%s", payment.ID, req.TripId)
	return response, nil
}

// 以下是临时的请求/响应结构体定义，实际应该从proto文件生成
// TODO: 创建payment.proto文件并使用protoc生成代码

type CreatePaymentSessionRequest struct {
	TripId   string  `json:"tripId"`
	UserId   string  `json:"userId"`
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

type CreatePaymentSessionResponse struct {
	PaymentId string  `json:"paymentId"`
	SessionId string  `json:"sessionId"`
	Amount    float64 `json:"amount"`
	Currency  string  `json:"currency"`
	Status    string  `json:"status"`
}

type GetPaymentByTripIdRequest struct {
	TripId string `json:"tripId"`
}

type GetPaymentByTripIdResponse struct {
	PaymentId string `json:"paymentId"`
	TripId    string `json:"tripId"`
	UserId    string `json:"userId"`
	Amount    float64 `json:"amount"`
	Currency  string `json:"currency"`
	Status    string `json:"status"`
	SessionId string `json:"sessionId"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

// 为了让代码能够编译，这里添加一些临时的gRPC接口实现
// 实际应用中应该从proto文件生成

func (h *gRPCHandler) ProcessWebhook(ctx context.Context, req *WebhookRequest) (*WebhookResponse, error) {
	log.Printf("收到支付Webhook请求: %+v", req)
	
	// 解析Webhook数据
	var webhookData map[string]interface{}
	if err := json.Unmarshal([]byte(req.Data), &webhookData); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "解析Webhook数据失败: %v", err)
	}
	
	// 提取会话ID和状态
	sessionID, ok := webhookData["sessionID"].(string)
	if !ok {
		return nil, status.Errorf(codes.InvalidArgument, "缺少会话ID")
	}
	
	status, ok := webhookData["status"].(string)
	if !ok {
		return nil, status.Errorf(codes.InvalidArgument, "缺少支付状态")
	}
	
	// 处理Webhook
	if err := h.service.ProcessPaymentWebhook(ctx, sessionID, status); err != nil {
		return nil, status.Errorf(codes.Internal, "处理Webhook失败: %v", err)
	}
	
	return &WebhookResponse{Success: true}, nil
}

type WebhookRequest struct {
	Data string `json:"data"`
}

type WebhookResponse struct {
	Success bool `json:"success"`
}