package core

import (
	"fmt"
	"github.com/Dasongzi1366/AutoGo/images"
	"image"
	"image/color"
	"image/draw"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"strconv"
	"strings"
	"time"
)

// 预定义的模式配置
var defaultPatterns = map[string]string{
	"琥珀盒-5万":  "0.9|0,27,e88f19;2,30,d8a312;6,29,f3d542;20,17,5a4e40;12,4,9a8e7f;24,6,bc9673;30,13,33190d;19,25,8f6b4e;19,30,43382f;6,0,7e634e",
	"高级诺恩的宝藏": "0.95|9,20,302f2c;9,5,d2cbc0;25,0,2d3b2a;39,13,223522;22,20,6d5c4c;4,18,100c0b;4,29,363430;0,38,1d2d1d",
	"烤鲑鱼":     "0.8|2,9,f18038;0,1,967e4e;4,0,704a1b;8,0,9eaa52;12,5,d5945c;13,10,ddbc47;13,13,b49a56;18,12,865524;24,12,b37c53;27,8,c3e79a;33,9,e2dfd8",
}

// Detection 检测结果结构
type Detection struct {
	Label   string  `json:"label"`   // 物体类别
	Score   float64 `json:"score"`   // 置信度 (0~1)
	X       int     `json:"x"`       // 左上角 X
	Y       int     `json:"y"`       // 左上角 Y
	Width   int     `json:"width"`   // 宽度
	Height  int     `json:"height"`  // 高度
	CenterX int     `json:"centerX"` // 中心坐标 X (自动计算)
	CenterY int     `json:"centerY"` // 中心坐标 Y (自动计算)
}

// Results 识别结果集合
type Results struct {
	Detections     []Detection `json:"detections"`
	MatchedTargets []string    `json:"matched_targets"` // 匹配到的目标名称列表
	Count          int         `json:"count"`
}

// ColorPoint 颜色点结构
type ColorPoint struct {
	X         int
	Y         int
	Color     color.RGBA
	Tolerance uint8
}

// MultiPointColorMatcher 多点识色匹配器
type MultiPointColorMatcher struct {
}

// NewMultiPointColorMatcher 创建新的多点识色匹配器
func NewMultiPointColorMatcher() *MultiPointColorMatcher {
	return &MultiPointColorMatcher{}
}

// FindPatterns 在指定区域内查找模式
func (m *MultiPointColorMatcher) FindPatterns(targetNames []string, searchRect *image.Rectangle, img image.Image) (Results, error) {
	results := Results{
		Detections:     make([]Detection, 0, len(targetNames)),
		MatchedTargets: make([]string, 0, len(targetNames)),
	}

	// 转换为RGBA图像
	var rgbaImg *image.RGBA
	if rgba, ok := img.(*image.RGBA); ok {
		rgbaImg = rgba
	} else {
		rgbaImg = image.NewRGBA(img.Bounds())
		draw.Draw(rgbaImg, rgbaImg.Bounds(), img, img.Bounds().Min, draw.Src)
	}

	// 获取搜索区域
	bounds := rgbaImg.Bounds()
	if searchRect != nil {
		bounds = *searchRect
	}

	// 用于记录已找到的道具
	foundItems := make(map[string]bool)

	for _, name := range targetNames {
		// 如果已经找到过这个道具，跳过
		if foundItems[name] {
			continue
		}

		patternStr, exists := defaultPatterns[name]
		if !exists {
			continue
		}

		_, offsets, err := m.parsePattern(patternStr)
		if err != nil {
			continue
		}

		minX, maxX, minY, maxY := m.calculatePatternBounds(offsets)
		searchWidth := maxX - minX
		searchHeight := maxY - minY

		// 标记是否找到该道具
		found := false

		for y := bounds.Min.Y; y < bounds.Max.Y-searchHeight && !found; y++ {
			for x := bounds.Min.X; x < bounds.Max.X-searchWidth && !found; x++ {
				match := true
				for _, p := range offsets {
					tx, ty := x+p.X, y+p.Y
					if tx < bounds.Min.X || tx >= bounds.Max.X || ty < bounds.Min.Y || ty >= bounds.Max.Y {
						match = false
						break
					}
					if !m.colorEqualsWithTolerance(rgbaImg.RGBAAt(tx, ty), p.Color, p.Tolerance) {
						match = false
						break
					}
				}

				if match {
					results.Detections = append(results.Detections, Detection{
						Label:   name,
						Score:   1.0,
						X:       x,
						Y:       y,
						Width:   searchWidth,
						Height:  searchHeight,
						CenterX: x + searchWidth/2,
						CenterY: y + searchHeight/2,
					})
					foundItems[name] = true
					found = true // 找到后立即停止搜索该道具
				}
			}
		}
	}

	// 设置结果数量和匹配目标
	results.Count = len(results.Detections)
	for _, det := range results.Detections {
		results.MatchedTargets = append(results.MatchedTargets, det.Label)
	}

	return results, nil
}

// parsePattern 解析模式字符串
func (m *MultiPointColorMatcher) parsePattern(patternStr string) (color.RGBA, []ColorPoint, error) {
	parts := strings.Split(patternStr, "|")
	if len(parts) != 2 {
		return color.RGBA{}, nil, fmt.Errorf("pattern must contain similarity|points")
	}

	// 解析相似度（这里不使用，保持兼容）
	_, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return color.RGBA{}, nil, fmt.Errorf("invalid similarity: %v", err)
	}

	// 解析颜色点
	pointStrs := strings.Split(parts[1], ";")
	pointStrs = m.removeEmptyStrings(pointStrs)
	if len(pointStrs) < 1 {
		return color.RGBA{}, nil, fmt.Errorf("pattern must contain base color")
	}

	var offsets []ColorPoint
	var baseColor color.RGBA

	for i, pointStr := range pointStrs {
		pointParts := strings.Split(pointStr, ",")
		if len(pointParts) != 3 {
			return color.RGBA{}, nil, fmt.Errorf("invalid point format: %s", pointStr)
		}

		x, err := strconv.Atoi(pointParts[0])
		if err != nil {
			return color.RGBA{}, nil, fmt.Errorf("invalid X offset: %v", err)
		}

		y, err := strconv.Atoi(pointParts[1])
		if err != nil {
			return color.RGBA{}, nil, fmt.Errorf("invalid Y offset: %v", err)
		}

		colorRGBA, tolerance, err := m.parseColorWithTolerance(pointParts[2])
		if err != nil {
			return color.RGBA{}, nil, fmt.Errorf("invalid color at position %d: %v", i, err)
		}

		if i == 0 {
			baseColor = colorRGBA
		}

		offsets = append(offsets, ColorPoint{
			X:         x,
			Y:         y,
			Color:     colorRGBA,
			Tolerance: tolerance,
		})
	}

	return baseColor, offsets, nil
}

// parseColorWithTolerance 解析带容差的颜色
func (m *MultiPointColorMatcher) parseColorWithTolerance(colorStr string) (color.RGBA, uint8, error) {
	parts := strings.Split(colorStr, "-")
	if len(parts[0]) != 6 {
		return color.RGBA{}, 0, fmt.Errorf("color must be 6-digit HEX")
	}

	r, _ := strconv.ParseUint(parts[0][0:2], 16, 8)
	g, _ := strconv.ParseUint(parts[0][2:4], 16, 8)
	b, _ := strconv.ParseUint(parts[0][4:6], 16, 8)

	tolerance := uint8(64) // 默认容差
	if len(parts) > 1 && len(parts[1]) >= 6 {
		tr, _ := strconv.ParseUint(parts[1][0:2], 16, 8)
		tg, _ := strconv.ParseUint(parts[1][2:4], 16, 8)
		tb, _ := strconv.ParseUint(parts[1][4:6], 16, 8)
		tolerance = uint8((tr + tg + tb) / 3)
	}

	return color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: 255}, tolerance, nil
}

// removeEmptyStrings 移除空字符串
func (m *MultiPointColorMatcher) removeEmptyStrings(slice []string) []string {
	var result []string
	for _, s := range slice {
		if s != "" {
			result = append(result, s)
		}
	}
	return result
}

// calculatePatternBounds 计算模式边界
func (m *MultiPointColorMatcher) calculatePatternBounds(points []ColorPoint) (minX, maxX, minY, maxY int) {
	minX, maxX = math.MaxInt32, math.MinInt32
	minY, maxY = math.MaxInt32, math.MinInt32
	for _, p := range points {
		if p.X < minX {
			minX = p.X
		}
		if p.X > maxX {
			maxX = p.X
		}
		if p.Y < minY {
			minY = p.Y
		}
		if p.Y > maxY {
			maxY = p.Y
		}
	}
	return
}

// colorEqualsWithTolerance 带容差的颜色比较
func (m *MultiPointColorMatcher) colorEqualsWithTolerance(a, b color.RGBA, tolerance uint8) bool {
	return m.colorDiff(a.R, b.R) <= tolerance &&
		m.colorDiff(a.G, b.G) <= tolerance &&
		m.colorDiff(a.B, b.B) <= tolerance
}

// colorDiff 颜色差值
func (m *MultiPointColorMatcher) colorDiff(a, b uint8) uint8 {
	if a > b {
		return a - b
	}
	return b - a
}

// nonMaxSuppression 非最大抑制，去除重叠检测
func (m *MultiPointColorMatcher) nonMaxSuppression(detections []Detection, overlapThresh float32) []Detection {
	if len(detections) == 0 {
		return detections
	}

	// 按置信度排序
	for i := 0; i < len(detections)-1; i++ {
		for j := i + 1; j < len(detections); j++ {
			if detections[i].Score < detections[j].Score {
				detections[i], detections[j] = detections[j], detections[i]
			}
		}
	}

	pick := make([]Detection, 0)
	active := make([]bool, len(detections))
	for i := range active {
		active[i] = true
	}

	for i := 0; i < len(detections); i++ {
		if !active[i] {
			continue
		}

		pick = append(pick, detections[i])
		active[i] = false

		for j := i + 1; j < len(detections); j++ {
			if !active[j] {
				continue
			}

			// 只对相同标签的检测进行抑制
			if detections[i].Label != detections[j].Label {
				continue
			}

			overlap := m.computeOverlap(detections[i], detections[j])
			if overlap > overlapThresh {
				active[j] = false
			}
		}
	}

	return pick
}

// computeOverlap 计算两个检测框的重叠率
func (m *MultiPointColorMatcher) computeOverlap(a, b Detection) float32 {
	x1 := maxVal(a.X, b.X)
	y1 := maxVal(a.Y, b.Y)
	x2 := minVal(a.X+a.Width, b.X+b.Width)
	y2 := minVal(a.Y+a.Height, b.Y+b.Height)

	intersectArea := maxVal(0, x2-x1) * maxVal(0, y2-y1)
	aArea := a.Width * a.Height
	bArea := b.Width * b.Height

	return float32(intersectArea) / float32(minVal(aArea, bArea))
}

// 辅助函数
func maxVal(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minVal(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// 全局实例
var MultiPointMatcher = initMultiPointMatcher()

// initMultiPointMatcher 初始化并加载预定义模式
func initMultiPointMatcher() *MultiPointColorMatcher {
	matcher := NewMultiPointColorMatcher()
	return matcher
}

// MultiPointMatcher 简化调用函数，搜索所有预定义模式
// 如果提供x1,y1,x2,y2参数，则在指定区域搜索；否则全屏搜索
func (c *ColorHandler) MultiPointMatcher(coords ...int) (Results, error) {
	startTime := time.Now() // 开始计时

	var targets []string

	// 搜索所有预定义模式
	for name := range defaultPatterns {
		targets = append(targets, name)
	}

	// 获取当前屏幕截图
	var img image.Image
	var err error
	var searchRect *image.Rectangle
	var offsetX, offsetY int

	if len(coords) == 4 {
		// 有坐标参数，截取指定区域
		offsetX, offsetY = coords[0], coords[1]
		img, err = c.GetScreenshotRegion(coords[0], coords[1], coords[2], coords[3])
		if err != nil {
			return Results{}, fmt.Errorf("获取区域截图失败: %v", err)
		}
		// 对于区域截图，搜索范围就是整个截图
		searchRect = nil
	} else {
		// 没有坐标参数，全屏截图
		offsetX, offsetY = 0, 0
		img, err = c.GetScreenshot()
		if err != nil {
			return Results{}, fmt.Errorf("获取截图失败: %v", err)
		}
		searchRect = nil
	}

	results, err := MultiPointMatcher.FindPatterns(targets, searchRect, img)

	// 如果是区域搜索，需要调整坐标偏移
	if len(coords) == 4 {
		for i := range results.Detections {
			results.Detections[i].X += offsetX
			results.Detections[i].Y += offsetY
			results.Detections[i].CenterX += offsetX
			results.Detections[i].CenterY += offsetY
		}
	}

	// 计算耗时
	duration := time.Since(startTime)
	fmt.Printf("多点识色执行时长: %v (找到 %d 个目标)\n", duration, results.Count)

	return results, err
}

// GetScreenshot 获取屏幕截图
func (c *ColorHandler) GetScreenshot() (image.Image, error) {
	// 截取主屏幕全屏
	img := images.CaptureScreen(0, 0, 0, 0, 0)
	if img == nil {
		return nil, fmt.Errorf("截图失败，返回空图像")
	}
	return img, nil
}

// GetScreenshotRegion 获取指定区域的屏幕截图
func (c *ColorHandler) GetScreenshotRegion(x1, y1, x2, y2 int) (image.Image, error) {
	// 截取指定区域
	img := images.CaptureScreen(x1, y1, x2, y2, 0)
	if img == nil {
		return nil, fmt.Errorf("区域截图失败，返回空图像")
	}
	return img, nil
}
