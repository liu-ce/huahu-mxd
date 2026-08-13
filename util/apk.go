package util

import (
	"app/core"

	"github.com/Dasongzi1366/AutoGo/utils"

	// "crypto/md5" // 暂时不校验MD5
	// "encoding/hex" // 暂时不校验MD5
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Dasongzi1366/AutoGo/files"
	"github.com/Dasongzi1366/AutoGo/storages"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// OSSDownloadConfig OSS下载配置结构体
type OSSDownloadConfig struct {
	Endpoint        string `json:"endpoint"`
	AccessKeyID     string `json:"access_key_id"`
	AccessKeySecret string `json:"access_key_secret"`
	Bucket          string `json:"bucket"`
	BaseURL         string `json:"base_url"`
}

// FileInfo 文件信息结构体
type FileInfo struct {
	ID          int    `json:"id"`
	DownloadURL string `json:"downloadUrl"`
	Name        string `json:"name"`
	FileSize    int64  `json:"fileSize"`
	MD5Hash     string `json:"md5Hash"`
	FileName    string `json:"fileName"`
}

// APKData APK数据结构体
type APKData struct {
	Game        string   `json:"game"`
	ARM64       FileInfo `json:"arm64"`
	X86         FileInfo `json:"x86"`
	APK         FileInfo `json:"apk"`
	VersionCode int      `json:"versionCode"`
}

// APKResponse APK更新响应结构体
type APKResponse struct {
	Code      int     `json:"code"`
	Message   string  `json:"message"`
	Data      APKData `json:"data"`
	Timestamp int64   `json:"timestamp"`
}

// APKCache APK缓存结构体
type APKCache struct {
	mu   sync.RWMutex
	Data *APKData
}

// APKManager APK管理器
type APKManager struct {
	httpClient *http.Client
	cache      *APKCache
}

var (
	apkManager *APKManager
	apkCache   = &APKCache{}
)

// NewAPKManager 创建新的APK管理器
func NewAPKManager() *APKManager {
	if apkManager == nil {
		apkManager = &APKManager{
			httpClient: &http.Client{
				Timeout: 120 * time.Second, // 增加API请求超时时间
			},
			cache: apkCache,
		}
	}
	return apkManager
}

// CheckUpdate 检查APK更新
func (a *APKManager) CheckUpdate(forceRefresh bool) (*APKResponse, error) {
	// 如果有缓存且不强制刷新，直接返回缓存
	if !forceRefresh {
		a.cache.mu.RLock()
		if a.cache.Data != nil {
			//fmt.Printf("使用缓存APK信息，版本号: %d\n",
			//	a.cache.Data.VersionCode)

			// 构造响应结构
			apkResp := &APKResponse{
				Code:    200,
				Message: "success",
				Data:    *a.cache.Data,
			}
			a.cache.mu.RUnlock()
			return apkResp, nil
		}
		a.cache.mu.RUnlock()
	}

	// 从配置中获取更新信息
	config := core.Config
	if config == nil {
		return nil, fmt.Errorf("配置未初始化")
	}

	// 获取服务器信息
	host := config.Get("server.host")
	updateEndpoint := config.Get("server.update_endpoint")
	game := config.Get("game")

	if host == "" || updateEndpoint == "" || game == "" {
		return nil, fmt.Errorf("配置信息不完整")
	}

	// 构建URL
	updateURL := fmt.Sprintf("%s%s?game=%s",
		host.(string),
		updateEndpoint.(string),
		url.QueryEscape(game.(string)))

	// 发起HTTP请求
	resp, err := a.httpClient.Get(updateURL)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("服务器返回错误状态码: %d", resp.StatusCode)
	}

	// 解析响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	var apkResp APKResponse
	if err := json.Unmarshal(body, &apkResp); err != nil {
		return nil, fmt.Errorf("解析JSON失败: %v", err)
	}

	// 缓存数据
	a.cache.mu.Lock()
	a.cache.Data = &apkResp.Data
	a.cache.mu.Unlock()

	return &apkResp, nil
}

// GetLatestAPK 获取最新APK信息（带缓存）
func (a *APKManager) GetLatestAPK() (*APKData, error) {
	resp, err := a.CheckUpdate(false)
	if err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

// ForceRefreshAPK 强制刷新APK信息
func (a *APKManager) ForceRefreshAPK() (*APKData, error) {
	resp, err := a.CheckUpdate(true)
	if err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

// IsUpdateRequired 检查是否需要强制更新
func (a *APKManager) IsUpdateRequired() (bool, error) {
	apkData, err := a.GetLatestAPK()
	if err != nil {
		return false, err
	}
	// 根据当前架构检查对应的文件是否可用
	arch := runtime.GOARCH
	if arch == "arm64" {
		return apkData.ARM64.ID > 0, nil
	} else if arch == "amd64" {
		return apkData.X86.ID > 0, nil
	}
	// 其他架构使用APK文件
	return apkData.APK.ID > 0, nil
}

// 全局函数，方便调用

// CheckAPKUpdate 检查APK更新（全局函数）
func CheckAPKUpdate() (*APKResponse, error) {
	manager := NewAPKManager()
	return manager.CheckUpdate(false)
}

// ForceCheckAPKUpdate 强制检查APK更新（全局函数）
func ForceCheckAPKUpdate() (*APKResponse, error) {
	manager := NewAPKManager()
	return manager.CheckUpdate(true)
}

// GetAPKInfo 获取APK信息（全局函数）
func GetAPKInfo() (*APKData, error) {
	manager := NewAPKManager()
	return manager.GetLatestAPK()
}

// DownloadOptimalBinary 根据架构下载最优的二进制文件
func DownloadOptimalBinary() error {
	// 获取APK信息
	apkData, err := GetAPKInfo()
	if err != nil {
		core.SLS_Log2("获取APK信息失败" + err.Error())
		return fmt.Errorf("获取APK信息失败: %v", err)
	}

	// 获取当前架构
	arch := runtime.GOARCH
	var targetFile FileInfo
	var fileType string

	// 根据架构选择合适的文件
	if arch == "arm64" && apkData.ARM64.ID > 0 {
		targetFile = apkData.ARM64
		fileType = "ARM64二进制包"
	} else if arch == "amd64" && apkData.X86.ID > 0 {
		targetFile = apkData.X86
		fileType = "X86二进制包"
	} else {
		targetFile = apkData.APK
		fileType = "APK文件"
	}

	fmt.Printf("当前架构: %s，选择下载: %s\n", arch, fileType)
	fmt.Printf("文件大小: %d bytes, MD5: %s\n", targetFile.FileSize, targetFile.MD5Hash)

	// 下载文件到内存
	sizeInMB := float64(targetFile.FileSize) / (1024 * 1024)
	utils.Toast(fmt.Sprintf("开始下载%s (%.1f MB)...", fileType, sizeInMB), 0, 0, 1000)

	// 带重试机制的下载
	fileData, err := downloadFileFromURLWithRetry(targetFile.DownloadURL, targetFile.FileSize, 10, 30)
	if err != nil {
		core.SLS_Log(fmt.Sprintf("下载%s失败: %v", fileType, err))
		return fmt.Errorf("下载%s失败: %v", fileType, err)
	}

	// 短暂延迟让用户看到完成信息
	time.Sleep(1 * time.Second)

	// 验证MD5（统一转换为大写进行比较）
	// 暂时跳过MD5校验
	/*
		hash := md5.Sum(fileData)
		downloadedMD5 := strings.ToUpper(hex.EncodeToString(hash[:]))
		expectedMD5 := strings.ToUpper(targetFile.MD5Hash)

		fmt.Printf("下载文件MD5: %s\n", downloadedMD5)
		fmt.Printf("预期文件MD5: %s\n", expectedMD5)

		if downloadedMD5 != expectedMD5 {
			errorMsg := "❌ 文件MD5校验失败，更新中止"
			fmt.Println(errorMsg)
			utils.Toast(errorMsg)
			return fmt.Errorf("MD5校验失败")
		}
	*/
	fmt.Println("已跳过MD5校验")

	// 保存当前任务状态
	SaveCurrentTaskState()

	files.WriteBytes(os.Args[0], fileData)
	Restart("更新完成，正在重启...")

	return nil
}

// loadOSSDownloadConfig 加载OSS下载配置
func loadOSSDownloadConfig() (*OSSDownloadConfig, error) {
	// 从配置文件加载OSS配置
	config := core.Config
	if config == nil {
		return nil, fmt.Errorf("配置未初始化")
	}

	// 获取OSS配置
	endpoint := config.Get("oss.endpoint")
	accessKeyID := config.Get("oss.access_key_id")
	accessKeySecret := config.Get("oss.access_key_secret")
	bucket := config.Get("oss.bucket")

	if endpoint == nil || accessKeyID == nil || accessKeySecret == nil || bucket == nil {
		return nil, fmt.Errorf("OSS配置信息不完整")
	}

	// 构造BaseURL：https://{bucket}.{region}.aliyuncs.com/
	// 从 endpoint "oss-cn-beijing.aliyuncs.com" 中提取 region 部分
	endpointStr := endpoint.(string)
	region := strings.TrimPrefix(endpointStr, "oss-")
	region = strings.TrimSuffix(region, ".aliyuncs.com")
	baseURL := fmt.Sprintf("https://%s.oss-%s.aliyuncs.com/", bucket.(string), region)

	return &OSSDownloadConfig{
		Endpoint:        endpoint.(string),
		AccessKeyID:     accessKeyID.(string),
		AccessKeySecret: accessKeySecret.(string),
		Bucket:          bucket.(string),
		BaseURL:         baseURL,
	}, nil
}

// generateOSSDownloadURL 生成OSS下载URL（带签名）
func generateOSSDownloadURL(relativePath string) (string, error) {
	config, err := loadOSSDownloadConfig()
	if err != nil {
		return "", err
	}

	// 移除相对路径开头的斜杠（如果有）
	relativePath = strings.TrimPrefix(relativePath, "/")

	// 创建OSS客户端
	client, err := oss.New(config.Endpoint, config.AccessKeyID, config.AccessKeySecret)
	if err != nil {
		return "", fmt.Errorf("创建OSS客户端失败: %v", err)
	}

	// 获取Bucket
	bucket, err := client.Bucket(config.Bucket)
	if err != nil {
		return "", fmt.Errorf("获取Bucket失败: %v", err)
	}

	// 生成带签名的URL（有效期1小时）
	signedURL, err := bucket.SignURL(relativePath, oss.HTTPGet, 3600)
	if err != nil {
		return "", fmt.Errorf("生成签名URL失败: %v", err)
	}

	return signedURL, nil
}

// downloadFileFromURL 通过URL下载文件到内存（带进度显示）
func downloadFileFromURL(downloadURL string, expectedSize int64) ([]byte, error) {
	// 检查是否为相对路径，如果是则转换为完整URL
	if !strings.HasPrefix(downloadURL, "http://") && !strings.HasPrefix(downloadURL, "https://") {
		// 尝试从配置获取下载基础URL（123网盘）
		config := core.Config
		var baseURL string

		if config != nil {
			downloadBase := config.Get("download.base_url")
			if downloadBase != nil {
				baseURL = downloadBase.(string)
			}
		}

		if baseURL == "" {
			return nil, fmt.Errorf("未配置 download.base_url")
		}

		// URL编码相对路径并拼接
		// 对路径进行编码，但保留斜杠
		pathParts := strings.Split(downloadURL, "/")
		for i, part := range pathParts {
			pathParts[i] = url.PathEscape(part)
		}
		encodedPath := strings.Join(pathParts, "/")

		downloadURL = baseURL + encodedPath
		fmt.Printf("相对路径转换为: %s\n", downloadURL)
	}

	fmt.Println("url:" + downloadURL)

	// 创建HTTP请求
	req, err := http.NewRequest("GET", downloadURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建HTTP请求失败: %v", err)
	}

	// OSS私有文件已通过签名URL认证，直接下载即可
	// 如果是OSS URL或123网盘URL，可以设置一些必要的请求头
	if strings.Contains(downloadURL, "oss-cn-beijing.aliyuncs.com") || strings.Contains(downloadURL, "autorun-oss") || strings.Contains(downloadURL, "123pan.cn") {
		req.Header.Set("User-Agent", "AutoRun-APK-Updater/1.0")
	}

	// 创建HTTP客户端进行下载
	client := &http.Client{Timeout: 300 * time.Second} // 5分钟超时
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		core.SLS_Log("下载失败，状态码:" + strconv.Itoa(resp.StatusCode))
		return nil, fmt.Errorf("下载失败，状态码: %d", resp.StatusCode)
	}

	// 准备下载到内存
	var fileData []byte
	var downloadedSize int64 = 0
	startTime := time.Now()

	// 启动进度显示协程
	done := make(chan bool)
	go func() {
		ticker := time.NewTicker(200 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				if expectedSize > 0 && downloadedSize > 0 {
					progress := int((downloadedSize * 100) / expectedSize)
					if progress > 100 {
						progress = 100
					}

					downloadedMB := float64(downloadedSize) / (1024 * 1024)
					totalMB := float64(expectedSize) / (1024 * 1024)
					elapsed := time.Since(startTime).Seconds()
					speed := downloadedMB / elapsed

					utils.Toast(fmt.Sprintf("下载中: %d%% (%.1f/%.1f MB) 速度: %.1f MB/s",
						progress, downloadedMB, totalMB, speed), 0, 0, 1000)
				} else {
					elapsed := time.Since(startTime).Seconds()
					utils.Toast(fmt.Sprintf("下载中... (已用时 %.0f 秒)", elapsed), 0, 0, 1000)
				}
			}
		}
	}()

	// 分块读取下载数据到内存
	buffer := make([]byte, 32*1024) // 32KB 缓冲区
	for {
		n, err := resp.Body.Read(buffer)
		if n > 0 {
			fileData = append(fileData, buffer[:n]...)
			downloadedSize += int64(n)
		}

		if err != nil {
			if err == io.EOF {
				break // 下载完成
			}
			done <- true
			return nil, fmt.Errorf("下载过程中出错: %v", err)
		}
	}

	// 停止进度显示
	done <- true

	// 显示下载完成信息
	downloadedSizeMB := float64(len(fileData)) / (1024 * 1024)
	elapsed := time.Since(startTime).Seconds()
	speed := downloadedSizeMB / elapsed

	utils.Toast(fmt.Sprintf("下载完成! %.1f MB，用时 %.1f 秒 (%.1f MB/s)",
		downloadedSizeMB, elapsed, speed), 0, 0, 1)

	return fileData, nil
}

// TaskState 任务状态结构
type TaskState struct {
	CurrentTaskName string `json:"current_task_name"`
	SaveTime        int64  `json:"save_time"`
}

// SaveCurrentTaskState 保存当前任务状态到存储
func SaveCurrentTaskState() {
	// 从当前任务状态storage获取信息
	taskName := storages.Get("task_state", "current_task_name")
	if taskName == "" {
		fmt.Println("没有正在运行的任务，跳过状态保存")
		return
	}

	// 只保存任务名称和保存时间
	storages.Put("task_resume", "task_name", taskName)
	storages.Put("task_resume", "save_time", strconv.FormatInt(time.Now().Unix(), 10))

	fmt.Printf("更新前保存任务状态: %s\n", taskName)
}

// RestoreTaskState 恢复任务状态
func RestoreTaskState() *TaskState {
	taskName := storages.Get("task_resume", "task_name")
	if taskName == "" {
		return nil
	}

	saveTimeStr := storages.Get("task_resume", "save_time")
	if saveTimeStr == "" {
		return nil
	}

	saveTime, err := strconv.ParseInt(saveTimeStr, 10, 64)
	if err != nil {
		return nil
	}

	// 检查保存时间是否在合理范围内（比如24小时内）
	if time.Now().Unix()-saveTime > 24*60*60 {
		fmt.Println("任务状态过期，跳过恢复")
		ClearTaskState()
		return nil
	}

	taskState := &TaskState{
		CurrentTaskName: taskName,
		SaveTime:        saveTime,
	}

	fmt.Printf("恢复任务状态: %s\n", taskState.CurrentTaskName)
	return taskState
}

// ClearTaskState 清除任务状态
func ClearTaskState() {
	storages.Remove("task_resume", "task_name")
	storages.Remove("task_resume", "save_time")
}

// CheckAndAutoUpdateWithForce 检查版本并自动更新（可强制刷新缓存）
func CheckAndAutoUpdateWithForce(forceRefresh bool) {
	var apkData *APKData
	var err error

	if forceRefresh {
		// 强制刷新，忽略缓存
		manager := NewAPKManager()
		apkData, err = manager.ForceRefreshAPK()
	} else {
		// 使用缓存
		apkData, err = GetAPKInfo()
	}

	if err != nil {
		core.SLS_Log("检查APK更新失败: " + err.Error())
	} else {
		localVersion := int(core.Config.Get("version").(float64))
		if apkData.VersionCode > localVersion {
			core.SLS("发现新版本! 远程 " + strconv.Itoa(apkData.VersionCode) + " > 本地 " + strconv.Itoa(localVersion))
			DownloadOptimalBinary()
		}
	}
}

// ClearAPKCache 清除APK缓存（用于测试或强制重新获取）
func ClearAPKCache() {
	manager := NewAPKManager()
	manager.cache.mu.Lock()
	manager.cache.Data = nil
	manager.cache.mu.Unlock()
}

// downloadFileFromURLWithRetry 带重试机制的文件下载函数
// url: 下载地址
// expectedSize: 预期文件大小
// maxRetries: 最大重试次数
// retryDelaySeconds: 重试间隔秒数
func downloadFileFromURLWithRetry(url string, expectedSize int64, maxRetries int, retryDelaySeconds int) ([]byte, error) {
	//var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		if attempt > 1 {
			core.SLS_Log(fmt.Sprintf("第%d次重试下载，等待%d秒...", attempt-1, retryDelaySeconds))
			time.Sleep(time.Duration(retryDelaySeconds) * time.Second)
		}

		core.SLS_Log(fmt.Sprintf("开始下载 (尝试 %d/%d)", attempt, maxRetries))

		fileData, err := downloadFileFromURL(url, expectedSize)
		if err == nil {
			// 下载成功
			if attempt > 1 {
				core.SLS_Log(fmt.Sprintf("重试成功！第%d次尝试下载完成", attempt))
			}
			return fileData, nil
		}

		// 下载失败，记录错误
		//lastErr = err
		core.SLS_Log(fmt.Sprintf("第%d次下载失败: %v", attempt, err))

		if attempt < maxRetries {
			core.SLS_Log(fmt.Sprintf("将在%d秒后进行第%d次重试", retryDelaySeconds, attempt))
		}
	}

	// 所有重试都失败了
	return nil, fmt.Errorf("下载失败，已重试%d次 请关闭代理或设置分应用代理重试", maxRetries)
}
