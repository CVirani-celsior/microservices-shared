package shared

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"
)

// Message represents a pub/sub message
type Message struct {
	Topic   string      `json:"topic"`
	Payload interface{} `json:"payload"`
}

// Subscriber is a function that handles messages
type Subscriber func(payload []byte) error

// PubSub is a simple in-memory pub/sub system
type PubSub struct {
	subscribers map[string][]Subscriber
	mutex       sync.RWMutex
}

// NewPubSub creates a new PubSub instance
func NewPubSub() *PubSub {
	return &PubSub{
		subscribers: make(map[string][]Subscriber),
	}
}

// Subscribe adds a subscriber to a topic
func (ps *PubSub) Subscribe(topic string, subscriber Subscriber) {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()

	ps.subscribers[topic] = append(ps.subscribers[topic], subscriber)
	logrus.WithField("topic", topic).Info("New subscriber added")
}

// Publish publishes a message to a topic
func (ps *PubSub) Publish(topic string, payload interface{}) error {
	ps.mutex.RLock()
	defer ps.mutex.RUnlock()

	subscribers, exists := ps.subscribers[topic]
	if !exists {
		logrus.WithField("topic", topic).Warn("No subscribers for topic")
		return nil
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	for _, subscriber := range subscribers {
		go func(sub Subscriber) {
			if err := sub(payloadBytes); err != nil {
				logrus.WithError(err).WithField("topic", topic).Error("Subscriber error")
			}
		}(subscriber)
	}

	logrus.WithField("topic", topic).Info("Message published")
	return nil
}

// Global PubSub instance
var GlobalPubSub = NewPubSub()
