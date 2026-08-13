package job

import (
	"app/core"
	"fmt"
	"strings"
)

const (
	closeTopBarX1, closeTopBarY1 = 166, 24
	closeTopBarX2, closeTopBarY2 = 998, 623
	closeTopBarTemplates         = "img/game/close1.png,img/game/close2.png"
	closeTopBarSim               = float32(0.8)
	closeTopBarMaxRounds         = 2
	closeTopBarBetweenMs         = 500
)

// CloseTopBarCloseButtons 在屏幕上部区域匹配 close1/close2，找到则点击，最多 closeTopBarMaxRounds 轮，直到不再出现。
func CloseTopBarCloseButtons(logTag string) {
	for i := 0; i < closeTopBarMaxRounds; i++ {

		// 对话弹框
		num := core.Color.GetColorCountInRegion(258, 462, 396, 485, "b4e700", 0.99)
		if num >= 30 {
			text := core.OCR.DetectText(286, 462, 371, 492)
			if strings.Contains(text, "停止") {
				core.SLS_Log2("遇到停止对话框 点击停止")
				core.RandomClickInArea(301, 469, 360, 480)
			}
		}

		cx, cy := core.OpenCV.FindImage(closeTopBarX1, closeTopBarY1, closeTopBarX2, closeTopBarY2, closeTopBarTemplates, false, 1.0, closeTopBarSim)
		if cx <= 0 || cy <= 0 {
			break
		}
		fmt.Printf("%s 关闭道具框 close1 x=%d y=%d i=%d\n", logTag, cx, cy, i)
		core.RandomClickInArea(cx, cy, cx+2, cy+2)
		core.Sleep(closeTopBarBetweenMs)

	}
}
