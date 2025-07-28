package doc2any

import (
	"strings"

	"server/model"

	"github.com/CocaineCong/tangseng/pkg/util/stringutils"
	"github.com/spf13/cast"
)

func Doc2Struct(docStr string) (doc *model.Document, err error) {
	docStr = strings.Replace(docStr, "\"", "", -1)
	d := strings.Split(docStr, ",")
	something2Str := make([]string, 0)

	for i := 2; i < 5; i++ {
		if len(d) > i && d[i] != "" {
			something2Str = append(something2Str, d[i])
		}
	}

	doc = &model.Document{
		DocId: cast.ToInt64(d[0]),
		Title: d[1],
		Body:  stringutils.StrConcat(something2Str),
	}
	return
}
