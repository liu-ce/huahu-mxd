package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Dasongzi1366/AutoGo/device"
	"github.com/Dasongzi1366/AutoGo/storages"
	"github.com/Dasongzi1366/AutoGo/utils"
)

// Login 用户登录
func (c *APIClient) Login(username, password, deviceID, windowID, game string) (*LoginResponse, error) {
	// 验证必填参数
	if username == "" {
		return nil, fmt.Errorf("用户名不能为空")
	}
	if password == "" {
		return nil, fmt.Errorf("密码不能为空")
	}
	if deviceID == "" {
		return nil, fmt.Errorf("设备ID不能为空")
	}
	if windowID == "" {
		return nil, fmt.Errorf("窗口ID不能为空")
	}
	if game == "" {
		return nil, fmt.Errorf("游戏名不能为空")
	}

	// 构建请求体
	request := LoginRequest{
		Username: username,
		Password: password,
		DeviceID: deviceID,
		WindowID: windowID,
		Game:     game,
	}

	// 序列化请求体
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("序列化请求数据失败: %v", err)
	}

	// 发送POST请求，支持重试
	url := c.baseURL + "/api/auth/mobile-login"
	fmt.Printf("登录请求URL: %s\n", url)
	fmt.Printf("登录请求体: %s\n", string(jsonData))

	// 先进行DNS解析测试
	fmt.Printf("开始DNS解析测试...\n")
	fmt.Println(url)
	err = testDNSResolution(url)
	if err != nil {
		SLS_Log2(fmt.Sprintf("DNS解析失败: %v", err))
		return nil, fmt.Errorf("DNS解析失败: %v", err)
	}
	fmt.Printf("DNS解析成功\n")

	// 进行TCP连接测试
	fmt.Printf("开始TCP连接测试...\n")
	err = testTCPConnection(url)
	if err != nil {
		SLS_Log2(fmt.Sprintf("TCP连接失败: %v", err))
		return nil, fmt.Errorf("TCP连接失败: %v", err)
	}
	fmt.Printf("TCP连接成功\n")

	var resp *http.Response
	maxRetries := 5

	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			waitTime := time.Duration(i*2) * time.Second // 指数退避: 2s, 4s
			fmt.Printf("第%d次重试，等待%v后重试...\n", i, waitTime)
			time.Sleep(waitTime)
		}

		fmt.Printf("开始第%d次HTTP请求...\n", i+1)

		startTime := time.Now()
		resp, err = c.httpClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
		duration := time.Since(startTime)

		fmt.Printf("第%d次请求耗时: %v\n", i+1, duration)

		if err == nil {
			fmt.Printf("请求成功！响应状态码: %d\n", resp.StatusCode)
			break // 请求成功
		}

		SLS_Log2(fmt.Sprintf("第%d次请求失败: %v", i+1, err))
		fmt.Printf("错误类型: %T\n", err)

		// 如果是最后一次重试，输出更详细的错误信息
		if i == maxRetries-1 {
			return nil, fmt.Errorf("发送登录请求失败，已重试%d次", maxRetries)
		}
	}
	defer resp.Body.Close()

	// 解析响应
	var loginResp LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		return nil, fmt.Errorf("解析登录响应失败: %v", err)
	}

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		return &loginResp, fmt.Errorf("登录请求失败，状态码: %d, 消息: %s", resp.StatusCode, loginResp.Message)
	}

	// 检查业务状态码
	if loginResp.Code != 200 {
		return &loginResp, fmt.Errorf("%s", loginResp.Message)
	}

	return &loginResp, nil
}

// syncRolePayloadFromLocal 组 RoleUpdate 请求前：采集内存/CPU、default.json 的 version、本地 data/silver（写入 datas）。
func syncRolePayloadFromLocal() {
	if roleInstance == nil {
		return
	}
	roleInstance.IdentifyMemory()
	roleInstance.IdentifyCpuUsage()
	if Config != nil && Config.Get("version") != nil {
		roleInstance.Version = strconv.Itoa(Config.GetInt("version"))
	}
	s := Storages.DataGet("silver")
	if s == "" {
		return
	}
	v, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil {
		return
	}
	if roleInstance.Datas == nil {
		roleInstance.Datas = make(map[string]interface{})
	}
	roleInstance.Datas["silver"] = v
}

// RoleUpdate 角色更新 - 使用全局Role实例
// lightMode: true表示轻量级更新（只发送ID, Game, DeviceID, WindowID, CurTask），false表示完整更新
func (c *APIClient) RoleUpdate(lightMode ...bool) (*APIResponse, error) {
	// 从全局存储中获取token
	tokenInterface := Get("token")
	if tokenInterface == nil {
		return nil, fmt.Errorf("未找到token，请先登录")
	}

	token, ok := tokenInterface.(string)
	if !ok {
		return nil, fmt.Errorf("token格式错误")
	}

	// 使用全局Role实例
	if roleInstance == nil {
		return nil, fmt.Errorf("Role实例未初始化")
	}

	// 检查是否使用轻量级模式（默认false，保持向后兼容）
	isLightMode := false
	if len(lightMode) > 0 {
		isLightMode = lightMode[0]
	}

	// 完整更新前：default.json 的 version + 本地 data/silver 写入请求体
	if !isLightMode {
		syncRolePayloadFromLocal()
	}

	var request *RoleUpdateRequest
	if isLightMode {
		request = roleInstance.toLightRequest()
	} else {
		request = roleInstance.toRequest()
	}

	return c.PostWithToken("/api/roles/update", request, token)
}

// RefreshToken 刷新token - 使用已存储的用户名和密码重新登录
func (c *APIClient) RefreshToken() error {
	// 从存储中获取用户名、密码和窗口ID
	username := Get("username")
	seq := Get("seq")

	if username == nil || seq == nil {
		return fmt.Errorf("未找到用户名或窗口ID，无法刷新token")
	}

	usernameStr, ok1 := username.(string)
	seqStr, ok2 := seq.(string)

	if !ok1 || !ok2 {
		return fmt.Errorf("用户名或窗口ID格式错误")
	}

	// 从storages获取密码
	password := storages.Get("data", "password")
	if password == "" {
		return fmt.Errorf("未找到密码，无法刷新token")
	}

	SLS_Log2("开始刷新token...")

	deviceId := device.Serial

	// 重新登录获取新token
	resp, err := c.Login(usernameStr, password, deviceId, seqStr, "冒险岛韩服")
	if err != nil {
		return fmt.Errorf("刷新token失败: %v", err)
	}

	// 更新token
	Set("token", resp.Data.Token)
	SLS_Log2(fmt.Sprintf("Token刷新成功: %s", resp.Data.Token))

	return nil
}

// LoginAndSetup 登录并设置全局变量 - 从main.go移过来的函数
func (c *APIClient) LoginAndSetup(username, password, seq string) error {
	deviceId := device.Serial

	if username == "" || password == "" || seq == "" {
		msg := "登录参数不完整：需要 storages 中的 username、password、windowId（请先走 UI /login 保存表单，或自行写入 data）"
		fmt.Printf("[login] %s (username空=%t password空=%t windowId空=%t)\n",
			msg, username == "", password == "", seq == "")
		utils.Toast(msg, 300, 0, 1000)
		return fmt.Errorf("%s", msg)
	}

	// 设置全局参数
	Set("username", username)
	Set("seq", seq)

	// 直接调用登录 + 更新
	resp, err := c.Login(username, password, deviceId, seq, "冒险岛韩服")
	if err != nil {
		fmt.Printf("[login] 请求失败: %v\n", err)
		utils.Toast(fmt.Sprintf("登录失败: %v", err), 0, 0, 1000)
		return fmt.Errorf("%v", err)
	}
	if resp == nil {
		fmt.Println("[login] 响应为空")
		return fmt.Errorf("登录响应为空")
	}
	if resp.Code != 200 {
		fmt.Printf("[login] 业务失败 code=%d message=%s\n", resp.Code, resp.Message)
	}

	// 检查响应中的code字段（即使HTTP状态码是200，业务code可能不是200）
	if resp.Code != 200 {
		errorMsg := fmt.Sprintf("登录失败: %s", resp.Message)
		utils.Toast(errorMsg, 0, 0, 1000)
		return fmt.Errorf("%s", resp.Message)
	}

	// 设置token
	Set("token", resp.Data.Token)
	fmt.Printf("Token: %s\n", resp.Data.Token)
	fmt.Printf("Role.ID: %d\n", resp.Data.RoleId)
	fmt.Printf("adb: %s\n", resp.Data.Adb)
	Set("adb", resp.Data.Adb)
	// 更新游戏角色信息
	if roleInstance == nil {
		roleInstance = &RoleInstance{
			Datas: make(map[string]interface{}),
		}
	}
	roleInstance.ID = resp.Data.RoleId

	// 同时设置公开的Role变量
	Role = roleInstance

	// 登录成功后拉取并打印服务端该角色的配置（/api/configs/role/:id?isAll=true）
	cfgResp, cfgErr := c.RefreshConfig()
	if cfgErr != nil {
		fmt.Printf("[api][config] 拉取账号配置失败: %v\n", cfgErr)
		return nil
	}
	fmt.Printf("[api][config] roleId=%d configId=%d version=%d name=%q game=%q\n",
		cfgResp.Data.RoleID, cfgResp.Data.ID, cfgResp.Data.Version, cfgResp.Data.Name, cfgResp.Data.Game)
	printConfigRoleInfo(cfgResp.Data)
	if cfgResp.Data.Websocket != "" {
		fmt.Printf("[api][config] websocket: %s\n", cfgResp.Data.Websocket)
	}
	if bankItems, bankErr := c.RefreshQuestionBank(); bankErr != nil {
		fmt.Printf("[api][question-bank] 拉取题库失败: %v\n", bankErr)
	} else {
		fmt.Printf("[api][question-bank] 已缓存 %d 题\n", len(bankItems))
	}
	pretty, err := json.MarshalIndent(cfgResp.Data, "", "  ")
	if err != nil {
		fmt.Printf("[api][config] 序列化配置 JSON 失败: %v\n", err)
		return nil
	}
	fmt.Printf("[api][config] 完整配置 JSON:\n%s\n", string(pretty))

	return nil
}
