package util

import (
	"app/core"
	"fmt"

	"github.com/Dasongzi1366/AutoGo/images"
)

// CachedItemImage 缓存的道具图像信息
type CachedItemImage struct {
	X1, Y1, X2, Y2 int    // 缓存区域坐标
	ImageData      []byte // 图像数据
}

// CacheItemImage 缓存指定区域的图像
// x1, y1: 原始坐标点
// offsetX1, offsetY1: 相对于x1,y1的偏移量（左上角）
// offsetX2, offsetY2: 相对于x1,y1的偏移量（右下角）
// 返回缓存后的图像信息，如果失败返回nil
func CacheItemImage(x1, y1, offsetX1, offsetY1, offsetX2, offsetY2 int) *CachedItemImage {
	cacheX1 := x1 + offsetX1
	cacheY1 := y1 + offsetY1
	cacheX2 := x1 + offsetX2
	cacheY2 := y1 + offsetY2

	// 验证区域坐标有效性
	if cacheX2 <= cacheX1 || cacheY2 <= cacheY1 {
		return nil // 区域无效：宽度或高度 <= 0
	}

	// 验证区域尺寸合理性（至少要有一定大小）
	width := cacheX2 - cacheX1
	height := cacheY2 - cacheY1
	if width < 5 || height < 5 {
		return nil // 区域太小，无法有效匹配
	}

	// 验证坐标不能为负数（至少左上角不能太小）
	if cacheX1 < 0 || cacheY1 < 0 {
		return nil // 坐标无效
	}

	// 截取图像区域
	img := images.CaptureScreen(cacheX1, cacheY1, cacheX2, cacheY2, 0)
	if img == nil {
		return nil
	}

	// 将图像编码为字节数组
	imageData := images.EncodeToBytes(img, "png", 100)
	if len(imageData) == 0 {
		return nil
	}

	// 返回缓存信息
	return &CachedItemImage{
		X1:        cacheX1,
		Y1:        cacheY1,
		X2:        cacheX2,
		Y2:        cacheY2,
		ImageData: imageData,
	}
}

// CacheItemImageWithLog 缓存指定区域的图像并记录日志
// x1, y1: 原始坐标点
// offsetX1, offsetY1: 相对于x1,y1的偏移量（左上角）
// offsetX2, offsetY2: 相对于x1,y1的偏移量（右下角）
// 返回缓存后的图像信息，如果失败返回nil
func CacheItemImageWithLog(x1, y1, offsetX1, offsetY1, offsetX2, offsetY2 int) *CachedItemImage {
	cached := CacheItemImage(x1, y1, offsetX1, offsetY1, offsetX2, offsetY2)
	if cached != nil {
		fmt.Printf("已缓存不需要收藏的道具图像，位置: (%d, %d, %d, %d)\n", cached.X1, cached.Y1, cached.X2, cached.Y2)
	}
	return cached
}

// IsCachedImageMatch 检查指定区域是否匹配缓存图像
// searchX1, searchY1, searchX2, searchY2: 搜索区域坐标
// cachedImages: 缓存的图像数组
// expandRange: 搜索范围扩展（±expandRange）
// similarity: 相似度阈值（0-1）
// 返回是否找到匹配的缓存图像
func IsCachedImageMatch(searchX1, searchY1, searchX2, searchY2 int, cachedImages []CachedItemImage, expandRange int, similarity float64) bool {
	if len(cachedImages) == 0 {
		return false
	}

	// 验证搜索区域坐标有效性
	if searchX2 <= searchX1 || searchY2 <= searchY1 {
		return false // 区域无效：宽度或高度 <= 0
	}

	// 计算区域大小
	regionWidth := searchX2 - searchX1
	regionHeight := searchY2 - searchY1

	// 验证区域尺寸合理性
	if regionWidth < 5 || regionHeight < 5 {
		return false // 区域太小，无法有效匹配
	}

	// 在 expandRange 范围内搜索匹配
	opencvHandler := core.NewOpenCVHandler()

	for offsetX := -expandRange; offsetX <= expandRange; offsetX++ {
		for offsetY := -expandRange; offsetY <= expandRange; offsetY++ {
			// 计算当前搜索区域（保持与缓存图像相同的尺寸）
			currentX1 := searchX1 + offsetX
			currentY1 := searchY1 + offsetY
			currentX2 := currentX1 + regionWidth
			currentY2 := currentY1 + regionHeight

			// 验证当前搜索区域坐标有效性
			if currentX2 <= currentX1 || currentY2 <= currentY1 {
				continue // 区域无效，跳过
			}

			// 验证坐标不能为负数
			if currentX1 < 0 || currentY1 < 0 {
				continue // 坐标无效，跳过
			}

			// 截取当前屏幕区域
			currentImg := images.CaptureScreen(currentX1, currentY1, currentX2, currentY2, 0)
			if currentImg == nil {
				continue
			}

			// 将当前图像编码为字节数组
			currentImageData := images.EncodeToBytes(currentImg, "png", 100)
			if len(currentImageData) == 0 {
				continue
			}

			// 与每个缓存图像比较
			for _, cached := range cachedImages {
				// 验证缓存图像数据有效性
				if len(cached.ImageData) == 0 {
					continue
				}

				// 计算缓存图像的尺寸
				cachedWidth := cached.X2 - cached.X1
				cachedHeight := cached.Y2 - cached.Y1

				// 确保当前搜索区域尺寸 >= 缓存图像尺寸（OpenCV matchTemplate 要求）
				if regionWidth < cachedWidth || regionHeight < cachedHeight {
					continue // 搜索区域小于缓存图像，无法匹配
				}

				// 计算相似度（需要两个图像尺寸相同）
				sim := opencvHandler.CalculateImageSimilarity(currentX1, currentY1, currentX2, currentY2, cached.ImageData, currentImageData)
				if sim >= similarity {
					return true
				}
			}
		}
	}

	return false
}
