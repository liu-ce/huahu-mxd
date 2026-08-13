package play

import (
	"fmt"
	"strconv"
)

const fashionLogTag = "[时尚大道]"

var farmLogTag = fashionLogTag

// SetFarmLogTag 切换挂机日志前缀（各 Play_* 入口调用）。
func SetFarmLogTag(tag string) {
	if tag != "" {
		farmLogTag = tag
	}
}

func farmLog(format string, args ...interface{}) {
	fmt.Printf(farmLogTag+" "+format+"\n", args...)
}

func cornerName(c cornerID) string {
	switch c {
	case cornerD:
		return "D"
	case cornerC:
		return "C"
	case cornerB:
		return "B"
	case cornerA:
		return "A"
	default:
		return "?"
	}
}

func legName(cur, tgt cornerID) string {
	return cornerName(cur) + "→" + cornerName(tgt)
}

func attackRegionLabel(cfg *mapConfig, lay, relX int) string {
	if cfg != nil && cfg.Type == MapTypeLinear {
		if x1, y1, x2, y2, ok := regionForLayer(cfg.YoloRegions, lay); ok {
			return "YOLO扫描区" + formatRegion(x1, y1, x2, y2)
		}
		return "YOLO扫描"
	}
	if cfg != nil && cfg.Type != MapTypeFashionAvenue {
		if x1, y1, x2, y2, ok := regionForLayer(cfg.AttackRegions, lay); ok {
			return "配置攻击区" + formatRegion(x1, y1, x2, y2)
		}
		return "YOLO全扫描"
	}
	switch lay {
	case 1:
		if relX > attackRelXRightThreshold {
			return "1层右侧固定区"
		}
		if relX < attackRelXLeftThreshold {
			return "1层左侧固定区"
		}
	case 2:
		if relX < attackRelXLeftThreshold {
			return "2层左侧固定区"
		}
		if relX > attackRelXRightThreshold {
			return "2层右侧固定区"
		}
	}
	return "中间默认区(layer" + strconv.Itoa(lay) + ")"
}

func formatRegion(x1, y1, x2, y2 int) string {
	return fmt.Sprintf("[%d,%d,%d,%d]", x1, y1, x2, y2)
}
