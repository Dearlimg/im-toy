package kafka

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"im-chat/internal/infra/config"

	"github.com/IBM/sarama"
)

var (
	producer sarama.AsyncProducer
	brokers  []string
)

type Message struct {
	Key   string
	Value interface{}
}

func InitProducer(cfg *config.KafkaConfig) error {
	brokers = cfg.Brokers
	if len(brokers) == 0 {
		return fmt.Errorf("kafka brokers not configured")
	}

	saramaConfig := sarama.NewConfig()
	saramaConfig.Producer.RequiredAcks = sarama.WaitForAll
	saramaConfig.Producer.Retry.Max = 3
	saramaConfig.Producer.Compression = sarama.CompressionSnappy

	var err error
	producer, err = sarama.NewAsyncProducer(brokers, saramaConfig)
	if err != nil {
		return fmt.Errorf("failed to create kafka producer: %w", err)
	}

	go func() {
		for err := range producer.Errors() {
			log.Printf("Kafka producer error: %v", err)
		}
	}()

	go func() {
		for success := range producer.Successes() {
			log.Printf("Message published to topic=%s, partition=%d, offset=%d",
				success.Topic, success.Partition, success.Offset)
		}
	}()

	log.Printf("Kafka producer initialized with brokers: %v", brokers)
	return nil
}

func Close() {
	if producer != nil {
		if err := producer.Close(); err != nil {
			log.Printf("Error closing kafka producer: %v", err)
		}
	}
}

func Publish(topic string, key string, value interface{}) error {
	if producer == nil {
		return fmt.Errorf("kafka producer not initialized")
	}

	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.ByteEncoder(data),
	}

	select {
	case producer.Input() <- msg:
		return nil
	case <-time.After(5 * time.Second):
		return fmt.Errorf("timeout sending message to kafka")
	}
}
