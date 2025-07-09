package shared

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/sirupsen/logrus"
)

// CloudPubSubClient wraps Google Cloud Pub/Sub functionality
type CloudPubSubClient struct {
	client    *pubsub.Client
	projectID string
	mutex     sync.RWMutex
	topics    map[string]*pubsub.Topic
	subs      map[string]*pubsub.Subscription
}

// CloudMessage represents a message for cloud pub/sub
type CloudMessage struct {
	ID        string                 `json:"id"`
	Topic     string                 `json:"topic"`
	Payload   map[string]interface{} `json:"payload"`
	Timestamp time.Time              `json:"timestamp"`
	Source    string                 `json:"source"`
}

// CloudSubscriberFunc is a function that handles cloud pub/sub messages
type CloudSubscriberFunc func(ctx context.Context, msg *CloudMessage) error

var (
	globalCloudPubSub *CloudPubSubClient
	initOnce          sync.Once
)

// InitCloudPubSub initializes the global Cloud Pub/Sub client
func InitCloudPubSub() error {
	var err error
	initOnce.Do(func() {
		projectID := os.Getenv("PUBSUB_PROJECT_ID")
		if projectID == "" {
			err = fmt.Errorf("PUBSUB_PROJECT_ID environment variable is required")
			return
		}

		ctx := context.Background()
		client, clientErr := pubsub.NewClient(ctx, projectID)
		if clientErr != nil {
			err = fmt.Errorf("failed to create pub/sub client: %w", clientErr)
			return
		}

		globalCloudPubSub = &CloudPubSubClient{
			client:    client,
			projectID: projectID,
			topics:    make(map[string]*pubsub.Topic),
			subs:      make(map[string]*pubsub.Subscription),
		}

		logrus.WithField("project_id", projectID).Info("Cloud Pub/Sub client initialized")
	})
	return err
}

// GetCloudPubSub returns the global Cloud Pub/Sub client
func GetCloudPubSub() *CloudPubSubClient {
	if globalCloudPubSub == nil {
		logrus.Fatal("Cloud Pub/Sub not initialized. Call InitCloudPubSub() first")
	}
	return globalCloudPubSub
}

// CreateTopicIfNotExists creates a topic if it doesn't exist
func (c *CloudPubSubClient) CreateTopicIfNotExists(ctx context.Context, topicID string) (*pubsub.Topic, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if topic, exists := c.topics[topicID]; exists {
		return topic, nil
	}

	topic := c.client.Topic(topicID)
	exists, err := topic.Exists(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check topic existence: %w", err)
	}

	if !exists {
		topic, err = c.client.CreateTopic(ctx, topicID)
		if err != nil {
			return nil, fmt.Errorf("failed to create topic %s: %w", topicID, err)
		}
		logrus.WithField("topic", topicID).Info("Created new pub/sub topic")
	}

	c.topics[topicID] = topic
	return topic, nil
}

// CreateSubscriptionIfNotExists creates a subscription if it doesn't exist
func (c *CloudPubSubClient) CreateSubscriptionIfNotExists(ctx context.Context, topicID, subscriptionID string) (*pubsub.Subscription, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	subKey := fmt.Sprintf("%s:%s", topicID, subscriptionID)
	if sub, exists := c.subs[subKey]; exists {
		return sub, nil
	}

	topic, err := c.CreateTopicIfNotExists(ctx, topicID)
	if err != nil {
		return nil, err
	}

	sub := c.client.Subscription(subscriptionID)
	exists, err := sub.Exists(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check subscription existence: %w", err)
	}

	if !exists {
		sub, err = c.client.CreateSubscription(ctx, subscriptionID, pubsub.SubscriptionConfig{
			Topic:       topic,
			AckDeadline: 10 * time.Second,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create subscription %s: %w", subscriptionID, err)
		}
		logrus.WithFields(logrus.Fields{
			"topic":        topicID,
			"subscription": subscriptionID,
		}).Info("Created new pub/sub subscription")
	}

	c.subs[subKey] = sub
	return sub, nil
}

// PublishMessage publishes a message to a topic
func (c *CloudPubSubClient) PublishMessage(ctx context.Context, topicID string, payload map[string]interface{}, source string) error {
	topic, err := c.CreateTopicIfNotExists(ctx, topicID)
	if err != nil {
		return err
	}

	message := CloudMessage{
		ID:        fmt.Sprintf("%s-%d", source, time.Now().UnixNano()),
		Topic:     topicID,
		Payload:   payload,
		Timestamp: time.Now(),
		Source:    source,
	}

	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	result := topic.Publish(ctx, &pubsub.Message{
		Data: data,
		Attributes: map[string]string{
			"source":    source,
			"topic":     topicID,
			"timestamp": message.Timestamp.Format(time.RFC3339),
		},
	})

	_, err = result.Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"topic":  topicID,
		"source": source,
		"id":     message.ID,
	}).Info("Message published to cloud pub/sub")

	return nil
}

// SubscribeToTopic subscribes to a topic with a given subscription ID
func (c *CloudPubSubClient) SubscribeToTopic(ctx context.Context, topicID, subscriptionID string, handler CloudSubscriberFunc) error {
	sub, err := c.CreateSubscriptionIfNotExists(ctx, topicID, subscriptionID)
	if err != nil {
		return err
	}

	logrus.WithFields(logrus.Fields{
		"topic":        topicID,
		"subscription": subscriptionID,
	}).Info("Starting to receive messages")

	return sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		var cloudMsg CloudMessage
		if err := json.Unmarshal(msg.Data, &cloudMsg); err != nil {
			logrus.WithError(err).Error("Failed to unmarshal cloud message")
			msg.Nack()
			return
		}

		if err := handler(ctx, &cloudMsg); err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"topic": cloudMsg.Topic,
				"id":    cloudMsg.ID,
			}).Error("Handler failed to process message")
			msg.Nack()
			return
		}

		msg.Ack()
		logrus.WithFields(logrus.Fields{
			"topic": cloudMsg.Topic,
			"id":    cloudMsg.ID,
		}).Debug("Message processed successfully")
	})
}

// Close closes the Cloud Pub/Sub client
func (c *CloudPubSubClient) Close() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Stop all topics
	for _, topic := range c.topics {
		topic.Stop()
	}

	// Close the client
	return c.client.Close()
}

// Convenience functions for common operations

// PublishOrderEvent publishes an order-related event
func PublishOrderEvent(ctx context.Context, eventType string, orderData map[string]interface{}, source string) error {
	client := GetCloudPubSub()
	topicID := os.Getenv("ORDERS_TOPIC")
	if topicID == "" {
		topicID = "orders-events" // default topic
	}

	payload := map[string]interface{}{
		"event_type": eventType,
		"order_data": orderData,
	}

	return client.PublishMessage(ctx, topicID, payload, source)
}

// PublishPaymentEvent publishes a payment-related event
func PublishPaymentEvent(ctx context.Context, eventType string, paymentData map[string]interface{}, source string) error {
	client := GetCloudPubSub()
	topicID := os.Getenv("PAYMENTS_TOPIC")
	if topicID == "" {
		topicID = "payments-events" // default topic
	}

	payload := map[string]interface{}{
		"event_type":   eventType,
		"payment_data": paymentData,
	}

	return client.PublishMessage(ctx, topicID, payload, source)
}

// SubscribeToOrderEvents subscribes to order events
func SubscribeToOrderEvents(ctx context.Context, subscriptionID string, handler CloudSubscriberFunc) error {
	client := GetCloudPubSub()
	topicID := os.Getenv("ORDERS_TOPIC")
	if topicID == "" {
		topicID = "orders-events"
	}

	return client.SubscribeToTopic(ctx, topicID, subscriptionID, handler)
}

// SubscribeToPaymentEvents subscribes to payment events
func SubscribeToPaymentEvents(ctx context.Context, subscriptionID string, handler CloudSubscriberFunc) error {
	client := GetCloudPubSub()
	topicID := os.Getenv("PAYMENTS_TOPIC")
	if topicID == "" {
		topicID = "payments-events"
	}

	return client.SubscribeToTopic(ctx, topicID, subscriptionID, handler)
}
