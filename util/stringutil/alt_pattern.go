package stringutil

import "strings"

// MapContainsAltPattern 将 pattern 按 | 切成多段（去空白），若 world 包含任一段则 true；无有效段则 false。
func MapContainsAltPattern(world, pattern string) bool {
	for _, p := range strings.Split(pattern, "|") {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if strings.Contains(world, p) {
			return true
		}
	}
	return false
}
