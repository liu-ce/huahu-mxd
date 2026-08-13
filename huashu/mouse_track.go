package huashu

import (
	"app/TaiBaiYoloV5Ncnn"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Dasongzi1366/AutoGo/utils"
)

const (
	// 易语言 滑鼠追踪 截图区域
	trackDetectX1 = 255
	trackDetectY1 = 93
	trackDetectX2 = 1027
	trackDetectY2 = 625

	trackInitModel     = 5
	trackInitThreshold = float32(0.7)
	trackThreshold     = float32(0.6)
	trackNmsThreshold  = float32(0.55)
	trackInputSize     = 416

	trackRunDuration = 20 * time.Second
	trackFrameDelay  = 50
	trackDragOffsetY = 97
)

// mapLabelToModelIndex 易语言：分类ID 0→model1, 1→model2, 2→model3, 3→model4。
// 本项目 label 也可能是 "1"~"4"。
func mapLabelToModelIndex(label string) int {
	label = strings.TrimSpace(label)
	switch strings.ToUpper(label) {
	case "1", "W":
		return 1
	case "2", "S":
		return 2
	case "3", "Y":
		return 3
	case "4", "Z":
		return 4
	}
	if n, err := strconv.Atoi(label); err == nil {
		if n >= 0 && n <= 3 {
			return n + 1
		}
		if n >= 1 && n <= 4 {
			return n
		}
	}
	return 0
}

func detectionCenter(d TaiBaiYoloV5Ncnn.DetectorResult) (int, int) {
	return d.X + d.W/2, d.Y + d.H/2
}

func identifyTrackModel(m *ModelManager) (int, error) {
	if err := m.LoadModel(trackInitModel); err != nil {
		return 0, fmt.Errorf("加载 model5: %w", err)
	}
	dets, err := m.DetectRegion(
		trackDetectX1, trackDetectY1, trackDetectX2, trackDetectY2,
		trackInitThreshold, trackNmsThreshold, trackInputSize,
	)
	if err != nil {
		return 0, err
	}
	if len(dets) == 0 {
		return 0, fmt.Errorf("model5 未识别到形状")
	}
	idx := mapLabelToModelIndex(dets[0].Label)
	if idx == 0 {
		return 0, fmt.Errorf("未知分类 label=%q", dets[0].Label)
	}
	fmt.Printf("[滑鼠] 识别形状 label=%q model=%d\n", dets[0].Label, idx)
	return idx, nil
}

// StartTrack 易语言 滑鼠追踪：model5 定类型 → 对应 model 逐帧取第一个框拖动。
func StartTrack(m *ModelManager) {
	if m == nil {
		var err error
		m, err = InitMouseModels()
		if err != nil {
			fmt.Println("[滑鼠] 模型初始化失败:", err)
			return
		}
	}

	deadline := time.Now().Add(trackRunDuration)
	trackModel := 0
	dragging := false

	for time.Now().Before(deadline) {
		if trackModel == 0 {
			idx, err := identifyTrackModel(m)
			if err != nil {
				utils.Sleep(trackFrameDelay)
				continue
			}
			trackModel = idx
			if err := m.LoadModel(trackModel); err != nil {
				fmt.Println("[滑鼠] 切换模型失败:", err)
				return
			}
			mousePressDown()
			dragging = true
			fmt.Println("[滑鼠] 开始追踪")
			continue
		}

		dets, err := m.DetectRegion(
			trackDetectX1, trackDetectY1, trackDetectX2, trackDetectY2,
			trackThreshold, trackNmsThreshold, trackInputSize,
		)
		if err != nil {
			utils.Sleep(trackFrameDelay)
			continue
		}
		if len(dets) > 0 {
			cx, cy := detectionCenter(dets[0])
			fmt.Printf("[滑鼠] target=(%d,%d)\n", cx, cy)
			showTrackTarget(cx, cy)
			touchMoveEx(cx, cy+trackDragOffsetY, 0)
		}
		utils.Sleep(trackFrameDelay)
	}

	if dragging {
		touchUpFinger()
	}
	fmt.Println("[滑鼠] 追踪结束")
}
