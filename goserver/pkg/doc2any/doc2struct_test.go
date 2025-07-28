package doc2any_test

import (
	"fmt"
	"os"
	"server/pkg/doc2any"
	"strings"
	"testing"
)

func TestDoc2struct(t *testing.T) {

	// read from csv file
	path := "./movies_data.csv"
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
		docStruct, _ := doc2any.Doc2Struct(line)
		fmt.Println(docStruct)
		os.Exit(0)
	}
}
