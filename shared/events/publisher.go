package events

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

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
	// 连接到RabbitMQ服务器
	conn, err := amqp091.Dial(amqpURL)
	if err != nil {
		return nil, fmt.Errorf("连接RabbitMQ失败: %w", err)
	}

	// 创建通道
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("创建通道失败: %w", err)
	}

	// 声明交换器
	err = ch.ExchangeDeclare(
		exchange, // 交换器名称
		"topic",  // 交换器类型
		true,     // 持久化
		false,    // 自动删除
		false,    // 内部使用
		false,    // 不等待
		nil,      // 参数
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("声明交换器失败: %w", err)
	}

	publisher := &RabbitMQPublisher{
		conn:    conn,
		channel: ch,
		exchange: exchange,
	}

	// 启动连接监控
	go publisher.monitorConnection()

	return publisher, nil
}

// PublishEvent 发布事件
func (p *RabbitMQPublisher) PublishEvent(eventType string, data interface{}) error {
	// 序列化数据
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("序列化事件数据失败: %w", err)
	}

	// 创建AMQP消息
	message := contracts.AmqpMessage{
		OwnerID: "trip-service", // TODO: 从配置获取服务名称
		Data:    jsonData,
	}

	// 序列化消息
	messageData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("序列化AMQP消息失败: %w", err)
	}

	// 发布消息
	err = p.channel.Publish(
		p.exchange, // 交换器
		eventType,  // 路由键
		false,      // 强制
		false,      // 立即
		amqp091.Publishing{
			ContentType: "application/json",
			Body:        messageData,
			Timestamp:   time.Now(),
		},
	)
	if err != nil {
		return fmt.Errorf("发布事件失败: %w", err)
	}

	log.Printf("成功发布事件: %s", eventType)
	return nil
}

// PublishCommand 发布命令
func (p *RabbitMQPublisher) PublishCommand(commandType string, data interface{}) error {
	// 序列化数据
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("序列化命令数据失败: %w", err)
	}

	// 创建AMQP消息
	message := contracts.AmqpMessage{
		OwnerID: "trip-service", // TODO: 从配置获取服务名称
		Data:    jsonData,
	}

	// 序列化消息
	messageData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("序列化AMQP消息失败: %w", err)
	}

	// 发布消息
	err = p.channel.Publish(
		p.exchange,   // 交换器
		commandType,  // 路由键
		false,        // 强制
		false,        // 立即
		amqp091.Publishing{
			ContentType: "application/json",
			Body:        messageData,
			Timestamp:   time.Now(),
		},
	)
	if err != nil {
		return fmt.Errorf("发布命令失败: %w", err)
	}

	log.Printf("成功发布命令: %s", commandType)
	return nil
}

// Close 关闭发布器
func (p *RabbitMQPublisher) Close() error {
	if p.channel != nil {
		p.channel.Close()
	}
	if p.conn != nil {
		p.conn.Close()
	}
	return nil
}

// monitorConnection 监控连接状态
func (p *RabbitMQPublisher) monitorConnection() {
	errChan := make(chan *amqp091.Error)
	p.channel.NotifyClose(errChan)

	err := <-errChan
	if err != nil {
		log.Printf("RabbitMQ连接关闭: %v", err)
		// 尝试重连
		go p.reconnect()
	}
}

// reconnect 尝试重新连接RabbitMQ
func (p *RabbitMQPublisher) reconnect() {
	log.Printf("尝试重新连接RabbitMQ...")
	
	// 需要从环境变量或配置获取URL
	amqpURL := "amqp://guest:guest@localhost:5672/" // 默认URL
	
	for i := 0; i < 5; i++ {
		// 等待一段时间后重试
		time.Sleep(time.Duration(i+1) * time.Second)
		
		// 尝试重新连接
		conn, err := amqp091.Dial(amqpURL)
		if err != nil {
			log.Printf("重连失败(尝试 %d/5): %v", i+1, err)
			continue
		}
		
		// 创建通道
		ch, err := conn.Channel()
		if err != nil {
			conn.Close()
			log.Printf("创建通道失败(尝试 %d/5): %v", i+1, err)
			continue
		}
		
		// 声明交换器
		err = ch.ExchangeDeclare(
			p.exchange, // 交换器名称
			"topic",    // 交换器类型
			true,       // 持久化
			false,      // 自动删除
			false,      // 内部使用
			false,      // 不等待
			nil,        // 参数
		)
		if err != nil {
			ch.Close()
			conn.Close()
			log.Printf("声明交换器失败(尝试 %d/5): %v", i+1, err)
			continue
		}
		
		// 更新连接
		p.conn.Close()
		p.channel.Close()
		p.conn = conn
		p.channel = ch
		
		log.Printf("RabbitMQ重连成功")
		go p.monitorConnection()
		return
	}
	
	log.Printf("RabbitMQ重连失败，已达到最大重试次数")
}