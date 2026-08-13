package play

import (
	"math/rand"
	"time"
)

const (
	PatrolFarmTeleportOnlyDur   = 8 * time.Second
	PatrolFarmWalkChancePercent = 35
	patrolFarmWalkHoldMinMs     = 600
	patrolFarmWalkHoldMaxMs     = 1200
)

// patrolFarmAllowWalk 进入该层/平台打怪前 8 秒只瞬移；之后 35% 概率改走路+攻击。
func patrolFarmAllowWalk(farmSince time.Time) bool {
	if farmSince.IsZero() {
		return false
	}
	if time.Since(farmSince) < PatrolFarmTeleportOnlyDur {
		return false
	}
	return rand.Intn(100) < PatrolFarmWalkChancePercent
}

func patrolFarmWalkMs() int {
	if patrolFarmWalkHoldMaxMs <= patrolFarmWalkHoldMinMs {
		return patrolFarmWalkHoldMinMs
	}
	return patrolFarmWalkHoldMinMs + rand.Intn(patrolFarmWalkHoldMaxMs-patrolFarmWalkHoldMinMs+1)
}

func patrolFarmWalkAndAttack(goRight bool, attackMinMs, attackMaxMs int) {
	if goRight {
		faceRight()
	} else {
		faceLeft()
	}
	walkHoldMs(goRight, patrolFarmWalkMs())
	if attackMinMs <= 0 {
		attackMinMs = keyHoldShortMin
	}
	if attackMaxMs < attackMinMs {
		attackMaxMs = attackMinMs
	}
	keyHoldPress(attackKeyCode(), attackMinMs, attackMaxMs)
}
