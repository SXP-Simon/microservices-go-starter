package websocket

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gorilla/websocket"
	"ride-sharing/shared/contracts"
)

// WebSocketManager WebSocket连接管理器
type WebSocketManager struct {
	riders  map[string]*websocket.Conn
	drivers map[string]*websocket.Conn
	mu      sync.RWMutex
}

// NewWebSocketManager 创建WebSocket管理器
func NewWebSocketManager() *WebSocketManager {
	return &WebSocketManager{
		riders:  make(map[string]*websocket.Conn),
		drivers: make(map[string]*websocket.Conn),
	}
}

// AddRiderConnection 添加乘客连接
func (m *WebSocketManager) AddRiderConnection(userID string, conn *websocket.Conn) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// 如果用户已有连接，先关闭旧连接
	if oldConn, exists := m.riders[userID]; exists {
		oldConn.Close()
	}
	
	m.riders[userID] = conn
	log.Printf("乘客连接已添加: %s", userID)
}

// AddDriverConnection 添加司机连接
func (m *WebSocketManager) AddDriverConnection(driverID string, conn *websocket.Conn) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// 如果司机已有连接，先关闭旧连接
	if oldConn, exists := m.drivers[driverID]; exists {
		oldConn.Close()
	}
	
	m.drivers[driverID] = conn
	log.Printf("司机连接已添加: %s", driverID)
}

// RemoveRiderConnection 移除乘客连接
func (m *WebSocketManager) RemoveRiderConnection(userID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if conn, exists := m.riders[userID]; exists {
		conn.Close()
		delete(m.riders, userID)
		log.Printf("乘客连接已移除: %s", userID)
	}
}

// RemoveDriverConnection 移除司机连接
func (m *WebSocketManager) RemoveDriverConnection(driverID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if conn, exists := m.drivers[driverID]; exists {
		conn.Close()
		delete(m.drivers, driverID)
		log.Printf("司机连接已移除: %s", driverID)
	}
}

// BroadcastToRiders 向乘客广播消息
func (m *WebSocketManager) BroadcastToRiders(message contracts.WSMessage) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}
	
	// 向所有乘客发送消息
	for userID, conn := range m.riders {
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			log.Printf("向乘客发送消息失败: %s, 错误: %v", userID, err)
			// 不移除连接，让重连机制处理
		}
	}
	
	return nil
}

// BroadcastToDrivers 向司机广播消息
func (m *WebSocketManager) BroadcastToDrivers(message contracts.WSMessage) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}
	
	// 向所有司机发送消息
	for driverID, conn := range m.drivers {
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			log.Printf("向司机发送消息失败: %s, 错误: %v", driverID, err)
			// 不移除连接，让重连机制处理
		}
	}
	
	return nil
}

// SendToRider 向特定乘客发送消息
func (m *WebSocketManager) SendToRider(userID string, message contracts.WSMessage) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	conn, exists := m.riders[userID]
	if !exists {
		log.Printf("乘客连接不存在: %s", userID)
		return nil
	}
	
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}
	
	if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
		log.Printf("向乘客发送消息失败: %s, 错误: %v", userID, err)
		// 不移除连接，让重连机制处理
	}
	
	return nil
}

// SendToDriver 向特定司机发送消息
func (m *WebSocketManager) SendToDriver(driverID string, message contracts.WSMessage) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	conn, exists := m.drivers[driverID]
	if !exists {
		log.Printf("司机连接不存在: %s", driverID)
		return nil
	}
	
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}
	
	if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
		log.Printf("向司机发送消息失败: %s, 错误: %v", driverID, err)
		// 不移除连接，让重连机制处理
	}
	
	return nil
}

// GetConnectedRidersCount 获取已连接的乘客数量
func (m *WebSocketManager) GetConnectedRidersCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.riders)
}

// GetConnectedDriversCount 获取已连接的司机数量
func (m *WebSocketManager) GetConnectedDriversCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.drivers)
}

// Close 关闭所有连接
func (m *WebSocketManager) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// 关闭所有乘客连接
	for userID, conn := range m.riders {
		conn.Close()
		log.Printf("乘客连接已关闭: %s", userID)
	}
	
	// 关闭所有司机连接
	for driverID, conn := range m.drivers {
		conn.Close()
		log.Printf("司机连接已关闭: %s", driverID)
	}
	
	// 清空连接映射
	m.riders = make(map[string]*websocket.Conn)
	m.drivers = make(map[string]*websocket.Conn)
}