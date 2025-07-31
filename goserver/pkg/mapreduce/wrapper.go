package mapreduce

import (
	"bufio"
	"fmt"
	"os"
	"server/model"
	"server/pkg/doc2any"
	"strings"

	"github.com/RoaringBitmap/roaring"
	cmap "github.com/orcaman/concurrent-map/v2"
	"github.com/spf13/cast"
)

// 理论上来讲pkg不应该过度耦合逻辑，但是这里貌似 异步kafka 送正排索引构建的任务必须由 mapreduce 此处完成，后期需要改进
func MapReduceWithFilePaths(filePaths []string) cmap.ConcurrentMap[string, *roaring.Bitmap] {
	invertedIndex := cmap.New[*roaring.Bitmap]()

	_, _ = MapReduce(func(source chan<- []byte) {
		for _, path := range filePaths {
			content, _ := os.ReadFile(path)
			source <- content

		}
	}, func(item []byte, writer Writer[[]*model.KeyValue], cancel func(error)) {

		fmt.Println("✅ mapper worker 启动")

		res := make([]*model.KeyValue, 0, 1e3)

		// content := string(item)
		// lines := strings.Split(content, "\r\n")

		reader := strings.NewReader(string(item))
		scanner := bufio.NewScanner(reader)

		// 可选：设置大行支持（默认 64KB 限制）
		const maxLineLen = 1 << 10 // 1MB
		scanner.Buffer(nil, maxLineLen)

		var firstLine = true

		for scanner.Scan() {
			line := scanner.Text()
			// 跳过表头
			if firstLine {
				firstLine = false
				continue
			}

			//fmt.Println(line)
			docStruct, _ := doc2any.Doc2Struct(line)

			if docStruct.DocId == 0 {
				continue
			}
			tokens, _ := doc2any.GseCutForBuildIndex(docStruct.DocId, docStruct.Body)
			// 一次性写入
			// for _, v := range tokens {
			// 	res = append(res, &model.KeyValue{Key: v.Token, Value: cast.ToString(v.DocId)})
			// }

			// 内存不够 batch写入
			// 逐个 token 添加到 batch
			for _, v := range tokens {
				if v.Token == "" {
					continue
				}
				res = append(res, &model.KeyValue{
					Key:   v.Token,
					Value: cast.ToString(v.DocId),
				})

				// 批量达到 1000 条就写一次，然后清空
				if len(res) >= 1000 {
					//fmt.Println("写入一次")
					writer.Write(res)
					res = res[:0] // 复用底层数组
				}
			}
		}
		if len(res) > 0 {
			writer.Write(res)
		}

	}, func(pipe <-chan []*model.KeyValue, writer Writer[string], cancel func(error)) {

		fmt.Println("✅ Reducer worker 启动")
		for values := range pipe {
			//fmt.Println("reduce")
			for _, v := range values {
				if value, ok := invertedIndex.Get(v.Key); ok {
					value.AddInt(cast.ToInt(v.Value))
					invertedIndex.Set(v.Key, value)
				} else {
					docIds := roaring.NewBitmap()
					docIds.AddInt(cast.ToInt(v.Value))
					invertedIndex.Set(v.Key, docIds)
				}
			}
		}

	}, WithWorkers(10))
	// keys := invertedIndex.Keys()

	// for _, v := range keys {
	// 	val, _ := invertedIndex.Get(v)
	// 	fmt.Println(v, val)
	// }
	return invertedIndex
}

// 此处的worker收集mapper的数据来的
func WithWorkers(n int) Option {
	return func(opts *mapReduceOptions) {
		if n < minWorkers {
			opts.workers = minWorkers
		} else {
			opts.workers = n
		}
	}
}
