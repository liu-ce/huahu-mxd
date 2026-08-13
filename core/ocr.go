package core

import (
	"app/TomatoOCR"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/Dasongzi1366/AutoGo/utils"

	"github.com/Dasongzi1366/AutoGo/storages"
)

// 用于保护OCR操作的全局互斥锁
var ocrMutex sync.Mutex

// OCRHandler 处理OCR相关操作
type OCRHandler struct {
	client *TomatoOCR.Client
	mu     sync.Mutex
	inited bool
}

// NewOCRHandler 创建一个新的OCRHandler实例
func NewOCRHandler() *OCRHandler {
	return &OCRHandler{}
}

// initClient 初始化OCR客户端（懒加载）
func (h *OCRHandler) initClient() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.inited {
		return nil
	}

	username := storages.Get("data", "username")
	windowId := storages.Get("data", "windowId")
	config := TomatoOCR.Config{
		LicenseKey: Config.GetString("ocr.license_key"),
		Remark:     "user" + username + "-" + windowId,
	}

	client, err := TomatoOCR.NewClient(config)
	if err != nil {
		return err
	}

	h.client = client
	h.inited = true
	return nil
}

// InitClient 公开的初始化方法，用于在程序启动时初始化OCR客户端
func (h *OCRHandler) InitClient() error {
	return h.initClient()
}

// DetectText 在指定区域识别文字，返回识别到的第一个文字内容，识别不到返回空字符串
// 可选参数 recType 用于设置识别语言/模型类型，如 "ch-3.0", "cht", "japan", "korean" 等
// 用法: DetectText(x1, y1, x2, y2) 或 DetectText(x1, y1, x2, y2, "korean")
func (h *OCRHandler) DetectText(x1, y1, x2, y2 int, recType ...string) string {
	ocrMutex.Lock()
	defer ocrMutex.Unlock()

	// 安全检查：确保客户端已初始化且不为nil
	if !h.inited || h.client == nil {
		log.Println("OCR客户端未初始化或为nil，尝试重新初始化")
		if err := h.initClient(); err != nil {
			log.Printf("OCR客户端重新初始化失败: %v", err)
			return ""
		}
	}

	opts := TomatoOCR.DefaultDetectOptions()
	if len(recType) > 0 && recType[0] != "" {
		opts.RecType = recType[0]
	}
	results, err := h.client.DetectInArea(x1, y1, x2, y2, opts)
	if err != nil {
		log.Println("OCR识别出错: %v", err)
		return ""
	}

	if len(results) == 0 {
		return ""
	}

	// 返回第一个识别到的文字
	return results[0].Words
}

// DetectMultilineText 在指定区域识别所有文字，按从上到下顺序拼接后返回。
// 可选参数 recType 用于设置识别语言/模型类型，如 "korean" 等。
func (h *OCRHandler) DetectMultilineText(x1, y1, x2, y2 int, recType ...string) string {
	ocrMutex.Lock()
	defer ocrMutex.Unlock()

	if !h.inited || h.client == nil {
		log.Println("OCR客户端未初始化或为nil，尝试重新初始化")
		if err := h.initClient(); err != nil {
			log.Printf("OCR客户端重新初始化失败: %v", err)
			return ""
		}
	}

	opts := TomatoOCR.DefaultDetectOptions()
	if len(recType) > 0 && recType[0] != "" {
		opts.RecType = recType[0]
	}
	results, err := h.client.DetectInArea(x1, y1, x2, y2, opts)
	if err != nil {
		log.Println("OCR识别出错: %v", err)
		return ""
	}
	if len(results) == 0 {
		return ""
	}

	sort.Slice(results, func(i, j int) bool {
		return ocrResultTopY(results[i]) < ocrResultTopY(results[j])
	})

	var b strings.Builder
	for _, result := range results {
		if result.Words != "" {
			b.WriteString(result.Words)
		}
	}
	return b.String()
}

func ocrResultTopY(result TomatoOCR.DetectResult) int {
	minY := result.Location[0][1]
	for _, p := range result.Location {
		if p[1] < minY {
			minY = p[1]
		}
	}
	return minY
}

func parseDetectNumberResults(results []TomatoOCR.DetectResult) (int, bool) {
	if len(results) == 0 {
		return 0, false
	}
	raw := results[0].Words
	if raw == "" {
		return 0, false
	}
	if n, err := strconv.Atoi(strings.TrimSpace(raw)); err == nil {
		return n, true
	}
	var b strings.Builder
	for i, r := range raw {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
			continue
		}
		if r == '-' && i == 0 {
			b.WriteRune(r)
		}
	}
	numStr := b.String()
	if numStr == "" || numStr == "-" {
		return 0, false
	}
	n, err := strconv.Atoi(numStr)
	if err != nil {
		return 0, false
	}
	return n, true
}

func (h *OCRHandler) detectNumberWithOpts(x1, y1, x2, y2 int, opts TomatoOCR.DetectOptions) (int, bool) {
	results, err := h.client.DetectInArea(x1, y1, x2, y2, opts)
	if err != nil {
		log.Printf("OCR数字识别出错: %v", err)
		return 0, false
	}
	return parseDetectNumberResults(results)
}

// DetectNumber 在指定区域识别数字（recType=number，返回如 {"words":"89"}）。
// 返回: (number, ok)；ok=false 表示未识别到任何数字。
func (h *OCRHandler) DetectNumber(x1, y1, x2, y2 int, recType ...string) (int, bool) {
	ocrMutex.Lock()
	defer ocrMutex.Unlock()

	if !h.inited || h.client == nil {
		log.Println("OCR客户端未初始化或为nil，尝试重新初始化")
		if err := h.initClient(); err != nil {
			log.Printf("OCR客户端重新初始化失败: %v", err)
			return 0, false
		}
	}

	opts := TomatoOCR.NumberDetectOptions()
	if len(recType) > 0 && recType[0] != "" {
		opts.RecType = recType[0]
	}
	return h.detectNumberWithOpts(x1, y1, x2, y2, opts)
}

// DetectLevelNumber 人物等级专用数字 OCR（含背景滤色）。
func (h *OCRHandler) DetectLevelNumber(x1, y1, x2, y2 int) (int, bool) {
	ocrMutex.Lock()
	defer ocrMutex.Unlock()

	if !h.inited || h.client == nil {
		log.Println("OCR客户端未初始化或为nil，尝试重新初始化")
		if err := h.initClient(); err != nil {
			log.Printf("OCR客户端重新初始化失败: %v", err)
			return 0, false
		}
	}

	return h.detectNumberWithOpts(x1, y1, x2, y2, TomatoOCR.LevelNumberDetectOptions())
}

// DetectAllText 在指定区域识别所有文字，返回所有识别到的文字内容切片，识别不到返回 nil
func (h *OCRHandler) DetectAllText(x1, y1, x2, y2 int) []string {
	ocrMutex.Lock()
	defer ocrMutex.Unlock()

	if err := h.initClient(); err != nil {
		log.Printf("OCR客户端初始化失败: %v", err)
		return nil
	}

	opts := TomatoOCR.DefaultDetectOptions()
	results, err := h.client.DetectInArea(x1, y1, x2, y2, opts)
	if err != nil {
		log.Printf("OCR识别出错: %v", err)
		return nil
	}

	if len(results) == 0 {
		return nil
	}

	var texts []string
	for _, result := range results {
		texts = append(texts, result.Words)
	}

	return texts
}

// FindText 查找指定文字的坐标，返回 x, y 坐标，找不到返回 -1, -1
func (h *OCRHandler) FindText(text string) (int, int) {
	ocrMutex.Lock()
	defer ocrMutex.Unlock()

	if err := h.initClient(); err != nil {
		log.Printf("OCR客户端初始化失败: %v", err)
		return -1, -1
	}

	point, err := h.client.FindSingleTapPoint(text)
	if err != nil {
		return -1, -1
	}

	return point[0], point[1]
}

// FindTextAtArea 在指定区域内查找指定文字的坐标，返回 x, y 坐标，找不到返回 -1, -1
// targetText: 目标文字
// x1, y1, x2, y2: 搜索区域坐标
// similarity: 文本相似度阈值 (0.0-1.0)，默认0.8
// 可选参数 recType 用于设置识别语言/模型类型，如 "ch-3.0", "cht", "japan", "korean" 等
func (h *OCRHandler) FindTextAtArea(targetText string, x1, y1, x2, y2 int, similarity float32, recType ...string) (int, int) {
	ocrMutex.Lock()
	defer ocrMutex.Unlock()

	if similarity <= 0 {
		similarity = 0.8
	}

	if err := h.initClient(); err != nil {
		log.Printf("OCR客户端初始化失败: %v", err)
		return -1, -1
	}

	opts := TomatoOCR.DefaultDetectOptions()
	if len(recType) > 0 && recType[0] != "" {
		opts.RecType = recType[0]
	}

	results, err := h.client.DetectInArea(x1, y1, x2, y2, opts)
	if err != nil {
		log.Printf("OCR识别出错: %v", err)
		return -1, -1
	}

	if len(results) == 0 {
		return -1, -1
	}

	// 遍历所有识别结果，查找匹配的文字
	for _, result := range results {
		if result.Words != "" {
			textSimilarity := CalculateTextSimilarity(targetText, result.Words)
			if textSimilarity >= similarity {
				// 计算文字区域的中心坐标
				// Location 是 [4][2]int，表示矩形的4个顶点
				centerX := (result.Location[0][0] + result.Location[1][0] + result.Location[2][0] + result.Location[3][0]) / 4
				centerY := (result.Location[0][1] + result.Location[1][1] + result.Location[2][1] + result.Location[3][1]) / 4
				// 返回相对于搜索区域 (x1, y1) 的相对坐标
				return centerX + x1, centerY + y1
			}
		}
	}

	return -1, -1
}

// WaitFor 等待在指定区域检测到目标文字，返回是否检测到
// targetText: 目标文字
// x1, y1, x2, y2: 检测区域坐标
// similarity: 文本相似度阈值 (0.0-1.0)
// interval: 检测间隔时间，默认1秒
// maxAttempts: 最大检测次数，默认60次
func (h *OCRHandler) WaitFor(targetText string, x1, y1, x2, y2 int, similarity float32, interval int, maxAttempts int, context string) bool {
	if interval <= 0 {
		interval = 2000
	}
	if maxAttempts <= 0 {
		maxAttempts = 60
	}

	for i := 0; i < maxAttempts; i++ {
		detectedText := h.DetectText(x1, y1, x2, y2)
		if detectedText != "" {
			textSimilarity := CalculateTextSimilarity(targetText, detectedText)
			if textSimilarity >= similarity {
				return true
			}
		}

		if context != "" {
			utils.Toast(context, 0, 0, 1000)
		}

		Sleep(interval)
	}
	return false
}

// ClickWhileExists 在指定区域内查找并点击目标文字，直到文字消失或达到最大点击次数
// targetText: 目标文字
// x1, y1, x2, y2: 检测区域坐标
// similarity: 文本相似度阈值 (0.0-1.0)
// interval: 检测间隔时间，默认1秒
// maxAttempts: 最大点击次数，默认60次
func (h *OCRHandler) ClickWhileExists(targetText string, x1, y1, x2, y2 int, similarity float32, interval int, maxAttempts int) bool {
	if interval <= 0 {
		interval = 1000
	}
	if maxAttempts <= 0 {
		maxAttempts = 60
	}

	for i := 0; i < maxAttempts; i++ {
		detectedText := h.DetectText(x1, y1, x2, y2)
		if detectedText != "" {
			textSimilarity := CalculateTextSimilarity(targetText, detectedText)
			if textSimilarity >= similarity {
				// 找到文字，点击区域内随机坐标
				RandomClickInArea(x1, y1, x2, y2)
				RandomSleep(interval-200, interval+200)
			} else {
				// 没有找到文字，停止点击
				return true
			}
		} else {
			// 没有检测到任何文字，停止点击
			return true
		}
	}
	return false
}

// ClickIfTextExists 判断指定区域内是否存在目标文字，如果存在则点击并返回true，否则返回false
// targetText: 目标文字
// x1, y1, x2, y2: 检测区域坐标
// similarity: 文本相似度阈值 (0.0-1.0)，默认0.8
func (h *OCRHandler) ClickIfTextExists(targetText string, x1, y1, x2, y2 int, similarity float32) bool {
	ocrMutex.Lock()
	defer ocrMutex.Unlock()

	if similarity <= 0 {
		similarity = 0.8
	}

	if err := h.initClient(); err != nil {
		log.Printf("OCR客户端初始化失败: %v", err)
		return false
	}

	// 直接执行OCR检测，避免重复加锁
	opts := TomatoOCR.DefaultDetectOptions()
	results, err := h.client.DetectInArea(x1, y1, x2, y2, opts)
	if err != nil {
		log.Printf("OCR识别出错: %v", err)
		return false
	}

	if len(results) == 0 {
		return false
	}

	detectedText := results[0].Words
	if detectedText != "" {
		textSimilarity := CalculateTextSimilarity(targetText, detectedText)
		if textSimilarity >= similarity {
			// 找到匹配的文字，点击区域内随机坐标
			RandomClickInArea(x1, y1, x2, y2)
			return true
		}
	}

	return false
}

// RetryClickUntilOCRSuccess 重试点击直到OCR识别成功 - 通用重试函数
// ocrX1, ocrY1, ocrX2, ocrY2: OCR识别区域坐标
// expectedText: 期望识别到的文本
// clickX1, clickY1, clickX2, clickY2: 点击按钮的区域坐标
// maxAttempts: 最大重试次数
// 返回: true表示成功，false表示失败
func (h *OCRHandler) RetryClickUntilOCRSuccess(ocrX1, ocrY1, ocrX2, ocrY2 int, expectedText string, clickX1, clickY1, clickX2, clickY2 int, maxAttempts int) bool {
	for attempt := 0; attempt < maxAttempts; attempt++ {
		// 点击按钮
		RandomClickInArea(clickX1, clickY1, clickX2, clickY2)
		RandomSleep(2000, 3000)

		// OCR检测
		text := h.DetectText(ocrX1, ocrY1, ocrX2, ocrY2)
		if text == expectedText {
			return true // 成功
		}

		// 如果是最后一次尝试还失败
		if attempt == maxAttempts-1 {
			fmt.Printf("%d次尝试都未能识别到预期文本，期望: %s，实际识别到: %s\n", maxAttempts, expectedText, text)
			return false
		}

		SLS_Log2(fmt.Sprintf("第%d次尝试失败，期望: %s，实际识别到: %s，准备重试", attempt+1, expectedText, text))
		RandomSleep(1000, 2000) // 重试前稍等一下
	}
	return false
}

// CalculateTextSimilarity 计算两个文本的相似度，使用编辑距离算法
func CalculateTextSimilarity(text1, text2 string) float32 {
	// 去除空格和转换为小写，提高匹配准确性
	s1 := strings.ReplaceAll(strings.ToLower(text1), " ", "")
	s2 := strings.ReplaceAll(strings.ToLower(text2), " ", "")

	if s1 == s2 {
		return 1.0
	}

	if len(s1) == 0 || len(s2) == 0 {
		return 0.0
	}

	// 使用编辑距离算法计算相似度
	distance := levenshteinDistance(s1, s2)
	maxLen := max(len(s1), len(s2))

	similarity := 1.0 - float32(distance)/float32(maxLen)
	if similarity < 0 {
		similarity = 0
	}

	return similarity
}

// levenshteinDistance 计算两个字符串的编辑距离
func levenshteinDistance(s1, s2 string) int {
	len1, len2 := len(s1), len(s2)
	if len1 == 0 {
		return len2
	}
	if len2 == 0 {
		return len1
	}

	// 创建距离矩阵
	matrix := make([][]int, len1+1)
	for i := range matrix {
		matrix[i] = make([]int, len2+1)
	}

	// 初始化第一行和第一列
	for i := 0; i <= len1; i++ {
		matrix[i][0] = i
	}
	for j := 0; j <= len2; j++ {
		matrix[0][j] = j
	}

	// 填充矩阵
	for i := 1; i <= len1; i++ {
		for j := 1; j <= len2; j++ {
			cost := 0
			if s1[i-1] != s2[j-1] {
				cost = 1
			}

			matrix[i][j] = min(
				min(matrix[i-1][j]+1, matrix[i][j-1]+1), // 删除和插入的最小值
				matrix[i-1][j-1]+cost,                   // 替换
			)
		}
	}

	return matrix[len1][len2]
}

// min 返回两个整数中的较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// max 返回两个整数中的较大值
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
