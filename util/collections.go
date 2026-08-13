package util

import (
	"sort"

	"golang.org/x/exp/rand"
)

func RandomBetween(min, max int) int {
	return rand.Intn(max-min+1) + min
}

// Contains 判断目标字符串是否存在于字符串切片中
func Contains(list []string, target string) bool {
	if target == "" {
		return false
	}

	for _, item := range list {
		if item == target {
			return true
		}
	}
	return false
}

// Rect 矩形结构体
type Rect struct {
	X1, Y1, X2, Y2 int
}

// SortRectsByPosition 按照从上到下、从左到右的顺序排序矩形
// 输入: OpenCV.FindImageAll 返回的结果 []map[string]interface{}
// 输出: 排序后的矩形列表 []Rect
func SortRectsByPosition(rets []map[string]interface{}) []Rect {
	var rects []Rect

	// 转换为 Rect 结构体
	for _, ret := range rets {
		rect := Rect{
			X1: ret["x1"].(int),
			Y1: ret["y1"].(int),
			X2: ret["x2"].(int),
			Y2: ret["y2"].(int),
		}
		rects = append(rects, rect)
	}

	// 排序：优先按 y1 从小到大，然后按 x1 从小到大
	sort.Slice(rects, func(i, j int) bool {
		if rects[i].Y1 != rects[j].Y1 {
			return rects[i].Y1 < rects[j].Y1 // 从上到下
		}
		return rects[i].X1 < rects[j].X1 // 从左到右
	})

	return rects
}

// CompleteGridRects 补全网格矩形，基于已识别的格子推测缺失的格子
// rects: 已排序的矩形列表
// cols: 每行的列数（例如背包每行4个格子）
// 返回补全后的矩形列表
func CompleteGridRects(rects []Rect, cols int) []Rect {
	if len(rects) == 0 || cols <= 0 {
		return rects
	}

	// 按行分组，每行最多cols个格子
	rows := make(map[int][]Rect)
	for _, rect := range rects {
		rows[rect.Y1] = append(rows[rect.Y1], rect)
	}

	// 对每行内的矩形按X坐标排序
	for y := range rows {
		sort.Slice(rows[y], func(i, j int) bool {
			return rows[y][i].X1 < rows[y][j].X1
		})
	}

	var completeRects []Rect

	// 处理每一行
	for _, y := range getSortedKeys(rows) {
		rowRects := rows[y]
		completedRow := completeRowRects(rowRects, cols)
		completeRects = append(completeRects, completedRow...)
	}

	return completeRects
}

// getSortedKeys 获取排序后的Y坐标键
func getSortedKeys(rows map[int][]Rect) []int {
	var keys []int
	for k := range rows {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	return keys
}

// completeRowRects 补全单行的矩形
func completeRowRects(rowRects []Rect, cols int) []Rect {
	if len(rowRects) == 0 {
		return rowRects
	}

	// 如果当前行已经有足够的格子，直接返回
	if len(rowRects) >= cols {
		return rowRects[:cols] // 只取前cols个
	}

	// 格子间距固定为73
	gridSpacing := 73
	width := rowRects[0].X2 - rowRects[0].X1
	height := rowRects[0].Y2 - rowRects[0].Y1

	// 创建完整的行
	var completeRow []Rect
	y := rowRects[0].Y1

	// 找到第一个格子的起始位置
	startX := rowRects[0].X1

	// 创建现有格子的位置映射
	existingPositions := make(map[int]Rect)
	for _, rect := range rowRects {
		existingPositions[rect.X1] = rect
	}

	// 生成完整的行（cols个格子）
	for i := 0; i < cols; i++ {
		expectedX := startX + i*gridSpacing

		// 检查这个位置是否有现有的格子
		if existingRect, exists := existingPositions[expectedX]; exists {
			completeRow = append(completeRow, existingRect)
		} else {
			// 检查是否有接近这个位置的格子
			closestRect := findClosestRect(expectedX, rowRects, gridSpacing/2)
			if closestRect != nil {
				completeRow = append(completeRow, *closestRect)
			} else {
				// 创建新的格子
				newRect := Rect{
					X1: expectedX,
					Y1: y,
					X2: expectedX + width,
					Y2: y + height,
				}
				completeRow = append(completeRow, newRect)
			}
		}
	}

	return completeRow
}

// findClosestRect 在指定范围内查找最接近目标X坐标的格子
func findClosestRect(targetX int, rects []Rect, tolerance int) *Rect {
	for _, rect := range rects {
		if abs(rect.X1-targetX) <= tolerance {
			return &rect
		}
	}
	return nil
}

// Abs 计算绝对值（导出函数）
func Abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
