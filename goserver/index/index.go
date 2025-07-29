package index

import (
	"fmt"
	"server/storage"

	"github.com/RoaringBitmap/roaring"
	cmap "github.com/orcaman/concurrent-map/v2"
)

// 索引服务

func GenerateIndex() {
	// mapreduce

	// store result

}

func StoreInvertedIndexToLevelDB(invertedIndex cmap.ConcurrentMap[string, *roaring.Bitmap], ldbPath string) (int, error) {
	if invertedIndex.Keys() == nil {
		return 0, fmt.Errorf("invertedIndex is nil")
	}

	ldb, err := storage.NewLevelDBWrapper(ldbPath, false)
	if err != nil {
		return 0, fmt.Errorf("failed to open LevelDB at %s: %w", ldbPath, err)
	}
	defer func() {
		if closeErr := ldb.Close(); closeErr != nil {
			// 记录关闭错误，但不中断主流程返回
			fmt.Printf("Warning: Failed to close LevelDB: %v\n", closeErr)
		}
	}()

	// 2. 初始化计数器和错误计数器
	totalKeysWritten := 0
	writeErrors := 0
	fmt.Printf("Starting to write index to LevelDB at %s...\n", ldbPath)
	for item := range invertedIndex.Items() {
		key := item.Key
		bitmap := item.Val

		// 4. 序列化 roaring.Bitmap
		bitmapBytes, marshalErr := bitmap.MarshalBinary()
		if marshalErr != nil {
			fmt.Printf("Error marshaling bitmap for key '%s': %v\n", key, marshalErr)
			writeErrors++
			// 跳过这个 key，继续处理下一个
			continue
		}

		// 5. 写入 LevelDB
		// 使用你封装的 Put 方法
		ldbErr := ldb.Put([]byte(key), bitmapBytes)
		if ldbErr != nil {
			fmt.Printf("Error writing key '%s' to LevelDB: %v\n", key, ldbErr)
			writeErrors++
			// 跳过这个 key，继续处理下一个
			continue
		} else {
			totalKeysWritten++
		}
	}

}

// ForwardIndex 的构建应当是从Kafka 收到消息之后
func StoreForwardIndexToLevelDB(forwardIndex) {
}

func ReceiveDataFromKafka() {

}
