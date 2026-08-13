package play

import (
	"app/assets"
	"app/core"
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/Dasongzi1366/AutoGo/yolo"
)

type cornerID int

const (
	cornerD cornerID = iota
	cornerC
	cornerB
	cornerA
)

var fashionRoute = []cornerID{cornerD, cornerC, cornerB, cornerA}

func nextCorner(c cornerID) cornerID {
	for i, x := range fashionRoute {
		if x == c {
			return fashionRoute[(i+1)%len(fashionRoute)]
		}
	}
	return cornerC
}

// Play_定制_时尚大道 两层图 D→C→B→A 循环挂机。
func Play_定制_时尚大道(mapAssetPath string) error {
	cfg, err := loadMapConfig(mapAssetPath)
	if err != nil {
		return err
	}
	if cfg.Ring == nil {
		return fmt.Errorf("时尚大道: 缺少 ring 配置")
	}
	cfg.Ring.normalize()
	SetFarmLogTag(fashionLogTag)
	applyMapMinimapRegions(cfg)
	defer clearMapMinimapRegions()

	var eng *yolo.Yolo
	if !cfg.DisableMonsterScan {
		labels := cfg.MonsterLabels
		if labels == "" {
			labels = core.DefaultFarmMonsterLabels
		}
		if paramPath, binPath, err := assets.InstallYoloOnDevice(); err == nil {
			eng = yolo.New("v8", 4, paramPath, binPath, labels)
			_ = paramPath
			_ = binPath
		}
	} else {
		if cfg.AfterTeleportAttackMax < cfg.AfterTeleportAttackMin {
			cfg.AfterTeleportAttackMax = cfg.AfterTeleportAttackMin
		}
		if cfg.AfterTeleportAttackMax == 0 && cfg.AfterTeleportAttackMin == 0 {
			cfg.AfterTeleportAttackMin = 1
			cfg.AfterTeleportAttackMax = 4
		}
		farmLog("模式: 关闭YOLO 边走边打 攻击%d～%d次", cfg.AfterTeleportAttackMin, cfg.AfterTeleportAttackMax)
	}

	StartFarmMaintainLoop(fashionLogTag)
	defer StopFarmMaintainLoop()

	for {
		core.BlockWhileCaptchaHold()
		TickFarmMainThreadTasks()
		mx, my, wx, wy := detectYellowThenWorld()
		if mx < 0 || my < 0 || wx < 0 || wy < 0 {
			farmLog("状态: 小地图未识别，等待 | %s", formatMinimapRelFail(mx, my, wx, wy))
			core.Sleep(200)
			continue
		}
		relX, relY := relativeToRef(mx, my, wx, wy)
		lay := detectRingLayer(cfg.Ring, relY)
		if lay == 0 {
			farmLog("状态: relY=%d 不在1/2层区间，等待", relY)
			core.Sleep(150)
			continue
		}

		cur := detectCorner(cfg.Ring, relX, relY, lay)
		tgt := nextCorner(cur)
		if cfg.DisableMonsterScan {
			farmLog("位置: relX=%d relY=%d lay=%d | 路线 %s | 无YOLO边走边打",
				relX, relY, lay, legName(cur, tgt))
		} else {
			monsters := countMonsters(cfg, eng, lay, relX)
			farmLog("位置: relX=%d relY=%d lay=%d | 路线 %s | 可攻击怪=%d | 攻击区=%s",
				relX, relY, lay, legName(cur, tgt), monsters, attackRegionLabel(cfg, lay, relX))
		}

		if tryApplyPositionRules(cfg, relX, relY) {
			continue
		}

		if !cfg.DisableMonsterScan && combatClearLayer(cfg, eng, lay, relX, "主循环清怪") {
			continue
		}

		leg := movementLeg(cfg, eng, cur, tgt, relX, relY, lay, cfg.Ring)
		if leg == legNeedHorizontal {
			horizontalTowardCorner(cfg, eng, cfg.Ring, relX, relY, lay, ringCorner(cfg.Ring, tgt), cfg.Ring.TeleportFarPx, cfg.Ring.WalkMs, cornerName(tgt))
		}
		core.Sleep(50)
	}
}

func loadMapConfig(path string) (*mapConfig, error) {
	data, err := assets.ConfigFile.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg mapConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

type moveLeg int

const (
	legDone moveLeg = iota
	legNeedHorizontal
)

func movementLeg(cfg *mapConfig, eng *yolo.Yolo, cur, tgt cornerID, relX, relY, lay int, ring *RingGraphConfig) moveLeg {
	far := ring.TeleportFarPx

	switch {
	// D → C：第二层直线（|dx|>far 瞬移+方向，否则长按走）
	case tgt == cornerC && lay == 2:
		farmLog("移动: D→C 二层横向")
		horizontalTowardCorner(cfg, eng, ring, relX, relY, lay, ring.C, far, ring.WalkMs, "C")
		return legDone

	// C → B：须在 C 点 x 走廊内，上+瞬移
	case tgt == cornerB:
		if lay == 2 {
			if !inCorner(relX, relY, lay, ring.C, ring) {
				farmLog("移动: C→B 未到C点，先横向到C")
				horizontalTowardCorner(cfg, eng, ring, relX, relY, lay, ring.C, far, ring.WalkMs, "C")
				return legDone
			}
			if combatBeforeMove(cfg, eng, lay, relX, "C→B上瞬移前") {
				return legDone
			}
			farmLog("移动: C→B 上+瞬移")
			relYBefore := relY
			tapUpTeleport()
			core.Sleep(500)
			if mx2, my2, wx2, wy2 := detectYellowThenWorld(); mx2 >= 0 && my2 >= 0 && wx2 >= 0 && wy2 >= 0 {
				_, relYAfter := relativeToRef(mx2, my2, wx2, wy2)
				if iabs(relYAfter-relYBefore) < 2 {
					farmLog("移动: C→B 层Y未变(%d→%d)，补左瞬移", relYBefore, relYAfter)
					teleportLeftAction()
					core.Sleep(80 + rand.Intn(60))
				}
			}
			combatAfterTeleport(cfg, eng, ring, "C→B上瞬移后")
			return legDone
		}
		if lay == 1 {
			farmLog("移动: 已在1层，横向到B")
			horizontalTowardCorner(cfg, eng, ring, relX, relY, lay, ring.B, far, ring.WalkMs, "B")
			return legDone
		}
		return legNeedHorizontal

	// B → A：第一层向左（B 在右 x≈70~100，A 在左 x≈-50~-40）
	case tgt == cornerA && lay == 1:
		farmLog("移动: B→A 一层向左")
		horizontalTowardCorner(cfg, eng, ring, relX, relY, lay, ring.A, far, ring.WalkMs, "A")
		return legDone

	// A → D：A 下跳落二层，再直线到 D
	case tgt == cornerD:
		if lay == 1 && inCorner(relX, relY, lay, ring.A, ring) {
			if combatBeforeMove(cfg, eng, lay, relX, "A→D下跳前") {
				return legDone
			}
			farmLog("移动: A→D 下键+跳跃落二层")
			tapDownJump()
			core.Sleep(300)
			if cfg.DisableMonsterScan {
				combatAfterTeleportQuick(cfg, eng, lay, relX, "A→D下跳后")
			}
			return legDone
		}
		if lay == 2 {
			farmLog("移动: A→D 二层横向到D")
			horizontalTowardCorner(cfg, eng, ring, relX, relY, lay, ring.D, far, ring.WalkMs, "D")
			return legDone
		}
		return legNeedHorizontal

	default:
		if lay > 0 {
			horizontalTowardCorner(cfg, eng, ring, relX, relY, lay, ringCorner(ring, tgt), far, ring.WalkMs, cornerName(tgt))
			return legDone
		}
	}
	return legNeedHorizontal
}

func detectRingLayer(ring *RingGraphConfig, relY int) int {
	if ring == nil {
		return 0
	}
	if matchRange(relY, ring.Layer1YMin, ring.Layer1YMax) {
		return 1
	}
	if matchRange(relY, ring.Layer2YMin, ring.Layer2YMax) {
		return 2
	}
	return 0
}

func detectCorner(ring *RingGraphConfig, relX, relY, lay int) cornerID {
	if inCorner(relX, relY, lay, ring.D, ring) {
		return cornerD
	}
	if inCorner(relX, relY, lay, ring.C, ring) {
		return cornerC
	}
	if inCorner(relX, relY, lay, ring.B, ring) {
		return cornerB
	}
	if inCorner(relX, relY, lay, ring.A, ring) {
		return cornerA
	}
	if lay == 2 {
		// D→C 走廊：relX 在 D 右边界以东、未到 C 角点前，视为在 D→C 段（当前=D，目标=C）
		if relX > ring.D.XMax {
			return cornerD
		}
		return cornerC
	}
	if lay == 1 {
		// B→A 走廊：relX 在 A 右边界以东、未到 B 角点前，视为在 B→A 段（当前=B，目标=A）
		if relX >= ring.B.XMin {
			return cornerB
		}
		if relX > ring.A.XMax {
			return cornerB
		}
		return cornerA
	}
	return cornerD
}

func ringCorner(ring *RingGraphConfig, c cornerID) RingCorner {
	switch c {
	case cornerA:
		return ring.A
	case cornerB:
		return ring.B
	case cornerC:
		return ring.C
	case cornerD:
		return ring.D
	default:
		return ring.D
	}
}

func inCorner(relX, relY, lay int, c RingCorner, ring *RingGraphConfig) bool {
	if c.Layer != lay {
		return false
	}
	var yMin, yMax int
	if lay == 1 {
		yMin, yMax = ring.Layer1YMin, ring.Layer1YMax
	} else if lay == 2 {
		yMin, yMax = ring.Layer2YMin, ring.Layer2YMax
	} else {
		return false
	}
	if !matchRange(relY, yMin, yMax) {
		return false
	}
	return matchRange(relX, c.XMin, c.XMax)
}

func matchRange(v, lo, hi int) bool {
	if lo > hi {
		lo, hi = hi, lo
	}
	return v >= lo && v <= hi
}

func horizontalTowardCorner(cfg *mapConfig, eng *yolo.Yolo, ring *RingGraphConfig, relX, relY, lay int, c RingCorner, farPx, walkMs int, tgtName string) {
	if inCorner(relX, relY, lay, c, ring) {
		farmLog("移动: 已在%s角点 x[%d,%d]，无需横移", tgtName, c.XMin, c.XMax)
		return
	}
	var dx int
	goRight := false
	if relX < c.XMin {
		dx = c.XMin - relX
		goRight = true
	} else if relX > c.XMax {
		dx = relX - c.XMax
		goRight = false
	} else {
		return
	}
	if combatBeforeMove(cfg, eng, lay, relX, "横移→"+tgtName+"前") {
		return
	}
	if dx <= farPx {
		ms := jitterWalkMs(walkMs)
		if dx < 4 {
			ms = ms * 60 / 100
		}
		dir := "左"
		if goRight {
			dir = "右"
		}
		farmLog("移动: 微调走向%s %s dx=%d ms=%d", tgtName, dir, dx, ms)
		walkHoldMs(goRight, ms)
		if cfg != nil && cfg.DisableMonsterScan {
			combatAfterTeleportQuick(cfg, eng, lay, relX, "横移走→"+tgtName+"后")
		}
		return
	}
	dir := "左瞬移"
	if goRight {
		dir = "右瞬移"
	}
	farmLog("移动: %s 朝%s x[%d,%d] dx=%d", dir, tgtName, c.XMin, c.XMax, dx)
	tapTeleportWithDirection(goRight)
	sleepAfterTeleport()
	combatAfterTeleportQuick(cfg, eng, lay, relX, dir+"后")
}
