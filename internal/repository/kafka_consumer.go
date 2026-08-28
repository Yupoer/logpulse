package repository

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/IBM/sarama"
	"github.com/Yupoer/logpulse/internal/domain"
	"github.com/Yupoer/logpulse/internal/metrics"
)

type KafkaConsumer struct {
	mysqlRepo domain.LogRepository
	esRepo    domain.LogSearchRepository
	metrics   *metrics.Metrics
}

func NewKafkaConsumer(mysqlRepo domain.LogRepository, esRepo domain.LogSearchRepository, appMetrics *metrics.Metrics) *KafkaConsumer {
	return &KafkaConsumer{
		mysqlRepo: mysqlRepo,
		esRepo:    esRepo,
		metrics:   appMetrics,
	}
}

func (c *KafkaConsumer) StartConsumerGroup(ctx context.Context, brokers []string, topic string, groupID string) error {
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRoundRobin()}
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	client, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		log.Printf("Error creating consumer group client: %v", err)
		return err
	}
	defer func() { _ = client.Close() }()

	for {
		// Consume is blocking, but our ConsumeClaim now handles the batch logic
		if err := client.Consume(ctx, []string{topic}, c); err != nil {
			if ctx.Err() != nil {
				return nil
			}
			log.Printf("Error from consumer: %v", err)
			// Small, cancellable backoff to avoid a tight loop on error.
			select {
			case <-time.After(2 * time.Second):
			case <-ctx.Done():
				return nil
			}
		}
		if ctx.Err() != nil {
			return nil
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
		flushCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := c.esRepo.BulkIndex(flushCtx, batch); err != nil {
			log.Printf("Failed to bulk index to ES: %v", err)
			if c.metrics != nil {
				c.metrics.RecordKafkaBatch(len(batch), "es_error")
			}
		} else {
			log.Printf("[Worker] Bulk Indexed %d logs to ES", len(batch))
			if c.metrics != nil {
				c.metrics.RecordKafkaBatch(len(batch), "success")
			}
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
				if c.metrics != nil {
					c.metrics.RecordKafka("unmarshal_error")
				}
				continue
			}

			// 2. Write to MySQL (Sync backup) - 保持逐筆寫入以確保資料安全性 (MVP)
			if err := c.mysqlRepo.Create(session.Context(), &entry); err != nil {
				log.Printf("Failed to save log to DB: %v", err)
				if c.metrics != nil {
					c.metrics.RecordKafka("mysql_error")
				}
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
