// connection/topic.go
package connection

import (
	"log"
	"sync"
)

// TopicManager handles topic subscriptions and message routing
type TopicManager struct {
	// Map of topic to map of client IDs
	subscriptions sync.Map // map[string]map[string]bool
	mu            sync.RWMutex
}

func NewTopicManager() *TopicManager {
	return &TopicManager{}
}

// Subscribe adds a client to a topic
func (tm *TopicManager) Subscribe(topic, clientID string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	// Get or create subscribers map for this topic
	value, _ := tm.subscriptions.LoadOrStore(topic, make(map[string]bool))
	subscribers := value.(map[string]bool)
	subscribers[clientID] = true

	log.Printf("Client %s subscribed to topic: %s", clientID, topic)
}

// Unsubscribe removes a client from a topic
func (tm *TopicManager) Unsubscribe(topic, clientID string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if value, ok := tm.subscriptions.Load(topic); ok {
		subscribers := value.(map[string]bool)
		delete(subscribers, clientID)

		// If no more subscribers, remove the topic
		if len(subscribers) == 0 {
			tm.subscriptions.Delete(topic)
		}

		log.Printf("Client %s unsubscribed from topic: %s", clientID, topic)
	}
}

// UnsubscribeAll removes a client from all topics
func (tm *TopicManager) UnsubscribeAll(clientID string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tm.subscriptions.Range(func(key, value interface{}) bool {
		topic := key.(string)
		subscribers := value.(map[string]bool)
		if subscribers[clientID] {
			delete(subscribers, clientID)
			if len(subscribers) == 0 {
				tm.subscriptions.Delete(topic)
			}
			log.Printf("Client %s unsubscribed from topic: %s", clientID, topic)
		}
		return true
	})
}

// GetSubscribers returns all client IDs subscribed to a topic
func (tm *TopicManager) GetSubscribers(topic string) []string {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	if value, ok := tm.subscriptions.Load(topic); ok {
		subscribers := value.(map[string]bool)
		clientIDs := make([]string, 0, len(subscribers))
		for clientID := range subscribers {
			clientIDs = append(clientIDs, clientID)
		}
		return clientIDs
	}
	return nil
}

// GetTopics returns all active topics
func (tm *TopicManager) GetTopics() []string {
	var topics []string
	tm.subscriptions.Range(func(key, _ interface{}) bool {
		topics = append(topics, key.(string))
		return true
	})
	return topics
}
