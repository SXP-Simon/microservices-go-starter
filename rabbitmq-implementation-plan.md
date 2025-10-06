# RabbitMQ事件处理系统实现计划

## 概述
本文档详细描述了为微服务打车应用实现RabbitMQ事件处理系统的计划，包括事件发布器、订阅器和相关配置。

## 实现组件

### 1. 共享事件处理组件

#### 1.1 事件发布器 (`shared/events/publisher.go`)
```go
package events

import (
	"encoding/json"
	"fmt"
	"log"
	"github.com/rabbitmq/amqp091-go"
	"ride-sharing/shared/contracts"
)

// Publisher 事件发布器接口
type Publisher interface {
	PublishEvent(eventType string, data interface{}) error
	PublishCommand(commandType string, data interface{}) error
	Close() error
}

// RabbitMQPublisher RabbitMQ事件发布器实现
type RabbitMQPublisher struct {
	conn    *amqp091.Connection
	channel *amqp091.Channel
	exchange string
}

// NewRabbitMQPublisher 创建新的RabbitMQ事件发布器
func NewRabbitMQPublisher(amqpURL, exchange string) (*RabbitMQPublisher, error) {
	// 实现连接和交换器声明
}

// PublishEvent 发布事件
func (p *RabbitMQPublisher) PublishEvent(eventType string, data interface{}) error {
	// 实现事件发布逻辑
}

// PublishCommand 发布命令
func (p *RabbitMQPublisher) PublishCommand(commandType string, data interface{}) error {
	// 实现命令发布逻辑
}
```

#### 1.2 事件订阅器 (`shared/events/subscriber.go`)
```go
package events

import (
	"encoding/json"
	"log"
	"github.com/rabbitmq/amqp091-go"
)

// Subscriber 事件订阅器接口
type Subscriber interface {
	Subscribe(queueName, routingKey string, handler func([]byte) error) error
	Close() error
}

// RabbitMQSubscriber RabbitMQ事件订阅器实现
type RabbitMQSubscriber struct {
	conn    *amqp091.Connection
	channel *amqp091.Channel
	exchange string
}

// NewRabbitMQSubscriber 创建新的RabbitMQ事件订阅器
func NewRabbitMQSubscriber(amqpURL, exchange string) (*RabbitMQSubscriber, error) {
	// 实现连接和队列声明
}

// Subscribe 订阅事件
func (s *RabbitMQSubscriber) Subscribe(queueName, routingKey string, handler func([]byte) error) error {
	// 实现订阅逻辑
}
```

#### 1.3 事件处理配置 (`shared/events/config.go`)
```go
package events

import (
	"os"
)

// Config RabbitMQ配置
type Config struct {
	URL      string
	Exchange string
}

// NewConfig 创建新的配置
func NewConfig() *Config {
	return &Config{
		URL:      getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		Exchange: getEnv("RABBITMQ_EXCHANGE", "ride_sharing"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
```

### 2. Trip服务事件集成

#### 2.1 Trip服务事件发布 (`services/trip-service/internal/infrastructure/events/publisher.go`)
```go
package events

import (
	"context"
	"encoding/json"
	"log"
	"ride-sharing/shared/events"
	"ride-sharing/shared/contracts"
)

// TripEventPublisher Trip服务事件发布器
type TripEventPublisher struct {
	publisher events.Publisher
}

// NewTripEventPublisher 创建Trip事件发布器
func NewTripEventPublisher(publisher events.Publisher) *TripEventPublisher {
	return &TripEventPublisher{
		publisher: publisher,
	}
}

// PublishTripCreated 发布行程创建事件
func (p *TripEventPublisher) PublishTripCreated(ctx context.Context, trip *domain.TripModel) error {
	// 实现行程创建事件发布
}

// PublishDriverAssigned 发布司机分配事件
func (p *TripEventPublisher) PublishDriverAssigned(ctx context.Context, trip *domain.TripModel) error {
	// 实现司机分配事件发布
}

// PublishNoDriversFound 发布未找到司机事件
func (p *TripEventPublisher) PublishNoDriversFound(ctx context.Context, tripID string) error {
	// 实现未找到司机事件发布
}
```

#### 2.2 Trip服务事件订阅 (`services/trip-service/internal/infrastructure/events/subscriber.go`)
```go
package events

import (
	"context"
	"encoding/json"
	"log"
	"ride-sharing/shared/events"
	"ride-sharing/shared/contracts"
	"ride-sharing/services/trip-service/internal/service"
)

// TripEventSubscriber Trip服务事件订阅器
type TripEventSubscriber struct {
	subscriber events.Subscriber
	service    domain.TripService
}

// NewTripEventSubscriber 创建Trip事件订阅器
func NewTripEventSubscriber(subscriber events.Subscriber, service domain.TripService) *TripEventSubscriber {
	return &TripEventSubscriber{
		subscriber: subscriber,
		service:    service,
	}
}

// SubscribeToDriverResponses 订阅司机响应
func (s *TripEventSubscriber) SubscribeToDriverResponses(ctx context.Context) error {
	// 实现司机响应订阅
}

// SubscribeToPaymentEvents 订阅支付事件
func (s *TripEventSubscriber) SubscribeToPaymentEvents(ctx context.Context) error {
	// 实现支付事件订阅
}
```

### 3. Driver服务事件集成

#### 3.1 Driver服务事件订阅 (`services/driver-service/events/subscriber.go`)
```go
package events

import (
	"context"
	"encoding/json"
	"log"
	"ride-sharing/shared/events"
	"ride-sharing/shared/contracts"
)

// DriverEventSubscriber Driver服务事件订阅器
type DriverEventSubscriber struct {
	subscriber events.Subscriber
	service    *Service
}

// NewDriverEventSubscriber 创建Driver事件订阅器
func NewDriverEventSubscriber(subscriber events.Subscriber, service *Service) *DriverEventSubscriber {
	return &DriverEventSubscriber{
		subscriber: subscriber,
		service:    service,
	}
}

// SubscribeToTripEvents 订阅行程事件
func (s *DriverEventSubscriber) SubscribeToTripEvents(ctx context.Context) error {
	// 实现行程事件订阅
}

// handleTripCreated 处理行程创建事件
func (s *DriverEventSubscriber) handleTripCreated(data []byte) error {
	// 实现行程创建事件处理
}
```

#### 3.2 Driver服务事件发布 (`services/driver-service/events/publisher.go`)
```go
package events

import (
	"context"
	"encoding/json"
	"log"
	"ride-sharing/shared/events"
	"ride-sharing/shared/contracts"
)

// DriverEventPublisher Driver服务事件发布器
type DriverEventPublisher struct {
	publisher events.Publisher
}

// NewDriverEventPublisher 创建Driver事件发布器
func NewDriverEventPublisher(publisher events.Publisher) *DriverEventPublisher {
	return &DriverEventPublisher{
		publisher: publisher,
	}
}

// PublishDriverResponse 发布司机响应
func (p *DriverEventPublisher) PublishDriverResponse(ctx context.Context, response *DriverTripResponse) error {
	// 实现司机响应发布
}

// PublishDriverLocation 发布司机位置
func (p *DriverEventPublisher) PublishDriverLocation(ctx context.Context, driverID string, location *Location) error {
	// 实现司机位置发布
}
```

### 4. API网关事件集成

#### 4.1 API网关事件订阅 (`services/api-gateway/events/subscriber.go`)
```go
package events

import (
	"context"
	"encoding/json"
	"log"
	"ride-sharing/shared/events"
	"ride-sharing/shared/contracts"
)

// GatewayEventSubscriber API网关事件订阅器
type GatewayEventSubscriber struct {
	subscriber   events.Subscriber
	wsManager    *WebSocketManager
}

// NewGatewayEventSubscriber 创建网关事件订阅器
func NewGatewayEventSubscriber(subscriber events.Subscriber, wsManager *WebSocketManager) *GatewayEventSubscriber {
	return &GatewayEventSubscriber{
		subscriber: subscriber,
		wsManager:  wsManager,
	}
}

// SubscribeToAllEvents 订阅所有事件
func (s *GatewayEventSubscriber) SubscribeToAllEvents(ctx context.Context) error {
	// 实现所有事件订阅
}

// handleTripCreated 处理行程创建事件
func (s *GatewayEventSubscriber) handleTripCreated(data []byte) error {
	// 实现行程创建事件处理，转发到WebSocket
}
```

### 5. WebSocket管理器

#### 5.1 WebSocket管理器 (`services/api-gateway/websocket/manager.go`)
```go
package websocket

import (
	"context"
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
	// 实现乘客连接添加
}

// AddDriverConnection 添加司机连接
func (m *WebSocketManager) AddDriverConnection(driverID string, conn *websocket.Conn) {
	// 实现司机连接添加
}

// BroadcastToRiders 向乘客广播消息
func (m *WebSocketManager) BroadcastToRiders(message contracts.WSMessage) error {
	// 实现乘客消息广播
}

// BroadcastToDrivers 向司机广播消息
func (m *WebSocketManager) BroadcastToDrivers(message contracts.WSMessage) error {
	// 实现司机消息广播
}

// SendToRider 向特定乘客发送消息
func (m *WebSocketManager) SendToRider(userID string, message contracts.WSMessage) error {
	// 实现特定乘客消息发送
}

// SendToDriver 向特定司机发送消息
func (m *WebSocketManager) SendToDriver(driverID string, message contracts.WSMessage) error {
	// 实现特定司机消息发送
}
```

### 6. 支付服务

#### 6.1 支付服务主程序 (`services/payment-service/main.go`)
```go
package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	
	"ride-sharing/shared/events"
	"ride-sharing/services/payment-service/internal/service"
)

var (
	grpcAddr = flag.String("grpc-addr", ":9094", "gRPC server address")
)

func main() {
	flag.Parse()
	
	// 初始化事件发布器和订阅器
	eventConfig := events.NewConfig()
	publisher, err := events.NewRabbitMQPublisher(eventConfig.URL, "payment_exchange")
	if err != nil {
		log.Fatalf("Failed to create event publisher: %v", err)
	}
	defer publisher.Close()
	
	subscriber, err := events.NewRabbitMQSubscriber(eventConfig.URL, "trip_exchange")
	if err != nil {
		log.Fatalf("Failed to create event subscriber: %v", err)
	}
	defer subscriber.Close()
	
	// 创建支付服务
	paymentService := service.NewPaymentService(publisher)
	
	// 启动gRPC服务器
	// ... gRPC服务器初始化代码
	
	// 订阅事件
	eventSubscriber := events.NewPaymentEventSubscriber(subscriber, paymentService)
	if err := eventSubscriber.SubscribeToTripEvents(context.Background()); err != nil {
		log.Fatalf("Failed to subscribe to events: %v", err)
	}
	
	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	
	log.Println("Shutting down payment service...")
}
```

## 实现步骤

1. **创建共享事件处理组件**
   - 实现事件发布器接口和RabbitMQ实现
   - 实现事件订阅器接口和RabbitMQ实现
   - 创建配置管理

2. **集成Trip服务事件处理**
   - 在Trip服务中添加事件发布功能
   - 实现Trip服务的事件订阅功能
   - 修改Trip服务主程序以初始化事件处理

3. **集成Driver服务事件处理**
   - 在Driver服务中添加事件订阅功能
   - 实现司机匹配逻辑
   - 添加司机位置更新功能
   - 修改Driver服务主程序以初始化事件处理

4. **增强API网关功能**
   - 实现WebSocket管理器
   - 添加事件订阅和消息转发功能
   - 修改WebSocket处理程序以支持事件驱动通信

5. **实现支付服务**
   - 创建支付服务基础结构
   - 实现Stripe集成
   - 添加事件发布和订阅功能

6. **更新服务配置**
   - 添加RabbitMQ连接配置
   - 更新Docker配置以包含RabbitMQ
   - 更新Kubernetes部署配置

## 事件流程图

```mermaid
graph TD
    subgraph Exchange[Trip Exchange]
        TE[trip]
    end

    subgraph Queues
        Q1[find_available_drivers]
        Q2[notify_new_trip]
        Q3[notify_driver_assignment]
        Q4[notify_no_drivers_found]
        Q5[driver_cmd_trip_request]
        Q6[driver_trip_response]
    end

    subgraph Events[Event Types]
        E1[trip.event.created]
        E2[trip.event.driver_assigned]
        E3[trip.event.no_drivers_found]
        E4[driver.cmd.trip_request]
        E5[driver.cmd.trip_accept]
        E6[driver.cmd.trip_decline]
    end

    subgraph Services
        TS[Trip Service]
        DS[Driver Service]
        AG[API Gateway]
    end

    %% Event Flow
    E1 --> Q1
    E1 --> Q2
    E2 --> Q3
    E3 --> Q4
    E4 --> Q5
    E5 --> Q6
    E6 --> Q6

    %% Service Interactions
    TS --> TE
    DS --> TE
    AG --> WS

    %% Queue to Service Flow
    Q1 --> DS
    Q2 --> AG
    Q3 --> AG
    Q4 --> AG
    Q5 --> DS
    Q6 --> TS
```

## 下一步行动

1. 切换到Code模式
2. 按照上述计划实现RabbitMQ事件处理系统
3. 从共享事件处理组件开始实现
4. 逐步集成到各个服务中
5. 测试事件流程
6. 提交代码并创建中文提交信息