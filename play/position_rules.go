package play

import (
	"app/core"
	"math/rand"
	"strings"
)

// PositionRule 小地图相对坐标命中矩形时执行 action（优先于打怪与路线移动）。
// 直线图中 action 为空时表示巡逻折返边界（x_min/x_max），不参与动作触发。
type PositionRule struct {
	Layer  int    `json:"layer"`
	YMin   int    `json:"y_min"`
	YMax   int    `json:"y_max"`
	XMin   int    `json:"x_min"`
	XMax   int    `json:"x_max"`
	Action string `json:"action"`
}

func (c *mapConfig) positionRules() []PositionRule {
	if c == nil {
		return nil
	}
	n := len(c.Rules) + len(c.Exceptions)
	if n == 0 {
		return nil
	}
	out := make([]PositionRule, 0, n)
	out = append(out, c.Rules...)
	out = append(out, c.Exceptions...)
	return out
}

func matchPositionRule(relX, relY int, rule PositionRule) bool {
	return matchRange(relX, rule.XMin, rule.XMax) && matchRange(relY, rule.YMin, rule.YMax)
}

func executePositionAction(action string) {
	switch strings.TrimSpace(strings.ToLower(action)) {
	case "teleport_left":
		teleportLeftAction()
		core.Sleep(80 + rand.Intn(60))
	case "teleport_right":
		teleportRightAction()
		core.Sleep(80 + rand.Intn(60))
	case "teleport_up", "up_teleport":
		tapUpTeleport()
		core.Sleep(500)
	default:
	}
}

// tryApplyPositionRules 命中任一规则则执行并返回 true（本 tick 不再打怪/走路）。
func tryApplyPositionRules(cfg *mapConfig, relX, relY int) bool {
	for _, rule := range cfg.positionRules() {
		if strings.TrimSpace(rule.Action) == "" {
			continue
		}
		if !matchPositionRule(relX, relY, rule) {
			continue
		}
		farmLog("规则: 命中 x[%d,%d] y[%d,%d] → %s", rule.XMin, rule.XMax, rule.YMin, rule.YMax, rule.Action)
		executePositionAction(rule.Action)
		return true
	}
	return false
}
