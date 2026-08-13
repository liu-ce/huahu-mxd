package play

import (
	"app/core"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/Dasongzi1366/AutoGo/motion"
	"github.com/Dasongzi1366/AutoGo/storages"
)

const (
	instituteC2AlcadnoMapName    = "알카드노 협회마가티아"
	instituteC2CentralLabMapName = "알카드도 연구연구소 중앙게"
	instituteC2FarmMapName       = "알카드도 연구연구소 C-2 구"

	instituteC2YellowX1 = 8
	instituteC2YellowY1 = 95
	instituteC2YellowX2 = 279
	instituteC2YellowY2 = 221

	instituteC2CentralLabYellowX1   = 46
	instituteC2CentralLabYellowY1   = 101
	instituteC2CentralLabYellowX2   = 186
	instituteC2CentralLabYellowY2   = 167
	instituteC2CentralLabTargetX    = 36
	instituteC2CentralLabPingMinX   = 35
	instituteC2CentralLabPingMaxX   = 37
	instituteC2CentralLabUpHoldMs   = 1000
	instituteC2CentralLabWalkStepMs = 200

	instituteC2FarmYellowX1       = 7
	instituteC2FarmYellowY1       = 95
	instituteC2FarmYellowX2       = 222
	instituteC2FarmYellowY2       = 173
	instituteC2FarmTargetX        = -20
	instituteC2FarmAlignXTol      = 4
	instituteC2FarmTeleportDist   = 7
	instituteC2FarmSuccessYMax    = 128
	instituteC2FarmNPressMin      = 11 // 原8-10 +3~5
	instituteC2FarmNPressMax      = 15
	instituteC2FarmNIntervalMinMs = 400
	instituteC2FarmNIntervalMaxMs = 700

	instituteC2FarmPatrolXMin                  = -70
	instituteC2FarmPatrolXMax                  = 130
	instituteC2FarmPatrolYMin                  = 130 // 下层挂机 relY 须 >130
	instituteC2FarmQuietXMin                   = 50  // 此区间尽量不攻击、不停顿
	instituteC2FarmQuietXMax                   = 85
	instituteC2FarmPatrolWalkChance            = 10 // 默认；中控「走路概率/瞬移概率」可覆盖
	instituteC2FarmPatrolTeleportChance        = 90
	instituteC2FarmPatrolJumpChance            = 1
	instituteC2FarmPatrolPauseChance           = 5
	instituteC2FarmPatrolTurnChance            = 3
	instituteC2FarmDoubleTeleportChance        = 10 // 连2次瞬移
	instituteC2FarmDownJumpWaitMs              = 400
	instituteC2FarmDownJumpMaxTry              = 8
	instituteC2FarmAttackHoldMinMs             = 300 // 默认；中控「攻击长按最短/最长毫秒」可覆盖
	instituteC2FarmAttackHoldMaxMs             = 400
	instituteC2FarmPlatformVisitMinSec         = 20 // 默认模式下层巡逻周期回站台；中控「巡逻时长最短/最长秒」可覆盖
	instituteC2FarmPlatformVisitMaxSec         = 30
	instituteC2FarmPlatformAttackMinSec        = 20 // 默认模式周期站撸时长；中控「站撸时长最短/最长秒」可覆盖
	instituteC2FarmPlatformAttackMaxSec        = 30
	instituteC2FarmPlatformStandTurnMinSec     = 20 // 站台站撸换向间隔
	instituteC2FarmPlatformStandTurnMaxSec     = 30
	instituteC2FarmPlatformStandTurnPauseMinMs = 600 // 换向前停止攻击
	instituteC2FarmPlatformStandTurnPauseMaxMs = 1000
	instituteC2FarmPlatformStandNBuffMinMin    = 29 // 站台站撸周期按N
	instituteC2FarmPlatformStandNBuffMaxMin    = 30
	instituteC2FarmPlatformStandNPressMin      = 11
	instituteC2FarmPlatformStandNPressMax      = 13
	instituteC2FarmPlatformStandXTol           = 3 // 站台站撸攻击 x≈-20±3
	instituteC2FarmMapCheckInterval            = 10 * time.Second
	instituteC2FarmYellowMissWalkThreshold     = 3               // 左上角OK但黄点连续未找到次数
	instituteC2FarmYellowMissWalkMs            = 300             // 恢复时左/右走时长
	instituteC2FarmRedColor                    = "237c63-b3b3b3" // RGB 35,124,99 偏色179
	instituteC2FarmRedSim                      = float32(0.95)
	instituteC2FarmLeaveConfirmReads           = 5
	instituteC2MagatiaJumpInterval             = 10 * time.Second
	instituteC2MagatiaPopupCheckInterval       = 3 * time.Second
	instituteC2MagatiaPopupColorX1             = 308
	instituteC2MagatiaPopupColorY1             = 475
	instituteC2MagatiaPopupColorX2             = 412
	instituteC2MagatiaPopupColorY2             = 491
	instituteC2MagatiaPopupColor               = "aacc15"
	instituteC2MagatiaPopupColorSim            = float32(0.95)
	instituteC2MagatiaPopupMinCount            = 30
	instituteC2MagatiaPopupClickX1             = 323
	instituteC2MagatiaPopupClickY1             = 478
	instituteC2MagatiaPopupClickX2             = 389
	instituteC2MagatiaPopupClickY2             = 488
	instituteC2MagatiaFallYThreshold           = 164 // relY>164 触发掉层恢复
	instituteC2MagatiaRecoverXMin              = 115 // 须 relX>115
	instituteC2MagatiaRecoverYTarget           = 161
	instituteC2MagatiaRecoverYTol              = 3
	instituteC2MagatiaExitWalkMs               = 200
	instituteC2MagatiaExitWalkNearMs           = 50   // x≈13 附近走出图，避免走多来回弹
	instituteC2MagatiaToAlcadnoDelayMs         = 3000 // 1→2 按上换图后等待

	instituteC2AlcadnoWhiteX1      = 95
	instituteC2AlcadnoWhiteY1      = 284
	instituteC2AlcadnoWhiteX2      = 1203
	instituteC2AlcadnoWhiteY2      = 490
	instituteC2AlcadnoWhiteColor   = "ffffff"
	instituteC2AlcadnoWhiteSim     = float32(0.97)
	instituteC2AlcadnoWhiteWinW    = 40
	instituteC2AlcadnoWhiteWinH    = 60
	instituteC2AlcadnoWhiteMinWin  = 1500
	instituteC2AlcadnoWhiteKeepMin = 3000 // 验图 OCR 未识别时 ffffff 仍≥此值视为仍在该图

	instituteC2AlcadnoWhiteWalkMs    = 8000
	instituteC2MagatiaUpTapMinMs     = 50
	instituteC2MagatiaUpTapMaxMs     = 80
	instituteC2MagatiaExitUpMinMs    = 200
	instituteC2MagatiaExitUpMaxMs    = 300
	instituteC2AlcadnoUpTapMinMs     = 220
	instituteC2AlcadnoUpTapMaxMs     = 300
	instituteC2AlcadnoWalkStepMs     = 2
	instituteC2AlcadnoMapCheckMs     = 1000
	instituteC2AlcadnoLeaveMapStreak = 2
	instituteC2AlcadnoRetryLeftTP    = 3
	instituteC2AlcadnoWhiteStuckMax  = 6 // ffffff已找到但未换图，连续此次数后左瞬移重试
)

type instituteC2AlcadnoWalkResult int

const (
	instituteC2AlcadnoWalkLeftMap instituteC2AlcadnoWalkResult = iota
	instituteC2AlcadnoWalkTimeoutOnMap
)

var instituteC2AttackHoldMinMs = instituteC2FarmAttackHoldMinMs
var instituteC2AttackHoldMaxMs = instituteC2FarmAttackHoldMaxMs
var instituteC2PlatformVisitMinSec = instituteC2FarmPlatformVisitMinSec
var instituteC2PlatformVisitMaxSec = instituteC2FarmPlatformVisitMaxSec
var instituteC2PlatformAttackMinMs = instituteC2FarmPlatformAttackMinSec * 1000
var instituteC2PlatformAttackMaxMs = instituteC2FarmPlatformAttackMaxSec * 1000
var instituteC2FarmPatrolWalkPct = instituteC2FarmPatrolWalkChance
var instituteC2FarmPatrolTeleportPct = instituteC2FarmPatrolTeleportChance
var instituteC2FarmYellowMissStreak int

type instituteC2FarmMode int

const (
	instituteC2FarmModeDefault instituteC2FarmMode = iota
	instituteC2FarmModeFullPatrol
	instituteC2FarmModePlatformStand
)

func instituteC2FarmModeFromConfig() instituteC2FarmMode {
	if core.API == nil {
		return instituteC2FarmModeDefault
	}
	switch strings.TrimSpace(core.API.GetConfigStringValue("挂机选项")) {
	case "全图巡逻":
		return instituteC2FarmModeFullPatrol
	case "站台站撸":
		return instituteC2FarmModePlatformStand
	default:
		return instituteC2FarmModeDefault
	}
}

func instituteC2FarmModeLabel(m instituteC2FarmMode) string {
	switch m {
	case instituteC2FarmModeFullPatrol:
		return "全图巡逻"
	case instituteC2FarmModePlatformStand:
		return "站台站撸"
	default:
		return "默认"
	}
}

func instituteC2InitDefaultModeTiming(logTag string, mode instituteC2FarmMode) {
	if mode != instituteC2FarmModeDefault {
		return
	}
	instituteC2PlatformVisitMinSec, instituteC2PlatformVisitMaxSec = IntSecRangeFromAPI(
		"巡逻时长最短秒", "巡逻时长最长秒",
		instituteC2FarmPlatformVisitMinSec, instituteC2FarmPlatformVisitMaxSec,
	)
	standMinSec, standMaxSec := IntSecRangeFromAPI(
		"站撸时长最短秒", "站撸时长最长秒",
		instituteC2FarmPlatformAttackMinSec, instituteC2FarmPlatformAttackMaxSec,
	)
	instituteC2PlatformAttackMinMs = standMinSec * 1000
	instituteC2PlatformAttackMaxMs = standMaxSec * 1000
	fmt.Printf("%s 默认模式 巡逻周期 %d-%ds 站撸 %d-%ds\n",
		logTag,
		instituteC2PlatformVisitMinSec, instituteC2PlatformVisitMaxSec,
		standMinSec, standMaxSec,
	)
}

func instituteC2InitPatrolMovePct(logTag string) {
	walk, tp := PercentPairFromAPI(
		"走路概率", "瞬移概率",
		instituteC2FarmPatrolWalkChance, instituteC2FarmPatrolTeleportChance,
	)
	instituteC2FarmPatrolWalkPct = walk
	instituteC2FarmPatrolTeleportPct = tp
	fmt.Printf("%s 巡逻移动 走路%d%% 瞬移%d%%\n", logTag, walk, tp)
}

func instituteC2FarmInPlatformStandAttackX(relX int) bool {
	return instituteC2FarmDist(relX) <= instituteC2FarmPlatformStandXTol
}

func instituteC2ScheduleNextPlatformStandTurn() time.Time {
	sec := instituteC2FarmPlatformStandTurnMinSec +
		rand.Intn(instituteC2FarmPlatformStandTurnMaxSec-instituteC2FarmPlatformStandTurnMinSec+1)
	return time.Now().Add(time.Duration(sec) * time.Second)
}

func instituteC2ScheduleNextPlatformStandNBuff() time.Time {
	minSec := instituteC2FarmPlatformStandNBuffMinMin * 60
	maxSec := instituteC2FarmPlatformStandNBuffMaxMin * 60
	if maxSec < minSec {
		maxSec = minSec
	}
	sec := minSec + rand.Intn(maxSec-minSec+1)
	return time.Now().Add(time.Duration(sec) * time.Second)
}

// Play_研究所C2 非挂机图每 200ms 条件 OCR；C-2 挂机图无间隔巡逻，每 10s 验图一次。
func Play_研究所C2(mapAssetPath string) error {
	cfg, err := loadMapConfig(mapAssetPath)
	if err != nil {
		return err
	}
	logTag := instituteC1LogTag(cfg.Name)
	instituteC2RequireAllowedAccount(cfg, logTag)
	SetFarmLogTag(logTag)
	StartFarmMaintainLoop(logTag)
	defer StopFarmMaintainLoop()
	SetFarmMaintainPotionEnabled(false)
	core.SetMinimapYellowRegion(instituteC2YellowX1, instituteC2YellowY1, instituteC2YellowX2, instituteC2YellowY2)
	defer core.ClearMinimapYellowRegion()
	fmt.Printf("%s 开始 非挂机图200ms OCR / C-2挂机10s验图\n", logTag)
	instituteC2AttackHoldMinMs, instituteC2AttackHoldMaxMs = AttackHoldMsFromAPI(
		instituteC2FarmAttackHoldMinMs, instituteC2FarmAttackHoldMaxMs,
	)
	fmt.Printf("%s 攻击长按 %d-%dms\n", logTag, instituteC2AttackHoldMinMs, instituteC2AttackHoldMaxMs)
	farmMode := instituteC2FarmModeFromConfig()
	fmt.Printf("%s 挂机选项: %s\n", logTag, instituteC2FarmModeLabel(farmMode))
	instituteC2InitDefaultModeTiming(logTag, farmMode)
	instituteC2InitPatrolMovePct(logTag)

	alcadnoMode := false
	alcadnoWhiteWalkDone := false
	alcadnoWhiteStuckCount := 0
	c2FarmMode := false
	c2FarmInitDone := false
	c2FarmGoRight := true
	var nextPlatformVisit time.Time
	c2FarmPlatformVisitActive := false
	var nextPlatformStandTurn time.Time
	var nextPlatformStandNBuff time.Time
	var lastFarmMapCheck time.Time
	lastOCRMap := ""
	magatiaMode := false
	var lastMagatiaJump time.Time
	var lastMagatiaPopupCheck time.Time
	clearBag := newAutoClearBagState()
	for {
		core.BlockWhileCaptchaHold()
		TickFarmMainThreadTasks()
		SetFarmMaintainPotionEnabled(c2FarmMode)

		if c2FarmMode {
			instituteC2HandleFarmMap(logTag, farmMode, &c2FarmInitDone, &clearBag, &c2FarmGoRight, &nextPlatformVisit, &c2FarmPlatformVisitActive, &nextPlatformStandTurn, &nextPlatformStandNBuff)
			if time.Since(lastFarmMapCheck) >= instituteC2FarmMapCheckInterval {
				lastFarmMapCheck = time.Now()
				mapText := instituteC2DetectMapText()
				display := mapText
				if display == "" {
					display = "（未识别）"
				}
				fmt.Printf("%s 挂机验图: %s\n", logTag, display)
				redN := core.CountMinimapColorDotsInRegion(
					instituteC2FarmYellowX1, instituteC2FarmYellowY1,
					instituteC2FarmYellowX2, instituteC2FarmYellowY2,
					instituteC2FarmRedColor, instituteC2FarmRedSim,
					5, 4,
				)
				fmt.Printf("%s C-2 红点数量: %d (区域同黄点)\n", logTag, redN)
				if !instituteC2IsFarmMap(mapText) {
					if left, text := instituteC2ConfirmLeftFarmMap(logTag); left {
						fmt.Printf("%s 连续%d次不在C-2 切换路线\n", logTag, instituteC2FarmLeaveConfirmReads)
						c2FarmMode = false
						instituteC2RouteAfterLeaveFarm(logTag, text, &alcadnoMode, &alcadnoWhiteWalkDone)
					}
				}
			}
			if c2FarmMode {
				continue
			}
		}

		mapText, mapRegion := instituteC2OCRMapText()
		if mapRegion != "" {
			if mapText == "" {
				mapText = "（未识别）"
			}
			fmt.Printf("%s 韩文OCR[%s]: %s\n", logTag, mapRegion, mapText)
			if instituteC2IsFarmMap(mapText) {
				if instituteC2IsCentralLab(lastOCRMap) {
					c2FarmInitDone = false
					nextPlatformVisit = time.Time{}
					c2FarmPlatformVisitActive = false
					nextPlatformStandTurn = time.Time{}
					nextPlatformStandNBuff = time.Time{}
					fmt.Printf("%s 从中央实验室进入C-2 重新初始化\n", logTag)
				}
				if !c2FarmMode {
					c2FarmGoRight = true
					lastFarmMapCheck = time.Now()
				}
				c2FarmMode = true
				alcadnoMode = false
				alcadnoWhiteWalkDone = false
				magatiaMode = false
			} else if instituteC2IsAlcadnoMagatia(mapText) {
				magatiaMode = false
				core.SetMinimapYellowRegion(instituteC2YellowX1, instituteC2YellowY1, instituteC2YellowX2, instituteC2YellowY2)
				if !alcadnoMode {
					alcadnoWhiteStuckCount = 0
				}
				alcadnoMode = true
			} else if instituteC2IsCentralLab(mapText) {
				alcadnoMode = false
				alcadnoWhiteWalkDone = false
				magatiaMode = false
				instituteC2HandleCentralLab(logTag)
			} else if instituteC2IsPlainMagatia(mapText) {
				alcadnoMode = false
				alcadnoWhiteWalkDone = false
				core.SetMinimapYellowRegion(instituteC2YellowX1, instituteC2YellowY1, instituteC2YellowX2, instituteC2YellowY2)
				if !magatiaMode {
					magatiaMode = true
					lastMagatiaJump = time.Now()
					lastMagatiaPopupCheck = time.Now()
				}
				instituteC2TickMagatiaJump(logTag, &lastMagatiaJump)
				instituteC2HandlePlainMagatia(logTag, mapText)
			} else {
				magatiaMode = false
			}
			if mapText != "（未识别）" {
				lastOCRMap = mapText
			}
		}

		if c2FarmMode {
			instituteC2HandleFarmMap(logTag, farmMode, &c2FarmInitDone, &clearBag, &c2FarmGoRight, &nextPlatformVisit, &c2FarmPlatformVisitActive, &nextPlatformStandTurn, &nextPlatformStandNBuff)
			continue
		}
		if alcadnoMode {
			instituteC2HandleAlcadnoMagatia(logTag, &alcadnoWhiteWalkDone, &alcadnoWhiteStuckCount)
		}
		if magatiaMode && !c2FarmMode {
			instituteC2TickMagatiaPopupDismiss(logTag, &lastMagatiaPopupCheck)
		}

		core.Sleep(200)
	}
}

func instituteC2LoginUsername() string {
	if u := storages.Get("data", "username"); u != "" {
		return u
	}
	if v, ok := core.Get("username").(string); ok {
		return v
	}
	return ""
}

func instituteC2RequireAllowedAccount(cfg *mapConfig, logTag string) {
	if cfg == nil || len(cfg.Accounts) == 0 {
		return
	}
	username := instituteC2LoginUsername()
	for _, a := range cfg.Accounts {
		if a == username {
			return
		}
	}
	msg := "暂未开发"
	fmt.Println(msg)
	core.SLS_Log2(msg)
	core.Sleep(3000)
	os.Exit(0)
}

func instituteC2IsPlainMagatia(text string) bool {
	return core.CalculateTextSimilarity(text, "마가티아") >= 0.8 && !instituteC2IsAlcadnoMagatia(text)
}

func instituteC2TickMagatiaJump(logTag string, last *time.Time) {
	if last == nil || time.Since(*last) < instituteC2MagatiaJumpInterval {
		return
	}
	*last = time.Now()
	fmt.Printf("%s 마가티아 定时右跳\n", logTag)
	tapJumpRight()
}

func instituteC2TickMagatiaPopupDismiss(logTag string, last *time.Time) {
	if last == nil || time.Since(*last) < instituteC2MagatiaPopupCheckInterval {
		return
	}
	*last = time.Now()
	n := core.Color.GetColorCountInRegion(
		instituteC2MagatiaPopupColorX1, instituteC2MagatiaPopupColorY1,
		instituteC2MagatiaPopupColorX2, instituteC2MagatiaPopupColorY2,
		instituteC2MagatiaPopupColor, instituteC2MagatiaPopupColorSim,
	)
	if n <= instituteC2MagatiaPopupMinCount {
		return
	}
	fmt.Printf("%s 마가티아 弹窗色块=%d>%d 点击[%d,%d,%d,%d]\n",
		logTag, n, instituteC2MagatiaPopupMinCount,
		instituteC2MagatiaPopupClickX1, instituteC2MagatiaPopupClickY1,
		instituteC2MagatiaPopupClickX2, instituteC2MagatiaPopupClickY2)
	core.RandomClickInArea(
		instituteC2MagatiaPopupClickX1, instituteC2MagatiaPopupClickY1,
		instituteC2MagatiaPopupClickX2, instituteC2MagatiaPopupClickY2,
	)
}

func instituteC2MagatiaRecoverYOK(relY int) bool {
	return matchRange(relY,
		instituteC2MagatiaRecoverYTarget-instituteC2MagatiaRecoverYTol,
		instituteC2MagatiaRecoverYTarget+instituteC2MagatiaRecoverYTol)
}

// instituteC2RecoverMagatiaFromFall relY>164：右瞬移到 relX>115，再上瞬移至 y=161±3。
func instituteC2RecoverMagatiaFromFall(logTag string, relX, relY int) bool {
	if relY <= instituteC2MagatiaFallYThreshold {
		return false
	}
	if instituteC2MagatiaRecoverYOK(relY) {
		return false
	}
	if relX <= instituteC2MagatiaRecoverXMin {
		fmt.Printf("%s 마가티아 掉层恢复 relX=%d relY=%d>%d 右瞬移→x>%d\n",
			logTag, relX, relY, instituteC2MagatiaFallYThreshold, instituteC2MagatiaRecoverXMin)
		tapTeleportWithDirection(true)
		sleepAfterTeleport()
		return true
	}
	fmt.Printf("%s 마가티아 掉层恢复 relX=%d>%d relY=%d 上瞬移→y=%d±%d\n",
		logTag, relX, instituteC2MagatiaRecoverXMin, relY,
		instituteC2MagatiaRecoverYTarget, instituteC2MagatiaRecoverYTol)
	tapUpTeleport()
	sleepAfterTeleport()
	return true
}

func instituteC2MagatiaDist(relX int) int {
	d := relX - 13
	if d < 0 {
		d = -d
	}
	return d
}

func instituteC2Near13(relX int) bool {
	return instituteC2MagatiaDist(relX) <= 4
}

func instituteC2StillOnMap() bool {
	text := core.OCR.DetectMultilineText(64, 26, 187, 85, "korean")
	if core.CalculateTextSimilarity(text, "마가티아") >= 0.8 {
		return true
	}
	a5 := core.Color.GetColorCountInRegion(6, 4, 291, 28, "a5a4a8", 0.95)
	dd := core.Color.GetColorCountInRegion(6, 4, 291, 28, "dd9911", 0.95)
	if a5 > 500 && dd > 200 {
		text = core.OCR.DetectMultilineText(4, 3, 198, 26, "korean")
		return core.CalculateTextSimilarity(text, "마가티아") >= 0.8
	}
	return false
}

func instituteC2AlignMagatiaX(logTag string, relX int) {
	dist := instituteC2MagatiaDist(relX)
	if dist <= 4 {
		return
	}
	if dist <= 7 {
		ms := dist * 45
		goRight := relX < 13
		dir := "左"
		if goRight {
			dir = "右"
		}
		fmt.Printf("%s relX=%d 近13 慢走%s %dms\n", logTag, relX, dir, ms)
		walkHoldMs(goRight, ms)
		return
	}
	if relX > 13 {
		fmt.Printf("%s relX=%d>13 左瞬移\n", logTag, relX)
		tapTeleportWithDirection(false)
	} else {
		fmt.Printf("%s relX=%d<13 右瞬移\n", logTag, relX)
		tapTeleportWithDirection(true)
	}
	sleepAfterTeleport()
}

func instituteC2MagatiaExitUpHoldMs() int {
	return instituteC2MagatiaExitUpMinMs + rand.Intn(instituteC2MagatiaExitUpMaxMs-instituteC2MagatiaExitUpMinMs+1)
}

func instituteC2MagatiaUpTapIntervalMs() int {
	return instituteC2MagatiaUpTapMinMs + rand.Intn(instituteC2MagatiaUpTapMaxMs-instituteC2MagatiaUpTapMinMs+1)
}

func instituteC2AlcadnoUpTapIntervalMs() int {
	return instituteC2AlcadnoUpTapMinMs + rand.Intn(instituteC2AlcadnoUpTapMaxMs-instituteC2AlcadnoUpTapMinMs+1)
}

func instituteC2MagatiaTapUpForMs(totalMs int) {
	elapsed := 0
	nextUpIn := 0
	step := instituteC2AlcadnoWalkStepMs
	for elapsed < totalMs {
		if remain := totalMs - elapsed; remain < step {
			step = remain
		}
		if nextUpIn <= 0 {
			motion.KeyAction(motion.KEYCODE_DPAD_UP, 0)
			nextUpIn = instituteC2MagatiaUpTapIntervalMs()
		}
		core.Sleep(step)
		nextUpIn -= step
		elapsed += step
	}
}

func instituteC2WalkWithUpMs(goRight bool, ms int) {
	dirCode := motion.KEYCODE_DPAD_LEFT
	if goRight {
		dirCode = motion.KEYCODE_DPAD_RIGHT
	}
	if ms < instituteC2AlcadnoWalkStepMs {
		ms = instituteC2AlcadnoWalkStepMs
	}
	elapsed := 0
	nextUpIn := 0
	step := instituteC2AlcadnoWalkStepMs
	for elapsed < ms {
		if remain := ms - elapsed; remain < step {
			step = remain
		}
		refreshDpadHold(dirCode, step)
		if nextUpIn <= 0 {
			motion.KeyAction(motion.KEYCODE_DPAD_UP, 0)
			nextUpIn = instituteC2MagatiaUpTapIntervalMs()
		}
		nextUpIn -= step
		elapsed += step
	}
	releaseDpadHold(dirCode)
}

func instituteC2MagatiaExitStepMs(relX int) int {
	if instituteC2Near13(relX) {
		return instituteC2MagatiaExitWalkNearMs
	}
	return instituteC2MagatiaExitWalkMs
}

func instituteC2WalkExitWithUp(logTag string) {
	goRight := true
	for instituteC2StillOnMap() {
		core.BlockWhileCaptchaHold()
		relX, _, ok := ReadMinimapRel()
		if ok {
			if relX <= 13 {
				goRight = true
			} else if relX >= 14 {
				goRight = false
			}
		}
		stepMs := instituteC2MagatiaExitStepMs(relX)
		dir := "左"
		if goRight {
			dir = "右"
		}
		fmt.Printf("%s 仍在地图 relX=%d 走%s+上 %dms\n", logTag, relX, dir, stepMs)
		instituteC2WalkWithUpMs(goRight, stepMs)
	}
	fmt.Printf("%s 已离开마가티아地图\n", logTag)
}

func instituteC2WaitAfterLeaveMagatia(logTag string) {
	fmt.Printf("%s 1→2 按上换图后等待%dms\n", logTag, instituteC2MagatiaToAlcadnoDelayMs)
	core.Sleep(instituteC2MagatiaToAlcadnoDelayMs)
}

func instituteC2HandleAtMagatia(logTag string, relX, relY int) {
	if !instituteC2Near13(relX) {
		return
	}
	if relY > 158 {
		fmt.Printf("%s relX=%d±4 relY=%d>158 跳跃\n", logTag, relX, relY)
		tapJump()
		return
	}
	holdMs := instituteC2MagatiaExitUpHoldMs()
	fmt.Printf("%s relX=%d 到13附近 按上%dms(50-80ms/次)\n", logTag, relX, holdMs)
	instituteC2MagatiaTapUpForMs(holdMs)
	if !instituteC2StillOnMap() {
		fmt.Printf("%s 按上后已离开地图\n", logTag)
		instituteC2WaitAfterLeaveMagatia(logTag)
		return
	}
	fmt.Printf("%s 按上%dms后仍在地图 继续走+上出图(50-80ms/次)\n", logTag, holdMs)
	instituteC2WalkExitWithUp(logTag)
	instituteC2WaitAfterLeaveMagatia(logTag)
}

func instituteC2IsAlcadnoMagatia(text string) bool {
	if text == "" || text == "（未识别）" {
		return false
	}
	if core.CalculateTextSimilarity(text, instituteC2AlcadnoMapName) >= 0.8 {
		return true
	}
	// OCR 常读成「마가티아 알카드노 협회」等混杂文本
	return strings.Contains(text, "알카드") && strings.Contains(text, "협회")
}

func instituteC2AlcadnoStillInMapDuringWalk(text string) bool {
	if instituteC2IsAlcadnoMagatia(text) {
		return true
	}
	if instituteC2IsCentralLab(text) || instituteC2IsFarmMap(text) {
		return false
	}
	return instituteC2CountAlcadnoWhiteRegion() >= instituteC2AlcadnoWhiteKeepMin
}

func instituteC2IsCentralLab(text string) bool {
	return core.CalculateTextSimilarity(text, instituteC2CentralLabMapName) >= 0.8
}

func instituteC2IsFarmMap(text string) bool {
	return core.CalculateTextSimilarity(text, instituteC2FarmMapName) >= 0.8
}

func instituteC2FarmDist(relX int) int {
	d := relX - instituteC2FarmTargetX
	if d < 0 {
		d = -d
	}
	return d
}

func instituteC2AlignFarmX(logTag string, relX int, alignTol int) {
	if alignTol <= 0 {
		alignTol = instituteC2FarmAlignXTol
	}
	dist := instituteC2FarmDist(relX)
	if dist <= alignTol {
		return
	}
	if dist <= instituteC2FarmTeleportDist {
		ms := dist * 45
		goRight := relX < instituteC2FarmTargetX
		dir := "左"
		if goRight {
			dir = "右"
		}
		fmt.Printf("%s %s relX=%d 近-20 慢走%s %dms\n", logTag, instituteC2FarmMapName, relX, dir, ms)
		walkHoldMs(goRight, ms)
		return
	}
	if relX > instituteC2FarmTargetX {
		fmt.Printf("%s %s relX=%d>-20 左瞬移\n", logTag, instituteC2FarmMapName, relX)
		tapTeleportWithDirection(false)
	} else {
		fmt.Printf("%s %s relX=%d<-20 右瞬移\n", logTag, instituteC2FarmMapName, relX)
		tapTeleportWithDirection(true)
	}
	sleepAfterTeleport()
}

func instituteC2AtFarmPlatform(relX, relY int) bool {
	return instituteC2FarmDist(relX) <= instituteC2FarmAlignXTol && relY < instituteC2FarmSuccessYMax
}

func instituteC2AtFarmShopSpot(relX, relY int) bool {
	return instituteC2AtFarmPlatform(relX, relY)
}

func instituteC2ScheduleNextPlatformVisit() time.Time {
	sec := instituteC2PlatformVisitMinSec + rand.Intn(instituteC2PlatformVisitMaxSec-instituteC2PlatformVisitMinSec+1)
	return time.Now().Add(time.Duration(sec) * time.Second)
}

func instituteC2PlatformStandAttack(logTag string) {
	fmt.Printf("%s %s 站台区站桩攻击 %d-%dms\n",
		logTag, instituteC2FarmMapName, instituteC2PlatformAttackMinMs, instituteC2PlatformAttackMaxMs)
	keyHoldPress(attackKeyCode(), instituteC2PlatformAttackMinMs, instituteC2PlatformAttackMaxMs)
}

// instituteC2TryPeriodicPlatformVisit 默认模式：下层巡逻周期回站台，站桩攻击 configurable 后下跳。
func instituteC2TryPeriodicPlatformVisit(logTag string, relX, relY int, nextVisit *time.Time, active *bool) bool {
	if nextVisit == nil {
		return false
	}
	nowDue := !nextVisit.IsZero() && !time.Now().Before(*nextVisit)
	if active == nil || (!*active && !nowDue) {
		return false
	}
	if !*active && nowDue {
		*active = true
		fmt.Printf("%s %s 周期回站台区开始 relX=%d relY=%d\n", logTag, instituteC2FarmMapName, relX, relY)
	}

	if instituteC2AtFarmPlatform(relX, relY) {
		fmt.Printf("%s %s 周期回站台区 到站 relX=%d relY=%d 站桩攻击\n",
			logTag, instituteC2FarmMapName, relX, relY)
		instituteC2PlatformStandAttack(logTag)
		instituteC2DownJumpUntilFarmLayer(logTag)
		*nextVisit = instituteC2ScheduleNextPlatformVisit()
		*active = false
		fmt.Printf("%s %s 周期回站台区完成 下次 %ds 后\n",
			logTag, instituteC2FarmMapName,
			int(time.Until(*nextVisit).Seconds()))
		return true
	}

	if instituteC2FarmDist(relX) > instituteC2FarmAlignXTol {
		fmt.Printf("%s %s 周期回站台区 对齐x≈-20 relX=%d relY=%d\n",
			logTag, instituteC2FarmMapName, relX, relY)
		instituteC2AlignFarmX(logTag, relX, 0)
		return true
	}

	if relY >= instituteC2FarmSuccessYMax {
		fmt.Printf("%s %s 周期回站台区 x≈-20就位 仅上瞬移 relX=%d relY=%d>=%d\n",
			logTag, instituteC2FarmMapName, relX, relY, instituteC2FarmSuccessYMax)
		tapUpTeleport()
		sleepAfterTeleport()
		return true
	}

	fmt.Printf("%s %s 周期回站台区 异常坐标 relX=%d relY=%d 尝试上瞬移\n",
		logTag, instituteC2FarmMapName, relX, relY)
	tapUpTeleport()
	sleepAfterTeleport()
	return true
}

func instituteC2TryFarmAutoShop(logTag string, s *autoClearBagState, relX, relY int) bool {
	if s == nil || !core.API.GetConfigBoolValue("自动清包") {
		return false
	}
	if !instituteC2AtFarmShopSpot(relX, relY) {
		return false
	}
	if !s.due() {
		return false
	}
	log := func(format string, args ...interface{}) {
		fmt.Printf(logTag+" "+format+"\n", args...)
	}
	log("自动清包: x=%d y=%d 准备清包", relX, relY)
	return s.tryAutoShop(log)
}

func instituteC2ReadFarmRel() (relX, relY int, ok bool) {
	core.SetMinimapYellowRegion(
		instituteC2FarmYellowX1, instituteC2FarmYellowY1,
		instituteC2FarmYellowX2, instituteC2FarmYellowY2,
	)
	return ReadMinimapRel()
}

func instituteC2PressNCount(logTag, reason string, nMin, nMax int) {
	n := nMin + rand.Intn(nMax-nMin+1)
	code := parseAutoKeyCode("N")
	fmt.Printf("%s %s %s 按N %d次 间隔400-700ms\n", logTag, instituteC2FarmMapName, reason, n)
	for i := 0; i < n; i++ {
		core.BlockWhileCaptchaHold()
		keyTapAction(code)
		if i < n-1 {
			core.RandomSleep(instituteC2FarmNIntervalMinMs, instituteC2FarmNIntervalMaxMs)
		}
	}
}

func instituteC2PressNFarmBuff(logTag string) {
	instituteC2PressNCount(logTag, "商店后", instituteC2FarmNPressMin, instituteC2FarmNPressMax)
}

func instituteC2PressNPlatformStandPeriodic(logTag string) {
	instituteC2PressNCount(logTag, "站台站撸周期", instituteC2FarmPlatformStandNPressMin, instituteC2FarmPlatformStandNPressMax)
}

func instituteC2TryPlatformStandPeriodicNBuff(logTag string, nextNBuff *time.Time) bool {
	if nextNBuff == nil {
		return false
	}
	if nextNBuff.IsZero() {
		*nextNBuff = instituteC2ScheduleNextPlatformStandNBuff()
		return false
	}
	if time.Now().Before(*nextNBuff) {
		return false
	}
	fmt.Printf("%s %s 站台站撸 周期按N到期 原地按N\n", logTag, instituteC2FarmMapName)
	instituteC2PressNPlatformStandPeriodic(logTag)
	*nextNBuff = instituteC2ScheduleNextPlatformStandNBuff()
	fmt.Printf("%s %s 站台站撸 下次周期按N %dmin 后\n",
		logTag, instituteC2FarmMapName, int(time.Until(*nextNBuff).Minutes()))
	return true
}

func instituteC2DownJumpUntilFarmLayer(logTag string) {
	for i := 0; i < instituteC2FarmDownJumpMaxTry; i++ {
		relX, relY, ok := instituteC2ReadFarmRel()
		if ok && relY > instituteC2FarmPatrolYMin {
			fmt.Printf("%s %s 下层就位 relX=%d relY=%d>130\n", logTag, instituteC2FarmMapName, relX, relY)
			return
		}
		y := relY
		if !ok {
			y = -1
		}
		fmt.Printf("%s %s 下跳 relY=%d需>130 (%d/%d)\n", logTag, instituteC2FarmMapName, y, i+1, instituteC2FarmDownJumpMaxTry)
		tapDownJump()
		core.Sleep(instituteC2FarmDownJumpWaitMs)
	}
}

func instituteC2AfterShopRoutine(logTag string, mode instituteC2FarmMode) {
	if mode == instituteC2FarmModePlatformStand {
		instituteC2PressNPlatformStandPeriodic(logTag)
		return
	}
	instituteC2PressNFarmBuff(logTag)
	instituteC2DownJumpUntilFarmLayer(logTag)
}

func instituteC2FinishFarmInit(logTag string, mode instituteC2FarmMode, clearBag *autoClearBagState, relX, relY int, nextPlatformVisit *time.Time) {
	if instituteC2TryFarmAutoShop(logTag, clearBag, relX, relY) {
		instituteC2AfterShopRoutine(logTag, mode)
	} else if mode == instituteC2FarmModePlatformStand {
		if clearBag != nil && clearBag.due() {
			fmt.Printf("%s %s 站台站撸就位 等待清包 跳过初始化按N\n", logTag, instituteC2FarmMapName)
		} else {
			fmt.Printf("%s %s 站台站撸就位 按N留站台\n", logTag, instituteC2FarmMapName)
			instituteC2PressNPlatformStandPeriodic(logTag)
		}
	} else {
		fmt.Printf("%s %s 站台区就位 按N+下跳(不清包也执行)\n", logTag, instituteC2FarmMapName)
		instituteC2PressNFarmBuff(logTag)
		instituteC2DownJumpUntilFarmLayer(logTag)
	}
	if mode == instituteC2FarmModeDefault && nextPlatformVisit != nil {
		*nextPlatformVisit = instituteC2ScheduleNextPlatformVisit()
	}
}

func instituteC2TickPlatformStandTurn(logTag string, goRight *bool, nextTurn *time.Time) {
	if goRight == nil || nextTurn == nil {
		return
	}
	if !nextTurn.IsZero() && time.Now().Before(*nextTurn) {
		return
	}
	first := nextTurn.IsZero()
	if !first {
		pauseMs := instituteC2FarmPlatformStandTurnPauseMinMs +
			rand.Intn(instituteC2FarmPlatformStandTurnPauseMaxMs-instituteC2FarmPlatformStandTurnPauseMinMs+1)
		fmt.Printf("%s %s 站台站撸 换向前停止攻击 %dms\n", logTag, instituteC2FarmMapName, pauseMs)
		core.Sleep(pauseMs)
		*goRight = !*goRight
	}
	*nextTurn = instituteC2ScheduleNextPlatformStandTurn()
	dir := "左"
	if *goRight {
		faceRight()
		dir = "右"
	} else {
		faceLeft()
	}
	if first {
		fmt.Printf("%s %s 站台站撸 初始朝向%s 下次换向 %ds 后\n",
			logTag, instituteC2FarmMapName, dir, int(time.Until(*nextTurn).Seconds()))
	} else {
		fmt.Printf("%s %s 站台站撸 换向%s 下次换向 %ds 后\n",
			logTag, instituteC2FarmMapName, dir, int(time.Until(*nextTurn).Seconds()))
	}
}

func instituteC2HandlePlatformStandFarm(logTag string, relX, relY int, goRight *bool, nextTurn *time.Time, nextNBuff *time.Time, clearBag *autoClearBagState) {
	if clearBag != nil && clearBag.pendingShop && !instituteC2AtFarmPlatform(relX, relY) {
		log := func(format string, args ...interface{}) {
			fmt.Printf(logTag+" "+format+"\n", args...)
		}
		log("自动清包: 已离开站台区 x=%d y=%d 放弃本次", relX, relY)
		clearBag.finishAttempt(log)
		return
	}
	if clearBag != nil && clearBag.due() && !instituteC2AtFarmPlatform(relX, relY) {
		instituteC2ReturnToFarmShopSpot(logTag, relX, relY)
		return
	}
	if instituteC2AtFarmPlatform(relX, relY) {
		if instituteC2TryFarmAutoShop(logTag, clearBag, relX, relY) {
			instituteC2PressNPlatformStandPeriodic(logTag)
			if nextTurn != nil {
				*nextTurn = time.Time{}
			}
			if nextNBuff != nil {
				*nextNBuff = instituteC2ScheduleNextPlatformStandNBuff()
			}
			return
		}
		if clearBag == nil || (!clearBag.due() && !clearBag.pendingShop) {
			if instituteC2TryPlatformStandPeriodicNBuff(logTag, nextNBuff) {
				return
			}
		}
		if !instituteC2FarmInPlatformStandAttackX(relX) {
			fmt.Printf("%s %s 站台站撸 对齐x≈-20±%d relX=%d\n",
				logTag, instituteC2FarmMapName, instituteC2FarmPlatformStandXTol, relX)
			instituteC2AlignFarmX(logTag, relX, instituteC2FarmPlatformStandXTol)
			return
		}
		instituteC2TickPlatformStandTurn(logTag, goRight, nextTurn)
		dir := "左"
		if goRight != nil && *goRight {
			dir = "右"
		}
		fmt.Printf("%s %s 站台站撸 攻击%s relX=%d relY=%d\n", logTag, instituteC2FarmMapName, dir, relX, relY)
		keyHoldPress(attackKeyCode(), instituteC2AttackHoldMinMs, instituteC2AttackHoldMaxMs)
		return
	}
	if relY >= instituteC2FarmSuccessYMax {
		if instituteC2FarmDist(relX) > instituteC2FarmAlignXTol {
			fmt.Printf("%s %s 站台站撸 回站台 对齐x≈-20 relX=%d relY=%d\n",
				logTag, instituteC2FarmMapName, relX, relY)
			instituteC2AlignFarmX(logTag, relX, 0)
			return
		}
		fmt.Printf("%s %s 站台站撸 回站台 上瞬移 relX=%d relY=%d\n", logTag, instituteC2FarmMapName, relX, relY)
		tapUpTeleport()
		sleepAfterTeleport()
		return
	}
	if instituteC2FarmDist(relX) > instituteC2FarmAlignXTol {
		instituteC2AlignFarmX(logTag, relX, 0)
		return
	}
	fmt.Printf("%s %s 站台站撸 回站台 上瞬移 relX=%d relY=%d\n", logTag, instituteC2FarmMapName, relX, relY)
	tapUpTeleport()
	sleepAfterTeleport()
}

func instituteC2ReturnToFarmShopSpot(logTag string, relX, relY int) {
	if instituteC2FarmDist(relX) > instituteC2FarmAlignXTol {
		instituteC2AlignFarmX(logTag, relX, 0)
		return
	}
	if relY >= instituteC2FarmSuccessYMax {
		fmt.Printf("%s %s 回站台区 relX=%d relY=%d 上瞬移\n", logTag, instituteC2FarmMapName, relX, relY)
		tapUpTeleport()
		sleepAfterTeleport()
	}
}

func instituteC2FarmPatrolMargin() int {
	return linearPatrolMargin(linearBounds{
		XMin: instituteC2FarmPatrolXMin,
		XMax: instituteC2FarmPatrolXMax,
	})
}

func instituteC2FarmInQuietZone(relX int) bool {
	return matchRange(relX, instituteC2FarmQuietXMin, instituteC2FarmQuietXMax)
}

func instituteC2FarmDoAttack(logTag, tag string, relX int) {
	if instituteC2FarmInQuietZone(relX) {
		fmt.Printf("%s %s 静默区[%d,%d] relX=%d 不攻击\n",
			logTag, instituteC2FarmMapName, instituteC2FarmQuietXMin, instituteC2FarmQuietXMax, relX)
		return
	}
	fmt.Printf("%s %s 攻击: %s 按住空格%d-%dms relX=%d\n",
		logTag, instituteC2FarmMapName, tag, instituteC2AttackHoldMinMs, instituteC2AttackHoldMaxMs, relX)
	keyHoldPress(attackKeyCode(), instituteC2AttackHoldMinMs, instituteC2AttackHoldMaxMs)
}

func instituteC2FarmDoAttackAfterMove(logTag, tag string) {
	relX, _, ok := instituteC2ReadFarmRel()
	if !ok {
		return
	}
	instituteC2FarmDoAttack(logTag, tag, relX)
}

func instituteC2FarmPatrolTeleport(logTag, label string, goRight bool, relX int, attackTag string) {
	times := 1
	if rand.Intn(100) < instituteC2FarmDoubleTeleportChance {
		times = 2
		label += "×2"
	}
	fmt.Printf("%s %s %s relX=%d\n", logTag, instituteC2FarmMapName, label, relX)
	for i := 0; i < times; i++ {
		tapTeleportWithDirection(goRight)
		sleepAfterTeleport()
	}
	instituteC2FarmDoAttackAfterMove(logTag, attackTag)
}

func instituteC2FarmPatrolStep(logTag string, relX int, goRight bool) bool {
	margin := instituteC2FarmPatrolMargin()
	xMin, xMax := instituteC2FarmPatrolXMin, instituteC2FarmPatrolXMax

	if relX < xMin {
		instituteC2FarmPatrolTeleport(logTag, "出界 右瞬移回区", true, relX, "回区")
		return true
	}
	if relX > xMax {
		instituteC2FarmPatrolTeleport(logTag, "出界 左瞬移回区", false, relX, "回区")
		return false
	}

	if goRight && relX >= xMax-margin {
		fmt.Printf("%s %s 近右界 relX=%d 改向左\n", logTag, instituteC2FarmMapName, relX)
		core.Sleep(80)
		return false
	}
	if !goRight && relX <= xMin+margin {
		fmt.Printf("%s %s 近左界 relX=%d 改向右\n", logTag, instituteC2FarmMapName, relX)
		core.Sleep(80)
		return true
	}

	if rand.Intn(100) < instituteC2FarmPatrolTurnChance {
		goRight = !goRight
		if goRight {
			fmt.Printf("%s %s 随机换向: 改向右 relX=%d\n", logTag, instituteC2FarmMapName, relX)
		} else {
			fmt.Printf("%s %s 随机换向: 改向左 relX=%d\n", logTag, instituteC2FarmMapName, relX)
		}
	}

	dir := "左"
	if goRight {
		dir = "右"
	}

	if instituteC2FarmInQuietZone(relX) {
		instituteC2FarmPatrolTeleport(logTag,
			fmt.Sprintf("静默区[%d,%d] %s穿过", instituteC2FarmQuietXMin, instituteC2FarmQuietXMax, dir),
			goRight, relX, dir+"瞬移")
		return goRight
	}

	roll := rand.Intn(instituteC2FarmPatrolWalkPct + instituteC2FarmPatrolTeleportPct)
	if roll < instituteC2FarmPatrolWalkPct {
		fmt.Printf("%s %s 移动: %s走 relX=%d\n", logTag, instituteC2FarmMapName, dir, relX)
		if goRight {
			faceRight()
		} else {
			faceLeft()
		}
		walkHoldMs(goRight, patrolFarmWalkMs())
		instituteC2FarmDoAttackAfterMove(logTag, dir+"走")
	} else {
		instituteC2FarmPatrolTeleport(logTag, fmt.Sprintf("移动: %s瞬移", dir), goRight, relX, dir+"瞬移")
	}

	if rand.Intn(100) < 30 {
		core.RandomSleep(50, 180)
	}
	return goRight
}

func instituteC2TryRecoverYellowMiss(logTag string, worldFound, yellowFound bool) bool {
	if yellowFound || !worldFound {
		instituteC2FarmYellowMissStreak = 0
		return false
	}
	instituteC2FarmYellowMissStreak++
	if instituteC2FarmYellowMissStreak < instituteC2FarmYellowMissWalkThreshold {
		return false
	}
	goRight := rand.Intn(2) == 0
	dir := "左"
	if goRight {
		dir = "右"
	}
	fmt.Printf("%s %s 左上角OK黄点连续%d次未找到 %s走%dms\n",
		logTag, instituteC2FarmMapName,
		instituteC2FarmYellowMissStreak, dir, instituteC2FarmYellowMissWalkMs)
	if goRight {
		faceRight()
	} else {
		faceLeft()
	}
	walkHoldMs(goRight, instituteC2FarmYellowMissWalkMs)
	instituteC2FarmYellowMissStreak = 0
	return true
}

func instituteC2HandleFarmMap(logTag string, mode instituteC2FarmMode, initDone *bool, clearBag *autoClearBagState, goRight *bool, nextPlatformVisit *time.Time, platformVisitActive *bool, nextPlatformStandTurn *time.Time, nextPlatformStandNBuff *time.Time) {
	core.SetMinimapYellowRegion(
		instituteC2FarmYellowX1, instituteC2FarmYellowY1,
		instituteC2FarmYellowX2, instituteC2FarmYellowY2,
	)
	relX, relY, ok, detail, worldFound, yellowFound := ReadMinimapRelWithDetailEx()
	if !ok {
		fmt.Printf("%s %s 小地图未识别: %s\n", logTag, instituteC2FarmMapName, detail)
		instituteC2TryRecoverYellowMiss(logTag, worldFound, yellowFound)
		return
	}
	instituteC2FarmYellowMissStreak = 0
	fmt.Printf("%s %s %s\n", logTag, instituteC2FarmMapName, detail)

	if initDone != nil && *initDone {
		if mode == instituteC2FarmModePlatformStand {
			instituteC2HandlePlatformStandFarm(logTag, relX, relY, goRight, nextPlatformStandTurn, nextPlatformStandNBuff, clearBag)
			return
		}

		visitActive := platformVisitActive != nil && *platformVisitActive

		if clearBag != nil && clearBag.pendingShop && !instituteC2AtFarmPlatform(relX, relY) {
			log := func(format string, args ...interface{}) {
				fmt.Printf(logTag+" "+format+"\n", args...)
			}
			log("自动清包: 已离开站台区 x=%d y=%d 放弃本次", relX, relY)
			clearBag.finishAttempt(log)
			return
		}

		if mode == instituteC2FarmModeDefault && visitActive {
			if instituteC2TryPeriodicPlatformVisit(logTag, relX, relY, nextPlatformVisit, platformVisitActive) {
				return
			}
		}

		if clearBag != nil && clearBag.due() && !instituteC2AtFarmPlatform(relX, relY) {
			instituteC2ReturnToFarmShopSpot(logTag, relX, relY)
			return
		}
		if instituteC2AtFarmPlatform(relX, relY) {
			if instituteC2TryFarmAutoShop(logTag, clearBag, relX, relY) {
				instituteC2AfterShopRoutine(logTag, mode)
				if mode == instituteC2FarmModeDefault && nextPlatformVisit != nil {
					*nextPlatformVisit = instituteC2ScheduleNextPlatformVisit()
				}
				if platformVisitActive != nil {
					*platformVisitActive = false
				}
			} else {
				fmt.Printf("%s %s 清包未到间隔 下跳至下层\n", logTag, instituteC2FarmMapName)
				instituteC2DownJumpUntilFarmLayer(logTag)
			}
			return
		}
		if relY <= instituteC2FarmPatrolYMin {
			if !instituteC2AtFarmPlatform(relX, relY) {
				instituteC2ReturnToFarmShopSpot(logTag, relX, relY)
				return
			}
			fmt.Printf("%s %s 清包未到间隔 下跳至下层\n", logTag, instituteC2FarmMapName)
			instituteC2DownJumpUntilFarmLayer(logTag)
			return
		}
		if mode == instituteC2FarmModeDefault {
			if instituteC2TryPeriodicPlatformVisit(logTag, relX, relY, nextPlatformVisit, platformVisitActive) {
				return
			}
		}
		if goRight != nil {
			*goRight = instituteC2FarmPatrolStep(logTag, relX, *goRight)
		}
		return
	}

	if initDone != nil && !*initDone {
		if instituteC2FarmDist(relX) > instituteC2FarmAlignXTol {
			instituteC2AlignFarmX(logTag, relX, 0)
			return
		}
		if relY >= instituteC2FarmSuccessYMax {
			fmt.Printf("%s %s 站台区 relX=%d relY=%d>=%d 上瞬移\n", logTag, instituteC2FarmMapName, relX, relY, instituteC2FarmSuccessYMax)
			tapUpTeleport()
			sleepAfterTeleport()
			return
		}
		fmt.Printf("%s %s 站台区就位 relX=%d relY=%d(<%d)\n", logTag, instituteC2FarmMapName, relX, relY, instituteC2FarmSuccessYMax)
		instituteC2FinishFarmInit(logTag, mode, clearBag, relX, relY, nextPlatformVisit)
		*initDone = true
		if mode == instituteC2FarmModePlatformStand && nextPlatformStandTurn != nil {
			*nextPlatformStandTurn = time.Time{}
		}
		if mode == instituteC2FarmModePlatformStand && nextPlatformStandNBuff != nil {
			*nextPlatformStandNBuff = instituteC2ScheduleNextPlatformStandNBuff()
		}
		if goRight != nil {
			*goRight = true
		}
	}
}

func instituteC2StillOnCentralLab() bool {
	return instituteC2IsCentralLab(instituteC2DetectMapText())
}

func instituteC2ReadCentralLabRel() (relX, relY int, ok bool) {
	core.SetMinimapYellowRegion(
		instituteC2CentralLabYellowX1, instituteC2CentralLabYellowY1,
		instituteC2CentralLabYellowX2, instituteC2CentralLabYellowY2,
	)
	return ReadMinimapRel()
}

func instituteC2TryUpLeaveCentralLab(logTag string) bool {
	fmt.Printf("%s %s 按上1s\n", logTag, instituteC2CentralLabMapName)
	refreshDpadUpHold(instituteC2CentralLabUpHoldMs)
	releaseDpadUp()
	if !instituteC2StillOnCentralLab() {
		fmt.Printf("%s %s 按上后已换图\n", logTag, instituteC2CentralLabMapName)
		return true
	}
	fmt.Printf("%s %s 按上后仍在该图\n", logTag, instituteC2CentralLabMapName)
	return false
}

func instituteC2AlignCentralLabX(logTag string, relX int) {
	dist := relX - instituteC2CentralLabTargetX
	if dist < 0 {
		dist = -dist
	}
	if dist == 0 {
		return
	}
	ms := dist * 45
	goRight := relX < instituteC2CentralLabTargetX
	dir := "左"
	if goRight {
		dir = "右"
	}
	fmt.Printf("%s %s relX=%d 走向36 走%s %dms\n", logTag, instituteC2CentralLabMapName, relX, dir, ms)
	walkHoldMs(goRight, ms)
}

func instituteC2WalkPingPongUpCentralLab(logTag string) {
	fmt.Printf("%s %s 35-37来回走按上直到换图\n", logTag, instituteC2CentralLabMapName)
	goRight := true
	for instituteC2StillOnCentralLab() {
		core.BlockWhileCaptchaHold()
		relX, _, ok := instituteC2ReadCentralLabRel()
		if ok {
			if relX <= instituteC2CentralLabPingMinX {
				goRight = true
			} else if relX >= instituteC2CentralLabPingMaxX {
				goRight = false
			}
		}
		dir := "左"
		if goRight {
			dir = "右"
		}
		fmt.Printf("%s %s relX=%d 走%s %dms\n", logTag, instituteC2CentralLabMapName, relX, dir, instituteC2CentralLabWalkStepMs)
		walkHoldMs(goRight, instituteC2CentralLabWalkStepMs)
		if instituteC2TryUpLeaveCentralLab(logTag) {
			return
		}
	}
}

func instituteC2HandleCentralLab(logTag string) {
	core.SetMinimapYellowRegion(
		instituteC2CentralLabYellowX1, instituteC2CentralLabYellowY1,
		instituteC2CentralLabYellowX2, instituteC2CentralLabYellowY2,
	)
	relX, relY, ok, detail := ReadMinimapRelWithDetail()
	if !ok {
		fmt.Printf("%s %s 小地图未识别: %s\n", logTag, instituteC2CentralLabMapName, detail)
		return
	}
	fmt.Printf("%s %s %s\n", logTag, instituteC2CentralLabMapName, detail)

	if relX != instituteC2CentralLabTargetX {
		instituteC2AlignCentralLabX(logTag, relX)
		if relX2, relY2, ok2 := instituteC2ReadCentralLabRel(); ok2 {
			relX, relY = relX2, relY2
			fmt.Printf("%s %s 对齐后黄点 relX=%d relY=%d\n", logTag, instituteC2CentralLabMapName, relX, relY)
		}
	}

	if !instituteC2StillOnCentralLab() {
		return
	}
	if instituteC2TryUpLeaveCentralLab(logTag) {
		return
	}
	instituteC2WalkPingPongUpCentralLab(logTag)
}

func instituteC2OCRMapText() (mapText, mapRegion string) {
	a5 := core.Color.GetColorCountInRegion(6, 4, 291, 28, "a5a4a8", 0.95)
	dd := core.Color.GetColorCountInRegion(6, 4, 291, 28, "dd9911", 0.95)
	if a5 > 500 && dd > 200 {
		return core.OCR.DetectMultilineText(4, 3, 198, 26, "korean"), "4,3,198,26"
	}
	n := core.Color.GetColorCountInRegion(174, 38, 261, 75, "99bbcc", 0.95)
	if n > 1000 {
		return core.OCR.DetectMultilineText(64, 26, 187, 85, "korean"), "64,26,187,85"
	}
	return "", ""
}

func instituteC2ConfirmLeftFarmMap(logTag string) (left bool, lastText string) {
	for i := 0; i < instituteC2FarmLeaveConfirmReads; i++ {
		text := instituteC2DetectMapText()
		display := text
		if display == "" {
			display = "（未识别）"
		}
		fmt.Printf("%s 离图确认 %d/%d: %s\n", logTag, i+1, instituteC2FarmLeaveConfirmReads, display)
		if instituteC2IsFarmMap(text) {
			fmt.Printf("%s 仍在C-2 继续挂机\n", logTag)
			return false, text
		}
		lastText = text
	}
	return true, lastText
}

func instituteC2RouteAfterLeaveFarm(logTag, mapText string, alcadnoMode, alcadnoWhiteWalkDone *bool) {
	if mapText == "" || mapText == "（未识别）" {
		mapText = instituteC2DetectMapText()
	}
	display := mapText
	if display == "" {
		display = "（未识别）"
	}
	fmt.Printf("%s 离图后路线: %s\n", logTag, display)
	if instituteC2IsAlcadnoMagatia(mapText) {
		core.SetMinimapYellowRegion(instituteC2YellowX1, instituteC2YellowY1, instituteC2YellowX2, instituteC2YellowY2)
		*alcadnoMode = true
		*alcadnoWhiteWalkDone = false
		return
	}
	*alcadnoMode = false
	*alcadnoWhiteWalkDone = false
	if instituteC2IsCentralLab(mapText) {
		instituteC2HandleCentralLab(logTag)
		return
	}
	core.SetMinimapYellowRegion(instituteC2YellowX1, instituteC2YellowY1, instituteC2YellowX2, instituteC2YellowY2)
	instituteC2HandlePlainMagatia(logTag, mapText)
}

func instituteC2DetectMapText() string {
	text, _ := instituteC2OCRMapText()
	return text
}

func instituteC2StillOnAlcadnoMap() bool {
	return instituteC2IsAlcadnoMagatia(instituteC2DetectMapText())
}

func instituteC2ReleaseRightAndUp() {
	releaseDpadHold(motion.KEYCODE_DPAD_RIGHT)
	releaseDpadUp()
}

func instituteC2AlcadnoLeftTeleportN(logTag string, n int) {
	fmt.Printf("%s %s 仍在地图 左瞬移%d次后重来\n", logTag, instituteC2AlcadnoMapName, n)
	for i := 0; i < n; i++ {
		core.BlockWhileCaptchaHold()
		fmt.Printf("%s %s 左瞬移 %d/%d\n", logTag, instituteC2AlcadnoMapName, i+1, n)
		tapTeleportWithDirection(false)
		sleepAfterTeleport()
	}
}

func instituteC2CountAlcadnoWhiteRegion() int {
	return core.Color.GetColorCountInRegion(
		instituteC2AlcadnoWhiteX1, instituteC2AlcadnoWhiteY1,
		instituteC2AlcadnoWhiteX2, instituteC2AlcadnoWhiteY2,
		instituteC2AlcadnoWhiteColor, instituteC2AlcadnoWhiteSim,
	)
}

func instituteC2FindAlcadnoWhiteCenter() (cx, cy int, ok bool) {
	cx, cy = core.Color.FindColorWindowPeakCenter(
		instituteC2AlcadnoWhiteX1, instituteC2AlcadnoWhiteY1,
		instituteC2AlcadnoWhiteX2, instituteC2AlcadnoWhiteY2,
		instituteC2AlcadnoWhiteColor, instituteC2AlcadnoWhiteSim,
		instituteC2AlcadnoWhiteWinW, instituteC2AlcadnoWhiteWinH, instituteC2AlcadnoWhiteMinWin,
	)
	return cx, cy, cx >= 0
}

func instituteC2HoldRightWalkWithUp(logTag string) instituteC2AlcadnoWalkResult {
	fmt.Printf("%s %s 首次ffffff 按住右走%dms 每%d-%dms按上 每秒验图\n",
		logTag, instituteC2AlcadnoMapName, instituteC2AlcadnoWhiteWalkMs,
		instituteC2AlcadnoUpTapMinMs, instituteC2AlcadnoUpTapMaxMs)
	elapsed := 0
	nextUpIn := 0
	nextMapCheckIn := instituteC2AlcadnoMapCheckMs
	mapMissStreak := 0
	upEnabled := true
	for elapsed < instituteC2AlcadnoWhiteWalkMs {
		core.BlockWhileCaptchaHold()
		step := instituteC2AlcadnoWalkStepMs
		if remain := instituteC2AlcadnoWhiteWalkMs - elapsed; remain < step {
			step = remain
		}
		refreshDpadHold(motion.KEYCODE_DPAD_RIGHT, step)
		if upEnabled && nextUpIn <= 0 {
			motion.KeyAction(motion.KEYCODE_DPAD_UP, 0)
			nextUpIn = instituteC2AlcadnoUpTapIntervalMs()
		}
		if upEnabled {
			nextUpIn -= step
		}
		elapsed += step
		nextMapCheckIn -= step
		if nextMapCheckIn <= 0 {
			mapText := instituteC2DetectMapText()
			if instituteC2AlcadnoStillInMapDuringWalk(mapText) {
				mapMissStreak = 0
			} else {
				mapMissStreak++
				fmt.Printf("%s %s 右走验图未识别地图 %d/%d\n",
					logTag, instituteC2AlcadnoMapName, mapMissStreak, instituteC2AlcadnoLeaveMapStreak)
				if mapMissStreak >= 1 && upEnabled {
					upEnabled = false
					releaseDpadUp()
					fmt.Printf("%s %s 验图未识别1/%d 停止按上\n",
						logTag, instituteC2AlcadnoMapName, instituteC2AlcadnoLeaveMapStreak)
				}
				if mapMissStreak >= instituteC2AlcadnoLeaveMapStreak {
					fmt.Printf("%s %s 连续%d次扫不到地图 松开右键\n",
						logTag, instituteC2AlcadnoMapName, instituteC2AlcadnoLeaveMapStreak)
					instituteC2ReleaseRightAndUp()
					return instituteC2AlcadnoWalkLeftMap
				}
			}
			nextMapCheckIn = instituteC2AlcadnoMapCheckMs
		}
	}
	instituteC2ReleaseRightAndUp()
	if mapMissStreak >= instituteC2AlcadnoLeaveMapStreak {
		return instituteC2AlcadnoWalkLeftMap
	}
	fmt.Printf("%s %s 右走%dms结束仍在地图(未连续%d次未识别)\n",
		logTag, instituteC2AlcadnoMapName, instituteC2AlcadnoWhiteWalkMs, instituteC2AlcadnoLeaveMapStreak)
	return instituteC2AlcadnoWalkTimeoutOnMap
}

func instituteC2HandleAlcadnoMagatia(logTag string, whiteWalkDone *bool, whiteStuckCount *int) {
	whiteN := instituteC2CountAlcadnoWhiteRegion()
	whiteX, whiteY, whiteOK := instituteC2FindAlcadnoWhiteCenter()
	if whiteOK {
		fmt.Printf("%s %s ffffff区域=%d ffffff中点(%d,%d)\n",
			logTag, instituteC2AlcadnoMapName, whiteN, whiteX, whiteY)
		if whiteWalkDone != nil && !*whiteWalkDone {
			*whiteWalkDone = true
			if whiteStuckCount != nil {
				*whiteStuckCount = 0
			}
			switch instituteC2HoldRightWalkWithUp(logTag) {
			case instituteC2AlcadnoWalkLeftMap:
				if whiteWalkDone != nil {
					*whiteWalkDone = false
				}
				if whiteStuckCount != nil {
					*whiteStuckCount = 0
				}
				mapText := instituteC2DetectMapText()
				if instituteC2IsCentralLab(mapText) || instituteC2IsFarmMap(mapText) {
					fmt.Printf("%s %s 右走验图已换图\n", logTag, instituteC2AlcadnoMapName)
					return
				}
				fmt.Printf("%s %s 右走结束未换图 左瞬移%d次重试\n",
					logTag, instituteC2AlcadnoMapName, instituteC2AlcadnoRetryLeftTP)
				instituteC2AlcadnoLeftTeleportN(logTag, instituteC2AlcadnoRetryLeftTP)
			case instituteC2AlcadnoWalkTimeoutOnMap:
				*whiteWalkDone = false
				if whiteStuckCount != nil {
					*whiteStuckCount = 0
				}
				instituteC2AlcadnoLeftTeleportN(logTag, instituteC2AlcadnoRetryLeftTP)
			}
			return
		}
		if whiteStuckCount != nil {
			*whiteStuckCount++
			fmt.Printf("%s %s ffffff已处理仍卡图 %d/%d\n",
				logTag, instituteC2AlcadnoMapName, *whiteStuckCount, instituteC2AlcadnoWhiteStuckMax)
			if *whiteStuckCount >= instituteC2AlcadnoWhiteStuckMax {
				fmt.Printf("%s %s 连续%d次未换图 左瞬移%d次重试\n",
					logTag, instituteC2AlcadnoMapName, *whiteStuckCount, instituteC2AlcadnoRetryLeftTP)
				*whiteStuckCount = 0
				if whiteWalkDone != nil {
					*whiteWalkDone = false
				}
				instituteC2AlcadnoLeftTeleportN(logTag, instituteC2AlcadnoRetryLeftTP)
			}
		}
		return
	}
	if whiteStuckCount != nil {
		*whiteStuckCount = 0
	}
	fmt.Printf("%s %s ffffff区域=%d ffffff未找到(40x60≥1500) 右瞬移\n",
		logTag, instituteC2AlcadnoMapName, whiteN)
	tapTeleportWithDirection(true)
	sleepAfterTeleport()
}

func instituteC2HandlePlainMagatia(logTag, text string) {
	if !instituteC2IsPlainMagatia(text) {
		return
	}
	relX, relY, ok, detail := ReadMinimapRelWithDetail()
	if ok {
		fmt.Printf("%s 마가티아 %s\n", logTag, detail)
		if instituteC2RecoverMagatiaFromFall(logTag, relX, relY) {
			return
		}
		instituteC2AlignMagatiaX(logTag, relX)
		if relX2, relY2, ok2 := ReadMinimapRel(); ok2 {
			instituteC2HandleAtMagatia(logTag, relX2, relY2)
		} else {
			instituteC2HandleAtMagatia(logTag, relX, relY)
		}
	} else {
		fmt.Printf("%s 마가티아 小地图未识别: %s\n", logTag, detail)
	}
}
