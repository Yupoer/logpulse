package repository

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/IBM/sarama"
	"github.com/Yupoer/logpulse/internal/domain"
)

type KafkaConsumer struct {
	mysqlRepo domain.LogRepository
	esRepo    domain.LogSearchRepository // [New] Dependency
}

// Updated Constructor
func NewKafkaConsumer(mysqlRepo domain.LogRepository, esRepo domain.LogSearchRepository) *KafkaConsumer {
	return &KafkaConsumer{
		mysqlRepo: mysqlRepo,
		esRepo:    esRepo,
	}
}

func (c *KafkaConsumer) StartConsumerGroup(ctx context.Context, brokers []string, topic string, groupID string) {
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRoundRobin()}
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	client, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		log.Printf("Error creating consumer group client: %v", err)
		return
	}

	for {
		// Consume is blocking, but our ConsumeClaim now handles the batch logic
		if err := client.Consume(ctx, []string{topic}, c); err != nil {
			log.Printf("Error from consumer: %v", err)
			// Small backoff to avoid tight loop on error
			time.Sleep(2 * time.Second)
		}
		if ctx.Err() != nil {
			return
		}
	}
}

func (c *KafkaConsumer) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (c *KafkaConsumer) Cleanup(sarama.ConsumerGroupSession) error { return nil }

// ConsumeClaim implements the Batch Processing Logic
func (c *KafkaConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	const batchSize = 100
	const flushInterval = 1 * time.Second

	// Buffer to hold logs
	batch := make([]*domain.LogEntry, 0, batchSize)

	// Ticker for time-based flush
	ticker := time.NewTicker(flushInterval)
	defer ticker.Stop()

	// Helper function to flush batch to ES
	flush := func() {
		if len(batch) == 0 {
			return
		}
		// Write to ES
		if err := c.esRepo.BulkIndex(context.Background(), batch); err != nil {
			log.Printf("Failed to bulk index to ES: %v", err)
		} else {
			log.Printf("[Worker] Bulk Indexed %d logs to ES", len(batch))
		}
		// Reset buffer (keep capacity)
		batch = batch[:0]
	}

	for {
		select {
		case msg, ok := <-claim.Messages():
			if !ok {
				flush() // Channel closed, flush remaining
				return nil
			}

			// 1. Unmarshal
			var entry domain.LogEntry
			if err := json.Unmarshal(msg.Value, &entry); err != nil {
				log.Printf("Failed to unmarshal log: %v", err)
				session.MarkMessage(msg, "") // Skip bad message
				continue
			}

			// 2. Write to MySQL (Sync backup) - 保持逐筆寫入以確保資料安全性 (MVP)
			if err := c.mysqlRepo.Create(context.Background(), &entry); err != nil {
				log.Printf("Failed to save log to DB: %v", err)
			} else {
				// Optional: Logging every single insert might be too noisy now
				// log.Printf("[Worker] Saved to MySQL: %s", entry.Message)
			}

			// 3. Add to Batch for ES
			batch = append(batch, &entry)

			// 4. Check Batch Size
			if len(batch) >= batchSize {
				flush()
				// Only mark offset after successful processing?
				// For simplicity, we mark here. ideally should be after flush success.
			}

			session.MarkMessage(msg, "")

		case <-ticker.C:
			// 5. Time Trigger
			flush()

		case <-session.Context().Done():
			// 6. Graceful Shutdown
			flush()
			return nil
		}
	}
}
