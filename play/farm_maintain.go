package play

import (
	"app/core"
	"app/job"
	"app/util"
	"strconv"
	"sync/atomic"
	"time"
)

const (
	hpBarFullCount    = 786
	mpBarFullCount    = 513
	autoHealConfigKey = "自动回血"
	hpBarRegionX1     = 289
	hpBarRegionY1     = 693
	hpBarRegionX2     = 433
	hpBarRegionY2     = 717
	mpBarRegionX1     = 433
	mpBarRegionY1     = 688
	mpBarRegionX2     = 577
	mpBarRegionY2     = 717

	// 血/蓝药点击区域（不用键盘，区域内随机点）
	hpPotionClickX1 = 940
	hpPotionClickY1 = 456
	hpPotionClickX2 = 968
	hpPotionClickY2 = 475
	mpPotionClickX1 = 1007
	mpPotionClickY1 = 451
	mpPotionClickX2 = 1036
	mpPotionClickY2 = 477

	potionLongClickMinMs = 80
	potionLongClickMaxMs = 120

	ossUploadInterval      = 120 * time.Minute
	closeShopInterval      = 1 * time.Minute
	farmItemOCRInterval    = 120 * time.Second
	topBannerCheckInterval = 30 * time.Second

	topBannerColorRegionX1 = 1080
	topBannerColorRegionY1 = 12
	topBannerColorRegionX2 = 1190
	topBannerColorRegionY2 = 75
	topBannerColor         = "1caebc"
	topBannerColorSim      = float32(0.95)
	topBannerColorMinCount = 6000
	topBannerDismissX1     = 1228
	topBannerDismissY1     = 41
	topBannerDismissX2     = 1237
	topBannerDismissY2     = 52
	exceptionCheckInterval = 10 * time.Second
)

type farmItemOCRRegion struct {
	name           string
	x1, y1, x2, y2 int
}

var farmItemOCRRegions = []farmItemOCRRegion{
	{"血瓶", 971, 334, 1019, 351},
	{"蓝瓶", 1038, 331, 1084, 353},
	{"宠物零食", 1104, 332, 1149, 354},
	{"混沌卷轴", 1228, 648, 1259, 666},
}

var farmMaintainOn int32
var farmMaintainPotionOn int32 // 默认关闭自动吃血/蓝药

// SetFarmMaintainPotionEnabled 曾用于挂机图开关自动吃药；现已全局关闭，调用无效。
func SetFarmMaintainPotionEnabled(on bool) {
	_ = on
	atomic.StoreInt32(&farmMaintainPotionOn, 0)
}

func farmMaintainPotionEnabled() bool {
	return false
}

// autoHealEnabled 仅当配置显式为 false 时关闭；缺省或 true 均自动回血。
func autoHealEnabled() bool {
	return core.API.GetConfigBoolValueOrDefault(autoHealConfigKey, true)
}

func formatOCRCount(n int, ok bool) string {
	if ok {
		return strconv.Itoa(n)
	}
	return "?"
}

func RunFarmItemCountsOCR() {
	runFarmItemCountsOCR()
}

func runFarmItemCountsOCR() {
	var hp, mp, pet, scroll int
	var hpOK, mpOK, petOK, scrollOK bool
	for i, r := range farmItemOCRRegions {
		raw, ok := core.OCR.DetectNumber(r.x1, r.y1, r.x2, r.y2)
		n, useOK := pushFarmItemCountFilter(i, raw, ok)
		switch i {
		case 0:
			hp, hpOK = n, useOK
		case 1:
			mp, mpOK = n, useOK
		case 2:
			pet, petOK = n, useOK
		case 3:
			scroll, scrollOK = n, useOK
		}
	}
	storeFarmItemCountsSnap(hp, mp, pet, scroll, hpOK, mpOK, petOK, scrollOK)
	farmLog("维护: 道具 %s", formatFarmItemCountsForLog(hp, mp, pet, scroll, hpOK, mpOK, petOK, scrollOK))
}

// StartFarmMaintainLoop 挂机期间：定时读人物数据、自动按键、HP/MP 维护。
func StartFarmMaintainLoop(logTag string) {
	if !atomic.CompareAndSwapInt32(&farmMaintainOn, 0, 1) {
		return
	}
	ReloadAutoKeysFromConfig()
	fireTimedAutoKeysOnceAtStart()
	go runFarmMaintainLoop(logTag)
	startScheduledLogoutLoop()
}

func StopFarmMaintainLoop() {
	atomic.StoreInt32(&farmMaintainOn, 0)
	resetFarmRoleUpdateSchedule()
}

func RunFarmMaintainLoop() {
	runFarmMaintainLoop("测试")
}

func runFarmMaintainLoop(logTag string) {
	nextOSSUpload := time.Now()
	nextCloseShop := time.Now()
	nextTopBannerCheck := time.Now()
	nextExceptionCheck := time.Now()
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for atomic.LoadInt32(&farmMaintainOn) == 1 {
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
			tryDismissTopBanner()
		}
		if !now.Before(nextCloseShop) {
			if !job.IsAutoShopRunning() {
				job.Close商店()
			}
			nextCloseShop = now.Add(closeShopInterval)
		}
		if !now.Before(nextOSSUpload) {
			if err := util.UploadOSS(); err != nil {
				farmLog("维护: UploadOSS 失败 %v", err)
			} else {
				farmLog("维护: UploadOSS 完成")
			}
			nextOSSUpload = now.Add(ossUploadInterval)
		}
		hp := core.Color.GetColorCountInRegion(hpBarRegionX1, hpBarRegionY1, hpBarRegionX2, hpBarRegionY2, "ef0000", 0.9)
		mp := core.Color.GetColorCountInRegion(mpBarRegionX1, mpBarRegionY1, mpBarRegionX2, mpBarRegionY2, "0081f2", 0.9)
		if !farmMaintainPaused() {
			if farmMaintainPotionEnabled() {
				if mp < mpBarFullCount*40/100 {
					farmLog("维护: MP偏低(%d) 点蓝药", mp)
					clickMpPotion()
				}
				if autoHealEnabled() {
					if hp < hpBarFullCount/2 {
						farmLog("维护: HP偏低(%d) 点血药", hp)
						clickHpPotion()
					}
				}
			}
		}
		// 定时自动按键单独判断：不能跟 IsCaptchaUIPresent 绑在一起，UI 改版后灰块误判会永久卡住不再触发。
		if !farmTimedAutoKeysPaused() {
			tickTimedAutoKeys(now)
		}
	}
}

func farmMaintainPaused() bool {
	if core.IsScheduledLogoutActive() {
		return true
	}
	if core.IsCaptchaHold() || core.IsCaptchaUIPresent() || core.IsCaptchaWSActive() {
		return true
	}
	if job.IsAutoShopRunning() || job.IsShopPanelOpen() {
		return true
	}
	return false
}

// farmTimedAutoKeysPaused 定时自动按键暂停条件（比维护暂停更窄，避免 UI 色块误判卡死）。
func farmTimedAutoKeysPaused() bool {
	if core.IsScheduledLogoutActive() {
		return true
	}
	if core.IsCaptchaHold() || core.IsCaptchaWSActive() {
		return true
	}
	if job.IsAutoShopRunning() {
		return true
	}
	return false
}

func clickHpPotion() {
	core.RandomLongClickInArea(hpPotionClickX1, hpPotionClickY1, hpPotionClickX2, hpPotionClickY2,
		potionLongClickMinMs, potionLongClickMaxMs)
}

func clickMpPotion() {
	core.RandomLongClickInArea(mpPotionClickX1, mpPotionClickY1, mpPotionClickX2, mpPotionClickY2,
		potionLongClickMinMs, potionLongClickMaxMs)
}

func tryDismissTopBanner() {
	n := core.Color.GetColorCountInRegion(
		topBannerColorRegionX1, topBannerColorRegionY1,
		topBannerColorRegionX2, topBannerColorRegionY2,
		topBannerColor, topBannerColorSim,
	)
	if n <= topBannerColorMinCount {
		return
	}
	farmLog("维护: 顶部公告色块=%d>%d 点击关闭", n, topBannerColorMinCount)
	core.RandomClickInArea(topBannerDismissX1, topBannerDismissY1, topBannerDismissX2, topBannerDismissY2)
}

// TestClickHpPotion 供 main 调试血药点击区域。
func TestClickHpPotion() { clickHpPotion() }

// TestClickMpPotion 供 main 调试蓝药点击区域。
func TestClickMpPotion() { clickMpPotion() }
