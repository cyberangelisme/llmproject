// bitmap.go - 封装 Redis Bitmap 操作工具
package storage

import (
	"context"
	"fmt"
	"log"

	"github.com/go-redis/redis/v8"
)

// BitmapClient 封装 Redis Bitmap 操作
type BitmapClient struct {
	client    *redis.Client
	ctx       context.Context
	namespace string // 命名空间，用于区分不同项目或用途
}

// Config Redis 配置
type Config struct {
	Addr      string // Redis 地址，如 "localhost:6379"
	Password  string // 密码（无则为空）
	DB        int    // 数据库编号
	Namespace string // 命名空间前缀，如 "projectA"
}

// NewBitmapClient 创建一个新的 Bitmap 客户端
func NewBitmapClient(config Config) (*BitmapClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     config.Addr,
		Password: config.Password,
		DB:       config.DB,
	})

	ctx := context.Background()
	// 检查连接
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("无法连接 Redis: %v", err)
	}

	return &BitmapClient{
		client:    client,
		ctx:       ctx,
		namespace: config.Namespace,
	}, nil
}

// 构建实际的 Redis key
func (b *BitmapClient) makeKey(key string) string {
	if b.namespace == "" {
		return key
	}
	return fmt.Sprintf("%s:%s", b.namespace, key)
}

// SetBit 设置某个偏移量的 bit 值（0 或 1）
// key: 逻辑 key 名（如 "user:active"）
// offset: 位偏移量（通常为 docID 或用户 ID）
// value: 0 或 1
func (b *BitmapClient) SetBit(key string, offset int64, value int) error {
	err := b.client.SetBit(b.ctx, b.makeKey(key), offset, int(value)).Err()
	if err != nil {
		return fmt.Errorf("SetBit 错误: %v", err)
	}
	return nil
}

// GetBit 获取某个偏移量的 bit 值
// 返回值：0, 1, 或错误
func (b *BitmapClient) GetBit(key string, offset int64) (int64, error) {
	val, err := b.client.GetBit(b.ctx, b.makeKey(key), offset).Result()
	if err != nil {
		return 0, fmt.Errorf("GetBit 错误: %v", err)
	}
	return val, nil
}

// Count 统计 key 中值为 1 的 bit 数量（即 BITCOUNT）
func (b *BitmapClient) Count(key string) (int64, error) {
	count, err := b.client.BitCount(b.ctx, b.makeKey(key), nil).Result()
	if err != nil {
		return 0, fmt.Errorf("Count 错误: %v", err)
	}
	return count, nil
}

// OpAnd 对多个 key 执行 AND 操作，结果存入 destKey
func (b *BitmapClient) OpAnd(destKey string, keys ...string) error {
	redisKeys := make([]string, len(keys))
	for i, k := range keys {
		redisKeys[i] = b.makeKey(k)
	}
	dest := b.makeKey(destKey)
	err := b.client.BitOpAnd(b.ctx, dest, redisKeys...).Err()
	if err != nil {
		return fmt.Errorf("OpAnd 错误: %v", err)
	}
	return nil
}

// OpOr 对多个 key 执行 OR 操作
func (b *BitmapClient) OpOr(destKey string, keys ...string) error {
	redisKeys := make([]string, len(keys))
	for i, k := range keys {
		redisKeys[i] = b.makeKey(k)
	}
	dest := b.makeKey(destKey)
	err := b.client.BitOpOr(b.ctx, dest, redisKeys...).Err()
	if err != nil {
		return fmt.Errorf("OpOr 错误: %v", err)
	}
	return nil
}

// OpXor 对多个 key 执行 XOR 操作
func (b *BitmapClient) OpXor(destKey string, keys ...string) error {
	redisKeys := make([]string, len(keys))
	for i, k := range keys {
		redisKeys[i] = b.makeKey(k)
	}
	dest := b.makeKey(destKey)
	err := b.client.BitOpXor(b.ctx, dest, redisKeys...).Err()
	if err != nil {
		return fmt.Errorf("OpXor 错误: %v", err)
	}
	return nil
}

// OpNot 对单个 key 执行 NOT 操作
func (b *BitmapClient) OpNot(destKey, key string) error {
	src := b.makeKey(key)
	dest := b.makeKey(destKey)
	err := b.client.BitOpNot(b.ctx, dest, src).Err()
	if err != nil {
		return fmt.Errorf("OpNot 错误: %v", err)
	}
	return nil
}

// Close 关闭 Redis 连接
func (b *BitmapClient) Close() error {
	return b.client.Close()
}

func (b *BitmapClient) FlushClient() error {
	var ctx = context.Background()
	// ✅ 清空当前数据库
	if err := b.client.FlushDB(ctx).Err(); err != nil {
		log.Fatal("FlushDB 失败:", err)
	}
	fmt.Println("✅ 当前数据库已清空")
	return nil
}
