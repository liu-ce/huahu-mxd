package job

import (
	"app/core"
	"app/util"
	"fmt"
	"math/rand"
	"strconv"
	"time"
)

const (
	roleUpdatePeriodMs       = 5 * 60 * 1000
	silverReadMinGapSec      = 2 * 3600
	silverReadMaxGapSec      = 3 * 3600
	keySilverReadNextUnix    = "silver_read_next_unix"
	keyCharacterReadNextUnix = "character_panel_read_next_unix"
	characterReadIntervalSec = 20 * 60
)

// StartRoleAndSilverMaintenanceLoop 与挂机同生命周期：与挂机入口同时启动（仅挂机时跑）。
// 每 5 分钟：RoleUpdate、读金币（若到点）、读人物面板（若到点）。
// 读金币 2～3 小时随机 → data/silver_read_next_unix；人物面板每 20 分钟 → data/character_panel_read_next_unix；
// OSS 截图仅在启动时上传一次；人物面板由挂机入口主线程先读一次（待 play 重写后恢复）。
// 每次完整 RoleUpdate 前 core 会合并 default.json 的 version 与 data/silver（见 syncRolePayloadFromLocal）。
func StartRoleAndSilverMaintenanceLoop(logTag string) {
	go func() {
		// 人物面板已在 RunAutoFarm 主线程读过；此处只排 20 分钟后下一轮
		now := time.Now().Unix()
		core.Storages.DataPut(keyCharacterReadNextUnix, strconv.FormatInt(now+characterReadIntervalSec, 10))
		fmt.Printf("%s[人物面板] 已排下次读取 unix=%d（约 20 分钟后）\n", logTag, now+characterReadIntervalSec)

		fmt.Printf("%s[oss] 启动时 UploadOSS\n", logTag)
		if err := util.UploadOSS(); err != nil {
			fmt.Printf("%s[oss] UploadOSS: %v\n", logTag, err)
		} else {
			fmt.Printf("%s[oss] UploadOSS 完成\n", logTag)
		}

		for {
			if core.API.GetConfigBoolValue("统计金币") {
				tryReadSilverIfDue(logTag)
			}

			tryReadCharacterPanelIfDue(logTag)
			core.Sleep(roleUpdatePeriodMs)
		}
	}()
}

func tryReadSilverIfDue(logTag string) {
	now := time.Now().Unix()
	nextStr := core.Storages.DataGet(keySilverReadNextUnix)
	if nextStr != "" {
		if nextTs, err := strconv.ParseInt(nextStr, 10, 64); err == nil && now < nextTs {
			return
		}
	}
	fmt.Printf("%s[silver] 执行读取金币 → 写入 data/silver\n", logTag)
	DO_读取金币()
	gap := int64(silverReadMinGapSec + rand.Intn(silverReadMaxGapSec-silverReadMinGapSec+1))
	core.Storages.DataPut(keySilverReadNextUnix, strconv.FormatInt(now+gap, 10))
	fmt.Printf("%s[silver] 下次读取约 %.1f 小时后 (unix=%d)\n", logTag, float64(gap)/3600.0, now+gap)
}

// tryReadCharacterPanelIfDue 每 20 分钟读一次人物等级/名字（首次 next 为空则本轮立刻读，与挂机线程同启）。
func tryReadCharacterPanelIfDue(logTag string) {
	now := time.Now().Unix()
	nextStr := core.Storages.DataGet(keyCharacterReadNextUnix)
	if nextStr != "" {
		if nextTs, err := strconv.ParseInt(nextStr, 10, 64); err == nil && now < nextTs {
			return
		}
	}
	fmt.Printf("%s[人物面板] 执行 OCR 读取等级\n", logTag)
	DO_读取人物等级()
	core.Storages.DataPut(keyCharacterReadNextUnix, strconv.FormatInt(now+characterReadIntervalSec, 10))
}
