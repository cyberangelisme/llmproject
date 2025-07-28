package model

// Document 文档格式
type Document struct {
	DocId int64  `json:"doc_id"`
	Title string `json:"title"`
	Body  string `json:"body"`
}

type KeyValue struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}


// Tokenization 分词返回结构
type Tokenization struct {
	Token string // 词条
	// Position int64  // 词条在文本的位置 // TODO 后面再补上
	// Offset   int64  // 偏移量
	DocId int64
}
