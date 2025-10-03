package main

import (
	"encoding/json"
	"log"
	"net/http"
	"ride-sharing/services/api-gateway/grpc_clients"
	"ride-sharing/shared/contracts"
)

func HandleTripStart(w http.ResponseWriter, r *http.Request) {
	var reqBody startTripRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		log.Println(err)
		http.Error(w, "failed to parse JSON data", http.StatusBadRequest)
		return
	}

	defer r.Body.Close()

	tripService, err := grpc_clients.NewTripServiceClient()

	if err != nil {
		log.Fatal(err)
	}

	defer tripService.Close()

	trip, err := tripService.Client.CreateTrip(r.Context(), reqBody.toProto())
	if err != nil {
		log.Printf("Failed to start a trip: %v", err)
		http.Error(w, "Failed to Start trip", http.StatusInternalServerError)
		return
	}

	response := contracts.APIResponse{Data: trip}

	writeJSON(w, http.StatusCreated, response)

}
func HandleTripPreview(w http.ResponseWriter, r *http.Request) {
	var reqBody previewTripRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		log.Println(err)
		http.Error(w, "failed to parse JSON data", http.StatusBadRequest)
		return
	}

	defer r.Body.Close()

	// validation

	if reqBody.UserID == "" {
		http.Error(w, "failed to parse JSON data", http.StatusBadRequest)
		return
	}

	tripService, err := grpc_clients.NewTripServiceClient()

	if err != nil {
		log.Fatal(err)
	}

	defer tripService.Close()

	// TODO: Call trip service
	//resp, err := http.Post("http://trip-service:8083/preview", "application/json", reader)
	//
	//if err != nil {
	//	log.Print(err)
	//	return
	//}
	//
	//defer resp.Body.Close()

	tripPreview, err := tripService.Client.PreviewTrip(r.Context(), reqBody.toProto())

	if err != nil {
		log.Printf("Failed to preveiw a trip: %v", err)
		http.Error(w, "Failed to Preview trip", http.StatusInternalServerError)
		return
	}

	response := contracts.APIResponse{Data: tripPreview}

	writeJSON(w, http.StatusCreated, response)

}
