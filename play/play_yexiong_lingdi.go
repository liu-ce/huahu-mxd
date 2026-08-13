package play

import (
	"app/core"
	"fmt"
	"math/rand"
	"time"
)

const yexiongLingdiLogTag = "[野熊的领地]"

const yxAlignStuckJumpAt = 5 // 对齐连续 N 次 x 未变才同向跳脱困

const (
	yxPhaseLowerFarm = iota
	yxPhaseGoMiddle
	yxPhaseMiddleFarm
	yxPhaseGoUpperRight
	yxPhaseUpperRightFarm
	yxPhaseGoLowerFromSide
	yxPhaseGoUpperLeft
	yxPhaseUpperLeftFarm
	yxPhaseGoUpper
	yxPhaseUpperClearBag
)

const (
	yxCycleMiddleA = iota
	yxCycleUpperRight
	yxCycleLowerA
	yxCycleMiddleB
	yxCycleUpperLeft
	yxCycleLowerB
	yxCycleMiddleC
)

func yxLog(format string, args ...interface{}) {
	fmt.Printf(yexiongLingdiLogTag+" "+format+"\n", args...)
}

func yxPatrolMargin(c *YexiongLingdiConfig) int {
	if c == nil || c.PatrolTeleportMargin <= 0 {
		return 12
	}
	return c.PatrolTeleportMargin
}

func yxIsLowerLayer(c *YexiongLingdiConfig, relY int) bool {
	return matchRange(relY, c.LowerYMin, c.LowerYMax)
}

func yxIsMiddleLayer(c *YexiongLingdiConfig, relY int) bool {
	return matchRange(relY, c.MiddleYMin, c.MiddleYMax)
}

func yxIsUpperClearPlatform(c *YexiongLingdiConfig, relY int) bool {
	return matchRange(relY, c.UpperYMin, c.UpperYMax)
}

func yxIsUpperRightPlatform(c *YexiongLingdiConfig, relX, relY int) bool {
	return matchRange(relY, c.UpperRightYMin, c.UpperRightYMax) &&
		matchRange(relX, c.UpperRightXMin, c.UpperRightXMax)
}

func yxIsUpperLeftPlatform(c *YexiongLingdiConfig, relX, relY int) bool {
	return matchRange(relY, c.UpperLeftYMin, c.UpperLeftYMax) &&
		matchRange(relX, c.UpperLeftXMin, c.UpperLeftXMax)
}

func yxIsUpperSidePlatform(c *YexiongLingdiConfig, relX, relY int) bool {
	return yxIsUpperRightPlatform(c, relX, relY) || yxIsUpperLeftPlatform(c, relX, relY)
}

func yxIsUpperSideLayerY(c *YexiongLingdiConfig, relY int) bool {
	return matchRange(relY, c.UpperRightYMin, c.UpperRightYMax)
}

func yxLayerName(c *YexiongLingdiConfig, relY int) string {
	switch {
	case yxIsLowerLayer(c, relY):
		return "下层"
	case yxIsMiddleLayer(c, relY):
		return "中层"
	case yxIsUpperSideLayerY(c, relY):
		return "侧上"
	case yxIsUpperSideRopeBetweenLayer(c, relY):
		return "侧上绳"
	case yxIsUpperClearPlatform(c, relY):
		return "清包上"
	case yxIsUpperRopeBetweenLayer(c, relY):
		return "清包绳"
	case yxIsRopeBetweenLayer(c, relY):
		return "中绳"
	default:
		return fmt.Sprintf("未知y=%d", relY)
	}
}

// yxReadyForUpperSideSideJump 侧上平台绳跳区：侧上 y 或侧上绳区，排除下层/中层。
func yxReadyForUpperSideSideJump(c *YexiongLingdiConfig, relY int) bool {
	if yxIsLowerLayer(c, relY) || yxIsMiddleLayer(c, relY) {
		return false
	}
	if yxIsUpperSideLayerY(c, relY) {
		return true
	}
	return relY > c.UpperSideClimbYAbove && relY < c.MiddleYMin
}

// yxWaitYSettled 每 layer_poll_ms 检测 y；连续 layer_stable_count 次不变视为落地。
func yxWaitYSettled(c *YexiongLingdiConfig, tag string) (relX, relY int, ok bool) {
	pollMs := c.LayerPollMs
	needStable := c.LayerStableCount
	maxPolls := c.DescendPollMax
	if maxPolls <= 0 {
		maxPolls = 40
	}
	lastY := -1
	stable := 0
	for i := 1; i <= maxPolls; i++ {
		core.Sleep(pollMs)
		x, y, readOk := readMinimapRel()
		if !readOk {
			stable = 0
			lastY = -1
			continue
		}
		yxLog("%s: 层检测 poll=%d relX=%d relY=%d 层=%s", tag, i, x, y, yxLayerName(c, y))
		if lastY >= 0 && y == lastY {
			stable++
			if stable >= needStable {
				return x, y, true
			}
		} else {
			stable = 1
			lastY = y
		}
	}
	return 0, 0, false
}

func yxOnClearBagX(c *YexiongLingdiConfig, relX int) bool {
	return matchRange(relX, c.ClearBagXMin, c.ClearBagXMax)
}

// yxIsRopeBetweenLayer 中层与下层之间的绳/过渡层（如 y=153）。
func yxIsRopeBetweenLayer(c *YexiongLingdiConfig, relY int) bool {
	return relY > c.MiddleYMax && relY < c.LowerYMin
}

// yxIsUpperRopeBetweenLayer 清包上平台与侧上平台之间的绳/过渡层（如 y=125）。
func yxIsUpperRopeBetweenLayer(c *YexiongLingdiConfig, relY int) bool {
	return relY > c.UpperYMax && relY < c.UpperRightYMin
}

// yxIsUpperSideRopeBetweenLayer 右上/左上平台与中层之间的过渡层。
func yxIsUpperSideRopeBetweenLayer(c *YexiongLingdiConfig, relY int) bool {
	return relY > c.UpperRightYMax && relY < c.MiddleYMin
}

func yxIsSpecialRopeZone(c *YexiongLingdiConfig, relX, relY int) bool {
	if yxIsUpperSideLayerY(c, relY) {
		return false
	}
	// 中层刷怪 x 走廊内、y 接近中层平台时不算爬绳过渡区（小地图 y 常偏 141~145）
	if matchRange(relX, c.MiddleXMin, c.MiddleXMax) &&
		matchRange(relY, c.MiddleYMin-5, c.MiddleYMax+2) {
		return false
	}
	return yxIsRopeBetweenLayer(c, relY) || yxIsUpperRopeBetweenLayer(c, relY) || yxIsUpperSideRopeBetweenLayer(c, relY)
}

func yxTryRecoverRopeBetweenLayer(tag string, c *YexiongLingdiConfig, relX, relY int) bool {
	if !yxIsSpecialRopeZone(c, relX, relY) {
		return false
	}
	split := c.AbnormalLayerJumpX
	if relX < split {
		yxLog("%s: 爬绳区 relX=%d relY=%d x<%d 右跳", tag, relX, relY, split)
		tapJumpRight()
	} else {
		yxLog("%s: 爬绳区 relX=%d relY=%d x>=%d 左跳", tag, relX, relY, split)
		tapJumpLeft()
	}
	core.Sleep(c.AbnormalLayerJumpLandMs)
	return true
}

func yxAlignWalkMs(c *YexiongLingdiConfig, dist int) int {
	if dist < 1 {
		dist = 1
	}
	ms := dist * c.AlignWalkMsPerDist
	if ms < c.AlignWalkMsMin {
		ms = c.AlignWalkMsMin
	}
	if c.AlignWalkMsMax > 0 && ms > c.AlignWalkMsMax {
		ms = c.AlignWalkMsMax
	}
	return ms
}

func yxRopeAlignWalkMs(c *YexiongLingdiConfig, dist int) int {
	if dist < 1 {
		dist = 1
	}
	ms := dist * c.RopeAlignWalkMsPerDist
	if ms < c.RopeAlignWalkMsMin {
		ms = c.RopeAlignWalkMsMin
	}
	if c.AlignWalkMsMax > 0 && ms > c.AlignWalkMsMax {
		ms = c.AlignWalkMsMax
	}
	return ms
}

func yxAlignWalkMsForTag(c *YexiongLingdiConfig, tag string, dist int) int {
	if tag == "爬绳" || tag == "爬上层" {
		return yxRopeAlignWalkMs(c, dist)
	}
	return yxAlignWalkMs(c, dist)
}

func yxRopeJumpSpotTol(c *YexiongLingdiConfig) int {
	if c.RopeJumpSpotXTolerance > 0 {
		return c.RopeJumpSpotXTolerance
	}
	return 2
}

func yxOnRopeJumpSpotReady(c *YexiongLingdiConfig, relX int) bool {
	if yxOnRopeJumpSpot(c, relX) {
		return true
	}
	tol := yxRopeJumpSpotTol(c)
	if matchRange(relX, c.RopeJumpRightXMin-tol, c.RopeJumpRightXMax+tol) {
		return true
	}
	return matchRange(relX, c.RopeJumpLeftXMin-tol, c.RopeJumpLeftXMax+tol)
}

func yxRopeJumpGoRight(c *YexiongLingdiConfig, relX int) bool {
	tol := yxRopeJumpSpotTol(c)
	if matchRange(relX, c.RopeJumpRightXMin-tol, c.RopeJumpRightXMax+tol) {
		return true
	}
	if matchRange(relX, c.RopeJumpLeftXMin-tol, c.RopeJumpLeftXMax+tol) {
		return false
	}
	return relX <= c.ropeJumpRightCenter()
}

type yxAlignStuckTracker struct {
	anchorX int
	streak  int
}

type yxAlignMoveState struct {
	lastTeleportDir int // -1 左瞬移, 0 无, 1 右瞬移
}

func (s *yxAlignMoveState) reset() {
	if s != nil {
		s.lastTeleportDir = 0
	}
}

func (t *yxAlignStuckTracker) reset() {
	t.anchorX = 0
	t.streak = 0
}

func (t *yxAlignStuckTracker) onNoMove(c *YexiongLingdiConfig, tag string, pass, beforeX int, goRight bool) {
	dir := "左"
	if goRight {
		dir = "右"
	}
	if t.streak == 0 || t.anchorX != beforeX {
		t.anchorX = beforeX
		t.streak = 1
		yxLog("%s: pass=%d x=%d未动 %d/%d", tag, pass, beforeX, t.streak, yxAlignStuckJumpAt)
		return
	}
	t.streak++
	if t.streak < yxAlignStuckJumpAt {
		yxLog("%s: pass=%d x=%d未动 %d/%d", tag, pass, beforeX, t.streak, yxAlignStuckJumpAt)
		return
	}
	yxLog("%s: 连续%d次对齐 x=%d 未动 %s跳", tag, t.streak, beforeX, dir)
	if goRight {
		tapJumpRight()
	} else {
		tapJumpLeft()
	}
	core.Sleep(c.AbnormalLayerJumpLandMs)
	t.reset()
}

// yxAlignMoveTowardX 走向/瞬移目标 x；禁止连续同向瞬移；walkOnly 时只走不瞬移；walkMsOverride>0 时用指定走步时长。
func yxAlignMoveTowardX(c *YexiongLingdiConfig, tag string, pass int, relX, targetX int, stuck *yxAlignStuckTracker, moveState *yxAlignMoveState, walkOnly bool, walkMsOverride int) {
	goRight := relX < targetX
	dist := relX - targetX
	if dist < 0 {
		dist = -dist
	}
	dir := "左"
	if goRight {
		dir = "右"
	}
	beforeX := relX
	useTeleport := !walkOnly && dist > c.RopeNearWalkDist
	if useTeleport && moveState != nil && moveState.lastTeleportDir != 0 {
		sameDir := (goRight && moveState.lastTeleportDir > 0) || (!goRight && moveState.lastTeleportDir < 0)
		if sameDir {
			useTeleport = false
			yxLog("%s: 对齐 pass=%d 连续同向瞬移改走", tag, pass)
		}
	}
	if useTeleport {
		yxLog("%s: 对齐 pass=%d relX=%d→%d 瞬移%s", tag, pass, relX, targetX, dir)
		yxTeleportOnly(goRight)
		if moveState != nil {
			if goRight {
				moveState.lastTeleportDir = 1
			} else {
				moveState.lastTeleportDir = -1
			}
		}
	} else {
		walkMs := walkMsOverride
		if walkMs <= 0 {
			walkMs = yxAlignWalkMsForTag(c, tag, dist)
		}
		yxLog("%s: 对齐 pass=%d relX=%d→%d 走%s %dms", tag, pass, relX, targetX, dir, walkMs)
		walkHoldMs(goRight, walkMs)
		if moveState != nil {
			moveState.reset()
		}
	}
	core.Sleep(80)
	afterX, _, ok := readMinimapRel()
	if ok && afterX == beforeX {
		stuck.onNoMove(c, tag, pass, beforeX, goRight)
	} else if ok {
		stuck.reset()
	}
}

func yxAlignPreAttack(c *YexiongLingdiConfig, tag string) {
	yxLog("%s: 对齐前攻击", tag)
	keyHoldPress(attackKeyCode(), c.AlignPreAttackMinMs, c.AlignPreAttackMaxMs)
	core.Sleep(c.AlignPreWaitMs)
}

func yxTeleportAndAttack(c *YexiongLingdiConfig, goRight bool) {
	if goRight {
		faceRight()
		teleportRightAction()
	} else {
		faceLeft()
		teleportLeftAction()
	}
	core.RandomSleep(c.TeleportAttackMinMs, c.TeleportAttackMaxMs)
	keyHoldPress(attackKeyCode(), c.AttackHoldMinMs, c.AttackHoldMaxMs)
}

func yxTeleportOnly(goRight bool) {
	if goRight {
		faceRight()
		teleportRightAction()
	} else {
		faceLeft()
		teleportLeftAction()
	}
}

func yxWalkAndAttack(c *YexiongLingdiConfig, goRight bool) {
	if goRight {
		faceRight()
	} else {
		faceLeft()
	}
	walkMs := c.PatrolWalkHoldMinMs
	if c.PatrolWalkHoldMaxMs > c.PatrolWalkHoldMinMs {
		walkMs = c.PatrolWalkHoldMinMs + rand.Intn(c.PatrolWalkHoldMaxMs-c.PatrolWalkHoldMinMs+1)
	}
	walkHoldMs(goRight, walkMs)
	keyHoldPress(attackKeyCode(), c.AttackHoldMinMs, c.AttackHoldMaxMs)
}

// yxPatrolFarmStep 同向瞬移+攻击；近边界提前折返，避免瞬移过冲；中层禁止连续同向瞬移；allowWalk 时按概率改成长按方向键走+攻击。
func yxPatrolFarmStep(tag string, c *YexiongLingdiConfig, relX, turnLeftX, turnRightX int, goRight *bool, laps *zzPatrolTracker, allowWalk bool, lastTpDir *int) {
	margin := yeqiuPatrolEffectiveMargin(turnLeftX, turnRightX, yxPatrolMargin(c))
	resetTpDir := func() {
		if lastTpDir != nil {
			*lastTpDir = 0
		}
	}

	if !*goRight && relX < turnLeftX {
		*goRight = true
		laps.onTurn()
		yxLog("%s: x=%d<%d 改向右 圈=%d/%d", tag, relX, turnLeftX, laps.lapsDone(), laps.targetLaps)
	} else if *goRight && relX > turnRightX {
		*goRight = false
		laps.onTurn()
		yxLog("%s: x=%d>%d 改向左 圈=%d/%d", tag, relX, turnRightX, laps.lapsDone(), laps.targetLaps)
	}

	if *goRight && relX >= turnRightX-margin {
		*goRight = false
		laps.onTurn()
		yxLog("%s: 近右界 relX=%d 改向左 圈=%d/%d", tag, relX, laps.lapsDone(), laps.targetLaps)
		resetTpDir()
		return
	}
	if !*goRight && relX <= turnLeftX+margin {
		*goRight = true
		laps.onTurn()
		yxLog("%s: 近左界 relX=%d 改向右 圈=%d/%d", tag, relX, laps.lapsDone(), laps.targetLaps)
		resetTpDir()
		return
	}

	dir := "左"
	if *goRight {
		dir = "右"
	}
	if allowWalk && c.PatrolWalkChancePercent > 0 && rand.Intn(100) < c.PatrolWalkChancePercent {
		yxLog("%s: %s走+攻击 relX=%d", tag, dir, relX)
		yxWalkAndAttack(c, *goRight)
		resetTpDir()
		return
	}
	if lastTpDir != nil && *lastTpDir != 0 {
		sameDir := (*lastTpDir > 0 && *goRight) || (*lastTpDir < 0 && !*goRight)
		if sameDir {
			yxLog("%s: 连续同向瞬移改走 relX=%d", tag, relX)
			yxWalkAndAttack(c, *goRight)
			resetTpDir()
			return
		}
	}
	yxLog("%s: %s瞬移+攻击 relX=%d", tag, dir, relX)
	yxTeleportAndAttack(c, *goRight)
	if lastTpDir != nil {
		if *goRight {
			*lastTpDir = 1
		} else {
			*lastTpDir = -1
		}
	}
	core.Sleep(c.PatrolTeleportSettleMs)
}

func yxOnRopeJumpRightSpot(c *YexiongLingdiConfig, relX int) bool {
	return matchRange(relX, c.RopeJumpRightXMin, c.RopeJumpRightXMax)
}

func yxOnRopeJumpLeftSpot(c *YexiongLingdiConfig, relX int) bool {
	return matchRange(relX, c.RopeJumpLeftXMin, c.RopeJumpLeftXMax)
}

func yxOnRopeAlignZone(c *YexiongLingdiConfig, relX int) bool {
	center := c.ropeJumpRightCenter()
	tol := c.RopeAlignXTolerance
	return matchRange(relX, center-tol, center+tol)
}

func yxOnXRange(xMin, xMax, relX int) bool {
	return matchRange(relX, xMin, xMax)
}

// yxOnSideJumpAlignSpot 侧上绳跳点：仅跳点及绳方向一侧少许容差（从右侧/左侧接近，不扩跳点内侧）。
func yxOnSideJumpAlignSpot(goRight bool, relX, xMin, xMax, overshootTol int) bool {
	return matchRange(relX, xMin, xMax+overshootTol)
}

func yxSideJumpOvershootTol(c *YexiongLingdiConfig, goRight bool) int {
	if c.UpperSideJumpXOvershoot > 0 {
		return c.UpperSideJumpXOvershoot
	}
	if goRight {
		return 2
	}
	return 1
}

func yxSideJumpAcceptXMax(xMax, overshootTol int) int {
	return xMax + overshootTol
}

func yxOnSideJumpAlignSpotOnMiddle(c *YexiongLingdiConfig, goRight bool, relX, relY, xMin, xMax, overshootTol int) bool {
	return yxIsMiddleLayer(c, relY) && yxOnSideJumpAlignSpot(goRight, relX, xMin, xMax, overshootTol)
}

func yxSideJumpAlignWalkOnly(c *YexiongLingdiConfig, goRight bool, relX, targetX, xMin, xMax int) bool {
	dist := relX - targetX
	if dist < 0 {
		dist = -dist
	}
	if dist <= c.RopeNearWalkDist*2 {
		return true
	}
	if goRight && relX >= xMin-c.RopeNearWalkDist {
		return true
	}
	if !goRight && relX <= xMax+c.RopeNearWalkDist {
		return true
	}
	// 左上：已过跳点仍在绳右侧（如 x=13），只走不瞬移
	if !goRight && relX > xMax && relX <= xMax+c.RopeNearWalkDist*2 {
		return true
	}
	return false
}

func yxSideJumpApproachOffset(c *YexiongLingdiConfig) int {
	if c.UpperSideAlignApproachOffset > 0 {
		return c.UpperSideAlignApproachOffset
	}
	return 2
}

// yxSideJumpAlignMoveTarget 远距先走到跳点前若干格（右上 66→64，左上 11→13），近距再微调。
func yxSideJumpAlignMoveTarget(c *YexiongLingdiConfig, goRight bool, relX, jumpX int) int {
	off := yxSideJumpApproachOffset(c)
	if goRight {
		if relX < jumpX-off {
			return jumpX - off
		}
		return jumpX
	}
	if relX > jumpX+off {
		return jumpX + off
	}
	return jumpX
}

func yxSideJumpAlignWalkMs(c *YexiongLingdiConfig, goRight bool, relX, moveTarget, jumpX int) int {
	dist := moveTarget - relX
	if dist < 0 {
		dist = -dist
	}
	if dist < 1 {
		dist = 1
	}
	off := yxSideJumpApproachOffset(c)
	finePhase := (goRight && relX >= jumpX-off) || (!goRight && relX <= jumpX+off)
	if finePhase {
		ms := dist * c.AlignWalkMsPerDist
		if ms < c.AlignWalkMsMin {
			ms = c.AlignWalkMsMin
		}
		maxFine := c.UpperSideAlignFineWalkMsMax
		if maxFine <= 0 {
			maxFine = 160
		}
		if ms > maxFine {
			ms = maxFine
		}
		return ms
	}
	perDist := c.UpperSideAlignCoarseWalkMsPerDist
	if perDist <= 0 {
		perDist = 55
	}
	ms := dist * perDist
	minCoarse := c.UpperSideAlignCoarseWalkMsMin
	if minCoarse <= 0 {
		minCoarse = 200
	}
	if ms < minCoarse {
		ms = minCoarse
	}
	maxCoarse := c.UpperSideAlignCoarseWalkMsMax
	if maxCoarse <= 0 {
		maxCoarse = 900
	}
	if ms > maxCoarse {
		ms = maxCoarse
	}
	return ms
}

func yxFineAlignToSideJumpSpot(c *YexiongLingdiConfig, tag string, goRight bool, xMin, xMax int) bool {
	overshootTol := yxSideJumpOvershootTol(c, goRight)
	acceptMax := yxSideJumpAcceptXMax(xMax, overshootTol)
	targetX := (xMin + xMax) / 2
	if relX, relY, ok := readMinimapRel(); ok && yxOnSideJumpAlignSpotOnMiddle(c, goRight, relX, relY, xMin, xMax, overshootTol) {
		yxLog("%s: 跳点就绪 relX=%d x∈[%d,%d] 直接跳", tag, relX, xMin, acceptMax)
		return true
	}
	var stuck yxAlignStuckTracker
	var moveState yxAlignMoveState
	for pass := 0; pass < c.RopeAlignMaxPass; pass++ {
		relX, relY, ok := readMinimapRel()
		if !ok {
			core.Sleep(80)
			continue
		}
		if !yxIsMiddleLayer(c, relY) {
			yxLog("%s: 对齐 pass=%d 不在中层 relY=%d 层=%s", tag, pass+1, relY, yxLayerName(c, relY))
			return false
		}
		if yxOnSideJumpAlignSpotOnMiddle(c, goRight, relX, relY, xMin, xMax, overshootTol) {
			yxLog("%s: 对齐就位 relX=%d x∈[%d,%d]", tag, relX, xMin, acceptMax)
			return true
		}
		moveTarget := yxSideJumpAlignMoveTarget(c, goRight, relX, targetX)
		walkMs := yxSideJumpAlignWalkMs(c, goRight, relX, moveTarget, targetX)
		walkOnly := yxSideJumpAlignWalkOnly(c, goRight, relX, targetX, xMin, xMax) || moveState.lastTeleportDir != 0
		yxAlignMoveTowardX(c, tag, pass+1, relX, moveTarget, &stuck, &moveState, walkOnly, walkMs)
		if _, relY, ok := readMinimapRel(); ok && !yxIsMiddleLayer(c, relY) {
			yxLog("%s: 对齐 pass=%d 后掉层 relY=%d 层=%s", tag, pass+1, relY, yxLayerName(c, relY))
			return false
		}
	}
	return false
}

func yxFineAlignToXRange(c *YexiongLingdiConfig, tag string, xMin, xMax int) bool {
	targetX := (xMin + xMax) / 2
	var stuck yxAlignStuckTracker
	for pass := 0; pass < c.RopeAlignMaxPass; pass++ {
		relX, _, ok := readMinimapRel()
		if !ok {
			core.Sleep(80)
			continue
		}
		if yxOnXRange(xMin, xMax, relX) {
			yxLog("%s: 对齐就位 relX=%d x∈[%d,%d]", tag, relX, xMin, xMax)
			return true
		}
		yxAlignMoveTowardX(c, tag, pass+1, relX, targetX, &stuck, nil, false, 0)
	}
	return false
}

func yxAscendToUpperSideFromMiddle(c *YexiongLingdiConfig, goRight bool) bool {
	tag := "右上"
	xMin, xMax := c.UpperRightJumpXMin, c.UpperRightJumpXMax
	checkOnPlatform := func(c *YexiongLingdiConfig, relX, relY int) bool {
		return yxIsUpperRightPlatform(c, relX, relY)
	}
	if !goRight {
		tag = "左上"
		xMin, xMax = c.UpperLeftJumpXMin, c.UpperLeftJumpXMax
		checkOnPlatform = func(c *YexiongLingdiConfig, relX, relY int) bool {
			return yxIsUpperLeftPlatform(c, relX, relY)
		}
	}
	maxAttempts := c.UpperSideClimbMaxAttempts
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		relX, relY, ok := readMinimapRel()
		if !ok {
			core.Sleep(100)
			continue
		}
		if checkOnPlatform(c, relX, relY) {
			yxLog("%s: 已在平台 relX=%d relY=%d", tag, relX, relY)
			return true
		}
		if yxIsLowerLayer(c, relY) {
			yxLog("%s: attempt=%d 在下层 relX=%d relY=%d 需先回中层", tag, attempt, relX, relY)
			return false
		}
		// 已在侧上高度：若 x 仍在跳点则补侧跳+瞬移，避免反复中层对齐
		if yxIsUpperSideLayerY(c, relY) {
			if checkOnPlatform(c, relX, relY) {
				return true
			}
			tol := yxSideJumpOvershootTol(c, goRight)
			if yxOnSideJumpAlignSpot(goRight, relX, xMin, xMax, tol) {
				yxLog("%s: attempt=%d 已在侧上y=%d 跳点x=%d 补侧跳+瞬移", tag, attempt, relY, relX)
				if goRight {
					yxSideRopeJumpOnly(true, tag+":补侧跳", relX)
				} else {
					yxSideRopeJumpOnly(false, tag+":补侧跳", relX)
				}
				core.Sleep(c.UpperSideArrivalSettleMs)
				yxTeleportOnly(goRight)
				relX, relY, settled := yxWaitYSettled(c, tag+":落地")
				if settled {
					yxLog("%s: 落地 relX=%d relY=%d 层=%s", tag, relX, relY, yxLayerName(c, relY))
					if checkOnPlatform(c, relX, relY) {
						return true
					}
					if yxIsLowerLayer(c, relY) {
						yxLog("%s: attempt=%d 补侧跳后在下层 失败", tag, attempt)
						return false
					}
				}
				yxLog("%s: attempt=%d 补侧跳后未到位", tag, attempt)
				return false
			}
			yxLog("%s: attempt=%d 已在侧上 relX=%d relY=%d 站位异常", tag, attempt, relX, relY)
			return false
		}
		if !yxIsMiddleLayer(c, relY) && !yxReadyForUpperSideSideJump(c, relY) {
			yxLog("%s: attempt=%d 站位异常 relX=%d relY=%d 层=%s", tag, attempt, relX, relY, yxLayerName(c, relY))
			return false
		}
		if attempt > 1 {
			yxUpperSideClimbRetryPrep(c, tag)
		}
		if !yxIsMiddleLayer(c, relY) {
			yxLog("%s: attempt=%d 不在中层 relY=%d 层=%s 跳过对齐", tag, attempt, relY, yxLayerName(c, relY))
			continue
		}
		if !yxFineAlignToSideJumpSpot(c, tag, goRight, xMin, xMax) {
			yxLog("%s: attempt=%d 未对齐 x∈[%d,%d]", tag, attempt, xMin, xMax)
			continue
		}
		relX, relY, ok = readMinimapRel()
		if !ok || !yxOnSideJumpAlignSpotOnMiddle(c, goRight, relX, relY, xMin, xMax, yxSideJumpOvershootTol(c, goRight)) {
			continue
		}
		if !yxIsMiddleLayer(c, relY) {
			yxLog("%s: attempt=%d 跳前不在中层 y=%d 层=%s", tag, attempt, relY, yxLayerName(c, relY))
			continue
		}
		if goRight {
			yxSideRopeJumpOnly(true, tag, relX)
		} else {
			yxSideRopeJumpOnly(false, tag, relX)
		}
		core.RandomSleep(c.RopePreUpWaitMinMs, c.RopePreUpWaitMaxMs)
		climbStart := time.Now()
		minClimbMs := time.Duration(c.UpperSideClimbMinMs) * time.Millisecond

		deadline := time.Now().Add(time.Duration(c.RopeClimbMaxSec) * time.Second)
		reached := false
		fellToLower := false
		for time.Now().Before(deadline) {
			refreshDpadUpHold(c.RopeUpHoldMs)
			core.Sleep(c.RopeClimbPollMs)
			curX, curY, ok := readMinimapRel()
			if !ok {
				continue
			}
			layer := yxLayerName(c, curY)
			if checkOnPlatform(c, curX, curY) {
				releaseDpadUp()
				yxLog("%s: 爬到平台 relX=%d relY=%d 层=%s", tag, curX, curY, layer)
				return true
			}
			if yxIsLowerLayer(c, curY) {
				yxLog("%s: 爬升掉到下层 y=%d 失败", tag, curY)
				fellToLower = true
				break
			}
			if time.Since(climbStart) >= minClimbMs && yxReadyForUpperSideSideJump(c, curY) {
				yxLog("%s: 爬绳%dms y=%d 层=%s 侧跳+瞬移", tag, c.UpperSideClimbMinMs, curY, layer)
				reached = true
				break
			}
		}
		releaseDpadUp()
		if fellToLower {
			return false
		}
		if !reached {
			yxLog("%s: attempt=%d 未爬到侧上绳区", tag, attempt)
			continue
		}
		if goRight {
			yxSideRopeJumpOnly(true, tag+":侧跳", relX)
		} else {
			yxSideRopeJumpOnly(false, tag+":侧跳", relX)
		}
		core.Sleep(c.UpperSideArrivalSettleMs)
		yxTeleportOnly(goRight)
		relX, relY, settled := yxWaitYSettled(c, tag+":落地")
		if settled {
			yxLog("%s: 落地 relX=%d relY=%d 层=%s", tag, relX, relY, yxLayerName(c, relY))
			if checkOnPlatform(c, relX, relY) {
				return true
			}
			if !goRight {
				yxLog("%s: 落地未进左上平台 x∈[%d,%d] y∈[%d,%d]", tag, c.UpperLeftXMin, c.UpperLeftXMax, c.UpperLeftYMin, c.UpperLeftYMax)
			} else {
				yxLog("%s: 落地未进右上平台 x∈[%d,%d] y∈[%d,%d]", tag, c.UpperRightXMin, c.UpperRightXMax, c.UpperRightYMin, c.UpperRightYMax)
			}
			if yxIsLowerLayer(c, relY) {
				yxLog("%s: attempt=%d 瞬移后在下层 失败", tag, attempt)
				return false
			}
		}
		yxLog("%s: attempt=%d 瞬移后未到位", tag, attempt)
	}
	return false
}

func yxDownJumpFromUpperSideToLower(c *YexiongLingdiConfig, fromRight bool) bool {
	tag := "左上"
	xMin, xMax := c.UpperLeftDescendXMin, c.UpperLeftDescendXMax
	if fromRight {
		tag = "右上"
		xMin, xMax = c.UpperRightDescendXMin, c.UpperRightDescendXMax
	}
	for attempt := 1; attempt <= c.DescendMaxRetry; attempt++ {
		relX, relY, ok := readMinimapRel()
		if !ok {
			core.Sleep(100)
			continue
		}
		if yxIsLowerLayer(c, relY) {
			yxLog("%s下跳: 已在下层 relX=%d relY=%d", tag, relX, relY)
			return true
		}
		if !yxIsUpperSideLayerY(c, relY) {
			yxLog("%s下跳: 站位异常 relX=%d relY=%d 层=%s", tag, relX, relY, yxLayerName(c, relY))
			core.Sleep(200)
			continue
		}
		if !yxOnXRange(xMin, xMax, relX) {
			if !yxFineAlignToXRange(c, tag+"下跳", xMin, xMax) {
				yxLog("%s下跳: attempt=%d 对齐 x∈[%d,%d] 失败", tag, attempt, xMin, xMax)
				continue
			}
			relX, relY, ok = readMinimapRel()
			if !ok || !yxOnXRange(xMin, xMax, relX) {
				continue
			}
		}
		yxLog("%s下跳: attempt=%d relX=%d relY=%d x∈[%d,%d] 下跳", tag, attempt, relX, relY, xMin, xMax)
		tapDownJump()
		landX, landY, settled := yxWaitYSettled(c, tag+"下跳")
		if !settled {
			yxLog("%s下跳: attempt=%d y未稳定", tag, attempt)
			continue
		}
		if yxIsLowerLayer(c, landY) {
			yxLog("%s下跳: 到达下层 relX=%d relY=%d", tag, landX, landY)
			return true
		}
		if yxIsUpperSideLayerY(c, landY) {
			yxLog("%s下跳: attempt=%d 落地仍在侧上 y=%d 重跳", tag, attempt, landY)
			continue
		}
		yxLog("%s下跳: attempt=%d 落地层=%s 重试", tag, attempt, yxLayerName(c, landY))
	}
	return false
}

func yxLowerRopeJumpTargetX(c *YexiongLingdiConfig, relX int) int {
	_ = relX
	return c.RopeJumpRightXMin
}

func yxOnLowerRopeJumpSpotExact(c *YexiongLingdiConfig, relX int) bool {
	return relX == c.RopeJumpRightXMin
}

func yxFineAlignToLowerRopeJumpSpot(c *YexiongLingdiConfig) bool {
	var stuck yxAlignStuckTracker
	for pass := 0; pass < c.RopeAlignMaxPass; pass++ {
		relX, _, ok := readMinimapRel()
		if !ok {
			core.Sleep(80)
			continue
		}
		if yxOnLowerRopeJumpSpotExact(c, relX) {
			yxLog("爬绳: 绳跳点就位 relX=%d (x=%d右跳)", relX, c.RopeJumpRightXMin)
			return true
		}
		targetX := yxLowerRopeJumpTargetX(c, relX)
		yxAlignMoveTowardX(c, "爬绳", pass+1, relX, targetX, &stuck, nil, false, 0)
	}
	return false
}

func yxLogBeforeRopeJump(tag string, goRight bool) {
	dir := "左"
	if goRight {
		dir = "右"
	}
	relX, relY, ok := readMinimapRel()
	if ok {
		yxLog("%s: 跳绳前 relX=%d relY=%d %s跳", tag, relX, relY, dir)
	} else {
		yxLog("%s: 跳绳前 读坐标失败 %s跳", tag, dir)
	}
}

func yxLowerRopeDirectionJump(c *YexiongLingdiConfig, goRight bool, relX int) {
	dir := "左"
	if goRight {
		dir = "右"
	}
	faceMs := c.RopeJumpFaceMinMs
	if c.RopeJumpFaceMaxMs > c.RopeJumpFaceMinMs {
		faceMs = c.RopeJumpFaceMinMs + rand.Intn(c.RopeJumpFaceMaxMs-c.RopeJumpFaceMinMs+1)
	}
	yxLog("爬绳: x=%d 先朝%s %dms", relX, dir, faceMs)
	walkHoldMs(goRight, faceMs)
	core.RandomSleep(c.RopeJumpBeforeMinMs, c.RopeJumpBeforeMaxMs)
	yxLogBeforeRopeJump("爬绳", goRight)
	if goRight {
		tapJumpRight()
	} else {
		tapJumpLeft()
	}
}

func yxLowerRopeGrabJump(c *YexiongLingdiConfig, relX int) {
	if relX == c.RopeJumpRightXMin {
		yxLowerRopeDirectionJump(c, true, relX)
	} else {
		yxLog("爬绳: x=%d 不在绳跳点 x=%d", relX, c.RopeJumpRightXMin)
	}
}

func yxAlignToRopeSpot(c *YexiongLingdiConfig) bool {
	targetX := c.ropeJumpRightCenter()
	if relX, relY, ok := readMinimapRel(); ok && yxOnRopeAlignZone(c, relX) {
		yxLog("爬绳: 对齐区就位 relX=%d relY=%d (目标x=%d±%d)", relX, relY, targetX, c.RopeAlignXTolerance)
		return true
	}
	var stuck yxAlignStuckTracker
	for pass := 0; pass < c.RopeAlignMaxPass; pass++ {
		relX, relY, ok := readMinimapRel()
		if !ok {
			core.Sleep(80)
			continue
		}
		if yxOnRopeAlignZone(c, relX) {
			yxLog("爬绳: 对齐区就位 relX=%d relY=%d (目标x=%d±%d)", relX, relY, targetX, c.RopeAlignXTolerance)
			return true
		}
		yxAlignMoveTowardX(c, "爬绳", pass+1, relX, targetX, &stuck, nil, false, 0)
	}
	yxLog("爬绳: 绳跳点对齐达最大轮次")
	return false
}

func yxOnRopeJumpSpot(c *YexiongLingdiConfig, relX int) bool {
	return yxOnRopeJumpRightSpot(c, relX) || yxOnRopeJumpLeftSpot(c, relX)
}

func yxRopeFaceBeforeJump(c *YexiongLingdiConfig, goRight bool) {
	ms := c.RopeJumpFaceMinMs
	if c.RopeJumpFaceMaxMs > c.RopeJumpFaceMinMs {
		ms = c.RopeJumpFaceMinMs + rand.Intn(c.RopeJumpFaceMaxMs-c.RopeJumpFaceMinMs+1)
	}
	dir := "左"
	if goRight {
		dir = "右"
	}
	yxLog("跳绳: 先朝%s %dms", dir, ms)
	walkHoldMs(goRight, ms)
}

func yxRopeDirectionJump(c *YexiongLingdiConfig, goRight bool, tag string, relX int) {
	_ = relX
	yxRopeFaceBeforeJump(c, goRight)
	yxLogBeforeRopeJump(tag, goRight)
	if goRight {
		tapJumpRight()
	} else {
		tapJumpLeft()
	}
}

// yxSideRopeJumpOnly 侧上绳跳：已在平台边缘，直接方向跳，不先按方向键（避免掉层）。
func yxSideRopeJumpOnly(goRight bool, tag string, relX int) {
	_ = relX
	yxLogBeforeRopeJump(tag, goRight)
	if goRight {
		tapJumpRight()
	} else {
		tapJumpLeft()
	}
}

func yxRopeJumpTargetX(c *YexiongLingdiConfig, relX int) int {
	if relX > c.RopeJumpRightXMax {
		return c.ropeJumpLeftCenter()
	}
	return c.ropeJumpRightCenter()
}

func yxFineAlignToRopeJumpSpot(c *YexiongLingdiConfig, tag string) bool {
	var stuck yxAlignStuckTracker
	for pass := 0; pass < c.RopeAlignMaxPass; pass++ {
		relX, _, ok := readMinimapRel()
		if !ok {
			core.Sleep(80)
			continue
		}
		if yxOnRopeJumpSpotReady(c, relX) {
			yxLog("%s: 绳跳点就位 relX=%d", tag, relX)
			return true
		}
		targetX := yxRopeJumpTargetX(c, relX)
		yxAlignMoveTowardX(c, tag, pass+1, relX, targetX, &stuck, nil, false, 0)
	}
	return false
}

func yxRopeGrabJump(c *YexiongLingdiConfig, relX int) {
	if !yxOnRopeJumpSpotReady(c, relX) {
		yxLog("爬绳: x=%d 不在绳跳点[%d,%d]/[%d,%d]", relX,
			c.RopeJumpRightXMin, c.RopeJumpRightXMax, c.RopeJumpLeftXMin, c.RopeJumpLeftXMax)
		return
	}
	if yxRopeJumpGoRight(c, relX) {
		yxRopeDirectionJump(c, true, "爬绳", relX)
	} else {
		yxRopeDirectionJump(c, false, "爬绳", relX)
	}
}

func yxRopeClimbRetryPrep(c *YexiongLingdiConfig) {
	yxLog("爬绳: 重试前攻击 %d~%dms", c.RopeClimbRetryAttackMinMs, c.RopeClimbRetryAttackMaxMs)
	keyHoldPress(attackKeyCode(), c.RopeClimbRetryAttackMinMs, c.RopeClimbRetryAttackMaxMs)
	core.Sleep(c.AlignPreWaitMs)
}

func yxClimbRopeToMiddle(c *YexiongLingdiConfig) bool {
	maxAttempts := c.RopeClimbMaxAttempts
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if _, curY, ok := readMinimapRel(); ok && yxIsMiddleLayer(c, curY) {
			yxLog("爬绳: 已在中层 y=%d", curY)
			return true
		}
		if attempt > 1 {
			yxRopeClimbRetryPrep(c)
		}
		if !yxFineAlignToLowerRopeJumpSpot(c) {
			yxLog("爬绳: attempt=%d 未站到绳跳点 x=%d", attempt, c.RopeJumpRightXMin)
			continue
		}
		relX, _, ok := readMinimapRel()
		if !ok || !yxOnLowerRopeJumpSpotExact(c, relX) {
			continue
		}
		yxLowerRopeGrabJump(c, relX)
		core.RandomSleep(c.RopePreUpWaitMinMs, c.RopePreUpWaitMaxMs)
		refreshDpadUpHold(c.RopeUpHoldMs)

		_, curY, ok := readMinimapRel()
		if !ok || yxIsLowerLayer(c, curY) {
			yxLog("爬绳: attempt=%d 按住上后仍在下层 y=%d 失败", attempt, curY)
			releaseDpadUp()
			continue
		}

		deadline := time.Now().Add(time.Duration(c.RopeClimbMaxSec) * time.Second)
		climbed := false
		for time.Now().Before(deadline) {
			refreshDpadUpHold(50)
			core.Sleep(c.RopeClimbPollMs)
			_, curY, ok := readMinimapRel()
			if !ok {
				continue
			}
			if yxIsMiddleLayer(c, curY) {
				yxLog("爬绳: 到中层 y=%d", curY)
				releaseDpadUp()
				if rand.Intn(2) == 0 {
					tapJumpLeft()
				} else {
					tapJumpRight()
				}
				core.Sleep(c.RopeArrivalSettleMs)
				climbed = true
				break
			}
			if yxIsLowerLayer(c, curY) {
				yxLog("爬绳: 爬升中掉回下层 y=%d", curY)
				releaseDpadUp()
				break
			}
		}
		if !climbed {
			releaseDpadUp()
			if _, curY, ok := readMinimapRel(); ok && yxIsMiddleLayer(c, curY) {
				yxLog("爬绳: 超时检测已在中层 y=%d", curY)
				return true
			}
			yxLog("爬绳: attempt=%d 未在%d秒内到中层", attempt, c.RopeClimbMaxSec)
			continue
		}
		return true
	}
	return false
}

func yxOnUpperRopeJumpRightSpot(c *YexiongLingdiConfig, relX int) bool {
	return matchRange(relX, c.UpperRopeJumpRightXMin, c.UpperRopeJumpRightXMax)
}

func yxOnUpperRopeJumpLeftSpot(c *YexiongLingdiConfig, relX int) bool {
	return matchRange(relX, c.UpperRopeJumpLeftXMin, c.UpperRopeJumpLeftXMax)
}

func yxOnUpperRopeJumpSpot(c *YexiongLingdiConfig, relX int) bool {
	return yxOnUpperRopeJumpRightSpot(c, relX) || yxOnUpperRopeJumpLeftSpot(c, relX)
}

func yxUpperRopeTargetX(c *YexiongLingdiConfig, relX int) int {
	rightC := c.upperRopeJumpRightCenter()
	leftC := c.upperRopeJumpLeftCenter()
	dRight := relX - rightC
	if dRight < 0 {
		dRight = -dRight
	}
	dLeft := relX - leftC
	if dLeft < 0 {
		dLeft = -dLeft
	}
	if dRight <= dLeft {
		return rightC
	}
	return leftC
}

func yxUpperSideClimbRetryPrep(c *YexiongLingdiConfig, tag string) {
	yxLog("%s: 重试前攻击 %dms", tag, c.UpperClimbRetryAttackMs)
	keyHoldPress(attackKeyCode(), c.UpperClimbRetryAttackMs, c.UpperClimbRetryAttackMs)
	core.Sleep(c.UpperClimbRetryWaitMs)
}

func yxUpperClimbRetryPrep(c *YexiongLingdiConfig) {
	yxUpperSideClimbRetryPrep(c, "爬上层")
}

func yxAlignToUpperRopeSpot(c *YexiongLingdiConfig) bool {
	if relX, relY, ok := readMinimapRel(); ok && yxOnUpperRopeJumpSpot(c, relX) {
		yxLog("爬上层: 绳跳点就位 relX=%d relY=%d", relX, relY)
		return true
	}
	var stuck yxAlignStuckTracker
	for pass := 0; pass < c.RopeAlignMaxPass; pass++ {
		relX, relY, ok := readMinimapRel()
		if !ok {
			core.Sleep(80)
			continue
		}
		if yxOnUpperRopeJumpSpot(c, relX) {
			yxLog("爬上层: 绳跳点就位 relX=%d relY=%d", relX, relY)
			return true
		}
		targetX := yxUpperRopeTargetX(c, relX)
		yxAlignMoveTowardX(c, "爬上层", pass+1, relX, targetX, &stuck, nil, false, 0)
	}
	yxLog("爬上层: 绳跳点对齐达最大轮次")
	return false
}

func yxUpperRopeGrabJump(c *YexiongLingdiConfig, relX int) {
	switch {
	case yxOnUpperRopeJumpRightSpot(c, relX):
		yxRopeDirectionJump(c, true, "爬上层", relX)
	case yxOnUpperRopeJumpLeftSpot(c, relX):
		yxRopeDirectionJump(c, false, "爬上层", relX)
	default:
		yxLog("爬上层: x=%d 不在绳跳点[%d,%d]/[%d,%d]", relX,
			c.UpperRopeJumpRightXMin, c.UpperRopeJumpRightXMax, c.UpperRopeJumpLeftXMin, c.UpperRopeJumpLeftXMax)
	}
}

func yxClimbRopeToUpper(c *YexiongLingdiConfig) bool {
	maxAttempts := c.UpperClimbMaxAttempts
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if attempt > 1 {
			yxUpperClimbRetryPrep(c)
		}
		if !yxAlignToUpperRopeSpot(c) {
			continue
		}
		relX, _, ok := readMinimapRel()
		if !ok || !yxOnUpperRopeJumpSpot(c, relX) {
			yxLog("爬上层: attempt=%d 未在绳跳点 relX=%d", attempt, relX)
			continue
		}
		yxUpperRopeGrabJump(c, relX)
		core.RandomSleep(c.RopePreUpWaitMinMs, c.RopePreUpWaitMaxMs)
		refreshDpadUpHold(c.RopeUpHoldMs)

		_, curY, ok := readMinimapRel()
		if !ok || yxIsMiddleLayer(c, curY) {
			yxLog("爬上层: attempt=%d 按住上后仍在中层 y=%d 失败", attempt, curY)
			releaseDpadUp()
			continue
		}

		deadline := time.Now().Add(time.Duration(c.RopeClimbMaxSec) * time.Second)
		climbed := false
		for time.Now().Before(deadline) {
			refreshDpadUpHold(50)
			core.Sleep(c.RopeClimbPollMs)
			_, curY, ok := readMinimapRel()
			if !ok {
				continue
			}
			if yxIsUpperClearPlatform(c, curY) {
				yxLog("爬上层: 到上平台 y=%d", curY)
				releaseDpadUp()
				core.Sleep(c.RopeArrivalSettleMs)
				climbed = true
				break
			}
			if yxIsMiddleLayer(c, curY) {
				yxLog("爬上层: 爬升中掉回中层 y=%d", curY)
				releaseDpadUp()
				break
			}
		}
		if !climbed {
			releaseDpadUp()
			if _, curY, ok := readMinimapRel(); ok && yxIsUpperClearPlatform(c, curY) {
				return true
			}
			yxLog("爬上层: attempt=%d 未在%d秒内到上平台", attempt, c.RopeClimbMaxSec)
			continue
		}
		if _, curY, ok := readMinimapRel(); ok && yxIsUpperClearPlatform(c, curY) {
			return true
		}
	}
	return false
}

var yxClearBagAlignStuck yxAlignStuckTracker

func yxAlignToClearBagSpot(c *YexiongLingdiConfig, relX int) bool {
	if yxOnClearBagX(c, relX) {
		yxClearBagAlignStuck.reset()
		return false
	}
	target := (c.ClearBagXMin + c.ClearBagXMax) / 2
	yxAlignMoveTowardX(c, "自动清包", 1, relX, target, &yxClearBagAlignStuck, nil, false, 0)
	return true
}

func tryAutoClearBagYexiong(s *autoClearBagState, c *YexiongLingdiConfig, relX, relY int) bool {
	if s.pendingShop {
		if !yxIsUpperClearPlatform(c, relY) {
			yxLog("自动清包: 已离开上平台 y=%d 放弃本次", relY)
			s.finishAttempt(yxLog)
			return true
		}
		return s.tryAutoShopSellMisc(yxLog, c.clearBagSellMisc())
	}
	if !s.due() {
		return false
	}
	if !yxIsUpperClearPlatform(c, relY) {
		return false
	}
	if !yxOnClearBagX(c, relX) {
		return yxAlignToClearBagSpot(c, relX)
	}
	yxLog("自动清包: 上平台 relX=%d relY=%d 卖货买药", relX, relY)
	return s.tryAutoShopSellMisc(yxLog, c.clearBagSellMisc())
}

func yxDescendFromUpperToMiddle(c *YexiongLingdiConfig) bool {
	for attempt := 1; attempt <= c.DescendMaxRetry; attempt++ {
		relX, relY, ok := readMinimapRel()
		if !ok {
			core.Sleep(100)
			continue
		}
		if yxIsMiddleLayer(c, relY) {
			yxLog("上平台下跳: 已在中层 relX=%d relY=%d", relX, relY)
			return true
		}
		if !yxIsUpperClearPlatform(c, relY) {
			yxLog("上平台下跳: 站位异常 relX=%d relY=%d 层=%s", relX, relY, yxLayerName(c, relY))
			core.Sleep(200)
			continue
		}
		if !yxOnXRange(c.UpperDescendXMin, c.UpperDescendXMax, relX) {
			if !yxFineAlignToXRange(c, "上平台下跳", c.UpperDescendXMin, c.UpperDescendXMax) {
				yxLog("上平台下跳: attempt=%d 对齐 x∈[%d,%d] 失败", attempt, c.UpperDescendXMin, c.UpperDescendXMax)
				continue
			}
			relX, relY, ok = readMinimapRel()
			if !ok || !yxOnXRange(c.UpperDescendXMin, c.UpperDescendXMax, relX) {
				continue
			}
		}
		yxLog("上平台下跳: attempt=%d relX=%d relY=%d x∈[%d,%d] 下跳", attempt, relX, relY, c.UpperDescendXMin, c.UpperDescendXMax)
		tapDownJump()
		landX, landY, settled := yxWaitYSettled(c, "上平台下跳")
		if !settled {
			yxLog("上平台下跳: attempt=%d y未稳定", attempt)
			continue
		}
		if yxIsMiddleLayer(c, landY) {
			yxLog("上平台下跳: 到达中层 relX=%d relY=%d", landX, landY)
			return true
		}
		if yxIsUpperClearPlatform(c, landY) {
			yxLog("上平台下跳: attempt=%d 仍在清包上 y=%d 重试", attempt, landY)
			continue
		}
		yxLog("上平台下跳: attempt=%d 落地层=%s 重试", attempt, yxLayerName(c, landY))
	}
	return false
}

// Play_野熊的领地 循环：中层→右上→下层→中层→左上→下层→中层；中层可爬绳清包。
func Play_野熊的领地(mapAssetPath string) error {
	cfg, err := loadMapConfig(mapAssetPath)
	if err != nil {
		return err
	}
	if cfg.YexiongLingdi == nil {
		return fmt.Errorf("野熊的领地: 缺少 yexiong_lingdi 配置")
	}
	c := cfg.YexiongLingdi
	c.normalize()
	if c.applyAPIConfigOverrides() {
		yxLog("API配置覆盖: 中层圈=%d~%d 下层圈=%d~%d 攻击=%d~%dms",
			c.MiddleLapsMin, c.MiddleLapsMax, c.LowerLapsMin, c.LowerLapsMax,
			c.AttackHoldMinMs, c.AttackHoldMaxMs)
	}
	c.applyDeleteYellowRegion()
	defer core.ClearMinimapYellowRegion()

	SetFarmLogTag(yexiongLingdiLogTag)
	StartFarmMaintainLoop(yexiongLingdiLogTag)
	defer StopFarmMaintainLoop()
	EnableFarmPeriodicLRJump()
	defer DisableFarmPeriodicLRJump()

	goRight := true
	faceRight()
	lastPatrolTpDir := 0
	var laps zzPatrolTracker
	clearBag := newAutoClearBagStateWithIntervalDefault(c.ClearBagIntervalMinMin, c.ClearBagIntervalMaxMin)
	cycleStep := yxCycleMiddleA
	phase := yxPhaseMiddleFarm
	laps.resetLaps(c.MiddleLapsMin, c.MiddleLapsMax)

	yxLog("开始挂机 循环: 中层→右上→下层→中层→左上→下层→中层")
	yxLog("  下层y=[%d,%d] 中层y=[%d,%d] 右上y=[%d,%d]x=[%d,%d] 左上y=[%d,%d]x=[%d,%d]",
		c.LowerYMin, c.LowerYMax, c.MiddleYMin, c.MiddleYMax,
		c.UpperRightYMin, c.UpperRightYMax, c.UpperRightXMin, c.UpperRightXMax,
		c.UpperLeftYMin, c.UpperLeftYMax, c.UpperLeftXMin, c.UpperLeftXMax)
	yxLog("  清包上平台y=[%d,%d]x=[%d,%d] 右上绳x=[%d,%d] 左上绳x=[%d,%d]",
		c.UpperYMin, c.UpperYMax, c.ClearBagXMin, c.ClearBagXMax,
		c.UpperRightJumpXMin, c.UpperRightJumpXMax, c.UpperLeftJumpXMin, c.UpperLeftJumpXMax)
	if core.API.GetConfigBoolValue("自动清包") {
		yxLog("自动清包: 上平台x=[%d,%d] 启动后优先清包 间隔%d~%d分钟", c.ClearBagXMin, c.ClearBagXMax, c.ClearBagIntervalMinMin, c.ClearBagIntervalMaxMin)
	}

	if relX, relY, ok := readMinimapRel(); ok {
		if clearBag.due() {
			switch {
			case yxIsUpperClearPlatform(c, relY):
				phase = yxPhaseUpperClearBag
				yxLog("启动在上平台 relX=%d relY=%d 优先清包", relX, relY)
			case yxIsMiddleLayer(c, relY):
				phase = yxPhaseGoUpper
				yxLog("启动在中层 relX=%d relY=%d 优先清包 去上平台", relX, relY)
			case yxIsLowerLayer(c, relY):
				phase = yxPhaseGoMiddle
				yxLog("启动在下层 relX=%d relY=%d 优先清包 先去中层", relX, relY)
			default:
				yxLog("启动站位异常 relX=%d relY=%d", relX, relY)
			}
		} else if yxIsUpperClearPlatform(c, relY) {
			phase = yxPhaseUpperClearBag
			yxLog("启动在上平台 relX=%d relY=%d", relX, relY)
		} else if yxIsUpperRightPlatform(c, relX, relY) {
			cycleStep = yxCycleUpperRight
			phase = yxPhaseUpperRightFarm
			laps.resetLaps(c.UpperSideLapsMin, c.UpperSideLapsMax)
			yxLog("启动在右上 relX=%d relY=%d", relX, relY)
		} else if yxIsUpperLeftPlatform(c, relX, relY) {
			cycleStep = yxCycleUpperLeft
			phase = yxPhaseUpperLeftFarm
			laps.resetLaps(c.UpperSideLapsMin, c.UpperSideLapsMax)
			yxLog("启动在左上 relX=%d relY=%d", relX, relY)
		} else if yxIsMiddleLayer(c, relY) {
			cycleStep = yxCycleMiddleA
			phase = yxPhaseMiddleFarm
			laps.resetLaps(c.MiddleLapsMin, c.MiddleLapsMax)
			yxLog("启动在中层 relX=%d relY=%d 先刷中层", relX, relY)
		} else if yxIsLowerLayer(c, relY) {
			phase = yxPhaseGoMiddle
			yxLog("启动在下层 relX=%d relY=%d 先去中层", relX, relY)
		}
	}

	for {
		core.BlockWhileCaptchaHold()
		TickFarmMainThreadTasks()

		switch phase {
		case yxPhaseGoMiddle:
			if relX, relY, ok := readMinimapRel(); ok && yxIsMiddleLayer(c, relY) {
				if clearBag.due() {
					phase = yxPhaseGoUpper
					yxLog("已在中层 优先清包 去上平台")
					continue
				}
				phase = yxPhaseMiddleFarm
				goRight = true
				faceRight()
				laps.resetLaps(c.MiddleLapsMin, c.MiddleLapsMax)
				yxLog("已在中层 relX=%d relY=%d 刷怪 %d圈", relX, relY, laps.targetLaps)
				continue
			}
			if yxClimbRopeToMiddle(c) {
				if clearBag.due() {
					phase = yxPhaseGoUpper
					yxLog("到中层 优先清包 去上平台")
				} else {
					phase = yxPhaseMiddleFarm
					goRight = true
					faceRight()
					laps.resetLaps(c.MiddleLapsMin, c.MiddleLapsMax)
					yxLog("到中层 刷怪 %d圈", laps.targetLaps)
				}
			} else {
				phase = yxPhaseLowerFarm
				laps.resetLaps(c.LowerLapsMin, c.LowerLapsMax)
				yxLog("爬绳失败 继续下层刷怪")
			}
			continue
		case yxPhaseGoUpperRight:
			if relX, relY, ok := readMinimapRel(); ok && yxIsUpperRightPlatform(c, relX, relY) {
				cycleStep = yxCycleUpperRight
				phase = yxPhaseUpperRightFarm
				goRight = true
				faceRight()
				laps.resetLaps(c.UpperSideLapsMin, c.UpperSideLapsMax)
				yxLog("已在右上 relX=%d relY=%d 刷怪 %d圈", relX, relY, laps.targetLaps)
				continue
			}
			if yxAscendToUpperSideFromMiddle(c, true) {
				cycleStep = yxCycleUpperRight
				phase = yxPhaseUpperRightFarm
				goRight = true
				faceRight()
				laps.resetLaps(c.UpperSideLapsMin, c.UpperSideLapsMax)
				yxLog("到右上 刷怪 %d圈", laps.targetLaps)
			} else if relX, relY, ok := readMinimapRel(); ok && yxIsLowerLayer(c, relY) {
				phase = yxPhaseGoMiddle
				yxLog("去右上失败且在下层 relX=%d relY=%d 先回中层", relX, relY)
			} else {
				phase = yxPhaseMiddleFarm
				laps.resetLaps(c.MiddleLapsMin, c.MiddleLapsMax)
				yxLog("去右上失败 继续中层刷怪")
			}
			continue
		case yxPhaseGoUpperLeft:
			if relX, relY, ok := readMinimapRel(); ok && yxIsUpperLeftPlatform(c, relX, relY) {
				cycleStep = yxCycleUpperLeft
				phase = yxPhaseUpperLeftFarm
				goRight = false
				faceLeft()
				laps.resetLaps(c.UpperSideLapsMin, c.UpperSideLapsMax)
				yxLog("已在左上 relX=%d relY=%d 刷怪 %d圈", relX, relY, laps.targetLaps)
				continue
			}
			if yxAscendToUpperSideFromMiddle(c, false) {
				cycleStep = yxCycleUpperLeft
				phase = yxPhaseUpperLeftFarm
				goRight = false
				faceLeft()
				laps.resetLaps(c.UpperSideLapsMin, c.UpperSideLapsMax)
				yxLog("到左上 刷怪 %d圈", laps.targetLaps)
			} else if relX, relY, ok := readMinimapRel(); ok && yxIsLowerLayer(c, relY) {
				phase = yxPhaseGoMiddle
				yxLog("去左上失败且在下层 relX=%d relY=%d 先回中层", relX, relY)
			} else if relX, relY, ok := readMinimapRel(); ok && yxIsUpperLeftPlatform(c, relX, relY) {
				cycleStep = yxCycleUpperLeft
				phase = yxPhaseUpperLeftFarm
				goRight = false
				faceLeft()
				laps.resetLaps(c.UpperSideLapsMin, c.UpperSideLapsMax)
				yxLog("去左上: 已在平台 relX=%d relY=%d 刷怪 %d圈", relX, relY, laps.targetLaps)
			} else {
				phase = yxPhaseMiddleFarm
				laps.resetLaps(c.MiddleLapsMin, c.MiddleLapsMax)
				yxLog("去左上失败 继续中层刷怪")
			}
			continue
		case yxPhaseGoLowerFromSide:
			if relX, relY, ok := readMinimapRel(); ok && yxIsLowerLayer(c, relY) {
				phase = yxPhaseLowerFarm
				goRight = true
				faceRight()
				laps.resetLaps(c.LowerLapsMin, c.LowerLapsMax)
				if cycleStep == yxCycleUpperRight {
					cycleStep = yxCycleLowerA
				} else if cycleStep == yxCycleUpperLeft {
					cycleStep = yxCycleLowerB
				}
				yxLog("已在下层 relX=%d relY=%d 刷怪 %d圈", relX, relY, laps.targetLaps)
				continue
			}
			fromRight := cycleStep == yxCycleUpperRight
			if yxDownJumpFromUpperSideToLower(c, fromRight) {
				phase = yxPhaseLowerFarm
				goRight = true
				faceRight()
				laps.resetLaps(c.LowerLapsMin, c.LowerLapsMax)
				if cycleStep == yxCycleUpperRight {
					cycleStep = yxCycleLowerA
				} else if cycleStep == yxCycleUpperLeft {
					cycleStep = yxCycleLowerB
				}
				yxLog("到下层 刷怪 %d圈", laps.targetLaps)
			} else if cycleStep == yxCycleUpperRight {
				phase = yxPhaseUpperRightFarm
				laps.resetLaps(c.UpperSideLapsMin, c.UpperSideLapsMax)
				yxLog("右上跳下失败 继续右上刷怪")
			} else {
				phase = yxPhaseUpperLeftFarm
				laps.resetLaps(c.UpperSideLapsMin, c.UpperSideLapsMax)
				yxLog("左上跳下失败 继续左上刷怪")
			}
			continue
		case yxPhaseGoUpper:
			if relX, relY, ok := readMinimapRel(); ok && yxIsUpperClearPlatform(c, relY) {
				phase = yxPhaseUpperClearBag
				yxLog("已在上平台 relX=%d relY=%d 清包", relX, relY)
				continue
			}
			if yxClimbRopeToUpper(c) {
				phase = yxPhaseUpperClearBag
				yxLog("到上平台 清包")
			} else {
				phase = yxPhaseMiddleFarm
				yxLog("爬上层失败 继续中层刷怪")
			}
			continue
		case yxPhaseUpperClearBag:
			relX, relY, ok := readMinimapRel()
			if !ok {
				core.Sleep(100)
				continue
			}
			if !yxIsUpperClearPlatform(c, relY) {
				if yxIsMiddleLayer(c, relY) {
					yxLog("清包阶段已在中层 继续刷怪")
					phase = yxPhaseMiddleFarm
					laps.resetLaps(c.MiddleLapsMin, c.MiddleLapsMax)
				} else if yxTryRecoverRopeBetweenLayer("清包", c, relX, relY) {
				} else {
					yxLog("清包阶段站位异常 relX=%d relY=%d", relX, relY)
					core.Sleep(200)
				}
				continue
			}
			if tryAutoClearBagYexiong(&clearBag, c, relX, relY) {
				continue
			}
			if clearBag.due() || clearBag.pendingShop {
				continue
			}
			if yxDescendFromUpperToMiddle(c) {
				phase = yxPhaseMiddleFarm
				goRight = true
				faceRight()
				laps.resetLaps(c.MiddleLapsMin, c.MiddleLapsMax)
				yxLog("清包完成 回中层刷怪")
			} else {
				yxLog("上平台下跳回中层失败")
				core.Sleep(200)
			}
			continue
		}

		relX, relY, ok := readMinimapRel()
		if !ok {
			core.Sleep(100)
			continue
		}

		if phase == yxPhaseLowerFarm {
			if !yxIsLowerLayer(c, relY) {
				if yxIsMiddleLayer(c, relY) {
					yxLog("下层阶段但已在中层 转中层刷怪")
					phase = yxPhaseMiddleFarm
					laps.resetLaps(c.MiddleLapsMin, c.MiddleLapsMax)
				} else if yxTryRecoverRopeBetweenLayer("下层", c, relX, relY) {
				} else {
					yxLog("下层阶段站位异常 relX=%d relY=%d", relX, relY)
					core.Sleep(200)
				}
				continue
			}
			if laps.finished() {
				switch cycleStep {
				case yxCycleLowerA:
					cycleStep = yxCycleMiddleB
					yxLog("下层刷够 %d圈 去中层B", laps.targetLaps)
				case yxCycleLowerB:
					cycleStep = yxCycleMiddleC
					yxLog("下层刷够 %d圈 去中层C", laps.targetLaps)
				default:
					yxLog("下层刷够 %d圈 去中层", laps.targetLaps)
				}
				phase = yxPhaseGoMiddle
				continue
			}
			yxPatrolFarmStep("下层", c, relX, c.LowerPatrolTurnXLeft, c.LowerPatrolTurnXRight, &goRight, &laps, false, nil)
			continue
		}

		if phase == yxPhaseMiddleFarm {
			if !yxIsMiddleLayer(c, relY) {
				if yxIsLowerLayer(c, relY) {
					yxLog("中层阶段但已在下层 爬绳回中层")
					phase = yxPhaseGoMiddle
				} else if yxIsUpperLeftPlatform(c, relX, relY) {
					cycleStep = yxCycleUpperLeft
					phase = yxPhaseUpperLeftFarm
					goRight = false
					faceLeft()
					laps.resetLaps(c.UpperSideLapsMin, c.UpperSideLapsMax)
					yxLog("中层阶段但已在左上 relX=%d relY=%d 改刷左上", relX, relY)
				} else if yxIsUpperRightPlatform(c, relX, relY) {
					cycleStep = yxCycleUpperRight
					phase = yxPhaseUpperRightFarm
					goRight = true
					faceRight()
					laps.resetLaps(c.UpperSideLapsMin, c.UpperSideLapsMax)
					yxLog("中层阶段但已在右上 relX=%d relY=%d 改刷右上", relX, relY)
				} else if yxIsUpperClearPlatform(c, relY) {
					yxLog("中层阶段但在上平台 relX=%d relY=%d 下跳回中层", relX, relY)
					if !yxDescendFromUpperToMiddle(c) {
						yxLog("中层阶段上平台下跳失败")
						core.Sleep(200)
					}
				} else if yxTryRecoverRopeBetweenLayer("中层", c, relX, relY) {
				} else {
					yxLog("中层阶段站位异常 relX=%d relY=%d", relX, relY)
					core.Sleep(200)
				}
				continue
			}
			if clearBag.due() {
				yxLog("中层: 到清包间隔 去上平台")
				phase = yxPhaseGoUpper
				continue
			}
			if laps.finished() {
				switch cycleStep {
				case yxCycleMiddleA:
					yxLog("中层A刷够 %d圈 去右上", laps.targetLaps)
					phase = yxPhaseGoUpperRight
				case yxCycleMiddleB:
					yxLog("中层B刷够 %d圈 去左上", laps.targetLaps)
					phase = yxPhaseGoUpperLeft
				case yxCycleMiddleC:
					yxLog("完成一轮")
					cycleStep = yxCycleMiddleA
					laps.resetLaps(c.MiddleLapsMin, c.MiddleLapsMax)
					yxLog("新一轮 中层A刷怪 %d圈", laps.targetLaps)
				default:
					yxLog("中层刷够 %d圈 去右上", laps.targetLaps)
					cycleStep = yxCycleMiddleA
					phase = yxPhaseGoUpperRight
				}
				continue
			}
			yxPatrolFarmStep("中层", c, relX, c.MiddlePatrolTurnXLeft, c.MiddlePatrolTurnXRight, &goRight, &laps, true, &lastPatrolTpDir)
			continue
		}

		if phase == yxPhaseUpperRightFarm {
			if !yxIsUpperSideLayerY(c, relY) {
				if yxIsMiddleLayer(c, relY) {
					yxLog("右上阶段但已在中层 转中层刷怪")
					phase = yxPhaseMiddleFarm
					laps.resetLaps(c.MiddleLapsMin, c.MiddleLapsMax)
				} else if yxIsLowerLayer(c, relY) {
					yxLog("右上阶段但已在下层 先回中层")
					phase = yxPhaseGoMiddle
				} else if yxTryRecoverRopeBetweenLayer("右上", c, relX, relY) {
				} else {
					yxLog("右上阶段站位异常 relX=%d relY=%d 层=%s", relX, relY, yxLayerName(c, relY))
					core.Sleep(200)
				}
				continue
			}
			if laps.finished() {
				yxLog("右上刷够 %d圈 下跳下层", laps.targetLaps)
				phase = yxPhaseGoLowerFromSide
				continue
			}
			yxPatrolFarmStep("右上", c, relX, c.UpperRightPatrolTurnXLeft, c.UpperRightPatrolTurnXRight, &goRight, &laps, false, nil)
			continue
		}

		if phase == yxPhaseUpperLeftFarm {
			if !yxIsUpperSideLayerY(c, relY) {
				if yxIsMiddleLayer(c, relY) {
					yxLog("左上阶段但已在中层 转中层刷怪")
					phase = yxPhaseMiddleFarm
					laps.resetLaps(c.MiddleLapsMin, c.MiddleLapsMax)
				} else if yxIsLowerLayer(c, relY) {
					yxLog("左上阶段但已在下层 先回中层")
					phase = yxPhaseGoMiddle
				} else if yxTryRecoverRopeBetweenLayer("左上", c, relX, relY) {
				} else {
					yxLog("左上阶段站位异常 relX=%d relY=%d 层=%s", relX, relY, yxLayerName(c, relY))
					core.Sleep(200)
				}
				continue
			}
			if laps.finished() {
				yxLog("左上刷够 %d圈 下跳下层", laps.targetLaps)
				phase = yxPhaseGoLowerFromSide
				continue
			}
			yxPatrolFarmStep("左上", c, relX, c.UpperLeftPatrolTurnXLeft, c.UpperLeftPatrolTurnXRight, &goRight, &laps, false, nil)
		}
	}
}
