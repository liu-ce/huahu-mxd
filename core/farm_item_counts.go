package core

import "sync"

// FarmItemCountsSnapshot 道具数量快照（经中位数滤波后的 OCR 结果）。
type FarmItemCountsSnapshot struct {
	HP, MP, Pet, Scroll         int
	HPOK, MPOK, PetOK, ScrollOK bool
}

var (
	farmItemCountsMu        sync.RWMutex
	farmItemCountsSnap      FarmItemCountsSnapshot
	farmItemCountsRefresher func()
)

// RegisterFarmItemCountsRefresher 注册道具数量 OCR 刷新函数（由 play 包在 init 时注册）。
func RegisterFarmItemCountsRefresher(fn func()) {
	farmItemCountsRefresher = fn
}

// RefreshFarmItemCounts 立即刷新道具数量快照。
func RefreshFarmItemCounts() {
	if farmItemCountsRefresher != nil {
		farmItemCountsRefresher()
	}
}

// SetFarmItemCountsSnapshot 写入道具数量快照。
func SetFarmItemCountsSnapshot(hp, mp, pet, scroll int, hpOK, mpOK, petOK, scrollOK bool) {
	farmItemCountsMu.Lock()
	farmItemCountsSnap = FarmItemCountsSnapshot{
		HP: hp, MP: mp, Pet: pet, Scroll: scroll,
		HPOK: hpOK, MPOK: mpOK, PetOK: petOK, ScrollOK: scrollOK,
	}
	farmItemCountsMu.Unlock()
}

// GetFarmItemCountsSnapshot 读取道具数量快照。
func GetFarmItemCountsSnapshot() FarmItemCountsSnapshot {
	farmItemCountsMu.RLock()
	defer farmItemCountsMu.RUnlock()
	return farmItemCountsSnap
}
