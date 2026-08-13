package play

import (
	"fmt"

	"app/core"
)

const (
	defaultWorldX1   = 0
	defaultWorldY1   = 0
	defaultWorldX2   = 397
	defaultWorldY2   = 44
	defaultWorldImgs = "img/game/左上角.png,img/game/左上角2.png,img/game/左上角3.png,img/game/左上角4.png"

	// 载入黑幕判定：左上角一片近似纯黑则本轮认为未就绪。
	minimapBlackProbeX1          = 32
	minimapBlackProbeY1          = 15
	minimapBlackProbeX2          = 115
	minimapBlackProbeY2          = 71
	minimapBlackProbeHex         = "000000"
	minimapBlackProbeSim float32 = 1.0
	minimapBlackPixelGe          = 4000
)

var (
	worldProbeX1   = defaultWorldX1
	worldProbeY1   = defaultWorldY1
	worldProbeX2   = defaultWorldX2
	worldProbeY2   = defaultWorldY2
	worldProbeImgs = defaultWorldImgs
)

// SetMinimapWorldProbe 设置世界锚点模板搜索区域（地图专用，结束时应 Reset）。
func SetMinimapWorldProbe(x1, y1, x2, y2 int, imgs string) {
	worldProbeX1, worldProbeY1 = x1, y1
	worldProbeX2, worldProbeY2 = x2, y2
	if imgs != "" {
		worldProbeImgs = imgs
	}
}

// ResetMinimapWorldProbe 恢复默认世界锚点搜索区域。
func ResetMinimapWorldProbe() {
	SetMinimapWorldProbe(defaultWorldX1, defaultWorldY1, defaultWorldX2, defaultWorldY2, defaultWorldImgs)
}

func minimapScreenBlocked() bool {
	n := core.Color.GetColorCountInRegion(
		minimapBlackProbeX1, minimapBlackProbeY1,
		minimapBlackProbeX2, minimapBlackProbeY2,
		minimapBlackProbeHex, minimapBlackProbeSim,
	)
	return n >= minimapBlackPixelGe
}

func iabs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func detectYellow() (int, int) {
	return core.FindMinimapYellowCenter()
}

func detectWorld() (int, int) {
	return core.OpenCV.FindImage(worldProbeX1, worldProbeY1, worldProbeX2, worldProbeY2, worldProbeImgs, false, 1.0, 0.6)
}

// detectYellowThenWorld 挂机热路径：两次 FindImage 连拍容易顶 JNI/截图缓冲；中间固定短睡仅让出时序，不改变坐标语义。
func detectYellowThenWorld() (mx, my, wx, wy int) {
	mx, my = detectYellow()
	core.Sleep(core.JitterMs(14, 0.25))
	wx, wy = detectWorld()
	core.RecordMinimapDetect(mx >= 0 && my >= 0 && wx >= 0 && wy >= 0)
	return
}

func DetectYellowThenWorld() (mx, my, wx, wy int) {
	return detectYellowThenWorld()
}

func relativeToRef(mx, my, wx, wy int) (int, int) {
	return mx - (wx - 50), my - wy
}

func formatMinimapRelFail(mx, my, wx, wy int) string {
	yx1, yy1, yx2, yy2 := core.MinimapYellowSearchRegion()
	var yellow, world string
	if mx < 0 || my < 0 {
		yellow = fmt.Sprintf("黄点:未找到 region=[%d,%d,%d,%d] color=ffff88 sim=%.2f",
			yx1, yy1, yx2, yy2, core.MinimapYellowCurrentSim())
	} else {
		yellow = fmt.Sprintf("黄点:OK(%d,%d)", mx, my)
	}
	if wx < 0 || wy < 0 {
		world = fmt.Sprintf("左上角:未找到 probe=[%d,%d,%d,%d] imgs=%s",
			worldProbeX1, worldProbeY1, worldProbeX2, worldProbeY2, worldProbeImgs)
	} else {
		world = fmt.Sprintf("左上角:OK(%d,%d)", wx, wy)
	}
	extra := ""
	if minimapScreenBlocked() {
		extra = " 黑幕探测=命中(可能载入中)"
	}
	return yellow + " | " + world + extra
}

func ReadMinimapRel() (relX, relY int, ok bool) {
	relX, relY, ok, _ = ReadMinimapRelWithDetail()
	return relX, relY, ok
}

// ReadMinimapRelWithDetail ok=false 时 detail 说明黄点/左上角哪一步失败。
func ReadMinimapRelWithDetail() (relX, relY int, ok bool, detail string) {
	relX, relY, ok, detail, _, _ = ReadMinimapRelWithDetailEx()
	return relX, relY, ok, detail
}

// ReadMinimapRelWithDetailEx 额外返回 worldFound/yellowFound，便于区分「有左上角无黄点」。
func ReadMinimapRelWithDetailEx() (relX, relY int, ok bool, detail string, worldFound, yellowFound bool) {
	mx, my, wx, wy := detectYellowThenWorld()
	yellowFound = mx >= 0 && my >= 0
	worldFound = wx >= 0 && wy >= 0
	if !yellowFound || !worldFound {
		return 0, 0, false, formatMinimapRelFail(mx, my, wx, wy), worldFound, yellowFound
	}
	relX, relY = relativeToRef(mx, my, wx, wy)
	return relX, relY, true,
		fmt.Sprintf("黄点(%d,%d) 左上角(%d,%d) rel=(%d,%d)", mx, my, wx, wy, relX, relY),
		true, true
}
