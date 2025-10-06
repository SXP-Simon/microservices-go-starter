package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"ride-sharing/api-gateway/events"
	"ride-sharing/api-gateway/websocket"
	sharedEvents "ride-sharing/shared/events"
	"syscall"
	"time"

	"ride-sharing/shared/env"
)

var (
	httpAddr = env.GetString("HTTP_ADDR", ":8081")
)

func main() {
	log.Println("启动API网关")

	// 初始化WebSocket管理器
	wsManager = websocket.NewWebSocketManager()
	defer wsManager.Close()

	// 初始化事件订阅器
	eventConfig := sharedEvents.NewTripExchangeConfig()
	subscriber, err := sharedEvents.NewRabbitMQSubscriber(eventConfig.URL, eventConfig.Exchange)
	if err != nil {
		log.Fatalf("创建事件订阅器失败: %v", err)
	}
	defer subscriber.Close()

	// 初始化事件发布器
	publisher, err := sharedEvents.NewRabbitMQPublisher(eventConfig.URL, eventConfig.Exchange)
	if err != nil {
		log.Fatalf("创建事件发布器失败: %v", err)
	}
	defer publisher.Close()

	// 创建事件订阅器并订阅事件
	gatewayEventSubscriber := events.NewGatewayEventSubscriber(subscriber, wsManager)
	if err := gatewayEventSubscriber.SubscribeToAllEvents(context.Background()); err != nil {
		log.Fatalf("订阅事件失败: %v", err)
	}
	
	// 创建API网关事件发布器
	apiEventPublisher := events.NewAPIEventPublisher(publisher)

	mux := http.NewServeMux()

	mux.HandleFunc("POST /trip/preview", enableCORS(HandleTripPreview))
	mux.HandleFunc("POST /trip/start", enableCORS(HandleTripStart))
	mux.HandleFunc("/ws/riders", func(w http.ResponseWriter, r *http.Request) {
		websocket.HandleRidersWebSocket(wsManager, apiEventPublisher, w, r)
	})
	mux.HandleFunc("/ws/drivers", func(w http.ResponseWriter, r *http.Request) {
		websocket.HandleDriversWebSocket(wsManager, apiEventPublisher, w, r)
	})

	server := &http.Server{
		Addr:    httpAddr,
		Handler: mux,
	}

	serverErrors := make(chan error, 1)

	go func() {
		log.Printf("服务器正在监听端口 %s", httpAddr)
		serverErrors <- server.ListenAndServe()
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		log.Printf("服务器启动错误: %v", err)
	case sig := <-shutdown:
		log.Printf("服务器正在关闭，收到信号: %v", sig)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Printf("无法优雅关闭服务器: %v", err)
			server.Close()
		}
	}
}
