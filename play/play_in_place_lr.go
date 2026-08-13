package play

import (
	"app/core"
	"fmt"
)

const inPlaceLRLogTag = "[原地左右打]"

func inPlaceLRLog(format string, args ...interface{}) {
	fmt.Printf(inPlaceLRLogTag+" "+format+"\n", args...)
}

// Play_原地左右打 原地循环：朝左打一下 → 600～800ms → 朝右打一下 → 600～800ms；后台维持喂宠与 HP/MP。
func Play_原地左右打(mapAssetPath string) error {
	if _, err := loadMapConfig(mapAssetPath); err != nil {
		return err
	}
	SetFarmLogTag(inPlaceLRLogTag)

	StartFarmMaintainLoop(inPlaceLRLogTag)
	defer StopFarmMaintainLoop()

	inPlaceLRLog("开始挂机")
	for {
		core.BlockWhileCaptchaHold()
		TickFarmMainThreadTasks()
		attackLeftOnce()
		core.RandomSleep(600, 800)
		attackRightOnce()
		core.RandomSleep(600, 800)
	}
}
