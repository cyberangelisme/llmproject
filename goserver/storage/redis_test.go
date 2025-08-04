// main.go - 测试 BitmapClient 实现倒排索引
package storage_test

import (
	"fmt"
	"log"
	"testing"

	"server/storage" // 替换为你的模块名
)

func TestRedis(t *testing.T) {
	// 初始化 BitmapClient
	config := storage.Config{
		Addr:      "localhost:6379",
		Password:  "123456",
		DB:        0,
		Namespace: "inverted_index", // 所有 key 会加上前缀
	}

	client, err := storage.NewBitmapClient(config)
	if err != nil {
		log.Fatal("Redis 连接失败:", err)
	}
	defer client.Close()

	fmt.Println("✅ Redis 连接成功")

	// 模拟文档数据
	// docID: 1 -> "go programming"
	// docID: 2 -> "go tutorial"
	// docID: 3 -> "rust programming"
	// docID: 4 -> "go rust"

	// 写入倒排索引（term -> docID）
	terms := map[string][]uint32{
		"go":          {1, 2, 4},
		"programming": {1, 3},
		"tutorial":    {2},
		"rust":        {3, 4},
	}

	for term, docIDs := range terms {
		for _, docID := range docIDs {
			err := client.SetBit(term, int64(docID), 1)
			if err != nil {
				log.Printf("写入 term=%s, docID=%d 失败: %v", term, docID, err)
			}
		}
		fmt.Printf("📌 已写入 '%s' -> docIDs: %v\n", term, docIDs)
	}

	// 统计每个 term 的文档数量
	fmt.Println("\n📊 倒排索引统计：")
	for term := range terms {
		count, _ := client.Count(term)
		fmt.Printf("'%s': %d 个文档\n", term, count)
	}

	// 查询：同时包含 "go" 和 "programming" 的文档数（AND）
	err = client.OpAnd("tmp:go_and_prog", "go", "programming")
	if err != nil {
		log.Fatal("AND 操作失败:", err)
	}
	countAnd, _ := client.Count("tmp:go_and_prog")
	fmt.Printf("\n🔍 同时包含 'go' 和 'programming' 的文档数: %d\n", countAnd) // 应为 1 (docID=1)

	// 查询：包含 "go" 或 "rust" 的文档数（OR）
	err = client.OpOr("tmp:go_or_rust", "go", "rust")
	if err != nil {
		log.Fatal("OR 操作失败:", err)
	}
	countOr, _ := client.Count("tmp:go_or_rust")
	fmt.Printf("🔍 包含 'go' 或 'rust' 的文档数: %d\n", countOr) // 应为 4 (1,2,3,4)

	// 查询某个 docID 是否存在（例如：docID=3 是否包含 "programming"）
	bit, _ := client.GetBit("programming", 3)
	if bit == 1 {
		fmt.Println("✅ docID=3 包含 'programming'")
	}

	// 可选：查看 Redis 中的 key（通过命令行）
	fmt.Printf(`
✅ 测试完成！
你可以在 redis-cli 中查看：
  keys inverted_index:*
  getbit inverted_index:go 1
  bitcount inverted_index:go
`)
}
