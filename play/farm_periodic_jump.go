package play

import (
	"app/core"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

const (
	farmPeriodicLRJumpMinSec = 40
	farmPeriodicLRJumpMaxSec = 60
)

var (
	farmPeriodicLRJumpOn   int32
	farmPeriodicLRJumpMu   sync.Mutex
	farmPeriodicLRJumpNext time.Time
)

func farmPeriodicLRJumpInterval() time.Duration {
	span := farmPeriodicLRJumpMaxSec - farmPeriodicLRJumpMinSec
	sec := farmPeriodicLRJumpMinSec
	if span > 0 {
		sec += rand.Intn(span + 1)
	}
	return time.Duration(sec) * time.Second
}

// EnableFarmPeriodicLRJump 启用挂机主循环内 40~60 秒一次的左跳或右跳（宝物岛、野熊的领地等）。
func EnableFarmPeriodicLRJump() {
	farmPeriodicLRJumpMu.Lock()
	farmPeriodicLRJumpNext = time.Now().Add(farmPeriodicLRJumpInterval())
	farmPeriodicLRJumpMu.Unlock()
	atomic.StoreInt32(&farmPeriodicLRJumpOn, 1)
}

// DisableFarmPeriodicLRJump 关闭定时左右跳。
func DisableFarmPeriodicLRJump() {
	atomic.StoreInt32(&farmPeriodicLRJumpOn, 0)
	farmPeriodicLRJumpMu.Lock()
	farmPeriodicLRJumpNext = time.Time{}
	farmPeriodicLRJumpMu.Unlock()
}

// TickFarmPeriodicLRJump 挂机主循环每轮调用。
func TickFarmPeriodicLRJump() {
	if atomic.LoadInt32(&farmPeriodicLRJumpOn) != 1 {
		return
	}
	if core.IsCaptchaHold() || farmMaintainPaused() {
		return
	}
	now := time.Now()
	farmPeriodicLRJumpMu.Lock()
	if now.Before(farmPeriodicLRJumpNext) {
		farmPeriodicLRJumpMu.Unlock()
		return
	}
	farmPeriodicLRJumpNext = now.Add(farmPeriodicLRJumpInterval())
	farmPeriodicLRJumpMu.Unlock()

	if rand.Intn(2) == 0 {
		farmLog("维护: 定时左跳")
		tapJumpLeft()
	} else {
		farmLog("维护: 定时右跳")
		tapJumpRight()
	}
}
