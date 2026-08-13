package util

import (
	"fmt"
	"runtime/debug"
	"sync"
	"time"

	"app/core"
)

// Timer 定时器结构
type Timer struct {
	ID       string
	Interval time.Duration
	Callback func()
	ticker   *time.Ticker
	done     chan bool
	paused   bool
	mu       sync.Mutex
}

// TimerManager 定时器管理器
type TimerManager struct {
	timers map[string]*Timer
	mu     sync.RWMutex
}

// 全局定时器管理器
var manager *TimerManager

func init() {
	manager = &TimerManager{
		timers: make(map[string]*Timer),
	}
}

// Every 周期性执行
func Every(interval time.Duration, callback func()) string {
	id := generateID()

	timer := &Timer{
		ID:       id,
		Interval: interval,
		Callback: callback,
		ticker:   time.NewTicker(interval),
		done:     make(chan bool),
		paused:   false,
	}

	manager.mu.Lock()
	manager.timers[id] = timer
	manager.mu.Unlock()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				core.SLS_Log2(fmt.Sprintf("定时器 %s 崩溃: %v\n堆栈:\n%s", timer.ID, r, debug.Stack()))
				core.Sleep(3000)
			}
		}()
		for {
			select {
			case <-timer.ticker.C:
				timer.mu.Lock()
				if !timer.paused {
					timer.Callback()
				}
				timer.mu.Unlock()
			case <-timer.done:
				timer.ticker.Stop()
				return
			}
		}
	}()

	return id
}

// 便捷函数 - 使用秒
func EverySec(sec int, callback func()) string {
	return Every(time.Duration(sec)*time.Second, callback)
}

// generateID 生成唯一ID
func generateID() string {
	return fmt.Sprintf("timer_%d", time.Now().UnixNano())
}
