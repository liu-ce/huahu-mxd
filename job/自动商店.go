package job

import (
	"app/core"
	"fmt"
	"math"
	"sync/atomic"
	"time"

	"github.com/Dasongzi1366/AutoGo/motion"
)

const (
	autoShopOpenMaxTry      = 8
	autoShopPosTolerance    = 2
	autoShopStableDuration  = 5 * time.Second
	autoShopPositionMaxWait = 2 * time.Minute
	autoShopWalkMsMin       = 60
	autoShopWalkMsMax       = 120

	autoShopInventoryPullUpImg      = "img/game/背包上拉.png"
	autoShopInventoryPullUpMaxClick = 3

	autoShopBuyAmount           = "300"
	autoShopBuyConfirmRounds    = 1
	autoShopBuyBluePotionRounds = 2
	autoShopBuyQtyRetryMax      = 5
	autoShopBuyConfirmColor     = "f28844"
	autoShopBuyConfirmSim       = float32(0.95)
	autoShopBuyConfirmMinPixels = 5 // >4
	autoShopBuyQtyDialogX1      = 700
	autoShopBuyQtyDialogY1      = 434
	autoShopBuyQtyDialogX2      = 746
	autoShopBuyQtyDialogY2      = 454
	autoShopBuyQtyInputX1       = 490
	autoShopBuyQtyInputY1       = 376
	autoShopBuyQtyInputX2       = 788
	autoShopBuyQtyInputY2       = 394
	autoShopBuyQtyClearKeys     = 8
	autoShopChatOKImg           = "img/game/聊天输入确认.png"
	autoShopChatOKFindX1        = 460
	autoShopChatOKFindY1        = 630
	autoShopChatOKFindX2        = 600
	autoShopChatOKFindY2        = 719
	autoShopChatOKSim           = float32(0.9)
	autoShopChatCloseMaxTry     = 2
	autoShopBuyDismissX1        = 755
	autoShopBuyDismissY1        = 425
	autoShopBuyDismissX2        = 799
	autoShopBuyDismissY2        = 439
	autoShopBuyDismissClickX1   = 764
	autoShopBuyDismissClickY1   = 429
	autoShopBuyDismissClickX2   = 789
	autoShopBuyDismissClickY2   = 438
	autoShopWhitePotionDblX1    = 494
	autoShopWhitePotionDblY1    = 426
	autoShopWhitePotionDblX2    = 529
	autoShopWhitePotionDblY2    = 443
	autoShopBluePotionDblX1     = 502
	autoShopBluePotionDblY1     = 526
	autoShopBluePotionDblX2     = 554
	autoShopBluePotionDblY2     = 551

	autoShopBuyPetSnackAmount        = "200"
	autoShopBuyPetSnackRounds        = 1
	autoShopBuyWhitePotionThreshold  = 500
	autoShopBuyBluePotionThreshold   = 2000
	autoShopBuyPetSnackThreshold     = 100
	autoShopPetSnackScrollClickX1    = 619
	autoShopPetSnackScrollClickY1    = 553
	autoShopPetSnackScrollClickX2    = 626
	autoShopPetSnackScrollClickY2    = 555
	autoShopPetSnackScrollFastClicks = 25
	autoShopPetSnackScrollSlowMax    = 40
	autoShopPetSnackPageX1           = 612
	autoShopPetSnackPageY1           = 486
	autoShopPetSnackPageX2           = 633
	autoShopPetSnackPageY2           = 524
	autoShopPetSnackPageColor        = "5588aa"
	autoShopPetSnackPageSim          = float32(0.9)
	autoShopPetSnackPageMinPixels    = 51 // >50
	autoShopPetSnackFindX1           = 329
	autoShopPetSnackFindY1           = 269
	autoShopPetSnackFindX2           = 428
	autoShopPetSnackFindY2           = 570
	autoShopPetSnackImg              = "img/game/宠物零食.png"
	autoShopPetSnackDblPad           = 10
	autoShopPetSnackFindRetries      = 8

	autoShopBuyConfigKey = "自动买药"

	worldProbeX1  = 8
	worldProbeY1  = 4
	worldProbeX2  = 447
	worldProbeY2  = 238
	worldProbeImg = "img/game/左上角.png,img/game/左上角2.png"
)

var autoShopRunning int32

func IsAutoShopRunning() bool {
	return atomic.LoadInt32(&autoShopRunning) == 1
}

// IsShopPanelOpen 随身商店界面是否打开。
func IsShopPanelOpen() bool {
	return core.Color.GetColorCountInRegion(530, 152, 620, 173, "ccdd44", 0.96) >= 50
}

func setAutoShopRunning(v bool) {
	if v {
		atomic.StoreInt32(&autoShopRunning, 1)
	} else {
		atomic.StoreInt32(&autoShopRunning, 0)
	}
}

// DO_自动商店 自动清包成功后打开商店并卖物（装备+杂货）。
// targetX/targetY 为相对小地图坐标；若任一 < 0 则首次识别到的位置作为目标点。
// 卖物前须在该点站稳且连续 5 秒内坐标无变化，否则先走过去再等待。
func DO_自动商店(targetX, targetY int) {
	DO_自动商店WithSellMisc(targetX, targetY, true)
}

// DO_自动商店WithSellMisc sellMisc 为 false 时只卖装备，不卖杂货。
func DO_自动商店WithSellMisc(targetX, targetY int, sellMisc bool) {
	setAutoShopRunning(true)
	defer setAutoShopRunning(false)

	if !ensureAutoShopPosition(targetX, targetY) {
		fmt.Println("[自动商店] 就位或站稳超时，放弃")
		return
	}

	core.KeyTap(motion.KEYCODE_I)
	core.RandomSleep(3000, 4000)

	anchorX, anchorY := core.OpenCV.FindImage(118, 55, 1248, 614, "img/game/背包下拉.png", false, 1.0, 0.7)
	if anchorX <= 0 || anchorY <= 0 {
		Close商店()
		return
	}

	fmt.Println(anchorX, anchorY)
	for i := 0; i < 3; i++ {
		num := core.Color.GetColorCountInRegion(anchorX-16, anchorY-281, anchorX+22, anchorY-255, "ff7799", 0.97)
		if num > 10 {
			core.RandomSleep(3000, 4000)
			break
		}
		if num == 0 {
			core.RandomClickInArea(anchorX-9, anchorY-275, anchorX+13, anchorY-267)
			core.RandomSleep(2000, 2500)
		}
		fmt.Println(num)
	}

	if !tryOpenPortableShop(anchorX, anchorY) {
		fmt.Println("[自动商店] 打开商店失败")
		Close商店()
		return
	}

	core.RandomSleep(1000, 1200)
	silver, _ := core.OCR.DetectNumber(817, 221, 930, 239)
	fmt.Println("silver:", silver)
	if silver >= 100000 {
		core.Role.Datas["silver"] = silver
	}

	do自动商店买药()
	do自动商店卖货(sellMisc)

	Close商店()
}

func readMinimapRel() (relX, relY int, ok bool) {
	mx, my := core.FindMinimapYellowCenter()
	if mx < 0 || my < 0 {
		return 0, 0, false
	}
	core.Sleep(core.JitterMs(14, 0.25))
	wx, wy := core.OpenCV.FindImage(
		worldProbeX1, worldProbeY1, worldProbeX2, worldProbeY2,
		worldProbeImg, false, 1.0, 0.6,
	)
	if wx < 0 || wy < 0 {
		return 0, 0, false
	}
	return mx - (wx - 50), my - wy, true
}

func autoShopNear(x1, y1, x2, y2 int) bool {
	dx := x1 - x2
	if dx < 0 {
		dx = -dx
	}
	dy := y1 - y2
	if dy < 0 {
		dy = -dy
	}
	return dx <= autoShopPosTolerance && dy <= autoShopPosTolerance
}

func autoShopWalkHold(goRight bool) {
	ms := autoShopWalkMsMin + int(time.Now().UnixNano()%int64(autoShopWalkMsMax-autoShopWalkMsMin+1))
	code := motion.KEYCODE_DPAD_LEFT
	if goRight {
		code = motion.KEYCODE_DPAD_RIGHT
	}
	motion.KeyActionDown(code, 0)
	core.Sleep(ms)
	motion.KeyActionUp(code, 0)
}

func autoShopWalkVertical(goDown bool) {
	ms := autoShopWalkMsMin + int(time.Now().UnixNano()%int64(autoShopWalkMsMax-autoShopWalkMsMin+1))
	code := motion.KEYCODE_DPAD_UP
	if goDown {
		code = motion.KEYCODE_DPAD_DOWN
	}
	motion.KeyActionDown(code, 0)
	core.Sleep(ms)
	motion.KeyActionUp(code, 0)
}

func walkTowardAutoShop(relX, relY, targetX, targetY int) {
	if relX < targetX-autoShopPosTolerance {
		fmt.Printf("[自动商店] 走向卖物点 右走 relX=%d→%d\n", relX, targetX)
		autoShopWalkHold(true)
		return
	}
	if relX > targetX+autoShopPosTolerance {
		fmt.Printf("[自动商店] 走向卖物点 左走 relX=%d→%d\n", relX, targetX)
		autoShopWalkHold(false)
		return
	}
	if relY < targetY-autoShopPosTolerance {
		fmt.Printf("[自动商店] 走向卖物点 下走 relY=%d→%d\n", relY, targetY)
		autoShopWalkVertical(true)
		return
	}
	if relY > targetY+autoShopPosTolerance {
		fmt.Printf("[自动商店] 走向卖物点 上走 relY=%d→%d\n", relY, targetY)
		autoShopWalkVertical(false)
	}
}

func ensureAutoShopPosition(targetX, targetY int) bool {
	if targetX < 0 || targetY < 0 {
		x, y, ok := readMinimapRel()
		if !ok {
			fmt.Println("[自动商店] 小地图未识别，无法确定卖物点")
			return false
		}
		targetX, targetY = x, y
		fmt.Printf("[自动商店] 以当前点为卖物点 x=%d y=%d\n", targetX, targetY)
	}

	deadline := time.Now().Add(autoShopPositionMaxWait)
	var stableStart time.Time
	var stableAnchorX, stableAnchorY int
	hasStableAnchor := false

	for time.Now().Before(deadline) {
		relX, relY, ok := readMinimapRel()
		if !ok {
			hasStableAnchor = false
			core.Sleep(200)
			continue
		}

		if !autoShopNear(relX, relY, targetX, targetY) {
			hasStableAnchor = false
			walkTowardAutoShop(relX, relY, targetX, targetY)
			core.Sleep(1000)
			continue
		}

		if !hasStableAnchor {
			stableAnchorX, stableAnchorY = relX, relY
			stableStart = time.Now()
			hasStableAnchor = true
			fmt.Printf("[自动商店] 已到卖物点 x=%d y=%d，开始计时站稳\n", relX, relY)
			core.Sleep(1000)
			continue
		}

		if !autoShopNear(relX, relY, stableAnchorX, stableAnchorY) {
			hasStableAnchor = false
			fmt.Printf("[自动商店] 坐标变化 (%d,%d)→(%d,%d)，重新站稳\n",
				stableAnchorX, stableAnchorY, relX, relY)
			core.Sleep(1000)
			continue
		}

		if time.Since(stableStart) >= autoShopStableDuration {
			fmt.Printf("[自动商店] 已站稳 %v x=%d y=%d，开始卖物\n",
				autoShopStableDuration, relX, relY)
			return true
		}

		core.Sleep(1000)
	}

	fmt.Printf("[自动商店] 超时未在 (%d,%d) 站稳 %v\n", targetX, targetY, autoShopStableDuration)
	return false
}

func tryOpenPortableShop(anchorX, anchorY int) bool {
	fmt.Println()
	autoShopPullUpInventory(anchorX, anchorY)
	for attempt := 1; attempt <= autoShopOpenMaxTry; attempt++ {
		shopX, shopY := core.OpenCV.FindImage(anchorX-205, anchorY-318, anchorX+33, anchorY+70, "img/game/随身商店.png,img/game/随身商店2.png", false, 1.0, 0.7)
		if shopX <= 0 || shopY <= 0 {
			fmt.Printf("[自动商店] 第%d次 未找到随身商店\n", attempt)
			core.RandomSleep(1000, 1200)
			continue
		}

		core.Click(shopX+5, shopY+5)
		core.RandomSleep(100, 200)
		core.Click(shopX+5, shopY+5)

		for i := 0; i < 10; i++ {
			core.RandomSleep(900, 1200)
			num := core.Color.GetColorCountInRegion(530, 152, 620, 173, "ccdd44", 0.96)
			if num > 50 {
				fmt.Printf("[自动商店] 第%d次 商店已打开\n", attempt)
				return true
			}
		}
		fmt.Printf("[自动商店] 第%d次 双击后未检测到商店界面\n", attempt)
	}
	return false
}

func autoShopPullUpInventory(anchorX, anchorY int) {
	for i := 1; i <= autoShopInventoryPullUpMaxClick; i++ {
		if !autoShopClickInventoryPullUp(anchorX, anchorY, i) {
			return
		}
	}
}

func autoShopClickInventoryPullUp(anchorX, anchorY, clickIndex int) bool {
	pullX, pullY := core.OpenCV.FindImage(anchorX-60, anchorY-360, anchorX+60, anchorY-180, autoShopInventoryPullUpImg, false, 1.0, 0.7)
	if pullX <= 0 || pullY <= 0 {
		pullX, pullY = core.OpenCV.FindImage(118, 55, 1248, 614, autoShopInventoryPullUpImg, false, 1.0, 0.7)
	}
	if pullX <= 0 || pullY <= 0 {
		fmt.Printf("[自动商店] 未找到背包上拉按钮，跳过上拉 第%d次\n", clickIndex)
		return false
	}
	fmt.Printf("[自动商店] 尝试背包上拉 第%d次 x=%d y=%d\n", clickIndex, pullX, pullY)
	core.RandomClickInArea(pullX, pullY, pullX+5, pullY+5)
	core.RandomSleep(500, 700)
	return true
}

func do自动商店卖货(sellMisc bool) {
	core.RandomSleep(2000, 2500)
	卖装备()
	if !sellMisc {
		fmt.Println("[自动商店] 配置只卖装备 跳过杂货")
		return
	}
	core.RandomSleep(1000, 1500)
	卖杂货()
}

func do自动商店买药() {
	if !autoShopBuyEnabled() {
		fmt.Println("[自动商店] 自动买药未启用 跳过")
		return
	}
	core.RefreshFarmItemCounts()
	counts := core.GetFarmItemCountsSnapshot()
	fmt.Printf("[自动商店] 开始买药 道具 %s\n", formatAutoShopItemCounts(counts))
	if autoShopBuyWhitePotionEnabled() {
		if autoShopShouldBuyByCount("白血瓶", counts.HP, counts.HPOK, autoShopBuyWhitePotionThreshold) {
			autoShopBuyPotion("白血瓶",
				autoShopWhitePotionDblX1, autoShopWhitePotionDblY1, autoShopWhitePotionDblX2, autoShopWhitePotionDblY2,
				false)
		}
	} else {
		fmt.Println("[自动商店] 配置关闭白血瓶 跳过")
	}
	if autoShopBuyBluePotionEnabled() {
		if autoShopShouldBuyByCount("蓝瓶", counts.MP, counts.MPOK, autoShopBuyBluePotionThreshold) {
			autoShopBuyItem("蓝瓶", autoShopBuyAmount, autoShopBuyBluePotionRounds,
				autoShopBluePotionDblX1, autoShopBluePotionDblY1, autoShopBluePotionDblX2, autoShopBluePotionDblY2,
				true)
		}
	} else {
		fmt.Println("[自动商店] 配置关闭蓝血瓶 跳过")
	}
	if autoShopBuyPetSnackEnabled() {
		if autoShopShouldBuyByCount("宠物零食", counts.Pet, counts.PetOK, autoShopBuyPetSnackThreshold) {
			autoShopBuyPetSnack()
		}
	} else {
		fmt.Println("[自动商店] 配置关闭宠物零食 跳过")
	}
	fmt.Println("[自动商店] 买药完成")
}

func formatAutoShopItemCounts(c core.FarmItemCountsSnapshot) string {
	return fmt.Sprintf("血瓶=%s 蓝瓶=%s 宠物零食=%s 混沌卷轴=%s",
		formatAutoShopItemCount(c.HP, c.HPOK),
		formatAutoShopItemCount(c.MP, c.MPOK),
		formatAutoShopItemCount(c.Pet, c.PetOK),
		formatAutoShopItemCount(c.Scroll, c.ScrollOK),
	)
}

func formatAutoShopItemCount(n int, ok bool) string {
	if ok {
		return fmt.Sprintf("%d", n)
	}
	return "?"
}

func autoShopShouldBuyByCount(name string, count int, ok bool, threshold int) bool {
	if !ok {
		fmt.Printf("[自动商店] %s 数量未知 跳过\n", name)
		return false
	}
	if count >= threshold {
		fmt.Printf("[自动商店] %s 当前%d >= %d 无需购买 跳过\n", name, count, threshold)
		return false
	}
	fmt.Printf("[自动商店] %s 当前%d < %d 需要购买\n", name, count, threshold)
	return true
}

func autoShopBuyEnabled() bool {
	return core.API.GetConfigBoolValueOrDefault(autoShopBuyConfigKey+".启用", true)
}

func autoShopBuyWhitePotionEnabled() bool {
	return core.API.GetConfigBoolValueOrDefault(autoShopBuyConfigKey+".白血瓶", true)
}

func autoShopBuyBluePotionEnabled() bool {
	return core.API.GetConfigBoolValueOrDefault(autoShopBuyConfigKey+".蓝血瓶", true)
}

func autoShopBuyPetSnackEnabled() bool {
	return core.API.GetConfigBoolValueOrDefault(autoShopBuyConfigKey+".宠物零食", true)
}

func autoShopBuyPotion(name string, dblX1, dblY1, dblX2, dblY2 int, waitBeforeDbl bool) {
	autoShopBuyItem(name, autoShopBuyAmount, autoShopBuyConfirmRounds, dblX1, dblY1, dblX2, dblY2, waitBeforeDbl)
}

func autoShopBuyPetSnack() {
	fmt.Println("[自动商店] 开始买宠物零食")
	if !autoShopScrollToPetSnackPage() {
		fmt.Println("[自动商店] 滚动到宠物零食页失败 跳过")
		return
	}
	core.RandomSleepAround(1200, 0.15)
	snackX, snackY := autoShopFindPetSnackIcon()
	if snackX <= 0 || snackY <= 0 {
		fmt.Println("[自动商店] 未找到宠物零食图标 跳过")
		return
	}
	dblX1 := snackX
	dblY1 := snackY
	dblX2 := snackX + autoShopPetSnackDblPad
	dblY2 := snackY + autoShopPetSnackDblPad
	fmt.Printf("[自动商店] 找到宠物零食 x=%d y=%d\n", snackX, snackY)
	autoShopBuyItem("宠物零食", autoShopBuyPetSnackAmount, autoShopBuyPetSnackRounds, dblX1, dblY1, dblX2, dblY2, true)
}

func autoShopFindPetSnackIcon() (int, int) {
	sims := []float32{0.7, 0.65, 0.6}
	for attempt := 1; attempt <= autoShopPetSnackFindRetries; attempt++ {
		for _, sim := range sims {
			x, y := core.OpenCV.FindImage(
				autoShopPetSnackFindX1, autoShopPetSnackFindY1,
				autoShopPetSnackFindX2, autoShopPetSnackFindY2,
				autoShopPetSnackImg, false, 1.0, sim,
			)
			if x > 0 && y > 0 {
				if attempt > 1 {
					fmt.Printf("[自动商店] 第%d次识别到宠物零食图标 x=%d y=%d\n", attempt, x, y)
				}
				return x, y
			}
		}
		if attempt < autoShopPetSnackFindRetries {
			fmt.Printf("[自动商店] 第%d次未找到宠物零食图标 等待后重试\n", attempt)
			core.RandomSleep(500, 800)
		}
	}
	return -1, -1
}

func autoShopScrollToPetSnackPage() bool {
	if autoShopPetSnackPageVisible() {
		fmt.Println("[自动商店] 已在宠物零食页")
		return true
	}
	for i := 1; i <= autoShopPetSnackScrollFastClicks; i++ {
		core.RandomClickInArea(
			autoShopPetSnackScrollClickX1, autoShopPetSnackScrollClickY1,
			autoShopPetSnackScrollClickX2, autoShopPetSnackScrollClickY2,
		)
		if autoShopPetSnackPageVisible() {
			fmt.Printf("[自动商店] 滚动到宠物零食页 快速点击第%d次\n", i)
			return true
		}
		core.RandomSleep(200, 300)
	}
	for i := 1; i <= autoShopPetSnackScrollSlowMax; i++ {
		core.RandomClickInArea(
			autoShopPetSnackScrollClickX1, autoShopPetSnackScrollClickY1,
			autoShopPetSnackScrollClickX2, autoShopPetSnackScrollClickY2,
		)
		if autoShopPetSnackPageVisible() {
			fmt.Printf("[自动商店] 滚动到宠物零食页 慢速点击第%d次\n", i)
			return true
		}
		core.RandomSleep(800, 1200)
	}
	return false
}

func autoShopPetSnackPageVisible() bool {
	return core.Color.GetColorCountInRegion(
		autoShopPetSnackPageX1, autoShopPetSnackPageY1,
		autoShopPetSnackPageX2, autoShopPetSnackPageY2,
		autoShopPetSnackPageColor, autoShopPetSnackPageSim,
	) >= autoShopPetSnackPageMinPixels
}

func autoShopBuyItem(name, amount string, rounds, dblX1, dblY1, dblX2, dblY2 int, waitBeforeDbl bool) {
	if waitBeforeDbl {
		core.RandomSleepAround(1000, 0.15)
	}
	autoShopMaybeDismissBuyDialog()
	autoShopDoubleClickArea(dblX1, dblY1, dblX2, dblY2)
	for round := 1; round <= rounds; round++ {
		autoShopMaybeDismissBuyDialog()
		if !autoShopWaitQtyDialog(name, round, dblX1, dblY1, dblX2, dblY2) {
			continue
		}
		fmt.Printf("[自动商店] 买%s 第%d轮 输入%s 确认\n", name, round, amount)
		autoShopFocusBuyQtyInput()
		autoShopInputDigits(amount)
		core.RandomSleepAround(1000, 0.15)
		motion.KeyAction(motion.KEYCODE_ENTER, 0)
		core.Sleep(2000)
	}
}

func autoShopWaitQtyDialog(name string, round, dblX1, dblY1, dblX2, dblY2 int) bool {
	core.RandomSleepAround(1500, 0.15)
	if autoShopBuyQtyDialogVisible() {
		return true
	}
	for retry := 1; retry <= autoShopBuyQtyRetryMax; retry++ {
		fmt.Printf("[自动商店] 买%s 第%d轮 未检测到数量框 重试双击 %d/%d\n", name, round, retry, autoShopBuyQtyRetryMax)
		autoShopDoubleClickArea(dblX1, dblY1, dblX2, dblY2)
		core.RandomSleepAround(1500, 0.15)
		if autoShopBuyQtyDialogVisible() {
			return true
		}
	}
	fmt.Printf("[自动商店] 买%s 第%d轮 重试%d次后仍未检测到数量框 跳过\n", name, round, autoShopBuyQtyRetryMax)
	return false
}

func autoShopBuyQtyDialogVisible() bool {
	return core.Color.GetColorCountInRegion(
		autoShopBuyQtyDialogX1, autoShopBuyQtyDialogY1,
		autoShopBuyQtyDialogX2, autoShopBuyQtyDialogY2,
		autoShopBuyConfirmColor, autoShopBuyConfirmSim,
	) >= autoShopBuyConfirmMinPixels
}

func autoShopMaybeDismissBuyDialog() {
	n := core.Color.GetColorCountInRegion(
		autoShopBuyDismissX1, autoShopBuyDismissY1,
		autoShopBuyDismissX2, autoShopBuyDismissY2,
		autoShopBuyConfirmColor, autoShopBuyConfirmSim,
	)
	if n >= autoShopBuyConfirmMinPixels {
		fmt.Println("[自动商店] 检测到买药提示 点击关闭")
		core.RandomClickInArea(
			autoShopBuyDismissClickX1, autoShopBuyDismissClickY1,
			autoShopBuyDismissClickX2, autoShopBuyDismissClickY2,
		)
		core.RandomSleepAround(1000, 0.15)
	}
}

func autoShopFocusBuyQtyInput() {
	core.RandomClickInArea(
		autoShopBuyQtyInputX1, autoShopBuyQtyInputY1,
		autoShopBuyQtyInputX2, autoShopBuyQtyInputY2,
	)
	core.RandomSleep(150, 250)
	motion.KeyAction(motion.KEYCODE_MOVE_END, 0)
	core.RandomSleep(80, 120)
	for i := 0; i < autoShopBuyQtyClearKeys; i++ {
		motion.KeyAction(motion.KEYCODE_DEL, 0)
		core.RandomSleep(35, 60)
	}
}

func autoShopDoubleClickArea(x1, y1, x2, y2 int) {
	x, y := core.RandomClickInArea(x1, y1, x2, y2)
	core.RandomSleep(140, 200)
	core.Click(x, y)
}

func autoShopInputDigits(s string) {
	for _, r := range s {
		if r < '0' || r > '9' {
			continue
		}
		motion.KeyAction(int(r-'0')+7, 0) // Android KEYCODE_0=7
		core.RandomSleep(80, 120)
	}
}

func 卖装备() {
	core.RandomClickInArea(662, 271, 682, 286)
	core.RandomSleep(1500, 2000)

	core.RandomClickInArea(852, 159, 905, 169)
	core.RandomSleep(2000, 3000)
	motion.KeyAction(motion.KEYCODE_ENTER, 0) // 单击回车键
	core.RandomSleep(2000, 3000)
}

// sellMiscItemSlotRGB 道具格区域内左/中/右三点的 RRGGBB。
type sellMiscItemSlotRGB struct {
	Left  string
	Mid   string
	Right string
}

// readSellMiscItemSlotRGB 读取卖杂货道具格（649,484,690,564）左/中/右三点颜色。
func readSellMiscItemSlotRGB() sellMiscItemSlotRGB {
	const x1, y1, x2, y2 = 649, 484, 690, 564
	cx := (x1 + x2) / 2
	cy := (y1 + y2) / 2
	return sellMiscItemSlotRGB{
		Left:  core.Color.Pixel(x1+4, cy),
		Mid:   core.Color.Pixel(cx, cy),
		Right: core.Color.Pixel(x2-4, cy),
	}
}

func countSellMiscItemSlotPrimaryColors() (red, yellow, blue int) {
	const x1, y1, x2, y2 = 649, 484, 690, 564
	const sim = float32(0.3)
	red = core.Color.GetColorCountInRegion(x1, y1, x2, y2, "ff0000", sim)
	yellow = core.Color.GetColorCountInRegion(x1, y1, x2, y2, "ffff00", sim)
	blue = core.Color.GetColorCountInRegion(x1, y1, x2, y2, "0000ff", sim)
	return red, yellow, blue
}

type sellMiscSlotSnapshot struct {
	rgb               sellMiscItemSlotRGB
	red, yellow, blue int
}

func captureSellMiscItemSlotSnapshot() sellMiscSlotSnapshot {
	rgb := readSellMiscItemSlotRGB()
	r, y, b := countSellMiscItemSlotPrimaryColors()
	return sellMiscSlotSnapshot{rgb: rgb, red: r, yellow: y, blue: b}
}

func intChangeOverPct(oldVal, newVal int, threshold float64) bool {
	if oldVal == 0 {
		return newVal != 0
	}
	return math.Abs(float64(newVal-oldVal))/float64(oldVal) > threshold
}

func hexChannelChangeOverPct(oldHex, newHex string, threshold float64) bool {
	if oldHex == newHex {
		return false
	}
	if len(oldHex) != 6 || len(newHex) != 6 {
		return true
	}
	or, og, ob := core.Color.HexToRGB(oldHex)
	nr, ng, nb := core.Color.HexToRGB(newHex)
	return intChangeOverPct(or, nr, threshold) ||
		intChangeOverPct(og, ng, threshold) ||
		intChangeOverPct(ob, nb, threshold)
}

// sellMiscSlotChangedEnough 相对基线，RGB 三点或红/黄/蓝计数任一变化超过 30%。
func sellMiscSlotChangedEnough(base, cur sellMiscSlotSnapshot) bool {
	const threshold = 0.10
	if intChangeOverPct(base.red, cur.red, threshold) ||
		intChangeOverPct(base.yellow, cur.yellow, threshold) ||
		intChangeOverPct(base.blue, cur.blue, threshold) {
		return true
	}
	return hexChannelChangeOverPct(base.rgb.Left, cur.rgb.Left, threshold) ||
		hexChannelChangeOverPct(base.rgb.Mid, cur.rgb.Mid, threshold) ||
		hexChannelChangeOverPct(base.rgb.Right, cur.rgb.Right, threshold)
}

func 卖杂货() {

	fmt.Println("开始卖杂货")

	for i := 0; i < 100; i++ {
		num := core.Color.GetColorCountInRegion(530, 152, 620, 173, "ccdd44", 0.96)
		if num < 50 {
			fmt.Println("说明已经关闭商人了")
			return
		}

		if core.Color.GetColorCountInRegion(829, 267, 870, 286, "ff8899", 0.9) < 5 {
			core.RandomClickInArea(829, 267, 870, 286)
			core.RandomSleep(1000, 2000)
		}

		if core.Color.GetColorCountInRegion(651, 308, 685, 345, "dddddd", 0.95) > 400 {
			core.SLS_Log2("道具已清空")
			return
		}

		baseSlot := captureSellMiscItemSlotSnapshot()
		fmt.Printf("[卖杂货] 道具格 RGB(%s,%s,%s) 红=%d 黄=%d 蓝=%d\n",
			baseSlot.rgb.Left, baseSlot.rgb.Mid, baseSlot.rgb.Right,
			baseSlot.red, baseSlot.yellow, baseSlot.blue)

		// 双击道具
		for d := 0; d < 2; d++ {
			core.RandomClickInArea(843, 316, 874, 339)
			core.RandomSleep(200, 300)
		}

		core.RandomSleep(800, 1000)

		// 点击确定
		if core.Color.GetColorCountInRegion(707, 436, 742, 449, "ff9944", 0.95) >= 20 {
			motion.KeyAction(motion.KEYCODE_ENTER, 0) // 单击回车键
			core.RandomSleep(300, 400)
		}

		// 点击确定
		if core.Color.GetColorCountInRegion(702, 426, 739, 438, "ee8822", 0.95) >= 4 {
			motion.KeyAction(motion.KEYCODE_ENTER, 0) // 单击回车键
			core.RandomSleep(300, 400)
		}

		// 点击确定
		if core.Color.GetColorCountInRegion(758, 423, 799, 439, "ee8822", 0.95) >= 5 {
			motion.KeyAction(motion.KEYCODE_ENTER, 0) // 单击回车键
			core.RandomSleep(300, 400)
		}

		// 点击取消
		if core.Color.GetColorCountInRegion(760, 437, 796, 450, "ee8822", 0.95) >= 5 {
			core.RandomClickInArea(767, 424, 791, 435)
			core.RandomSleep(300, 400)
		}

		sold := false
		for poll := 0; poll < 6; poll++ {
			core.RandomSleep(300, 500)
			curSlot := captureSellMiscItemSlotSnapshot()
			if sellMiscSlotChangedEnough(baseSlot, curSlot) {
				fmt.Printf("[卖杂货] 道具格变化超30%% poll=%d 下一格\n", poll+1)
				sold = true
				core.RandomSleep(300, 400)
				break
			}
		}
		if sold {
			continue
		}

	}
}

func Close商店() {
	for i := 0; i < 50; i++ {
		num := core.Color.GetColorCountInRegion(530, 152, 620, 173, "ccdd44", 0.96)
		if num >= 50 {
			// 离开商店
			core.RandomClickInArea(549, 158, 599, 166)
			core.RandomSleep(2000, 3000)
			continue
		}

		// 点击取消
		if core.Color.GetColorCountInRegion(760, 437, 796, 450, "ee8822", 0.95) >= 5 {
			core.RandomClickInArea(767, 424, 791, 435)
			core.RandomSleep(1500, 2000)
			continue
		}

		// 点击确定
		if core.Color.GetColorCountInRegion(707, 436, 742, 449, "ff9944", 0.95) >= 20 {
			core.RandomClickInArea(707, 436, 742, 449)
			core.RandomSleep(1500, 2000)
			continue
		}

		// 点击确定
		if core.Color.GetColorCountInRegion(702, 426, 739, 438, "ee8822", 0.95) >= 4 {
			core.RandomClickInArea(702, 426, 739, 438)
			core.RandomSleep(1500, 2000)
			continue
		}

		// 点击确定
		if core.Color.GetColorCountInRegion(758, 423, 799, 439, "ee8822", 0.95) >= 5 {
			core.RandomClickInArea(767, 424, 791, 435)
			core.RandomSleep(1500, 2000)
			continue
		}

		anchorX, anchorY := core.OpenCV.FindImage(118, 55, 1248, 614, "img/game/背包下拉.png", false, 1.0, 0.7)
		if anchorX > 0 || anchorY > 0 {
			fmt.Println("关一次")
			core.KeyTap(motion.KEYCODE_I)
			core.RandomSleep(1500, 2000)
			continue
		}

		break
	}
	autoShopCloseChatIfOpen()
}

// autoShopCloseChatIfOpen 避免确认弹框时误按回车留下聊天输入框。
// 商店打开时，聊天收起把手会被商店界面覆盖，只能由回车切换聊天状态。
func autoShopCloseChatIfOpen() {
	for i := 1; i <= autoShopChatCloseMaxTry; i++ {
		chatOKX, chatOKY := core.OpenCV.FindImage(
			autoShopChatOKFindX1, autoShopChatOKFindY1,
			autoShopChatOKFindX2, autoShopChatOKFindY2,
			autoShopChatOKImg, false, 1.0, autoShopChatOKSim,
		)
		if chatOKX <= 0 || chatOKY <= 0 {
			return
		}

		fmt.Printf("[自动商店] 检测到聊天输入框，按回车收起 %d/%d\\n", i, autoShopChatCloseMaxTry)
		motion.KeyAction(motion.KEYCODE_ENTER, 0)
		core.RandomSleep(500, 800)
	}

	fmt.Println("[自动商店] 聊天输入框仍未关闭")
}
