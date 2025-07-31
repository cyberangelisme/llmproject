package index_test

import (
	"bufio"
	"os"
	"server/index"
	"server/pkg/mapreduce"
	"testing"
)

func TestStoreInvertedIndexToLevelDB(t *testing.T) {
	// cmap_inverted := mapreduce.MapReduceWithFilePaths([]string{"./part_1.csv",
	// 	"./part_2.csv",
	// 	"./part_3.csv",
	// 	"./part_4.csv",
	// 	"./part_5.csv",
	// 	"./part_6.csv",
	// 	"./part_7.csv",
	// 	"./part_8.csv"})
	cmap_inverted := mapreduce.MapReduceWithFilePaths([]string{"./movies_metadata.csv"})
	index.StoreInvertedIndexToLevelDB(cmap_inverted, "./index.ldb")
}

// 尝试分割文件
func TestSpiltData(t *testing.T) {
	// 输入配置
	const inputFile = "movies_metadata.csv"
	const numFiles = 8 // 用户修改这里：想分成几个文件

	file, _ := os.Open(inputFile)
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var lines []string

	// 读取所有数据行（跳过表头）
	if scanner.Scan() {
		lines = append(lines, scanner.Text()) // 表头
	}
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	// 计算每个文件最多放多少行数据（不含表头）
	totalDataLines := len(lines) - 1
	linesPerFile := (totalDataLines + numFiles - 1) / numFiles // 向上取整
	itoa := func(n int) string {
		if n == 0 {
			return "0"
		}
		s := ""
		for n > 0 {
			s = string(rune('0'+n%10)) + s
			n /= 10
		}
		return s
	}

	// 分割写入
	for i := 0; i < numFiles; i++ {
		start := 1 + i*linesPerFile   // 数据起始
		end := 1 + (i+1)*linesPerFile // 数据结束
		if start >= len(lines) {
			break
		}
		if end > len(lines) {
			end = len(lines)
		}

		f, _ := os.Create("part_" + itoa(i+1) + ".csv")
		w := bufio.NewWriter(f)
		w.WriteString(lines[0] + "\n") // 写入表头
		for j := start; j < end; j++ {
			w.WriteString(lines[1:][j-1] + "\n")
		}
		w.Flush()
		f.Close()
	}
}
