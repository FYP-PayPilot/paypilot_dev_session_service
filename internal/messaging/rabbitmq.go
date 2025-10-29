package messaging

import (
	"context"
	"fmt"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"github.com/villageFlower/paypilot_dev_session_service/pkg/config"
	"go.uber.org/zap"
)

// RabbitMQ represents a RabbitMQ connection
type RabbitMQ struct {
	conn    *amqp091.Connection
	channel *amqp091.Channel
	config  *config.RabbitMQConfig
	log     *zap.Logger
}

// NewRabbitMQ creates a new RabbitMQ instance
func NewRabbitMQ(cfg *config.RabbitMQConfig, log *zap.Logger) (*RabbitMQ, error) {
	url := cfg.GetRabbitMQURL()

	conn, err := amqp091.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	rmq := &RabbitMQ{
		conn:    conn,
		channel: channel,
		config:  cfg,
		log:     log,
	}

	// Declare exchange
	err = channel.ExchangeDeclare(
		cfg.Exchange, // name
		"topic",      // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		rmq.Close()
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	// Declare queue
	_, err = channel.QueueDeclare(
		cfg.Queue, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		rmq.Close()
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	// Bind queue to exchange
	err = channel.QueueBind(
		cfg.Queue,    // queue name
		"#",          // routing key
		cfg.Exchange, // exchange
		false,
		nil,
	)
	if err != nil {
		rmq.Close()
		return nil, fmt.Errorf("failed to bind queue: %w", err)
	}

	log.Info("RabbitMQ connection established",
		zap.String("exchange", cfg.Exchange),
		zap.String("queue", cfg.Queue))

	return rmq, nil
}

// Publish publishes a message to the exchange
func (r *RabbitMQ) Publish(ctx context.Context, routingKey string, body []byte) error {
	err := r.channel.PublishWithContext(
		ctx,
		r.config.Exchange, // exchange
		routingKey,        // routing key
		false,             // mandatory
		false,             // immediate
		amqp091.Publishing{
			ContentType: "application/json",
			Body:        body,
			Timestamp:   time.Now(),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	r.log.Debug("Message published",
		zap.String("routing_key", routingKey),
		zap.Int("size", len(body)))

	return nil
}

// Consume starts consuming messages from the queue
func (r *RabbitMQ) Consume(ctx context.Context, handler func([]byte) error) error {
	msgs, err := r.channel.Consume(
		r.config.Queue, // queue
		"",             // consumer
		false,          // auto-ack
		false,          // exclusive
		false,          // no-local
		false,          // no-wait
		nil,            // args
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	r.log.Info("Started consuming messages from queue", zap.String("queue", r.config.Queue))

	for {
		select {
		case <-ctx.Done():
			r.log.Info("Stopping message consumer")
			return ctx.Err()
		case msg, ok := <-msgs:
			if !ok {
				return fmt.Errorf("message channel closed")
			}

			r.log.Debug("Received message", zap.Int("size", len(msg.Body)))

			if err := handler(msg.Body); err != nil {
				r.log.Error("Failed to handle message", zap.Error(err))
				msg.Nack(false, true) // Requeue message
			} else {
				msg.Ack(false)
			}
		}
	}
}

// Close closes the RabbitMQ connection
func (r *RabbitMQ) Close() error {
	if r.channel != nil {
		if err := r.channel.Close(); err != nil {
			return err
		}
	}
	if r.conn != nil {
		if err := r.conn.Close(); err != nil {
			return err
		}
	}
	r.log.Info("RabbitMQ connection closed")
	return nil
}
