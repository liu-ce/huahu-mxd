package util

import (
	"app/core"
	"github.com/Dasongzi1366/AutoGo/hud"
)

func RedAlert(context string) {

	core.Set("暂停", true)
	screenWidth := 1280
	screenHeight := 720
	hudWidth := 400
	hudHeight := 300

	// 计算居中位置
	x1 := (screenWidth - hudWidth) / 2
	y1 := (screenHeight - hudHeight) / 2
	x2 := x1 + hudWidth
	y2 := y1 + hudHeight

	hud.New().SetPosition(x1, y1, x2, y2).SetBackgroundColor("#ff0000").SetTextSize(30).SetText([]hud.TextItem{
		{TextColor: "#FFFFFF", Text: context},
	}).Show()

	for {
		core.RandomSleep(10000, 20000)
	}
}
