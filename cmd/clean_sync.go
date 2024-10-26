package main

import (
	"bufio"
	"os"
	"strings"
)

// DedupLines 根据每行第一个字段去重并保留最后出现的行
func DedupLines(lines []string) []string {
	lineMap := make(map[string]string)
	for _, line := range reverse(lines) {
		//fields := strings.Fields(line)
		fields := strings.Split(line, ",")
		if len(fields[0]) > 0 {
			lineMap[fields[0]] = line
		}
	}

	dedupedLines := make([]string, 0, len(lineMap))
	for _, line := range lineMap {
		dedupedLines = append(dedupedLines, line)
	}

	return dedupedLines
}

// reverse 反转切片顺序
func reverse(slice []string) []string {
	for i, j := 0, len(slice)-1; i < j; i, j = i+1, j-1 {
		slice[i], slice[j] = slice[j], slice[i]
	}
	return slice
}

// InPlaceModifyFile 原地修改文件，实现去重功能
func InPlaceModifyFile(filePath string) error {
	// 打开文件用于读取
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// 使用缓冲读取器读取文件内容到切片
	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	// 对读取的行进行去重处理
	dedupedLines := DedupLines(lines)

	// 重新打开文件用于写入（会截断原有内容）
	file, err = os.OpenFile(filePath, os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// 使用缓冲写入器将去重后的行写入文件
	writer := bufio.NewWriter(file)
	for _, line := range dedupedLines {
		_, err := writer.WriteString(line + "\n")
		if err != nil {
			return err
		}
	}

	// 刷新缓冲写入器，确保数据写入文件
	err = writer.Flush()
	if err != nil {
		return err
	}

	return nil
}
