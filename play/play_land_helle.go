package play

import (
	"app/core"
	"fmt"
	"math/rand"
	"time"

	"github.com/Dasongzi1366/AutoGo/motion"
)

const (
	landHelleWalkLogTag     = "[land赫勒走路版]"
	landHelleTeleportLogTag = "[land赫勒瞬移版]"

	landHelleWalkPatrolXMin = 40
	landHelleWalkPatrolXMax = 70

	landHelleTeleportPatrolXMin    = 20
	landHelleTeleportPatrolXMax    = 90
	landHelleTeleportClickX1       = 702
	landHelleTeleportClickY1       = 511
	landHelleTeleportClickX2       = 718
	landHelleTeleportClickY2       = 524
	landHelleTeleportAttackMax     = 3
	landHelleTeleportChancePercent = 20 // 瞬移 20%，走路 80%

	landHelleWalkAttackDurMinMs = 1000
	landHelleWalkAttackDurMaxMs = 2000

	landHelleWalkUnitsMin            = 5
	landHelleWalkUnitsMax            = 6
	landHelleWalkMsPerUnit           = 45
	landHelleAttackDurMinMs          = 4000
	landHelleAttackDurMaxMs          = 8000
	landHelleAfterAttackTurnGapMinMs = 800
	landHelleAfterAttackTurnGapMaxMs = 1000

	landHelleSkill2MinInterval  = 2 * time.Minute
	landHelleSkill2MinPreWaitMs = 1000
	landHelleSkill2MinClickX1   = 746
	landHelleSkill2MinClickY1   = 580
	landHelleSkill2MinClickX2   = 763
	landHelleSkill2MinClickY2   = 596

	landHelleSkill15MinInterval = 15 * time.Minute
	landHelleSkill15MinClickX1  = 660
	landHelleSkill15MinClickY1  = 503
	landHelleSkill15MinClickX2  = 670
	landHelleSkill15MinClickY2  = 526

	landHelleSkillClickHoldMinMs = 200
	landHelleSkillClickHoldMaxMs = 300
)

type landHelleTimers struct {
	nextSkill2Min  time.Time
	nextSkill15Min time.Time
}

func (t *landHelleTimers) init() {
	now := time.Now()
	t.nextSkill2Min = now.Add(landHelleSkill2MinInterval)
	t.nextSkill15Min = now.Add(landHelleSkill15MinInterval)
}

type landHelleRun struct {
	logTag   string
	xMin     int
	xMax     int
	teleport bool
	timers   landHelleTimers
	goRight  bool
}

func (r *landHelleRun) log(format string, args ...interface{}) {
	fmt.Printf(r.logTag+" "+format+"\n", args...)
}

func (r *landHelleRun) longClickSkill(tag string, x1, y1, x2, y2 int) {
	r.log("定时: %s 长按[%d,%d,%d,%d] %d-%dms",
		tag, x1, y1, x2, y2, landHelleSkillClickHoldMinMs, landHelleSkillClickHoldMaxMs)
	core.RandomLongClickInArea(x1, y1, x2, y2, landHelleSkillClickHoldMinMs, landHelleSkillClickHoldMaxMs)
}

func (r *landHelleRun) tick15Min() {
	now := time.Now()
	if now.Before(r.timers.nextSkill15Min) {
		return
	}
	r.timers.nextSkill15Min = now.Add(landHelleSkill15MinInterval)
	r.longClickSkill("15分钟", landHelleSkill15MinClickX1, landHelleSkill15MinClickY1,
		landHelleSkill15MinClickX2, landHelleSkill15MinClickY2)
}

func (r *landHelleRun) tick2MinAfterAttack() {
	now := time.Now()
	if now.Before(r.timers.nextSkill2Min) {
		return
	}
	r.timers.nextSkill2Min = now.Add(landHelleSkill2MinInterval)
	r.log("定时: 2分钟 攻击后等%dms", landHelleSkill2MinPreWaitMs)
	core.Sleep(landHelleSkill2MinPreWaitMs)
	r.longClickSkill("2分钟", landHelleSkill2MinClickX1, landHelleSkill2MinClickY1,
		landHelleSkill2MinClickX2, landHelleSkill2MinClickY2)
}

func landHelleWalkUnits() int {
	return landHelleWalkUnitsMin + rand.Intn(landHelleWalkUnitsMax-landHelleWalkUnitsMin+1)
}

func (r *landHelleRun) walkStep(goRight bool) {
	ms := landHelleWalkUnits() * landHelleWalkMsPerUnit
	units := ms / landHelleWalkMsPerUnit
	dir := "左"
	if goRight {
		dir = "右"
	}
	r.log("走位: %s走%d格(%dms)", dir, units, ms)
	if goRight {
		holdDpad(motion.KEYCODE_DPAD_RIGHT, ms)
	} else {
		holdDpad(motion.KEYCODE_DPAD_LEFT, ms)
	}
}

func landHelleFace(goRight bool) {
	if goRight {
		landFaceRight()
	} else {
		landFaceLeft()
	}
	core.Sleep(landFaceSettleMs)
}

func landHelleClickTeleport(goRight bool) {
	dirCode := motion.KEYCODE_DPAD_LEFT
	if goRight {
		dirCode = motion.KEYCODE_DPAD_RIGHT
	}
	if goRight {
		releaseDpadHold(motion.KEYCODE_DPAD_LEFT)
	} else {
		releaseDpadHold(motion.KEYCODE_DPAD_RIGHT)
	}
	motion.KeyActionDown(dirCode, 0)
	core.Sleep(40)
	core.RandomClickInArea(
		landHelleTeleportClickX1, landHelleTeleportClickY1,
		landHelleTeleportClickX2, landHelleTeleportClickY2,
	)
	core.Sleep(40)
	motion.KeyActionUp(dirCode, 0)
}

func landHelleAttackTimes() int {
	return rand.Intn(landHelleTeleportAttackMax + 1)
}

func (r *landHelleRun) doAttacks(n int, tag string) {
	if n <= 0 {
		r.log("攻击: %s 跳过(0次)", tag)
		return
	}
	r.log("攻击: %s ×%d", tag, n)
	for i := 0; i < n; i++ {
		core.BlockWhileCaptchaHold()
		landClickAttackHold()
		if i < n-1 {
			core.Sleep(landAttackIntervalMs)
		}
	}
	r.tick2MinAfterAttack()
}

func (r *landHelleRun) teleportAndAttack(goRight bool, tag string) {
	dir := "左"
	if goRight {
		dir = "右"
	}
	r.log("瞬移: %s", dir)
	landHelleClickTeleport(goRight)
	core.Sleep(50)
	n := landHelleAttackTimes()
	r.doAttacks(n, tag)
}

func (r *landHelleRun) patrolMargin() int {
	span := r.xMax - r.xMin
	if span <= 20 {
		return 4
	}
	if span <= 40 {
		return 8
	}
	return 12
}

func (r *landHelleRun) correctXWalk(relX int) bool {
	if relX < r.xMin {
		r.log("relX=%d<%d 右走", relX, r.xMin)
		r.walkStep(true)
		return true
	}
	if relX > r.xMax {
		r.log("relX=%d>%d 左走", relX, r.xMax)
		r.walkStep(false)
		return true
	}
	return false
}

func (r *landHelleRun) attackForDurationMs(goRight bool, minMs, maxMs int, tag string) {
	landHelleFace(goRight)
	dur := minMs + rand.Intn(maxMs-minMs+1)
	dir := "左"
	if goRight {
		dir = "右"
	}
	r.log("攻击: 朝%s %dms(%s)", dir, dur, tag)
	deadline := time.Now().Add(time.Duration(dur) * time.Millisecond)
	for time.Now().Before(deadline) {
		core.BlockWhileCaptchaHold()
		landClickAttackHold()
		if time.Now().Before(deadline) {
			core.Sleep(landAttackIntervalMs)
		}
	}
	r.tick2MinAfterAttack()
}

func (r *landHelleRun) attackForDuration(goRight bool) {
	r.attackForDurationMs(goRight, landHelleAttackDurMinMs, landHelleAttackDurMaxMs, "左右打")
	core.RandomSleep(landHelleAfterAttackTurnGapMinMs, landHelleAfterAttackTurnGapMaxMs)
	r.log("攻击结束 等待%d-%dms后换向", landHelleAfterAttackTurnGapMinMs, landHelleAfterAttackTurnGapMaxMs)
}

func (r *landHelleRun) patrolMoveStep(goRight bool, tag string) {
	if rand.Intn(100) < landHelleTeleportChancePercent {
		r.teleportAndAttack(goRight, tag)
		return
	}
	dir := "左"
	if goRight {
		dir = "右"
	}
	r.log("走路: %s(%s)", dir, tag)
	r.walkStep(goRight)
	r.attackForDurationMs(goRight, landHelleWalkAttackDurMinMs, landHelleWalkAttackDurMaxMs, tag)
}

func (r *landHelleRun) patrolTeleportStep(relX int) {
	margin := r.patrolMargin()

	if relX < r.xMin {
		r.patrolMoveStep(true, "回区")
		r.goRight = true
		return
	}
	if relX > r.xMax {
		r.patrolMoveStep(false, "回区")
		r.goRight = false
		return
	}

	if r.goRight && relX >= r.xMax-margin {
		r.log("近右界 relX=%d 改向左", relX)
		r.goRight = false
		core.Sleep(80)
		return
	}
	if !r.goRight && relX <= r.xMin+margin {
		r.log("近左界 relX=%d 改向右", relX)
		r.goRight = true
		core.Sleep(80)
		return
	}

	dir := "左"
	if r.goRight {
		dir = "右"
	}
	r.patrolMoveStep(r.goRight, dir+"移动")
}

func runLandHelle(r *landHelleRun) error {
	mode := "走路"
	if r.teleport {
		mode = "瞬移"
	}
	r.log("开始 %s x=[%d,%d] 瞬移%d%%走路%d%% 2分钟+15分钟技能",
		mode, r.xMin, r.xMax, landHelleTeleportChancePercent, 100-landHelleTeleportChancePercent)
	r.timers.init()
	r.goRight = true

	for {
		core.BlockWhileCaptchaHold()
		r.tick15Min()

		relX, relY, ok := ReadMinimapRel()
		if !ok {
			r.log("小地图未识别 等待")
			core.Sleep(200)
			continue
		}

		if r.teleport {
			r.patrolTeleportStep(relX)
			continue
		}

		if r.correctXWalk(relX) {
			continue
		}

		r.log("就位 relX=%d relY=%d 左右攻击", relX, relY)
		r.attackForDuration(false)
		r.attackForDuration(true)
	}
}

func startLandHelle(mapAssetPath string) error {
	cfg, err := loadMapConfig(mapAssetPath)
	if err != nil {
		return err
	}
	applyMapMinimapRegions(cfg)
	return nil
}

// Play_land赫勒地区走路版 x∈[40,70] 走路 + 左右攻击各 4～8s。
func Play_land赫勒地区走路版(mapAssetPath string) error {
	if err := startLandHelle(mapAssetPath); err != nil {
		return err
	}
	defer clearMapMinimapRegions()
	return runLandHelle(&landHelleRun{
		logTag:   landHelleWalkLogTag,
		xMin:     landHelleWalkPatrolXMin,
		xMax:     landHelleWalkPatrolXMax,
		teleport: false,
	})
}

// Play_land赫勒地区瞬移版 x∈[20,90] 巡逻 20% 瞬移+0～3 攻击、80% 走路+攻击 1～2s。
func Play_land赫勒地区瞬移版(mapAssetPath string) error {
	if err := startLandHelle(mapAssetPath); err != nil {
		return err
	}
	defer clearMapMinimapRegions()
	return runLandHelle(&landHelleRun{
		logTag:   landHelleTeleportLogTag,
		xMin:     landHelleTeleportPatrolXMin,
		xMax:     landHelleTeleportPatrolXMax,
		teleport: true,
	})
}
