package play

import (
	"app/core"
	"fmt"
	"math/rand"
	"time"

	"github.com/Dasongzi1366/AutoGo/motion"
)

const downTeleportMinInterval = 2 * time.Second

var lastDownTeleportAt time.Time

const (
	teleportAttackIntervalMinKey = "瞬移和攻击间隔最短毫秒"
	teleportAttackIntervalMaxKey = "瞬移和攻击间隔最长毫秒"
)

func configuredTeleportAttackInterval() (minMs, maxMs int, ok bool) {
	minV, errMin := core.API.GetConfigInt(teleportAttackIntervalMinKey)
	maxV, errMax := core.API.GetConfigInt(teleportAttackIntervalMaxKey)
	if errMin != nil || errMax != nil || minV <= 0 || maxV <= 0 {
		return 0, 0, false
	}
	if maxV < minV {
		maxV = minV
	}
	return minV, maxV, true
}

func sleepTeleportAttackIntervalOr(defaultMin, defaultMax int) {
	if minMs, maxMs, ok := configuredTeleportAttackInterval(); ok {
		core.RandomSleep(minMs, maxMs)
		return
	}
	if defaultMax <= defaultMin {
		core.Sleep(defaultMin)
		return
	}
	core.RandomSleep(defaultMin, defaultMax)
}

// sleepAfterTeleport 瞬移后、攻击前的等待；配置「瞬移和攻击间隔*毫秒」优先。
func sleepAfterTeleport() {
	sleepTeleportAttackIntervalOr(50, 79)
}

// sleepAttackComboInterval 连击攻击次间等待；配置「瞬移和攻击间隔*毫秒」优先。
func sleepAttackComboInterval() {
	sleepTeleportAttackIntervalOr(100, 140)
}

const (
	keyHoldShortMin = 100
	keyHoldShortMax = 150
	keyHoldPetMin   = 100
	keyHoldPetMax   = 300
	keyHoldLongMin  = 500
	keyHoldLongMax  = 1000
	keyHoldSkillMin = 200
	keyHoldSkillMax = 300

	// 模拟器按键：Down+Sleep+Up 无效，需循环 KeyAction / KeyActionDown。
	emulatorComboActionEvery = 5
	emulatorComboSleepMs     = 2
)

func emulatorHoldMs(minMs, maxMs int) int {
	if maxMs <= minMs {
		return minMs
	}
	return minMs + rand.Intn(maxMs-minMs+1)
}

func emulatorRepeats(holdMs int) int {
	reps := holdMs / emulatorComboSleepMs
	if reps < 1 {
		reps = 1
	}
	return reps
}

// keyHoldPress 模拟器兼容：循环 KeyAction 点按功能键（攻击/宠物/瞬移键等）。
func keyHoldPress(code, minMs, maxMs int) {
	if core.IsCaptchaHold() {
		return
	}
	reps := emulatorRepeats(emulatorHoldMs(minMs, maxMs))
	for i := 0; i < reps; i++ {
		if core.IsCaptchaHold() {
			motion.KeyActionUp(code, 0)
			return
		}
		if i%emulatorComboActionEvery == 0 {
			motion.KeyAction(code, 0)
		}
		core.Sleep(emulatorComboSleepMs)
	}
}

// keyHoldDirection 模拟器兼容：循环重按方向键，最后 KeyActionUp（转向/走路）。
func keyHoldDirection(dirCode, minMs, maxMs int) {
	if core.IsCaptchaHold() {
		return
	}
	reps := emulatorRepeats(emulatorHoldMs(minMs, maxMs))
	for i := 0; i < reps; i++ {
		if core.IsCaptchaHold() {
			motion.KeyActionUp(dirCode, 0)
			return
		}
		motion.KeyActionDown(dirCode, 0)
		core.Sleep(emulatorComboSleepMs)
	}
	motion.KeyActionUp(dirCode, 0)
}

// refreshDpadHold 模拟器兼容：持续按住方向键时，每周期刷新 ms 毫秒（不松开）。
func refreshDpadHold(dirCode, ms int) {
	if core.IsCaptchaHold() {
		return
	}
	if ms < emulatorComboSleepMs {
		ms = emulatorComboSleepMs
	}
	reps := emulatorRepeats(ms)
	for i := 0; i < reps; i++ {
		if core.IsCaptchaHold() {
			motion.KeyActionUp(dirCode, 0)
			return
		}
		motion.KeyActionDown(dirCode, 0)
		core.Sleep(emulatorComboSleepMs)
	}
}

func releaseDpadHold(dirCode int) {
	motion.KeyActionUp(dirCode, 0)
}

func refreshDpadUpHold(ms int) {
	refreshDpadHold(motion.KEYCODE_DPAD_UP, ms)
}

func releaseDpadUp() {
	releaseDpadHold(motion.KEYCODE_DPAD_UP)
}

func tapPetFeed() {
	keyHoldPress(petKeyCode(), keyHoldPetMin, keyHoldPetMax)
}

func tapJump() {
	keyTapAction(jumpKeyCode())
}

func tapJumpLeft() {
	keyHoldDirectionAction(motion.KEYCODE_DPAD_LEFT, jumpKeyCode())
}

func tapJumpRight() {
	keyHoldDirectionAction(motion.KEYCODE_DPAD_RIGHT, jumpKeyCode())
}

func tapTeleportKey() {
	keyHoldPress(teleportKeyCode(), keyHoldShortMin, keyHoldShortMax)
}

func tapAttackOnce() {
	keyHoldPress(attackKeyCode(), keyHoldShortMin, keyHoldShortMax)
}

func attackLeftOnce() {
	faceLeft()
	core.Sleep(40)
	tapAttackOnce()
}

func attackRightOnce() {
	faceRight()
	core.Sleep(40)
	tapAttackOnce()
}

// tapAttack 随机攻击 3～5 次：每次按住攻击键，次间间隔 100～140ms。
func tapAttack() int {
	n := 3 + rand.Intn(3)
	code := attackKeyCode()
	for i := 0; i < n; i++ {
		keyHoldPress(code, keyHoldShortMin, keyHoldShortMax)
		if i < n-1 {
			sleepAttackComboInterval()
		}
	}
	return n
}

// PressAttackHold 按住攻击键一次（main 测试用）。
func PressAttackHold() {
	keyHoldPress(attackKeyCode(), keyHoldLongMin, keyHoldLongMax)
}

// keyTapAction 模拟器兼容：短按单键（纯跳跃等无方向组合时用）。
func keyTapAction(code int) {
	keyHoldPress(code, keyHoldShortMin, keyHoldShortMax)
}

// keyHoldDirectionAction 模拟器兼容：循环重按方向键，间隔 KeyAction 点按功能键（跳跃/瞬移）。
func keyHoldDirectionAction(dirCode, actionCode int) {
	keyHoldDirectionActionMs(dirCode, actionCode, keyHoldShortMin, keyHoldShortMax)
}

func keyHoldDirectionActionMs(dirCode, actionCode, minMs, maxMs int) {
	if core.IsCaptchaHold() {
		return
	}
	reps := emulatorRepeats(emulatorHoldMs(minMs, maxMs))
	for i := 0; i < reps; i++ {
		if core.IsCaptchaHold() {
			motion.KeyActionUp(dirCode, 0)
			motion.KeyActionUp(actionCode, 0)
			return
		}
		motion.KeyActionDown(dirCode, 0)
		if i%emulatorComboActionEvery == 0 {
			motion.KeyAction(actionCode, 0)
		}
		motion.KeyActionDown(dirCode, 0)
		core.Sleep(emulatorComboSleepMs)
	}
	motion.KeyActionUp(dirCode, 0)
}

func tapTeleportWithDirection(goRight bool) {
	if goRight {
		teleportRightAction()
	} else {
		teleportLeftAction()
	}
}

func teleportRightAction() {
	keyHoldDirectionAction(motion.KEYCODE_DPAD_RIGHT, teleportKeyCode())
}

func teleportLeftAction() {
	keyHoldDirectionAction(motion.KEYCODE_DPAD_LEFT, teleportKeyCode())
}

func teleportAndAttack(goRight bool) {
	if goRight {
		faceRight()
		teleportRightAction()
	} else {
		faceLeft()
		teleportLeftAction()
	}
	sleepAfterTeleport()
	tapAttackOnce()
}

// TestTeleportRight 供 main 调试右瞬移。
func TestTeleportRight() {
	teleportRightAction()
}

// TestTeleportLeft 供 main 调试左瞬移。
func TestTeleportLeft() {
	teleportLeftAction()
}

// TestJump 供 main 调试跳跃。
func TestJump() {
	tapJump()
}

// TestTeleportLeftAndAttack 供 main 调试左瞬移+攻击（与叶秋001巡逻一致）。
func TestTeleportLeftAndAttack() {
	teleportAndAttack(false)
}

// TestTeleportRightAndAttack 供 main 调试右瞬移+攻击。
func TestTeleportRightAndAttack() {
	teleportAndAttack(true)
}

func tapUpTeleport() {
	keyHoldDirectionAction(motion.KEYCODE_DPAD_UP, teleportKeyCode())
}

func tapDownTeleport() {
	if !lastDownTeleportAt.IsZero() && time.Since(lastDownTeleportAt) < downTeleportMinInterval {
		return
	}
	keyHoldDirectionAction(motion.KEYCODE_DPAD_DOWN, teleportKeyCode())
	lastDownTeleportAt = time.Now()
}

func tapAttackTwice() {
	code := attackKeyCode()
	for i := 0; i < 2; i++ {
		keyHoldPress(code, keyHoldShortMin, keyHoldShortMax)
		if i < 1 {
			sleepAttackComboInterval()
		}
	}
}

func tapDownJump() {
	keyHoldDirectionAction(motion.KEYCODE_DPAD_DOWN, jumpKeyCode())
}

// holdDpad 长按方向键走路。
func holdDpad(code int, ms int) {
	if ms < 40 {
		ms = 40
	}
	keyHoldDirection(code, ms, ms)
}

// holdDpadRandom 按住方向键 minMs～maxMs 后松开。
func holdDpadRandom(code, minMs, maxMs int) {
	keyHoldDirection(code, minMs, maxMs)
}

func walkHoldMs(goRight bool, ms int) {
	if goRight {
		holdDpad(motion.KEYCODE_DPAD_RIGHT, ms)
	} else {
		holdDpad(motion.KEYCODE_DPAD_LEFT, ms)
	}
}

func faceLeft() {
	keyHoldDirection(motion.KEYCODE_DPAD_LEFT, 80, 120)
}

func faceRight() {
	keyHoldDirection(motion.KEYCODE_DPAD_RIGHT, 80, 120)
}

func jitterWalkMs(base int) int {
	if base <= 0 {
		base = 400
	}
	lo := base * 70 / 100
	hi := base * 130 / 100
	if hi <= lo {
		hi = lo + 50
	}
	return lo + rand.Intn(hi-lo+1)
}

const (
	releaseKeyGapMinMs = 100
	releaseKeyGapMaxMs = 130
)

// ReleaseAllHeldKeys 只做 KeyActionUp（连抬两次），绝不 KeyAction 再点按。
// 旧逻辑 Up 后再 KeyAction 会在 GM 输入框里打出空格/x。
func ReleaseAllHeldKeys() {
	ensureAutoKeysLoaded()
	autoKeysMu.RLock()
	codes := collectHeldKeyCodesLocked()
	autoKeysMu.RUnlock()

	for _, code := range codes {
		motion.KeyActionUp(code, 0)
		core.RandomSleep(releaseKeyGapMinMs, releaseKeyGapMaxMs)
		motion.KeyActionUp(code, 0)
		core.RandomSleep(releaseKeyGapMinMs, releaseKeyGapMaxMs)
	}
	if len(codes) > 0 {
		fmt.Printf("[按键] 已松开 %d 个键\n", len(codes))
	}
}

func collectHeldKeyCodesLocked() []int {
	seen := make(map[int]struct{})
	var codes []int
	add := func(c int) {
		if c <= 0 {
			return
		}
		if _, ok := seen[c]; ok {
			return
		}
		seen[c] = struct{}{}
		codes = append(codes, c)
	}
	add(motion.KEYCODE_DPAD_LEFT)
	add(motion.KEYCODE_DPAD_RIGHT)
	add(motion.KEYCODE_DPAD_UP)
	add(motion.KEYCODE_DPAD_DOWN)
	add(keyAttackCode)
	add(keyJumpCode)
	add(keyTeleportCode)
	add(keyPetCode)
	for _, s := range timedAutoKeys {
		add(s.code)
	}
	// 默认攻击/瞬移键，配置改键后仍强制抬起，避免输入框连打空格/x
	add(motion.KEYCODE_SPACE)
	add(motion.KEYCODE_X)
	add(motion.KEYCODE_SHIFT_LEFT)
	add(motion.KEYCODE_SHIFT_RIGHT)
	return codes
}
