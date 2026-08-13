package huashu

import (
	"sync"

	"github.com/Dasongzi1366/AutoGo/hud"
	"github.com/Dasongzi1366/AutoGo/utils"
)

const (
	trackHUDPointHalf  = 8
	trackHUDBackground = "#ff00ff"
	trackHUDIntervalMs = 16
)

var (
	trackHUDOnce  sync.Once
	trackHUD      *hud.HUD
	trackHUDPosMu sync.Mutex
	trackHUDPos   struct{ x, y int }
	trackHUDDirty bool
)

func initTrackHUD() {
	trackHUDOnce.Do(func() {
		trackHUD = hud.New()
		trackHUD.SetBackgroundColor(trackHUDBackground)
		trackHUD.Show()
		go func() {
			for {
				utils.Sleep(trackHUDIntervalMs)
				trackHUDPosMu.Lock()
				if !trackHUDDirty {
					trackHUDPosMu.Unlock()
					continue
				}
				x, y := trackHUDPos.x, trackHUDPos.y
				trackHUDDirty = false
				trackHUDPosMu.Unlock()
				half := trackHUDPointHalf
				trackHUD.SetPosition(x-half, y-half, x+half, y+half)
			}
		}()
	})
}

func showTrackTarget(x, y int) {
	initTrackHUD()
	trackHUDPosMu.Lock()
	trackHUDPos.x, trackHUDPos.y = x, y
	trackHUDDirty = true
	trackHUDPosMu.Unlock()
}
