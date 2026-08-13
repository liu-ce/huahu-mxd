package core

import (
	"fmt"
	"math/rand"
	"time"
)

type farmAutoSkillSlot struct {
	id     string
	x1, y1 int
	x2, y2 int
	minSec int
	maxSec int
	nextAt time.Time
}

var farmAutoSkillSlots []farmAutoSkillSlot

func randSecInRange(minS, maxS int) int {
	if maxS < minS {
		maxS = minS
	}
	if maxS <= minS {
		return minS
	}
	return minS + rand.Intn(maxS-minS+1)
}

// InitFarmAutoSkillsFromConfig 从角色配置 content 读取「自动释放技能」；仅 enabled 的槽位参与。
// 路径示例：自动释放技能.A.enabled、.minSec、.maxSec（秒，下次点击的随机间隔范围）。
func InitFarmAutoSkillsFromConfig() {
	farmAutoSkillSlots = nil
	if API == nil {
		return
	}
	order := []string{"A", "B", "C", "D"}
	var slots []farmAutoSkillSlot
	for _, id := range order {
		base := "自动释放技能." + id
		if !API.GetConfigBoolValue(base + ".enabled") {
			continue
		}
		rect, ok := FarmAutoSkillRects[id]
		if !ok {
			continue
		}
		minS := API.GetConfigIntValueOrDefault(base+".minSec", 60)
		maxS := API.GetConfigIntValueOrDefault(base+".maxSec", 90)
		if minS <= 0 {
			minS = 60
		}
		if maxS <= 0 {
			maxS = 90
		}
		if maxS < minS {
			maxS = minS
		}
		// nextAt 为零：进入挂机主循环后首轮 Tick 即释放；之后按 minSec~maxSec 随机间隔。
		slots = append(slots, farmAutoSkillSlot{
			id: id, x1: rect[0], y1: rect[1], x2: rect[2], y2: rect[3],
			minSec: minS, maxSec: maxS,
			nextAt: time.Time{},
		})
	}
	farmAutoSkillSlots = slots
	if len(slots) > 0 {
		fmt.Printf("[farm-auto-skill] 已启用 %d 个槽位；首轮进入挂机循环即尝试点击，之后按 min~max 秒间隔\n", len(slots))
	}
}

// ResetFarmAutoSkills 挂机流程结束时清空状态。
func ResetFarmAutoSkills() {
	farmAutoSkillSlots = nil
}

const farmAutoSkillPostClickMs = 3000

// 挂机诊断：区域 dd0033 像素计数；仅当 10<n<40 时点血瓶补到 ≥600（硬编码矩形，不配 JSON）。
const (
	farmColorProbeX1, farmColorProbeY1 = 555, 615
	farmColorProbeX2, farmColorProbeY2 = 800, 629
	farmColorProbeHex                  = "dd0033"
	farmColorProbeSim                  = float32(1.0)
	// 喝血：仅当 10 < 色块 < 40（过低可能是界面变化/检测异常，避免乱点）
	farmColorProbePotionBelow    = 450
	farmColorProbePotionAboveMin = 10
	farmColorProbeHealTarget     = 550
	farmPotionX1, farmPotionY1   = 1122, 290
	farmPotionX2, farmPotionY2   = 1134, 303
	farmPotionMaxClicks          = 12
)

var farmColorProbeNextLog time.Time

// 颜色探测间隔不宜过短：每圈挂机还会跑 YOLO/截图，缩短间隔会叠加 native 负载。
const farmColorProbeIntervalMs = 900

func tickFarmColorProbeMaybeLog() {
	now := time.Now()
	if now.Before(farmColorProbeNextLog) {
		return
	}
	farmColorProbeNextLog = now.Add(time.Duration(JitterMs(farmColorProbeIntervalMs, 0.25)) * time.Millisecond)
	n := Color.GetColorCountInRegion(farmColorProbeX1, farmColorProbeY1, farmColorProbeX2, farmColorProbeY2, farmColorProbeHex, farmColorProbeSim)
	fmt.Printf("[farm] GetColorCountInRegion(%d,%d,%d,%d %s sim=%v) = %d\n",
		farmColorProbeX1, farmColorProbeY1, farmColorProbeX2, farmColorProbeY2, farmColorProbeHex, farmColorProbeSim, n)

	// 自动吃血药已关闭，仅保留血条色块日志。
}

// TickFarmAutoSkillsDuringFarm 在主挂机循环内调用：到期则点击技能区，再暂停 3 秒，并安排下次随机间隔。
func TickFarmAutoSkillsDuringFarm() {
	if IsScheduledLogoutActive() {
		return
	}
	tickFarmColorProbeMaybeLog()
	if len(farmAutoSkillSlots) == 0 {
		return
	}
	now := time.Now()
	for i := range farmAutoSkillSlots {
		s := &farmAutoSkillSlots[i]
		// nextAt 为零表示「尚未排期」= 首轮立即释放
		if !s.nextAt.IsZero() && now.Before(s.nextAt) {
			continue
		}
		fmt.Printf("[farm-auto-skill] 点击槽 %s 区域 [%d,%d,%d,%d]\n", s.id, s.x1, s.y1, s.x2, s.y2)
		RandomClickInArea(s.x1, s.y1, s.x2, s.y2)
		RandomSleepAround(farmAutoSkillPostClickMs, 0.3)
		nextDelay := randSecInRange(s.minSec, s.maxSec)
		s.nextAt = time.Now().Add(time.Duration(nextDelay) * time.Second)
		fmt.Printf("[farm-auto-skill] 槽 %s 下次触发约 %d 秒后\n", s.id, nextDelay)
	}
}
