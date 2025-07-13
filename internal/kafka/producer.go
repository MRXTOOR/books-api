package kafka

import (
	"context"

	"github.com/segmentio/kafka-go"
)

func NewProducer(brokers []string, topic string) *kafka.Writer {
	return kafka.NewWriter(kafka.WriterConfig{
		Brokers: brokers,
		Topic:   topic,
	})
}

func SendMessage(writer *kafka.Writer, value string) error {
	return writer.WriteMessages(context.Background(), kafka.Message{
		Value: []byte(value),
	})
}
