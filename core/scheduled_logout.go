package core

import (
	"sync/atomic"
)

var scheduledLogoutActive int32

// SetScheduledLogoutActive 定时下线/登录流程进行中。
func SetScheduledLogoutActive(on bool) {
	if on {
		atomic.StoreInt32(&scheduledLogoutActive, 1)
	} else {
		atomic.StoreInt32(&scheduledLogoutActive, 0)
	}
}

// IsScheduledLogoutActive 是否处于定时下线流程。
func IsScheduledLogoutActive() bool {
	return atomic.LoadInt32(&scheduledLogoutActive) == 1
}
