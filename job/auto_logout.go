package job

import "app/core"

// DoAutoLogout 从游戏内退出到大厅「游戏」按钮界面。
func DoAutoLogout() bool {
	for i := 0; i < 30; i++ {
		core.SLS_Log2NoToast("自动离线进行中")

		colorNumber := core.Color.GetColorCountInRegion(1217, 22, 1259, 69, "ffffff", 0.9)
		colorNumber2 := core.Color.GetColorCountInRegion(23, 312, 68, 358, "ffbf33", 0.9)
		if colorNumber >= 150 && colorNumber <= 250 && colorNumber2 >= 3 {
			core.RandomClickInArea(1228, 42, 1243, 57)
			core.RandomSleep(2000, 3000)
			continue
		}

		colorNumber = core.Color.GetColorCountInRegion(543, 626, 777, 682, "323232", 0.98)
		if colorNumber >= 12000 {
			text := core.OCR.DetectText(1106, 642, 1198, 676)
			if text == "前往大厅" {
				core.RandomClickInArea(1106, 642, 1198, 676)
				core.RandomSleep(3000, 4000)
				continue
			}
		}

		colorNumber = core.Color.GetColorCountInRegion(730, 462, 881, 520, "ffb62a", 0.98)
		if colorNumber >= 7000 {
			text := core.OCR.DetectText(765, 471, 823, 510)
			if text == "是" {
				core.RandomClickInArea(765, 471, 823, 510)
				core.RandomSleep(3000, 4000)
				continue
			}
		}

		colorNumber = core.Color.GetColorCountInRegion(380, 1118, 575, 1175, "f9b225", 0.98)
		if colorNumber >= 8000 {
			text := core.OCR.DetectText(420, 1125, 529, 1171)
			if text == "游戏" {
				return true
			}
		}

		core.RandomSleep(2000, 3000)
	}
	return false
}

// DoAutoLogin 从大厅「游戏」按钮登录并进入角色。
func DoAutoLogin() bool {
	for i := 0; i < 90; i++ {
		colorNumber := core.Color.GetColorCountInRegion(380, 1118, 575, 1175, "f9b225", 0.98)
		if colorNumber >= 8000 {
			text := core.OCR.DetectText(420, 1125, 529, 1171)
			if text == "游戏" {
				core.RandomClickInArea(420, 1125, 529, 1171)
				core.RandomSleep(5000, 6000)
				continue
			}
		}

		colorNumber = core.Color.GetColorCountInRegion(723, 517, 803, 613, "c0a17b", 0.95)
		if colorNumber >= 1500 {
			core.SLS_Log2NoToast("点击登录")
			core.RandomClickInArea(798, 549, 826, 574)
			core.RandomSleep(3000, 4000)
			continue
		}

		colorNumber = core.Color.GetColorCountInRegion(1035, 435, 1106, 467, "cc934a", 0.95)
		if colorNumber >= 1000 {
			core.SLS_Log2NoToast("选择角色")
			core.RandomClickInArea(332, 472, 352, 515)
			core.RandomSleep(100, 120)
			core.RandomClickInArea(332, 472, 352, 515)
		}

		colorNumber = core.Color.GetColorCountInRegion(1217, 22, 1259, 69, "ffffff", 0.9)
		colorNumber2 := core.Color.GetColorCountInRegion(23, 312, 68, 358, "ffbf33", 0.9)
		if colorNumber >= 150 && colorNumber <= 250 && colorNumber2 >= 3 {
			return true
		}

		core.RandomSleep(3000, 5000)
	}
	return false
}
