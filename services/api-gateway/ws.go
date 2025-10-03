package main

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"ride-sharing/services/api-gateway/grpc_clients"
	"ride-sharing/shared/contracts"
	"ride-sharing/shared/proto/driver"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func HandleRidersWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	userID := r.URL.Query().Get("userID")

	if userID == "" {
		log.Println("No user ID provided")
		return
	}

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading message: %v", err)
			break
		}
		log.Printf("Recieved messsage: %s", message)
	}

}

func HandleDriversWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	userID := r.URL.Query().Get("userID")

	if userID == "" {
		log.Println("No user ID provided")
		return
	}
	packageSlug := r.URL.Query().Get("packageSlug")

	if packageSlug == "" {
		log.Println("No user ID provided")
		return
	}

	//type Driver struct {
	//	Id             string `json:"id"`
	//	Name           string `json:"name"`
	//	ProfilePicture string `json:"profilePicture"`
	//	CarPlate       string `json:"carPlate"`
	//	PackageSlug    string `json:"packageSlug"`
	//}

	ctx := r.Context()

	driverService, err := grpc_clients.NewDriverServiceClient()
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		driverService.Client.UnRegisterDriver(ctx, &driver.RegisterDriverRequest{
			DriverID:    userID,
			PackageSlug: packageSlug,
		})
		driverService.Close()
		log.Println("Driver unregistered ", userID)
	}()

	driverData, err := driverService.Client.RegisterDriver(ctx, &driver.RegisterDriverRequest{
		DriverID:    userID,
		PackageSlug: packageSlug,
	})

	if err != nil {

		log.Printf("Error registering driver: %v", err)
		return
	}

	msg := contracts.WSMessage{
		Type: "driver.cmd.register",
		Data: driverData.Driver,
	}

	if err := conn.WriteJSON(msg); err != nil {
		log.Printf("Error sending message: %v", err)
	}

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading message: %v", err)
			break
		}
		log.Printf("Recieved messsage: %s", message)
	}

}
