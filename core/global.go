package core

import (
	"math/rand"
	"sync"
	"time"
)

// 直接暴露的全局实例
var OCR = NewOCRHandler()
var OpenCV = NewOpenCVHandler()
var Color = NewColorHandler()
var Storages = NewStoragesHandler()

// API 将在 api.go 的 init 函数中初始化

// Role 全局角色变量，在API初始化后设置
var Role *RoleInstance

// DetectExceptionFunc 异常检测函数，由task包注册
var DetectExceptionFunc func(status int) string

// IsGameExceptionFunc / HandleGameExceptionFunc 由 main 注册，挂机循环中周期性检测并恢复游戏
var IsGameExceptionFunc func() bool
var HandleGameExceptionFunc func()

// ExitOnGameExceptionIfNeeded 由 main 注册；命中 IsException 时打日志并 os.Exit。
var ExitOnGameExceptionIfNeeded func()

// 全局参数存储
var globalParams = make(map[string]interface{})
var paramsMutex sync.RWMutex

// Set 设置全局参数
func Set(key string, value interface{}) {
	paramsMutex.Lock()
	defer paramsMutex.Unlock()
	globalParams[key] = value
}

// Get 获取全局参数
func Get(key string) interface{} {
	paramsMutex.RLock()
	defer paramsMutex.RUnlock()
	return globalParams[key]
}

func ConfigInit() {
}

// RandomTime 存储随机生成的时刻（分钟和秒）
type RandomTime struct {
	Minute int // 分钟（0-59）
	Second int // 秒（0-59）
}

// GetOrGenerateRandomTime 获取或生成00:00-09:59范围内的随机时刻（保存到内存中）
// 00:00-09:59 指的是任意小时的0-9分钟和0-59秒
// 如果内存中已有值，直接返回；如果没有，生成一个新的随机时刻并保存
// 返回值：随机时刻（分钟和秒）
func GetOrGenerateRandomTime() RandomTime {
	key := "boss_random_time"

	// 先尝试从内存获取
	if value := Get(key); value != nil {
		if rt, ok := value.(RandomTime); ok {
			return rt
		}
	}

	// 内存中没有，生成新的随机时刻（00:00-09:59范围内的任意分钟和秒）
	// 分钟：0-9，秒：0-59
	rand.Seed(time.Now().UnixNano())
	randomMinute := rand.Intn(10) // 0-9分钟
	randomSecond := rand.Intn(60) // 0-59秒

	rt := RandomTime{
		Minute: randomMinute,
		Second: randomSecond,
	}

	// 保存到内存
	Set(key, rt)

	return rt
}
