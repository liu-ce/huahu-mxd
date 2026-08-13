package util

import (
	"app/core"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/Dasongzi1366/AutoGo/app"
	"github.com/Dasongzi1366/AutoGo/system"
	"github.com/Dasongzi1366/AutoGo/utils"
)

// GetGamePackage 获取游戏包名（ymir）
func GetGamePackage() string {
	return core.Config.GetString("app_packages.game")
}

const vaGamePackage = "io.busniess.va"

// AllowedGamePackages 前台视为游戏仍在运行的包名（含虚拟应用容器）。
func AllowedGamePackages() []string {
	p := GetGamePackage()
	if p == "" {
		return []string{vaGamePackage}
	}
	if p == vaGamePackage {
		return []string{p}
	}
	return []string{p, vaGamePackage}
}

// FormatExpectedGamePackages 日志用期望包名列表。
func FormatExpectedGamePackages() string {
	return strings.Join(AllowedGamePackages(), ",")
}

// IsAllowedGamePackage 当前前台是否为游戏或允许的容器包名。
func IsAllowedGamePackage(pkg string) bool {
	for _, p := range AllowedGamePackages() {
		if pkg == p {
			return true
		}
	}
	return false
}

// ExecuteWithProbability 根据给定的概率执行指定的函数
// probability: 执行概率，范围0.0-1.0，例如0.05表示5%的概率
// fn: 要执行的函数
func ExecuteWithProbability(fn func(), probability float64) {
	if probability <= 0 {
		return // 概率为0或负数，不执行
	}
	if probability >= 1.0 {
		fn() // 概率为1或更大，必定执行
		return
	}

	// 生成0.0-1.0之间的随机数
	rand.Seed(time.Now().UnixNano())
	randomValue := rand.Float64()

	if randomValue < probability {
		fn()
	}
}

// RetryUntilTrue 重试直到条件为true，如果连续maxRetries次都为false则返回false
// condition: 条件检查函数，返回true表示成功
// maxRetries: 最大重试次数
// sleepMin: 每次重试之间的最小延时（毫秒）
// sleepMax: 每次重试之间的最大延时（毫秒）
// 返回: true表示条件满足，false表示连续maxRetries次都未满足
func RetryUntilTrue(condition func() bool, maxRetries int, sleepMin, sleepMax int) bool {
	for i := 0; i < maxRetries; i++ {
		if condition() {
			return true
		}
		if i < maxRetries-1 {
			// 最后一次不需要sleep
			core.RandomSleep(sleepMin, sleepMax)
		}
	}
	return false
}

// 概率执行 中文别名函数
func 概率执行(fn func(), probability float64) {
	ExecuteWithProbability(fn, probability)
}

func StopProxy() {
	agent_flag := core.API.GetConfigBoolValue("启用代理")
	if !agent_flag {
		return
	}

	agent := core.API.GetConfigStringValue("代理类型")
	if agent == "" {
		fmt.Println("未设置代理类型")
		return
	}

	// 通用的停止逻辑
	packageKey := fmt.Sprintf("app_packages.%s", agent)
	proxyPackage := core.Config.GetString(packageKey)
	app.ForceStop(proxyPackage)
	fmt.Printf("已停止%s代理\n", agent)
}

func HandleKitProxy(force bool) {

	agent_flag := core.API.GetConfigBoolValue("自动代理")
	fmt.Println(agent_flag)
	if !agent_flag {
		return
	}

	agent := core.API.GetConfigStringValue("代理类型")
	fmt.Println(agent)
	if agent == "" {
		fmt.Println("未设置代理类型")
		return
	}

	// 通用的代理处理逻辑
	packageKey := fmt.Sprintf("app_packages.%s", agent)
	var imagePath string
	if agent == "kit" {
		imagePath = fmt.Sprintf("img/sys/%s_start.png,img/sys/%s_start2.png", agent, agent)

	} else {
		imagePath = fmt.Sprintf("img/sys/%s_start.png", agent)
	}

	proxyPackage := core.Config.GetString(packageKey)
	core.SLS_Log("启动代理:" + agent)
	app.Launch(proxyPackage, 0)
	core.Sleep(3000)

	if agent == "kit" {
		// 先点击2次开启关闭
		core.SLS_Log2("点击2次开启和关闭 防止假连")
		for i := 0; i < 2; i++ {
			core.RandomClickInArea(605, 1168, 650, 1213)
			core.RandomSleep(500, 600)
		}

	}

	// 如果是nekobox 先关闭 再开启
	if agent == "nekobox" {
		core.RandomClickInArea(339, 967, 369, 1001)
		core.RandomSleep(3000, 3500)
	}

	// 处理多个图片路径（逗号分隔）
	imagePaths := strings.Split(imagePath, ",")
	var x, y int = -1, -1
	for _, path := range imagePaths {
		path = strings.TrimSpace(path)
		x, y = core.OpenCV.FindImage(0, 0, 0, 0, path, false, 1.0, 0.9)
		fmt.Printf("代理按钮(%s): %d, %d\n", path, x, y)
		if x != -1 && y != -1 {
			break
		}
	}

	if x != -1 && y != -1 {
		core.Click(x, y)
		fmt.Printf("已点击%s启动按钮\n", agent)
	}
	utils.Sleep(3000)
}

func ExitSystem(state string) {
	core.SLS_Log2("程序退出:" + state)
	core.RandomSleep(2000, 3000)
	os.Exit(0)
}

func Restart(state string) {
	core.SLS_Log("程序重启:" + state)
	core.RandomSleep(4000, 5000)
	system.RestartSelf()
}
