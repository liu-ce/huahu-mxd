package play

import (
	"app/core"
	"app/job"
	"app/util"
	"fmt"
	"math/rand"
	"sync/atomic"
	"time"

	"github.com/Dasongzi1366/AutoGo/motion"
)

const landLRLogTag = "[land原地左右打]"

// land 专用点击与血蓝检测（不走 attackKeyCode / farm_maintain 默认区域）。
const (
	landAttackClickX1 = 1009
	landAttackClickY1 = 492
	landAttackClickX2 = 1034
	landAttackClickY2 = 509

	landHpPotionClickX1 = 1131
	landHpPotionClickY1 = 417
	landHpPotionClickX2 = 1160
	landHpPotionClickY2 = 443

	landMpPotionClickX1 = 1151
	landMpPotionClickY1 = 580
	landMpPotionClickX2 = 1176
	landMpPotionClickY2 = 613

	landAttackHoldMinMs    = 100
	landAttackHoldMaxMs    = 200
	landFaceHoldMinMs      = 100
	landFaceHoldMaxMs      = 150
	landFaceSettleMs       = 50
	landAttackIntervalMs   = 200
	landAttacksPerDirMin   = 3
	landAttacksPerDirMax   = 4
	landBeforeTurnGapMinMs = 500
	landBeforeTurnGapMaxMs = 600

	landPotionClickMinMs = 80
	landPotionClickMaxMs = 120

	landHpBarFullCount = 700
	landMpBarFullCount = 800
	landHpBarRegionX1  = 472
	landHpBarRegionY1  = 682
	landHpBarRegionX2  = 575
	landHpBarRegionY2  = 694
	landHpBarColor     = "ee0000"
	landHpBarSim       = float32(0.9)
	landMpBarRegionX1  = 582
	landMpBarRegionY1  = 679
	landMpBarRegionX2  = 683
	landMpBarRegionY2  = 694
	landMpBarColor     = "0088ff"
	landMpBarSim       = float32(0.8)
)

var landMaintainOn int32

func landLRLog(format string, args ...interface{}) {
	fmt.Printf(landLRLogTag+" "+format+"\n", args...)
}

func landFaceLeft() {
	releaseDpadHold(motion.KEYCODE_DPAD_RIGHT)
	keyHoldDirection(motion.KEYCODE_DPAD_LEFT, landFaceHoldMinMs, landFaceHoldMaxMs)
}

func landFaceRight() {
	releaseDpadHold(motion.KEYCODE_DPAD_LEFT)
	keyHoldDirection(motion.KEYCODE_DPAD_RIGHT, landFaceHoldMinMs, landFaceHoldMaxMs)
}

func landClickAttackHold() {
	core.RandomLongClickInArea(
		landAttackClickX1, landAttackClickY1, landAttackClickX2, landAttackClickY2,
		landAttackHoldMinMs, landAttackHoldMaxMs,
	)
}

func landAttacksPerBurst() int {
	return landAttacksPerDirMin + rand.Intn(landAttacksPerDirMax-landAttacksPerDirMin+1)
}

func landAttackBurstLeft() {
	landFaceLeft()
	core.Sleep(landFaceSettleMs)
	n := landAttacksPerBurst()
	for i := 0; i < n; i++ {
		landClickAttackHold()
		if i < n-1 {
			core.Sleep(landAttackIntervalMs)
		}
	}
	core.RandomSleep(landBeforeTurnGapMinMs, landBeforeTurnGapMaxMs)
}

func landAttackBurstRight() {
	landFaceRight()
	core.Sleep(landFaceSettleMs)
	n := landAttacksPerBurst()
	for i := 0; i < n; i++ {
		landClickAttackHold()
		if i < n-1 {
			core.Sleep(landAttackIntervalMs)
		}
	}
	core.RandomSleep(landBeforeTurnGapMinMs, landBeforeTurnGapMaxMs)
}

func landClickHpPotion() {
	core.RandomLongClickInArea(
		landHpPotionClickX1, landHpPotionClickY1, landHpPotionClickX2, landHpPotionClickY2,
		landPotionClickMinMs, landPotionClickMaxMs,
	)
}

func landClickMpPotion() {
	core.RandomLongClickInArea(
		landMpPotionClickX1, landMpPotionClickY1, landMpPotionClickX2, landMpPotionClickY2,
		landPotionClickMinMs, landPotionClickMaxMs,
	)
}

func landTickPotionIfNeeded() {
	// 自动吃血/蓝药已关闭
}

func landTryDismissTopBanner() {
	n := core.Color.GetColorCountInRegion(
		topBannerColorRegionX1, topBannerColorRegionY1,
		topBannerColorRegionX2, topBannerColorRegionY2,
		topBannerColor, topBannerColorSim,
	)
	if n <= topBannerColorMinCount {
		return
	}
	landLRLog("维护: 顶部公告色块=%d>%d 点击关闭", n, topBannerColorMinCount)
	core.RandomClickInArea(topBannerDismissX1, topBannerDismissY1, topBannerDismissX2, topBannerDismissY2)
}

func landMaintainPaused() bool {
	return farmMaintainPaused()
}

func StartLandMaintainLoop(logTag string) {
	if !atomic.CompareAndSwapInt32(&landMaintainOn, 0, 1) {
		return
	}
	go runLandMaintainLoop(logTag)
	startScheduledLogoutLoop()
}

func StopLandMaintainLoop() {
	atomic.StoreInt32(&landMaintainOn, 0)
	resetFarmRoleUpdateSchedule()
}

func runLandMaintainLoop(logTag string) {
	nextOSSUpload := time.Now()
	nextCloseShop := time.Now()
	nextTopBannerCheck := time.Now()
	nextExceptionCheck := time.Now()
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for atomic.LoadInt32(&landMaintainOn) == 1 {
		<-ticker.C
		now := time.Now()
		if !now.Before(nextExceptionCheck) {
			nextExceptionCheck = now.Add(exceptionCheckInterval)
			if !core.IsScheduledLogoutActive() && core.ExitOnGameExceptionIfNeeded != nil {
				core.ExitOnGameExceptionIfNeeded()
			}
		}
		if core.IsCaptchaHold() {
			continue
		}
		if !now.Before(nextTopBannerCheck) {
			nextTopBannerCheck = now.Add(topBannerCheckInterval)
			landTryDismissTopBanner()
		}
		if !now.Before(nextCloseShop) {
			if !job.IsAutoShopRunning() {
				job.Close商店()
			}
			nextCloseShop = now.Add(closeShopInterval)
		}
		if !now.Before(nextOSSUpload) {
			if err := util.UploadOSS(); err != nil {
				landLRLog("维护: UploadOSS 失败 %v", err)
			} else {
				landLRLog("维护: UploadOSS 完成")
			}
			nextOSSUpload = now.Add(ossUploadInterval)
		}
		if !landMaintainPaused() {
			landTickPotionIfNeeded()
		}
	}
}

// Play_land原地左右打 同向连打3～4次(间隔200ms) → 停500～600ms → 换向。
func Play_land原地左右打(mapAssetPath string) error {
	if _, err := loadMapConfig(mapAssetPath); err != nil {
		return err
	}
	SetFarmLogTag(landLRLogTag)

	StartLandMaintainLoop(landLRLogTag)
	defer StopLandMaintainLoop()

	landLRLog("开始 攻击×%d-%d 间隔%dms 换向前停%d-%dms",
		landAttacksPerDirMin, landAttacksPerDirMax, landAttackIntervalMs,
		landBeforeTurnGapMinMs, landBeforeTurnGapMaxMs)
	landLRLog("血条[%d,%d,%d,%d] %s≥%.0f%% 满=%d 50%%加血",
		landHpBarRegionX1, landHpBarRegionY1, landHpBarRegionX2, landHpBarRegionY2,
		landHpBarColor, landHpBarSim*100, landHpBarFullCount)
	landLRLog("蓝条[%d,%d,%d,%d] %s≥%.0f%% 满=%d 50%%加蓝",
		landMpBarRegionX1, landMpBarRegionY1, landMpBarRegionX2, landMpBarRegionY2,
		landMpBarColor, landMpBarSim*100, landMpBarFullCount)

	for {
		core.BlockWhileCaptchaHold()
		TickFarmRoleUpdateOnMainThread()
		landAttackBurstLeft()
		landAttackBurstRight()
	}
}
