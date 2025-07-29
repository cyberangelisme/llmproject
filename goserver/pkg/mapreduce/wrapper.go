package mapreduce

import (
	"fmt"
	"os"
	"server/model"
	"server/pkg/doc2any"
	"strings"

	"github.com/RoaringBitmap/roaring"
	cmap "github.com/orcaman/concurrent-map/v2"
	"github.com/spf13/cast"
)

func MapReduceWithFilePaths(filePaths []string) cmap.ConcurrentMap[string, *roaring.Bitmap] {
	invertedIndex := cmap.New[*roaring.Bitmap]()

	_, _ = MapReduce(func(source chan<- []byte) {
		for _, path := range filePaths {
			content, _ := os.ReadFile(path)
			source <- content
		}
	}, func(item []byte, writer Writer[[]*model.KeyValue], cancel func(error)) {
		res := make([]*model.KeyValue, 0, 1e3)
		lines := strings.Split(string(item), "\r\n")
		for _, line := range lines[1:] {
			docStruct, _ := doc2any.Doc2Struct(line)
			if docStruct.DocId == 0 {
				continue
			}
			tokens, _ := doc2any.GseCutForBuildIndex(docStruct.DocId, docStruct.Body)
			for _, v := range tokens {
				res = append(res, &model.KeyValue{Key: v.Token, Value: cast.ToString(v.DocId)})
			}
		}
		writer.Write(res)
	}, func(pipe <-chan []*model.KeyValue, writer Writer[string], cancel func(error)) {
		for values := range pipe {
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
	})
	keys := invertedIndex.Keys()
	for _, v := range keys {
		val, _ := invertedIndex.Get(v)
		fmt.Println(v, val)
	}
	return invertedIndex
}
