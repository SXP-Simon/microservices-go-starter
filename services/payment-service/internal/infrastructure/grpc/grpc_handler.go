package grpc

import (
	"context"
	"encoding/json"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	pb "ride-sharing/shared/proto/payment"
	"ride-sharing/payment-service/internal/service"
)

// gRPCHandler gRPC处理器
type gRPCHandler struct {
	pb.UnimplementedPaymentServiceServer
	service service.PaymentService
}

// NewGRPCHandler 创建新的gRPC处理器
func NewGRPCHandler(server *grpc.Server, paymentService service.PaymentService) *gRPCHandler {
	handler := &gRPCHandler{
		service: paymentService,
	}

	// 注册gRPC服务
	pb.RegisterPaymentServiceServer(server, handler)

	return handler
}

// CreatePaymentSession 创建支付会话
func (h *gRPCHandler) CreatePaymentSession(ctx context.Context, req *pb.CreatePaymentSessionRequest) (*pb.CreatePaymentSessionResponse, error) {
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
	response := &pb.CreatePaymentSessionResponse{
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
func (h *gRPCHandler) GetPaymentByTripId(ctx context.Context, req *pb.GetPaymentByTripIdRequest) (*pb.GetPaymentByTripIdResponse, error) {
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
	response := &pb.GetPaymentByTripIdResponse{
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

func (h *gRPCHandler) ProcessWebhook(ctx context.Context, req *pb.WebhookRequest) (*pb.WebhookResponse, error) {
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
