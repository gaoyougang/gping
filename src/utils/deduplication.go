package utils

import (
	"sort"
	"strconv"
	"strings"
)

// 移除[]string重复项
func RemoveDuplicates(slice []string) []string {
    encountered := map[string]bool{}
    result := []string{}
    
    for _, v := range slice {
        if !encountered[v] {
            encountered[v] = true
            result = append(result, v)
        }
    }
    return result
}

// 判断[]string是否为空
func IsNonEmptyStrings(slice []string) bool {
    if len(slice) == 0 {
        return false
    }
    for _, s := range slice {
        if s != "" {
            return true
        }
    }
    return false
}

// 通过指定列,对[][]string进行排序
func SortByNumericColumn(data [][]string, column int) [][]string {
	sort.Slice(data, func(i, j int) bool {
		a, _ := strconv.Atoi(data[i][column])
		b, _ := strconv.Atoi(data[j][column])
		return a < b
	})
	return data
}

// 对[]string进行过滤
func FilterLines(lines []string) []string {
    var result []string
    for _, line := range lines {
        trimmed := strings.TrimSpace(line)
        // 跳过空行和以#开头的行
        if trimmed != "" && !strings.HasPrefix(trimmed, "#") {
            result = append(result, line)
        }
    }
    return result
}