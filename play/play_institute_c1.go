package play

import (
	"app/core"
	"fmt"
	"math/rand"
	"time"

	"github.com/Dasongzi1366/AutoGo/motion"
)

func instituteC1LogTag(name string) string {
	return fmt.Sprintf("[%s]", name)
}

func instituteC1Log(tag, format string, args ...interface{}) {
	fmt.Printf(tag+" "+format+"\n", args...)
}

func readMinimapRel() (relX, relY int, ok bool) {
	return ReadMinimapRel()
}

func stairsXInTarget(st *StairsFarmConfig, relX int) bool {
	return relX >= st.TargetX-st.TargetXTolerance && relX <= st.TargetX+st.TargetXTolerance
}

func stairsXDelta(relX, targetX int) int {
	d := relX - targetX
	if d < 0 {
		return -d
	}
	return d
}

func logMinimapPos(tag, phase string) (relX, relY int, ok bool) {
	relX, relY, ok = readMinimapRel()
	if ok {
		instituteC1Log(tag, "%s relX=%d relY=%d", phase, relX, relY)
	} else {
		instituteC1Log(tag, "%s 小地图未识别", phase)
	}
	return relX, relY, ok
}

func stairsMoveX(tag string, st *StairsFarmConfig, relX, targetX int, belowPlatform bool) {
	delta := stairsXDelta(relX, targetX)
	goRight := relX < targetX

	if belowPlatform && delta > st.TeleportXDeltaMin {
		dir := "左瞬移"
		if goRight {
			dir = "右瞬移"
		}
		instituteC1Log(tag, "【掉阶】x距离=%d>%d relX=%d→%d %s", delta, st.TeleportXDeltaMin, relX, targetX, dir)
		tapTeleportWithDirection(goRight)
		sleepAfterTeleport()
		return
	}

	msMin, msMax := st.SlowWalkMsMin, st.SlowWalkMsMax
	if belowPlatform {
		msMin *= st.RecoverWalkMultiplier
		msMax *= st.RecoverWalkMultiplier
	}
	ms := msMin + rand.Intn(msMax-msMin+1)
	if belowPlatform {
		dir := "左"
		if goRight {
			dir = "右"
		}
		instituteC1Log(tag, "【掉阶】台阶下慢走%s %dms relX=%d→%d", dir, ms, relX, targetX)
	}
	walkHoldMs(goRight, ms)
	core.Sleep(30)
}

func alignXSlowToTargetWithPrefix(tag string, st *StairsFarmConfig, targetX int, prefix string) {
	for pass := 0; pass < st.AlignMaxPass; pass++ {
		relX, _, ok := readMinimapRel()
		if !ok {
			instituteC1Log(tag, "%s: pass=%d 小地图未识别", prefix, pass+1)
			core.Sleep(80)
			continue
		}
		if relX >= targetX-st.TargetXTolerance && relX <= targetX+st.TargetXTolerance {
			instituteC1Log(tag, "%s: 就位 relX=%d 目标=%d", prefix, relX, targetX)
			return
		}
		dir := "右"
		if relX > targetX {
			dir = "左"
		}
		instituteC1Log(tag, "%s: pass=%d relX=%d→%d 慢走%s", prefix, pass+1, relX, targetX, dir)
		stairsMoveX(tag, st, relX, targetX, false)
	}
	instituteC1Log(tag, "%s: 达最大轮次仍未就位", prefix)
}

func alignXSlowToTarget(tag string, st *StairsFarmConfig, targetX int) {
	alignXSlowToTargetWithPrefix(tag, st, targetX, "x对齐")
}

// alignXFastThenSlow 偏左时先快走再慢走；不含左走大步。
func alignXFastThenSlow(tag string, st *StairsFarmConfig, targetX int, prefix string) {
	relX, _, ok := readMinimapRel()
	if ok && relX < targetX-st.TargetXTolerance {
		instituteC1Log(tag, "%s: 先快走右 %dms relX=%d→%d", prefix, st.AlignFastWalkMs, relX, targetX)
		walkHoldMs(true, st.AlignFastWalkMs)
		core.Sleep(50)
	}
	alignXSlowToTargetWithPrefix(tag, st, targetX, prefix)
}

// alignXAfterLeftWalk 左走后先往右快走一段，再慢走微调 x。
func alignXAfterLeftWalk(tag string, st *StairsFarmConfig, targetX int) {
	alignXFastThenSlow(tag, st, targetX, "x对齐")
}

func alignXToRecoverRange(tag string, st *StairsFarmConfig) {
	for pass := 0; pass < st.RecoverMaxPass; pass++ {
		relX, relY, ok := readMinimapRel()
		if !ok {
			instituteC1Log(tag, "【掉阶】阶段1 pass=%d 小地图未识别", pass+1)
			core.Sleep(80)
			continue
		}
		if relX >= st.RecoverXMin && relX <= st.RecoverXMax {
			instituteC1Log(tag, "【掉阶】阶段1完成 x就位 relX=%d relY=%d 区间[%d,%d]", relX, relY, st.RecoverXMin, st.RecoverXMax)
			return
		}
		targetX := st.RecoverXMin
		if relX > st.RecoverXMax {
			targetX = st.RecoverXMax
		}
		dir := "右"
		if relX > targetX {
			dir = "左"
		}
		instituteC1Log(tag, "【掉阶】阶段1 pass=%d relX=%d relY=%d 朝%s→[%d,%d]", pass+1, relX, relY, dir, st.RecoverXMin, st.RecoverXMax)
		stairsMoveX(tag, st, relX, targetX, true)
	}
	instituteC1Log(tag, "【掉阶】阶段1 达最大轮次 x可能未完全就位")
}

func recoverFromFall(tag string, st *StairsFarmConfig, relX, relY int) {
	instituteC1Log(tag, "【掉阶】开始 relX=%d relY=%d (y>%d) target_x=%d y目标=[%d,%d]",
		relX, relY, st.YMax, st.TargetX, st.YMin, st.YMax)
	instituteC1Log(tag, "【掉阶】阶段1: 台阶下x移到[%d,%d]", st.RecoverXMin, st.RecoverXMax)
	alignXToRecoverRange(tag, st)

	logMinimapPos(tag, "【掉阶】阶段1结束")
	instituteC1Log(tag, "【掉阶】阶段2: 上+瞬移")
	tapUpTeleport()
	core.Sleep(st.FallTeleportWaitMs)

	logMinimapPos(tag, "【掉阶】上台阶后")
	instituteC1Log(tag, "【掉阶】阶段3: 回到target_x=%d（不左走，避免掉边）", st.TargetX)
	alignXFastThenSlow(tag, st, st.TargetX, "【掉阶】回位")

	logMinimapPos(tag, "【掉阶】恢复完成")
}

func recoverFromStairs(tag string, st *StairsFarmConfig, relX, relY int) {
	instituteC1Log(tag, "上台阶: y=%d<%d relX=%d → 下跳", relY, st.YMin, relX)
	tapDownJump()
	core.Sleep(st.StairsDownJumpWaitMs)
}

func nudgeXIfNeeded(tag string, st *StairsFarmConfig, relX int) bool {
	if stairsXInTarget(st, relX) {
		return false
	}
	dir := "右"
	if relX > st.TargetX {
		dir = "左"
	}
	instituteC1Log(tag, "微调x: relX=%d 目标=%d 朝%s", relX, st.TargetX, dir)
	stairsMoveX(tag, st, relX, st.TargetX, false)
	return true
}

func handleStairsPosition(tag string, st *StairsFarmConfig, relX, relY int, skipDownJumpUntil time.Time) bool {
	if relY > st.YMax {
		recoverFromFall(tag, st, relX, relY)
		return true
	}
	if relY < st.YMin {
		if time.Now().Before(skipDownJumpUntil) {
			instituteC1Log(tag, "y=%d<%d 台阶调整冷却中，跳过下跳", relY, st.YMin)
			return false
		}
		recoverFromStairs(tag, st, relX, relY)
		return true
	}
	return nudgeXIfNeeded(tag, st, relX)
}

func faceSwitchInterval(st *StairsFarmConfig) time.Duration {
	span := st.FaceSwitchSecMax - st.FaceSwitchSecMin
	sec := st.FaceSwitchSecMin
	if span > 0 {
		sec += rand.Intn(span + 1)
	}
	return time.Duration(sec) * time.Second
}

func maybeSwitchAttackFace(tag string, st *StairsFarmConfig, facingRight *bool, nextSwitch *time.Time) {
	if time.Now().Before(*nextSwitch) {
		return
	}
	pauseLo := st.FaceSwitchPauseMs * 90 / 100
	pauseHi := st.FaceSwitchPauseMs * 110 / 100
	if pauseHi <= pauseLo {
		pauseHi = pauseLo + 50
	}
	instituteC1Log(tag, "换向: 暂停攻击 %dms 左右", st.FaceSwitchPauseMs)
	core.RandomSleep(pauseLo, pauseHi)
	*facingRight = !*facingRight
	if *facingRight {
		faceRight()
		instituteC1Log(tag, "换向: 朝右")
	} else {
		faceLeft()
		instituteC1Log(tag, "换向: 朝左")
	}
	*nextSwitch = time.Now().Add(faceSwitchInterval(st))
}

func stairAdjustInterval(st *StairsFarmConfig) time.Duration {
	span := st.AdjustIntervalSecMax - st.AdjustIntervalSecMin
	if span < 0 {
		span = 0
	}
	sec := st.AdjustIntervalSecMin
	if span > 0 {
		sec += rand.Intn(span + 1)
	}
	return time.Duration(sec) * time.Second
}

func alignXForStairAdjust(tag string, st *StairsFarmConfig) {
	ax, tol := st.AdjustTargetX, st.AdjustTargetXTolerance
	relX, _, ok := readMinimapRel()
	if ok && relX < ax-tol {
		instituteC1Log(tag, "台阶调整: 先快走右 %dms relX=%d→%d", st.AlignFastWalkMs, relX, ax)
		walkHoldMs(true, st.AlignFastWalkMs)
		core.Sleep(50)
	}
	for pass := 0; pass < st.AlignMaxPass; pass++ {
		relX, _, ok = readMinimapRel()
		if !ok {
			instituteC1Log(tag, "台阶调整: x对齐 pass=%d 小地图未识别", pass+1)
			core.Sleep(80)
			continue
		}
		if relX >= ax-tol && relX <= ax+tol {
			instituteC1Log(tag, "台阶调整: x就位 relX=%d 目标=%d", relX, ax)
			return
		}
		dir := "右"
		if relX > ax {
			dir = "左"
		}
		instituteC1Log(tag, "台阶调整: x对齐 pass=%d relX=%d→%d 慢走%s", pass+1, relX, ax, dir)
		stairsMoveX(tag, st, relX, ax, false)
	}
	instituteC1Log(tag, "台阶调整: x对齐达最大轮次")
}

func doStairAdjustRoutine(tag string, st *StairsFarmConfig) time.Time {
	instituteC1Log(tag, "台阶调整: 等待%dms", st.AdjustWaitMs)
	core.Sleep(st.AdjustWaitMs)
	instituteC1Log(tag, "台阶调整: 对齐 x→%d", st.AdjustTargetX)
	alignXForStairAdjust(tag, st)
	instituteC1Log(tag, "台阶调整: 先下 %d～%dms（朝右）", st.DownHoldMsMin, st.DownHoldMsMax)
	faceRight()
	core.Sleep(40)
	holdDpadRandom(motion.KEYCODE_DPAD_DOWN, st.DownHoldMsMin, st.DownHoldMsMax)
	instituteC1Log(tag, "台阶调整: 后上 %d～%dms", st.UpHoldMsMin, st.UpHoldMsMax)
	holdDpadRandom(motion.KEYCODE_DPAD_UP, st.UpHoldMsMin, st.UpHoldMsMax)
	instituteC1Log(tag, "台阶调整: 上来后左跳")
	tapJumpLeft()
	core.Sleep(st.AfterUpLeftJumpWaitMs)
	until := time.Now().Add(time.Duration(st.AfterAdjustNoDownJumpMs) * time.Millisecond)
	instituteC1Log(tag, "台阶调整: 完成 %dms 内不触发下跳", st.AfterAdjustNoDownJumpMs)
	return until
}

// Play_研究所C1 持续按空格；按 JSON stairs 维持站位并周期调整台阶。
func Play_研究所C1(mapAssetPath string) error {
	cfg, err := loadMapConfig(mapAssetPath)
	if err != nil {
		return err
	}
	if cfg.Stairs == nil {
		return fmt.Errorf("%s: 缺少 stairs 配置", cfg.Name)
	}
	st := cfg.Stairs
	st.normalize()

	logTag := instituteC1LogTag(cfg.Name)
	SetFarmLogTag(logTag)
	applyMapMinimapRegions(cfg)
	defer clearMapMinimapRegions()

	StartFarmMaintainLoop(logTag)
	defer StopFarmMaintainLoop()

	nextStair := time.Now().Add(stairAdjustInterval(st))
	var skipDownJumpUntil time.Time
	clearBag := newAutoClearBagState()
	if clearBag.startupPending {
		instituteC1Log(logTag, "自动清包: 启动后优先清包 (需 x=[%d,%d] y=[%d,%d])",
			st.TargetX-st.TargetXTolerance, st.TargetX+st.TargetXTolerance, st.YMin, st.YMax)
	}
	facingRight := true
	nextFaceSwitch := time.Now().Add(faceSwitchInterval(st))
	faceRight()
	instituteC1Log(logTag, "开始挂机 目标 y=[%d,%d] x=[%d,%d]",
		st.YMin, st.YMax, st.TargetX-st.TargetXTolerance, st.TargetX+st.TargetXTolerance)

	for {
		core.BlockWhileCaptchaHold()
		TickFarmMainThreadTasks()

		if !time.Now().Before(nextStair) && !clearBag.startupPending {
			skipDownJumpUntil = doStairAdjustRoutine(logTag, st)
			nextStair = time.Now().Add(stairAdjustInterval(st))
			continue
		}

		if relX, relY, ok := readMinimapRel(); ok {
			if handleStairsPosition(logTag, st, relX, relY, skipDownJumpUntil) {
				continue
			}
			if tryAutoClearBagInstituteC1(&clearBag, logTag, st, relX, relY) {
				continue
			}
		}

		maybeSwitchAttackFace(logTag, st, &facingRight, &nextFaceSwitch)
		if !clearBag.startupPending {
			keyHoldPress(attackKeyCode(), keyHoldShortMin, keyHoldShortMax)
		}
	}
}
