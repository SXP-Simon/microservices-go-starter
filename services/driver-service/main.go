package main

import (
	"context"
	grpcserver "google.golang.org/grpc"
	"log"
	"net"
	"os"
	"os/signal"
	"ride-sharing/driver-service/events"
	sharedEvents "ride-sharing/shared/events"
	"syscall"
)

var GrpcAddr = ":9092"

func main() {
	// 创建服务
	svc := NewService()

	// 初始化事件发布器
	eventConfig := sharedEvents.NewTripExchangeConfig()
	publisher, err := sharedEvents.NewRabbitMQPublisher(eventConfig.URL, eventConfig.Exchange)
	if err != nil {
		log.Fatalf("创建事件发布器失败: %v", err)
	}
	defer publisher.Close()

	// 创建Driver事件发布器
	driverEventPublisher := events.NewDriverEventPublisher(publisher)

	// 初始化事件订阅器
	subscriber, err := sharedEvents.NewRabbitMQSubscriber(eventConfig.URL, eventConfig.Exchange)
	if err != nil {
		log.Fatalf("创建事件订阅器失败: %v", err)
	}
	defer subscriber.Close()

	// 创建事件订阅器并订阅事件
	eventSubscriber := events.NewDriverEventSubscriber(subscriber, svc, driverEventPublisher)
	if err := eventSubscriber.SubscribeToTripEvents(context.Background()); err != nil {
		log.Fatalf("订阅行程事件失败: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		<-sigCh
		cancel()
	}()

	lis, err := net.Listen("tcp", GrpcAddr)
	if err != nil {
		log.Fatalf("监听端口失败: %v", err)
	}

	// 启动gRPC服务器
	grpcServer := grpcserver.NewServer()
	NewGrpcHandler(grpcServer, svc)

	log.Printf("启动Driver服务gRPC服务器，端口: %s", lis.Addr().String())

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Printf("gRPC服务启动失败: %v", err)
			cancel()
		}
	}()

	// 等待关闭信号
	<-ctx.Done()

	log.Println("正在关闭服务器...")
	grpcServer.GracefulStop()
}
