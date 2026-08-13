package util

import (
	"app/assets"
	"app/core"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// ConfigChecker 配置检查器
type ConfigChecker struct {
	configCheckURL string
}

// ConfigRequest 配置检查请求结构
type ConfigRequest struct {
	ConfigID int `json:"config_id"`
	Version  int `json:"version"`
}

// NewConfigChecker 创建配置检查器
func NewConfigChecker() *ConfigChecker {
	config := loadConfigCheckerConfig()
	return &ConfigChecker{
		configCheckURL: config.ConfigCheckURL,
	}
}

// loadConfigCheckerConfig 加载配置检查器配置
func loadConfigCheckerConfig() *ConfigCheckerConfig {
	// 使用嵌入文件系统读取配置文件
	data, err := assets.ConfigFile.ReadFile("config/default.json")
	if err != nil {
		log.Fatalf("读取配置文件失败: %v", err)
	}

	// 解析配置文件
	var config map[string]interface{}
	err = json.Unmarshal(data, &config)
	if err != nil {
		log.Fatalf("解析配置文件失败: %v", err)
	}

	// 提取配置检查器配置
	var checkerConfig *ConfigCheckerConfig

	if server, ok := config["server"].(map[string]interface{}); ok {
		host := server["host"].(string)
		port := server["port"].(string)
		configPath := server["config_check_url"].(string)

		checkerConfig = &ConfigCheckerConfig{
			ConfigCheckURL: fmt.Sprintf("http://%s:%s%s", host, port, configPath),
		}
	} else {
		log.Fatalf("配置文件中缺少server配置")
	}

	return checkerConfig
}

// ConfigCheckerConfig 配置检查器配置
type ConfigCheckerConfig struct {
	ConfigCheckURL string `json:"config_check_url"`
}

// CheckUpdate 检查配置更新
func (c *ConfigChecker) CheckUpdate(configID, version int) (map[string]interface{}, error) {
	// 请求数据
	requestData := ConfigRequest{
		ConfigID: configID,
		Version:  version,
	}

	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return nil, fmt.Errorf("序列化请求数据失败: %v", err)
	}

	// 发送POST请求
	resp, err := http.Post(c.configCheckURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("服务器返回错误状态码: %d", resp.StatusCode)
	}

	// 读取响应
	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	return result, nil
}

// CheckUpdateAndPrint 检查配置更新并格式化打印结果
func (c *ConfigChecker) CheckUpdateAndPrint(configID, version int) error {
	result, err := c.CheckUpdate(configID, version)
	if err != nil {
		core.SLS_Log2(fmt.Sprintf("配置更新检查失败: %v", err))
		return err
	}

	// 格式化输出JSON
	prettyJSON, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Printf("配置更新检查结果: %v\n", result)
	} else {
		fmt.Printf("配置更新检查结果:\n%s\n", string(prettyJSON))
	}

	return nil
}

// CheckConfigUpdate 全局函数：检查配置更新
func CheckConfigUpdate(configID, version int) error {
	checker := NewConfigChecker()
	return checker.CheckUpdateAndPrint(configID, version)
}
