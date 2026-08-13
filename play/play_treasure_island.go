package play

import (
	"app/assets"
	"app/core"
	"fmt"
	"time"

	"github.com/Dasongzi1366/AutoGo/yolo"
)

const treasureIslandLogTag = "[抢夺宝物岛]"

const (
	tiPhaseFarm    = iota // 当前层打怪
	tiPhaseDescend        // 往下层走
	tiPhaseAscend         // 往上层走
)

const tiAdoptOffLayerMinSec = 5 // 意外换层后在新层至少停留此秒数才认层

type tiRunState struct {
	phase              int
	farmLayer          int
	ascendTarget       int
	descendTarget      int  // 下落阶段目标层（正常换层=farmLayer+1；偏层恢复=farmLayer）
	afterL3Ascend      bool // L3 打完上升后，L2 打完再回 L1
	layerEnteredAt     time.Time
	offLayer           int       // 意外偏离目标层时检测到的层
	offLayerSince      time.Time // 首次检测到 offLayer 的时间
	patrolGoRight      bool
	leavingLayer       bool // 已决定换层，不再回头打怪
	clearBag           autoClearBagState
	ascendAlignSince   time.Time // 上升 x 对齐开始时刻；零值表示未在对齐
	unknownLayerStreak int       // 连续 0 层计数（仅打怪阶段）
	lastUnknownJumpAt  time.Time // 上次 0 层右跳时刻
}

const (
	tiAscendAlignTimeoutSec       = 3
	tiAscendAlignAttackHoldMinMs  = 1000
	tiAscendAlignAttackHoldMaxMs  = 2000
	tiUnknownLayerJumpCooldownSec = 3
	tiAscendWalkMsPerDist         = 45 // walk_ms ≈ b×|relX-ascendX|；可被配置 ascend_near_walk_ms_per_dist 覆盖
)

func (s *tiRunState) beginAscendAlign() {
	if s.ascendAlignSince.IsZero() {
		s.ascendAlignSince = time.Now()
	}
}

func (s *tiRunState) clearAscendAlign() {
	s.ascendAlignSince = time.Time{}
}

func (s *tiRunState) ascendAlignTimedOut() bool {
	if s.ascendAlignSince.IsZero() {
		return false
	}
	return time.Since(s.ascendAlignSince) >= tiAscendAlignTimeoutSec*time.Second
}

func (s *tiRunState) resetAscendAlignTimer() {
	s.ascendAlignSince = time.Now()
}

func tiLog(format string, args ...interface{}) {
	fmt.Printf(treasureIslandLogTag+" "+format+"\n", args...)
}

func tiPhaseName(p int) string {
	switch p {
	case tiPhaseFarm:
		return "打怪"
	case tiPhaseDescend:
		return "下落"
	case tiPhaseAscend:
		return "上升"
	default:
		return "?"
	}
}

func normalizeTreasureIslandConfig(cfg *mapConfig) {
	if cfg == nil {
		return
	}
	if cfg.SkipMonsterScanRelXAbove <= defaultSkipMonsterScanRelXAbove {
		cfg.SkipMonsterScanRelXAbove = 9999
	}
	if cfg.TreasureIsland != nil {
		cfg.TreasureIsland.normalize()
		applyTreasureIslandStayFromContent(cfg.TreasureIsland)
	}
}

func applyTreasureIslandStayFromContent(ti *TreasureIslandConfig) {
	if ti == nil {
		return
	}
	type layerStayKeys struct {
		layer          int
		minKey, maxKey string
	}
	for _, ls := range []layerStayKeys{
		{1, "下层最少打怪秒数", "下层最多打怪秒数"},
		{2, "中层最少打怪秒数", "中层最多打怪秒数"},
		{3, "最上层最少打怪秒数", "最上层最多打怪秒数"},
	} {
		def := ti.layerDef(ls.layer)
		if def == nil {
			continue
		}
		if v, err := core.API.GetConfigInt(ls.minKey); err == nil && v > 0 {
			def.MinFarmStaySec = v
		}
		if v, err := core.API.GetConfigInt(ls.maxKey); err == nil && v > 0 {
			def.MaxStaySec = v
		}
	}
}

func logTreasureIslandLayerStay(ti *TreasureIslandConfig) {
	if ti == nil {
		return
	}
	for _, lay := range []int{1, 2, 3} {
		def := ti.layerDef(lay)
		if def == nil {
			continue
		}
		name := fmt.Sprintf("L%d", lay)
		switch lay {
		case 1:
			name = "下层"
		case 2:
			name = "中层"
		case 3:
			name = "最上层"
		}
		tiLog("%s 打怪秒数 最少=%ds 最多=%ds", name, def.MinFarmStaySec, def.MaxStaySec)
	}
}

func tiDetectLayer(ti *TreasureIslandConfig, relY int) int {
	if ti != nil {
		if lay := ti.detectLayer(relY); lay > 0 {
			return lay
		}
	}
	return 0
}

func tiLayerMonsterCounts(cfg *mapConfig, eng *yolo.Yolo, lay, relX int) (monsterN, attackN int) {
	if lay <= 0 || eng == nil || !shouldScanMonsters(cfg, relX) {
		return 0, 0
	}
	attackable, _, labeled := detectMonsters(cfg, eng, lay, relX)
	return len(labeled), len(attackable)
}

func tiInXRange(relX, center, tol int) bool {
	return relX >= center-tol && relX <= center+tol
}

func tiInFightRange(def *TreasureIslandLayerConfig, relX int) bool {
	if def == nil {
		return true
	}
	return matchRange(relX, def.FightXMin, def.FightXMax)
}

func tiScanLayer(st *tiRunState, lay int) int {
	if lay > 0 {
		return lay
	}
	if st.phase == tiPhaseAscend {
		return st.ascendTarget + 1
	}
	if st.phase == tiPhaseDescend && st.descendTarget > 1 {
		return st.descendTarget - 1
	}
	return st.farmLayer
}

func (s *tiRunState) minFarmStaySec(ti *TreasureIslandConfig) int {
	minStay := ti.MinFarmStaySec
	if def := ti.layerDef(s.farmLayer); def != nil && def.MinFarmStaySec > 0 {
		minStay = def.MinFarmStaySec
	}
	return minStay
}

func (s *tiRunState) minFarmStayLeft(ti *TreasureIslandConfig) time.Duration {
	minStay := time.Duration(s.minFarmStaySec(ti)) * time.Second
	elapsed := time.Since(s.layerEnteredAt)
	if elapsed >= minStay {
		return 0
	}
	return minStay - elapsed
}

func tiStartLeaveLayer(s *tiRunState) {
	s.leavingLayer = true
}

func tiShouldTransitionFight(def *TreasureIslandLayerConfig, ti *TreasureIslandConfig, monsterN, attackN int) bool {
	if def != nil && def.FightAttackMin > 0 {
		return attackN >= def.FightAttackMin
	}
	return monsterN >= ti.TransitionFightMonsterMin
}

func tiHandleTransitionFight(cfg *mapConfig, eng *yolo.Yolo, ti *TreasureIslandConfig, st *tiRunState, lay, relX int, when string) bool {
	if st.leavingLayer || st.phase == tiPhaseFarm || ti == nil {
		return false
	}
	scanLay := tiScanLayer(st, lay)
	def := ti.layerDef(scanLay)
	if def == nil {
		return false
	}
	monsterN, attackN := tiLayerMonsterCounts(cfg, eng, scanLay, relX)
	if !tiShouldTransitionFight(def, ti, monsterN, attackN) {
		return false
	}
	if !tiInFightRange(def, relX) {
		tiLog("换层%s L%d 需攻击=%d x=%d不在[%d,%d] 对齐打怪区", when, scanLay, attackN, relX, def.FightXMin, def.FightXMax)
		tiAlignXByTeleport("", relX, def.FightXMin, def.FightXMax)
		return true
	}
	tiLog("换层%s L%d 需攻击=%d x在[%d,%d] 攻击", when, scanLay, attackN, def.FightXMin, def.FightXMax)
	tapAttack()
	return true
}

func tiTeleportTransition(cfg *mapConfig, eng *yolo.Yolo, ti *TreasureIslandConfig, st *tiRunState, lay, relX int, goRight bool) bool {
	if tiHandleTransitionFight(cfg, eng, ti, st, lay, relX, "瞬移前") {
		return true
	}
	if goRight {
		tapTeleportWithDirection(true)
	} else {
		tapTeleportWithDirection(false)
	}
	sleepAfterTeleport()
	tiHandleTransitionFight(cfg, eng, ti, st, lay, relX, "瞬移后")
	return true
}

func tiAbsDist(a, b int) int {
	if a > b {
		return a - b
	}
	return b - a
}

func tiAscendAlignWalkMs(ti *TreasureIslandConfig, dist int) int {
	if dist <= 0 {
		return ti.AscendNearWalkMsMin
	}
	b := tiAscendWalkMsPerDist
	if ti != nil && ti.AscendNearWalkMsPerDist > 0 {
		b = ti.AscendNearWalkMsPerDist
	}
	ms := b * dist
	if ti != nil {
		if ti.AscendNearWalkMsMin > 0 && ms < ti.AscendNearWalkMsMin {
			ms = ti.AscendNearWalkMsMin
		}
		if ti.AscendNearWalkMsMax > 0 && ms > ti.AscendNearWalkMsMax {
			ms = ti.AscendNearWalkMsMax
		}
	}
	return ms
}

func tiAlignAscendX(cfg *mapConfig, eng *yolo.Yolo, ti *TreasureIslandConfig, st *tiRunState, lay, relX int, def *TreasureIslandLayerConfig) bool {
	if def == nil {
		return false
	}
	xMin := def.AscendX - def.AscendXTolerance
	xMax := def.AscendX + def.AscendXTolerance
	if relX >= xMin && relX <= xMax {
		return false
	}
	dist := tiAbsDist(relX, def.AscendX)
	if dist <= ti.AscendNearWalkDistMax {
		goRight := relX < def.AscendX
		ms := tiAscendAlignWalkMs(ti, dist)
		b := tiAscendWalkMsPerDist
		if ti != nil && ti.AscendNearWalkMsPerDist > 0 {
			b = ti.AscendNearWalkMsPerDist
		}
		dir := "右走"
		if !goRight {
			dir = "左走"
		}
		tiLog("上升对齐 x=%d 距目标%d仅%d %s%dms(b=%d)", relX, def.AscendX, dist, dir, ms, b)
		walkHoldMs(goRight, ms)
		core.Sleep(100)
		return true
	}
	tiLog("上升对齐 x=%d 距目标%d=%d 瞬移→[%d,%d]", relX, def.AscendX, dist, xMin, xMax)
	return tiAlignXByTeleportTransition(cfg, eng, ti, st, lay, relX, xMin, xMax)
}

func tiAlignXByTeleport(tag string, relX, xMin, xMax int) bool {
	if relX >= xMin && relX <= xMax {
		return false
	}
	dir := "右瞬移"
	if relX > xMax {
		dir = "左瞬移"
		tapTeleportWithDirection(false)
	} else {
		tapTeleportWithDirection(true)
	}
	tiLog("%s relX=%d→[%d,%d]", dir, relX, xMin, xMax)
	sleepAfterTeleport()
	return true
}

func tiAlignXByTeleportTransition(cfg *mapConfig, eng *yolo.Yolo, ti *TreasureIslandConfig, st *tiRunState, lay, relX, xMin, xMax int) bool {
	if relX >= xMin && relX <= xMax {
		return false
	}
	dir := "右瞬移"
	goRight := true
	if relX > xMax {
		dir = "左瞬移"
		goRight = false
	}
	tiLog("%s relX=%d→[%d,%d]", dir, relX, xMin, xMax)
	return tiTeleportTransition(cfg, eng, ti, st, lay, relX, goRight)
}

const tiMinimapSettleWaitMs = 1000 // 下跳/上瞬移前等待小地图黄点刷新

func tiReadMinimapRel() (relX, relY int, ok bool) {
	mx, my, wx, wy := detectYellowThenWorld()
	if mx < 0 || my < 0 || wx < 0 || wy < 0 {
		return 0, 0, false
	}
	relX, relY = relativeToRef(mx, my, wx, wy)
	return relX, relY, true
}

func tiTryDownJump(tag string, ti *TreasureIslandConfig, relY int) bool {
	if ti == nil || !matchRange(relY, ti.DownJumpYMin, ti.DownJumpYMax) {
		return false
	}
	core.Sleep(tiMinimapSettleWaitMs)
	_, relY2, ok := tiReadMinimapRel()
	if !ok {
		tiLog("下跳: 等待%dms后小地图未识别", tiMinimapSettleWaitMs)
		return true
	}
	if !matchRange(relY2, ti.DownJumpYMin, ti.DownJumpYMax) {
		tiLog("下跳: 等待后 y=%d 不在[%d,%d] 跳过", relY2, ti.DownJumpYMin, ti.DownJumpYMax)
		return false
	}
	tiLog("y=%d 在[%d,%d] 下跳", relY2, ti.DownJumpYMin, ti.DownJumpYMax)
	tapDownJump()
	core.Sleep(ti.DownJumpWaitMs)
	return true
}

func tiTrySpecialRightTeleport(relX, relY int) bool {
	if matchRange(relX, -10, 10) && matchRange(relY, 152, 156) {
		tiLog("特殊点 x=%d y=%d 右瞬移", relX, relY)
		tapTeleportWithDirection(true)
		sleepAfterTeleport()
		return true
	}
	if matchRange(relX, 10, 20) && matchRange(relY, 139, 143) {
		tiLog("特殊点 x=%d y=%d 右瞬移", relX, relY)
		tapTeleportWithDirection(true)
		sleepAfterTeleport()
		return true
	}
	return false
}

func tiDoFarmPatrolStep(st *tiRunState, def *TreasureIslandLayerConfig, relX int) bool {
	xMin, xMax := def.patrolXBounds()
	tol := 4
	if relX >= xMax-tol {
		st.patrolGoRight = false
	} else if relX <= xMin+tol {
		st.patrolGoRight = true
	}
	goRight := st.patrolGoRight
	dir := "左"
	if goRight {
		dir = "右"
	}
	if patrolFarmAllowWalk(st.layerEnteredAt) {
		tiLog("L%d 巡逻 x=%d %s走+攻击 范围[%d,%d]", st.farmLayer, relX, dir, xMin, xMax)
		if goRight {
			faceRight()
		} else {
			faceLeft()
		}
		walkHoldMs(goRight, patrolFarmWalkMs())
		tapAttackTwice()
		return true
	}
	tiLog("L%d 巡逻 x=%d %s瞬移 范围[%d,%d]", st.farmLayer, relX, dir, xMin, xMax)
	tapTeleportWithDirection(goRight)
	sleepAfterTeleport()
	tapAttackTwice()
	return true
}

func (s *tiRunState) enterFarmLayer(layer int, ti *TreasureIslandConfig) {
	s.phase = tiPhaseFarm
	s.farmLayer = layer
	s.layerEnteredAt = time.Now()
	s.leavingLayer = false
	s.offLayer = 0
	if layer == 1 || layer == 2 || layer == 3 {
		s.patrolGoRight = true
	}
}

func (s *tiRunState) startDescend(target int) {
	s.phase = tiPhaseDescend
	s.descendTarget = target
	s.leavingLayer = true
}

func (s *tiRunState) startAscend(target int) {
	s.phase = tiPhaseAscend
	s.ascendTarget = target
	s.leavingLayer = true
	s.clearAscendAlign()
}

func (s *tiRunState) layerStayExceeded(def *TreasureIslandLayerConfig) bool {
	if def == nil || def.MaxStaySec <= 0 {
		return false
	}
	return time.Since(s.layerEnteredAt) >= time.Duration(def.MaxStaySec)*time.Second
}

func (s *tiRunState) refreshAscendZoneEnabled(ti *TreasureIslandConfig, lay int) {
	if ti == nil {
		return
	}
	ti.AscendZoneEnabled = false
	if lay != 3 {
		return
	}
	def := ti.layerDef(3)
	if def == nil || def.MaxStaySec <= 0 {
		return
	}
	ti.AscendZoneEnabled = s.layerStayExceeded(def)
}

func tiTryL3AscendZoneFight(cfg *mapConfig, eng *yolo.Yolo, ti *TreasureIslandConfig, lay, relX int) bool {
	if lay != 3 || ti == nil || !ti.AscendZoneEnabled {
		return false
	}
	def := ti.layerDef(3)
	if def == nil || !def.inAscendZone(relX) {
		return false
	}
	_, attackN := tiLayerMonsterCounts(cfg, eng, 3, relX)
	if attackN >= 1 {
		tiLog("L3上升区 x=%d 攻击区怪物=%d 优先攻击", relX, attackN)
		tapAttack()
		return true
	}
	return false
}

func tiDoAscendStep(cfg *mapConfig, eng *yolo.Yolo, tag string, ti *TreasureIslandConfig, st *tiRunState, lay, relX int, target int) bool {
	if lay <= target {
		return false
	}
	def := ti.layerDef(lay)
	if def == nil {
		return false
	}
	if tiTryL3AscendZoneFight(cfg, eng, ti, lay, relX) {
		return true
	}
	if !tiInXRange(relX, def.AscendX, def.AscendXTolerance) {
		tiLog("上升 L%d→L%d: x未到 %d±%d (relX=%d)", lay, target, def.AscendX, def.AscendXTolerance, relX)
		st.beginAscendAlign()
		if st.ascendAlignTimedOut() {
			tiLog("上升对齐超%d秒 攻击长按清怪", tiAscendAlignTimeoutSec)
			keyHoldPress(attackKeyCode(), tiAscendAlignAttackHoldMinMs, tiAscendAlignAttackHoldMaxMs)
			st.resetAscendAlignTimer()
			return true
		}
		if tiHandleTransitionFight(cfg, eng, ti, st, lay, relX, "对齐前") {
			return true
		}
		tiAlignAscendX(cfg, eng, ti, st, lay, relX, def)
		return true
	}
	st.clearAscendAlign()
	if tiTryL3AscendZoneFight(cfg, eng, ti, lay, relX) {
		return true
	}
	tiLog("上升 L%d→L%d: 上瞬移+右瞬移 x=%d", lay, target, relX)
	tapUpTeleport()
	core.RandomSleep(ti.AscendWaitMsMin, ti.AscendWaitMsMax)
	if tiHandleTransitionFight(cfg, eng, ti, st, lay, relX, "上瞬移后") {
		return true
	}
	tapTeleportWithDirection(true)
	sleepAfterTeleport()
	core.Sleep(200)
	return true
}

func tiDoDescendStep(cfg *mapConfig, eng *yolo.Yolo, tag string, ti *TreasureIslandConfig, st *tiRunState, lay, relX int) bool {
	fromLay := lay
	if fromLay <= 0 {
		fromLay = st.descendTarget - 1
		if fromLay < 1 {
			fromLay = st.farmLayer
		}
	}
	tiLog("下落 L%d→L%d: 下瞬移 x=%d", fromLay, st.descendTarget, relX)
	tapDownTeleport()
	core.RandomSleep(ti.AscendWaitMsMin, ti.AscendWaitMsMax)
	core.Sleep(200)
	return true
}

func tiStartLeaveAndTransition(s *tiRunState) {
	tiStartLeaveLayer(s)
	switch s.farmLayer {
	case 3:
		s.afterL3Ascend = true
		s.startAscend(2)
	case 2:
		if s.afterL3Ascend {
			s.startAscend(1)
			s.afterL3Ascend = false
		} else {
			s.startDescend(s.farmLayer + 1)
		}
	default:
		s.startDescend(s.farmLayer + 1)
	}
}

func tiDoFarmStep(tag string, cfg *mapConfig, ti *TreasureIslandConfig, s *tiRunState, lay, relX, relY int) bool {
	if lay != s.farmLayer && lay > 0 {
		if s.offLayer != lay {
			s.offLayer = lay
			s.offLayerSince = time.Now()
		}
		offSec := int(time.Since(s.offLayerSince).Seconds())
		if offSec >= tiAdoptOffLayerMinSec {
			tiLog("打怪: 在L%d 目标L%d 意外换层 已停留%d秒 从L%d开始", lay, s.farmLayer, offSec, lay)
			s.enterFarmLayer(lay, ti)
			return true
		}
		if lay > s.farmLayer {
			tiLog("打怪: 在L%d 目标L%d 先上升 (已%d/%ds)", lay, s.farmLayer, offSec, tiAdoptOffLayerMinSec)
			tiStartLeaveLayer(s)
			s.startAscend(s.farmLayer)
			return true
		}
		tiLog("打怪: 在L%d 目标L%d 先下落 (已%d/%ds)", lay, s.farmLayer, offSec, tiAdoptOffLayerMinSec)
		tiStartLeaveLayer(s)
		s.startDescend(s.farmLayer)
		return true
	}
	s.offLayer = 0

	def := ti.layerDef(s.farmLayer)
	if def == nil {
		return false
	}

	if s.layerStayExceeded(def) {
		tiLog("L%d 停留>%ds 强制换层", s.farmLayer, def.MaxStaySec)
		tiStartLeaveAndTransition(s)
		return true
	}

	if s.farmLayer == 3 {
		return tiDoFarmPatrolStep(s, def, relX)
	}

	if s.minFarmStayLeft(ti) > 0 {
		return tiDoFarmPatrolStep(s, def, relX)
	}

	tiLog("L%d 最少停留%d秒已满 换层", s.farmLayer, s.minFarmStaySec(ti))
	tiStartLeaveAndTransition(s)
	return true
}

// SetupTreasureIslandMinimap 与 Play_抢夺宝物岛 使用相同的小地图黄点区域（调试用）。
func SetupTreasureIslandMinimap(mapAssetPath string) error {
	cfg, err := loadMapConfig(mapAssetPath)
	if err != nil {
		return err
	}
	if cfg.TreasureIsland == nil {
		return fmt.Errorf("抢夺宝物岛: 缺少 treasure_island 配置")
	}
	normalizeTreasureIslandConfig(cfg)
	cfg.TreasureIsland.applyDeleteYellowRegion()
	return nil
}

// ResetTreasureIslandMinimap 恢复默认小地图黄点检测区域。
func ResetTreasureIslandMinimap() {
	core.ClearMinimapYellowRegion()
}

// Play_抢夺宝物岛 三层循环：打怪→下落→上升，参数见 JSON treasure_island。
func Play_抢夺宝物岛(mapAssetPath string) error {
	cfg, err := loadMapConfig(mapAssetPath)
	if err != nil {
		return err
	}
	if cfg.TreasureIsland == nil {
		return fmt.Errorf("抢夺宝物岛: 缺少 treasure_island 配置")
	}
	ti := cfg.TreasureIsland
	normalizeTreasureIslandConfig(cfg)
	ti.applyDeleteYellowRegion()
	defer core.ClearMinimapYellowRegion()
	SetFarmLogTag(treasureIslandLogTag)

	labels := cfg.MonsterLabels
	if labels == "" {
		labels = core.DefaultFarmMonsterLabels
	}
	var eng *yolo.Yolo
	if paramPath, binPath, err := assets.InstallYoloOnDevice(); err == nil {
		eng = yolo.New("v8", 4, paramPath, binPath, labels)
	}

	StartFarmMaintainLoop(treasureIslandLogTag)
	defer StopFarmMaintainLoop()
	EnableFarmPeriodicLRJump()
	defer DisableFarmPeriodicLRJump()

	st := tiRunState{clearBag: newAutoClearBagState()}
	if st.clearBag.startupPending {
		tiLog("自动清包: 启动后优先到 x=[%d,%d] 清包", tiAutoClearBagXMin, tiAutoClearBagXMax)
	}
	st.enterFarmLayer(1, ti)
	tiLog("开始挂机 目标从第1层")
	logTreasureIslandLayerStay(ti)

	for {
		core.BlockWhileCaptchaHold()
		TickFarmMainThreadTasks()

		mx, my, wx, wy := detectYellowThenWorld()
		if mx < 0 || my < 0 || wx < 0 || wy < 0 {
			tiLog("第?层 | 怪物=0 需攻击=0 | %s (小地图未识别)", tiPhaseName(st.phase))
			continue
		}
		relX, relY := relativeToRef(mx, my, wx, wy)

		if tiTryDownJump(treasureIslandLogTag, ti, relY) {
			continue
		}
		if tiTrySpecialRightTeleport(relX, relY) {
			continue
		}

		lay := tiDetectLayer(ti, relY)
		st.refreshAscendZoneEnabled(ti, lay)
		if lay <= 0 {
			if st.phase == tiPhaseFarm {
				st.unknownLayerStreak++
				cooldownOk := st.lastUnknownJumpAt.IsZero() ||
					time.Since(st.lastUnknownJumpAt) >= tiUnknownLayerJumpCooldownSec*time.Second
				if st.unknownLayerStreak >= ti.UnknownLayerJumpStreak && cooldownOk {
					tiLog("连续%d次0层 右跳", st.unknownLayerStreak)
					tapJumpRight()
					st.lastUnknownJumpAt = time.Now()
					st.unknownLayerStreak = 0
					continue
				}
			}
		} else {
			st.unknownLayerStreak = 0
		}

		var monsterN, attackN int
		if st.phase != tiPhaseFarm {
			monsterN, attackN = tiLayerMonsterCounts(cfg, eng, lay, relX)
		}
		if st.phase == tiPhaseFarm {
			minSec := st.minFarmStaySec(ti)
			elapsed := int(time.Since(st.layerEnteredAt).Seconds())
			tiLog("第%d层 relX=%d relY=%d | 打怪中 已停留=%ds/%ds | 目标L%d",
				lay, relX, relY, elapsed, minSec, st.farmLayer)
		} else {
			phaseTarget := st.farmLayer
			switch st.phase {
			case tiPhaseAscend:
				phaseTarget = st.ascendTarget
			case tiPhaseDescend:
				phaseTarget = st.descendTarget
			}
			tiLog("第%d层 relX=%d relY=%d | 怪物=%d 需攻击=%d | %s 目标L%d",
				lay, relX, relY, monsterN, attackN, tiPhaseName(st.phase), phaseTarget)
		}

		switch st.phase {
		case tiPhaseAscend:
			if lay > 0 && lay <= st.ascendTarget {
				st.clearAscendAlign()
				st.enterFarmLayer(st.ascendTarget, ti)
				tiLog("上升完成 进入L%d打怪", st.ascendTarget)
				continue
			}
			tiDoAscendStep(cfg, eng, treasureIslandLogTag, ti, &st, lay, relX, st.ascendTarget)
		case tiPhaseDescend:
			if lay > 0 && lay >= st.descendTarget {
				st.enterFarmLayer(st.descendTarget, ti)
				tiLog("下落完成 进入L%d打怪", st.descendTarget)
				continue
			}
			tiDoDescendStep(cfg, eng, treasureIslandLogTag, ti, &st, lay, relX)
		default:
			if lay > 0 {
				if tryAlignTreasureIslandAutoClearBag(&st.clearBag, lay, st.farmLayer, relX) {
					continue
				}
				if tryAutoClearBagTreasureIsland(&st.clearBag, ti, lay, relX, relY) {
					continue
				}
				tiDoFarmStep(treasureIslandLogTag, cfg, ti, &st, lay, relX, relY)
			}
		}
	}
}
