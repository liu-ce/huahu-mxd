package huashu

import (
	"math/rand"

	"github.com/Dasongzi1366/AutoGo/utils"
)

const (
	trackTouchEnabled = false
	mousePressX       = 638
	mousePressY       = 252
)

// const mouseTouchFinger = 2

func randomSleep(baseMs int) {
	if !trackTouchEnabled && baseMs > 0 {
		return
	}
	utils.Sleep(baseMs + utils.Random(20, 50))
}

func randomPointInArea(x1, y1, x2, y2 int) (int, int) {
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	if y1 > y2 {
		y1, y2 = y2, y1
	}
	return utils.Random(x1, x2), utils.Random(y1, y2)
}

func mousePressDown() {
	if !trackTouchEnabled {
		return
	}
	// motion.TouchDown(mousePressX, mousePressY, mouseTouchFinger, 0)
}

func touchUpFinger() {
	if !trackTouchEnabled {
		return
	}
	// motion.TouchUp(0, 0, mouseTouchFinger, 0)
}

func randomClickInArea(x1, y1, x2, y2, holdMs int) {
	if !trackTouchEnabled {
		return
	}
	// x, y := randomPointInArea(x1, y1, x2, y2)
	// motion.TouchDown(x, y, mouseTouchFinger, 0)
	randomSleep(holdMs)
	// motion.TouchUp(x, y, mouseTouchFinger, 0)
}

func randomTouchDownInArea(x1, y1, x2, y2, holdMs int) {
	if !trackTouchEnabled {
		return
	}
	// x, y := randomPointInArea(x1, y1, x2, y2)
	// motion.TouchDown(x, y, mouseTouchFinger, 0)
	randomSleep(holdMs)
}

func touchMoveEx(x, y, durationMs int) {
	if !trackTouchEnabled {
		return
	}
	// motion.TouchMove(x, y, mouseTouchFinger, 0)
	if durationMs > 0 {
		utils.Sleep(durationMs)
	}
}

func randomTouchMoveJitter(baseX, baseY, durationMs int) {
	if !trackTouchEnabled {
		return
	}
	r := rand.Intn(5) + 1
	touchMoveEx(baseX+r, baseY+r, durationMs)
}
