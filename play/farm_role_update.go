package play

import (
	"app/core"
	"app/job"
	"fmt"
	"sync"
	"time"
)

const farmRoleUpdateInterval = 20 * time.Minute

const (
	roleDataKeyHPBottle    = "血瓶数量"
	roleDataKeyMPBottle    = "蓝瓶数量"
	roleDataKeyPetSnack    = "宠物零食数量"
	roleDataKeyChaosScroll = "卷轴数量"
)

func init() {
	core.RegisterFarmItemCountsRefresher(runFarmItemCountsOCR)
}

var (
	farmRoleUpdateMu   sync.Mutex
	farmRoleUpdateNext time.Time
	farmRoleUpdateInit bool
)

func storeFarmItemCountsSnap(hp, mp, pet, scroll int, hpOK, mpOK, petOK, scrollOK bool) {
	core.SetFarmItemCountsSnapshot(hp, mp, pet, scroll, hpOK, mpOK, petOK, scrollOK)
}

func applyFarmItemCountsToRoleDatas() {
	if core.Role == nil {
		return
	}
	snap := core.GetFarmItemCountsSnapshot()

	if core.Role.Datas == nil {
		core.Role.Datas = make(map[string]interface{})
	}
	setRoleDataCount(roleDataKeyHPBottle, snap.HP, snap.HPOK)
	setRoleDataCount(roleDataKeyMPBottle, snap.MP, snap.MPOK)
	setRoleDataCount(roleDataKeyPetSnack, snap.Pet, snap.PetOK)
	setRoleDataCount(roleDataKeyChaosScroll, snap.Scroll, snap.ScrollOK)
}

func setRoleDataCount(key string, n int, ok bool) {
	if !ok {
		delete(core.Role.Datas, key)
		return
	}
	core.Role.Datas[key] = n
}

// TickFarmRoleUpdateOnMainThread 挂机主循环调用：定时读等级并 RoleUpdate（含道具数量 datas）。
func TickFarmRoleUpdateOnMainThread() {
	if !core.API.GetConfigBoolValueOrDefault("中控通讯", true) {
		return
	}
	now := time.Now()
	farmRoleUpdateMu.Lock()
	if !farmRoleUpdateInit {
		farmRoleUpdateInit = true
		farmRoleUpdateNext = now
	}
	if now.Before(farmRoleUpdateNext) {
		farmRoleUpdateMu.Unlock()
		return
	}
	farmRoleUpdateNext = now.Add(farmRoleUpdateInterval)
	farmRoleUpdateMu.Unlock()

	job.DO_读取人物等级()
	applyFarmItemCountsToRoleDatas()

	if core.API == nil || core.Role == nil {
		return
	}
	if _, err := core.API.RoleUpdate(); err != nil {
		farmLog("维护: RoleUpdate %v", err)
	} else {
		farmLog("维护: RoleUpdate 已推送")
	}
}

func resetFarmRoleUpdateSchedule() {
	farmRoleUpdateMu.Lock()
	farmRoleUpdateInit = false
	farmRoleUpdateNext = time.Time{}
	farmRoleUpdateMu.Unlock()
	resetFarmItemOCRSchedule()
	resetFarmItemCountFilters()
}

var (
	farmItemOCRMu   sync.Mutex
	farmItemOCRNext time.Time
	farmItemOCRInit bool
)

func resetFarmItemOCRSchedule() {
	farmItemOCRMu.Lock()
	farmItemOCRInit = false
	farmItemOCRNext = time.Time{}
	farmItemOCRMu.Unlock()
}

// TickFarmItemCountsOnMainThread 挂机主循环调用：OCR 道具数量并写入快照（供 RoleUpdate datas）。
func TickFarmItemCountsOnMainThread() {
	now := time.Now()
	farmItemOCRMu.Lock()
	if !farmItemOCRInit {
		farmItemOCRInit = true
		farmItemOCRNext = now
	}
	if now.Before(farmItemOCRNext) {
		farmItemOCRMu.Unlock()
		return
	}
	farmItemOCRNext = now.Add(farmItemOCRInterval)
	farmItemOCRMu.Unlock()

	runFarmItemCountsOCR()
}

// TickFarmMainThreadTasks 挂机主循环每轮调用：道具 OCR + RoleUpdate + 定时左/右跳。
func TickFarmMainThreadTasks() {
	TickFarmItemCountsOnMainThread()
	TickFarmRoleUpdateOnMainThread()
	TickFarmPeriodicLRJump()
}

func formatFarmItemCountsForLog(hp, mp, pet, scroll int, hpOK, mpOK, petOK, scrollOK bool) string {
	return fmt.Sprintf("血瓶=%s 蓝瓶=%s 宠物零食=%s 混沌卷轴=%s",
		formatOCRCount(hp, hpOK),
		formatOCRCount(mp, mpOK),
		formatOCRCount(pet, petOK),
		formatOCRCount(scroll, scrollOK),
	)
}
