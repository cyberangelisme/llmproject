// consume.go
package kafkago

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

// KafkaConsumer 封装消费者
type KafkaConsumer struct {
	reader *kafka.Reader
}

// NewKafkaConsumer 创建一个新的消费者（属于某个消费组）
func NewKafkaConsumer(brokers []string, groupID, topic string) *KafkaConsumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:         brokers,
		GroupID:         groupID,
		Topic:           topic,
		MinBytes:        16,                // 每次拉取最小数据量
		MaxBytes:        10e6,              // 每次拉取最大数据量
		MaxWait:         1 * time.Second,   // 最大等待时间
		ReadLagInterval: -1,                // 不定期报告 lag
		CommitInterval:  1 * time.Second,   // 自动提交 offset 间隔
		StartOffset:     kafka.FirstOffset, // 可设为 kafka.LastOffset 从最新开始
	})

	return &KafkaConsumer{
		reader: reader,
	}
}

// Consume 运行消费者，持续拉取消息
// handler 是处理每条消息的函数 func(key, value []byte) error
func (c *KafkaConsumer) Consume(ctx context.Context, handler func(key, value []byte) error) error {
	for {
		msg, err := c.reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() == context.Canceled || ctx.Err() == context.DeadlineExceeded {
				log.Println("Consumer context cancelled or timeout, exiting...")
				return nil
			}
			return fmt.Errorf("error reading message: %w", err)
		}

		log.Printf("Received: topic=%s partition=%d offset=%d timestamp=%s\n",
			msg.Topic, msg.Partition, msg.Offset, msg.Time.String())

		// 调用业务处理函数
		if err := handler(msg.Key, msg.Value); err != nil {
			log.Printf("Error processing message at offset %d: %v", msg.Offset, err)
			// 可以选择 continue 或重试机制
			continue
		}

		// offset 已通过 CommitInterval 自动提交
		// 如需手动提交，使用 c.reader.CommitMessages(ctx, msg)
	}
}

// Close 关闭消费者
func (c *KafkaConsumer) Close() error {
	return c.reader.Close()
}
