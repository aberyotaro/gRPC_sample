package main

import (
	"encoding/json"
	"log"
	"net/http"

	gatewaypb "github.com/aberyotaro/grpc-sample/gen/gateway"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var gatewayClient gatewaypb.GatewayServiceClient

func processHandler(w http.ResponseWriter, r *http.Request) {
	text := r.URL.Query().Get("text")
	if text == "" {
		http.Error(w, `{"error":"text query parameter is required"}`, http.StatusBadRequest)
		return
	}

	res, err := gatewayClient.Process(r.Context(), &gatewaypb.ProcessRequest{Text: text})
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"original":  text,
		"uppercase": res.Uppercase,
		"count":     res.Count,
	})
}

func main() {
	conn, err := grpc.NewClient("gateway:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to gateway: %v", err)
	}
	defer conn.Close()

	gatewayClient = gatewaypb.NewGatewayServiceClient(conn)

	http.HandleFunc("/process", processHandler)
	log.Println("client service listening on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
