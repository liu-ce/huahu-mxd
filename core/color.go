package core

import (
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Dasongzi1366/AutoGo/images"
)

// ColorHandler 处理颜色识别相关操作
type ColorHandler struct {
	mu sync.Mutex
}

// NewColorHandler 创建一个新的ColorHandler实例
func NewColorHandler() *ColorHandler {
	return &ColorHandler{}
}

// Pixel 获取指定坐标点的颜色值
// x, y: 坐标点的位置
// 返回值: 表示颜色值的 "RRGGBB" 格式字符串
func (h *ColorHandler) Pixel(x, y int) string {
	h.mu.Lock()
	defer h.mu.Unlock()

	color := images.Pixel(x, y, 0)
	if color == "" {
		log.Printf("获取坐标(%d, %d)的颜色失败", x, y)
		return ""
	}

	return color
}

// CmpColor 比较指定坐标点的颜色
// x, y: 坐标点的位置
// colorStr: 颜色字符串，格式如 "FFFFFF|CCCCCC-101010"
// sim: 相似度，范围 0.1 - 1.0
// 返回值: true 表示颜色匹配，false 表示颜色不匹配
func (h *ColorHandler) CmpColor(x, y int, colorStr string, sim float32) bool {
	h.mu.Lock()
	defer h.mu.Unlock()

	if sim < 0.1 || sim > 1.0 {
		log.Printf("相似度参数无效: %f，应该在0.1-1.0范围内", sim)
		return false
	}

	matched := images.CmpColor(x, y, colorStr, sim, 0)
	return matched
}

// FindColor 在指定区域内查找目标颜色
// x1, y1: 区域左上角的坐标
// x2, y2: 区域右下角的坐标，当 x2 或 y2 为 0 时，表示使用屏幕的最大宽度或高度
// colorStr: 颜色格式字符串，例如 "FFFFFF|CCCCCC-101010"
// sim: 相似度，范围 0.1 - 1.0
// dir: 查找方向，0-从左到右/从上到下，1-从右到左/从上到下，2-从左到右/从下到上，3-从右到左/从下到上
// 返回值: 找到颜色的坐标 (x, y)，找不到返回 (-1, -1)
func (h *ColorHandler) FindColor(x1, y1, x2, y2 int, colorStr string, sim float32, dir int) (int, int) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if sim < 0.1 || sim > 1.0 {
		log.Printf("相似度参数无效: %f，应该在0.1-1.0范围内", sim)
		return -1, -1
	}

	if dir < 0 || dir > 3 {
		log.Printf("查找方向参数无效: %d，应该在0-3范围内", dir)
		return -1, -1
	}

	x, y := images.FindColor(x1, y1, x2, y2, colorStr, sim, dir, 0)
	return x, y
}

// GetColorCountInRegion 计算指定区域内符合颜色条件的像素数量
// x1, y1: 区域左上角的坐标
// x2, y2: 区域右下角的坐标，当 x2 或 y2 为 0 时，表示使用屏幕的最大宽度或高度
// colorStr: 要查找的颜色字符串，格式为 "FFFFFF|CCCCCC-101010"
// sim: 相似度，范围 0.1 - 1.0
// 返回值: 符合条件的颜色像素数量，如果未找到符合条件的像素，则返回 0
func (h *ColorHandler) GetColorCountInRegion(x1, y1, x2, y2 int, colorStr string, sim float32) int {
	h.mu.Lock()
	defer h.mu.Unlock()

	if sim < 0.1 || sim > 1.0 {
		log.Printf("相似度参数无效: %f，应该在0.1-1.0范围内", sim)
		return 0
	}

	count := images.GetColorCountInRegion(x1, y1, x2, y2, colorStr, sim, 0)
	return count
}

// FindColorWindowPeakCenter 在区域内找 color 命中最多的 winW×winH 窗口中心；不足 minInWin 返回 (-1,-1)。
func (h *ColorHandler) FindColorWindowPeakCenter(x1, y1, x2, y2 int, colorStr string, sim float32, winW, winH, minInWin int) (int, int) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if sim < 0.1 || sim > 1.0 || winW <= 0 || winH <= 0 || minInWin <= 0 {
		return -1, -1
	}
	w := x2 - x1 + 1
	hh := y2 - y1 + 1
	if w < winW || hh < winH {
		return -1, -1
	}

	grid := make([]byte, w*hh)
	for y := y1; y <= y2; y++ {
		for x := x1; x <= x2; x++ {
			if images.CmpColor(x, y, colorStr, sim, 0) {
				grid[(y-y1)*w+(x-x1)] = 1
			}
		}
	}

	best, cx, cy := 0, -1, -1
	for wy := 0; wy <= hh-winH; wy++ {
		for wx := 0; wx <= w-winW; wx++ {
			n := 0
			base := wy*w + wx
			for dy := 0; dy < winH; dy++ {
				row := base + dy*w
				for dx := 0; dx < winW; dx++ {
					if grid[row+dx] != 0 {
						n++
					}
				}
			}
			if n > best {
				best = n
				cx = x1 + wx + winW/2
				cy = y1 + wy + winH/2
			}
		}
	}
	if best < minInWin {
		return -1, -1
	}
	return cx, cy
}

// DetectsMultiColors 根据指定的颜色串信息在屏幕进行多点颜色比对
// colors: 颜色模板字符串，例如 "369,1220,ffab2d-101010,370,1221,24b1ff-101010,380,390,907efd-101010"
// sim: 相似度，范围 0.1 - 1.0
// 返回值: true 表示比对成功，false 表示比对失败
func (h *ColorHandler) DetectsMultiColors(colors string, sim float32) bool {
	h.mu.Lock()
	defer h.mu.Unlock()

	if sim < 0.1 || sim > 1.0 {
		log.Printf("相似度参数无效: %f，应该在0.1-1.0范围内", sim)
		return false
	}

	result := images.DetectsMultiColors(colors, sim, 0)
	return result
}

// FindMultiColors 在指定区域内查找匹配的多点颜色序列
// x1, y1: 区域左上角的坐标
// x2, y2: 区域右下角的坐标，当 x2 或 y2 为 0 时，表示使用屏幕的最大宽度或高度
// colors: 颜色模板字符串，例如 "ffccff-151515,635,978,ffab2d-101010,6,29,24b1ff-101010,68,35,907efd-101010"
// sim: 相似度，范围 0.1 - 1.0
// dir: 查找方向，0-从左到右/从上到下，1-从右到左/从上到下，2-从左到右/从下到上，3-从右到左/从下到上
// 返回值: 匹配的首个颜色点的屏幕坐标 (x, y)，如果未找到则返回 (-1, -1)
func (h *ColorHandler) FindMultiColors(x1, y1, x2, y2 int, colors string, sim float32, dir int) (int, int) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if sim < 0.1 || sim > 1.0 {
		log.Printf("相似度参数无效: %f，应该在0.1-1.0范围内", sim)
		return -1, -1
	}

	if dir < 0 || dir > 3 {
		log.Printf("查找方向参数无效: %d，应该在0-3范围内", dir)
		return -1, -1
	}

	x, y := images.FindMultiColors(x1, y1, x2, y2, colors, sim, dir, 0)
	return x, y
}

// WaitForColor 等待在指定区域内检测到目标颜色
// x1, y1, x2, y2: 检测区域坐标
// colorStr: 目标颜色字符串
// sim: 相似度阈值 (0.1-1.0)
// maxAttempts: 最大检测次数，默认60次
// 返回值: 找到颜色的坐标 (x, y)，找不到返回 (-1, -1)
func (h *ColorHandler) WaitForColor(x1, y1, x2, y2 int, colorStr string, sim float32, maxAttempts int) (int, int) {
	if maxAttempts <= 0 {
		maxAttempts = 60
	}

	for i := 0; i < maxAttempts; i++ {
		x, y := h.FindColor(x1, y1, x2, y2, colorStr, sim, 0)
		if x != -1 && y != -1 {
			return x, y
		}
		time.Sleep(time.Second)
	}

	return -1, -1
}

// IsExistByMultipoints 多点识图方法
// points: 点数组，格式为 [][]interface{}{{x1, y1, color1}, {x2, y2, color2}, ...}
// 其中 x, y 为 int 类型，color 为 string 类型
// sim: 相似度，范围 0.1 - 1.0
// 返回值: true 表示所有点都匹配，false 表示有任意一个点不匹配
func (h *ColorHandler) IsExistByMultipoints(points [][]interface{}, sim float32) bool {
	h.mu.Lock()
	defer h.mu.Unlock()

	if sim < 0.1 || sim > 1.0 {
		return false
	}

	if len(points) == 0 {
		return false
	}

	matchedCount := 0
	// 遍历每个点进行颜色匹配
	for i, point := range points {
		if len(point) != 3 {
			return false
		}

		// 类型断言：前两个为int，最后一个为string
		x, ok1 := point[0].(int)
		y, ok2 := point[1].(int)
		colorStr, ok3 := point[2].(string)

		if !ok1 || !ok2 || !ok3 {
			SLS_Log2(fmt.Sprintf("多点识图: 第%d个点类型断言失败", i+1))
			return false
		}

		// 使用底层方法进行颜色匹配
		matched := images.CmpColor(x, y, colorStr, sim, 0)
		if matched {
			matchedCount++
		} else {
			//fmt.Printf("多点识图: 结果 %d/%d 匹配，返回 false\n", matchedCount, len(points))
			return false
		}
	}

	//fmt.Printf("多点识图: 结果 %d/%d 匹配，返回 true\n", matchedCount, len(points))
	return true
}

// FindByMultipoints 多点识图搜索方法，在屏幕上搜索匹配的点模式并返回坐标
// points: 点数组，格式为 [][]interface{}{{x1, y1, color1}, {x2, y2, color2}, ...}
// 其中 x, y 为 int 类型，color 为 string 类型（如 "897e54"）
// sim: 相似度，范围 0.1 - 1.0
// coords: 可选参数
//   - 如果提供 4 个参数：x1, y1, x2, y2 - 搜索区域坐标，当 x2 或 y2 为 0 时表示全屏搜索
//   - 如果提供 5 个参数：x1, y1, x2, y2, dir - 搜索区域和方向（0-3），默认为0
//
// 返回值: 找到匹配模式时返回第一个点的坐标 (x, y)，找不到返回 (-1, -1)
func (h *ColorHandler) FindByMultipoints(points [][]interface{}, sim float32, coords ...int) (int, int) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if sim < 0.1 || sim > 1.0 {
		log.Printf("相似度参数无效: %f，应该在0.1-1.0范围内", sim)
		return -1, -1
	}

	if len(points) == 0 {
		return -1, -1
	}

	// 解析第一个点作为基准点
	if len(points[0]) != 3 {
		return -1, -1
	}

	baseX, ok1 := points[0][0].(int)
	baseY, ok2 := points[0][1].(int)
	baseColor, ok3 := points[0][2].(string)

	if !ok1 || !ok2 || !ok3 {
		return -1, -1
	}

	// 构建 FindMultiColors 所需的字符串格式
	// 格式: "基准颜色,offsetX1,offsetY1,颜色1,offsetX2,offsetY2,颜色2,..."
	var colorStr strings.Builder

	// 第一个颜色作为基准颜色（参考点）
	colorStr.WriteString(baseColor)

	// 将其他点转换为相对坐标
	for i := 1; i < len(points); i++ {
		if len(points[i]) != 3 {
			continue
		}

		x, ok1 := points[i][0].(int)
		y, ok2 := points[i][1].(int)
		color, ok3 := points[i][2].(string)

		if !ok1 || !ok2 || !ok3 {
			continue
		}

		// 计算相对于基准点的偏移
		offsetX := x - baseX
		offsetY := y - baseY

		// 添加偏移和颜色: ",offsetX,offsetY,color"
		colorStr.WriteString(fmt.Sprintf(",%d,%d,%s", offsetX, offsetY, color))
	}

	// 解析搜索区域和方向参数
	var x1, y1, x2, y2, dir int
	if len(coords) >= 4 {
		x1, y1, x2, y2 = coords[0], coords[1], coords[2], coords[3]
	}
	if len(coords) >= 5 {
		dir = coords[4]
		if dir < 0 || dir > 3 {
			dir = 0
		}
	}

	// 调用 FindMultiColors 搜索
	x, y := images.FindMultiColors(x1, y1, x2, y2, colorStr.String(), sim, dir, 0)
	return x, y
}

// IsExistByMultipointsWithRatio 多点识图方法（支持比例匹配）
// points: 点数组，格式为 [][]interface{}{{x1, y1, color1}, {x2, y2, color2}, ...}
// sim: 颜色相似度，范围 0.1 - 1.0
// ratio: 匹配比例阈值，范围 0.1 - 1.0，例如 0.8 表示 80% 的点匹配即可
// 返回值: true 表示匹配比例达到阈值，false 表示未达到
func (h *ColorHandler) IsExistByMultipointsWithRatio(points [][]interface{}, sim float32, ratio float32) bool {
	h.mu.Lock()
	defer h.mu.Unlock()

	if sim < 0.1 || sim > 1.0 || ratio < 0.1 || ratio > 1.0 {
		return false
	}

	if len(points) == 0 {
		return false
	}

	matchedCount := 0
	validPoints := 0
	// 遍历每个点进行颜色匹配
	for _, point := range points {
		if len(point) != 3 {
			continue // 跳过格式错误的点
		}

		// 类型断言：前两个为int，最后一个为string
		x, ok1 := point[0].(int)
		y, ok2 := point[1].(int)
		colorStr, ok3 := point[2].(string)

		if !ok1 || !ok2 || !ok3 {
			continue // 跳过类型断言失败的点
		}

		validPoints++ // 有效点计数

		// 使用底层方法进行颜色匹配
		matched := images.CmpColor(x, y, colorStr, sim, 0)
		if matched {
			matchedCount++
		}
	}

	// 如果没有有效点，返回 false
	if validPoints == 0 {
		return false
	}

	// 计算匹配比例（基于有效点数）
	matchRatio := float32(matchedCount) / float32(validPoints)
	result := matchRatio >= ratio

	//fmt.Printf("多点识图(比例): 结果 %d/%d 匹配，比例 %.2f，阈值 %.2f，返回 %t\n",
	//	matchedCount, validPoints, matchRatio, ratio, result)

	return result
}

// ========= 工具函数 =========
func (h *ColorHandler) HexToRGB(hex string) (int, int, int) {
	var r, g, b int
	if len(hex) >= 6 {
		fmt.Sscanf(hex[len(hex)-6:], "%02x%02x%02x", &r, &g, &b)
	}
	return r, g, b
}

func (h *ColorHandler) Rgb2hsv(r, g, b int) (H, S, V float64) {
	R := float64(r) / 255.0
	G := float64(g) / 255.0
	B := float64(b) / 255.0

	max := math.Max(R, math.Max(G, B))
	min := math.Min(R, math.Min(G, B))
	d := max - min

	switch {
	case d == 0:
		H = 0
	case max == R:
		H = math.Mod((60*((G-B)/d) + 360), 360)
	case max == G:
		H = math.Mod((60*((B-R)/d) + 120), 360)
	default:
		H = math.Mod((60*((R-G)/d) + 240), 360)
	}

	if max == 0 {
		S = 0
	} else {
		S = d / max
	}
	V = max
	return
}

// CheckRGBChangeRate 检测指定区域在延迟前后RGB的变化率
// x1, y1, x2, y2: 检测区域坐标
// delayMs: 延迟时间（毫秒），默认200ms
// 返回值: R变化率, G变化率, B变化率, 是否成功
func (h *ColorHandler) CheckRGBChangeRate(x1, y1, x2, y2 int, delayMs int) (float64, float64, float64, bool) {
	if delayMs <= 0 {
		delayMs = 200
	}

	// 在区域内采样多个点，计算RGB平均值
	samplePoints := [][]int{
		{x1 + (x2-x1)/4, y1 + (y2-y1)/4},
		{x1 + (x2-x1)/2, y1 + (y2-y1)/2},
		{x1 + (x2-x1)*3/4, y1 + (y2-y1)*3/4},
		{x1 + (x2-x1)/4, y1 + (y2-y1)*3/4},
		{x1 + (x2-x1)*3/4, y1 + (y2-y1)/4},
	}

	// 第一次采样：计算RGB平均值
	var r1, g1, b1 int
	validSamples1 := 0
	for _, point := range samplePoints {
		colorHex := h.Pixel(point[0], point[1])
		if len(colorHex) == 6 {
			r, err1 := strconv.ParseInt(colorHex[0:2], 16, 0)
			g, err2 := strconv.ParseInt(colorHex[2:4], 16, 0)
			b, err3 := strconv.ParseInt(colorHex[4:6], 16, 0)
			if err1 == nil && err2 == nil && err3 == nil {
				r1 += int(r)
				g1 += int(g)
				b1 += int(b)
				validSamples1++
			}
		}
	}

	if validSamples1 == 0 {
		return 0, 0, 0, false
	}

	// 计算平均值
	avgR1 := float64(r1) / float64(validSamples1)
	avgG1 := float64(g1) / float64(validSamples1)
	avgB1 := float64(b1) / float64(validSamples1)

	// 延迟
	time.Sleep(time.Duration(delayMs) * time.Millisecond)

	// 第二次采样：计算RGB平均值
	var r2, g2, b2 int
	validSamples2 := 0
	for _, point := range samplePoints {
		colorHex := h.Pixel(point[0], point[1])
		if len(colorHex) == 6 {
			r, err1 := strconv.ParseInt(colorHex[0:2], 16, 0)
			g, err2 := strconv.ParseInt(colorHex[2:4], 16, 0)
			b, err3 := strconv.ParseInt(colorHex[4:6], 16, 0)
			if err1 == nil && err2 == nil && err3 == nil {
				r2 += int(r)
				g2 += int(g)
				b2 += int(b)
				validSamples2++
			}
		}
	}

	if validSamples2 == 0 {
		return 0, 0, 0, false
	}

	// 计算平均值
	avgR2 := float64(r2) / float64(validSamples2)
	avgG2 := float64(g2) / float64(validSamples2)
	avgB2 := float64(b2) / float64(validSamples2)

	// 计算三种颜色的变化率
	var changeRateR, changeRateG, changeRateB float64
	maxR := avgR1
	if avgR2 > maxR {
		maxR = avgR2
	}
	if maxR > 0 {
		changeRateR = math.Abs(avgR1-avgR2) / maxR * 100
	}

	maxG := avgG1
	if avgG2 > maxG {
		maxG = avgG2
	}
	if maxG > 0 {
		changeRateG = math.Abs(avgG1-avgG2) / maxG * 100
	}

	maxB := avgB1
	if avgB2 > maxB {
		maxB = avgB2
	}
	if maxB > 0 {
		changeRateB = math.Abs(avgB1-avgB2) / maxB * 100
	}

	return changeRateR, changeRateG, changeRateB, true
}

// ======================== 多点识色区域 ========================
