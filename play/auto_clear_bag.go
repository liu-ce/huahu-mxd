package play

import (
	"app/core"
	"app/job"
	"fmt"
	"math/rand"
	"time"
)

const (
	autoClearBagMaxRetry       = 10
	autoClearBagIntervalMinKey = "自动清包间隔最短分钟"
	autoClearBagIntervalMaxKey = "自动清包间隔最长分钟"
	autoClearBagIntervalMinDef = 30
	autoClearBagIntervalMaxDef = 40
	tiAutoClearBagXMin         = 43
	tiAutoClearBagXMax         = 54
	tiAutoClearBagShopYBelow   = 130
	autoClearBagStableSec      = 5
)

type autoClearBagState struct {
	nextAt          time.Time
	startupPending  bool // 脚本启动后优先到清包点清一次
	pendingShop     bool // 已到卖物区(y<130)，等待站稳后卖
	intervalMinSec  int
	intervalSpanSec int
}

func autoClearBagIntervalMinutes() (minMin, maxMin int) {
	return autoClearBagIntervalMinutesWithDefault(autoClearBagIntervalMinDef, autoClearBagIntervalMaxDef)
}

func autoClearBagIntervalMinutesWithDefault(minDef, maxDef int) (minMin, maxMin int) {
	minMin = core.API.GetConfigIntValueOrDefault(autoClearBagIntervalMinKey, minDef)
	maxMin = core.API.GetConfigIntValueOrDefault(autoClearBagIntervalMaxKey, maxDef)
	if minMin <= 0 {
		minMin = minDef
	}
	if maxMin <= 0 {
		maxMin = maxDef
	}
	if maxMin < minMin {
		maxMin = minMin
	}
	return minMin, maxMin
}

func newAutoClearBagState() autoClearBagState {
	return newAutoClearBagStateWithIntervalDefault(autoClearBagIntervalMinDef, autoClearBagIntervalMaxDef)
}

func newAutoClearBagStateWithIntervalDefault(minDef, maxDef int) autoClearBagState {
	minMin, maxMin := autoClearBagIntervalMinutesWithDefault(minDef, maxDef)
	s := autoClearBagState{
		intervalMinSec:  minMin * 60,
		intervalSpanSec: (maxMin - minMin) * 60,
	}
	if core.API.GetConfigBoolValue("自动清包") {
		s.startupPending = true
	}
	return s
}

func (s *autoClearBagState) due() bool {
	if !core.API.GetConfigBoolValue("自动清包") {
		return false
	}
	if s.startupPending {
		return true
	}
	if s.pendingShop {
		return true
	}
	return !s.nextAt.IsZero() && !time.Now().Before(s.nextAt)
}

func (s *autoClearBagState) finishAttempt(log func(string, ...interface{})) {
	s.startupPending = false
	s.pendingShop = false
	span := s.intervalSpanSec
	if span < 0 {
		span = 0
	}
	sec := s.intervalMinSec + rand.Intn(span+1)
	s.nextAt = time.Now().Add(time.Duration(sec) * time.Second)
	if log != nil {
		log("自动清包: 下次可执行约 %d 分钟后", sec/60)
	}
}

func readMinimapXY() (relX, relY int, ok bool) {
	mx, my, wx, wy := detectYellowThenWorld()
	if mx < 0 || my < 0 || wx < 0 || wy < 0 {
		return 0, 0, false
	}
	relX, relY = relativeToRef(mx, my, wx, wy)
	return relX, relY, true
}

func waitAutoClearBagPositionStable(log func(string, ...interface{})) bool {
	relX, relY, ok := readMinimapXY()
	if !ok {
		log("自动清包: 小地图未识别 取消卖物")
		return false
	}
	log("自动清包: 检测站位稳定 %ds (x=%d y=%d)", autoClearBagStableSec, relX, relY)
	deadline := time.Now().Add(autoClearBagStableSec * time.Second)
	lastX, lastY := relX, relY
	for time.Now().Before(deadline) {
		core.Sleep(500)
		curX, curY, ok := readMinimapXY()
		if !ok {
			log("自动清包: 等待中丢失小地图 取消卖物")
			return false
		}
		if curX != lastX || curY != lastY {
			log("自动清包: 站位变化 (%d,%d)→(%d,%d) 取消卖物", lastX, lastY, curX, curY)
			return false
		}
	}
	return true
}

// tryAutoShop 站位 5 秒内 x/y 不变才卖；否则放弃本次并排下次间隔。
func (s *autoClearBagState) tryAutoShop(log func(string, ...interface{})) bool {
	return s.tryAutoShopSellMisc(log, true)
}

func (s *autoClearBagState) tryAutoShopSellMisc(log func(string, ...interface{}), sellMisc bool) bool {
	if !waitAutoClearBagPositionStable(log) {
		if s.startupPending {
			log("自动清包: 站位未稳定 启动清包稍后重试")
			return false
		}
		log("自动清包: 站位未稳定 放弃本次 等下次间隔")
		s.finishAttempt(log)
		return false
	}
	relX, relY, ok := readMinimapXY()
	if !ok {
		log("自动清包: 卖物前小地图未识别 放弃本次")
		s.finishAttempt(log)
		return false
	}
	s.finishAttempt(log)
	job.DO_自动商店WithSellMisc(relX, relY, sellMisc)
	return true
}

func tryAutoClearBagUpTeleport(s *autoClearBagState, log func(string, ...interface{}), relX int, successYBelow int, stillOnSpot func(curY int) bool) bool {
	return tryAutoClearBagUpTeleportWait(s, log, relX, successYBelow, 500, stillOnSpot)
}

func tryAutoClearBagUpTeleportWait(s *autoClearBagState, log func(string, ...interface{}), relX int, successYBelow, waitMs int, stillOnSpot func(curY int) bool) bool {
	if !s.due() {
		return false
	}
	if waitMs <= 0 {
		waitMs = 500
	}

	log("自动清包: x=%d 开始上瞬移", relX)
	for attempt := 1; attempt <= autoClearBagMaxRetry; attempt++ {
		tapUpTeleport()
		core.Sleep(waitMs)

		mx, my, wx, wy := detectYellowThenWorld()
		if mx < 0 || my < 0 || wx < 0 || wy < 0 {
			log("自动清包: 第%d次 小地图未识别", attempt)
			continue
		}
		curX, curY := relativeToRef(mx, my, wx, wy)
		if curY < successYBelow {
			log("自动清包: 第%d次 到达卖物区 y=%d x=%d", attempt, curY, curX)
			s.pendingShop = true
			return s.tryAutoShop(log)
		}
		if stillOnSpot(curY) {
			log("自动清包: 第%d次 仍在一层 y=%d x=%d 重试", attempt, curY, curX)
			continue
		}
		log("自动清包: 第%d次 未到卖物区 y=%d 重试", attempt, curY)
	}

	log("自动清包: 上瞬移%d次未到卖物区 放弃本次", autoClearBagMaxRetry)
	s.finishAttempt(log)
	return true
}

func tryAutoClearBagTreasureIsland(s *autoClearBagState, ti *TreasureIslandConfig, lay, relX, relY int) bool {
	if s.pendingShop {
		if relY >= tiAutoClearBagShopYBelow {
			tiLog("自动清包: 已离开卖物区 y=%d 放弃本次", relY)
			s.finishAttempt(tiLog)
			return false
		}
		return s.tryAutoShop(tiLog)
	}
	if lay != 1 || !matchRange(relX, tiAutoClearBagXMin, tiAutoClearBagXMax) {
		return false
	}
	return tryAutoClearBagUpTeleport(s, tiLog, relX, tiAutoClearBagShopYBelow, func(curY int) bool {
		return tiDetectLayer(ti, curY) == 1
	})
}

func tryAlignTreasureIslandAutoClearBag(s *autoClearBagState, lay, farmLayer, relX int) bool {
	if !s.startupPending || lay != 1 || farmLayer != 1 {
		return false
	}
	if matchRange(relX, tiAutoClearBagXMin, tiAutoClearBagXMax) {
		return false
	}
	tiLog("自动清包: 启动对齐 x=%d → [%d,%d]", relX, tiAutoClearBagXMin, tiAutoClearBagXMax)
	tiAlignXByTeleport("", relX, tiAutoClearBagXMin, tiAutoClearBagXMax)
	return true
}

func tryAutoClearBagInstituteC1(s *autoClearBagState, tag string, st *StairsFarmConfig, relX, relY int) bool {
	if !stairsXInTarget(st, relX) || relY < st.YMin || relY > st.YMax {
		return false
	}
	if !s.due() {
		return false
	}
	log := func(format string, args ...interface{}) {
		instituteC1Log(tag, format, args...)
	}
	log("自动清包: x=%d y=%d 准备清包", relX, relY)
	return s.tryAutoShop(log)
}

// RunStartupAutoClearBag 非宝物岛/研究所 C1 地图：脚本启动后在当前位置尝试清包一次。
func RunStartupAutoClearBag(logTag string) {
	if !core.API.GetConfigBoolValue("自动清包") {
		return
	}
	s := newAutoClearBagState()
	if !s.startupPending {
		return
	}
	log := func(format string, args ...interface{}) {
		fmt.Printf(logTag+" "+format+"\n", args...)
	}
	log("自动清包: 脚本启动 尝试清包一次")
	s.tryAutoShop(log)
}
