package doc2any

// doc2tokens 可以借助大模型或者深度学习nlp的技术进行文档分词，毕竟是异步的，有一定的cost容忍
import (
	"errors"
	"server/model"
	"strings"

	"github.com/go-ego/gse"
)

var seg gse.Segmenter

func init() {
	seg.LoadDict()
}

func GseCutForBuildIndex(docID int64, content string) ([]*model.Tokenization, error) {
	// 可以在这里添加针对索引构建的预处理，例如：
	// - 去除特殊字符
	replacer := strings.NewReplacer(
		"!", "", "@", "", "#", "", "$", "", "%", "", "^", "", "&", "", "*", "", "(", "", ")", "",
		"-", "", "+", "", "=", "", "[", "", "]", "", "{", "", "}", "", "|", "", "\\", "",
		":", "", ";", "", "\"", "", "'", "", "<", "", ">", "", ",", "", ".", "", "?", "", "/", "",
		"`", "", "~", "", "《", "", "》", "", "。", "", "，", "", "、", "",
		// 可以根据需要添加更多
	)
	processedContent := replacer.Replace(content)
	// - 转换为小写 (对中文影响不大)
	// - 过滤停用词 (需要额外的停用词表)

	// 调用基础分词函数
	segments := seg.Segment([]byte(processedContent))
	tokens := gse.ToSlice(segments, false)
	tokens_with_docID := make([]*model.Tokenization, len(tokens))
	for k, v := range tokens {
		tokens_with_docID[k] = &model.Tokenization{
			Token: v,
			DocId: docID,
		}
	}
	// 在更复杂的实现中，这里可能还会：
	// - 过滤掉长度小于某值的词
	// - 过滤掉停用词

	// 因为 GseCut 本身不返回错误，我们简单地返回 nil
	// 如果后续加入可能出错的逻辑（如加载词典失败），需要处理 error
	return tokens_with_docID, nil
}

// 用于倒排索引的tokens 构建
func Doc2Tokens(docStr string) ([]*model.Tokenization, error) {
	docStruct, err := Doc2Struct(docStr)
	if err != nil {
		return nil, errors.New("Doc2Struct err!")
	}
	tokens, err := GseCutForBuildIndex(docStruct.DocId, docStruct.Body)
	if err != nil {
		return nil, errors.New("Doc2Tokens err!")
	}
	return tokens, err
}
