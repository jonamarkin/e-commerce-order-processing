package kafka

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/jonamarkin/e-commerce-order-processing/internal/orderservice/service"
	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	reader *kafka.Reader
}

// NewConsumer creates a new Kafka consumer.
func NewConsumer(brokers []string, topic, groupID string) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        brokers,
		Topic:          topic,
		GroupID:        groupID,         // Consumer group ID
		MinBytes:       10e3,            // 10KB
		MaxBytes:       10e6,            // 10MB
		MaxWait:        1 * time.Second, // Maximum amount of time to wait for new data to come to a partition
		CommitInterval: 1 * time.Second, // Periodically commit offsets
		Logger:         kafka.LoggerFunc(log.Printf),
		ErrorLogger:    kafka.LoggerFunc(log.Printf),
	})
	return &Consumer{reader: reader}
}

// StartConsuming begins consuming messages from Kafka.
func (c *Consumer) StartConsuming(ctx context.Context) {
	log.Printf("Starting Kafka consumer for topic %s, group %s...", c.reader.Config().Topic, c.reader.Config().GroupID)
	for {
		select {
		case <-ctx.Done():
			log.Println("Kafka consumer context cancelled. Shutting down.")
			return
		default:
			msg, err := c.reader.FetchMessage(ctx) // Fetch one message at a time
			if err != nil {
				if ctx.Err() != nil { // Check if context was cancelled
					return // Context cancelled, gracefully exit
				}
				log.Printf("Error fetching message: %v", err)
				time.Sleep(time.Second) // Small backoff before retrying
				continue
			}

			// Simulate processing the message
			var event service.OrderPlacedEvent // Reusing the event struct from order service
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				log.Printf("Error unmarshaling message from topic %s, partition %d, offset %d: %v",
					msg.Topic, msg.Partition, msg.Offset, err)
			} else {
				log.Printf("Inventory Service: Received OrderPlaced event | OrderID: %s, CustomerID: %s, TotalPrice: %.2f",
					event.OrderID, event.CustomerID, event.TotalPrice)
			}

			// Commit the offset only after successful processing
			if err := c.reader.CommitMessages(ctx, msg); err != nil {
				log.Printf("Error committing offset for message from topic %s, partition %d, offset %d: %v",
					msg.Topic, msg.Partition, msg.Offset, err)
			}
		}
	}
}

// Close closes the Kafka consumer connection.
func (c *Consumer) Close() error {
	log.Println("Closing Kafka consumer...")
	return c.reader.Close()
}
