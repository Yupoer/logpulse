package repository

import (
	"context"
	"encoding/json"
	"log"

	"github.com/IBM/sarama"
	"github.com/Yupoer/logpulse/internal/domain"
)

// KafkaConsumer represents the worker that consumes logs.
type KafkaConsumer struct {
	repo domain.LogRepository // Dependency: Needs to write to MySQL
}

// NewKafkaConsumer creates a new consumer logic instance.
func NewKafkaConsumer(repo domain.LogRepository) *KafkaConsumer {
	return &KafkaConsumer{repo: repo}
}

// StartConsumerGroup starts the infinite loop to consume messages.
// This is a blocking function, so it should be run in a goroutine.
func (c *KafkaConsumer) StartConsumerGroup(ctx context.Context, brokers []string, topic string, groupID string) {
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRoundRobin()}
	config.Consumer.Offsets.Initial = sarama.OffsetOldest // Start from the beginning if no offset is saved

	client, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		log.Printf("Error creating consumer group client: %v", err)
		return
	}

	// Loop to keep consuming even after rebalancing
	for {
		if err := client.Consume(ctx, []string{topic}, c); err != nil {
			log.Printf("Error from consumer: %v", err)
			return
		}
		// check if context was cancelled, signaling that the consumer should stop
		if ctx.Err() != nil {
			return
		}
	}
}

// --- Implementing sarama.ConsumerGroupHandler Interface ---

// Setup is run at the beginning of a new session, before ConsumeClaim.
func (c *KafkaConsumer) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited.
func (c *KafkaConsumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (c *KafkaConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		// 1. Deserialize JSON
		var entry domain.LogEntry
		if err := json.Unmarshal(msg.Value, &entry); err != nil {
			log.Printf("Failed to unmarshal log: %v", err)
			continue
		}

		// 2. Write to MySQL (Idempotency check could be added here)
		// We use context.Background() because the worker lifecycle is independent of the HTTP request
		if err := c.repo.Create(context.Background(), &entry); err != nil {
			log.Printf("Failed to save log to DB: %v", err)
			// Decide: Retry? Dead Letter Queue? For MVP, we just log error.
		} else {
			log.Printf("[Worker] Consumed & Saved: %s", entry.Message)
		}

		// 3. Mark message as processed
		session.MarkMessage(msg, "")
	}
	return nil
}
