package core

import (
	"app/assets"
	"encoding/json"
	"fmt"
	"strings"
)

// ConfigManager 配置管理器
type ConfigManager struct {
	data map[string]interface{} `json:"-"`
}

// Get 获取配置值，支持多层路径如 "app.packages.kit"
func (c *ConfigManager) Get(path string) interface{} {
	if c.data == nil {
		return nil
	}

	keys := strings.Split(path, ".")
	current := c.data

	for i, key := range keys {
		if i == len(keys)-1 {
			// 最后一层，直接返回值
			return current[key]
		}

		// 中间层，需要继续向下查找
		if next, ok := current[key].(map[string]interface{}); ok {
			current = next
		} else {
			return nil
		}
	}

	return nil
}

// GetString 获取字符串类型的配置值
func (c *ConfigManager) GetString(path string) string {
	if value := c.Get(path); value != nil {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return ""
}

// GetInt 获取整数类型的配置值
func (c *ConfigManager) GetInt(path string) int {
	if value := c.Get(path); value != nil {
		if num, ok := value.(float64); ok {
			return int(num)
		}
	}
	return 0
}

// GetBool 获取布尔类型的配置值
func (c *ConfigManager) GetBool(path string) bool {
	if value := c.Get(path); value != nil {
		if b, ok := value.(bool); ok {
			return b
		}
	}
	return false
}

// GetMap 获取map类型的配置值
func (c *ConfigManager) GetMap(path string) map[string]interface{} {
	if value := c.Get(path); value != nil {
		if m, ok := value.(map[string]interface{}); ok {
			return m
		}
	}
	return nil
}

// GetArray 获取数组类型的配置值
func (c *ConfigManager) GetArray(path string) []interface{} {
	if value := c.Get(path); value != nil {
		if arr, ok := value.([]interface{}); ok {
			return arr
		}
	}
	return nil
}

// Exists 检查路径是否存在
func (c *ConfigManager) Exists(path string) bool {
	return c.Get(path) != nil
}

// UserInfo 用户信息结构体
type UserInfo struct {
	UserName string `json:"username"`
	PassWord int    `json:"password"`
	Seq      bool   `json:"seq"`
}

// DataStore 全局数据存储结构体
type DataStore struct {
	UserInfo UserInfo `json:"user_info"`

	// 保留一个通用的map用于临时数据
	TempData map[string]interface{} `json:"temp_data"`
}

// NewDataStore 创建新的数据存储
func NewDataStore() *DataStore {
	return &DataStore{
		UserInfo: UserInfo{},
		TempData: make(map[string]interface{}),
	}
}

// SetTempData 设置临时数据
func (d *DataStore) SetTempData(key string, value interface{}) {
	d.TempData[key] = value
}

// GetTempData 获取临时数据
func (d *DataStore) GetTempData(key string) interface{} {
	return d.TempData[key]
}

// ClearTempData 清空临时数据
func (d *DataStore) ClearTempData() {
	d.TempData = make(map[string]interface{})
}

// 全局数据存储实例
var GlobalStore *DataStore

// 全局配置实例
var Config *ConfigManager

// init 函数在包加载时自动执行
func init() {
	Config = &ConfigManager{}
	GlobalStore = NewDataStore()
	loadConfig()
	// 配置加载完成后，初始化API客户端
	initAPIClient()
}

// initAPIClient 初始化API客户端
func initAPIClient() {
	API = NewAPIClient()
	// 初始化全局角色实例
	initRole()
}

// initRole 初始化全局角色实例
func initRole() {
	roleInstance = &RoleInstance{
		Datas: make(map[string]interface{}),
	}
	// 将实例赋值给全局变量
	Role = roleInstance
}

// loadConfig 加载配置文件
func loadConfig() {
	// 尝试从嵌入文件系统读取默认配置
	data, err := assets.ConfigFile.ReadFile("config/default.json")
	if err != nil {
		return
	}

	var configData map[string]interface{}
	if err := json.Unmarshal(data, &configData); err != nil {
		return
	}

	Config.data = configData
	fmt.Println("嵌入配置文件加载成功: config/default.json")
}

// GetAppPackage 保持向后兼容的函数
func GetAppPackage(appName string) string {
	return Config.GetString("app_packages." + appName)
}

// GetText 获取配置中的文本映射
func GetText(key string) string {
	return Config.GetString("text_mapping." + key)
}
