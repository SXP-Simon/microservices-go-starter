package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"ride-sharing/shared/events"
	sharedContracts "ride-sharing/shared/contracts"
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

// PublishDriverTripRequest 发布司机行程请求命令
func (p *DriverEventPublisher) PublishDriverTripRequest(ctx context.Context, request DriverTripRequest) error {
	// 发布命令
	err := p.publisher.PublishCommand(sharedContracts.DriverCmdTripRequest, request)
	if err != nil {
		return fmt.Errorf("发布司机行程请求命令失败: %w", err)
	}
	
	log.Printf("成功发布司机行程请求命令: 司机ID=%s, 行程ID=%s", request.DriverID, request.TripID)
	return nil
}

// PublishDriverResponse 发布司机响应命令
func (p *DriverEventPublisher) PublishDriverResponse(ctx context.Context, response DriverTripResponse) error {
	var commandType string
	if response.Accept {
		commandType = sharedContracts.DriverCmdTripAccept
	} else {
		commandType = sharedContracts.DriverCmdTripDecline
	}
	
	// 发布命令
	err := p.publisher.PublishCommand(commandType, response)
	if err != nil {
		return fmt.Errorf("发布司机响应命令失败: %w", err)
	}
	
	action := "接受"
	if !response.Accept {
		action = "拒绝"
	}
	log.Printf("成功发布司机响应命令: 司机ID=%s, 行程ID=%s, 操作=%s", response.DriverID, response.TripID, action)
	return nil
}

// PublishDriverLocationUpdate 发布司机位置更新命令
func (p *DriverEventPublisher) PublishDriverLocationUpdate(ctx context.Context, update DriverLocationUpdate) error {
	// 发布命令
	err := p.publisher.PublishCommand(sharedContracts.DriverCmdLocation, update)
	if err != nil {
		return fmt.Errorf("发布司机位置更新命令失败: %w", err)
	}
	
	log.Printf("成功发布司机位置更新命令: 司机ID=%s", update.DriverID)
	return nil
}

// Close 关闭发布器
func (p *DriverEventPublisher) Close() error {
	return p.publisher.Close()
}

// DriverTripResponse 司机行程响应
type DriverTripResponse struct {
	TripID   string `json:"tripID"`
	RiderID  string `json:"riderID"`
	DriverID string `json:"driverID"`
	Accept   bool   `json:"accept"`
}

// DriverLocationUpdate 司机位置更新
type DriverLocationUpdate struct {
	DriverID string  `json:"driverID"`
	Latitude float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Timestamp int64  `json:"timestamp"`
}