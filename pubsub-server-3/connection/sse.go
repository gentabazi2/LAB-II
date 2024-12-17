// connection/sse.go
package connection

import (
	"fmt"
	"log"
	"net/http"
	"sync"
)

// SSEHandler implements the Handler interface for Server-Sent Events
type SSEHandler struct {
	clients sync.Map
}

func NewSSEHandler() *SSEHandler {
	return &SSEHandler{}
}

func (h *SSEHandler) HandleConnection(w http.ResponseWriter, r *http.Request, clientID string) error {
	log.Printf("New SSE connection request for client: %s", clientID)

	// Set headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Important: flush the headers immediately
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	} else {
		return fmt.Errorf("streaming not supported")
	}

	client := &Client{
		ID:      clientID,
		Channel: make(chan string, 10),
	}

	h.clients.Store(clientID, client)
	log.Printf("SSE client connected: %s", clientID)

	// Send an initial message to confirm connection
	fmt.Fprintf(w, "data: Connected as %s\n\n", clientID)
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}

	// Setup cleanup when connection is closed
	notify := r.Context().Done()
	go func() {
		<-notify
		log.Printf("SSE client disconnected: %s", clientID)
		h.Disconnect(clientID)
	}()

	// Send events to client
	for msg := range client.Channel {
		_, err := fmt.Fprintf(w, "data: %s\n\n", msg)
		if err != nil {
			log.Printf("Error sending message to SSE client %s: %v", clientID, err)
			return err
		}
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
		log.Printf("Sent message to SSE client %s: %s", clientID, msg)
	}

	return nil
}

func (h *SSEHandler) SendMessage(clientID string, message string) error {
	log.Printf("Attempting to send message to SSE client %s: %s", clientID, message)

	// If clientID is empty, broadcast to all clients
	if clientID == "" {
		var sent bool
		h.clients.Range(func(key, value interface{}) bool {
			if client, ok := value.(*Client); ok {
				select {
				case client.Channel <- message:
					log.Printf("Broadcasted message to SSE client %s", client.ID)
					sent = true
				default:
					log.Printf("Failed to send message to SSE client %s: channel full", client.ID)
				}
			}
			return true
		})
		if !sent {
			return fmt.Errorf("no clients received the message")
		}
		return nil
	}

	// Send to specific client
	if clientInt, ok := h.clients.Load(clientID); ok {
		if client, ok := clientInt.(*Client); ok {
			select {
			case client.Channel <- message:
				log.Printf("Successfully sent message to SSE client %s", clientID)
				return nil
			default:
				return fmt.Errorf("client channel full")
			}
		}
	}
	return fmt.Errorf("client not found: %s", clientID)
}

func (h *SSEHandler) Disconnect(clientID string) error {
	if clientInt, ok := h.clients.Load(clientID); ok {
		if client, ok := clientInt.(*Client); ok {
			close(client.Channel)
			h.clients.Delete(clientID)
			log.Printf("SSE client disconnected and cleaned up: %s", clientID)
			return nil
		}
	}
	return fmt.Errorf("client not found: %s", clientID)
}
