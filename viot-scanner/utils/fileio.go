package utils

import (
	"encoding/csv"
	"io"
	"log"
	"os"
)

// ReadCSVByColumn 讀取 CSV 文件並返回指定列的值
func ReadCSVByColumn(filename, columnName string) []string {
	// 檢查文件是否存在
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		log.Printf("⚠️ 文件不存在: %s", filename)
		return []string{}
	}

	// 打開文件
	file, err := os.Open(filename)
	if err != nil {
		log.Printf("⚠️ 打開文件失敗: %v", err)
		return []string{}
	}
	defer file.Close()

	// 創建 CSV reader
	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1 // 允許每行欄位數不同

	// 讀取標題行
	headers, err := reader.Read()
	if err != nil {
		log.Printf("⚠️ 讀取標題行失敗: %v", err)
		return []string{}
	}

	// 找到指定列的索引
	columnIndex := -1
	for i, header := range headers {
		if header == columnName {
			columnIndex = i
			break
		}
	}
	if columnIndex == -1 {
		log.Printf("⚠️ 未找到列: %s", columnName)
		return []string{}
	}

	// 讀取數據行
	var values []string
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		// 只要能讀取到數據行就加入
		values = append(values, record[columnIndex])
	}

	return values
}
