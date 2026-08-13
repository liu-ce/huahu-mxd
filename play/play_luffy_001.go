package play

import (
	"app/core"
	"time"
)

const (
	luffy001LogTag       = "[路飞001]"
	luffyWalkAttackMinMs = 1000
	luffyWalkAttackMaxMs = 2000
)

// luffyPatrolFarmStep 打怪区往返：全程走路+攻击 1～2s（无瞬移）。
func luffyPatrolFarmStep(tag string, y *Yeqiu001Config, relX, xMin, xMax int, goRight *bool, _ time.Time) {
	margin := yeqiuPatrolEffectiveMargin(xMin, xMax, yeqiuPatrolMargin(y))

	if relX < xMin {
		*goRight = true
		yeqiuLog("%s: x=%d 超出左界%d 右走+攻击", tag, relX, xMin)
		patrolFarmWalkAndAttack(true, luffyWalkAttackMinMs, luffyWalkAttackMaxMs)
		return
	}
	if relX > xMax {
		*goRight = false
		yeqiuLog("%s: x=%d 超出右界%d 左走+攻击", tag, relX, xMax)
		patrolFarmWalkAndAttack(false, luffyWalkAttackMinMs, luffyWalkAttackMaxMs)
		return
	}

	if *goRight && relX >= xMax-margin {
		*goRight = false
		yeqiuLog("%s: 近右界 relX=%d 改向左", tag, relX)
		core.Sleep(50)
		return
	}
	if !*goRight && relX <= xMin+margin {
		*goRight = true
		yeqiuLog("%s: 近左界 relX=%d 改向右", tag, relX)
		core.Sleep(50)
		return
	}

	dir := "左"
	if *goRight {
		dir = "右"
	}
	yeqiuLog("%s: %s走+攻击 relX=%d", tag, dir, relX)
	patrolFarmWalkAndAttack(*goRight, luffyWalkAttackMinMs, luffyWalkAttackMaxMs)
}

// Play_路飞001 同叶秋001，全程无瞬移：巡逻与绳子对齐均走路+攻击。
func Play_路飞001(mapAssetPath string) error {
	return runYeqiu001(mapAssetPath, luffy001LogTag, luffyPatrolFarmStep, true)
}
