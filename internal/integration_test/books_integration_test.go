package integration_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/jackc/pgx/v5/pgxpool"
)

func waitForKafka(brokers []string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		config := sarama.NewConfig()
		config.Producer.Return.Successes = true
		producer, err := sarama.NewSyncProducer(brokers, config)
		if err == nil {
			producer.Close()
			return nil
		}
		time.Sleep(2 * time.Second)
	}
	return fmt.Errorf("kafka not ready after %s", timeout)
}

func TestKafkaAndPostgresIntegration(t *testing.T) {
	dsn := os.Getenv("DATABASE_DSN")
	brokers := []string{os.Getenv("KAFKA_BROKERS")}

	// Проверяем подключение к Postgres
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatalf("cannot connect to db: %v", err)
	}
	defer pool.Close()

	// Ждём Kafka
	if err := waitForKafka(brokers, 120*time.Second); err != nil {
		t.Fatalf("cannot connect to kafka: %v", err)
	}

	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		t.Fatalf("cannot connect to kafka: %v", err)
	}
	defer producer.Close()

	// Пример: отправить сообщение в Kafka
	msg := &sarama.ProducerMessage{
		Topic: "test-topic",
		Value: sarama.StringEncoder("hello from integration test"),
	}
	_, _, err = producer.SendMessage(msg)
	if err != nil {
		t.Fatalf("cannot send kafka message: %v", err)
	}

	// Пример: создать запись в БД
	_, err = pool.Exec(context.Background(), "CREATE TABLE IF NOT EXISTS integration_test (id SERIAL PRIMARY KEY, val TEXT)")
	if err != nil {
		t.Fatalf("cannot create table: %v", err)
	}

	// Пример: вставить и прочитать запись
	_, err = pool.Exec(context.Background(), "INSERT INTO integration_test (val) VALUES ($1)", "test-value")
	if err != nil {
		t.Fatalf("cannot insert: %v", err)
	}

	row := pool.QueryRow(context.Background(), "SELECT val FROM integration_test WHERE val=$1", "test-value")
	var val string
	if err := row.Scan(&val); err != nil {
		t.Fatalf("cannot select: %v", err)
	}
	if val != "test-value" {
		t.Fatalf("unexpected value: %s", val)
	}
}
