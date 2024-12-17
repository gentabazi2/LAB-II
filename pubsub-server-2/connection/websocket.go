// connection/websocket.go
package connection

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// WebSocketHandler implements the Handler interface for WebSocket connections
type WebSocketHandler struct {
	clients  sync.Map
	upgrader websocket.Upgrader
}

func NewWebSocketHandler() *WebSocketHandler {
	return &WebSocketHandler{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // For development
			},
		},
	}
}

func (h *WebSocketHandler) HandleConnection(w http.ResponseWriter, r *http.Request, clientID string) error {
	log.Printf("Handling new WebSocket connection for client: %s", clientID)

	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return fmt.Errorf("websocket upgrade failed: %v", err)
	}

	client := &Client{
		ID:      clientID,
		Channel: make(chan string, 10),
	}

	wsClient := &wsClient{
		Client: client,
		conn:   conn,
	}
	h.clients.Store(clientID, wsClient)
	log.Printf("WebSocket client connected: %s", clientID)

	// Handle incoming messages
	go func() {
		defer h.Disconnect(clientID)
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				log.Printf("WebSocket read error for client %s: %v", clientID, err)
				return
			}
		}
	}()

	// Send messages to client
	go func() {
		for msg := range client.Channel {
			log.Printf("Sending WebSocket message to client %s: %s", clientID, msg)
			err := conn.WriteMessage(websocket.TextMessage, []byte(msg))
			if err != nil {
				log.Printf("WebSocket write error for client %s: %v", clientID, err)
				return
			}
		}
	}()

	return nil
}

func (h *WebSocketHandler) SendMessage(clientID string, message string) error {
	log.Printf("Attempting to send WebSocket message: clientID=%s, message=%s", clientID, message)

	// If clientID is empty, broadcast to all clients
	if clientID == "" {
		var sent bool
		h.clients.Range(func(key, value interface{}) bool {
			if wsClient, ok := value.(*wsClient); ok {
				select {
				case wsClient.Channel <- message:
					log.Printf("Successfully sent message to WebSocket client %s", wsClient.ID)
					sent = true
				default:
					log.Printf("Failed to send message to WebSocket client %s: channel full", wsClient.ID)
				}
			}
			return true
		})
		if !sent {
			return fmt.Errorf("no WebSocket clients received the message")
		}
		return nil
	}

	// Send to specific client
	if clientInt, ok := h.clients.Load(clientID); ok {
		if wsClient, ok := clientInt.(*wsClient); ok {
			select {
			case wsClient.Channel <- message:
				log.Printf("Successfully sent message to WebSocket client %s", clientID)
				return nil
			default:
				return fmt.Errorf("client channel full")
			}
		}
	}
	return fmt.Errorf("WebSocket client not found: %s", clientID)
}

func (h *WebSocketHandler) Disconnect(clientID string) error {
	log.Printf("Disconnecting WebSocket client: %s", clientID)
	if clientInt, ok := h.clients.Load(clientID); ok {
		wsClient := clientInt.(*wsClient)
		close(wsClient.Channel)
		wsClient.conn.Close()
		h.clients.Delete(clientID)
		log.Printf("WebSocket client disconnected and cleaned up: %s", clientID)
		return nil
	}
	return fmt.Errorf("WebSocket client not found: %s", clientID)
}

type wsClient struct {
	*Client
	conn *websocket.Conn
}
