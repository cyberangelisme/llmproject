package doc2any_test

import (
	"fmt"
	"os"
	"server/pkg/doc2any"
	"strings"
	"testing"
)

func TestDoc2Tokens(t *testing.T) {

	// read from csv file
	path := "./movies_metadata.csv"
	// filename := "movie.csv"
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Error(err)
	}

	// cut2lines
	// res = make([]*model.KeyValue, 0)
	lines := strings.Split(string(contents), "\r\n")

	// a line2struct
	for _, line := range lines[1:] {
		docTokens, _ := doc2any.Doc2Tokens(line)
		for _, v := range docTokens {
			fmt.Println(v.Token, v.DocId)
		}

		// doc2tokens
		os.Exit(0)
	}
}
