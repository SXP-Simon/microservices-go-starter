package main

import (
	"encoding/json"
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
	wsManager *websocket.WebSocketManager
)

func HandleRidersWebSocket(w http.ResponseWriter, r *http.Request) {
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

func HandleDriversWebSocket(w http.ResponseWriter, r *http.Request) {
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
		if err := handleDriverMessage(userID, message); err != nil {
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
		return err
	}
	
	log.Printf("收到乘客消息: 用户ID=%s, 类型=%s", userID, wsMessage.Type)
	
	// TODO: 根据消息类型处理不同的乘客消息
	// 例如：取消行程、更新位置等
	
	return nil
}

// handleDriverMessage 处理司机消息
func handleDriverMessage(driverID string, message []byte) error {
	var wsMessage contracts.WSMessage
	if err := json.Unmarshal(message, &wsMessage); err != nil {
		return err
	}
	
	log.Printf("收到司机消息: 司机ID=%s, 类型=%s", driverID, wsMessage.Type)
	
	// TODO: 根据消息类型处理不同的司机消息
	// 例如：接受/拒绝行程、更新位置等
	
	return nil
}
