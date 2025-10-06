package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"ride-sharing/shared/events"
	"ride-sharing/shared/contracts"
)

// GatewayEventPublisher API网关事件发布器
type GatewayEventPublisher struct {
	publisher events.Publisher
}

// NewGatewayEventPublisher 创建网关事件发布器
func NewGatewayEventPublisher(publisher events.Publisher) *GatewayEventPublisher {
	return &GatewayEventPublisher{
		publisher: publisher,
	}
}

// PublishDriverTripResponse 发布司机行程响应命令
func (p *GatewayEventPublisher) PublishDriverTripResponse(ctx context.Context, response DriverTripResponse) error {
	var commandType string
	if response.Accept {
		commandType = contracts.DriverCmdTripAccept
	} else {
		commandType = contracts.DriverCmdTripDecline
	}
	
	// 发布命令
	err := p.publisher.PublishCommand(commandType, response)
	if err != nil {
		return fmt.Errorf("发布司机行程响应命令失败: %w", err)
	}
	
	action := "接受"
	if !response.Accept {
		action = "拒绝"
	}
	log.Printf("成功发布司机行程响应命令: 司机ID=%s, 行程ID=%s, 操作=%s", response.DriverID, response.TripID, action)
	return nil
}

// PublishDriverLocationUpdate 发布司机位置更新命令
func (p *GatewayEventPublisher) PublishDriverLocationUpdate(ctx context.Context, update DriverLocationUpdate) error {
	// 发布命令
	err := p.publisher.PublishCommand(contracts.DriverCmdLocation, update)
	if err != nil {
		return fmt.Errorf("发布司机位置更新命令失败: %w", err)
	}
	
	log.Printf("成功发布司机位置更新命令: 司机ID=%s", update.DriverID)
	return nil
}

// Close 关闭发布器
func (p *GatewayEventPublisher) Close() error {
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