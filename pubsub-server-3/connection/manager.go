// connection/manager.go
package connection

import (
	"encoding/json"
	"log"
	"pubsub-server/matching"
	"sync"
)

type Manager struct {
	Handlers map[string]Handler
	matcher  *matching.Matcher
	topics   map[string]map[string]bool
	mu       sync.RWMutex
}

func NewManager() *Manager {
	return &Manager{
		Handlers: make(map[string]Handler),
		matcher:  matching.NewMatcher(),
		topics:   make(map[string]map[string]bool),
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

// Legacy topic-based methods

func (m *Manager) Subscribe(topic, clientID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.topics[topic]; !exists {
		m.topics[topic] = make(map[string]bool)
	}
	m.topics[topic][clientID] = true
}

func (m *Manager) Unsubscribe(topic, clientID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if clients, exists := m.topics[topic]; exists {
		delete(clients, clientID)
		if len(clients) == 0 {
			delete(m.topics, topic)
		}
	}
}

func (m *Manager) GetSubscribers(topic string) []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if clients, exists := m.topics[topic]; exists {
		subscribers := make([]string, 0, len(clients))
		for clientID := range clients {
			subscribers = append(subscribers, clientID)
		}
		return subscribers
	}
	return nil
}

func (m *Manager) GetTopics() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	topics := make([]string, 0, len(m.topics))
	for topic := range m.topics {
		topics = append(topics, topic)
	}
	return topics
}

func (m *Manager) PublishToTopic(topic, message string) error {
	subscribers := m.GetSubscribers(topic)
	if len(subscribers) == 0 {
		return nil
	}

	for _, clientID := range subscribers {
		for _, handler := range m.Handlers {
			_ = handler.SendMessage(clientID, message)
		}
	}

	log.Printf("Message published to topic %s with %d subscribers", topic, len(subscribers))
	return nil
}

// New pattern-based methods

func (m *Manager) SubscribeWithPatterns(clientID string, types []string, keywords []string, patterns map[string]string) {
	m.matcher.Subscribe(clientID, types, keywords, patterns)
	log.Printf("Client %s subscribed with %d types, %d keywords, %d patterns",
		clientID, len(types), len(keywords), len(patterns))
}

func (m *Manager) PublishEvent(event *matching.Event) {
	matches := m.matcher.FindMatches(event, 0.5)

	for _, clientID := range matches {
		eventJSON, _ := json.Marshal(event)
		for _, handler := range m.Handlers {
			handler.SendMessage(clientID, string(eventJSON))
		}
	}

	log.Printf("Event published to %d matching subscribers", len(matches))
}
