package kafkago_test

import (
	"context"
	"log"
	"server/pkg/kfk/kafkago"
	"testing"
)

func TestKafkaProducer(t *testing.T) {
	consumer := kafkago.NewKafkaConsumer(
		[]string{"localhost:9092"},
		"my-group",
		"my-topic",
	)
	defer consumer.Close()

	ctx := context.Background()

	handler := func(key, value []byte) error {
		log.Printf("Processed: %s => %s", key, value)
		return nil // 返回 error 可用于重试
	}

	if err := consumer.Consume(ctx, handler); err != nil {
		log.Fatal(err)
	}
}
