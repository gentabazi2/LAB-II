package controllers

import (
	"encoding/json"
	"log"
	"net/http"
	"pubsub-server/connection"
	"pubsub-server/utils"
)

var (
	connectionManager = connection.NewManager()
)

func init() {
	connectionManager.RegisterHandler("sse", connection.NewSSEHandler())
	connectionManager.RegisterHandler("ws", connection.NewWebSocketHandler())
}

// SubscribeRequest represents a subscription request
type SubscribeRequest struct {
	Topics []string `json:"topics"`
	ID     string   `json:"id"`
}

// PublishRequest represents a publish request
type PublishRequest struct {
	Topic   string `json:"topic"`
	Message string `json:"message"`
}

// Subscribe handles topic subscription
func Subscribe(w http.ResponseWriter, r *http.Request) {
	var req SubscribeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSONError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	for _, topic := range req.Topics {
		connectionManager.Subscribe(topic, req.ID)
	}

	utils.JSONResponse(w, http.StatusOK, map[string]interface{}{
		"status":  "success",
		"message": "Subscribed to topics",
		"topics":  req.Topics,
	})
}

// Publish handles publishing messages to topics
func Publish(w http.ResponseWriter, r *http.Request) {
	var req PublishRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSONError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if req.Topic == "" {
		utils.JSONError(w, http.StatusBadRequest, "Topic is required")
		return
	}

	log.Printf("Publishing message to topic %s: %s", req.Topic, req.Message)
	err := connectionManager.PublishToTopic(req.Topic, req.Message)
	if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "Failed to publish message")
		return
	}

	utils.JSONResponse(w, http.StatusOK, map[string]string{
		"status":  "success",
		"message": "Message published to topic",
	})
}

// GetTopics returns all active topics
func GetTopics(w http.ResponseWriter, r *http.Request) {
	topics := connectionManager.GetTopics()
	utils.JSONResponse(w, http.StatusOK, map[string]interface{}{
		"topics": topics,
	})
}

// SubscribeSSE handles SSE connections
func SubscribeSSE(w http.ResponseWriter, r *http.Request) {
	clientID := r.URL.Query().Get("id")
	if clientID == "" {
		utils.JSONError(w, http.StatusBadRequest, "Missing 'id' parameter")
		return
	}

	log.Printf("New SSE subscription request for client: %s", clientID)

	handler, exists := connectionManager.GetHandler("sse")
	if !exists {
		utils.JSONError(w, http.StatusInternalServerError, "SSE handler not found")
		return
	}

	if err := handler.HandleConnection(w, r, clientID); err != nil {
		log.Printf("Error handling SSE connection for client %s: %v", clientID, err)
		utils.JSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
}

// SubscribeWS handles WebSocket connections
func SubscribeWS(w http.ResponseWriter, r *http.Request) {
	clientID := r.URL.Query().Get("id")
	if clientID == "" {
		utils.JSONError(w, http.StatusBadRequest, "Missing 'id' parameter")
		return
	}

	log.Printf("New WebSocket subscription request for client: %s", clientID)

	handler, exists := connectionManager.GetHandler("ws")
	if !exists {
		utils.JSONError(w, http.StatusInternalServerError, "WebSocket handler not found")
		return
	}

	if err := handler.HandleConnection(w, r, clientID); err != nil {
		log.Printf("Error handling WebSocket connection for client %s: %v", clientID, err)
		utils.JSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
}
