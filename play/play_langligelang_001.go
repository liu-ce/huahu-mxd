package play

import (
	"app/core"
	"fmt"
	"math/rand"

	"github.com/Dasongzi1366/AutoGo/motion"
)

const (
	langligelang001LogTag      = "[浪里个浪001]"
	llgDoubleTeleportChancePct = 10 // 连2次瞬移概率
)

func llgLog(format string, args ...interface{}) {
	fmt.Printf(langligelang001LogTag+" "+format+"\n", args...)
}

func llgIsLowerLayer(c *Langligelang001Config, relY int) bool {
	return matchRange(relY, c.LowerYMin, c.LowerYMax)
}

func llgIsUpperLayer(c *Langligelang001Config, relY int) bool {
	return matchRange(relY, c.UpperYMin, c.UpperYMax)
}

func llgPatrolMargin(c *Langligelang001Config) int {
	if c == nil || c.PatrolTeleportMargin <= 0 {
		return 12
	}
	return c.PatrolTeleportMargin
}

func llgDoTeleport(c *Langligelang001Config, goRight bool) {
	dirCode := motion.KEYCODE_DPAD_LEFT
	if goRight {
		faceRight()
		dirCode = motion.KEYCODE_DPAD_RIGHT
	} else {
		faceLeft()
	}
	keyHoldDirectionActionMs(dirCode, teleportKeyCode(), c.TeleportHoldMinMs, c.TeleportHoldMaxMs)
	core.RandomSleep(c.AfterTeleportWaitMinMs, c.AfterTeleportWaitMaxMs)
}

func llgTeleportAndAttack(c *Langligelang001Config, goRight bool) {
	teleports := 1
	if rand.Intn(100) < llgDoubleTeleportChancePct {
		teleports = 2
		llgLog("连2次瞬移")
	}
	for i := 0; i < teleports; i++ {
		llgDoTeleport(c, goRight)
	}
	keyHoldPress(attackKeyCode(), c.AttackHoldMinMs, c.AttackHoldMaxMs)
}

// llgPatrolFarmStep 下层 x 区间两头来回瞬移+攻击。
func llgPatrolFarmStep(c *Langligelang001Config, relX int, goRight *bool) {
	xMin, xMax := c.LowerXMin, c.LowerXMax
	margin := yeqiuPatrolEffectiveMargin(xMin, xMax, llgPatrolMargin(c))

	if relX < xMin {
		*goRight = true
		llgLog("x=%d 超出左界%d 右瞬移回区", relX, xMin)
		llgTeleportAndAttack(c, true)
		return
	}
	if relX > xMax {
		*goRight = false
		llgLog("x=%d 超出右界%d 左瞬移回区", relX, xMax)
		llgTeleportAndAttack(c, false)
		return
	}

	if *goRight && relX >= xMax-margin {
		*goRight = false
		llgLog("近右界 relX=%d 改向左", relX)
		core.Sleep(50)
		return
	}
	if !*goRight && relX <= xMin+margin {
		*goRight = true
		llgLog("近左界 relX=%d 改向右", relX)
		core.Sleep(50)
		return
	}

	dir := "左"
	if *goRight {
		dir = "右"
	}
	llgLog("%s瞬移+攻击 relX=%d", dir, relX)
	llgTeleportAndAttack(c, *goRight)
}

func llgOnClearBagJumpSpot(c *Langligelang001Config, relX int) bool {
	return matchRange(relX, c.ClearBagXMin, c.ClearBagXMax)
}

func llgClearBagJumpAlignTarget(relX, xMin, xMax int) (target int, inRange bool) {
	if matchRange(relX, xMin, xMax) {
		return relX, true
	}
	if relX < xMin {
		return xMin, false
	}
	return xMax, false
}

func llgClearBagAlignDistToTarget(relX, target int) int {
	d := relX - target
	if d < 0 {
		d = -d
	}
	return d
}

func llgClearBagAlignWalkMs(c *Langligelang001Config, dist int) int {
	ms := c.ClearBagAlignWalkMsPerDist * dist
	if c.ClearBagAlignWalkMsMin > 0 && ms < c.ClearBagAlignWalkMsMin {
		ms = c.ClearBagAlignWalkMsMin
	}
	if c.ClearBagAlignWalkMsMax > 0 && ms > c.ClearBagAlignWalkMsMax {
		ms = c.ClearBagAlignWalkMsMax
	}
	return ms
}

// llgAlignToClearBagJumpX 清包跳点前对齐到 x=[clear_bag_x_min, clear_bag_x_max]。
func llgAlignToClearBagJumpX(c *Langligelang001Config, relX int) bool {
	xMin, xMax := c.ClearBagXMin, c.ClearBagXMax
	target, inRange := llgClearBagJumpAlignTarget(relX, xMin, xMax)
	if inRange {
		return false
	}
	dist := llgClearBagAlignDistToTarget(relX, target)
	goRight := relX < target
	if dist <= c.ClearBagAlignNearWalkDistMax {
		ms := llgClearBagAlignWalkMs(c, dist)
		dir := "右走"
		if !goRight {
			dir = "左走"
		}
		llgLog("自动清包: 走到跳点 x=%d→[%d,%d] 距%d %s%dms", relX, xMin, xMax, dist, dir, ms)
		walkHoldMs(goRight, ms)
		core.Sleep(100)
		return true
	}
	if relX > target {
		llgLog("自动清包: 对齐跳点 x=%d→[%d,%d] 距%d 左瞬移", relX, xMin, xMax, dist)
		llgDoTeleport(c, false)
	} else {
		llgLog("自动清包: 对齐跳点 x=%d→[%d,%d] 距%d 右瞬移", relX, xMin, xMax, dist)
		llgDoTeleport(c, true)
	}
	return true
}

func llgIsClearBagMidLayer(c *Langligelang001Config, relY int) bool {
	return relY > c.UpperYMax && relY < c.LowerYMin
}

func llgTryClearBagJumpUp(s *autoClearBagState, c *Langligelang001Config, relX int) bool {
	xMin, xMax := c.ClearBagXMin, c.ClearBagXMax
	for attempt := 1; attempt <= autoClearBagMaxRetry; attempt++ {
		curX, curY, ok := readMinimapRel()
		if !ok {
			llgLog("自动清包: 第%d次 小地图未识别", attempt)
			core.Sleep(100)
			continue
		}
		if llgIsUpperLayer(c, curY) {
			llgLog("自动清包: 第%d次 到上层 y=%d x=%d", attempt, curY, curX)
			s.pendingShop = true
			return s.tryAutoShopSellMisc(llgLog, c.clearBagSellMisc())
		}
		if !llgIsLowerLayer(c, curY) && !llgIsClearBagMidLayer(c, curY) {
			llgLog("自动清包: 第%d次 站位异常 y=%d", attempt, curY)
			core.Sleep(100)
			continue
		}
		if !llgOnClearBagJumpSpot(c, curX) {
			llgLog("自动清包: 第%d次 跳前 x=%d 不在[%d,%d] 先对齐", attempt, curX, xMin, xMax)
			llgAlignToClearBagJumpX(c, curX)
			continue
		}
		llgLog("自动清包: 第%d次 x=%d 向上跳", attempt, curX)
		tapJump()
		core.Sleep(c.ClearBagJumpCheckMs)

		curX, curY, ok = readMinimapRel()
		if !ok {
			llgLog("自动清包: 第%d次 跳后小地图未识别", attempt)
			continue
		}
		if llgIsUpperLayer(c, curY) {
			llgLog("自动清包: 第%d次 到上层 y=%d x=%d", attempt, curY, curX)
			s.pendingShop = true
			return s.tryAutoShopSellMisc(llgLog, c.clearBagSellMisc())
		}
		if curY >= c.LowerYMin {
			llgLog("自动清包: 第%d次 仍在下层 y=%d x=%d 重跳", attempt, curY, curX)
			continue
		}
		llgLog("自动清包: 第%d次 上跳成功 y=%d<%d 等%dms看上上层", attempt, curY, c.LowerYMin, c.ClearBagUpperWaitMs)
		core.Sleep(c.ClearBagUpperWaitMs)

		curX, curY, ok = readMinimapRel()
		if !ok {
			llgLog("自动清包: 第%d次 等待后小地图未识别", attempt)
			continue
		}
		if llgIsUpperLayer(c, curY) {
			llgLog("自动清包: 第%d次 到达上层 y=%d x=%d", attempt, curY, curX)
			s.pendingShop = true
			return s.tryAutoShopSellMisc(llgLog, c.clearBagSellMisc())
		}
		llgLog("自动清包: 第%d次 未到上层 y=%d x=%d 继续上跳", attempt, curY, curX)
	}
	llgLog("自动清包: 向上跳%d次未到上层 放弃本次", autoClearBagMaxRetry)
	s.finishAttempt(llgLog)
	return true
}

func tryAutoClearBagLangligelang(s *autoClearBagState, c *Langligelang001Config, relX, relY int) bool {
	if s.pendingShop {
		if llgIsUpperLayer(c, relY) {
			return s.tryAutoShopSellMisc(llgLog, c.clearBagSellMisc())
		}
		llgLog("自动清包: 已离开上层 y=%d 放弃本次", relY)
		s.finishAttempt(llgLog)
		return true
	}
	if !s.due() {
		return false
	}
	if llgIsClearBagMidLayer(c, relY) {
		if !llgOnClearBagJumpSpot(c, relX) {
			llgAlignToClearBagJumpX(c, relX)
			return true
		}
		return llgTryClearBagJumpUp(s, c, relX)
	}
	if !llgIsLowerLayer(c, relY) {
		return false
	}
	if !llgOnClearBagJumpSpot(c, relX) {
		llgAlignToClearBagJumpX(c, relX)
		return true
	}
	return llgTryClearBagJumpUp(s, c, relX)
}

func llgTryDownJumpToLower(c *Langligelang001Config, relX, relY int) bool {
	if !llgIsUpperLayer(c, relY) {
		return false
	}
	waitMs := c.DownJumpCheckMs
	for attempt := 1; attempt <= c.DownJumpMaxRetry; attempt++ {
		llgLog("上层 relX=%d relY=%d 下跳 attempt=%d", relX, relY, attempt)
		tapDownJump()
		core.Sleep(waitMs)
		curX, curY, ok := readMinimapRel()
		if !ok {
			return true
		}
		if llgIsLowerLayer(c, curY) {
			llgLog("已到下层 relX=%d relY=%d", curX, curY)
			return true
		}
		if !llgIsUpperLayer(c, curY) {
			return true
		}
		relY = curY
	}
	return true
}

// Play_浪里个浪001 下层 x=[-85,110] 来回瞬移+攻击；y=145~155 为下层。
func Play_浪里个浪001(mapAssetPath string) error {
	cfg, err := loadMapConfig(mapAssetPath)
	if err != nil {
		return err
	}
	if cfg.Langligelang001 == nil {
		return fmt.Errorf("浪里个浪001: 缺少 langligelang001 配置")
	}
	c := cfg.Langligelang001
	c.normalize()
	c.applyDeleteYellowRegion()
	defer core.ClearMinimapYellowRegion()

	SetFarmLogTag(langligelang001LogTag)
	StartFarmMaintainLoop(langligelang001LogTag)
	defer StopFarmMaintainLoop()

	goRight := true
	faceRight()
	clearBag := newAutoClearBagStateWithIntervalDefault(c.ClearBagIntervalMinMin, c.ClearBagIntervalMaxMin)
	if core.API.GetConfigBoolValue("自动清包") {
		llgLog("自动清包: 间隔默认%d~%d分钟 下层x=[%d,%d] 上跳后卖货",
			c.ClearBagIntervalMinMin, c.ClearBagIntervalMaxMin, c.ClearBagXMin, c.ClearBagXMax)
	}
	llgLog("开始挂机 下层y=[%d,%d] x=[%d,%d] 上层y=[%d,%d]",
		c.LowerYMin, c.LowerYMax, c.LowerXMin, c.LowerXMax,
		c.UpperYMin, c.UpperYMax)

	for {
		core.BlockWhileCaptchaHold()
		TickFarmMainThreadTasks()

		relX, relY, ok := readMinimapRel()
		if !ok {
			core.Sleep(100)
			continue
		}

		if tryAutoClearBagLangligelang(&clearBag, c, relX, relY) {
			continue
		}
		if llgIsUpperLayer(c, relY) {
			llgTryDownJumpToLower(c, relX, relY)
			continue
		}
		if !llgIsLowerLayer(c, relY) {
			llgLog("站位异常 relX=%d relY=%d 等待", relX, relY)
			core.Sleep(200)
			continue
		}
		llgPatrolFarmStep(c, relX, &goRight)
	}
}
