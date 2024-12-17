package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"pubsub-server/utils"
	"sync"
)

var connections = struct {
	sync.RWMutex
	clients map[string]chan string
}{
	clients: make(map[string]chan string),
}

func PrintConnections() {
	connections.RLock()
	defer connections.RUnlock()

	if len(connections.clients) == 0 {
		fmt.Println("No active connections")
		return
	}

	fmt.Println("Active connections:")
	for id := range connections.clients {
		fmt.Printf(" - Client ID: %s\n", id)
	}
}

// Publish handles publishing messages
func Publish(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Topic   string `json:"topic"`
		Message string `json:"message"`
	}

	// Parse request body
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		utils.JSONError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	connections.RLock()
	defer connections.RUnlock()

	for id, ch := range connections.clients {
		select {
		case ch <- data.Message:
			fmt.Printf("Sent message to client %s\n", id)
		default:
			fmt.Printf("Client %s not ready to receive message\n", id)
		}
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Message broadcasted to all subscribers")
	// PrintConnections()
}

// Subscribe handles subscribing to a topic
func Subscribe(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Topic string `json:"topic"`
		ID    string `json:"id"`
	}

	// Parse request body
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		utils.JSONError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Subscribe logic (placeholder)
	// You would add the logic for adding a subscriber here.
	utils.JSONResponse(w, http.StatusOK, map[string]string{
		"status":  "success",
		"message": "Subscribed to topic",
	})
}

func SubscribeSSE(w http.ResponseWriter, r *http.Request) {
	// Set headers for SSE
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Expose-Headers", "Content-Type")
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Validate query parameter
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Missing 'id' parameter", http.StatusBadRequest)
		return
	}

	clientChan := make(chan string, 10)

	connections.Lock()
	connections.clients[id] = clientChan
	connections.Unlock()

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	// Listen for client disconnection
	ctx := r.Context()
	go func() {
		<-ctx.Done()
		connections.Lock()
		delete(connections.clients, id)
		connections.Unlock()
		close(clientChan)
		fmt.Printf("Client %s disconnected\n", id)
	}()

	// Send events
	for msg := range clientChan {
		if _, err := fmt.Fprintf(w, "data: %s\n\n", msg); err != nil {
			fmt.Printf("Error sending message to client %s: %v\n", id, err)
			break
		}
		flusher.Flush()
	}
}
