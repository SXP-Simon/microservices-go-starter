package events

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/rabbitmq/amqp091-go"
	"ride-sharing/shared/contracts"
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

	subscriber := &RabbitMQSubscriber{
		conn:    conn,
		channel: ch,
		exchange: exchange,
	}

	return subscriber, nil
}

// Subscribe 订阅事件
func (s *RabbitMQSubscriber) Subscribe(queueName, routingKey string, handler func([]byte) error) error {
	// 声明队列
	q, err := s.channel.QueueDeclare(
		queueName, // 队列名称
		true,      // 持久化
		false,     // 自动删除
		false,     // 独占
		false,     // 不等待
		nil,       // 参数
	)
	if err != nil {
		return fmt.Errorf("声明队列失败: %w", err)
	}

	// 绑定队列到交换器
	err = s.channel.QueueBind(
		q.Name,     // 队列名称
		routingKey, // 路由键
		s.exchange, // 交换器名称
		false,      // 不等待
		nil,        // 参数
	)
	if err != nil {
		return fmt.Errorf("绑定队列失败: %w", err)
	}

	// 设置QoS
	err = s.channel.Qos(
		1,     // 预取数量
		0,     // 预取大小
		false, // 全局设置
	)
	if err != nil {
		return fmt.Errorf("设置QoS失败: %w", err)
	}

	// 消费消息
	msgs, err := s.channel.Consume(
		q.Name, // 队列名称
		"",     // 消费者标签
		false,  // 自动确认
		false,  // 独占
		false,  // 不等待
		false,  // 参数
		nil,
	)
	if err != nil {
		return fmt.Errorf("开始消费失败: %w", err)
	}

	log.Printf("成功订阅队列: %s, 路由键: %s", queueName, routingKey)

	// 处理消息
	go func() {
		for msg := range msgs {
			if err := s.handleMessage(msg, handler); err != nil {
				log.Printf("处理消息失败: %v", err)
				// 拒绝消息并重新入队
				msg.Nack(false, true)
			} else {
				// 确认消息
				msg.Ack(false)
			}
		}
		log.Printf("消息消费通道已关闭: %s", queueName)
	}()

	return nil
}

// handleMessage 处理接收到的消息
func (s *RabbitMQSubscriber) handleMessage(msg amqp091.Delivery, handler func([]byte) error) error {
	// 解析AMQP消息
	var amqpMsg contracts.AmqpMessage
	if err := json.Unmarshal(msg.Body, &amqpMsg); err != nil {
		return fmt.Errorf("解析AMQP消息失败: %w", err)
	}

	// 调用处理函数
	if err := handler(amqpMsg.Data); err != nil {
		return fmt.Errorf("消息处理函数执行失败: %w", err)
	}

	log.Printf("成功处理消息: %s", msg.RoutingKey)
	return nil
}

// Close 关闭订阅器
func (s *RabbitMQSubscriber) Close() error {
	if s.channel != nil {
		s.channel.Close()
	}
	if s.conn != nil {
		s.conn.Close()
	}
	return nil
}