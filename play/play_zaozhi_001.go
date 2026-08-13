package play

import (
	"app/core"
	"fmt"
	"math/rand"
	"time"
)

const zaozhi001LogTag = "[早吱定制001]"

const (
	zzPhaseUpperFarm = iota
	zzPhaseGoLower
	zzPhaseLowerFarm
	zzPhaseGoUpper

	zzPatrolRightStuckJumpAt = 3
)

type zzPatrolTracker struct {
	targetLaps int
	turnCount  int
}

func (t *zzPatrolTracker) resetLaps(minLaps, maxLaps int) {
	span := maxLaps - minLaps
	t.targetLaps = minLaps
	if span > 0 {
		t.targetLaps += rand.Intn(span + 1)
	}
	t.turnCount = 0
}

func (t *zzPatrolTracker) lapsDone() int {
	return t.turnCount / 2
}

func (t *zzPatrolTracker) finished() bool {
	return t.lapsDone() >= t.targetLaps
}

func (t *zzPatrolTracker) onTurn() {
	t.turnCount++
}

type zzPatrolStuckTracker struct {
	anchorX int
	streak  int
}

func (t *zzPatrolStuckTracker) reset() {
	t.anchorX = 0
	t.streak = 0
}

func (t *zzPatrolStuckTracker) onRightTeleportNoMove(tag string, beforeX int) bool {
	curX, _, ok := readMinimapRel()
	if !ok {
		return false
	}
	if curX != beforeX {
		t.reset()
		return false
	}
	if t.streak == 0 || t.anchorX != beforeX {
		t.anchorX = beforeX
		t.streak = 1
		return false
	}
	t.streak++
	if t.streak < zzPatrolRightStuckJumpAt {
		zzLog("%s: 右瞬移 x=%d 未动 %d/%d", tag, beforeX, t.streak, zzPatrolRightStuckJumpAt)
		return false
	}
	zzLog("%s: 连续%d次右瞬移 x=%d 不变 左跳", tag, t.streak, beforeX)
	tapJumpLeft()
	core.Sleep(200)
	t.reset()
	return true
}

func zzLog(format string, args ...interface{}) {
	fmt.Printf(zaozhi001LogTag+" "+format+"\n", args...)
}

func zzIsUpperLayer(z *Zaozhi001Config, relY int) bool {
	return matchRange(relY, z.upperFarmYMin(), z.UpperYMax)
}

func zzIsLowerLayer(z *Zaozhi001Config, relY int) bool {
	return matchRange(relY, z.LowerYMin, z.LowerYMax)
}

func zzIsMidLayer(z *Zaozhi001Config, relY int) bool {
	return relY > z.UpperYMax && relY < z.LowerYMin
}

// zzIsAboveUpperLayer 明显高于上层挂机面（如 y=125），才需要下跳落回。
func zzIsAboveUpperLayer(z *Zaozhi001Config, relY int) bool {
	return relY < z.upperFarmYMin()
}

func zzReleaseClimbAtUpper(tag string) {
	releaseDpadUp()
	core.Sleep(300)
}

// zzTryDownJumpToUpper y 高于上层平台时下跳落回上层。
func zzTryDownJumpToUpper(tag string, z *Zaozhi001Config, relX, relY int) bool {
	maxRetry := z.DownJumpMaxRetry
	if maxRetry <= 0 {
		maxRetry = 5
	}
	for attempt := 1; attempt <= maxRetry; attempt++ {
		zzLog("%s: y=%d<%d 明显高于上层 relX=%d 下跳 attempt=%d", tag, relY, z.upperFarmYMin(), relX, attempt)
		tapDownJump()
		core.Sleep(z.DownJumpCheckMs)
		curX, curY, ok := readMinimapRel()
		if !ok {
			continue
		}
		if zzIsUpperLayer(z, curY) {
			zzLog("%s: 下跳后到上层 relX=%d relY=%d", tag, curX, curY)
			return true
		}
		relY = curY
		if zzIsLowerLayer(z, curY) || zzIsMidLayer(z, curY) {
			zzLog("%s: 下跳后 y=%d 未到上层", tag, curY)
			return false
		}
	}
	return false
}

type zzMidLayerTracker struct {
	anchorY  int
	since    time.Time
	tracking bool
}

func (t *zzMidLayerTracker) reset() {
	t.tracking = false
}

// zzClimbRopeUntilUpper 已在绳上时持续上爬直到上层。
func zzClimbRopeUntilUpper(z *Zaozhi001Config, tag string, relX, relY int) bool {
	zzLog("%s: 中间层 relX=%d relY=%d y不变%d秒 判定在绳上 持续上爬", tag, relX, relY, z.MidLayerStableSec)
	grabAt := time.Now()
	deadline := grabAt.Add(time.Duration(z.ClimbMaxSec) * time.Second)
	refreshDpadUpHold(z.ClimbUpTapMs)
	for time.Now().Before(deadline) {
		refreshDpadUpHold(50)
		_, curY, ok := readMinimapRel()
		if !ok {
			continue
		}
		if zzIsUpperLayer(z, curY) {
			zzLog("%s: 绳上到上层 y=%d", tag, curY)
			zzReleaseClimbAtUpper(tag)
			return true
		}
		if zzIsLowerLayer(z, curY) {
			zzLog("%s: 绳上掉到下层 y=%d", tag, curY)
			releaseDpadUp()
			return false
		}
	}
	releaseDpadUp()
	if _, curY, ok := readMinimapRel(); ok && zzIsUpperLayer(z, curY) {
		return true
	}
	zzLog("%s: 绳上爬%d秒未到上层", tag, z.ClimbMaxSec)
	return false
}

// zzTryHandleMidLayerRope 上下层之间且 y 稳定超过阈值时按绳上处理。
func zzTryHandleMidLayerRope(tag string, z *Zaozhi001Config, relX, relY int, mid *zzMidLayerTracker) bool {
	if !zzIsMidLayer(z, relY) {
		mid.reset()
		return false
	}
	stable := time.Duration(z.MidLayerStableSec) * time.Second
	if !mid.tracking || relY != mid.anchorY {
		mid.tracking = true
		mid.anchorY = relY
		mid.since = time.Now()
		zzLog("%s: 中间层 relX=%d relY=%d 监测y稳定", tag, relX, relY)
		core.Sleep(200)
		return true
	}
	if time.Since(mid.since) < stable {
		core.Sleep(200)
		return true
	}
	mid.reset()
	zzClimbRopeUntilUpper(z, tag, relX, relY)
	return true
}

func zzPatrolMargin(z *Zaozhi001Config) int {
	if z == nil || z.PatrolTeleportMargin <= 0 {
		return 12
	}
	return z.PatrolTeleportMargin
}

func zzTeleportAndAttack(z *Zaozhi001Config, goRight bool) {
	if goRight {
		faceRight()
		teleportRightAction()
	} else {
		faceLeft()
		teleportLeftAction()
	}
	keyHoldPress(attackKeyCode(), z.AttackHoldMinMs, z.AttackHoldMaxMs)
}

func zzTeleportOnly(goRight bool) {
	if goRight {
		faceRight()
		teleportRightAction()
	} else {
		faceLeft()
		teleportLeftAction()
	}
}

// zzPatrolReturnInBounds 出界只瞬移回区，不攻击。
func zzPatrolReturnInBounds(tag string, relX, xMin, xMax int, goRight *bool) {
	const maxTry = 6
	for attempt := 1; attempt <= maxTry; attempt++ {
		if relX < xMin {
			*goRight = true
			zzLog("%s: x=%d 超出左界%d 右瞬移回区 attempt=%d", tag, relX, xMin, attempt)
			zzTeleportOnly(true)
		} else if relX > xMax {
			*goRight = false
			zzLog("%s: x=%d 超出右界%d 左瞬移回区 attempt=%d", tag, relX, xMax, attempt)
			zzTeleportOnly(false)
		} else {
			return
		}
		core.Sleep(80)
		curX, _, ok := readMinimapRel()
		if !ok {
			return
		}
		relX = curX
	}
}

// zzPatrolFarmStep 平台往返瞬移+长按攻击；边界折返计圈。
func zzPatrolFarmStep(tag string, z *Zaozhi001Config, relX, xMin, xMax int, goRight *bool, laps *zzPatrolTracker, stuck *zzPatrolStuckTracker, farmSince time.Time) {
	margin := yeqiuPatrolEffectiveMargin(xMin, xMax, zzPatrolMargin(z))

	if relX < xMin {
		if stuck != nil {
			stuck.reset()
		}
		zzPatrolReturnInBounds(tag, relX, xMin, xMax, goRight)
		return
	}
	if relX > xMax {
		if stuck != nil {
			stuck.reset()
		}
		zzPatrolReturnInBounds(tag, relX, xMin, xMax, goRight)
		return
	}

	if *goRight && relX >= xMax-margin {
		*goRight = false
		if stuck != nil {
			stuck.reset()
		}
		laps.onTurn()
		zzLog("%s: 近右界 relX=%d 改向左 圈=%d/%d", tag, relX, laps.lapsDone(), laps.targetLaps)
		core.Sleep(50)
		return
	}
	if !*goRight && relX <= xMin+margin {
		*goRight = true
		if stuck != nil {
			stuck.reset()
		}
		laps.onTurn()
		zzLog("%s: 近左界 relX=%d 改向右 圈=%d/%d", tag, relX, laps.lapsDone(), laps.targetLaps)
		core.Sleep(50)
		return
	}

	dir := "左"
	if *goRight {
		dir = "右"
	}
	beforeX := relX
	if patrolFarmAllowWalk(farmSince) {
		zzLog("%s: %s走+攻击 relX=%d", tag, dir, relX)
		patrolFarmWalkAndAttack(*goRight, z.AttackHoldMinMs, z.AttackHoldMaxMs)
		if stuck != nil {
			stuck.reset()
		}
		return
	}
	zzLog("%s: %s瞬移+攻击 relX=%d", tag, dir, relX)
	zzTeleportAndAttack(z, *goRight)
	if *goRight && stuck != nil {
		stuck.onRightTeleportNoMove(tag, beforeX)
	} else if stuck != nil {
		stuck.reset()
	}
}

func zzOnDownJumpSpot(z *Zaozhi001Config, relX int) bool {
	return matchRange(relX, z.downJumpSpot1Min(), z.downJumpSpot1Max()) ||
		matchRange(relX, z.downJumpSpot2Min(), z.downJumpSpot2Max())
}

func zzOnClearBagSpot(z *Zaozhi001Config, relX int) bool {
	return matchRange(relX, z.clearBagXMin(), z.clearBagXMax())
}

func zzAlignToClearBagSpot(z *Zaozhi001Config, relX int) bool {
	if zzOnClearBagSpot(z, relX) {
		return false
	}
	target := z.ClearBagXCenter
	goRight := relX < target
	dist := zzStairDist(relX, target)
	dir := "左"
	if goRight {
		dir = "右"
	}
	if dist <= z.StairNearWalkDist {
		ms := zzStairWalkMs(z, dist, 100)
		zzLog("自动清包: 对齐 x=%d→[%d,%d] 走%s %dms", relX, z.clearBagXMin(), z.clearBagXMax(), dir, ms)
		walkHoldMs(goRight, ms)
	} else {
		zzLog("自动清包: 对齐 x=%d→[%d,%d] 瞬移%s", relX, z.clearBagXMin(), z.clearBagXMax(), dir)
		zzTeleportOnly(goRight)
	}
	core.Sleep(80)
	return true
}

func tryAutoClearBagZaozhi(s *autoClearBagState, z *Zaozhi001Config, relX, relY int) bool {
	if s.pendingShop {
		if zzIsUpperLayer(z, relY) {
			zzLog("自动清包: 已离开站台 y=%d 放弃本次", relY)
			s.finishAttempt(zzLog)
			return true
		}
		if z.isShopPlatform(relY) {
			return s.tryAutoShopSellMisc(zzLog, z.clearBagSellMisc())
		}
		return false
	}
	if !s.due() {
		return false
	}
	if !zzIsUpperLayer(z, relY) {
		return false
	}
	if !zzOnClearBagSpot(z, relX) {
		return zzAlignToClearBagSpot(z, relX)
	}
	return tryAutoClearBagUpTeleportWait(s, zzLog, relX, z.clearBagShopYBelow(), z.ClearBagUpWaitMs, func(curY int) bool {
		return zzIsUpperLayer(z, curY)
	})
}

func zzAlignToDownJumpSpot(z *Zaozhi001Config, relX int) bool {
	target := z.nearestDownJumpTarget(relX)
	tol := z.DownJumpX1Tol
	if target == z.DownJumpX2Center && z.DownJumpX2Tol > 0 {
		tol = z.DownJumpX2Tol
	}
	if matchRange(relX, target-tol, target+tol) {
		return false
	}
	goRight := relX < target
	dist := relX - target
	if dist < 0 {
		dist = -dist
	}
	dir := "左"
	if goRight {
		dir = "右"
	}
	if dist <= z.StairNearWalkDist {
		ms := z.StairNearWalkMs
		if dist <= 2 {
			ms = z.StairNearMicroStepMs
		} else if dist <= 4 {
			ms = z.StairNearStepMs
		}
		zzLog("下跳对齐: x=%d→%d 慢走%s %dms", relX, target, dir, ms)
		walkHoldMs(goRight, ms)
	} else {
		zzLog("下跳对齐: x=%d→%d 瞬移%s", relX, target, dir)
		zzTeleportAndAttack(z, goRight)
	}
	return true
}

func zzWaitReachLower(z *Zaozhi001Config) bool {
	for i := 1; i <= z.DescendPollMax; i++ {
		core.Sleep(z.DescendPollMs)
		_, curY, ok := readMinimapRel()
		if !ok {
			continue
		}
		if zzIsLowerLayer(z, curY) {
			zzLog("下跳: 第%d次检测到达下层 y=%d", i, curY)
			return true
		}
	}
	zzLog("下跳: 检测%d次未到下层", z.DescendPollMax)
	return false
}

func zzDescendToLower(z *Zaozhi001Config, mid *zzMidLayerTracker) bool {
	for attempt := 1; attempt <= z.DownJumpMaxRetry; attempt++ {
		relX, relY, ok := readMinimapRel()
		if !ok {
			core.Sleep(100)
			continue
		}
		if zzIsLowerLayer(z, relY) {
			zzLog("下跳: 已在下层 relX=%d relY=%d", relX, relY)
			return true
		}
		if !zzIsUpperLayer(z, relY) {
			if zzTryHandleMidLayerRope("下跳", z, relX, relY, mid) {
				if _, curY, ok := readMinimapRel(); ok && zzIsUpperLayer(z, curY) {
					continue
				}
				if _, curY, ok := readMinimapRel(); ok && zzIsLowerLayer(z, curY) {
					return true
				}
				continue
			}
			zzLog("下跳: 站位异常 relX=%d relY=%d", relX, relY)
			core.Sleep(200)
			continue
		}
		if !zzOnDownJumpSpot(z, relX) {
			if zzAlignToDownJumpSpot(z, relX) {
				continue
			}
		}
		target := z.nearestDownJumpTarget(relX)
		zzLog("下跳: attempt=%d relX=%d relY=%d 在跳点x≈%d 下跳", attempt, relX, relY, target)
		tapDownJump()
		core.Sleep(z.DownJumpCheckMs)
		curX, curY, ok := readMinimapRel()
		if !ok {
			continue
		}
		if curY <= z.DownJumpSuccessYAbove {
			zzLog("下跳: y=%d<=%d 未离开上层 重试", curY, z.DownJumpSuccessYAbove)
			continue
		}
		zzLog("下跳: y=%d>%d 离开上层 检测下层", curY, z.DownJumpSuccessYAbove)
		if zzWaitReachLower(z) {
			zzLog("下跳: 成功 relX=%d relY=%d", curX, curY)
			return true
		}
	}
	zzLog("下跳: %d次失败", z.DownJumpMaxRetry)
	return false
}

func zzOnStairJumpSpot(z *Zaozhi001Config, relX, stairX int) bool {
	if z.StairJumpXTolerance <= 0 {
		return relX == stairX
	}
	return matchRange(relX, z.stairJumpXMin(stairX), z.stairJumpXMax(stairX))
}

func zzStairWalkMs(z *Zaozhi001Config, dist, scalePct int) int {
	if scalePct <= 0 {
		scalePct = 100
	}
	var ms int
	switch {
	case dist <= 1:
		ms = z.StairNearMicroStepMs
	case dist <= 3:
		ms = z.StairNearStepMs
	default:
		ms = dist * z.StairAlignWalkMsPerDist
	}
	ms = ms * scalePct / 100
	if ms < z.StairAlignWalkMsMin {
		ms = z.StairAlignWalkMsMin
	}
	if ms > z.StairAlignWalkMsMax {
		ms = z.StairAlignWalkMsMax
	}
	return ms
}

func zzStairWalkOvershot(prevX, curX, stairX int) bool {
	if prevX == curX {
		return false
	}
	if prevX < stairX {
		return curX > stairX
	}
	if prevX > stairX {
		return curX < stairX
	}
	return false
}

func zzStairWalkToward(tag string, z *Zaozhi001Config, relX, stairX int, scalePct *int) int {
	dist := zzStairDist(relX, stairX)
	goRight := relX < stairX
	dir := "左"
	if goRight {
		dir = "右"
	}
	ms := zzStairWalkMs(z, dist, *scalePct)
	zzLog("%s: 楼梯走 relX=%d→%d 距%d %s%dms", tag, relX, stairX, dist, dir, ms)
	walkHoldMs(goRight, ms)
	core.Sleep(60)
	curX, _, ok := readMinimapRel()
	if !ok {
		return relX
	}
	if zzStairWalkOvershot(relX, curX, stairX) {
		next := *scalePct * 55 / 100
		if next < 40 {
			next = 40
		}
		zzLog("%s: 走过头 %d→%d 目标%d 步长缩至%d%%", tag, relX, curX, stairX, next)
		*scalePct = next
	} else if curX == relX && dist > 0 {
		next := *scalePct * 130 / 100
		if next > 160 {
			next = 160
		}
		*scalePct = next
	}
	return curX
}

// zzFineAlignStairX 微调到楼梯正下方（relX 必须等于 stairX）。
func zzFineAlignStairX(tag string, z *Zaozhi001Config, stairX int) bool {
	walkScale := 100
	var relX int
	for i := 0; i < 25; i++ {
		curX, _, ok := readMinimapRel()
		if !ok {
			core.Sleep(60)
			continue
		}
		relX = curX
		if zzOnStairJumpSpot(z, relX, stairX) {
			zzLog("%s: 楼梯精对齐就位 relX=%d 楼梯x=%d", tag, relX, stairX)
			return true
		}
		relX = zzStairWalkToward(tag, z, relX, stairX, &walkScale)
	}
	if zzOnStairJumpSpot(z, relX, stairX) {
		return true
	}
	zzLog("%s: 楼梯精对齐失败 relX=%d 目标x=%d", tag, relX, stairX)
	return false
}

func zzStairDist(relX, stairX int) int {
	d := relX - stairX
	if d < 0 {
		d = -d
	}
	return d
}

func zzWaitAfterAttackBeforeJump(z *Zaozhi001Config) {
	core.RandomSleep(z.ClimbAttackJumpWaitMsMin, z.ClimbAttackJumpWaitMsMax)
}

func zzAlignToStair(tag string, z *Zaozhi001Config, stairX int) bool {
	walkScale := 100
	var relX int
	for pass := 0; pass < z.AlignMaxPass; pass++ {
		curX, relY, ok := readMinimapRel()
		if !ok {
			core.Sleep(80)
			continue
		}
		relX = curX
		if zzIsUpperLayer(z, relY) {
			zzLog("%s: 已在上层 relX=%d relY=%d 无需楼梯对齐", tag, relX, relY)
			return true
		}
		if zzOnStairJumpSpot(z, relX, stairX) {
			return zzFineAlignStairX(tag, z, stairX)
		}
		dist := zzStairDist(relX, stairX)
		if dist <= z.StairNearWalkDist {
			zzLog("%s: 楼梯对齐 pass=%d relX=%d→%d", tag, pass+1, relX, stairX)
			relX = zzStairWalkToward(tag, z, relX, stairX, &walkScale)
			if zzOnStairJumpSpot(z, relX, stairX) {
				return zzFineAlignStairX(tag, z, stairX)
			}
			continue
		}
		goRight := relX < stairX
		dir := "左"
		if goRight {
			dir = "右"
		}
		zzLog("%s: 楼梯对齐 pass=%d relX=%d→%d 瞬移%s", tag, pass+1, relX, stairX, dir)
		if goRight {
			faceRight()
			teleportRightAction()
		} else {
			faceLeft()
			teleportLeftAction()
		}
		sleepAfterTeleport()
		tapAttackOnce()
		zzWaitAfterAttackBeforeJump(z)
	}
	if zzFineAlignStairX(tag, z, stairX) {
		return true
	}
	zzLog("%s: 楼梯对齐达最大轮次", tag)
	return false
}

func zzStartClimbGrab(tag string, z *Zaozhi001Config, relX int) {
	zzLog("%s: 爬梯 x=%d 跳→上 等%dms", tag, relX, z.ClimbJumpWaitMs)
	tapJump()
	core.Sleep(z.ClimbJumpWaitMs)
	refreshDpadUpHold(z.ClimbUpTapMs)
}

func zzClimbStillOnLower(z *Zaozhi001Config, minY, curY int, sinceGrab time.Duration) bool {
	if !zzIsLowerLayer(z, curY) {
		return false
	}
	if minY < z.LowerYMin {
		return true
	}
	return sinceGrab >= time.Duration(z.ClimbLowerRetryMs)*time.Millisecond
}

func zzClimbStair(tag string, z *Zaozhi001Config, stairX int) bool {
	if _, relY, ok := readMinimapRel(); ok && zzIsUpperLayer(z, relY) {
		zzLog("%s: 已在上层 y=%d 无需爬梯", tag, relY)
		return true
	}
	minClimb := time.Duration(z.ClimbMinSec) * time.Second
	maxClimb := time.Duration(z.ClimbMaxSec) * time.Second
	for attempt := 0; attempt < 5; attempt++ {
		if _, relY, ok := readMinimapRel(); ok && zzIsUpperLayer(z, relY) {
			zzLog("%s: attempt=%d 已在上层 y=%d", tag, attempt+1, relY)
			return true
		}
		zzLog("%s: 对齐前攻击清怪 attempt=%d 楼梯x=%d", tag, attempt+1, stairX)
		tapAttackOnce()
		zzWaitAfterAttackBeforeJump(z)
		if !zzAlignToStair(tag, z, stairX) {
			return false
		}
		relX, startY, ok := readMinimapRel()
		if !ok {
			continue
		}
		if zzIsUpperLayer(z, startY) {
			zzLog("%s: 对齐后已在上层 relX=%d relY=%d", tag, relX, startY)
			return true
		}
		if !zzFineAlignStairX(tag, z, stairX) {
			zzLog("%s: 跳前未精确站到楼梯 x=%d", tag, stairX)
			continue
		}
		relX, startY, ok = readMinimapRel()
		if !ok {
			continue
		}
		zzWaitAfterAttackBeforeJump(z)
		zzStartClimbGrab(tag, z, relX)

		minY := startY
		failedOnLower := false
		grabAt := time.Now()
		deadline := grabAt.Add(maxClimb)
		for time.Now().Before(deadline) {
			refreshDpadUpHold(50)
			_, curY, ok := readMinimapRel()
			if !ok {
				continue
			}
			if curY < minY {
				minY = curY
			}
			elapsed := time.Since(grabAt)
			if zzClimbStillOnLower(z, minY, curY, elapsed) {
				zzLog("%s: 爬梯 y=%d 仍在下层 minY=%d 重爬 attempt=%d", tag, curY, minY, attempt+1)
				releaseDpadUp()
				failedOnLower = true
				break
			}
			if elapsed >= minClimb && zzIsUpperLayer(z, curY) {
				zzLog("%s: 爬梯到上层 y=%d", tag, curY)
				zzReleaseClimbAtUpper(tag)
				return true
			}
		}
		releaseDpadUp()
		if _, relY, ok := readMinimapRel(); ok && zzIsUpperLayer(z, relY) {
			zzLog("%s: 爬梯结束已在上层 y=%d", tag, relY)
			return true
		}
		if failedOnLower {
			continue
		}
		zzLog("%s: 爬梯超时%d秒未到达上层 attempt=%d", tag, z.ClimbMaxSec, attempt+1)
	}
	return false
}

func zzAscendToUpper(z *Zaozhi001Config) bool {
	relX, relY, ok := readMinimapRel()
	if ok && zzIsUpperLayer(z, relY) {
		zzLog("上爬: 已在上层 relX=%d relY=%d 无需爬梯", relX, relY)
		return true
	}
	stairX := 44
	if ok {
		stairX = z.nearestStairX(relX)
	}
	zzLog("上爬: 选楼梯 x=%d", stairX)
	if zzClimbStair("上爬", z, stairX) {
		return true
	}
	for _, alt := range z.StairXs {
		if alt == stairX {
			continue
		}
		zzLog("上爬: 换楼梯 x=%d", alt)
		if zzClimbStair("上爬", z, alt) {
			return true
		}
	}
	return false
}

// Play_早吱定制001 上层3-5圈→下跳下层→下层1-2圈→爬梯回上层。
func Play_早吱定制001(mapAssetPath string) error {
	cfg, err := loadMapConfig(mapAssetPath)
	if err != nil {
		return err
	}
	if cfg.Zaozhi001 == nil {
		return fmt.Errorf("早吱定制001: 缺少 zaozhi001 配置")
	}
	z := cfg.Zaozhi001
	z.normalize()
	z.applyDeleteYellowRegion()
	defer core.ClearMinimapYellowRegion()

	SetFarmLogTag(zaozhi001LogTag)
	StartFarmMaintainLoop(zaozhi001LogTag)
	defer StopFarmMaintainLoop()

	goRight := true
	faceRight()
	var laps zzPatrolTracker
	var patrolStuck zzPatrolStuckTracker
	var midLayer zzMidLayerTracker
	var upperArrivalGrace time.Time
	var farmSince time.Time
	markFarmSince := func() {
		farmSince = time.Now()
	}
	clearBag := newAutoClearBagStateWithIntervalDefault(z.ClearBagIntervalMinMin, z.ClearBagIntervalMaxMin)
	phase := zzPhaseUpperFarm
	laps.resetLaps(z.UpperLapsMin, z.UpperLapsMax)
	markFarmSince()

	markUpperGrace := func() {
		upperArrivalGrace = time.Now().Add(8 * time.Second)
	}

	zzLog("开始挂机 上层y=[%d,%d](检测≥%d)x=[%d,%d] 下层y=[%d,%d]x=[%d,%d] 楼梯x=%v",
		z.UpperYMin, z.UpperYMax, z.upperFarmYMin(), z.UpperXMin, z.UpperXMax,
		z.LowerYMin, z.LowerYMax, z.LowerXMin, z.LowerXMax, z.StairXs)
	if core.API.GetConfigBoolValue("自动清包") {
		zzLog("自动清包: 上层x=[%d,%d] 上+瞬移到站台 默认间隔%d~%d分钟",
			z.clearBagXMin(), z.clearBagXMax(), z.ClearBagIntervalMinMin, z.ClearBagIntervalMaxMin)
	}

	if relX, relY, ok := readMinimapRel(); ok {
		if zzIsLowerLayer(z, relY) {
			phase = zzPhaseLowerFarm
			laps.resetLaps(z.LowerLapsMin, z.LowerLapsMax)
			markFarmSince()
			zzLog("启动在下层 relX=%d relY=%d 先刷下层", relX, relY)
		} else if zzIsUpperLayer(z, relY) {
			zzLog("启动在上层 relX=%d relY=%d 先刷上层", relX, relY)
		}
	}

	for {
		core.BlockWhileCaptchaHold()
		TickFarmMainThreadTasks()

		switch phase {
		case zzPhaseGoLower:
			if relX, relY, ok := readMinimapRel(); ok && zzIsLowerLayer(z, relY) {
				midLayer.reset()
				phase = zzPhaseLowerFarm
				goRight = true
				faceRight()
				laps.resetLaps(z.LowerLapsMin, z.LowerLapsMax)
				markFarmSince()
				zzLog("已在下层 relX=%d relY=%d 转下层刷怪 %d圈", relX, relY, laps.targetLaps)
				continue
			}
			if zzDescendToLower(z, &midLayer) {
				phase = zzPhaseLowerFarm
				goRight = true
				faceRight()
				laps.resetLaps(z.LowerLapsMin, z.LowerLapsMax)
				markFarmSince()
				zzLog("转下层刷怪 %d圈", laps.targetLaps)
			} else {
				phase = zzPhaseUpperFarm
				laps.resetLaps(z.UpperLapsMin, z.UpperLapsMax)
				markFarmSince()
				zzLog("下跳失败 继续上层刷怪")
			}
			continue
		case zzPhaseGoUpper:
			if relX, relY, ok := readMinimapRel(); ok && zzIsUpperLayer(z, relY) {
				midLayer.reset()
				phase = zzPhaseUpperFarm
				goRight = true
				faceRight()
				laps.resetLaps(z.UpperLapsMin, z.UpperLapsMax)
				markUpperGrace()
				markFarmSince()
				zzLog("已在上层 relX=%d relY=%d 转上层刷怪 %d圈", relX, relY, laps.targetLaps)
				continue
			}
			if zzAscendToUpper(z) {
				phase = zzPhaseUpperFarm
				goRight = true
				faceRight()
				laps.resetLaps(z.UpperLapsMin, z.UpperLapsMax)
				markUpperGrace()
				markFarmSince()
				zzLog("到上层 刷怪 %d圈", laps.targetLaps)
			} else {
				phase = zzPhaseLowerFarm
				laps.resetLaps(z.LowerLapsMin, z.LowerLapsMax)
				markFarmSince()
				zzLog("上爬失败 继续下层刷怪")
			}
			continue
		}

		relX, relY, ok := readMinimapRel()
		if !ok {
			core.Sleep(100)
			continue
		}

		if phase == zzPhaseUpperFarm {
			if !zzIsUpperLayer(z, relY) {
				if zzIsLowerLayer(z, relY) {
					midLayer.reset()
					zzLog("上层阶段但已在下层 转下层刷怪")
					phase = zzPhaseLowerFarm
					laps.resetLaps(z.LowerLapsMin, z.LowerLapsMax)
					markFarmSince()
				} else if !upperArrivalGrace.IsZero() && time.Now().Before(upperArrivalGrace) {
					zzLog("上层: 刚到上层 relX=%d relY=%d 跳过纠偏", relX, relY)
				} else if zzIsAboveUpperLayer(z, relY) {
					midLayer.reset()
					zzTryDownJumpToUpper("上层", z, relX, relY)
				} else if zzTryHandleMidLayerRope("上层", z, relX, relY, &midLayer) {
					continue
				} else {
					zzLog("上层阶段站位异常 relX=%d relY=%d", relX, relY)
					core.Sleep(200)
				}
				continue
			}
			midLayer.reset()
			if tryAutoClearBagZaozhi(&clearBag, z, relX, relY) {
				continue
			}
			if laps.finished() {
				zzLog("上层刷够 %d圈 去下层", laps.targetLaps)
				phase = zzPhaseGoLower
				continue
			}
			zzPatrolFarmStep("上层", z, relX, z.UpperXMin, z.UpperXMax, &goRight, &laps, &patrolStuck, farmSince)
			continue
		}

		if phase == zzPhaseLowerFarm {
			if !zzIsLowerLayer(z, relY) {
				if zzIsUpperLayer(z, relY) {
					midLayer.reset()
					zzLog("下层阶段但已在上层 转上层刷怪")
					phase = zzPhaseUpperFarm
					laps.resetLaps(z.UpperLapsMin, z.UpperLapsMax)
					markFarmSince()
				} else if zzIsAboveUpperLayer(z, relY) {
					midLayer.reset()
					zzTryDownJumpToUpper("下层", z, relX, relY)
				} else if zzTryHandleMidLayerRope("下层", z, relX, relY, &midLayer) {
					continue
				} else {
					zzLog("下层阶段站位异常 relX=%d relY=%d", relX, relY)
					core.Sleep(200)
				}
				continue
			}
			midLayer.reset()
			if laps.finished() {
				zzLog("下层刷够 %d圈 去上层", laps.targetLaps)
				phase = zzPhaseGoUpper
				continue
			}
			zzPatrolFarmStep("下层", z, relX, z.LowerXMin, z.LowerXMax, &goRight, &laps, &patrolStuck, farmSince)
		}
	}
}
