package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"slices"
	"sync"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQClient struct {
	conn            *amqp.Connection
	channel         *amqp.Channel
	notificationQ   string       // Name of client's notification queue
	pendingRequests []string     // Map to track correlation IDs of pending requests
	requestsMutex   sync.RWMutex // Mutex to handle concurrent access to pendingRequests
}

func NewRabbitMQClient(url string) (*RabbitMQClient, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// Declare the topic exchange
	err = ch.ExchangeDeclare(
		CarExchange, // name
		"topic",     // type
		true,        // durable
		false,       // auto-deleted
		false,       // internal
		false,       // no-wait
		nil,         // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	// Create an exclusive notification queue for this client
	q, err := ch.QueueDeclare(
		"",    // name (empty = let server generate name)
		false, // durable
		true,  // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("failed to declare notification queue: %w", err)
	}

	return &RabbitMQClient{
		conn:          conn,
		channel:       ch,
		notificationQ: q.Name,
	}, nil
}

func (c *RabbitMQClient) IsPendingRequest(correlationID string) bool {
	c.requestsMutex.RLock()
	defer c.requestsMutex.RUnlock()
	return slices.Contains(c.pendingRequests, correlationID)
}

// Close closes the RabbitMQ connection and channel
func (c *RabbitMQClient) Close() {
	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		c.conn.Close()
	}
}

func (c *RabbitMQClient) ClearPendingRequests(correlationID string) {
	c.requestsMutex.Lock()
	defer c.requestsMutex.Unlock()
	c.pendingRequests = slices.DeleteFunc(c.pendingRequests, func(id string) bool {
		return id == correlationID
	})
}

// PublishMetadataAsync publishes car metadata asynchronously
func (c *RabbitMQClient) PublishMetadataAsync(ctx context.Context, metadata CarMetadata) error {
	correlationID := uuid.New().String()

	c.requestsMutex.Lock()
	c.pendingRequests = append(c.pendingRequests, correlationID)
	c.requestsMutex.Unlock()

	data, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Publish the message
	err = c.channel.PublishWithContext(ctx,
		CarExchange,      // exchange
		CarMetadataQueue, // routing key
		false,            // mandatory
		false,            // immediate
		amqp.Publishing{
			ContentType:   "application/json",
			CorrelationId: correlationID,
			ReplyTo:       c.notificationQ, // Use our notification queue for replies
			Body:          data,
		})
	if err != nil {
		return fmt.Errorf("failed to publish metadata: %w", err)
	}

	return nil
}

// ConsumeNotifications starts consuming notifications from the client's notification queue
func (c *RabbitMQClient) ConsumeNotifications(handler func([]byte) error) error {
	// Bind the notification queue to the exchange with the notification routing key
	err := c.channel.QueueBind(
		c.notificationQ,                     // queue name
		"car.notification."+c.notificationQ, // routing key
		CarExchange,                         // exchange
		false,                               // no-wait
		nil,                                 // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to bind queue: %w", err)
	}

	msgs, err := c.channel.Consume(
		c.notificationQ, // queue
		"",              // consumer
		true,            // auto-ack
		false,           // exclusive
		false,           // no-local
		false,           // no-wait
		nil,             // args
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	go func() {
		for msg := range msgs {
			if err := handler(msg.Body); err != nil {
				log.Printf("Error handling notification: %v", err)
			}
		}
	}()

	return nil
}

// GetNotificationQueue returns the name of this client's notification queue
func (c *RabbitMQClient) GetNotificationQueue() string {
	return c.notificationQ
}

// ConsumeMessages starts consuming messages from a queue with the given routing key
func (c *RabbitMQClient) ConsumeMessages(queueName, routingKey string, handler func([]byte) error) error {
	q, err := c.channel.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	err = c.channel.QueueBind(
		q.Name,      // queue name
		routingKey,  // routing key
		CarExchange, // exchange
		false,       // no-wait
		nil,         // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to bind queue: %w", err)
	}

	msgs, err := c.channel.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	go func() {
		for msg := range msgs {
			if err := handler(msg.Body); err != nil {
				log.Printf("Error handling message: %v", err)
			}
		}
	}()

	return nil
}
