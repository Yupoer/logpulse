package repository

import (
	"context"
	"encoding/json"
	"log"

	"github.com/IBM/sarama"
	"github.com/Yupoer/logpulse/internal/domain"
)

type kafkaProducer struct {
	producer sarama.SyncProducer
	topic    string
}

// NewKafkaProducer initializes a new Sarama SyncProducer.
func NewKafkaProducer(brokers []string, topic string) (domain.LogProducer, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true          // Must be true for SyncProducer
	config.Producer.RequiredAcks = sarama.WaitForAll // Strongest consistency guarantee
	config.Producer.Retry.Max = 5                    // Retry up to 5 times on failure

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, err
	}

	return &kafkaProducer{
		producer: producer,
		topic:    topic,
	}, nil
}

func (p *kafkaProducer) SendLog(ctx context.Context, entry *domain.LogEntry) error {
	// 1. Serialize struct to JSON
	bytes, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	// 2. Build Kafka Message
	msg := &sarama.ProducerMessage{
		Topic: p.topic,
		// Using ServiceName as Key ensures logs from the same service go to the same partition (Ordering Guarantee)
		Key:   sarama.StringEncoder(entry.ServiceName),
		Value: sarama.ByteEncoder(bytes),
	}

	// 3. Send Message
	partition, offset, err := p.producer.SendMessage(msg)
	if err != nil {
		return err
	}

	log.Printf("Message sent to partition %d at offset %d", partition, offset)
	return nil
}

func (p *kafkaProducer) Close() error {
	return p.producer.Close()
}
