package main

import (
	"app/core"
	"app/play"
	"app/util"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/Dasongzi1366/AutoGo/app"
	"github.com/Dasongzi1366/AutoGo/device"
	"github.com/Dasongzi1366/AutoGo/system"
)

// ExitOnGameExceptionIfNeeded 挂机维护循环周期性调用；游戏不存在时直接退出脚本。
func ExitOnGameExceptionIfNeeded() {
	if !IsGameMissing() {
		return
	}
	reason := DetectException()
	msg := fmt.Sprintf("[异常] 游戏不存在，退出脚本: %s", reason)
	fmt.Println(msg)
	core.SLS_Log2(msg)
	os.Exit(0)
}

func DetectException() string {
	gamePackage := util.GetGamePackage()
	currentPackage := app.CurrentPackage()
	expected := util.FormatExpectedGamePackages()

	pid := system.GetPid(gamePackage)
	if !util.IsAllowedGamePackage(currentPackage) {
		return fmt.Sprintf("游戏进程已停止运行 期望包名=%s 当前包名=%s pid=%d",
			expected, currentPackage, pid)
	}
	if currentPackage == gamePackage && (pid == 0 || pid == -1) {
		return fmt.Sprintf("游戏进程已停止运行 期望包名=%s 当前包名=%s pid=%d",
			expected, currentPackage, pid)
	}

	//if 是否在启动页() {
	//	return "位于启动页"
	//}
	//
	//if 是否在启动页_启动游戏() {
	//	return "位于启动页启动游戏按钮处"
	//}
	//
	//if 游戏启动加载() {
	//	return "游戏加载界面"
	//}
	//
	//if 请点击荧幕开始游戏() {
	//	return "请点击荧幕开始游戏"
	//}
	//
	//if 选择角色界面() {
	//	return "选择角色界面"
	//}
	//
	//if 没网游戏提示() {
	//	return "没网游戏提示"
	//}
	//
	//if 屏幕上方公告() {
	//	return "屏幕上方公告"
	//}
	//
	//if 小地图缩() {
	//	return "小地图缩放"
	//}
	//if 是否进入游戏() {
	//	return "进入游戏"
	//}
	//
	//if 不支持谷歌Play() {
	//	return "不支持谷歌Play"
	//}
	//
	//if 是否回城() {
	//	return "城内"
	//}
	//
	//if 白屏() {
	//	return "白屏"
	//}
	return ""
}

func IsException() bool {
	status := DetectException()
	if strings.HasPrefix(status, "游戏进程已停止运行") ||
		strings.Contains(status, "检测到游戏进程不存在") ||
		status == "位于启动页" ||
		status == "位于启动页启动游戏按钮处" ||
		status == "游戏加载界面" ||
		status == "请点击荧幕开始游戏" ||
		status == "选择角色界面" ||
		status == "没网游戏提示" ||
		status == "不支持谷歌Play" ||
		status == "白屏" ||
		status == "城内" {
		return true
	}

	return false
}

// IsGameMissing 游戏进程不在前台或已停止。
func IsGameMissing() bool {
	status := DetectException()
	return strings.HasPrefix(status, "游戏进程已停止运行") ||
		strings.Contains(status, "检测到游戏进程不存在")
}

func HandleException() {
	gamePackage := util.GetGamePackage()

	maxCount := 90

	for i := 0; i < maxCount; i++ {
		fmt.Println("循环登录中.." + strconv.Itoa(i))
		status := DetectException()
		if strings.HasPrefix(status, "游戏进程已停止运行") {
			// 游戏没运行 启动游戏
			core.SLS_Log2("游戏没运行 启动游戏")

			if app.Launch(gamePackage, 0) {
				core.SLS_Log2("已启动...准备进入游戏")
			} else {
				core.SLS_Log2("启动游戏失败 请检查是否安装")
			}

			core.RandomSleep(5000, 6000)
		}

		if status == "没网游戏提示" {
			core.SLS_Log2("没网游戏提示")
			core.RandomClickInArea(322, 705, 383, 730)
			core.RandomSleep(2000, 3000)
			app.ForceStop(gamePackage)
			if core.API.GetConfigBoolValue("自动代理") {
				启动代理()
			} else {
				core.SLS_Log2("中控未开启自动代理 停止脚本")
			}
		}

		if status == "不支持谷歌Play" {
			core.RandomClickInArea(568, 698, 598, 711)
			core.RandomSleep(2000, 2300)
		}

		if status == "位于启动页" {
			core.SLS_Log2("位于启动页")
			core.Swipe2(538, 307, 595, 762, 500, 1000)
			core.RandomSleep(2000, 2300)
			core.Swipe2(555, 646, 484, 401, 500, 1000)
			core.RandomSleep(2000, 2300)

			for j := 0; j < 3; j++ {
				x, y := core.OpenCV.FindImage(18, 214, 688, 1110, "img/game/枫星.png", false, 1.0, 0.8)
				if x > 0 && y > 0 {
					core.RandomClickInArea(x, y, x+20, y+20)
				} else {
					break
				}
				core.RandomSleep(4000, 5000)
			}

		}

		if status == "位于启动页启动游戏按钮处" {
			core.SLS_Log2("点击启动游戏")

			text1 := core.OCR.DetectText(429, 1099, 518, 1148)
			text2 := core.OCR.DetectText(425, 1181, 521, 1230)
			if text1 == "游戏" {
				core.RandomClickInArea(420, 1100, 524, 1148)
			}

			if text2 == "游戏" {
				core.RandomClickInArea(425, 1181, 521, 1230)
			}

			core.RandomSleep(10000, 15000)
		}

		if status == "游戏加载界面" {
			core.SLS_Log2("游戏加载界面")
			core.RandomSleep(10000, 15000)
			maxCount = 600
		}

		if status == "请点击荧幕开始游戏" {
			core.SLS_Log2("点击荧幕开始游戏")
			core.RandomClickInArea(581, 555, 664, 580)
			core.RandomSleep(5000, 6000)
		}

		if status == "选择角色界面" {
			core.SLS_Log2("选择角色界面")
			core.RandomClickInArea(1070, 203, 1107, 224)
			core.RandomSleep(5000, 6000)
		}

		if status == "进入游戏" {
			break
		}

		if status == "白屏" {
			core.SLS_Log2("白屏")
			if i >= 10 {
				app.ForceStop(gamePackage)
				i = 0
			}
		}

		core.RandomSleep(5000, 6000)
	}
}

func 是否回城() bool {
	_, _, wx, wy := play.DetectYellowThenWorld()
	fmt.Println(wx, wy)

	num := core.Color.GetColorCountInRegion(7, wy+77, wx+88, wy+193, "00ff00", 0.99)
	fmt.Println(num)
	if num >= 25 {
		return true
	}

	return false
}

func 是否竖屏() bool {
	_, _, _, rotation := device.GetDisplayInfo(0)
	if rotation == 0 {
		return true
	}

	return false

}

func 是否在启动页() bool {
	if !是否竖屏() {
		return false
	}

	if core.Color.GetColorCountInRegion(27, 1132, 102, 1272, "0095ff", 0.99) < 300 {
		return false
	}

	if core.Color.GetColorCountInRegion(154, 1129, 212, 1278, "913cff", 0.99) < 500 {
		return false
	}

	return true
}

func 不支持谷歌Play() bool {
	if !是否竖屏() {
		return false
	}

	if core.Color.GetColorCountInRegion(212, 673, 482, 738, "f2f0f4", 0.99) < 15000 {
		return false
	}

	text := core.OCR.DetectText(548, 686, 620, 7282)
	if text != "确定" {
		return false
	}

	return true
}

func 是否在启动页_启动游戏() bool {
	if !是否竖屏() {
		return false
	}

	if core.Color.GetColorCountInRegion(516, 1101, 604, 1246, "f9b225", 0.99) < 3000 {
		return false
	}

	text1 := core.OCR.DetectText(429, 1099, 518, 1148)
	text2 := core.OCR.DetectText(425, 1181, 521, 1230)

	if text1 == "游戏" || text2 == "游戏" {
		return true
	}

	return false
}

func 游戏启动加载() bool {
	if core.Color.GetColorCountInRegion(14, 11, 139, 74, "000000", 0.99) < 3000 {
		return false
	}

	text := core.OCR.DetectText(606, 388, 670, 422)
	if text != "楓星" {
		return false
	}

	return true
}

func 请点击荧幕开始游戏() bool {
	if core.Color.GetColorCountInRegion(567, 637, 719, 659, "ffffff", 0.99) < 1000 {
		return false
	}

	for i := 0; i < 5; i++ {
		text := core.OCR.DetectText(545, 550, 737, 583)
		similarity := util.CalculateTextSimilarity(text, "請點擊螢幕開始遊戲")
		if similarity > 0.4 {
			return true
		}
	}

	return false
}

func 选择角色界面() bool {
	// 蓝色背景
	if core.Color.GetColorCountInRegion(830, 547, 1016, 639, "627ede", 0.99) < 3000 {
		return false
	}

	text := core.OCR.DetectText(19, 660, 71, 691)
	if text != "返回" {
		return false
	}

	return true
}

func 屏幕上方公告() bool {
	if 是否竖屏() {
		return false
	}

	// 灰色公告
	if core.Color.GetColorCountInRegion(1055, 17, 1186, 61, "666a6d", 0.99) >= 4000 {
		core.RandomClickInArea(1224, 40, 1237, 52)
		core.SLS_Log2("关闭公告")
		return true
	}

	// 蓝色公告
	if core.Color.GetColorCountInRegion(062, 20, 1174, 74, "1caebc", 0.99) >= 4000 {
		core.RandomClickInArea(1224, 40, 1237, 52)
		core.SLS_Log2("关闭公告")
		return true
	}

	return false
}

func 小地图缩() bool {
	for i := 0; i < 3; i++ {
		// 灰色公告
		x, y := core.OpenCV.FindImage(0, 0, 335, 200, "img/game/小地图+.png,img/game/小地图+2.png", false, 1.0, 0.7)
		if x > 0 && y > 0 {
			core.RandomClickInArea(x, y, x+2, y+2)
			core.RandomSleep(1800, 2000)
			core.SLS_Log2("小地图缩放")
		} else {
			break
		}

	}

	return false

}

// 当前正在维护 可以在网页上查看维护内容
func 没网游戏提示() bool {
	if core.Color.GetColorCountInRegion(204, 691, 281, 740, "fbb124", 0.99) <= 3000 {
		return false
	}

	if core.Color.GetColorCountInRegion(194, 533, 276, 585, "faf8f6", 0.99) <= 3000 {
		return false
	}

	text := core.OCR.DetectText(324, 691, 396, 739)
	if text != "确认" {
		return false
	}

	return true
}

func 白屏() bool {
	if !是否竖屏() {
		return false
	}

	if core.Color.GetColorCountInRegion(156, 226, 512, 894, "ffffff", 0.99) <= 200000 {
		return false
	}

	return true

}

func 是否进入游戏() bool {
	if core.Color.GetColorCountInRegion(23, 260, 44, 278, "ffffff", 0.99) <= 100 {
		return false
	}

	if core.Color.GetColorCountInRegion(607, 586, 648, 603, "111111", 0.99) <= 30 {
		return false
	}

	x, y := core.OpenCV.FindImage(401, 560, 892, 622, "img/game/LV.png,img/game/LV2.png", false, 1.0, 0.7)
	if x > 0 && y > 0 {
		return true
	}

	return false

}

func 启动代理() {
	util.HandleKitProxy(true)
}
