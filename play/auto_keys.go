package play

import (
	"app/core"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/Dasongzi1366/AutoGo/motion"
)

const (
	autoKeyModeInstant = "即时"
	autoKeyModeTimed   = "定时"

	autoKeyRolePet      = "宠物药水"
	autoKeyRoleJump     = "跳跃键"
	autoKeyRoleTeleport = "瞬移键"
	autoKeyRoleAttack   = "攻击键"

	timedAutoKeyBatchGapMinMs = 400
	timedAutoKeyBatchGapMaxMs = 500

	screenHotkeyHoldMinMs = 100
	screenHotkeyHoldMaxMs = 180
)

type timedAutoKeySlot struct {
	name           string
	code           int
	minSec, maxSec int
	nextAt         time.Time
	longHold       bool
	touch          bool // true=长按屏幕区域，非键盘
	x1, y1, x2, y2 int
}

// 快捷键1~4：UI 槽位，长按屏幕而非键盘。
var screenHotkeyRegions = map[string][4]int{
	"快捷键1": {987, 321, 1009, 336},
	"快捷键2": {1057, 318, 1067, 335},
	"快捷键3": {1124, 321, 1137, 334},
	"快捷键4": {1191, 315, 1207, 331},
}

var (
	autoKeysMu      sync.RWMutex
	autoKeysLoaded  bool
	keyAttackCode   = motion.KEYCODE_SPACE
	keyJumpCode     = motion.KEYCODE_Z
	keyTeleportCode = motion.KEYCODE_X
	keyPetCode      = 42
	timedAutoKeys   []timedAutoKeySlot
)

func ensureAutoKeysLoaded() {
	autoKeysMu.RLock()
	loaded := autoKeysLoaded
	autoKeysMu.RUnlock()
	if loaded {
		return
	}
	loadAutoKeysFromConfig()
}

func loadAutoKeysFromConfig() {
	autoKeysMu.Lock()
	defer autoKeysMu.Unlock()

	keyAttackCode = motion.KEYCODE_SPACE
	keyJumpCode = motion.KEYCODE_Z
	keyTeleportCode = motion.KEYCODE_X
	keyPetCode = 42
	timedAutoKeys = nil

	if core.API == nil {
		applyDefaultTimedPetLocked()
		autoKeysLoaded = true
		return
	}

	arr, err := core.API.GetConfigArray("自动按键")
	if err != nil || len(arr) == 0 {
		applyDefaultTimedPetLocked()
		autoKeysLoaded = true
		return
	}

	for _, item := range arr {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		name := configString(m, "名称")
		if name == "" || !configBool(m, "启用", true) {
			continue
		}
		keyStr := configString(m, "按键")
		region, isTouch := parseScreenHotkey(keyStr)
		code := 0
		if !isTouch {
			code = parseAutoKeyCode(keyStr)
			if code <= 0 {
				continue
			}
			switch name {
			case autoKeyRolePet:
				keyPetCode = code
			case autoKeyRoleJump:
				keyJumpCode = code
			case autoKeyRoleTeleport:
				keyTeleportCode = code
			case autoKeyRoleAttack:
				keyAttackCode = code
			}
		}
		if configString(m, "模式") != autoKeyModeTimed {
			continue
		}
		minSec := configInt(m, "间隔最短秒", 0)
		maxSec := configInt(m, "间隔最长秒", 0)
		if minSec <= 0 {
			minSec = 60
		}
		if maxSec <= 0 {
			maxSec = minSec
		}
		if maxSec < minSec {
			maxSec = minSec
		}
		slot := timedAutoKeySlot{
			name:     name,
			code:     code,
			minSec:   minSec,
			maxSec:   maxSec,
			longHold: name == autoKeyRolePet || isTouch,
		}
		if isTouch {
			slot.touch = true
			slot.x1, slot.y1, slot.x2, slot.y2 = region[0], region[1], region[2], region[3]
		}
		timedAutoKeys = append(timedAutoKeys, slot)
	}

	if len(timedAutoKeys) == 0 {
		applyDefaultTimedPetLocked()
	}

	autoKeysLoaded = true
	touchN := 0
	for _, s := range timedAutoKeys {
		if s.touch {
			touchN++
		}
	}
	fmt.Printf("[自动按键] 攻击=%d 跳跃=%d 瞬移=%d 宠物=%d 定时=%d(屏幕槽=%d)\n",
		keyAttackCode, keyJumpCode, keyTeleportCode, keyPetCode, len(timedAutoKeys), touchN)
}

func applyDefaultTimedPetLocked() {
	timedAutoKeys = []timedAutoKeySlot{{
		name: "宠物药水", code: keyPetCode, minSec: 600, maxSec: 700, longHold: true,
	}}
}

func parseScreenHotkey(s string) (region [4]int, ok bool) {
	s = strings.TrimSpace(s)
	region, ok = screenHotkeyRegions[s]
	return region, ok
}

func configString(m map[string]interface{}, key string) string {
	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}
	switch x := v.(type) {
	case string:
		return strings.TrimSpace(x)
	default:
		return strings.TrimSpace(fmt.Sprint(x))
	}
}

func configBool(m map[string]interface{}, key string, def bool) bool {
	v, ok := m[key]
	if !ok || v == nil {
		return def
	}
	switch x := v.(type) {
	case bool:
		return x
	case string:
		s := strings.TrimSpace(strings.ToLower(x))
		return s == "true" || s == "1" || s == "是"
	case float64:
		return x != 0
	case int:
		return x != 0
	default:
		return def
	}
}

func configInt(m map[string]interface{}, key string, def int) int {
	v, ok := m[key]
	if !ok || v == nil {
		return def
	}
	switch x := v.(type) {
	case float64:
		return int(x)
	case int:
		return x
	case int64:
		return int(x)
	case string:
		var n int
		fmt.Sscanf(strings.TrimSpace(x), "%d", &n)
		return n
	default:
		return def
	}
}

func parseAutoKeyCode(s string) int {
	s = strings.TrimSpace(s)
	switch strings.ToUpper(s) {
	case "空格", "SPACE":
		return motion.KEYCODE_SPACE
	case "左SHIFT", "LEFT_SHIFT", "LSHIFT", "SHIFT_LEFT":
		return motion.KEYCODE_SHIFT_LEFT
	case "右SHIFT", "RIGHT_SHIFT", "RSHIFT", "SHIFT_RIGHT":
		return motion.KEYCODE_SHIFT_RIGHT
	}
	if s == "" {
		return 0
	}
	r := []rune(s)
	if len(r) == 1 {
		c := r[0]
		if c >= 'A' && c <= 'Z' {
			return int(c-'A') + 29
		}
		if c >= 'a' && c <= 'z' {
			return int(c-'a') + 29
		}
	}
	return 0
}

func randSecDuration(minSec, maxSec int) time.Duration {
	if maxSec < minSec {
		maxSec = minSec
	}
	if maxSec <= minSec {
		return time.Duration(minSec) * time.Second
	}
	return time.Duration(minSec+rand.Intn(maxSec-minSec+1)) * time.Second
}

func initTimedAutoKeysSchedule() {
	ensureAutoKeysLoaded()
	autoKeysMu.Lock()
	defer autoKeysMu.Unlock()
	now := time.Now()
	for i := range timedAutoKeys {
		timedAutoKeys[i].nextAt = now.Add(randSecDuration(timedAutoKeys[i].minSec, timedAutoKeys[i].maxSec))
	}
}

func sleepTimedAutoKeyBatchGap() {
	core.RandomSleep(timedAutoKeyBatchGapMinMs, timedAutoKeyBatchGapMaxMs)
}

func pressDueTimedAutoKeys(now time.Time, startup bool) {
	batchIdx := 0
	for i := range timedAutoKeys {
		s := &timedAutoKeys[i]
		if !startup && now.Before(s.nextAt) {
			continue
		}
		if batchIdx > 0 {
			autoKeysMu.Unlock()
			sleepTimedAutoKeyBatchGap()
			autoKeysMu.Lock()
		}
		pressTimedAutoKeySlot(s, startup)
		wait := randSecDuration(s.minSec, s.maxSec)
		s.nextAt = time.Now().Add(wait)
		farmLog("维护: %s 下次约 %.0fs 后", s.name, wait.Seconds())
		batchIdx++
	}
}

func pressTimedAutoKeySlot(s *timedAutoKeySlot, startup bool) {
	if s.touch {
		if startup {
			farmLog("维护: 启动 %s(屏幕长按[%d,%d,%d,%d])", s.name, s.x1, s.y1, s.x2, s.y2)
		} else {
			farmLog("维护: %s(屏幕长按[%d,%d,%d,%d])", s.name, s.x1, s.y1, s.x2, s.y2)
		}
		core.RandomLongClickInArea(s.x1, s.y1, s.x2, s.y2, screenHotkeyHoldMinMs, screenHotkeyHoldMaxMs)
		return
	}
	if s.longHold {
		if startup {
			farmLog("维护: 启动 %s(长按)", s.name)
		} else {
			farmLog("维护: %s(长按)", s.name)
		}
		keyHoldPress(s.code, keyHoldPetMin, keyHoldPetMax)
		return
	}
	if startup {
		farmLog("维护: 启动 %s(技能长按)", s.name)
	} else {
		farmLog("维护: %s(技能长按)", s.name)
	}
	keyHoldPress(s.code, keyHoldSkillMin, keyHoldSkillMax)
}

// fireTimedAutoKeysOnceAtStart 挂机开始时立即按一次各定时自动键，并安排下次触发。
func fireTimedAutoKeysOnceAtStart() {
	ensureAutoKeysLoaded()
	autoKeysMu.Lock()
	defer autoKeysMu.Unlock()
	if len(timedAutoKeys) == 0 {
		return
	}
	pressDueTimedAutoKeys(time.Now(), true)
}

func tickTimedAutoKeys(now time.Time) {
	autoKeysMu.Lock()
	defer autoKeysMu.Unlock()
	pressDueTimedAutoKeys(now, false)
}

// ReloadAutoKeysFromConfig 重新读取「自动按键」配置（如登录后）。
func ReloadAutoKeysFromConfig() {
	autoKeysMu.Lock()
	autoKeysLoaded = false
	autoKeysMu.Unlock()
	loadAutoKeysFromConfig()
	initTimedAutoKeysSchedule()
}

func attackKeyCode() int {
	ensureAutoKeysLoaded()
	autoKeysMu.RLock()
	defer autoKeysMu.RUnlock()
	return keyAttackCode
}

func jumpKeyCode() int {
	ensureAutoKeysLoaded()
	autoKeysMu.RLock()
	defer autoKeysMu.RUnlock()
	return keyJumpCode
}

func teleportKeyCode() int {
	ensureAutoKeysLoaded()
	autoKeysMu.RLock()
	defer autoKeysMu.RUnlock()
	return keyTeleportCode
}

func petKeyCode() int {
	ensureAutoKeysLoaded()
	autoKeysMu.RLock()
	defer autoKeysMu.RUnlock()
	return keyPetCode
}
