package kafka

import (
	"testing"
)

func TestNewProducer(t *testing.T) {
	brokers := []string{"localhost:9092"}
	writer := NewProducer(brokers, "test-topic")
	if writer == nil {
		t.Fatal("NewProducer returned nil")
	}
	// Проверяем базовые поля
	if writer.Topic != "test-topic" {
		t.Errorf("expected topic 'test-topic', got %s", writer.Topic)
	}
	if len(writer.Addr.String()) == 0 {
		t.Errorf("expected non-empty address")
	}
	writer.Close()
}

func TestSendMessage(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when writer is nil")
		}
	}()
	_ = SendMessage(nil, "test")
}
