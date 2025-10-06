package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"ride-sharing/api-gateway/events"
	"ride-sharing/api-gateway/grpc_clients"
	"ride-sharing/api-gateway/websocket"
	"ride-sharing/shared/contracts"
	"ride-sharing/shared/proto/driver"
)

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

func HandleRidersWebSocket(wsManager *websocket.WebSocketManager, eventPublisher *events.GatewayEventPublisher, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket升级失败: %v", err)
		return
	}

	userID := r.URL.Query().Get("userID")
	if userID == "" {
		log.Println("未提供用户ID")
		conn.Close()
		return
	}

	// 添加连接到管理器
	wsManager.AddRiderConnection(userID, conn)
	defer wsManager.RemoveRiderConnection(userID)

	log.Printf("乘客WebSocket连接已建立: %s", userID)

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("读取乘客消息失败: %v", err)
			break
		}
		
		// 处理来自乘客的消息
		if err := handleRiderMessage(userID, message); err != nil {
			log.Printf("处理乘客消息失败: %v", err)
		}
	}
}

func HandleDriversWebSocket(wsManager *websocket.WebSocketManager, eventPublisher *events.GatewayEventPublisher, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket升级失败: %v", err)
		return
	}

	userID := r.URL.Query().Get("userID")
	if userID == "" {
		log.Println("未提供用户ID")
		conn.Close()
		return
	}
	
	packageSlug := r.URL.Query().Get("packageSlug")
	if packageSlug == "" {
		log.Println("未提供套餐类型")
		conn.Close()
		return
	}

	// 添加连接到管理器
	wsManager.AddDriverConnection(userID, conn)
	defer wsManager.RemoveDriverConnection(userID)

	ctx := r.Context()

	driverService, err := grpc_clients.NewDriverServiceClient()
	if err != nil {
		log.Fatalf("创建司机服务客户端失败: %v", err)
	}
	defer driverService.Close()

	// 注册司机
	driverData, err := driverService.Client.RegisterDriver(ctx, &driver.RegisterDriverRequest{
		DriverID:    userID,
		PackageSlug: packageSlug,
	})
	if err != nil {
		log.Printf("注册司机失败: %v", err)
		return
	}

	// 发送注册成功消息
	msg := contracts.WSMessage{
		Type: contracts.DriverRegister,
		Data: driverData.Driver,
	}

	if err := conn.WriteJSON(msg); err != nil {
		log.Printf("发送注册消息失败: %v", err)
	}

	log.Printf("司机WebSocket连接已建立: %s, 套餐: %s", userID, packageSlug)

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("读取司机消息失败: %v", err)
			break
		}
		
		// 处理来自司机的消息
		if err := handleDriverMessage(userID, message, eventPublisher); err != nil {
			log.Printf("处理司机消息失败: %v", err)
		}
	}

	// 注销司机
	if _, err := driverService.Client.UnRegisterDriver(ctx, &driver.RegisterDriverRequest{
		DriverID:    userID,
		PackageSlug: packageSlug,
	}); err != nil {
		log.Printf("注销司机失败: %v", err)
	} else {
		log.Printf("司机已注销: %s", userID)
	}
}

// handleRiderMessage 处理乘客消息
func handleRiderMessage(userID string, message []byte) error {
	var wsMessage contracts.WSMessage
	if err := json.Unmarshal(message, &wsMessage); err != nil {
		log.Printf("解析乘客消息失败: 用户ID=%s, 错误=%v", userID, err)
		return err
	}
	
	log.Printf("收到乘客消息: 用户ID=%s, 类型=%s", userID, wsMessage.Type)
	
	// TODO: 根据消息类型处理不同的乘客消息
	// 例如：取消行程、更新位置等
	
	return nil
}

// handleDriverMessage 处理司机消息
func handleDriverMessage(driverID string, message []byte, eventPublisher *events.GatewayEventPublisher) error {
	var wsMessage contracts.WSMessage
	if err := json.Unmarshal(message, &wsMessage); err != nil {
		log.Printf("解析司机消息失败: 司机ID=%s, 错误=%v", driverID, err)
		return err
	}
	
	log.Printf("收到司机消息: 司机ID=%s, 类型=%s", driverID, wsMessage.Type)
	
	switch wsMessage.Type {
	case contracts.DriverCmdTripAccept, contracts.DriverCmdTripDecline:
		return handleDriverTripResponse(driverID, wsMessage, eventPublisher)
	case contracts.DriverCmdLocation:
		return handleDriverLocationUpdate(driverID, wsMessage, eventPublisher)
	default:
		log.Printf("未知的司机消息类型: %s", wsMessage.Type)
	}
	
	return nil
}

// handleDriverTripResponse 处理司机行程响应
func handleDriverTripResponse(driverID string, wsMessage contracts.WSMessage, eventPublisher *events.GatewayEventPublisher) error {
	// 解析响应数据
	var responseData map[string]interface{}
	if err := json.Unmarshal(wsMessage.Data.([]byte), &responseData); err != nil {
		log.Printf("解析司机响应数据失败: 司机ID=%s, 错误=%v", driverID, err)
		return err
	}
	
	tripID, ok := responseData["tripID"].(string)
	if !ok {
		log.Printf("司机响应中缺少行程ID: 司机ID=%s", driverID)
		return fmt.Errorf("缺少行程ID")
	}
	
	riderID, ok := responseData["riderID"].(string)
	if !ok {
		log.Printf("司机响应中缺少乘客ID: 司机ID=%s", driverID)
		return fmt.Errorf("缺少乘客ID")
	}
	
	accept := wsMessage.Type == contracts.DriverCmdTripAccept
	
	log.Printf("处理司机行程响应: 司机ID=%s, 行程ID=%s, 接受=%v", driverID, tripID, accept)
	
	// 发布司机响应命令到RabbitMQ
	if eventPublisher != nil {
		response := events.DriverTripResponse{
			TripID:   tripID,
			RiderID:  riderID,
			DriverID: driverID,
			Accept:   accept,
		}
		
		if err := eventPublisher.PublishDriverTripResponse(context.Background(), response); err != nil {
			log.Printf("发布司机响应命令失败: %v", err)
			return err
		}
		
		log.Printf("成功发布司机响应命令: 司机ID=%s, 行程ID=%s", driverID, tripID)
	} else {
		log.Printf("事件发布器未初始化，无法发布司机响应")
		return fmt.Errorf("事件发布器未初始化")
	}
	
	return nil
}

// handleDriverLocationUpdate 处理司机位置更新
func handleDriverLocationUpdate(driverID string, wsMessage contracts.WSMessage, eventPublisher *events.GatewayEventPublisher) error {
	// 解析位置数据
	var locationData map[string]interface{}
	if err := json.Unmarshal(wsMessage.Data.([]byte), &locationData); err != nil {
		log.Printf("解析司机位置数据失败: 司机ID=%s, 错误=%v", driverID, err)
		return err
	}
	
	latitude, ok := locationData["latitude"].(float64)
	if !ok {
		log.Printf("司机位置更新中缺少纬度信息: 司机ID=%s", driverID)
		return fmt.Errorf("缺少纬度信息")
	}
	
	longitude, ok := locationData["longitude"].(float64)
	if !ok {
		log.Printf("司机位置更新中缺少经度信息: 司机ID=%s", driverID)
		return fmt.Errorf("缺少经度信息")
	}
	
	log.Printf("处理司机位置更新: 司机ID=%s, 位置=(%.6f, %.6f)", driverID, latitude, longitude)
	
	// 发布司机位置更新事件
	if eventPublisher != nil {
		locationUpdate := events.DriverLocationUpdate{
			DriverID:  driverID,
			Latitude:  latitude,
			Longitude: longitude,
			Timestamp: time.Now().Unix(),
		}
		
		if err := eventPublisher.PublishDriverLocationUpdate(context.Background(), locationUpdate); err != nil {
			log.Printf("发布司机位置更新命令失败: %v", err)
			return err
		}
		
		log.Printf("成功发布司机位置更新命令: 司机ID=%s", driverID)
	} else {
		log.Printf("事件发布器未初始化，无法发布司机位置更新")
		return fmt.Errorf("事件发布器未初始化")
	}
	
	return nil
}
