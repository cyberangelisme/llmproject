package mapreduce_test

import (
	"server/pkg/mapreduce"
	"testing"
)

func TestMapReduceWithFilePaths(t *testing.T) {
	_ = mapreduce.MapReduceWithFilePaths([]string{"./movies_data.csv"})
}
