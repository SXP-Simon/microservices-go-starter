package main

import (
	"context"
	"flag"
	grpcserver "google.golang.org/grpc"
	"log"
	"net"
	"os"
	"os/signal"
	"ride-sharing/payment-service/internal/infrastructure/events"
	"ride-sharing/payment-service/internal/infrastructure/grpc"
	"ride-sharing/payment-service/internal/infrastructure/repository"
	"ride-sharing/payment-service/internal/service"
	sharedEvents "ride-sharing/shared/events"
	"syscall"
)

var (
	grpcAddr = flag.String("grpc-addr", ":9094", "gRPC服务器地址")
)

func main() {
	flag.Parse()
	
	log.Println("启动支付服务")

	// 初始化事件发布器
	eventConfig := sharedEvents.NewPaymentExchangeConfig()
	publisher, err := sharedEvents.NewRabbitMQPublisher(eventConfig.URL, eventConfig.Exchange)
	if err != nil {
		log.Fatalf("创建事件发布器失败: %v", err)
	}
	defer publisher.Close()

	// 创建支付事件发布器
	paymentEventPublisher := events.NewPaymentEventPublisher(publisher)

	// 初始化事件订阅器
	subscriber, err := sharedEvents.NewRabbitMQSubscriber(eventConfig.URL, "trip") // 订阅trip交换器
	if err != nil {
		log.Fatalf("创建事件订阅器失败: %v", err)
	}
	defer subscriber.Close()

	// 初始化存储库
	repo := repository.NewInmemRepository()

	// 创建支付服务
	paymentService := service.NewPaymentService(repo, paymentEventPublisher)

	// 创建事件订阅器并订阅事件
	eventSubscriber := events.NewPaymentEventSubscriber(subscriber, paymentService)
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

	lis, err := net.Listen("tcp", *grpcAddr)
	if err != nil {
		log.Fatalf("监听端口失败: %v", err)
	}

	// 启动gRPC服务器
	grpcServer := grpcserver.NewServer()
	grpc.NewGRPCHandler(grpcServer, paymentService)

	log.Printf("启动支付服务gRPC服务器，端口: %s", lis.Addr().String())

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