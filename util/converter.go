package util

import (
	"strconv"
	"strings"
)

// SafeStringToInt 安全地将字符串转换为 int，转换失败返回 0
func SafeStringToInt(s string) int {
	if val, err := strconv.Atoi(strings.TrimSpace(s)); err == nil {
		return val
	}
	return 0
}

// SafeStringToInt64 安全地将字符串转换为 int64，转换失败返回 0
func SafeStringToInt64(s string) int64 {
	if val, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64); err == nil {
		return val
	}
	return 0
}

// SafeStringToIntWithSeparator 安全地将字符串按分隔符分割后取第一部分转换为 int
// 例如: "1234(+100)" 按 "(" 分割取 "1234" 转换为 int
func SafeStringToIntWithSeparator(s, separator string) int {
	parts := strings.Split(s, separator)
	if len(parts) > 0 {
		return SafeStringToInt(parts[0])
	}
	return 0
}

// SafeStringToInt64WithSeparator 安全地将字符串按分隔符分割后取第一部分转换为 int64
// 例如: "1234(+100)" 按 "(" 分割取 "1234" 转换为 int64
func SafeStringToInt64WithSeparator(s, separator string) int64 {
	parts := strings.Split(s, separator)
	if len(parts) > 0 {
		return SafeStringToInt64(parts[0])
	}
	return 0
}
