package play

import (
	"app/core"
	"app/job"
	"fmt"
	"sync/atomic"
	"time"
)

const (
	scheduledLogoutConfigKey   = "定时下线"
	scheduledLogoutDurationKey = "定时下线时长分钟"
	scheduledLogoutIntervalKey = "定时下线间隔小时"
)

func scheduledLogoutEnabled() bool {
	return core.API.GetConfigBoolValue(scheduledLogoutConfigKey)
}

func scheduledLogoutDuration() time.Duration {
	min := core.API.GetConfigIntValueOrDefault(scheduledLogoutDurationKey, 20)
	if min <= 0 {
		min = 20
	}
	return time.Duration(min) * time.Minute
}

func scheduledLogoutInterval() time.Duration {
	h := core.API.GetConfigIntValueOrDefault(scheduledLogoutIntervalKey, 3)
	if h <= 0 {
		h = 3
	}
	return time.Duration(h) * time.Hour
}

func startScheduledLogoutLoop() {
	if !scheduledLogoutEnabled() {
		return
	}
	go runScheduledLogoutLoop()
}

func runScheduledLogoutLoop() {
	interval := scheduledLogoutInterval()
	offlineDur := scheduledLogoutDuration()
	nextAt := time.Now().Add(interval)
	farmLog("定时下线: 已启用 间隔=%s 离线=%s", interval, offlineDur)

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for atomic.LoadInt32(&farmMaintainOn) == 1 {
		<-ticker.C
		if !scheduledLogoutEnabled() {
			continue
		}
		if core.IsCaptchaHold() || core.IsScheduledLogoutActive() {
			continue
		}
		if time.Now().Before(nextAt) {
			continue
		}
		runScheduledLogoutOnce(offlineDur)
		interval = scheduledLogoutInterval()
		offlineDur = scheduledLogoutDuration()
		nextAt = time.Now().Add(interval)
		farmLog("定时下线: 下次触发 %s 后", interval)
	}
}

func runScheduledLogoutOnce(offlineDur time.Duration) {
	farmLog("定时下线: 开始")
	core.SetScheduledLogoutActive(true)
	defer core.SetScheduledLogoutActive(false)

	if job.DoAutoLogout() {
		farmLog("定时下线: 离线成功，等待 %s", offlineDur)
		deadline := time.Now().Add(offlineDur)
		lastLogMin := -1
		for time.Now().Before(deadline) {
			if atomic.LoadInt32(&farmMaintainOn) == 0 {
				return
			}
			remainMin := int(time.Until(deadline).Minutes() + 0.999)
			if remainMin != lastLogMin && remainMin%5 == 0 {
				lastLogMin = remainMin
				core.SLS_Log2NoToast(fmt.Sprintf("自动离线中，剩余约 %d 分钟", remainMin))
			}
			core.Sleep(5000)
		}
	} else {
		farmLog("定时下线: 离线失败")
		core.SLS_Log2("定时下线: 自动离线失败")
	}

	if job.DoAutoLogin() {
		farmLog("定时下线: 登录成功")
		ReloadAutoKeysFromConfig()
	} else {
		farmLog("定时下线: 登录失败")
		core.SLS_Log2("定时下线: 自动登录失败")
	}
}

// TestScheduledLogoutOnce 手动测一轮定时下线：退出大厅 → 等 offlineMinutes 分钟 → 自动登录。
// offlineMinutes<=0 时默认 1 分钟。需已在游戏内（角色界面）。
func TestScheduledLogoutOnce(offlineMinutes int) {
	if offlineMinutes <= 0 {
		offlineMinutes = 1
	}
	SetFarmLogTag("[定时下线测试]")
	atomic.StoreInt32(&farmMaintainOn, 1)
	defer atomic.StoreInt32(&farmMaintainOn, 0)

	offlineDur := time.Duration(offlineMinutes) * time.Minute
	farmLog("手动测试: 离线等待=%s", offlineDur)
	runScheduledLogoutOnce(offlineDur)
	farmLog("手动测试: 结束")
}
