package core

import (
	"app/assets"
	"bytes"
	"fmt"
	"image"
	"image/color"
	"math"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/Dasongzi1366/AutoGo/utils"

	"github.com/Dasongzi1366/AutoGo/images"
	"github.com/Dasongzi1366/AutoGo/opencv"
)

// 用于保护缓存的互斥锁
var cacheMutex sync.RWMutex

// 用于保护OpenCV操作的全局互斥锁
var opencvMutex sync.Mutex

// maxTemplateMatCacheEntries 限制 byte2mat 全局模板缓存条数；超出则淘汰一条已有项，避免挂机久了模板 Native 内存只增不减。
const maxTemplateMatCacheEntries = 256

func closeTemplateMatPairUnlocked(key string) {
	if m, ok := assets.TemplateMap[key]; ok {
		if !m.Empty() {
			m.Close()
		}
		delete(assets.TemplateMap, key)
	}
	if m, ok := assets.MaskMap[key]; ok {
		if !m.Empty() {
			m.Close()
		}
		delete(assets.MaskMap, key)
	}
}

// 调用方需持有 cacheMutex 写锁。
func evictArbitraryTemplateMatUnlocked() {
	for k := range assets.TemplateMap {
		closeTemplateMatPairUnlocked(k)
		return
	}
}

func makeRoomForNewTemplateMatUnlocked(sign string) {
	if _, exists := assets.TemplateMap[sign]; exists {
		return
	}
	for len(assets.TemplateMap) >= maxTemplateMatCacheEntries {
		evictArbitraryTemplateMatUnlocked()
	}
}

// OpenCVHandler 处理OpenCV相关操作
type OpenCVHandler struct{}

// ImageDetectResult 图像检测结果
type ImageDetectResult struct {
	PicName string  // 匹配的图片名称
	X       int     // 匹配位置X坐标
	Y       int     // 匹配位置Y坐标
	Score   float32 // 匹配度得分
}

// NewOpenCVHandler 创建一个新的OpenCVHandler实例
func NewOpenCVHandler() *OpenCVHandler {
	return &OpenCVHandler{}
}

// ClearTemplateCache 清空模板缓存
func (h *OpenCVHandler) ClearTemplateCache() {
	opencvMutex.Lock()
	defer opencvMutex.Unlock()

	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	// 关闭所有缓存的Mat对象
	for _, mat := range assets.TemplateMap {
		if !mat.Empty() {
			mat.Close()
		}
	}
	for _, mat := range assets.MaskMap {
		if !mat.Empty() {
			mat.Close()
		}
	}

	// 清空缓存映射
	assets.TemplateMap = make(map[string]opencv.Mat)
	assets.MaskMap = make(map[string]opencv.Mat)

	fmt.Println("模板缓存已清空")
}

// FindImage 在指定区域内查找单个图像匹配，支持逗号分隔的多个图片
func (h *OpenCVHandler) FindImage(x1, y1, x2, y2 int, picName string, isGray bool, scale, sim float32) (int, int) {
	opencvMutex.Lock()
	defer opencvMutex.Unlock()

	if scale < 0.1 {
		scale = 1
	}

	// 解析逗号分隔的图片名称
	picNames := strings.Split(picName, ",")
	for i := range picNames {
		picNames[i] = strings.TrimSpace(picNames[i])
	}

	// 截取屏幕图像一次，避免重复截图
	img := images.CaptureScreen(x1, y1, x2, y2, 0)
	if img == nil {
		return -1, -1
	}

	bounds := img.Bounds()
	mat1, err := opencv.NewMatFromBytes(bounds.Dy(), bounds.Dx(), opencv.MatTypeCV8UC4, img.Pix)
	defer mat1.Close()
	if err != nil {
		return -1, -1
	}

	if isGray {
		originalMat1 := mat1
		mat1 = matGray(mat1)
		originalMat1.Close() // 关闭原始的mat1，使用新的灰度图mat1
	}

	// 依次检测每个图片，第一个识别到就返回
	for _, currentPicName := range picNames {
		template, err := assets.ImageFile.ReadFile(currentPicName)
		if err != nil {
			continue // 跳过无法读取的图片
		}

		mat2, mat3 := byte2mat(&template, isGray, scale, currentPicName)
		if mat2.Empty() {
			continue // 跳过空的模板
		}

		result := opencv.NewMat()
		opencv.MatchTemplate(mat1, mat2, &result, opencv.TmCcoeffNormed, mat3)
		_, maxVal, _, maxLoc := opencv.MinMaxLoc(result)
		hit := maxVal >= 0.5+sim*0.5
		result.Close()
		if hit {
			return int(float32(maxLoc.X)/scale) + x1, int(float32(maxLoc.Y)/scale) + y1
		}
	}

	return -1, -1
}

// FindImageAll 在指定区域内查找所有图像匹配
func (h *OpenCVHandler) FindImageAll(x1, y1, x2, y2 int, picName string, isGray bool, scale, sim float32) []map[string]interface{} {
	opencvMutex.Lock()
	defer opencvMutex.Unlock()

	if scale < 0.1 {
		scale = 1
	}

	resultT := make([]map[string]interface{}, 0)

	template, err := assets.ImageFile.ReadFile(picName)
	if err != nil {
		SLS_Log2(fmt.Sprintf("读取模板图片失败: %v", err))
		return nil
	}

	mat2, mat3 := byte2mat(&template, isGray, scale, picName)
	if mat2.Empty() {
		fmt.Println("模板图片为空")
		return nil
	}

	img := images.CaptureScreen(x1, y1, x2, y2, 0)
	if img == nil {
		return nil
	}

	bounds := img.Bounds()
	mat1, err := opencv.NewMatFromBytes(bounds.Dy(), bounds.Dx(), opencv.MatTypeCV8UC4, img.Pix)
	defer mat1.Close()
	if err != nil {
		SLS_Log2(fmt.Sprintf("屏幕图片转换为 Mat 失败: %v", err))
		return nil
	}

	if isGray {
		originalMat1 := mat1
		mat1 = matGray(mat1)
		originalMat1.Close() // 关闭原始的mat1，使用新的灰度图mat1
	}

	result := opencv.NewMat()
	defer result.Close()

	if !mat2.Empty() {
		name := extractFileName(picName)
		for {
			opencv.MatchTemplate(mat1, mat2, &result, opencv.TmCcoeffNormed, mat3)
			_, maxVal, _, maxLoc := opencv.MinMaxLoc(result)

			if maxVal >= sim {
				x, y := int(float32(maxLoc.X)/scale), int(float32(maxLoc.Y)/scale)
				rect := image.Rectangle{
					Min: image.Point{X: x, Y: y},
					Max: image.Point{X: x + mat2.Cols(), Y: y + mat2.Rows()},
				}

				opencv.Rectangle(&mat1, rect, color.RGBA{R: 0, G: 255, B: 0, A: 255}, -1)
				resultT = append(resultT, map[string]interface{}{
					"name": name,
					"x1":   x,
					"y1":   y,
					"x2":   x + mat2.Cols(),
					"y2":   y + mat2.Rows(),
					"zx":   x + mat2.Cols()/2,
					"zy":   y + mat2.Rows()/2,
					"sim":  maxVal,
				})
			} else {
				break
			}
		}
	}

	if len(resultT) > 0 {
		return resultT
	}
	return nil
}

// byte2mat 将字节数据转换为OpenCV矩阵
func byte2mat(pngData *[]byte, isGray bool, scale float32, picName string) (opencv.Mat, opencv.Mat) {
	// 输入参数验证
	if pngData == nil || len(*pngData) == 0 {
		fmt.Println("byte2mat: 输入数据为空：" + picName)
		return opencv.NewMat(), opencv.NewMat()
	}

	// 使用文件路径作为缓存键，而不是指针地址
	sign := fmt.Sprintf("%s-%t-%.2f", picName, isGray, scale)

	// 先尝试读取缓存
	cacheMutex.RLock()
	if cachedMat, ok := assets.TemplateMap[sign]; ok {
		// 验证缓存的Mat是否有效
		if !cachedMat.Empty() {
			maskMat := assets.MaskMap[sign]
			cacheMutex.RUnlock()
			return cachedMat, maskMat
		} else {
			// 如果缓存的Mat无效，清理并重新创建
			delete(assets.TemplateMap, sign)
			delete(assets.MaskMap, sign)
		}
	}
	cacheMutex.RUnlock()

	// 获取写锁来创建新的缓存条目
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	// 双重检查，避免并发创建同一个缓存项
	if cachedMat, ok := assets.TemplateMap[sign]; ok {
		if !cachedMat.Empty() {
			return cachedMat, assets.MaskMap[sign]
		}
	}

	img, _, err := image.Decode(bytes.NewReader(*pngData))
	if err != nil {
		SLS_Log2("图像解码失败")
		return opencv.NewMat(), opencv.NewMat()
	}
	imgNrgba := ImageToNRGBA(img)

	bounds := imgNrgba.Bounds()
	templateMat, _ := opencv.NewMatFromBytes(bounds.Dy(), bounds.Dx(), opencv.MatTypeCV8UC4, imgNrgba.Pix)

	isTransparent := checkTransparent(imgNrgba)

	if isGray {
		originalTemplateMat := templateMat
		templateMat = matGray(templateMat)
		originalTemplateMat.Close()
	}

	if scale != 1.0 {
		originalTemplateMat := templateMat
		templateMat = matScale(templateMat, scale)
		// 只有当scale确实改变了Mat时才关闭原始的
		if templateMat.Ptr() != originalTemplateMat.Ptr() {
			originalTemplateMat.Close()
		}
	}

	var maskMat opencv.Mat
	if isTransparent {
		maskMat = createMask(imgNrgba)
	} else {
		maskMat = opencv.NewMat()
	}

	makeRoomForNewTemplateMatUnlocked(sign)
	assets.TemplateMap[sign] = templateMat
	assets.MaskMap[sign] = maskMat

	return assets.TemplateMap[sign], assets.MaskMap[sign]
}

// matGray 将图像转换为灰度图
func matGray(mat opencv.Mat) opencv.Mat {
	if mat.Empty() {
		fmt.Println("matGray: 输入Mat为空")
		return opencv.NewMat()
	}

	grayMat := opencv.NewMat()
	opencv.CvtColor(mat, &grayMat, opencv.ColorBGRToGray)

	// 验证转换结果
	if grayMat.Empty() {
		SLS_Log2("matGray: 颜色转换失败")
		grayMat.Close()
		return opencv.NewMat()
	}

	// 只有当mat不是来自缓存时才关闭
	// 在这个上下文中，我们不应该关闭传入的mat，因为它可能来自缓存
	// 或者调用者还需要使用它
	return grayMat
}

// matScale 缩放图像
func matScale(mat opencv.Mat, scale float32) opencv.Mat {
	if mat.Empty() {
		fmt.Println("matScale: 输入Mat为空")
		return opencv.NewMat()
	}

	const epsilon = 1e-6
	if math.Abs(float64(scale-1)) < epsilon {
		return mat
	}

	fmt.Printf("缩放图像，比例: %.2f\n", scale)
	scaledMat := opencv.NewMat()
	newSize := image.Point{
		X: int(float32(mat.Cols()) * scale),
		Y: int(float32(mat.Rows()) * scale),
	}

	if newSize.X <= 0 || newSize.Y <= 0 {
		fmt.Printf("matScale: 无效的目标尺寸: %dx%d\n", newSize.X, newSize.Y)
		return opencv.NewMat()
	}

	opencv.Resize(mat, &scaledMat, newSize, 0, 0, opencv.InterpolationLinear)

	// 验证缩放结果
	if scaledMat.Empty() {
		SLS_Log2("matScale: 缩放失败")
		scaledMat.Close()
		return opencv.NewMat()
	}

	// 注意：这里不关闭传入的mat，因为它可能来自缓存或调用者还需要使用
	return scaledMat
}

// checkTransparent 检查图像是否为透明图
func checkTransparent(img image.Image) bool {
	width := img.Bounds().Dx()
	height := img.Bounds().Dy()

	if width < 2 || height < 2 {
		return false
	}

	c0 := getRGB(img.At(0, 0))
	c1 := getRGB(img.At(width-1, 0))
	c2 := getRGB(img.At(0, height-1))
	c3 := getRGB(img.At(width-1, height-1))

	if c0 != c1 || c0 != c2 || c0 != c3 {
		return false
	}

	transparentCount := 0
	totalPixels := width * height
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if getRGB(img.At(x, y)) == c0 {
				transparentCount++
			}
		}
	}

	if transparentCount >= int(float32(totalPixels)*0.3) && transparentCount < totalPixels {
		return true
	}

	return false
}

// createMask 创建透明图遮罩
func createMask(img image.Image) opencv.Mat {
	width := img.Bounds().Dx()
	height := img.Bounds().Dy()

	c0 := getRGB(img.At(0, 0))

	mask := opencv.NewMatWithSize(height, width, opencv.MatTypeCV8U)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if getRGB(img.At(x, y)) == c0 {
				mask.SetUCharAt(y, x, 1)
			} else {
				mask.SetUCharAt(y, x, 0)
			}
		}
	}

	return mask
}

// getRGB 获取RGB颜色值
func getRGB(c color.Color) color.RGBA {
	r, g, b, _ := c.RGBA() // 忽略 Alpha 通道
	return color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), 255}
}

// ImageToNRGBA 将图像转换为NRGBA格式
func ImageToNRGBA(img image.Image) *image.NRGBA {
	bounds := img.Bounds()
	nrgbaImg := image.NewNRGBA(bounds)

	// 简单的顺序处理，移除不必要的goroutine并发
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			srcColor := img.At(x, y)
			r, g, b, a := srcColor.RGBA()
			i := (y-bounds.Min.Y)*nrgbaImg.Stride + (x-bounds.Min.X)*4
			nrgbaImg.Pix[i] = uint8(r >> 8)
			nrgbaImg.Pix[i+1] = uint8(g >> 8)
			nrgbaImg.Pix[i+2] = uint8(b >> 8)
			nrgbaImg.Pix[i+3] = uint8(a >> 8)
		}
	}

	return nrgbaImg
}

// extractFileName 从路径中提取文件名
func extractFileName(path string) string {
	parts := strings.Split(path, "/")
	fileName := parts[len(parts)-1]
	return strings.TrimSuffix(fileName, filepath.Ext(fileName))
}

// WaitFor 等待在指定区域检测到目标图像，返回检测到的坐标和是否成功
// x1, y1, x2, y2: 检测区域坐标
// picName: 图片名称
// isGray: 是否灰度匹配
// scale: 缩放比例
// sim: 相似度阈值
// interval: 检测间隔时间，默认1秒
// maxAttempts: 最大检测次数，默认60次
func (h *OpenCVHandler) WaitFor(x1, y1, x2, y2 int, picName string, isGray bool, scale, sim float32, interval time.Duration, maxAttempts int, context string) (int, int, bool) {
	if interval <= 0 {
		interval = time.Second
	}
	if maxAttempts <= 0 {
		maxAttempts = 60
	}

	for i := 0; i < maxAttempts; i++ {
		x, y := h.FindImage(x1, y1, x2, y2, picName, isGray, scale, sim)
		if x != -1 && y != -1 {
			return x, y, true
		}
		time.Sleep(interval)
		utils.Toast(context, 300, 0, 1000)
	}
	return -1, -1, false
}

// ClickWhileExists 在指定区域内查找并点击目标图像，直到图像消失或达到最大点击次数
// x1, y1, x2, y2: 检测区域坐标
// picName: 图片名称
// isGray: 是否灰度匹配
// scale: 缩放比例
// sim: 相似度阈值
// interval: 检测间隔时间，默认1秒
// maxAttempts: 最大点击次数，默认60次
func (h *OpenCVHandler) ClickWhileExists(x1, y1, x2, y2 int, picName string, isGray bool, scale, sim float32, interval int, maxAttempts int) bool {
	if interval <= 0 {
		interval = 1000
	}
	if maxAttempts <= 0 {
		maxAttempts = 60
	}

	for i := 0; i < maxAttempts; i++ {
		x, y := h.FindImage(x1, y1, x2, y2, picName, isGray, scale, sim)
		if x != -1 && y != -1 {
			// 找到图像，点击该位置
			Click(x, y)
			Sleep(interval)
		} else {
			// 没有找到图像，停止点击
			return true
		}
	}
	return false
}

func (h *OpenCVHandler) ClickIfExist(x1, y1, x2, y2 int, picName string, isGray bool, scale, sim float32) {
	x, y := h.FindImage(x1, y1, x2, y2, picName, isGray, scale, sim)
	if x > 0 && y > 0 {
		RandomClickInArea(x, y, x+5, y+5)
		RandomSleep(1000, 1200)
	}
}

// FindMultipleImages 在指定区域内检测多个图片，返回检测到的第一个图片信息
// picNames: 图片名称数组
// 返回: 检测到的图片信息，未检测到返回空结果
func (h *OpenCVHandler) FindMultipleImages(x1, y1, x2, y2 int, picNames []string, isGray bool, scale, sim float32) ImageDetectResult {
	opencvMutex.Lock()
	defer opencvMutex.Unlock()

	if scale < 0.1 {
		scale = 1
	}

	// 截取屏幕图像一次，避免重复截图
	img := images.CaptureScreen(x1, y1, x2, y2, 0)
	if img == nil {
		return ImageDetectResult{}
	}

	bounds := img.Bounds()
	mat1, err := opencv.NewMatFromBytes(bounds.Dy(), bounds.Dx(), opencv.MatTypeCV8UC4, img.Pix)
	defer mat1.Close()
	if err != nil {
		SLS_Log2(fmt.Sprintf("屏幕图片转换为 Mat 失败: %v", err))
		return ImageDetectResult{}
	}

	if isGray {
		originalMat1 := mat1
		mat1 = matGray(mat1)
		originalMat1.Close() // 关闭原始的mat1，使用新的灰度图mat1
	}

	// 遍历每个图片进行检测
	for _, picName := range picNames {
		template, err := assets.ImageFile.ReadFile(picName)
		if err != nil {
			SLS_Log2(fmt.Sprintf("读取模板图片失败 %s: %v", picName, err))
			continue
		}

		mat2, mat3 := byte2mat(&template, isGray, scale, picName)
		if mat2.Empty() {
			//fmt.Printf("模板图片为空: %s\n", picName)
			continue
		}

		result := opencv.NewMat()
		// 注意：mat2和mat3来自缓存，不应该在这里关闭
		if !mat3.Empty() {
			opencv.MatchTemplate(mat1, mat2, &result, opencv.TmCcoeffNormed, mat3)
		} else {
			emptyMat := opencv.NewMat()
			opencv.MatchTemplate(mat1, mat2, &result, opencv.TmCcoeffNormed, emptyMat)
			emptyMat.Close()
		}
		_, maxVal, _, maxLoc := opencv.MinMaxLoc(result)
		if float32(maxVal) >= sim {
			result.Close()
			return ImageDetectResult{
				PicName: picName,
				X:       x1 + maxLoc.X,
				Y:       y1 + maxLoc.Y,
				Score:   float32(maxVal),
			}
		}
		result.Close()
	}

	// 没有找到任何匹配的图片
	return ImageDetectResult{}
}

// FindAllMultipleImages 在指定区域内检测多个图片，返回所有检测到的图片信息
// picNames: 图片名称数组
// 返回: 所有检测到的图片信息数组
func (h *OpenCVHandler) FindAllMultipleImages(x1, y1, x2, y2 int, picNames []string, isGray bool, scale, sim float32) []ImageDetectResult {
	opencvMutex.Lock()
	defer opencvMutex.Unlock()

	if scale < 0.1 {
		scale = 1
	}

	var results []ImageDetectResult

	// 截取屏幕图像一次，避免重复截图
	img := images.CaptureScreen(x1, y1, x2, y2, 0)
	if img == nil {
		return results
	}

	bounds := img.Bounds()
	mat1, err := opencv.NewMatFromBytes(bounds.Dy(), bounds.Dx(), opencv.MatTypeCV8UC4, img.Pix)
	defer mat1.Close()
	if err != nil {
		SLS_Log2(fmt.Sprintf("屏幕图片转换为 Mat 失败: %v", err))
		return results
	}

	if isGray {
		originalMat1 := mat1
		mat1 = matGray(mat1)
		originalMat1.Close() // 关闭原始的mat1，使用新的灰度图mat1
	}

	// 遍历每个图片进行检测
	for _, picName := range picNames {
		template, err := assets.ImageFile.ReadFile(picName)
		if err != nil {
			SLS_Log2(fmt.Sprintf("读取模板图片失败 %s: %v", picName, err))
			continue
		}

		mat2, mat3 := byte2mat(&template, isGray, scale, picName)
		if mat2.Empty() {
			//fmt.Printf("模板图片为空: %s\n", picName)
			continue
		}

		result := opencv.NewMat()
		if !mat3.Empty() {
			opencv.MatchTemplate(mat1, mat2, &result, opencv.TmCcoeffNormed, mat3)
		} else {
			emptyMat := opencv.NewMat()
			opencv.MatchTemplate(mat1, mat2, &result, opencv.TmCcoeffNormed, emptyMat)
			emptyMat.Close()
		}
		_, maxVal, _, maxLoc := opencv.MinMaxLoc(result)
		if float32(maxVal) >= sim {
			results = append(results, ImageDetectResult{
				PicName: picName,
				X:       x1 + maxLoc.X,
				Y:       y1 + maxLoc.Y,
				Score:   float32(maxVal),
			})
		}
		result.Close()
	}

	return results
}

// 全局移动检测缓存，key为区域坐标字符串，value为上一次的截图数据
var movementCache = make(map[string][]byte)
var movementCacheMutex sync.RWMutex

// CheckMovement 检查指定区域是否在移动（全局函数）
// 参数：x1, y1, x2, y2 检测区域坐标，threshold 相似度阈值
// 返回：true表示在移动，false表示静止
func CheckMovement(x1, y1, x2, y2 int, threshold float64) bool {
	// 生成缓存key
	cacheKey := fmt.Sprintf("%d,%d,%d,%d", x1, y1, x2, y2)

	// 截取当前区域
	img := images.CaptureScreen(x1, y1, x2, y2, 0)
	if img == nil {
		SLS_Log2("移动检测：截屏失败")
		return false
	}

	// 将图像转换为字节数组
	currentScreenshot := images.EncodeToBytes(img, "png", 100)
	if len(currentScreenshot) == 0 {
		SLS_Log2("移动检测：图像编码失败")
		return false
	}

	// 获取读锁检查缓存
	movementCacheMutex.RLock()
	lastScreenshot, exists := movementCache[cacheKey]
	movementCacheMutex.RUnlock()

	// 如果有上一次的截图，进行比较
	if exists && len(lastScreenshot) > 0 {
		handler := NewOpenCVHandler()
		similarity := handler.CalculateImageSimilarity(x1, y1, x2, y2, lastScreenshot, currentScreenshot)

		isMoving := similarity < threshold
		//fmt.Printf("移动检测：相似度=%.3f, 阈值=%.3f, 移动状态=%t\n", similarity, threshold, isMoving)

		// 更新缓存
		movementCacheMutex.Lock()
		movementCache[cacheKey] = currentScreenshot
		movementCacheMutex.Unlock()

		return isMoving
	}

	// 第一次检测，保存截图到缓存
	movementCacheMutex.Lock()
	movementCache[cacheKey] = currentScreenshot
	movementCacheMutex.Unlock()

	return true
}

// CalculateImageSimilarity 计算两个图像区域的相似度
// 返回相似度值，范围0-1，1表示完全相同
func (h *OpenCVHandler) CalculateImageSimilarity(x1, y1, x2, y2 int, img1, img2 []byte) float64 {
	opencvMutex.Lock()
	defer opencvMutex.Unlock()

	if len(img1) == 0 || len(img2) == 0 {
		return 0.0
	}

	// 创建Mat对象
	mat1, err1 := h.bytesToMat(img1)
	if err1 != nil || mat1.Empty() {
		return 0.0
	}
	defer mat1.Close()

	mat2, err2 := h.bytesToMat(img2)
	if err2 != nil || mat2.Empty() {
		return 0.0
	}
	defer mat2.Close()

	// 确保两个图像尺寸相同
	if mat1.Rows() != mat2.Rows() || mat1.Cols() != mat2.Cols() {
		return 0.0
	}

	// 使用模板匹配计算相似度
	result := opencv.NewMat()
	defer result.Close()

	// 创建空的mask
	emptyMask := opencv.NewMat()
	defer emptyMask.Close()

	opencv.MatchTemplate(mat1, mat2, &result, opencv.TmCcoeffNormed, emptyMask)
	_, maxVal, _, _ := opencv.MinMaxLoc(result)

	return float64(maxVal)
}

// bytesToMat 将字节数组转换为Mat对象
func (h *OpenCVHandler) bytesToMat(imgData []byte) (opencv.Mat, error) {
	img, _, err := image.Decode(bytes.NewReader(imgData))
	if err != nil {
		return opencv.NewMat(), err
	}

	nrgbaImg := ImageToNRGBA(img)
	bounds := nrgbaImg.Bounds()

	mat, err := opencv.NewMatFromBytes(bounds.Dy(), bounds.Dx(), opencv.MatTypeCV8UC4, nrgbaImg.Pix)
	return mat, err
}

// MovementDetector 移动检测器
type MovementDetector struct {
	x1, y1, x2, y2 int     // 检测区域坐标
	threshold      float64 // 相似度阈值，低于此值认为在移动
	lastScreenshot []byte  // 上一次的截图数据
}

// NewMovementDetector 创建新的移动检测器
func NewMovementDetector(x1, y1, x2, y2 int, threshold float64) *MovementDetector {
	return &MovementDetector{
		x1:        x1,
		y1:        y1,
		x2:        x2,
		y2:        y2,
		threshold: threshold,
	}
}

// CheckMovement 检查移动状态，返回是否在移动
func (md *MovementDetector) CheckMovement() bool {
	// 截取当前区域
	img := images.CaptureScreen(md.x1, md.y1, md.x2, md.y2, 0)
	if img == nil {
		SLS_Log2("移动检测：截屏失败")
		return false
	}

	// 将图像转换为字节数组
	currentScreenshot := images.EncodeToBytes(img, "png", 100)
	if len(currentScreenshot) == 0 {
		SLS_Log2("移动检测：图像编码失败")
		return false
	}

	// 如果有上一次的截图，进行比较
	if len(md.lastScreenshot) > 0 {
		handler := NewOpenCVHandler()
		similarity := handler.CalculateImageSimilarity(md.x1, md.y1, md.x2, md.y2, md.lastScreenshot, currentScreenshot)

		isMoving := similarity < md.threshold
		fmt.Printf("移动检测：相似度=%.3f, 阈值=%.3f, 移动状态=%t\n", similarity, md.threshold, isMoving)

		// 更新上一次的截图
		md.lastScreenshot = currentScreenshot
		return isMoving
	}

	// 第一次检测，保存截图
	md.lastScreenshot = currentScreenshot
	return false
}
