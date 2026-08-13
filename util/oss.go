package util

import (
	"app/assets"
	"app/core"
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Dasongzi1366/AutoGo/images"
	"github.com/Dasongzi1366/AutoGo/storages"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// OSSConfig OSS配置结构体
type OSSConfig struct {
	AccessKeyID     string `json:"access_key_id"`
	AccessKeySecret string `json:"access_key_secret"`
	Endpoint        string `json:"endpoint"`
	Bucket          string `json:"bucket"`
	Region          string `json:"region"`
}

// OSSUploader OSS上传器
type OSSUploader struct {
	client *oss.Client
	bucket *oss.Bucket
	config *OSSConfig
}

// NewOSSUploader 创建OSS上传器
func NewOSSUploader() (*OSSUploader, error) {
	config, err := loadOSSConfig()
	if err != nil {
		return nil, err
	}

	client, err := oss.New(config.Endpoint, config.AccessKeyID, config.AccessKeySecret)
	if err != nil {
		return nil, err
	}

	bucket, err := client.Bucket(config.Bucket)
	if err != nil {
		return nil, err
	}

	return &OSSUploader{
		client: client,
		bucket: bucket,
		config: config,
	}, nil
}

// Upload 上传图片到OSS
func (u *OSSUploader) Upload(game, username, windowsID string, img *image.NRGBA, quality ...int) error {
	if img == nil {
		return fmt.Errorf("图片对象为空")
	}

	imgQuality := 10
	if len(quality) > 0 {
		imgQuality = quality[0]
	}

	// 将图片转换为JPEG字节数据
	var buf bytes.Buffer
	err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: imgQuality})
	if err != nil {
		return err
	}

	// 构建文件路径: capture/ymir/{username}/{windowsID}.jpg
	objectKey := fmt.Sprintf("capture/mxd/%s/%s.jpg", username, windowsID)

	// 上传文件到OSS
	err = u.bucket.PutObject(objectKey, bytes.NewReader(buf.Bytes()), oss.ContentType("image/jpeg"))
	return err
}

// UploadWithTimestamp 使用时间戳上传图片
func (u *OSSUploader) UploadWithTimestamp(game, username string, img *image.NRGBA, quality ...int) error {
	timestamp := time.Now().Unix()
	windowsID := strconv.FormatInt(timestamp, 10)
	return u.Upload(game, username, windowsID, img, quality...)
}

// UploadWithParams 使用传入的参数上传图片
func UploadWithParams(game, username, windowsID string, img *image.NRGBA, quality ...int) error {
	if img == nil {
		return fmt.Errorf("图片对象为空")
	}

	uploader, err := NewOSSUploader()
	if err != nil {
		return err
	}

	return uploader.Upload(game, username, windowsID, img, quality...)
}

// loadOSSConfig 从配置文件加载OSS配置
func loadOSSConfig() (*OSSConfig, error) {
	data, err := assets.ConfigFile.ReadFile("config/default.json")
	if err != nil {
		return nil, err
	}

	var config map[string]interface{}
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	ossConfigMap, ok := config["oss"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("未配置 oss")
	}

	return &OSSConfig{
		AccessKeyID:     ossConfigMap["access_key_id"].(string),
		AccessKeySecret: ossConfigMap["access_key_secret"].(string),
		Endpoint:        ossConfigMap["endpoint"].(string),
		Bucket:          ossConfigMap["bucket"].(string),
		Region:          ossConfigMap["region"].(string),
	}, nil
}

// UploadOSS 从core全局参数获取配置并上传截图
func UploadOSS() error {
	username, _ := core.Get("username").(string)
	windowID, _ := core.Get("seq").(string)
	if username == "" {
		username = "default"
	}

	if windowID == "" {
		windowID = "1"
	}

	img := images.CaptureScreen(0, 0, 0, 0, 0)
	if img == nil {
		return fmt.Errorf("截图失败")
	}

	// 正常上传
	return UploadWithParams("mxd", username, windowID, img)
}

// UploadHighQualityDebugScreenshot 上传高质量调试截图
// 质量100%，路径格式: capture/ymir_debug/yyyymmddhhmmssSSS_username_windowsid.jpg
func UploadHighQualityDebugScreenshot(suffix ...string) error {
	username, _ := core.Get("username").(string)
	windowID, _ := core.Get("seq").(string)
	if username == "" {
		username = "default"
	}

	if windowID == "" {
		windowID = "1"
	}

	// 截图
	img := images.CaptureScreen(0, 0, 0, 0, 0)
	if img == nil {
		return fmt.Errorf("高质量截图失败")
	}

	// 生成带毫秒的时间戳: yyyymmddhhmmssSSS (使用UTC+8北京时间)
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		// 如果加载时区失败，使用固定偏移UTC+8
		loc = time.FixedZone("CST", 8*3600)
	}
	now := time.Now().In(loc)
	timestamp := fmt.Sprintf("%04d%02d%02d%02d%02d%02d%03d",
		now.Year(), now.Month(), now.Day(),
		now.Hour(), now.Minute(), now.Second(),
		now.Nanosecond()/1000000) // 转换为毫秒

	// 可选文件名后缀
	extra := ""
	if len(suffix) > 0 && strings.TrimSpace(suffix[0]) != "" {
		extra = "_" + strings.TrimSpace(suffix[0])
	}

	// 构建文件名: yyyymmddhhmmssSSS_username_windowsid[_suffix].jpg
	filename := fmt.Sprintf("%s_%s_%s%s.jpg", timestamp, username, windowID, extra)

	// 构建完整路径: capture/ymir_debug/yyyymmddhhmmssSSS_username_windowsid.jpg
	objectKey := fmt.Sprintf("capture/mxd_debug/%s", filename)

	fmt.Printf("📸 准备上传高质量调试截图: %s\n", objectKey)

	// 正常上传高质量图片
	return uploadHighQualityImage(objectKey, img)
}

const (
	captchaScreenshotBurstDur      = 30 * time.Second
	captchaScreenshotBurstInterval = 3 * time.Second
)

// UploadGMPatrolScreenshotAsync GM巡逻告警时异步上传一张100%质量全屏截图。
func UploadGMPatrolScreenshotAsync() {
	go func() {
		if err := UploadHighQualityDebugScreenshot("gm_patrol"); err != nil {
			core.SLS_Log2NoToast("[截图] GM巡逻高清图上传失败: " + err.Error())
		} else {
			core.SLS_Log2NoToast("[截图] GM巡逻高清图已上传")
		}
	}()
}

// UploadAnswerScreenshot 上传 GM 巡逻答题最终截图，路径 answer-screenshots/YYYYMMDD/xxx.jpg。
func UploadAnswerScreenshot() (string, error) {
	username, _ := core.Get("username").(string)
	windowID := storages.Get("data", "windowId")
	if username == "" {
		username = "default"
	}
	if windowID == "" {
		windowID = "1"
	}

	img := images.CaptureScreen(0, 0, 0, 0, 0)
	if img == nil {
		return "", fmt.Errorf("截图失败")
	}

	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		loc = time.FixedZone("CST", 8*3600)
	}
	now := time.Now().In(loc)
	dateDir := fmt.Sprintf("%04d%02d%02d", now.Year(), now.Month(), now.Day())
	filename := fmt.Sprintf("%04d%02d%02d%02d%02d%02d%03d_%s_%s.jpg",
		now.Year(), now.Month(), now.Day(),
		now.Hour(), now.Minute(), now.Second(),
		now.Nanosecond()/1000000,
		username, windowID,
	)
	objectKey := fmt.Sprintf("answer-screenshots/%s/%s", dateDir, filename)

	if err := uploadHighQualityImage(objectKey, img); err != nil {
		return "", err
	}
	return objectKey, nil
}

// StartCaptchaScreenshotBurst 钉钉告警后 30 秒内每 3 秒上传一次调试截图，完成后退出脚本。
func StartCaptchaScreenshotBurst() {
	go func() {
		deadline := time.Now().Add(captchaScreenshotBurstDur)
		n := 0
		for time.Now().Before(deadline) {
			n++
			suffix := fmt.Sprintf("captcha_%d", n)
			if err := UploadHighQualityDebugScreenshot(suffix); err != nil {
				core.SLS_Log2("[截图] 测谎连拍失败: " + err.Error())
			} else {
				core.SLS_Log2(fmt.Sprintf("[截图] 测谎连拍 #%d 已上传", n))
			}
			time.Sleep(captchaScreenshotBurstInterval)
		}
		core.SLS_Log2(fmt.Sprintf("[截图] 测谎连拍完成 共%d张 脚本退出", n))
		os.Exit(0)
	}()
}

// UploadHighQualityDebugScreenshotLimited 上传高质量调试截图（每6小时最多2次）
// 时间段划分: 0-6, 6-12, 12-18, 18-24（UTC+8）
func UploadHighQualityDebugScreenshotLimited(suffix ...string) error {
	// 获取当前时间（UTC+8）
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		loc = time.FixedZone("CST", 8*3600)
	}
	now := time.Now().In(loc)

	// 计算当前时间段索引（每6小时一个段）
	slotIndex := now.Hour() / 12 // 0..3

	// 构建存储key: yyyymmdd_slotIndex，例如: 20241020_0, 20241020_1
	today := fmt.Sprintf("%04d%02d%02d", now.Year(), now.Month(), now.Day())
	storageKey := fmt.Sprintf("%s_%d", today, slotIndex)

	// 读取当前时间段的计数
	storedKey := storages.Get("debug_screenshot", "slot_key")
	storedCountStr := storages.Get("debug_screenshot", "slot_count")

	currentCount := 0
	if storedKey == storageKey && storedCountStr != "" {
		if v, err := strconv.Atoi(storedCountStr); err == nil {
			currentCount = v
		}
	} else {
		// 新时间段，重置
		storages.Put("debug_screenshot", "slot_key", storageKey)
		storages.Put("debug_screenshot", "slot_count", "0")
		currentCount = 0
	}

	// 限制：每6小时最多2次
	if currentCount >= 2 {
		fmt.Sprintf("当前12小时段(%s 第%d段)高质量截图已达上限: %d/2", today, slotIndex, currentCount)
		return fmt.Errorf("high quality screenshot limited: %s", storageKey)
	}

	// 调用实际上传
	if err := UploadHighQualityDebugScreenshot(suffix...); err != nil {
		return err
	}

	// 计数+1
	currentCount++
	storages.Put("debug_screenshot", "slot_key", storageKey)
	storages.Put("debug_screenshot", "slot_count", strconv.Itoa(currentCount))

	fmt.Printf("✅ 高质量截图次数 - %s 段%d: %d/2\n", today, slotIndex, currentCount)
	core.SLS_Log2(fmt.Sprintf("高质量截图上传成功 - %s 段%d: %d/2", today, slotIndex, currentCount))
	return nil
}

// uploadHighQualityImage 上传100%质量的图片
func uploadHighQualityImage(objectKey string, img *image.NRGBA) error {
	if img == nil {
		return fmt.Errorf("图片对象为空")
	}

	// 获取OSS配置
	config, err := loadOSSConfig()
	if err != nil {
		return fmt.Errorf("加载OSS配置失败: %v", err)
	}

	// 创建OSS客户端
	client, err := oss.New(config.Endpoint, config.AccessKeyID, config.AccessKeySecret)
	if err != nil {
		return fmt.Errorf("创建OSS客户端失败: %v", err)
	}

	bucket, err := client.Bucket(config.Bucket)
	if err != nil {
		return fmt.Errorf("获取Bucket失败: %v", err)
	}

	// 将图片编码为100%质量的JPEG
	var buf bytes.Buffer
	err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: 100})
	if err != nil {
		return fmt.Errorf("图片编码失败: %v", err)
	}

	// 上传到OSS
	err = bucket.PutObject(objectKey, bytes.NewReader(buf.Bytes()), oss.ContentType("image/jpeg"))
	if err != nil {
		return fmt.Errorf("上传OSS失败: %v", err)
	}

	fmt.Printf("✅ 高质量截图上传成功: %s (大小: %d KB)\n", objectKey, buf.Len()/1024)
	return nil
}
