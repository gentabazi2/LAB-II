// matching/matcher.go
package matching

import (
	"strings"
	"sync"
)

type Event struct {
	Type       string            `json:"type"`
	Content    string            `json:"content"`
	Attributes map[string]string `json:"attributes"`
}

type Subscription struct {
	ClientID string
	Types    []string          // Event types to match
	Keywords []string          // Keywords to match in content
	Patterns map[string]string // Attribute patterns to match
}

type Matcher struct {
	subscriptions map[string]*Subscription // clientID -> subscription
	mu            sync.RWMutex
}

func NewMatcher() *Matcher {
	return &Matcher{
		subscriptions: make(map[string]*Subscription),
	}
}

func (m *Matcher) Subscribe(clientID string, types []string, keywords []string, patterns map[string]string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.subscriptions[clientID] = &Subscription{
		ClientID: clientID,
		Types:    types,
		Keywords: keywords,
		Patterns: patterns,
	}
}

func (m *Matcher) Unsubscribe(clientID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.subscriptions, clientID)
}

// Match returns a score between 0 and 1 indicating how well the event matches
func (m *Matcher) matchScore(event *Event, sub *Subscription) float64 {
	// Check event type
	if len(sub.Types) > 0 {
		typeMatch := false
		for _, t := range sub.Types {
			if t == event.Type {
				typeMatch = true
				break
			}
		}
		if !typeMatch {
			return 0
		}
	}

	totalScore := 0.0
	components := 0.0

	// Check for keyword matches in content
	if len(sub.Keywords) > 0 {
		matches := 0
		content := strings.ToLower(event.Content)
		for _, keyword := range sub.Keywords {
			if strings.Contains(content, strings.ToLower(keyword)) {
				matches++
			}
		}
		keywordScore := float64(matches) / float64(len(sub.Keywords))
		totalScore += keywordScore
		components++
	}

	// Check attribute patterns
	if len(sub.Patterns) > 0 {
		matches := 0
		for key, pattern := range sub.Patterns {
			if value, exists := event.Attributes[key]; exists {
				if strings.Contains(strings.ToLower(value), strings.ToLower(pattern)) {
					matches++
				}
			}
		}
		patternScore := float64(matches) / float64(len(sub.Patterns))
		totalScore += patternScore
		components++
	}

	if components == 0 {
		return 1.0 // If no patterns or keywords specified, consider it a match
	}

	return totalScore / components
}

// FindMatches returns clientIDs of matching subscriptions with scores above threshold
func (m *Matcher) FindMatches(event *Event, threshold float64) []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	matches := make([]string, 0)
	for clientID, sub := range m.subscriptions {
		score := m.matchScore(event, sub)
		if score >= threshold {
			matches = append(matches, clientID)
		}
	}
	return matches
}
