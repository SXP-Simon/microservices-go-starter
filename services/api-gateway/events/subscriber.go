package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"ride-sharing/api-gateway/websocket"
	"ride-sharing/shared/events"
	"ride-sharing/shared/contracts"
	pb "ride-sharing/shared/proto/trip"
)

// GatewayEventSubscriber API网关事件订阅器
type GatewayEventSubscriber struct {
	subscriber events.Subscriber
	wsManager  *websocket.WebSocketManager
}

// NewGatewayEventSubscriber 创建网关事件订阅器
func NewGatewayEventSubscriber(subscriber events.Subscriber, wsManager *websocket.WebSocketManager) *GatewayEventSubscriber {
	return &GatewayEventSubscriber{
		subscriber: subscriber,
		wsManager:  wsManager,
	}
}

// SubscribeToAllEvents 订阅所有事件
func (s *GatewayEventSubscriber) SubscribeToAllEvents(ctx context.Context) error {
	// 订阅行程创建事件
	err := s.subscriber.Subscribe(
		"notify_new_trip_queue",
		contracts.TripEventCreated,
		s.handleTripCreated,
	)
	if err != nil {
		return fmt.Errorf("订阅行程创建事件失败: %w", err)
	}

	// 订阅司机分配事件
	err = s.subscriber.Subscribe(
		"notify_driver_assignment_queue",
		contracts.TripEventDriverAssigned,
		s.handleDriverAssigned,
	)
	if err != nil {
		return fmt.Errorf("订阅司机分配事件失败: %w", err)
	}

	// 订阅未找到司机事件
	err = s.subscriber.Subscribe(
		"notify_no_drivers_found_queue",
		contracts.TripEventNoDriversFound,
		s.handleNoDriversFound,
	)
	if err != nil {
		return fmt.Errorf("订阅未找到司机事件失败: %w", err)
	}

	// 订阅司机行程请求事件
	err = s.subscriber.Subscribe(
		"notify_drivers_queue",
		contracts.DriverCmdTripRequest,
		s.handleDriverTripRequest,
	)
	if err != nil {
		return fmt.Errorf("订阅司机行程请求事件失败: %w", err)
	}

	// 订阅支付会话创建事件
	err = s.subscriber.Subscribe(
		"notify_payment_status_queue",
		contracts.PaymentEventSessionCreated,
		s.handlePaymentSessionCreated,
	)
	if err != nil {
		return fmt.Errorf("订阅支付会话创建事件失败: %w", err)
	}

	// 订阅支付成功事件
	err = s.subscriber.Subscribe(
		"notify_payment_success_queue",
		contracts.PaymentEventSuccess,
		s.handlePaymentSuccess,
	)
	if err != nil {
		return fmt.Errorf("订阅支付成功事件失败: %w", err)
	}

	// 订阅支付失败事件
	err = s.subscriber.Subscribe(
		"notify_payment_failed_queue",
		contracts.PaymentEventFailed,
		s.handlePaymentFailed,
	)
	if err != nil {
		return fmt.Errorf("订阅支付失败事件失败: %w", err)
	}

	log.Println("成功订阅所有事件")
	return nil
}

// handleTripCreated 处理行程创建事件
func (s *GatewayEventSubscriber) handleTripCreated(data []byte) error {
	var trip pb.Trip
	if err := json.Unmarshal(data, &trip); err != nil {
		return fmt.Errorf("解析行程创建事件失败: %w", err)
	}

	// 向乘客发送行程创建确认
	message := contracts.WSMessage{
		Type: contracts.TripEventCreated,
		Data: trip,
	}

	// 发送给特定乘客
	if err := s.wsManager.SendToRider(trip.UserId, message); err != nil {
		log.Printf("向乘客发送行程创建事件失败: %v", err)
	}

	log.Printf("已向乘客发送行程创建事件: 乘客ID=%s, 行程ID=%s", trip.UserId, trip.Id)
	return nil
}

// handleDriverAssigned 处理司机分配事件
func (s *GatewayEventSubscriber) handleDriverAssigned(data []byte) error {
	var trip pb.Trip
	if err := json.Unmarshal(data, &trip); err != nil {
		return fmt.Errorf("解析司机分配事件失败: %w", err)
	}

	// 向乘客发送司机分配通知
	message := contracts.WSMessage{
		Type: contracts.TripEventDriverAssigned,
		Data: trip,
	}

	// 发送给特定乘客
	if err := s.wsManager.SendToRider(trip.UserId, message); err != nil {
		log.Printf("向乘客发送司机分配事件失败: %v", err)
	}

	// 向司机发送行程确认
	driverMessage := contracts.WSMessage{
		Type: contracts.TripEventDriverAssigned,
		Data: trip,
	}

	if trip.Driver != nil {
		if err := s.wsManager.SendToDriver(trip.Driver.Id, driverMessage); err != nil {
			log.Printf("向司机发送行程确认失败: %v", err)
		}
	}

	log.Printf("已发送司机分配事件: 乘客ID=%s, 司机ID=%s, 行程ID=%s", 
		trip.UserId, trip.Driver.Id, trip.Id)
	return nil
}

// handleNoDriversFound 处理未找到司机事件
func (s *GatewayEventSubscriber) handleNoDriversFound(data []byte) error {
	var eventData map[string]string
	if err := json.Unmarshal(data, &eventData); err != nil {
		return fmt.Errorf("解析未找到司机事件失败: %w", err)
	}

	tripID, ok := eventData["tripID"]
	if !ok {
		return fmt.Errorf("事件数据中缺少tripID")
	}

	// TODO: 需要获取乘客ID，这里简化处理
	// 在实际实现中，可能需要从数据库查询行程信息获取乘客ID
	
	message := contracts.WSMessage{
		Type: contracts.TripEventNoDriversFound,
		Data: map[string]string{
			"tripID": tripID,
			"message": "未找到可用司机，请稍后再试",
		},
	}

	// 向所有乘客广播（简化处理）
	if err := s.wsManager.BroadcastToRiders(message); err != nil {
		log.Printf("广播未找到司机事件失败: %v", err)
	}

	log.Printf("已发送未找到司机事件: 行程ID=%s", tripID)
	return nil
}

// handlePaymentSessionCreated 处理支付会话创建事件
func (s *GatewayEventSubscriber) handlePaymentSessionCreated(data []byte) error {
	var paymentData map[string]interface{}
	if err := json.Unmarshal(data, &paymentData); err != nil {
		return fmt.Errorf("解析支付会话创建事件失败: %w", err)
	}

	tripID, ok := paymentData["tripID"].(string)
	if !ok {
		return fmt.Errorf("事件数据中缺少tripID")
	}

	// TODO: 需要获取乘客ID，这里简化处理
	
	message := contracts.WSMessage{
		Type: contracts.PaymentEventSessionCreated,
		Data: paymentData,
	}

	// 向所有乘客广播（简化处理）
	if err := s.wsManager.BroadcastToRiders(message); err != nil {
		log.Printf("广播支付会话创建事件失败: %v", err)
	}

	log.Printf("已发送支付会话创建事件: 行程ID=%s", tripID)
	return nil
}

// handleDriverTripRequest 处理司机行程请求事件
func (s *GatewayEventSubscriber) handleDriverTripRequest(data []byte) error {
	var requestData map[string]interface{}
	if err := json.Unmarshal(data, &requestData); err != nil {
		return fmt.Errorf("解析司机行程请求事件失败: %w", err)
	}

	driverID, ok := requestData["driverID"].(string)
	if !ok {
		return fmt.Errorf("事件数据中缺少driverID")
	}

	// 向特定司机发送行程请求
	message := contracts.WSMessage{
		Type: contracts.DriverCmdTripRequest,
		Data: requestData,
	}

	if err := s.wsManager.SendToDriver(driverID, message); err != nil {
		log.Printf("向司机发送行程请求失败: %v", err)
	}

	log.Printf("已向司机发送行程请求: 司机ID=%s, 行程ID=%v",
		driverID, requestData["tripID"])
	return nil
}

// handlePaymentSuccess 处理支付成功事件
func (s *GatewayEventSubscriber) handlePaymentSuccess(data []byte) error {
	var paymentData map[string]interface{}
	if err := json.Unmarshal(data, &paymentData); err != nil {
		return fmt.Errorf("解析支付成功事件失败: %w", err)
	}

	tripID, ok := paymentData["tripID"].(string)
	if !ok {
		return fmt.Errorf("事件数据中缺少tripID")
	}

	message := contracts.WSMessage{
		Type: contracts.PaymentEventSuccess,
		Data: paymentData,
	}

	// 向所有乘客广播（简化处理）
	if err := s.wsManager.BroadcastToRiders(message); err != nil {
		log.Printf("广播支付成功事件失败: %v", err)
	}

	log.Printf("已发送支付成功事件: 行程ID=%s", tripID)
	return nil
}

// handlePaymentFailed 处理支付失败事件
func (s *GatewayEventSubscriber) handlePaymentFailed(data []byte) error {
	var paymentData map[string]interface{}
	if err := json.Unmarshal(data, &paymentData); err != nil {
		return fmt.Errorf("解析支付失败事件失败: %w", err)
	}

	tripID, ok := paymentData["tripID"].(string)
	if !ok {
		return fmt.Errorf("事件数据中缺少tripID")
	}

	message := contracts.WSMessage{
		Type: contracts.PaymentEventFailed,
		Data: paymentData,
	}

	// 向所有乘客广播（简化处理）
	if err := s.wsManager.BroadcastToRiders(message); err != nil {
		log.Printf("广播支付失败事件失败: %v", err)
	}

	log.Printf("已发送支付失败事件: 行程ID=%s", tripID)
	return nil
}

// Close 关闭订阅器
func (s *GatewayEventSubscriber) Close() error {
	return s.subscriber.Close()
}