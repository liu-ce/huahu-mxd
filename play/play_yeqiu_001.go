package play

import (
	"app/core"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

const yeqiu001LogTag = "[叶秋001]"

var yeqiuActiveLogTag = yeqiu001LogTag
var yeqiuActiveNoTeleport bool

type yeqiuPatrolStepFunc func(tag string, y *Yeqiu001Config, relX, xMin, xMax int, goRight *bool, farmSince time.Time)

func yeqiuLog(format string, args ...interface{}) {
	fmt.Printf(yeqiuActiveLogTag+" "+format+"\n", args...)
}

func yeqiuRandSec(minS, maxS int) time.Duration {
	span := maxS - minS
	sec := minS
	if span > 0 {
		sec += rand.Intn(span + 1)
	}
	return time.Duration(sec) * time.Second
}

func yeqiuToastPos(y *Yeqiu001Config, relX, relY int) {
	if y == nil || !y.ShowPosToast {
		return
	}
	ms := y.PosToastMs
	if ms <= 0 {
		ms = 500
	}
	//utils.Toast(strconv.Itoa(relX)+"  "+strconv.Itoa(relY), 0, 0, ms)
}

func yeqiuReadMinimapRel(y *Yeqiu001Config) (relX, relY int, ok bool) {
	relX, relY, ok = readMinimapRel()
	if ok {
		yeqiuToastPos(y, relX, relY)
	}
	return relX, relY, ok
}

func yeqiuIsSmallPlatform(y *Yeqiu001Config, relX, relY int) bool {
	return matchRange(relY, y.SmallPlatformYMin, y.SmallPlatformYMax) &&
		matchRange(relX, y.smallPlatformXMin(), y.smallPlatformXMax())
}

func yeqiuIsLowerLayerY(y *Yeqiu001Config, relY int) bool {
	return matchRange(relY, y.LowerYMin, y.LowerYMax)
}

func yeqiuIsUpperLayerY(y *Yeqiu001Config, relY int) bool {
	return matchRange(relY, y.UpperYMin, y.UpperYMax)
}

func yeqiuInLowerPatrolX(y *Yeqiu001Config, relX int) bool {
	return matchRange(relX, y.LowerXMin, y.LowerXMax)
}

func yeqiuInUpperPatrolX(y *Yeqiu001Config, relX int) bool {
	return matchRange(relX, y.UpperXMin, y.UpperXMax)
}

func yeqiuPatrolMargin(y *Yeqiu001Config) int {
	if y == nil || y.PatrolTeleportMargin <= 0 {
		return 12
	}
	return y.PatrolTeleportMargin
}

// yeqiuPatrolEffectiveMargin 折返区不可重叠，否则会在中间位置只改方向不瞬移。
func yeqiuPatrolEffectiveMargin(xMin, xMax, margin int) int {
	span := xMax - xMin
	if span < 4 {
		return 0
	}
	maxMargin := (span - 2) / 2
	if maxMargin < 1 {
		return 0
	}
	if margin <= 0 || margin > maxMargin {
		return maxMargin
	}
	return margin
}

// yeqiuPatrolFarmStep 打怪区往返：出界只朝区内瞬移；近边界提前折返，避免左右来回过冲。
func yeqiuPatrolFarmStep(tag string, y *Yeqiu001Config, relX, xMin, xMax int, goRight *bool, farmSince time.Time) {
	margin := yeqiuPatrolEffectiveMargin(xMin, xMax, yeqiuPatrolMargin(y))

	if relX < xMin {
		*goRight = true
		yeqiuLog("%s: x=%d 超出左界%d 右瞬移回区", tag, relX, xMin)
		tapTeleportWithDirection(true)
		core.Sleep(50)
		return
	}
	if relX > xMax {
		*goRight = false
		yeqiuLog("%s: x=%d 超出右界%d 左瞬移回区", tag, relX, xMax)
		tapTeleportWithDirection(false)
		core.Sleep(50)
		return
	}

	if *goRight && relX >= xMax-margin {
		*goRight = false
		yeqiuLog("%s: 近右界 relX=%d 改向左", tag, relX)
		core.Sleep(50)
		return
	}
	if !*goRight && relX <= xMin+margin {
		*goRight = true
		yeqiuLog("%s: 近左界 relX=%d 改向右", tag, relX)
		core.Sleep(50)
		return
	}

	dir := "左"
	if *goRight {
		dir = "右"
	}
	if patrolFarmAllowWalk(farmSince) {
		yeqiuLog("%s: %s走+攻击 relX=%d", tag, dir, relX)
		patrolFarmWalkAndAttack(*goRight, keyHoldShortMin, keyHoldShortMax)
		return
	}
	yeqiuLog("%s: %s瞬移+攻击 relX=%d", tag, dir, relX)
	yeqiuTeleportAndAttack(*goRight)
}

func yeqiuIsKnownLayer(y *Yeqiu001Config, relX, relY int) bool {
	return yeqiuIsUpperLayerY(y, relY) || yeqiuIsLowerLayerY(y, relY) || yeqiuIsSmallPlatform(y, relX, relY)
}

type yeqiuUnknownLayerTracker struct {
	since  time.Time
	active bool
}

func (t *yeqiuUnknownLayerTracker) tryJumpLeft(tag string, y *Yeqiu001Config, relX, relY int) bool {
	if yeqiuIsKnownLayer(y, relX, relY) {
		t.active = false
		return false
	}
	sec := 3
	if y != nil && y.UnknownLayerJumpSec > 0 {
		sec = y.UnknownLayerJumpSec
	}
	if !t.active {
		t.active = true
		t.since = time.Now()
		return false
	}
	if time.Since(t.since) < time.Duration(sec)*time.Second {
		return false
	}
	yeqiuLog("%s: 中间层 relX=%d relY=%d %ds 左跳", tag, relX, relY, sec)
	tapJumpLeft()
	core.Sleep(200)
	t.active = false
	return true
}

func yeqiuOnRopeSpot(y *Yeqiu001Config, relX int) bool {
	return matchRange(relX, y.ropeJumpXMin(), y.ropeJumpXMax())
}

func yeqiuRopeDist(y *Yeqiu001Config, relX int) int {
	d := relX - y.RopeX
	if d < 0 {
		d = -d
	}
	return d
}

func yeqiuRopeNearStepMs(y *Yeqiu001Config) int {
	if y == nil || y.RopeNearStepMs <= 0 {
		return 120
	}
	return y.RopeNearStepMs
}

func yeqiuRopeMicroStepMs(y *Yeqiu001Config) int {
	if y == nil || y.RopeNearMicroStepMs <= 0 {
		return 65
	}
	return y.RopeNearMicroStepMs
}

func yeqiuRopeAlignStepMs(y *Yeqiu001Config, dist int) int {
	if dist <= 1 {
		return yeqiuRopeMicroStepMs(y)
	}
	if dist <= 2 {
		ms := yeqiuRopeMicroStepMs(y) + 25
		if ms > yeqiuRopeNearStepMs(y) {
			return yeqiuRopeNearStepMs(y)
		}
		return ms
	}
	return yeqiuRopeNearStepMs(y)
}

func yeqiuDownJumpCheckMs(y *Yeqiu001Config) int {
	if y == nil || y.DownJumpCheckMs <= 0 {
		return 200
	}
	return y.DownJumpCheckMs
}

func yeqiuDownJumpMaxRetry(y *Yeqiu001Config) int {
	if y == nil || y.DownJumpMaxRetry <= 0 {
		return 5
	}
	return y.DownJumpMaxRetry
}

// yeqiuTryDownJumpLeaveUpper 上层下跳；200ms 后 y 仍<=upper_y_max 则重跳。
func yeqiuTryDownJumpLeaveUpper(tag string, y *Yeqiu001Config, relX, relY int) bool {
	if !yeqiuIsUpperLayerY(y, relY) {
		return false
	}
	waitMs := yeqiuDownJumpCheckMs(y)
	for attempt := 1; attempt <= yeqiuDownJumpMaxRetry(y); attempt++ {
		yeqiuLog("%s: 上层 relX=%d relY=%d 下跳 attempt=%d", tag, relX, relY, attempt)
		tapDownJump()
		core.Sleep(waitMs)
		curX, curY, ok := yeqiuReadMinimapRel(y)
		if !ok {
			return true
		}
		if yeqiuIsLowerLayerY(y, curY) {
			yeqiuLog("%s: 已到下层 relX=%d relY=%d", tag, curX, curY)
			return true
		}
		if !yeqiuIsUpperLayerY(y, curY) {
			return true
		}
		yeqiuLog("%s: 下跳失败 y=%d 仍在上层 重跳", tag, curY)
		relY = curY
	}
	return true
}

func yeqiuTrySmallPlatformDown(tag string, y *Yeqiu001Config, relX, relY int) bool {
	if !yeqiuIsSmallPlatform(y, relX, relY) {
		return false
	}
	waitMs := yeqiuDownJumpCheckMs(y)
	maxAttempt := 5
	if y.DownJumpMaxRetry > 0 {
		maxAttempt = y.DownJumpMaxRetry
	}
	for attempt := 1; attempt <= maxAttempt; attempt++ {
		yeqiuLog("%s: 小站台 relX=%d relY=%d 下跳 attempt=%d", tag, relX, relY, attempt)
		tapDownJump()
		core.Sleep(waitMs)
		curX, curY, ok := yeqiuReadMinimapRel(y)
		if ok && !yeqiuIsSmallPlatform(y, curX, curY) {
			yeqiuLog("%s: 小站台下跳成功 y=%d", tag, curY)
			return true
		}
		if ok {
			yeqiuLog("%s: 小站台下跳失败 y=%d 仍在小站台 重跳", tag, curY)
		}
	}
	return true
}

// yeqiuHandleSmallPlatform 小站台：到清包间隔则卖货（含买药），否则下跳离开。
func yeqiuHandleSmallPlatform(s *autoClearBagState, tag string, y *Yeqiu001Config, relX, relY int) bool {
	if !yeqiuIsSmallPlatform(y, relX, relY) {
		return false
	}
	if s != nil && s.due() {
		yeqiuLog("%s: 小站台 relX=%d relY=%d 到达清包间隔 自动卖货", tag, relX, relY)
		s.tryAutoShop(yeqiuLog)
		return true
	}
	return yeqiuTrySmallPlatformDown(tag, y, relX, relY)
}

func yeqiuTeleportAndAttack(goRight bool) {
	teleportAndAttack(goRight)
}

func yeqiuAlignToRope(tag string, y *Yeqiu001Config) bool {
	ropeXTol := y.RopeXTolerance
	if ropeXTol <= 0 {
		ropeXTol = 3
	}
	for pass := 0; pass < y.AlignMaxPass; pass++ {
		relX, relY, ok := yeqiuReadMinimapRel(y)
		if !ok {
			core.Sleep(80)
			continue
		}
		dist := yeqiuRopeDist(y, relX)
		if yeqiuOnRopeSpot(y, relX) {
			yeqiuLog("%s: 绳子跳点就位 relX=%d relY=%d (跳抓x=[%d,%d])", tag, relX, relY, y.ropeJumpXMin(), y.ropeJumpXMax())
			return true
		}

		goRight := relX < y.RopeX
		dir := "左"
		if goRight {
			dir = "右"
		}
		switch {
		case dist <= ropeXTol:
			stepMs := yeqiuRopeAlignStepMs(y, dist)
			yeqiuLog("%s: 绳子对齐 pass=%d relX=%d→%d 微调%s %dms", tag, pass+1, relX, y.RopeX, dir, stepMs)
			walkHoldMs(goRight, stepMs)
		case dist <= y.RopeNearWalkDist:
			yeqiuLog("%s: 绳子对齐 pass=%d relX=%d→%d 慢走%s %dms", tag, pass+1, relX, y.RopeX, dir, y.RopeNearWalkMs)
			walkHoldMs(goRight, y.RopeNearWalkMs)
		default:
			if yeqiuActiveNoTeleport {
				ms := y.RopeNearWalkMs
				if ms <= 0 {
					ms = 300
				}
				yeqiuLog("%s: 绳子对齐 pass=%d relX=%d→%d 快走%s %dms", tag, pass+1, relX, y.RopeX, dir, ms)
				walkHoldMs(goRight, ms)
			} else {
				yeqiuLog("%s: 绳子对齐 pass=%d relX=%d→%d 瞬移%s", tag, pass+1, relX, y.RopeX, dir)
				tapTeleportWithDirection(goRight)
			}
		}
		if dist > ropeXTol {
			tapAttackOnce()
			yeqiuWaitAfterAttackBeforeJump(y)
		} else {
			core.Sleep(200)
		}
	}
	yeqiuLog("%s: 绳子对齐达最大轮次", tag)
	return false
}

func yeqiuWaitAfterAttackBeforeJump(y *Yeqiu001Config) {
	minMs := 300
	maxMs := 500
	if y != nil {
		if y.ClimbAttackJumpWaitMsMin > 0 {
			minMs = y.ClimbAttackJumpWaitMsMin
		}
		if y.ClimbAttackJumpWaitMsMax >= minMs {
			maxMs = y.ClimbAttackJumpWaitMsMax
		}
	}
	core.RandomSleep(minMs, maxMs)
}

func yeqiuStartClimbGrab(tag string, y *Yeqiu001Config, relX int) {
	waitMs := y.ClimbJumpWaitMs
	if waitMs <= 0 {
		waitMs = 200
	}
	yeqiuLog("%s: 爬绳 x=%d 跳→上 等%dms", tag, relX, waitMs)
	tapJump()
	core.Sleep(waitMs)
	tapMs := y.ClimbUpTapMs
	if tapMs <= 0 {
		tapMs = 300
	}
	refreshDpadUpHold(tapMs)
}

func yeqiuClimbReachedTop(y *Yeqiu001Config, curY int) bool {
	return yeqiuIsUpperLayerY(y, curY) || curY <= y.UpperYMax
}

func yeqiuClimbLowerRetryMs(y *Yeqiu001Config) int {
	if y == nil || y.ClimbLowerRetryMs <= 0 {
		return 500
	}
	return y.ClimbLowerRetryMs
}

func yeqiuClimbStillOnLower(y *Yeqiu001Config, minY, curY int, sinceGrab time.Duration) bool {
	if !yeqiuIsLowerLayerY(y, curY) {
		return false
	}
	if minY < y.LowerYMin {
		return true
	}
	return sinceGrab >= time.Duration(yeqiuClimbLowerRetryMs(y))*time.Millisecond
}

func yeqiuReleaseUp() {
	releaseDpadUp()
}

func yeqiuClimbRope(tag string, y *Yeqiu001Config) bool {
	maxSec := y.ClimbMaxSec
	if maxSec <= 0 {
		maxSec = 25
	}
	for attempt := 0; attempt < 5; attempt++ {
		yeqiuLog("%s: 对齐前攻击清怪 attempt=%d", tag, attempt+1)
		tapAttackOnce()
		yeqiuWaitAfterAttackBeforeJump(y)
		if !yeqiuAlignToRope(tag, y) {
			return false
		}
		relX, startY, ok := yeqiuReadMinimapRel(y)
		if !ok {
			continue
		}
		yeqiuWaitAfterAttackBeforeJump(y)
		yeqiuStartClimbGrab(tag, y, relX)

		minY := startY
		climbed := false
		failedOnLower := false
		grabAt := time.Now()
		deadline := time.Now().Add(time.Duration(maxSec) * time.Second)
		for time.Now().Before(deadline) {
			refreshDpadUpHold(50)
			_, curY, ok := yeqiuReadMinimapRel(y)
			if !ok {
				continue
			}
			if curY < minY {
				minY = curY
			}
			if yeqiuClimbStillOnLower(y, minY, curY, time.Since(grabAt)) {
				yeqiuLog("%s: 爬绳 y=%d 仍在下层 minY=%d 重爬 attempt=%d", tag, curY, minY, attempt+1)
				yeqiuReleaseUp()
				failedOnLower = true
				break
			}
			if yeqiuClimbReachedTop(y, curY) {
				yeqiuLog("%s: 爬绳到顶 y=%d 左跳", tag, curY)
				yeqiuReleaseUp()
				tapJumpLeft()
				core.Sleep(200)
				climbed = true
				break
			}
		}
		if !climbed {
			yeqiuReleaseUp()
			if _, relY, ok := yeqiuReadMinimapRel(y); ok && yeqiuIsUpperLayerY(y, relY) {
				yeqiuLog("%s: 爬绳超时但已在上层 y=%d", tag, relY)
				return true
			}
			if failedOnLower {
				continue
			}
			yeqiuLog("%s: 爬绳超时未检测到到顶 attempt=%d", tag, attempt+1)
			continue
		}
		if relX, relY, ok := yeqiuReadMinimapRel(y); ok && yeqiuIsUpperLayerY(y, relY) {
			yeqiuLog("%s: 已到上层 relX=%d relY=%d", tag, relX, relY)
			return true
		}
	}
	return false
}

func yeqiuDoDescend(y *Yeqiu001Config, s *autoClearBagState) {
	yeqiuLog("下落: 下跳回下层开始")
	deadline := time.Now().Add(time.Duration(y.DescendMaxSec) * time.Second)
	waitMs := yeqiuDownJumpCheckMs(y)
	for time.Now().Before(deadline) {
		relX, relY, ok := yeqiuReadMinimapRel(y)
		if !ok {
			core.Sleep(100)
			continue
		}
		if yeqiuIsLowerLayerY(y, relY) {
			yeqiuLog("下落: 已到下层 relX=%d relY=%d", relX, relY)
			return
		}
		if yeqiuTryDownJumpLeaveUpper("下落", y, relX, relY) {
			if _, curY, ok := yeqiuReadMinimapRel(y); ok && yeqiuIsLowerLayerY(y, curY) {
				return
			}
			core.Sleep(y.DescendPollMs)
			continue
		}
		if yeqiuHandleSmallPlatform(s, "下落", y, relX, relY) {
			core.Sleep(y.DescendPollMs)
			continue
		}
		yeqiuLog("下落: 中间层 relX=%d relY=%d 下跳", relX, relY)
		tapDownJump()
		core.Sleep(waitMs)
	}
	yeqiuLog("下落: 下跳回下层超时")
}

func yeqiuDoUpperFarm(tag string, y *Yeqiu001Config, until time.Time, unknown *yeqiuUnknownLayerTracker, s *autoClearBagState, patrolStep yeqiuPatrolStepFunc) {
	goRight := true
	faceRight()
	farmSince := time.Now()
	yeqiuLog("%s: 开始刷怪", tag)
	for time.Now().Before(until) {
		core.BlockWhileCaptchaHold()
		relX, relY, ok := yeqiuReadMinimapRel(y)
		if !ok {
			core.Sleep(100)
			continue
		}
		if yeqiuHandleSmallPlatform(s, tag, y, relX, relY) {
			continue
		}
		if !yeqiuIsUpperLayerY(y, relY) {
			if unknown != nil && unknown.tryJumpLeft(tag, y, relX, relY) {
				continue
			}
			core.Sleep(200)
			continue
		}
		if unknown != nil {
			unknown.active = false
		}
		if patrolStep != nil {
			patrolStep(tag, y, relX, y.UpperXMin, y.UpperXMax, &goRight, farmSince)
		} else {
			yeqiuPatrolFarmStep(tag, y, relX, y.UpperXMin, y.UpperXMax, &goRight, farmSince)
		}
	}
}

func yeqiuRunUpperThenDescend(y *Yeqiu001Config, unknown *yeqiuUnknownLayerTracker, s *autoClearBagState, patrolStep yeqiuPatrolStepFunc) {
	upperUntil := time.Now().Add(yeqiuRandSec(y.UpperFarmSecMin, y.UpperFarmSecMax))
	yeqiuDoUpperFarm("上层", y, upperUntil, unknown, s, patrolStep)
	yeqiuDoDescend(y, s)
}

func runYeqiu001(mapAssetPath, logTag string, patrolStep yeqiuPatrolStepFunc, noTeleport bool) error {
	cfg, err := loadMapConfig(mapAssetPath)
	if err != nil {
		return err
	}
	if cfg.Yeqiu001 == nil {
		return fmt.Errorf("%s: 缺少 yeqiu001 配置", strings.Trim(logTag, "[]"))
	}
	y := cfg.Yeqiu001
	y.normalize()
	y.applyDeleteYellowRegion()
	defer core.ClearMinimapYellowRegion()

	prevLogTag := yeqiuActiveLogTag
	yeqiuActiveLogTag = logTag
	prevNoTeleport := yeqiuActiveNoTeleport
	yeqiuActiveNoTeleport = noTeleport
	defer func() {
		yeqiuActiveLogTag = prevLogTag
		yeqiuActiveNoTeleport = prevNoTeleport
	}()

	SetFarmLogTag(logTag)
	StartFarmMaintainLoop(logTag)
	defer StopFarmMaintainLoop()

	goRight := true
	faceRight()
	lowerUntil := time.Now().Add(yeqiuRandSec(y.LowerFarmSecMin, y.LowerFarmSecMax))
	lowerFarmSince := time.Now()
	var unknownLayer yeqiuUnknownLayerTracker
	clearBag := newAutoClearBagState()
	if core.API.GetConfigBoolValue("自动清包") {
		clearBag.startupPending = false
		clearBag.finishAttempt(yeqiuLog)
		yeqiuLog("自动清包: 小站台 y=[%d,%d] x=[%d,%d] 到间隔可卖货",
			y.SmallPlatformYMin, y.SmallPlatformYMax,
			y.smallPlatformXMin(), y.smallPlatformXMax())
	}

	yeqiuLog("开始挂机 下层y=[%d,%d]打怪x=[%d,%d] 上层y=[%d,%d]打怪x=[%d,%d] 绳x=%d",
		y.LowerYMin, y.LowerYMax, y.LowerXMin, y.LowerXMax,
		y.UpperYMin, y.UpperYMax, y.UpperXMin, y.UpperXMax, y.RopeX)

	for {
		core.BlockWhileCaptchaHold()
		TickFarmMainThreadTasks()

		if !time.Now().Before(lowerUntil) {
			yeqiuLog("下层刷够 去爬绳")
			if yeqiuClimbRope("爬绳", y) {
				yeqiuRunUpperThenDescend(y, &unknownLayer, &clearBag, patrolStep)
				goRight = true
				faceRight()
				lowerUntil = time.Now().Add(yeqiuRandSec(y.LowerFarmSecMin, y.LowerFarmSecMax))
				lowerFarmSince = time.Now()
			} else {
				yeqiuLog("爬绳失败 继续下层")
				lowerUntil = time.Now().Add(30 * time.Second)
				lowerFarmSince = time.Now()
			}
			continue
		}

		relX, relY, ok := yeqiuReadMinimapRel(y)
		if ok && yeqiuHandleSmallPlatform(&clearBag, "下层", y, relX, relY) {
			continue
		}
		if ok && yeqiuIsUpperLayerY(y, relY) {
			unknownLayer.active = false
			yeqiuLog("检测到上层 relX=%d relY=%d 转上层刷怪", relX, relY)
			yeqiuRunUpperThenDescend(y, &unknownLayer, &clearBag, patrolStep)
			goRight = true
			faceRight()
			lowerUntil = time.Now().Add(yeqiuRandSec(y.LowerFarmSecMin, y.LowerFarmSecMax))
			lowerFarmSince = time.Now()
			continue
		}
		if ok && !yeqiuIsKnownLayer(y, relX, relY) {
			if unknownLayer.tryJumpLeft("下层", y, relX, relY) {
				continue
			}
			yeqiuLog("下层: 站位异常 relX=%d relY=%d 等待", relX, relY)
			core.Sleep(200)
			continue
		}
		if ok {
			unknownLayer.active = false
			if patrolStep != nil {
				patrolStep("下层", y, relX, y.LowerXMin, y.LowerXMax, &goRight, lowerFarmSince)
			} else {
				yeqiuPatrolFarmStep("下层", y, relX, y.LowerXMin, y.LowerXMax, &goRight, lowerFarmSince)
			}
		}
	}
}

// Play_叶秋001 下层瞬移刷怪→爬绳→上层短刷→下跳回下层。
func Play_叶秋001(mapAssetPath string) error {
	return runYeqiu001(mapAssetPath, yeqiu001LogTag, nil, false)
}
