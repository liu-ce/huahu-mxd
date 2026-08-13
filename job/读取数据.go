package job

import (
	"app/core"
	"fmt"
)

// DO_读取人物等级 先找 LV 图钉锚点，再 OCR 等级写入 core.Role（不上传，RoleUpdate 由挂机主线程推送）。
func DO_读取人物等级() {
	if !core.API.GetConfigBoolValueOrDefault("中控通讯", true) {
		return
	}

	num := core.Color.GetColorCountInRegion(6, 645, 82, 674, "f4d65c", 0.97)
	if num <= 10 {
		return
	}
	//x, y := core.OpenCV.FindImage(0, 630, 90, 685, "img/game/LV.png", false, 1.0, 0.7)
	//fmt.Printf("[job][人物数据] LV锚点 x=%d y=%d\n", x, y)
	//
	//if x <= 0 || y <= 0 {
	//	return
	//}
	lv, lvOK := core.OCR.DetectLevelNumber(37, 648, 84, 670)
	fmt.Println(lv)

	if core.Role != nil {
		if lvOK && lv >= 10 && lv <= 250 {
			core.Role.Level = lv
		}
	}
}

// DO_读取人物数据 兼容旧调用名。
func DO_读取人物数据() {
	DO_读取人物等级()
}
