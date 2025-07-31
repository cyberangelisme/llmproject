package storage_test

import (
	"encoding/binary"
	"fmt"
	"log"
	"sort"
	"sync"
	"time"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/storage"
	"golang.org/x/exp/rand"
)

// encode 将 []uint64 编码为字节流（Varint）
func encode(ids []uint64) []byte {
	var buf []byte
	for _, id := range ids {
		var b [binary.MaxVarintLen64]byte
		n := binary.PutUvarint(b[:], id)
		buf = append(buf, b[:n]...)
	}
	return buf
}

// decode 将字节流解码为 []uint64
func decode(data []byte) []uint64 {
	var ids []uint64
	for len(data) > 0 {
		id, n := binary.Uvarint(data)
		if n <= 0 {
			break
		}
		ids = append(ids, id)
		data = data[n:]
	}
	return ids
}

// 去重并排序
func uniqueAndSort(ids []uint64) []uint64 {
	sort.Slice(ids, func(i, j int) bool {
		return ids[i] < ids[j]
	})
	j := 0
	for i := 1; i < len(ids); i++ {
		if ids[i] != ids[j] {
			j++
			ids[j] = ids[i]
		}
	}
	return ids[:j+1]
}

// PostingsMergeComparer 实现 Comparer 接口，并提供 Merge 方法
type PostingsMergeComparer struct{}

func (p *PostingsMergeComparer) Name() string {
	return "postings-merge-comparer"
}

func (p *PostingsMergeComparer) Compare(a, b []byte) int {
	// 简单字节比较
	for i := 0; i < len(a) && i < len(b); i++ {
		if a[i] != b[i] {
			if a[i] < b[i] {
				return -1
			}
			return 1
		}
	}
	if len(a) < len(b) {
		return -1
	}
	if len(a) > len(b) {
		return 1
	}
	return 0
}

// ✅ 关键：这个方法名必须是 Merge，且签名正确
// 当 LevelDB 检测到 Comparer 有 Merge 方法时，自动启用 Merge Operator
func (p *PostingsMergeComparer) Merge(key, existing, delta []byte) ([]byte, bool, error) {
	var result []uint64

	// 解码已有值
	if existing != nil {
		result = decode(existing)
	}

	// 解码增量
	newIDs := decode(delta)
	result = append(result, newIDs...)

	// 去重排序
	result = uniqueAndSort(result)

	// 编码返回
	return encode(result), true, nil
}
func main() {
	memStor := storage.NewMemStorage()

	// ✅ 正确方式：通过 Comparer 注入 Merge Operator
	db, err := leveldb.Open(memStor, &opt.Options{
		Comparer: &PostingsMergeComparer{}, // Merge 是通过这个结构体的 Merge 方法启用的
	})
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	key := []byte("中国")

	// === 并发 Merge 写入 ===
	var wg sync.WaitGroup
	n := 10
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(docID uint64) {
			defer wg.Done()
			time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)
			delta := encode([]uint64{docID})
			err := db.Merge(key, delta, nil)
			if err != nil {
				log.Printf("Merge error: %v", err)
			} else {
				fmt.Printf("写入: Merge('中国', [%d])\n", docID)
			}
		}(uint64(1000 + i))
	}
	wg.Wait()

	// === 读取验证 ===
	data, err := db.Get(key, nil)
	if err != nil {
		log.Fatal(err)
	}
	ids := decode(data)
	fmt.Printf("\nGet('中国') = %v (共 %d 个文档)\n", ids, len(ids))

	// 验证是否完整
	expectedCount := 10
	if len(ids) == expectedCount {
		fmt.Println("✅ 所有增量都正确合并，无丢失")
	} else {
		fmt.Printf("❌ 实际 %d 个，期望 %d 个\n", len(ids), expectedCount)
	}

	// === 触发 Compaction ===
	fmt.Println("\n=== 触发 Compaction ===")
	err = db.CompactRange(key, key)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("✅ Compaction 完成")

	// === 再次读取验证一致性 ===
	data2, _ := db.Get(key, nil)
	ids2 := decode(data2)
	fmt.Printf("Compaction 后: %v\n", ids2)

	match := len(ids) == len(ids2)
	for i := range ids {
		if ids[i] != ids2[i] {
			match = false
			break
		}
	}
	if match {
		fmt.Println("✅ Compaction 前后结果一致")
	} else {
		fmt.Println("❌ 结果不一致！")
	}
}
