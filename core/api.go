package core

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Dasongzi1366/AutoGo/device"
	"github.com/Dasongzi1366/AutoGo/system"
)

// testDNSResolution 测试DNS解析，支持重试
func testDNSResolution(urlStr string) error {
	u, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("解析URL失败: %v", err)
	}

	fmt.Printf("解析主机名: %s\n", u.Host)

	// 移除端口号（如果有）
	host := u.Host
	if strings.Contains(host, ":") {
		host, _, err = net.SplitHostPort(host)
		if err != nil {
			return fmt.Errorf("分割主机名和端口失败: %v", err)
		}
	}

	// DNS解析重试机制
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			waitTime := time.Duration(i) * time.Second // 1s, 2s
			fmt.Printf("DNS第%d次重试，等待%v...\n", i, waitTime)
			time.Sleep(waitTime)
		}

		fmt.Printf("DNS第%d次解析尝试\n", i+1)

		startTime := time.Now()
		ips, err := net.LookupHost(host)
		duration := time.Since(startTime)

		if err == nil {
			fmt.Printf("DNS解析成功，耗时%v，解析到IP: %v\n", duration, ips)
			return nil
		}

		SLS_Log2(fmt.Sprintf("DNS第%d次解析失败，耗时%v: %v", i+1, duration, err))

		// 如果是最后一次重试
		if i == maxRetries-1 {
			return fmt.Errorf("DNS解析最终失败，已重试%d次: %v", maxRetries, err)
		}
	}

	return nil
}

// testTCPConnection 测试TCP连接
func testTCPConnection(urlStr string) error {
	u, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("解析URL失败: %v", err)
	}

	// 确定端口
	port := u.Port()
	if port == "" {
		if u.Scheme == "https" {
			port = "443"
		} else {
			port = "80"
		}
	}

	address := net.JoinHostPort(u.Hostname(), port)
	fmt.Printf("测试TCP连接到: %s\n", address)

	startTime := time.Now()
	conn, err := net.DialTimeout("tcp", address, 15*time.Second)
	duration := time.Since(startTime)

	if err != nil {
		return fmt.Errorf("TCP连接失败，耗时%v: %v", duration, err)
	}

	conn.Close()
	fmt.Printf("TCP连接成功，耗时%v\n", duration)
	return nil
}

// extractDigits 提取字符串中的所有数字
func extractDigits(text string) string {
	digitsOnly := ""
	for _, r := range text {
		if r >= '0' && r <= '9' {
			digitsOnly += string(r)
		}
	}
	return digitsOnly
}

// APIClient API客户端
type APIClient struct {
	baseURL    string
	httpClient *http.Client
}

// LoginRequest 登录请求结构体
type LoginRequest struct {
	Username string `json:"username"` // 用户名不能为空
	Password string `json:"password"` // 密码不能为空
	DeviceID string `json:"deviceId"` // 设备ID不能为空
	WindowID string `json:"windowId"` // 窗口ID不能为空
	Game     string `json:"game"`     // 游戏名不能为空
}

// LoginResponse 登录响应结构体
type LoginResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Token    string `json:"token"`
		UserID   string `json:"userId"`
		Username string `json:"username"`
		RoleId   int    `json:"roleId"`
		Adb      string `json:"adb"`
	} `json:"data"`
}

// APIResponse 通用API响应结构体
type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// ConfigData 配置数据结构体
type ConfigData struct {
	ID                int                    `json:"id"`                  // 配置ID
	Name              string                 `json:"name"`                // 配置名称
	Game              string                 `json:"game"`                // 游戏名称
	Content           map[string]interface{} `json:"content"`             // 配置内容JSON
	Description       string                 `json:"description"`         // 配置描述
	Version           int                    `json:"version"`             // 配置版本号（更新次数）
	RoleID            int                    `json:"roleId"`              // 角色ID (如果需要的话)
	UserID            int                    `json:"userId"`              // 用户ID (如果需要的话)
	CreateAt          string                 `json:"createAt"`            // 创建时间
	UpdateAt          string                 `json:"updateAt"`            // 更新时间
	Websocket         string                 `json:"websocket"`           // 设备运行信息
	RoleName          string                 `json:"role_name"`           // 角色名称（服务端）
	VocationName      string                 `json:"vocation_name"`       // 职业名称（服务端）
	AIAnswerCustom    string                 `json:"ai_answer_custom"`    // GM 问答自定义参考
	AutoAnswerEnabled *bool                  `json:"auto_answer_enabled"` // GM 巡逻是否自动点确定；false 时仍 AI+输入+钉钉
}

// ConfigResponse 配置响应结构体
type ConfigResponse struct {
	Code    int        `json:"code"`
	Message string     `json:"message"`
	Data    ConfigData `json:"data"`
}

// NotificationData 通知数据结构体
type NotificationData struct {
	ID         int    `json:"id"`
	RoleID     int    `json:"roleId"`
	Type       string `json:"type"`
	State      int    `json:"state"`
	CreateTime string `json:"createTime"`
	UpdateTime string `json:"updateTime"`
}

// NotificationResponse 通知响应结构体
type NotificationResponse struct {
	Code      int                `json:"code"`
	Message   string             `json:"message"`
	Data      []NotificationData `json:"data"`
	Timestamp int64              `json:"timestamp"`
}

// RoleUpdateRequest 角色更新请求结构体
type RoleUpdateRequest struct {
	ID          int                    `json:"id"`                    // 角色ID (必填)
	Game        *string                `json:"game,omitempty"`        // 游戏名称
	DeviceID    *string                `json:"deviceId,omitempty"`    // 设备ID (IMEI)
	WindowID    *string                `json:"windowId,omitempty"`    // 窗口ID
	Name        *string                `json:"name,omitempty"`        // 角色名称
	Region      *string                `json:"region,omitempty"`      // 大区
	Level       *int                   `json:"level,omitempty"`       // 等级
	CombatPower *int64                 `json:"combatPower,omitempty"` // 战斗力
	Version     *string                `json:"version,omitempty"`     // 版本号
	ConfigID    *int                   `json:"configId,omitempty"`    // 配置ID
	CurTask     *string                `json:"curTask,omitempty"`     // 当前任务
	Websocket   *string                `json:"websocket,omitempty"`   // 设备运行信息
	Datas       map[string]interface{} `json:"datas,omitempty"`       // 其他数据 (JSON格式)
}

// 全局API客户端实例
var API *APIClient

// RoleInstance 全局角色实例，支持直接字段访问
type RoleInstance struct {
	ID          int                    // 角色ID
	Game        string                 // 游戏名称
	DeviceID    string                 // 设备ID
	WindowID    string                 // 窗口ID
	Name        string                 // 角色名称
	Region      string                 // 大区
	Level       int                    // 等级
	CombatPower int64                  // 战斗力
	Version     string                 // 版本号
	ConfigID    int                    // 配置ID
	CurTask     string                 // 当前任务
	Websocket   string                 // 设备运行信息
	Datas       map[string]interface{} // 其他数据
}

// toRequest 转换为API请求格式
func (r *RoleInstance) toRequest() *RoleUpdateRequest {
	var datasClone map[string]interface{}
	if len(r.Datas) > 0 {
		datasClone = make(map[string]interface{})
		for k, v := range r.Datas {
			datasClone[k] = v
		}
	}

	// 只有非零值才传递指针
	req := &RoleUpdateRequest{
		ID:    r.ID,
		Datas: datasClone,
	}

	if r.Game != "" {
		req.Game = &r.Game
	}
	if r.DeviceID != "" {
		req.DeviceID = &r.DeviceID
	}
	if r.WindowID != "" {
		req.WindowID = &r.WindowID
	}
	if r.Name != "" {
		req.Name = &r.Name
	}
	if r.Region != "" {
		req.Region = &r.Region
	}
	if r.Level != 0 {
		req.Level = &r.Level
	}
	if r.CombatPower != 0 {
		req.CombatPower = &r.CombatPower
	}
	if r.Version != "" {
		req.Version = &r.Version
	}
	if r.ConfigID != 0 {
		req.ConfigID = &r.ConfigID
	}

	if r.CurTask != "" {
		req.CurTask = &r.CurTask
	}
	if r.Websocket != "" {
		req.Websocket = &r.Websocket
	}

	return req
}

// toLightRequest 转换为轻量级API请求格式（只包含ID, Game, DeviceID, WindowID, CurTask）
func (r *RoleInstance) toLightRequest() *RoleUpdateRequest {
	req := &RoleUpdateRequest{
		ID: r.ID,
	}

	if r.Game != "" {
		req.Game = &r.Game
	}
	if r.DeviceID != "" {
		req.DeviceID = &r.DeviceID
	}
	if r.WindowID != "" {
		req.WindowID = &r.WindowID
	}
	if r.CurTask != "" {
		req.CurTask = &r.CurTask
	}

	return req
}

// IdentifyMemory 计算并更新内存使用百分比（整数）
func (r *RoleInstance) IdentifyMemory() int {
	totalMem := device.GetTotalMem()
	availMem := device.GetAvailMem()

	if totalMem <= 0 {
		return 0
	}

	usedMem := totalMem - availMem
	percentage := (usedMem * 100) / totalMem

	// 存储到 Datas 中
	if r.Datas == nil {
		r.Datas = make(map[string]interface{})
	}
	r.Datas["memory_usage"] = int(percentage)

	return int(percentage)
}

// IdentifyCpuUsage 计算并更新CPU使用率（返回百分比整数）
func (r *RoleInstance) IdentifyCpuUsage() int {
	// 从配置中获取包名
	packageName := Config.GetString("app_packages.game")
	if packageName == "" {
		return 0
	}

	// 获取进程PID
	pid := system.GetPid(packageName)
	if pid <= 0 {
		return 0
	}

	// 获取CPU使用率
	cpuUsage := system.GetCpuUsage(pid)
	percentage := int(cpuUsage) / 8

	// 存储到 Datas 中
	if r.Datas == nil {
		r.Datas = make(map[string]interface{})
	}
	r.Datas["cpu_usage"] = percentage

	return percentage
}

// 识别名称 登录角色的时候识别
func (r *RoleInstance) IdentifyRoleName() {
	roleName := OCR.DetectText(91, 99, 289, 134)
	SLS_Log2("识别角色名称:" + roleName)
	r.Name = roleName
}

// 识别职业 登录角色的时候识别
func (r *RoleInstance) IdentifyVocationName() {
	vocationName := OCR.DetectText(207, 144, 323, 172)
	r.Datas["vocation_name"] = vocationName
}

// 全局角色实例
var roleInstance *RoleInstance

// 配置缓存
var configCache struct {
	mu      sync.RWMutex // 读写锁，保证并发安全
	Data    *ConfigData  // 缓存的配置数据
	ID      int          // 缓存的配置ID
	Version int          // 缓存的版本号
}

// NewAPIClient 创建新的API客户端
func NewAPIClient() *APIClient {
	// 从配置中获取服务器信息
	host := Config.GetString("server.host")
	baseURL := host
	fmt.Printf("API客户端初始化 - BaseURL: %s\n", baseURL)

	// 创建一个最简化的HTTP客户端
	simpleTransport := &http.Transport{
		// 最基本的连接配置
		DialContext: (&net.Dialer{
			Timeout:   15 * time.Second, // 缩短连接超时
			KeepAlive: 30 * time.Second,
		}).DialContext,

		// 简化的TLS配置
		TLSHandshakeTimeout: 15 * time.Second, // 缩短TLS握手超时
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // 跳过证书验证
		},

		// 禁用HTTP/2，强制使用HTTP/1.1
		ForceAttemptHTTP2: false,

		// 基本超时配置
		ResponseHeaderTimeout: 30 * time.Second,
		ExpectContinueTimeout: 5 * time.Second,

		// 简化连接池
		IdleConnTimeout:     60 * time.Second,
		DisableKeepAlives:   false,
		MaxIdleConns:        5,
		MaxIdleConnsPerHost: 2,
		MaxConnsPerHost:     5,

		// 禁用压缩
		DisableCompression: true,
	}

	client := &APIClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout:   45 * time.Second, // 缩短总超时时间
			Transport: simpleTransport,
		},
	}

	return client
}

// Post 通用POST请求方法
func (c *APIClient) Post(endpoint string, data interface{}) (*APIResponse, error) {
	// 序列化请求体
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("序列化请求数据失败: %v", err)
	}

	// 发送POST请求
	url := c.baseURL + endpoint
	resp, err := c.httpClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("发送POST请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 解析响应
	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		return &apiResp, fmt.Errorf("请求失败，状态码: %d, 消息: %s", resp.StatusCode, apiResp.Message)
	}

	return &apiResp, nil
}

// PostWithToken 带Token认证的POST请求方法
func (c *APIClient) PostWithToken(endpoint string, data interface{}, token string) (*APIResponse, error) {
	// 序列化请求体
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("序列化请求数据失败: %v", err)
	}

	// 创建请求
	url := c.baseURL + endpoint
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	fmt.Println("Bearer " + token)
	// 发送请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送POST请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 解析响应
	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		return &apiResp, fmt.Errorf("请求失败，状态码: %d, 消息: %s", resp.StatusCode, apiResp.Message)
	}

	return &apiResp, nil
}

// Get 通用GET请求方法
func (c *APIClient) Get(endpoint string) (*APIResponse, error) {
	// 发送GET请求
	url := c.baseURL + endpoint
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("发送GET请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 解析响应
	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		return &apiResp, fmt.Errorf("请求失败，状态码: %d, 消息: %s", resp.StatusCode, apiResp.Message)
	}

	return &apiResp, nil
}

// GetWithToken 带Token认证的GET请求方法
func (c *APIClient) GetWithToken(endpoint string, token string) (*APIResponse, error) {
	// 创建请求
	url := c.baseURL + endpoint
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置请求头
	req.Header.Set("Authorization", "Bearer "+token)

	// 发送请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送GET请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 解析响应
	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		return &apiResp, fmt.Errorf("请求失败，状态码: %d, 消息: %s", resp.StatusCode, apiResp.Message)
	}

	return &apiResp, nil
}

// GetConfig 获取角色配置信息（带缓存）
func (c *APIClient) GetConfig() (*ConfigResponse, error) {
	return c.getConfigWithIsAll(false, true)
}

// GetConfigWithRetry 获取角色配置信息（带重试机制）
// 重试20次，每次失败后等待15秒再重试
// 如果所有重试都失败，程序将自动重启
func (c *APIClient) GetConfigWithRetry() (*ConfigResponse, error) {
	const maxRetries = 20
	const retryInterval = 15 * time.Second

	//var lastErr error

	for i := 0; i < maxRetries; i++ {
		config, err := c.GetConfig()
		if err == nil {
			if i > 0 {
				SLS_Log2(fmt.Sprintf("✅ 配置加载成功 (第 %d 次尝试)", i+1))
			} else {
				SLS_Log2("✅ 配置加载成功")
			}

			// 打印websocket信息
			if config.Data.Websocket != "" {
				fmt.Printf("设备运行信息(Websocket): %s\n", config.Data.Websocket)
			}

			return config, nil
		}

		//lastErr = err
		SLS_Log2(fmt.Sprintf("⚠️ 配置加载失败 (尝试 %d/%d): %v", i+1, maxRetries, err))

		if i < maxRetries-1 {
			SLS_Log2(fmt.Sprintf("等待 %d 秒后重试...", int(retryInterval.Seconds())))
			time.Sleep(retryInterval)
		}
	}

	// 所有重试都失败
	SLS_Log2(fmt.Sprintf("配置加载失败，已重试 %d 次 请关闭代理或设置分应用代理重试", maxRetries))
	return nil, fmt.Errorf("配置加载失败，已重试 %d 次 请关闭代理或设置分应用代理重试", maxRetries)
}

// GetConfigLight 获取角色配置信息（仅基本字段，用于版本检查）
func (c *APIClient) GetConfigLight() (*ConfigResponse, error) {
	return c.getConfigWithIsAll(true, false) // 强制刷新，用于版本检查
}

// RefreshConfig 强制刷新配置缓存
func (c *APIClient) RefreshConfig() (*ConfigResponse, error) {
	return c.getConfigWithIsAll(true, true)
}

// GetNotifications 获取角色通知信息
func (c *APIClient) GetNotifications() (*NotificationResponse, error) {
	// 从全局存储中获取token
	tokenInterface := Get("token")
	if tokenInterface == nil {
		return nil, fmt.Errorf("未找到token，请先登录")
	}

	token, ok := tokenInterface.(string)
	if !ok {
		return nil, fmt.Errorf("token格式错误")
	}

	// 获取roleId
	if roleInstance == nil {
		return nil, fmt.Errorf("Role实例未初始化")
	}

	roleId := roleInstance.ID
	if roleId == 0 {
		return nil, fmt.Errorf("角色ID未设置")
	}

	// 构建endpoint
	endpoint := fmt.Sprintf("/api/notifications/%d", roleId)

	// 创建请求
	url := c.baseURL + endpoint
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置请求头
	req.Header.Set("Authorization", "Bearer "+token)

	// 发送请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 解析响应
	var notificationResp NotificationResponse
	if err := json.NewDecoder(resp.Body).Decode(&notificationResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		return &notificationResp, fmt.Errorf("请求失败，状态码: %d, 消息: %s", resp.StatusCode, notificationResp.Message)
	}

	return &notificationResp, nil
}

// ClearConfigCache 清除配置缓存
func (c *APIClient) ClearConfigCache() {
	configCache.mu.Lock()
	defer configCache.mu.Unlock()

	configCache.Data = nil
	configCache.ID = 0
	configCache.Version = 0
	fmt.Println("配置缓存已清除")
}

// GetCachedConfigInfo 获取缓存的配置ID和版本号
func (c *APIClient) GetCachedConfigInfo() (int, int) {
	configCache.mu.RLock()
	defer configCache.mu.RUnlock()

	return configCache.ID, configCache.Version
}

func printConfigRoleInfo(data ConfigData) {
	autoAnswer := true
	if data.AutoAnswerEnabled != nil {
		autoAnswer = *data.AutoAnswerEnabled
	}
	fmt.Printf("[api][config] role_name=%q vocation_name=%q ai_answer_custom=%q auto_answer_enabled=%v\n",
		data.RoleName, data.VocationName, data.AIAnswerCustom, autoAnswer)
}

// getConfigWithIsAll 内部方法，支持缓存、强制刷新和isAll参数
func (c *APIClient) getConfigWithIsAll(forceRefresh bool, isAll bool) (*ConfigResponse, error) {
	// 如果有缓存且不强制刷新
	if !forceRefresh {
		configCache.mu.RLock()
		if configCache.Data != nil {
			// 如果isAll=false，只返回基本信息（不包含content）
			if !isAll {
				configResp := &ConfigResponse{
					Code:    200,
					Message: "success",
					Data: ConfigData{
						ID:                configCache.Data.ID,
						Name:              configCache.Data.Name,
						Game:              configCache.Data.Game,
						Description:       configCache.Data.Description,
						Version:           configCache.Data.Version,
						RoleID:            configCache.Data.RoleID,
						UserID:            configCache.Data.UserID,
						CreateAt:          configCache.Data.CreateAt,
						UpdateAt:          configCache.Data.UpdateAt,
						Websocket:         configCache.Data.Websocket,
						RoleName:          configCache.Data.RoleName,
						VocationName:      configCache.Data.VocationName,
						AIAnswerCustom:    configCache.Data.AIAnswerCustom,
						AutoAnswerEnabled: configCache.Data.AutoAnswerEnabled,
						// Content 字段故意不设置，节省带宽
					},
				}
				configCache.mu.RUnlock()
				return configResp, nil
			} else {
				// isAll=true，返回完整缓存
				configResp := &ConfigResponse{
					Code:    200,
					Message: "success",
					Data:    *configCache.Data,
				}
				configCache.mu.RUnlock()
				return configResp, nil
			}
		}
		configCache.mu.RUnlock()
	}

	// 从全局存储中获取token
	tokenInterface := Get("token")
	if tokenInterface == nil {
		return nil, fmt.Errorf("未找到token，请先登录")
	}

	token, ok := tokenInterface.(string)
	if !ok {
		return nil, fmt.Errorf("token格式错误")
	}

	// 获取roleId
	if roleInstance == nil {
		return nil, fmt.Errorf("Role实例未初始化")
	}

	roleId := roleInstance.ID
	if roleId == 0 {
		return nil, fmt.Errorf("角色ID未设置")
	}

	// 构建endpoint，添加isAll参数
	endpoint := fmt.Sprintf("/api/configs/role/%d?isAll=%t", roleId, isAll)

	// 创建请求
	url := c.baseURL + endpoint
	//fmt.Printf("请求配置API: %s (isAll=%t)\n", url, isAll)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置请求头
	req.Header.Set("Authorization", "Bearer "+token)

	// 发送请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送GET请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 解析响应
	var configResp ConfigResponse
	if err := json.NewDecoder(resp.Body).Decode(&configResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		return &configResp, fmt.Errorf("请求失败，状态码: %d, 消息: %s", resp.StatusCode, configResp.Message)
	}

	// 检查业务状态码
	if configResp.Code != 200 {
		return &configResp, fmt.Errorf("获取配置失败: %s", configResp.Message)
	}

	// 检查配置ID和版本是否有变化（需要读锁）
	configCache.mu.RLock()
	configChanged := configCache.Data == nil ||
		configCache.ID != configResp.Data.ID ||
		configCache.Version != configResp.Data.Version

	oldID := configCache.ID
	oldVersion := configCache.Version
	configCache.mu.RUnlock()

	if configChanged {
		if configCache.Data == nil {
			fmt.Printf("首次获取配置: ID=%d, 版本=%d\n", configResp.Data.ID, configResp.Data.Version)
		} else if oldID != configResp.Data.ID {
			fmt.Printf("配置ID变化: %d -> %d，版本: %d -> %d，更新缓存\n",
				oldID, configResp.Data.ID, oldVersion, configResp.Data.Version)
		} else if oldVersion != configResp.Data.Version {
			fmt.Printf("配置版本更新: ID=%d, %d -> %d，更新缓存\n",
				configResp.Data.ID, oldVersion, configResp.Data.Version)
		}
	} else {
		//fmt.Printf("配置未变化: ID=%d, 版本=%d\n", configResp.Data.ID, configResp.Data.Version)
	}

	// 更新缓存（需要写锁）
	configCache.mu.Lock()
	if isAll {
		// 完整更新缓存
		configCache.Data = &configResp.Data
	} else {
		// 只更新版本信息，保持原有的完整配置内容
		if configCache.Data != nil {
			// 只更新基本字段，保留Content
			configCache.Data.ID = configResp.Data.ID
			configCache.Data.Name = configResp.Data.Name
			configCache.Data.Game = configResp.Data.Game
			configCache.Data.Description = configResp.Data.Description
			configCache.Data.Version = configResp.Data.Version
			configCache.Data.RoleID = configResp.Data.RoleID
			configCache.Data.UserID = configResp.Data.UserID
			configCache.Data.CreateAt = configResp.Data.CreateAt
			configCache.Data.UpdateAt = configResp.Data.UpdateAt
			configCache.Data.Websocket = configResp.Data.Websocket
			configCache.Data.RoleName = configResp.Data.RoleName
			configCache.Data.VocationName = configResp.Data.VocationName
			configCache.Data.AIAnswerCustom = configResp.Data.AIAnswerCustom
			configCache.Data.AutoAnswerEnabled = configResp.Data.AutoAnswerEnabled
			// Content 字段保持不变
		} else {
			// 如果没有缓存，则完整设置（但Content为空）
			configCache.Data = &configResp.Data
		}
	}
	configCache.ID = configResp.Data.ID
	configCache.Version = configResp.Data.Version
	configCache.mu.Unlock()

	//fmt.Printf("成功获取角色配置，配置ID: %d, 配置名称: %s, 游戏: %s, 版本: %d\n",
	//	configResp.Data.ID, configResp.Data.Name, configResp.Data.Game, configResp.Data.Version)

	// 只在配置变化时打印JSON格式的配置内容和websocket信息
	if configChanged {
		printConfigRoleInfo(configResp.Data)

		// 打印websocket信息
		if configResp.Data.Websocket != "" {
			fmt.Printf("设备运行信息(Websocket): %s\n", configResp.Data.Websocket)
		}

		// 打印配置内容
		if configResp.Data.Content != nil {
			jsonData, err := json.MarshalIndent(configResp.Data.Content, "", "  ")
			if err == nil {
				fmt.Println("配置内容(JSON格式):")
				fmt.Println(string(jsonData))
			}
		}
	}

	return &configResp, nil
}

// GetConfigValue 获取配置中的特定值 - 仅从缓存读取
// 使用示例: GetConfigValue("副本.女武神.启用") 或 GetConfigValue("挂机设置.血量阈值")
// 注意: 此方法不会触发API调用，需要先调用GetConfig()初始化缓存
func (c *APIClient) GetConfigValue(path string) (interface{}, error) {
	// 只从缓存读取，不触发API调用，使用读锁保证安全
	configCache.mu.RLock()
	defer configCache.mu.RUnlock()

	if configCache.Data == nil {
		return nil, fmt.Errorf("配置缓存为空，请先调用GetConfig()初始化配置")
	}

	if configCache.Data.Content == nil {
		return nil, fmt.Errorf("配置内容为空")
	}

	// 分割路径
	keys := strings.Split(path, ".")
	current := configCache.Data.Content

	// 逐级访问
	for i, key := range keys {
		if i == len(keys)-1 {
			// 最后一个键，返回值
			return current[key], nil
		} else {
			// 中间键，继续深入
			if next, ok := current[key].(map[string]interface{}); ok {
				current = next
			} else {
				return nil, fmt.Errorf("路径 '%s' 在键 '%s' 处不存在或类型错误", path, key)
			}
		}
	}

	return nil, fmt.Errorf("路径为空")
}

// GetConfigBool 获取配置中的布尔值
func (c *APIClient) GetConfigBool(path string) (bool, error) {
	value, err := c.GetConfigValue(path)
	if err != nil {
		return false, err
	}

	if boolValue, ok := value.(bool); ok {
		return boolValue, nil
	}
	if stringValue, ok := value.(string); ok {
		switch strings.ToLower(strings.TrimSpace(stringValue)) {
		case "true", "1", "yes", "on":
			return true, nil
		case "false", "0", "no", "off":
			return false, nil
		}
	}
	if f, ok := value.(float64); ok {
		return f != 0, nil
	}

	return false, fmt.Errorf("路径 '%s' 的值不是布尔类型", path)
}

// GetConfigString 获取配置中的字符串值
func (c *APIClient) GetConfigString(path string) (string, error) {
	value, err := c.GetConfigValue(path)
	if err != nil {
		return "", err
	}

	if stringValue, ok := value.(string); ok {
		return stringValue, nil
	}

	return "", fmt.Errorf("路径 '%s' 的值不是字符串类型", path)
}

// GetConfigInt 获取配置中的整数值
func (c *APIClient) GetConfigInt(path string) (int, error) {
	value, err := c.GetConfigValue(path)
	if err != nil {
		return 1, err
	}

	// JSON中的数字可能是float64类型
	switch v := value.(type) {
	case int:
		return v, nil
	case float64:
		return int(v), nil
	case string:
		// 尝试从字符串转换
		if intValue, err := strconv.Atoi(v); err == nil {
			return intValue, nil
		}
	}

	return 0, fmt.Errorf("路径 '%s' 的值不是整数类型", path)
}

// GetConfigIntValue 获取配置中的整数值，失败时返回默认值0
func (c *APIClient) GetConfigIntValue(path string) int {
	value, _ := c.GetConfigInt(path)
	return value
}

// GetConfigIntValueOrDefault 获取配置中的整数值，如果值为0或获取失败，返回指定的默认值
func (c *APIClient) GetConfigIntValueOrDefault(path string, defaultValue int) int {
	value, _ := c.GetConfigInt(path)
	if value == 0 {
		return defaultValue
	}
	return value
}

// GetConfigTime 获取配置中的时间间隔（分钟），返回带有±15%随机幅度的值
// path: 配置路径，如 "间隔时间.仓库存储间隔"
// defaultValue: 默认值（当配置值为0或不存在时使用）
func (c *APIClient) GetConfigTime(path string, defaultValue int) int {
	value := c.GetConfigIntValue(path)
	if value <= 1 {
		value = defaultValue
	}

	// 计算±15%的范围
	min := int(float64(value) * 0.85) // 85%
	max := int(float64(value) * 1.15) // 115%

	// 返回随机值
	if max <= min {
		return value
	}
	return rand.Intn(max-min+1) + min
}

// farmMapTrailingSpaceDigits 匹配「挂机地图」里常见的末尾展示：空格 + 阿拉伯数字频道（如 "萌孢山丘 10"）。
var farmMapTrailingSpaceDigits = regexp.MustCompile(`\s+\d+$`)

// NormalizeFarmMapLabel 去掉地图名末尾的「空格+数字频道」；可与按 `-` 取线号前的名称一起用。
func NormalizeFarmMapLabel(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	return strings.TrimSpace(farmMapTrailingSpaceDigits.ReplaceAllString(s, ""))
}

// GetConfigStringValue 获取配置中的字符串值，失败时返回空字符串
func (c *APIClient) GetConfigStringValue(path string) string {
	value, _ := c.GetConfigString(path)
	if value == "" {
		return ""
	}
	// 按-切分，取第0位（线号等多在 `-` 之后）
	parts := strings.Split(value, "-")
	out := strings.TrimSpace(parts[0])
	if path == "挂机地图" {
		out = NormalizeFarmMapLabel(out)
	}
	return out
}

// GetConfigBoolValue 获取配置中的布尔值，失败时返回false
func (c *APIClient) GetConfigBoolValue(path string) bool {
	value, _ := c.GetConfigBool(path)
	return value
}

// GetConfigBoolValueOrDefault 获取配置中的布尔值，如果获取失败，返回指定的默认值
func (c *APIClient) GetConfigBoolValueOrDefault(path string, defaultValue bool) bool {
	value, err := c.GetConfigBool(path)
	if err != nil {
		return defaultValue
	}
	return value
}

// GetConfigArray 获取配置中的数组值
func (c *APIClient) GetConfigArray(path string) ([]interface{}, error) {
	value, err := c.GetConfigValue(path)
	if err != nil {
		return nil, err
	}

	if arrayValue, ok := value.([]interface{}); ok {
		return arrayValue, nil
	}

	return nil, fmt.Errorf("路径 '%s' 的值不是数组类型", path)
}

// GetConfigStringArray 获取配置中的字符串数组值
func (c *APIClient) GetConfigStringArray(path string) ([]string, error) {
	arrayValue, err := c.GetConfigArray(path)
	if err != nil {
		return nil, err
	}

	var stringArray []string
	for i, item := range arrayValue {
		if str, ok := item.(string); ok {
			stringArray = append(stringArray, str)
		} else {
			return nil, fmt.Errorf("路径 '%s' 的数组第%d个元素不是字符串类型", path, i)
		}
	}

	return stringArray, nil
}

// GetConfigStringArrayValue 获取配置中的字符串数组值，失败时返回空数组
func (c *APIClient) GetConfigStringArrayValue(path string) []string {
	value, err := c.GetConfigStringArray(path)
	if err != nil {
		return []string{} // 返回空数组而不是nil
	}
	return value
}

// GetConfigFloat 获取配置中的浮点数值
func (c *APIClient) GetConfigFloat(path string) (float64, error) {
	value, err := c.GetConfigValue(path)
	if err != nil {
		return 0, err
	}

	switch v := value.(type) {
	case float64:
		return v, nil
	case int:
		return float64(v), nil
	case string:
		if floatValue, err := strconv.ParseFloat(v, 64); err == nil {
			return floatValue, nil
		}
	}

	return 0, fmt.Errorf("路径 '%s' 的值不是数字类型", path)
}

// SetBaseURL 设置基础URL
func (c *APIClient) SetBaseURL(baseURL string) {
	c.baseURL = baseURL
}

// GetBaseURL 获取基础URL
func (c *APIClient) GetBaseURL() string {
	return c.baseURL
}

// GetWebsocketInfo 获取缓存的websocket设备运行信息
func (c *APIClient) GetWebsocketInfo() string {
	configCache.mu.RLock()
	defer configCache.mu.RUnlock()

	if configCache.Data != nil {
		return configCache.Data.Websocket
	}
	return ""
}
