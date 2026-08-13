package play

import (
	"app/core"
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	"github.com/Dasongzi1366/AutoGo/yolo"
)

const (
	defaultMinMonsterScore          = 0.35
	defaultMinMonstersToAttack      = 2
	defaultSkipMonsterScanRelXAbove = 95
	combatClearMaxPass              = 24 // 原 12 加倍
	teleportSweepRoundsMin          = 2
	teleportSweepIntervalMs         = 100
)

func skipMonsterScanRelXMax(cfg *mapConfig) int {
	if cfg != nil && cfg.SkipMonsterScanRelXAbove > 0 {
		return cfg.SkipMonsterScanRelXAbove
	}
	return defaultSkipMonsterScanRelXAbove
}

func shouldScanMonsters(cfg *mapConfig, relX int) bool {
	if cfg != nil && cfg.DisableMonsterScan {
		return false
	}
	return relX <= skipMonsterScanRelXMax(cfg)
}

func attackWithoutYolo(cfg *mapConfig, phase string) {
	n := linearAfterTeleportAttackCount(cfg)
	if n <= 0 {
		farmLog("战斗: [%s] 无YOLO 跳过(攻击次数=0)", phase)
		return
	}
	farmLog("战斗: [%s] 无YOLO 空格×%d", phase, n)
	attackTimes(n)
}

func regionForLayer(regions map[string][]int, lay int) (x1, y1, x2, y2 int, ok bool) {
	if len(regions) == 0 {
		return 0, 0, 0, 0, false
	}
	r, ok := regions[strconv.Itoa(lay)]
	if !ok || len(r) < 4 {
		return 0, 0, 0, 0, false
	}
	return r[0], r[1], r[2], r[3], true
}

func yoloScanRegionForLayer(cfg *mapConfig, lay, relX int) (x1, y1, x2, y2 int, ok bool) {
	if cfg == nil {
		return 0, 0, 0, 0, false
	}
	if cfg.Type == MapTypeTreasureIsland && cfg.TreasureIsland != nil {
		if x1, y1, x2, y2, ok := cfg.TreasureIsland.yoloScanRegion(lay, relX); ok {
			return x1, y1, x2, y2, true
		}
	}
	return regionForLayer(cfg.YoloRegions, lay)
}

const (
	attackRelXLeftThreshold  = -70
	attackRelXRightThreshold = 90
)

// attackRegionForLayerAndRelX 时尚大道按 relX 用固定攻击区；直线图不过滤；抢夺宝物岛上升区有专用攻击区。
func attackRegionForLayerAndRelX(cfg *mapConfig, lay, relX int) (x1, y1, x2, y2 int, ok bool) {
	if cfg == nil || cfg.Type == MapTypeLinear {
		return 0, 0, 0, 0, false
	}
	if cfg.Type == MapTypeTreasureIsland && cfg.TreasureIsland != nil {
		if x1, y1, x2, y2, ok := cfg.TreasureIsland.attackRegion(lay, relX); ok {
			return x1, y1, x2, y2, true
		}
	}
	if cfg.Type == MapTypeFashionAvenue {
		switch lay {
		case 1:
			if relX > attackRelXRightThreshold {
				return 716, 161, 1234, 466, true
			}
			if relX < attackRelXLeftThreshold {
				return 53, 106, 627, 467, true
			}
		case 2:
			if relX < attackRelXLeftThreshold {
				return 78, 291, 564, 578, true
			}
			if relX > attackRelXRightThreshold {
				return 685, 333, 1220, 613, true
			}
		}
	}
	return regionForLayer(cfg.AttackRegions, lay)
}

// resultOverlapsRegion 检测框与区域是否有任意像素重叠。
func resultOverlapsRegion(r yolo.Result, x1, y1, x2, y2 int) bool {
	ax1, ay1 := r.X, r.Y
	ax2, ay2 := r.X+r.Width, r.Y+r.Height
	if r.Width <= 0 || r.Height <= 0 {
		return r.CenterX >= x1 && r.CenterX < x2 && r.CenterY >= y1 && r.CenterY < y2
	}
	return ax1 < x2 && ax2 > x1 && ay1 < y2 && ay2 > y1
}

func filterMonstersInAttackRegion(results []yolo.Result, x1, y1, x2, y2 int) []yolo.Result {
	if len(results) == 0 {
		return nil
	}
	out := make([]yolo.Result, 0, len(results))
	for _, r := range results {
		if resultOverlapsRegion(r, x1, y1, x2, y2) {
			out = append(out, r)
		}
	}
	return out
}

func minMonsterScore(cfg *mapConfig) float64 {
	if cfg != nil && cfg.MinMonsterScore > 0 {
		return cfg.MinMonsterScore
	}
	return defaultMinMonsterScore
}

func minMonstersToAttack(cfg *mapConfig) int {
	if cfg != nil && cfg.MinMonstersToAttack > 0 {
		return cfg.MinMonstersToAttack
	}
	return defaultMinMonstersToAttack
}

func filterMonsters(cfg *mapConfig, results []yolo.Result) []yolo.Result {
	if len(results) == 0 {
		return nil
	}
	minScore := minMonsterScore(cfg)
	allowAll := cfg == nil || len(cfg.MonsterAllowlist) == 0
	allow := map[string]struct{}{}
	if !allowAll {
		for _, l := range cfg.MonsterAllowlist {
			allow[strings.TrimSpace(l)] = struct{}{}
		}
	}
	out := make([]yolo.Result, 0, len(results))
	for _, r := range results {
		if r.Score < minScore {
			continue
		}
		label := strings.TrimSpace(r.Label)
		if allowAll {
			r.Label = label
			out = append(out, r)
			continue
		}
		if _, ok := allow[label]; ok {
			r.Label = label
			out = append(out, r)
		}
	}
	return out
}

func formatYoloMonsterList(results []yolo.Result) string {
	if len(results) == 0 {
		return "无"
	}
	parts := make([]string, 0, len(results))
	for _, r := range results {
		parts = append(parts, fmt.Sprintf("%s@%.2f(%d,%d)", strings.TrimSpace(r.Label), r.Score, r.CenterX, r.CenterY))
	}
	return strings.Join(parts, ", ")
}

func detectMonsters(cfg *mapConfig, eng *yolo.Yolo, lay, relX int) (attackable, raw, labeled []yolo.Result) {
	if eng == nil {
		return nil, nil, nil
	}
	sx1, sy1, sx2, sy2, ok := yoloScanRegionForLayer(cfg, lay, relX)
	if !ok {
		return nil, nil, nil
	}
	raw = eng.Detect(sx1, sy1, sx2, sy2, 0)
	labeled = filterMonsters(cfg, raw)
	ax1, ay1, ax2, ay2, attackOK := attackRegionForLayerAndRelX(cfg, lay, relX)
	if !attackOK {
		return labeled, raw, labeled
	}
	return filterMonstersInAttackRegion(labeled, ax1, ay1, ax2, ay2), raw, labeled
}

// countMonsters 当前层过滤后的怪物数量；relX 超过配置阈值时不检测，返回 0。
func countMonsters(cfg *mapConfig, eng *yolo.Yolo, lay, relX int) int {
	if !shouldScanMonsters(cfg, relX) || eng == nil || lay <= 0 {
		return 0
	}
	rs, _, _ := detectMonsters(cfg, eng, lay, relX)
	return len(rs)
}

func attackFaceSettleSleep(cfg *mapConfig) {
	if cfg != nil && cfg.AttackFaceSettleMs == 0 {
		return
	}
	minMs, maxMs := 50, 120
	if cfg != nil && cfg.AttackFaceSettleMs > 0 {
		minMs = cfg.AttackFaceSettleMs
		maxMs = cfg.AttackFaceSettleMs + 30
	}
	core.RandomSleep(minMs, maxMs)
}

func attackOneMonster(cfg *mapConfig, idx int, r yolo.Result) {
	face := "右"
	if r.CenterX < 640 {
		faceLeft()
		face = "左"
	} else {
		faceRight()
	}
	attackFaceSettleSleep(cfg)
	tapAttack()
	farmLog("战斗: 攻击 #%d %s score=%.2f center=(%d,%d) 朝%s 长按攻击",
		idx, r.Label, r.Score, r.CenterX, r.CenterY, face)
}

// combatBeforeMove 移动/瞬移前按层 YOLO 扫怪；relX>95 时不扫。
func combatBeforeMove(cfg *mapConfig, eng *yolo.Yolo, lay, relX int, phase string) bool {
	if !shouldScanMonsters(cfg, relX) {
		farmLog("战斗: [%s] 跳过(relX=%d>%d)", phase, relX, skipMonsterScanRelXMax(cfg))
		return false
	}
	return combatClearLayer(cfg, eng, lay, relX, phase)
}

func combatRegionDesc(cfg *mapConfig, lay, relX int) string {
	label := attackRegionLabel(cfg, lay, relX)
	if cfg != nil && cfg.Type == MapTypeFashionAvenue {
		if x1, y1, x2, y2, ok := attackRegionForLayerAndRelX(cfg, lay, relX); ok {
			return label + formatRegion(x1, y1, x2, y2)
		}
	}
	return label
}

func combatAttackWave(cfg *mapConfig, eng *yolo.Yolo, lay, relX int, phase string) bool {
	attackable, raw, labeled := detectMonsters(cfg, eng, lay, relX)
	regionStr := combatRegionDesc(cfg, lay, relX)
	minAtk := minMonstersToAttack(cfg)
	if len(attackable) < minAtk {
		farmLog("战斗: [%s] 不打 YOLO检出=%d 白名单=%d 可打=%d(需>=%d) | %s",
			phase, len(raw), len(labeled), len(attackable), minAtk, regionStr)
		if len(raw) > 0 {
			farmLog("战斗: [%s] YOLO明细: %s", phase, formatYoloMonsterList(raw))
			if len(labeled) > 0 && len(labeled) != len(raw) {
				farmLog("战斗: [%s] 白名单明细: %s", phase, formatYoloMonsterList(labeled))
			}
		}
		return false
	}
	farmLog("战斗: [%s] 开打 YOLO检出=%d 白名单=%d 可打=%d只 | %s",
		phase, len(raw), len(labeled), len(attackable), regionStr)
	farmLog("战斗: [%s] 目标: %s", phase, formatYoloMonsterList(attackable))
	for i, r := range attackable {
		attackOneMonster(cfg, i, r)
	}
	return true
}

// combatAfterTeleportQuick 左右/横向瞬移后：单次 YOLO，够数量才打，避免多轮扫怪卡顿。
func combatAfterTeleportQuick(cfg *mapConfig, eng *yolo.Yolo, lay, relX int, phase string) {
	if cfg != nil && cfg.DisableMonsterScan {
		attackWithoutYolo(cfg, phase)
		return
	}
	if eng == nil || lay <= 0 {
		farmLog("战斗: [%s] 跳过(YOLO未就绪或lay=0)", phase)
		return
	}
	if !shouldScanMonsters(cfg, relX) {
		farmLog("战斗: [%s] 跳过(relX=%d)", phase, relX)
		return
	}
	if combatAttackWave(cfg, eng, lay, relX, phase) {
		core.Sleep(teleportSweepIntervalMs)
	}
}

// combatAfterTeleport 上瞬移等：2～3 轮扫怪，每轮「检测→有怪就打→100ms→再检测」直到无怪。
func combatAfterTeleport(cfg *mapConfig, eng *yolo.Yolo, ring *RingGraphConfig, phase string) {
	if cfg != nil && cfg.DisableMonsterScan {
		attackWithoutYolo(cfg, phase)
		return
	}
	if eng == nil || ring == nil {
		return
	}
	mx, my, wx, wy := detectYellowThenWorld()
	if mx < 0 || my < 0 || wx < 0 || wy < 0 {
		farmLog("战斗: [%s] 跳过(小地图未识别)", phase)
		return
	}
	relX, relY := relativeToRef(mx, my, wx, wy)
	lay := detectRingLayer(ring, relY)
	if lay == 0 || !shouldScanMonsters(cfg, relX) {
		farmLog("战斗: [%s] 跳过 lay=%d relX=%d", phase, lay, relX)
		return
	}
	sweepRounds := teleportSweepRoundsMin + rand.Intn(2) // 2～3
	farmLog("战斗: [%s] 多轮扫怪 %d轮 lay=%d", phase, sweepRounds, lay)
	for round := 0; round < sweepRounds; round++ {
		if round > 0 {
			core.Sleep(teleportSweepIntervalMs)
		}
		roundFought := false
		for wave := 0; wave < combatClearMaxPass; wave++ {
			if !combatAttackWave(cfg, eng, lay, relX, fmt.Sprintf("%s R%d", phase, round+1)) {
				break
			}
			roundFought = true
			core.Sleep(teleportSweepIntervalMs)
		}
		if round == 0 && !roundFought {
			farmLog("战斗: [%s] 首轮无怪，结束", phase)
			return
		}
	}
	farmLog("战斗: [%s] 扫怪结束", phase)
}

// combatClearLayer 有怪则空格攻击，100ms 后再扫，直到清完或达上限；relX 超过配置阈值时不扫。
func combatClearLayer(cfg *mapConfig, eng *yolo.Yolo, lay, relX int, phase string) bool {
	if !shouldScanMonsters(cfg, relX) {
		return false
	}
	if eng == nil || lay <= 0 {
		farmLog("战斗: [%s] 跳过(YOLO未就绪)", phase)
		return false
	}
	fought := false
	for pass := 0; pass < combatClearMaxPass; pass++ {
		if !combatAttackWave(cfg, eng, lay, relX, fmt.Sprintf("%s#%d", phase, pass+1)) {
			if pass == 0 {
				return false
			}
			return fought
		}
		fought = true
		core.Sleep(teleportSweepIntervalMs)
	}
	farmLog("战斗: [%s] 达清怪轮次上限", phase)
	return fought
}
