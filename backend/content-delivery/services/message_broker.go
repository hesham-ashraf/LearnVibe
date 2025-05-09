package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/cenkalti/backoff/v4"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sony/gobreaker"
)

// MessageBroker is a service for publishing and consuming messages from RabbitMQ
type MessageBroker struct {
	conn         *amqp.Connection
	ch           *amqp.Channel
	exchangeName string
	queues       map[string]bool
}

// MessagePayload represents the message payload structure
type MessagePayload struct {
	EventType string      `json:"event_type"`
	Data      interface{} `json:"data"`
	UserID    string      `json:"user_id,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// NewMessageBroker creates a new MessageBroker instance
func NewMessageBroker(rabbitmqURL, exchangeName string) (*MessageBroker, error) {
	var conn *amqp.Connection
	var err error

	// Use exponential backoff to retry connection
	operation := func() error {
		conn, err = amqp.Dial(rabbitmqURL)
		if err != nil {
			return fmt.Errorf("failed to connect to RabbitMQ: %v", err)
		}
		return nil
	}

	err = backoff.Retry(operation, backoff.NewExponentialBackOff())
	if err != nil {
		return nil, err
	}

	// Create a circuit breaker for RabbitMQ operations
	settings := gobreaker.Settings{
		Name:    "RabbitMQCircuitBreaker",
		Timeout: 30 * time.Second,
	}
	cb := gobreaker.NewCircuitBreaker(settings)

	var ch *amqp.Channel
	_, err = cb.Execute(func() (interface{}, error) {
		ch, err = conn.Channel()
		if err != nil {
			return nil, fmt.Errorf("failed to open a channel: %v", err)
		}

		// Declare exchange
		err = ch.ExchangeDeclare(
			exchangeName, // name
			"topic",      // type
			true,         // durable
			false,        // auto-deleted
			false,        // internal
			false,        // no-wait
			nil,          // arguments
		)
		if err != nil {
			return nil, fmt.Errorf("failed to declare an exchange: %v", err)
		}
		return nil, nil
	})

	if err != nil {
		return nil, err
	}

	return &MessageBroker{
		conn:         conn,
		ch:           ch,
		exchangeName: exchangeName,
		queues:       make(map[string]bool),
	}, nil
}

// PublishMessage publishes a message to the specified routing key
func (mb *MessageBroker) PublishMessage(ctx context.Context, routingKey string, message interface{}) error {
	// Marshal the message
	payload := MessagePayload{
		EventType: routingKey,
		Data:      message,
		Timestamp: time.Now(),
	}

	jsonMessage, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %v", err)
	}

	// Create a circuit breaker for the publish operation
	settings := gobreaker.Settings{
		Name:    "PublishMessage",
		Timeout: 5 * time.Second,
	}
	cb := gobreaker.NewCircuitBreaker(settings)

	_, err = cb.Execute(func() (interface{}, error) {
		return nil, mb.ch.PublishWithContext(
			ctx,
			mb.exchangeName, // exchange
			routingKey,      // routing key
			false,           // mandatory
			false,           // immediate
			amqp.Publishing{
				ContentType: "application/json",
				Body:        jsonMessage,
			})
	})

	if err != nil {
		return fmt.Errorf("failed to publish message: %v", err)
	}

	log.Printf("Message published to exchange %s with routing key %s", mb.exchangeName, routingKey)
	return nil
}

// ConsumeMessages sets up a consumer for a queue with the given routing key pattern
func (mb *MessageBroker) ConsumeMessages(queueName, routingKeyPattern string, handler func(MessagePayload) error) error {
	// Declare a queue
	q, err := mb.ch.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare a queue: %v", err)
	}

	// Bind the queue to the exchange with the routing key pattern
	err = mb.ch.QueueBind(
		q.Name,            // queue name
		routingKeyPattern, // routing key
		mb.exchangeName,   // exchange
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind a queue: %v", err)
	}

	// Set up the consumer
	msgs, err := mb.ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		return fmt.Errorf("failed to register a consumer: %v", err)
	}

	// Mark this queue as successfully set up
	mb.queues[queueName] = true

	// Start a goroutine to consume messages
	go func() {
		for d := range msgs {
			var payload MessagePayload
			if err := json.Unmarshal(d.Body, &payload); err != nil {
				log.Printf("Error deserializing message: %v", err)
				d.Nack(false, true) // Negative acknowledgement, requeue
				continue
			}

			// Process the message
			err := handler(payload)
			if err != nil {
				log.Printf("Error processing message: %v", err)
				d.Nack(false, true) // Negative acknowledgement, requeue
			} else {
				d.Ack(false) // Acknowledge the message
			}
		}
		log.Printf("Consumer for queue %s has been closed", queueName)
	}()

	log.Printf("Started consuming messages from queue %s with routing key pattern %s", queueName, routingKeyPattern)
	return nil
}

// Close closes the connection to RabbitMQ
func (mb *MessageBroker) Close() error {
	if err := mb.ch.Close(); err != nil {
		return fmt.Errorf("failed to close channel: %v", err)
	}
	if err := mb.conn.Close(); err != nil {
		return fmt.Errorf("failed to close connection: %v", err)
	}
	return nil
}
