package core

import (
	"sync"
	"sync/atomic"

	"github.com/Dasongzi1366/AutoGo/utils"
)

var (
	captchaHoldMu   sync.RWMutex
	captchaHold     bool
	captchaWSActive int32
)

const (
	captchaGrayX1 = 662
	captchaGrayY1 = 497
	captchaGrayX2 = 759
	captchaGrayY2 = 599
)

// SetCaptchaHold 验证码弹出期间设为 true，关闭后设为 false。
func SetCaptchaHold(on bool) {
	captchaHoldMu.Lock()
	captchaHold = on
	captchaHoldMu.Unlock()
}

// IsCaptchaHold 是否处于验证码暂停窗口。
func IsCaptchaHold() bool {
	captchaHoldMu.RLock()
	defer captchaHoldMu.RUnlock()
	return captchaHold
}

// SetCaptchaWSActive 验证码 WS 推流/拖动过程中为 true。
func SetCaptchaWSActive(on bool) {
	if on {
		atomic.StoreInt32(&captchaWSActive, 1)
	} else {
		atomic.StoreInt32(&captchaWSActive, 0)
	}
}

// IsCaptchaWSActive 是否正在执行验证码 WS 流程。
func IsCaptchaWSActive() bool {
	return atomic.LoadInt32(&captchaWSActive) == 1
}

// IsCaptchaUIPresent 验证码灰块区域已出现（含 hold 尚未置 true 的窗口期）。
func IsCaptchaUIPresent() bool {
	return Color.GetColorCountInRegion(captchaGrayX1, captchaGrayY1, captchaGrayX2, captchaGrayY2, "323232", 0.98) > 5000
}

// BlockWhileCaptchaHold 验证码或定时下线期间阻塞；内部只用 utils.Sleep。
func BlockWhileCaptchaHold() {
	for {
		captchaHoldMu.RLock()
		on := captchaHold
		captchaHoldMu.RUnlock()
		if !on && !IsScheduledLogoutActive() {
			return
		}
		utils.Sleep(50)
	}
}
