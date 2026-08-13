package core

import (
	"sync"

	"github.com/Dasongzi1366/AutoGo/images"
)

const (
	//minimapYellowX1 = 0
	//minimapYellowY1 = 3
	//minimapYellowX2 = 285
	//minimapYellowY2 = 208
	minimapYellowX1         = 6
	minimapYellowY1         = 65
	minimapYellowX2         = 117
	minimapYellowY2         = 135
	minimapYellowColor      = "ffff21"
	MinimapYellowSimDefault = float32(0.85) // 韩服等小地图
	MinimapYellowSimLand    = float32(0.92) // land 客户端小地图
	minimapYellowWindow     = 5
	minimapYellowMinInWin   = 4 // 5×5 内至少 4 个像素命中才认为找到黄点
)

var (
	minimapYellowRegionMu sync.RWMutex
	minimapYellowRegion   [4]int // x1,y1,x2,y2；全 0 表示用默认区域
	minimapYellowSimMu    sync.RWMutex
	minimapYellowSim      = MinimapYellowSimDefault
)

// SetMinimapYellowRegion 覆盖小地图黄点检测区域；传 x2<=x1 或 y2<=y1 则恢复默认。
func SetMinimapYellowRegion(x1, y1, x2, y2 int) {
	minimapYellowRegionMu.Lock()
	defer minimapYellowRegionMu.Unlock()
	if x2 <= x1 || y2 <= y1 {
		minimapYellowRegion = [4]int{}
		return
	}
	minimapYellowRegion = [4]int{x1, y1, x2, y2}
}

// ClearMinimapYellowRegion 恢复默认黄点检测区域与相似度。
func ClearMinimapYellowRegion() {
	SetMinimapYellowRegion(0, 0, 0, 0)
	ClearMinimapYellowSim()
}

// SetMinimapYellowSim 覆盖小地图黄点颜色相似度；sim<=0 则恢复默认。
func SetMinimapYellowSim(sim float32) {
	minimapYellowSimMu.Lock()
	defer minimapYellowSimMu.Unlock()
	if sim <= 0 {
		minimapYellowSim = MinimapYellowSimDefault
		return
	}
	minimapYellowSim = sim
}

// ClearMinimapYellowSim 恢复默认黄点相似度（韩服 0.85）。
func ClearMinimapYellowSim() {
	SetMinimapYellowSim(0)
}

// MinimapYellowCurrentSim 当前黄点颜色相似度（含 SetMinimapYellowSim 覆盖）。
func MinimapYellowCurrentSim() float32 {
	return currentMinimapYellowSim()
}

func currentMinimapYellowSim() float32 {
	minimapYellowSimMu.RLock()
	defer minimapYellowSimMu.RUnlock()
	return minimapYellowSim
}

func minimapYellowBounds() (x1, y1, x2, y2 int) {
	minimapYellowRegionMu.RLock()
	r := minimapYellowRegion
	minimapYellowRegionMu.RUnlock()
	if r[2] > r[0] && r[3] > r[1] {
		return r[0], r[1], r[2], r[3]
	}
	return minimapYellowX1, minimapYellowY1, minimapYellowX2, minimapYellowY2
}

// MinimapYellowSearchRegion 当前黄点搜索区域（含 SetMinimapYellowRegion 覆盖）。
func MinimapYellowSearchRegion() (x1, y1, x2, y2 int) {
	return minimapYellowBounds()
}

// FindMinimapYellowCenter 在指定区域内用颜色匹配找黄点，返回 5×5 窗口匹配数最多的中心坐标；未找到返回 (-1,-1)。
func FindMinimapYellowCenter() (int, int) {
	x1, y1, x2, y2 := minimapYellowBounds()
	w := x2 - x1 + 1
	h := y2 - y1 + 1
	grid := make([]byte, w*h)

	Color.mu.Lock()
	sim := currentMinimapYellowSim()
	for y := y1; y <= y2; y++ {
		for x := x1; x <= x2; x++ {
			if images.CmpColor(x, y, minimapYellowColor, sim, 0) {
				grid[(y-y1)*w+(x-x1)] = 1
			}
		}
	}
	Color.mu.Unlock()

	win := minimapYellowWindow
	best, cx, cy := 0, -1, -1
	for wy := 0; wy <= h-win; wy++ {
		for wx := 0; wx <= w-win; wx++ {
			n := 0
			base := wy*w + wx
			for dy := 0; dy < win; dy++ {
				row := base + dy*w
				for dx := 0; dx < win; dx++ {
					if grid[row+dx] != 0 {
						n++
					}
				}
			}
			if n > best {
				best = n
				cx = x1 + wx + win/2
				cy = y1 + wy + win/2
			}
		}
	}
	if best < minimapYellowMinInWin {
		return -1, -1
	}
	return cx, cy
}

// CountMinimapColorDotsInRegion 在指定区域内统计独立色点数量（5×5 窗口峰值，逐个抹除后重扫）。
func CountMinimapColorDotsInRegion(x1, y1, x2, y2 int, colorStr string, sim float32, win, minInWin int) int {
	if x2 <= x1 || y2 <= y1 || win <= 0 || minInWin <= 0 {
		return 0
	}
	w := x2 - x1 + 1
	h := y2 - y1 + 1
	grid := make([]byte, w*h)

	Color.mu.Lock()
	for y := y1; y <= y2; y++ {
		for x := x1; x <= x2; x++ {
			if images.CmpColor(x, y, colorStr, sim, 0) {
				grid[(y-y1)*w+(x-x1)] = 1
			}
		}
	}
	Color.mu.Unlock()

	count := 0
	half := win / 2
	for {
		best, bx, by := 0, -1, -1
		for wy := 0; wy <= h-win; wy++ {
			for wx := 0; wx <= w-win; wx++ {
				n := 0
				base := wy*w + wx
				for dy := 0; dy < win; dy++ {
					row := base + dy*w
					for dx := 0; dx < win; dx++ {
						if grid[row+dx] != 0 {
							n++
						}
					}
				}
				if n > best {
					best = n
					bx = wx
					by = wy
				}
			}
		}
		if best < minInWin {
			break
		}
		count++
		cx := bx + half
		cy := by + half
		for y := cy - half; y <= cy+half; y++ {
			if y < 0 || y >= h {
				continue
			}
			for x := cx - half; x <= cx+half; x++ {
				if x < 0 || x >= w {
					continue
				}
				grid[y*w+x] = 0
			}
		}
	}
	return count
}

// CountMinimapColorDots 在当前小地图区域（同 delete_yellow / SetMinimapYellowRegion）统计色点数量。
func CountMinimapColorDots(colorStr string, sim float32) int {
	x1, y1, x2, y2 := minimapYellowBounds()
	return CountMinimapColorDotsInRegion(x1, y1, x2, y2, colorStr, sim, minimapYellowWindow, minimapYellowMinInWin)
}

// RecordMinimapDetect 记录一次小地图黄点+世界锚点联合检测是否成功（供统计/扩展）。
func RecordMinimapDetect(ok bool) {
	_ = ok
}
