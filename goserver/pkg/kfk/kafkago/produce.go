// produce.go
package kafkago

import (
	"context"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
)

// KafkaProducer 封装生产者
type KafkaProducer struct {
	writer *kafka.Writer
}

// NewKafkaProducer 创建一个新的生产者
func NewKafkaProducer(brokers []string, topic string) *KafkaProducer {
	writer := &kafka.Writer{
		Addr:                   kafka.TCP(brokers...),
		Topic:                  topic,
		Balancer:               &kafka.LeastBytes{}, // 分区选择策略
		WriteTimeout:           10 * time.Second,
		ReadTimeout:            10 * time.Second,
		RequiredAcks:           kafka.RequireAll, // 确保所有副本确认
		AllowAutoTopicCreation: true,             // 允许自动创建主题（生产环境建议关闭）
	}

	return &KafkaProducer{
		writer: writer,
	}
}

// SendMessage 发送一条消息
func (p *KafkaProducer) SendMessage(key, value []byte) error {
	msg := kafka.Message{
		Key:   key,
		Value: value,
		Time:  time.Now(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := p.writer.WriteMessages(ctx, msg)
	if err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	return nil
}

// SendMessages 批量发送多条消息
func (p *KafkaProducer) SendMessages(messages []kafka.Message) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for i := range messages {
		messages[i].Time = time.Now()
	}

	err := p.writer.WriteMessages(ctx, messages...)
	if err != nil {
		return fmt.Errorf("failed to write messages: %w", err)
	}

	return nil
}

// Close 关闭生产者
func (p *KafkaProducer) Close() error {
	return p.writer.Close()
}
