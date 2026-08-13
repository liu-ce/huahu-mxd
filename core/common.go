package core

import (
	"strconv"
	"strings"
)

// GetTaskInterval 获取任务间隔时间（分钟） - 从配置文件中读取，如果没有配置则返回默认值
// 支持解析表达式，如 "4*60", "24*60", "6*24*60" 等
func GetTaskInterval(key string, defaultMinutes int) int {
	configValue := API.GetConfigStringValue("task_intervals." + key)
	if configValue == "" {
		return defaultMinutes
	}

	// 解析表达式，如 "4*60", "24*60", "6*24*60"
	result := parseTimeExpression(configValue)
	if result == 0 {
		return defaultMinutes
	}
	return result
}

// parseTimeExpression 解析时间表达式，如 "4*60", "24*60", "6*24*60"
func parseTimeExpression(expr string) int {
	// 移除空格
	expr = strings.ReplaceAll(expr, " ", "")

	// 按 * 分割
	parts := strings.Split(expr, "*")
	if len(parts) == 0 {
		return 0
	}

	result := 1
	for _, part := range parts {
		num, err := strconv.Atoi(part)
		if err != nil {
			return 0 // 解析失败
		}
		result *= num
	}

	return result
}
