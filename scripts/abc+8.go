/*
透過命令行傳遞目標abc記譜檔案路徑 相容相對和絕對路徑
對目標檔案的所有以`[V:2]`開頭行 將該行進行升八度
*/
package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("請提供目標 .abc 檔案路徑作為參數")
		return
	}

	inputPath, err := filepath.Abs(os.Args[1])
	if err != nil {
		fmt.Printf("無法解析路徑: %v\n", err)
		return
	}

	file, err := os.Open(inputPath)
	if err != nil {
		fmt.Printf("無法開啟檔案: %v\n", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var outputLines []string

	// 正則：匹配所有前導區段 [XXX:YYY]
	headerRegex := regexp.MustCompile(`^(\s*(\[[^\]]+\])+)(.*)$`)

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "[V:2]") {
			// 拆分前導區段與音符內容
			matches := headerRegex.FindStringSubmatch(line)
			if len(matches) == 4 {
				prefix := matches[1]  // 所有前導區段
				content := matches[3] // 音符內容
				converted := transposeOctave(content)
				outputLines = append(outputLines, prefix+converted)
			} else {
				// 無法解析，整行處理
				outputLines = append(outputLines, transposeOctave(line))
			}
		} else {
			outputLines = append(outputLines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("讀取檔案時發生錯誤: %v\n", err)
		return
	}

	err = os.WriteFile(inputPath, []byte(strings.Join(outputLines, "\n")), 0644)
	if err != nil {
		fmt.Printf("寫入檔案失敗: %v\n", err)
		return
	}

	fmt.Println("處理完成，原始檔案已更新。")
}

func transposeOctave(input string) string {
	// 正則：感嘆號標記（如 !arpeggio!）
	exclamRegex := regexp.MustCompile(`![^!]+!`)
	// 正則：區段標記（只忽略含冒號的，如 [K:Cm]、[V:2]）
	sectionRegex := regexp.MustCompile(`\[[^:\]]+:[^\]]+\]`)
	// 正則：音符（含升降還原符號）
	noteRegex := regexp.MustCompile(`([_=^]?)([A-Ga-g])(,)?`)

	// 暫存感嘆號標記
	exclamations := exclamRegex.FindAllString(input, -1)
	for i, ex := range exclamations {
		placeholder := fmt.Sprintf("§E%d§", i)
		input = strings.Replace(input, ex, placeholder, 1)
	}

	// 暫存區段標記（只含冒號的）
	sections := sectionRegex.FindAllString(input, -1)
	for i, sec := range sections {
		placeholder := fmt.Sprintf("§S%d§", i)
		input = strings.Replace(input, sec, placeholder, 1)
	}

	// 音符升八度（和弦內也會處理）
	converted := noteRegex.ReplaceAllStringFunc(input, func(note string) string {
		m := noteRegex.FindStringSubmatch(note)
		if len(m) != 4 {
			return note
		}
		modifier := m[1]
		base := m[2]
		comma := m[3]

		if comma == "," {
			return modifier + base // 去掉逗號
		}
		if base >= "A" && base <= "G" {
			return modifier + strings.ToLower(base)
		}
		if base >= "a" && base <= "g" {
			return modifier + base + "'"
		}
		return note
	})

	// 還原區段標記
	for i, sec := range sections {
		placeholder := fmt.Sprintf("§S%d§", i)
		converted = strings.Replace(converted, placeholder, sec, 1)
	}

	// 還原感嘆號標記
	for i, ex := range exclamations {
		placeholder := fmt.Sprintf("§E%d§", i)
		converted = strings.Replace(converted, placeholder, ex, 1)
	}

	return converted
}
