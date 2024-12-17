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

// Manager handles client connections and message distribution
type Manager struct {
	Handlers     map[string]Handler
	TopicManager *TopicManager
}

// NewManager creates a new connection manager
func NewManager() *Manager {
	return &Manager{
		Handlers:     make(map[string]Handler),
		TopicManager: NewTopicManager(),
	}
}

// RegisterHandler adds a new connection handler
func (m *Manager) RegisterHandler(connectionType string, handler Handler) {
	m.Handlers[connectionType] = handler
}

// GetHandler retrieves a connection handler by type
func (m *Manager) GetHandler(connectionType string) (Handler, bool) {
	handler, exists := m.Handlers[connectionType]
	return handler, exists
}

// PublishToTopic publishes a message to all subscribers of a topic
func (m *Manager) PublishToTopic(topic, message string) error {
	subscribers := m.TopicManager.GetSubscribers(topic)
	if len(subscribers) == 0 {
		return nil
	}

	for _, clientID := range subscribers {
		// Try to send to all handlers (client might be connected via different protocols)
		for _, handler := range m.Handlers {
			_ = handler.SendMessage(clientID, message)
		}
	}

	return nil
}
