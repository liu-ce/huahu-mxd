package captcha

import (
	"app/core"
	"app/job"
	"app/play"
	"app/util"
	"fmt"
	"time"

	"github.com/Dasongzi1366/AutoGo/utils"
)

// 验证码：两种弹窗样式 + GM 巡逻，命中后暂停打怪；消失后恢复。

type colorCheck struct {
	x1, y1    int
	x2, y2    int
	color     string
	sim       float32
	minPixels int
}

type popupRule struct {
	name          string
	x1, y1        int
	x2, y2        int
	color         string
	sim           float32
	minPixels     int
	extraColors   []colorCheck
	colorOnly     bool
	patrol        bool
	ocrRecType    string
	ocrX1         int
	ocrY1         int
	ocrX2         int
	ocrY2         int
	ocrOriginalX1 int
	ocrOriginalY1 int
	ocrOriginalX2 int
	ocrOriginalY2 int
}

var popupRules = []popupRule{
	{
		name: "样式1", x1: 662, y1: 497, x2: 759, y2: 599,
		color: "323232", sim: 0.94, minPixels: 5000,
		ocrX1: 835, ocrY1: 182, ocrX2: 1058, ocrY2: 227,
	},
	{
		name: "样式2", x1: 918, y1: 498, x2: 999, y2: 589,
		color: "333133", sim: 0.94, minPixels: 4000,
		ocrX1: 1007, ocrY1: 364, ocrX2: 1158, ocrY2: 391,
	},
	{
		name: "GM巡逻", x1: 498, y1: 237, x2: 960, y2: 439,
		color: "f4f9f4", sim: 0.95, minPixels: 50000, patrol: true,
		extraColors: []colorCheck{
			{x1: 317, y1: 381, x2: 473, y2: 402, color: "213356", sim: 0.95, minPixels: 300},
			{x1: 458, y1: 58, x2: 817, y2: 146, color: "fff478", sim: 0.95, minPixels: 20},
		},
		ocrRecType:    "korean",
		ocrOriginalX1: 493, ocrOriginalY1: 177, ocrOriginalX2: 960, ocrOriginalY2: 473,
	},
}

const gmPatrolPauseDur = 2 * time.Minute

func detectAreaOCR(x1, y1, x2, y2 int, recType string, multiline bool) string {
	if multiline {
		if recType != "" {
			return core.OCR.DetectMultilineText(x1, y1, x2, y2, recType)
		}
		return core.OCR.DetectMultilineText(x1, y1, x2, y2)
	}
	if recType != "" {
		return core.OCR.DetectText(x1, y1, x2, y2, recType)
	}
	return core.OCR.DetectText(x1, y1, x2, y2)
}

func detectRuleOCR(rule popupRule) string {
	return detectAreaOCR(rule.ocrX1, rule.ocrY1, rule.ocrX2, rule.ocrY2, rule.ocrRecType, false)
}

func detectPopup(holding bool, rule popupRule) (popup bool, pixelCount int, ocrText string) {
	pixelCount = core.Color.GetColorCountInRegion(rule.x1, rule.y1, rule.x2, rule.y2, rule.color, rule.sim)
	extraCounts := make([]int, len(rule.extraColors))
	for i, extra := range rule.extraColors {
		extraCounts[i] = core.Color.GetColorCountInRegion(extra.x1, extra.y1, extra.x2, extra.y2, extra.color, extra.sim)
	}
	if rule.patrol {
		anyNonZero := pixelCount > 0
		for _, n := range extraCounts {
			if n > 0 {
				anyNonZero = true
				break
			}
		}
		if anyNonZero {
			msg := fmt.Sprintf("[GM巡逻] 主区域=%d/%d", pixelCount, rule.minPixels)
			for i, extra := range rule.extraColors {
				msg += fmt.Sprintf(" 附加%d=%d/%d", i+1, extraCounts[i], extra.minPixels)
			}
		}
	}
	if pixelCount <= rule.minPixels {
		return false, pixelCount, ""
	}
	for i, extra := range rule.extraColors {
		if extraCounts[i] <= extra.minPixels {
			return false, pixelCount, ""
		}
	}
	if holding || rule.colorOnly {
		return true, pixelCount, ""
	}
	if rule.patrol {
		return true, pixelCount, ""
	}
	ocrText = detectRuleOCR(rule)
	fmt.Printf("[验证码][%s] %s\n", rule.name, ocrText)
	return util.CalculateTextSimilarity(ocrText, "准备寻找透明图形") > 0.5, pixelCount, ocrText
}

// Run 后台检测验证码/GM 巡逻弹窗，命中后暂停打怪并在关闭后恢复。
func Run() {
	holding := false
	gmPatrolActive := false
	gmPatrolHoldUntil := time.Time{}
	for {
		popup := false
		triggerName := ""
		triggerPixels := 0
		triggerPatrol := false
		var triggerRule popupRule
		skipGMPatrol := job.IsAutoShopRunning() || job.IsShopPanelOpen() || core.IsScheduledLogoutActive()
		for _, rule := range popupRules {
			if rule.patrol && skipGMPatrol {
				continue
			}
			ok, n, _ := detectPopup(holding, rule)
			if ok {
				popup = true
				triggerName = rule.name
				triggerPixels = n
				triggerPatrol = rule.patrol
				triggerRule = rule
				break
			}
		}
		if popup {
			if !holding {
				logFn := core.SLS_Log2
				if triggerPatrol {
					logFn = core.SLS_Log2NoToast
				}
				logFn(fmt.Sprintf("验证码弹出 %s pixel=%d 暂停打怪", triggerName, triggerPixels))
				core.SetCaptchaHold(true)
				holding = true
				if triggerPatrol {
					gmPatrolActive = true
					gmPatrolHoldUntil = time.Now().Add(gmPatrolPauseDur)
					core.GMPatrolBeforeInput = play.ReleaseAllHeldKeys
					play.ReleaseAllHeldKeys()
					core.NotifyGMPatrolDingTalkAlert()
					core.SLS_Log2NoToast("GM巡逻 已暂停 动态等待文字展示 6~10s")
					job.DO_读取人物等级()
					ocrText := waitGMPatrolOCRStable(triggerRule)
					fmt.Printf("[验证码][%s][原文] %s\n", triggerRule.name, ocrText)
					play.ReleaseAllHeldKeys()
					core.StartGMPatrolAIAnswer(ocrText)
					util.UploadGMPatrolScreenshotAsync()
				} else {
					core.NotifyCaptchaDingTalk()
				}
			}
		} else if holding {
			if gmPatrolActive {
				if wait := time.Until(gmPatrolHoldUntil); wait > 0 {
					core.SLS_Log2NoToast(fmt.Sprintf("GM巡逻 脚本暂停 %ds", int(wait.Seconds()+0.5)))
					time.Sleep(wait)
				}
				gmPatrolActive = false
			} else {
				core.RandomSleep(34000, 36000)
			}
			core.SLS_Log2("验证码关闭 恢复打怪")
			core.SetCaptchaHold(false)
			holding = false
		}
		utils.Sleep(500)
	}
}

// StartPeriodicCloseTopBar 每 30 秒尝试关顶栏（验证码暂停期间跳过）。
func StartPeriodicCloseTopBar() {
	go func() {
		t := time.NewTicker(30 * time.Second)
		defer t.Stop()
		for range t.C {
			if core.IsCaptchaHold() {
				continue
			}
			job.CloseTopBarCloseButtons("[job][定时关顶栏]")
		}
	}()
}
