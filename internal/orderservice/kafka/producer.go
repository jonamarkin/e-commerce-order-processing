package kafka

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

type Producer struct {
	writer *kafka.Writer
}

// NewProducer creates a new Kafka producer.
// brokers should be a slice of "host:port" strings.
// topic is the default topic to write to.
func NewProducer(brokers []string, topic string) *Producer {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},          // Distribute messages among partitions
		RequiredAcks: kafka.RequiredAcks(1),        // Wait for leader acknowledgement
		MaxAttempts:  3,                            // Retry up to 3 times
		WriteTimeout: 5 * time.Second,              // Timeout for write operations
		BatchTimeout: 1 * time.Second,              // Max time before a batch is sent
		BatchSize:    100,                          // Max number of messages in a batch
		Logger:       kafka.LoggerFunc(log.Printf), // Integrate with standard logger
		ErrorLogger:  kafka.LoggerFunc(log.Printf), // Integrate with standard error logger
	}
	return &Producer{writer: writer}
}

// PublishMessage sends a key-value message to the Kafka topic.
func (p *Producer) PublishMessage(ctx context.Context, key, value []byte) error {
	msg := kafka.Message{
		Key:   key,
		Value: value,
		Time:  time.Now(),
	}

	err := p.writer.WriteMessages(ctx, msg)
	if err != nil {
		return fmt.Errorf("failed to write message to Kafka: %w", err)
	}
	return nil
}

// Close closes the Kafka producer connection.
func (p *Producer) Close() error {
	log.Println("Closing Kafka producer...")
	return p.writer.Close()
}
