package play

import (
	"app/assets"
	"app/core"
	"fmt"
	"time"

	"github.com/Dasongzi1366/AutoGo/yolo"
)

// YoloDebugRegion 调试用识别区域；x1,y1,x2,y2 为 0,0,0,0 时表示全屏（与 yolo.Detect 约定一致）。
type YoloDebugRegion struct {
	Label          string
	X1, Y1, X2, Y2 int
}

// YoloScanAttackPair 扫描区 Detect + 攻击区 box 重叠过滤（与挂机 yolo_regions + attack_regions 一致）。
type YoloScanAttackPair struct {
	ScanX1, ScanY1, ScanX2, ScanY2         int
	AttackX1, AttackY1, AttackX2, AttackY2 int
}

func detectMonstersInRegion(cfg *mapConfig, eng *yolo.Yolo, x1, y1, x2, y2 int) (rawN, filteredN int, filtered []yolo.Result, detectMs int64) {
	if eng == nil {
		return 0, 0, nil, 0
	}
	start := time.Now()
	raw := eng.Detect(x1, y1, x2, y2, 0)
	detectMs = time.Since(start).Milliseconds()
	filtered = filterMonsters(cfg, raw)
	return len(raw), len(filtered), filtered, detectMs
}

// detectScanThenAttackRegion 大区域 Detect + 标签/score 过滤 + 攻击区重叠过滤（与挂机 combat 一致）。
func detectScanThenAttackRegion(cfg *mapConfig, eng *yolo.Yolo, scanX1, scanY1, scanX2, scanY2, atkX1, atkY1, atkX2, atkY2 int, useAttack bool) (rawN, labelN, attackN int, attackRs []yolo.Result, detectMs int64) {
	rawN, labelN, rs, detectMs := detectMonstersInRegion(cfg, eng, scanX1, scanY1, scanX2, scanY2)
	if !useAttack {
		return rawN, labelN, labelN, rs, detectMs
	}
	attackRs = filterMonstersInAttackRegion(rs, atkX1, atkY1, atkX2, atkY2)
	return rawN, labelN, len(attackRs), attackRs, detectMs
}

func printRegionScan(label string, x1, y1, x2, y2 int, rawN, monsters int, rs []yolo.Result, detectMs int64) {
	fmt.Printf("[yolo-test] [%s] region=[%d,%d,%d,%d] raw=%d monsters=%d yolo_ms=%d\n",
		label, x1, y1, x2, y2, rawN, monsters, detectMs)
	for i, r := range rs {
		fmt.Printf("[yolo-test] [%s]   #%d %s score=%.3f center=(%d,%d) box=(%d,%d,%d,%d)\n",
			label, i, r.Label, r.Score, r.CenterX, r.CenterY,
			r.X, r.Y, r.X+r.Width, r.Y+r.Height)
	}
}

// RunYoloMultiRegionDebugLoop 每轮依次扫描多个区域；scanAttack 非 nil 时额外打印扫描+攻击区结果。
func RunYoloMultiRegionDebugLoop(mapAssetPath string, intervalMs int, regions []YoloDebugRegion, scanAttack *YoloScanAttackPair) error {
	if len(regions) == 0 {
		return fmt.Errorf("yolo 调试: 未指定区域")
	}
	cfg, err := loadMapConfig(mapAssetPath)
	if err != nil {
		return err
	}
	labels := cfg.MonsterLabels
	if labels == "" {
		labels = core.DefaultFarmMonsterLabels
	}
	paramPath, binPath, err := assets.InstallYoloOnDevice()
	if err != nil {
		return fmt.Errorf("YOLO 安装失败: %w", err)
	}
	eng := yolo.New("v8", 4, paramPath, binPath, labels)
	if eng == nil {
		return fmt.Errorf("YOLO 初始化失败")
	}
	if intervalMs <= 0 {
		intervalMs = 500
	}
	fmt.Printf("[yolo-test] regions=%d interval=%dms allowlist=%v min_score=%.2f\n",
		len(regions), intervalMs, cfg.MonsterAllowlist, minMonsterScore(cfg))
	for _, reg := range regions {
		fmt.Printf("[yolo-test]   %s: [%d,%d,%d,%d]\n", reg.Label, reg.X1, reg.Y1, reg.X2, reg.Y2)
	}
	if scanAttack != nil {
		fmt.Printf("[yolo-test]   scan+attack: scan=[%d,%d,%d,%d] attack=[%d,%d,%d,%d]\n",
			scanAttack.ScanX1, scanAttack.ScanY1, scanAttack.ScanX2, scanAttack.ScanY2,
			scanAttack.AttackX1, scanAttack.AttackY1, scanAttack.AttackX2, scanAttack.AttackY2)
	}
	tick := 0
	for {
		tick++
		fmt.Printf("[yolo-test] ===== tick %d =====\n", tick)
		for _, reg := range regions {
			rawN, monsters, rs, detectMs := detectMonstersInRegion(cfg, eng, reg.X1, reg.Y1, reg.X2, reg.Y2)
			printRegionScan(reg.Label, reg.X1, reg.Y1, reg.X2, reg.Y2, rawN, monsters, rs, detectMs)
		}
		if scanAttack != nil {
			rawN, labelN, attackN, rs, detectMs := detectScanThenAttackRegion(cfg, eng,
				scanAttack.ScanX1, scanAttack.ScanY1, scanAttack.ScanX2, scanAttack.ScanY2,
				scanAttack.AttackX1, scanAttack.AttackY1, scanAttack.AttackX2, scanAttack.AttackY2, true)
			fmt.Printf("[yolo-test] [scan+attack] scan=[%d,%d,%d,%d] attack=[%d,%d,%d,%d] raw=%d label=%d attack=%d yolo_ms=%d\n",
				scanAttack.ScanX1, scanAttack.ScanY1, scanAttack.ScanX2, scanAttack.ScanY2,
				scanAttack.AttackX1, scanAttack.AttackY1, scanAttack.AttackX2, scanAttack.AttackY2,
				rawN, labelN, attackN, detectMs)
			for i, r := range rs {
				fmt.Printf("[yolo-test] [scan+attack]   #%d %s score=%.3f center=(%d,%d) box=(%d,%d,%d,%d)\n",
					i, r.Label, r.Score, r.CenterX, r.CenterY,
					r.X, r.Y, r.X+r.Width, r.Y+r.Height)
			}
		}
		core.Sleep(intervalMs)
	}
}

// RunYoloRegionDebugLoop 单区域调试（兼容旧调用）。
func RunYoloRegionDebugLoop(mapAssetPath string, x1, y1, x2, y2, intervalMs int) error {
	return RunYoloMultiRegionDebugLoop(mapAssetPath, intervalMs, []YoloDebugRegion{{
		Label: "single",
		X1:    x1, Y1: y1, X2: x2, Y2: y2,
	}}, nil)
}
