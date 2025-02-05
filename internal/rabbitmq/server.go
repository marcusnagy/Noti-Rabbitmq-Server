package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/marcus.jonathan/noti-rabbitmq/internal/models"
)

type RabbitMQServer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

// NotificationHandler is a function that processes metadata and returns a channel of notifications
type NotificationHandler func(delivery amqp.Delivery) (<-chan Notification, error)

func NewRabbitMQServer(url string) (*RabbitMQServer, error) {
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

	// Declare the fanout exchange
	err = ch.ExchangeDeclare(
		FanoutExchange, // name
		"fanout",       // type
		true,           // durable
		false,          // auto-deleted
		false,          // internal
		false,          // no-wait
		nil,            // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	return &RabbitMQServer{
		conn:    conn,
		channel: ch,
	}, nil
}

func (s *RabbitMQServer) Close() {
	if s.channel != nil {
		s.channel.Close()
	}
	if s.conn != nil {
		s.conn.Close()
	}
}

func (s *RabbitMQServer) PublishNotification(ctx context.Context, notification models.Notification, routingKey string) error {
	body, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("failed to marshal notification: %w", err)
	}

	return s.channel.PublishWithContext(
		ctx,
		FanoutExchange,
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent, // make the message persistent
			Body:         body,
		},
	)
}

// ConsumeRPC consumes messages and processes them with the handler
func (s *RabbitMQServer) ConsumeRPC(queueName, routingKey string, handler NotificationHandler) error {
	// Declare a queue
	q, err := s.channel.QueueDeclare(
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

	// Bind the queue to the exchange
	err = s.channel.QueueBind(
		q.Name,      // queue name
		routingKey,  // routing key
		CarExchange, // exchange
		false,       // no-wait
		nil,         // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to bind queue: %w", err)
	}

	// Set QoS
	err = s.channel.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	msgs, err := s.channel.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack is disabled, meaning you need to ack manually
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		return fmt.Errorf("failed to consume queue: %w", err)
	}

	go func() {
		for d := range msgs {
			// Create a new context for each message with a timeout
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

			go func(delivery amqp.Delivery) {
				defer cancel() // Ensure context cancellation after processing

				// Process the message with your custom handler
				notificationChan, err := handler(delivery)
				if err != nil {
					log.Printf("Error handling message: %v", err)
					// Optionally, send a negative ack to requeue the message.
					if nackErr := delivery.Nack(false, true); nackErr != nil {
						log.Printf("Failed to nack message: %v", nackErr)
					}
					return
				}

				processedSuccessfully := true

				// Process notifications that your handler returns
				for {
					select {
					case notification, ok := <-notificationChan:
						if !ok {
							// Notification channel is closed: processing complete
							goto finish
						}
						data, err := json.Marshal(notification)
						if err != nil {
							log.Printf("Failed to marshal notification: %v", err)
							processedSuccessfully = false
							goto finish
						}

						if err := s.channel.PublishWithContext(ctx,
							"",               // Using default exchange here
							delivery.ReplyTo, // routing key
							false,            // mandatory
							false,            // immediate
							amqp.Publishing{
								ContentType:   "application/json",
								CorrelationId: delivery.CorrelationId,
								Body:          data,
							}); err != nil {
							log.Printf("Error publishing notification: %v", err)
							processedSuccessfully = false
							goto finish
						} else {
							log.Printf("Successfully published notification to %s", delivery.ReplyTo)
						}

					case <-ctx.Done():
						log.Printf("Context timeout while processing notifications")
						processedSuccessfully = false
						goto finish
					}
				}

			finish:
				// Acknowledge or negatively acknowledge the message based on success
				if processedSuccessfully {
					if err := delivery.Ack(false); err != nil {
						log.Printf("Failed to ack message: %v", err)
					}
				} else {
					if err := delivery.Nack(false, true); err != nil {
						log.Printf("Failed to nack message: %v", err)
					}
				}
			}(d)
		}
	}()

	return nil
}
