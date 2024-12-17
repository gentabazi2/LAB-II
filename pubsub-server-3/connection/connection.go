// connection/connection.go
package connection

import (
	"net/http"
)

// Handler defines the interface for different connection types
type Handler interface {
	HandleConnection(w http.ResponseWriter, r *http.Request, clientID string) error
	SendMessage(clientID string, message string) error
	Disconnect(clientID string) error
}

// Client represents a connected client
type Client struct {
	ID      string
	Channel chan string
}
