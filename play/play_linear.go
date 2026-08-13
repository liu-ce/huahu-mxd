package play

import (
	"app/assets"
	"app/core"
	"fmt"
	"math/rand"
	"strings"

	"github.com/Dasongzi1366/AutoGo/yolo"
)

const linearLogTag = "[直线图]"

func linearLog(format string, args ...interface{}) {
	fmt.Printf(linearLogTag+" "+format+"\n", args...)
}

type linearBounds struct {
	Layer      int
	YMin, YMax int
	XMin, XMax int
}

func linearBoundsFromConfig(cfg *mapConfig) (linearBounds, bool) {
	if cfg == nil {
		return linearBounds{}, false
	}
	for _, r := range cfg.Rules {
		if strings.TrimSpace(r.Action) != "" {
			continue
		}
		if r.XMax <= r.XMin {
			continue
		}
		layer := r.Layer
		if layer <= 0 {
			layer = 1
		}
		return linearBounds{
			Layer: layer,
			YMin:  r.YMin,
			YMax:  r.YMax,
			XMin:  r.XMin,
			XMax:  r.XMax,
		}, true
	}
	return linearBounds{}, false
}

func normalizeLinearConfig(cfg *mapConfig) {
	if cfg == nil {
		return
	}
	if cfg.MinMonstersToAttack <= 0 {
		cfg.MinMonstersToAttack = 1
	}
	if cfg.SkipMonsterScanRelXAbove <= defaultSkipMonsterScanRelXAbove {
		cfg.SkipMonsterScanRelXAbove = 9999
	}
	if cfg.DisableMonsterScan {
		if cfg.AfterTeleportAttackMax < cfg.AfterTeleportAttackMin {
			cfg.AfterTeleportAttackMax = cfg.AfterTeleportAttackMin
		}
		if cfg.AfterTeleportAttackMax == 0 && cfg.AfterTeleportAttackMin == 0 {
			cfg.AfterTeleportAttackMin = 0
			cfg.AfterTeleportAttackMax = 8
		}
	}
}

func linearAfterTeleportAttackCount(cfg *mapConfig) int {
	minN, maxN := 0, 8
	if cfg != nil && cfg.DisableMonsterScan {
		minN, maxN = cfg.AfterTeleportAttackMin, cfg.AfterTeleportAttackMax
	}
	if maxN < minN {
		maxN = minN
	}
	if maxN <= minN {
		return minN
	}
	return minN + rand.Intn(maxN-minN+1)
}

func attackTimes(n int) {
	if n <= 0 {
		return
	}
	for i := 0; i < n; i++ {
		tapAttackOnce()
		if i < n-1 {
			sleepAttackComboInterval()
		}
	}
}

const (
	linearPatrolWalkChance       = 35
	linearPatrolJumpChance       = 5
	linearPatrolPauseChance      = 5
	linearPatrolRandomTurnChance = 3
)

func linearPatrolMargin(bounds linearBounds) int {
	span := bounds.XMax - bounds.XMin
	if span <= 20 {
		return 4
	}
	if span <= 40 {
		return 8
	}
	return 12
}

func linearDoAttack(cfg *mapConfig, eng *yolo.Yolo, lay, relX int, tag string) {
	if cfg != nil && cfg.DisableMonsterScan {
		attacks := linearAfterTeleportAttackCount(cfg)
		if attacks > 0 {
			linearLog("攻击: %s 空格×%d", tag, attacks)
			attackTimes(attacks)
		}
		return
	}
	combatAfterTeleportQuick(cfg, eng, lay, relX, tag)
}

func linearPatrolStep(cfg *mapConfig, eng *yolo.Yolo, bounds linearBounds, lay, relX int, goRight bool) bool {
	margin := linearPatrolMargin(bounds)

	if relX < bounds.XMin {
		linearLog("出界: relX=%d<%d 右瞬移回区", relX, bounds.XMin)
		tapTeleportWithDirection(true)
		sleepAfterTeleport()
		linearDoAttack(cfg, eng, lay, relX, "回区")
		return true
	}
	if relX > bounds.XMax {
		linearLog("出界: relX=%d>%d 左瞬移回区", relX, bounds.XMax)
		tapTeleportWithDirection(false)
		sleepAfterTeleport()
		linearDoAttack(cfg, eng, lay, relX, "回区")
		return false
	}

	if goRight && relX >= bounds.XMax-margin {
		linearLog("近右界 relX=%d 改向左", relX)
		core.Sleep(80)
		return false
	}
	if !goRight && relX <= bounds.XMin+margin {
		linearLog("近左界 relX=%d 改向右", relX)
		core.Sleep(80)
		return true
	}

	if rand.Intn(100) < linearPatrolRandomTurnChance {
		goRight = !goRight
		if goRight {
			linearLog("随机换向: 改向右 relX=%d", relX)
		} else {
			linearLog("随机换向: 改向左 relX=%d", relX)
		}
	}

	dir := "左"
	if goRight {
		dir = "右"
	}

	roll := rand.Intn(100)
	switch {
	case roll < linearPatrolWalkChance:
		linearLog("移动: %s走+攻击 relX=%d", dir, relX)
		if goRight {
			faceRight()
		} else {
			faceLeft()
		}
		walkHoldMs(goRight, patrolFarmWalkMs())
		linearDoAttack(cfg, eng, lay, relX, dir+"走")
	case roll < linearPatrolWalkChance+linearPatrolJumpChance:
		linearLog("移动: %s跳+攻击 relX=%d", dir, relX)
		if goRight {
			tapJumpRight()
		} else {
			tapJumpLeft()
		}
		core.RandomSleep(200, 400)
		linearDoAttack(cfg, eng, lay, relX, dir+"跳")
	case roll < linearPatrolWalkChance+linearPatrolJumpChance+linearPatrolPauseChance:
		pauseMs := 400 + rand.Intn(501)
		linearLog("移动: 停顿%dms+攻击 relX=%d", pauseMs, relX)
		core.Sleep(pauseMs)
		linearDoAttack(cfg, eng, lay, relX, "停顿")
	default:
		linearLog("移动: %s瞬移 relX=%d", dir, relX)
		tapTeleportWithDirection(goRight)
		sleepAfterTeleport()
		linearDoAttack(cfg, eng, lay, relX, dir+"瞬移")
	}

	if rand.Intn(100) < 30 {
		core.RandomSleep(50, 180)
	}
	return goRight
}

func detectLinearLayer(bounds linearBounds, relY int) int {
	if bounds.YMin == 0 && bounds.YMax == 0 {
		return bounds.Layer
	}
	if matchRange(relY, bounds.YMin, bounds.YMax) {
		return bounds.Layer
	}
	return 0
}

// Play_直线图 单层左右往返：35%走路 + 随机跳/停顿/换向，其余瞬移；有 YOLO 时先清怪。
func Play_直线图(mapAssetPath string) error {
	cfg, err := loadMapConfig(mapAssetPath)
	if err != nil {
		return err
	}
	bounds, ok := linearBoundsFromConfig(cfg)
	if !ok {
		return fmt.Errorf("直线图: 缺少巡逻边界 rules（无 action 的 x_min/x_max）")
	}
	normalizeLinearConfig(cfg)
	applyMapMinimapRegions(cfg)
	defer clearMapMinimapRegions()
	SetFarmLogTag(linearLogTag)

	var eng *yolo.Yolo
	if !cfg.DisableMonsterScan {
		labels := cfg.MonsterLabels
		if labels == "" {
			labels = core.DefaultFarmMonsterLabels
		}
		if paramPath, binPath, err := assets.InstallYoloOnDevice(); err == nil {
			eng = yolo.New("v8", 4, paramPath, binPath, labels)
		}
	}

	StartFarmMaintainLoop(linearLogTag)
	defer StopFarmMaintainLoop()

	goRight := true

	if cfg.DisableMonsterScan {
		linearLog("开始挂机 边界 x[%d,%d] y[%d,%d] layer=%d 无YOLO 走35%%/跳5%%/停5%%/瞬移52%% 攻击%d～%d次",
			bounds.XMin, bounds.XMax, bounds.YMin, bounds.YMax, bounds.Layer,
			cfg.AfterTeleportAttackMin, cfg.AfterTeleportAttackMax)
	} else {
		linearLog("开始挂机 边界 x[%d,%d] y[%d,%d] layer=%d 走35%%/跳5%%/停5%%/瞬移52%%",
			bounds.XMin, bounds.XMax, bounds.YMin, bounds.YMax, bounds.Layer)
	}

	for {
		core.BlockWhileCaptchaHold()
		TickFarmMainThreadTasks()

		mx, my, wx, wy := detectYellowThenWorld()
		if mx < 0 || my < 0 || wx < 0 || wy < 0 {
			linearLog("状态: 小地图未识别，等待")
			core.Sleep(200)
			continue
		}
		relX, relY := relativeToRef(mx, my, wx, wy)
		lay := detectLinearLayer(bounds, relY)
		if lay == 0 {
			linearLog("状态: relY=%d 不在层区间 [%d,%d]，等待", relY, bounds.YMin, bounds.YMax)
			core.Sleep(150)
			continue
		}

		if tryApplyPositionRules(cfg, relX, relY) {
			continue
		}

		if !cfg.DisableMonsterScan && combatClearLayer(cfg, eng, lay, relX, "巡逻清怪") {
			continue
		}

		goRight = linearPatrolStep(cfg, eng, bounds, lay, relX, goRight)

		if cfg.AttackLoopPauseMs > 0 {
			core.Sleep(cfg.AttackLoopPauseMs)
		}
	}
}
