package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/segmentio/kafka-go"
)

type KafkaProducer interface {
	PublishMessage(ctx context.Context, key, value []byte) error
	Close() error
}

type Producer struct {
	writer *kafka.Writer
}

func NewProducer(brokers []string, topic string) *Producer {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequiredAcks(1),
		MaxAttempts:  3,
		WriteTimeout: 5 * time.Second,
		BatchTimeout: 1 * time.Second,
		BatchSize:    100,
		Logger:       kafka.LoggerFunc(log.Printf),
		ErrorLogger:  kafka.LoggerFunc(log.Printf),
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
	log.Info().Msg("Closing Kafka producer...")
	return p.writer.Close()
}
