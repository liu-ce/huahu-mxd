package play

import (
	"app/core"
	"strings"
)

// MapTypeFashionAvenue 时尚大道定制两层环形（D→C→B→A）。
const MapTypeFashionAvenue = "定制_时尚大道"

// MapTypeInPlaceLR 原地左右交替攻击（无移动/YOLO）。
const MapTypeInPlaceLR = "原地左右打"

// MapTypeLandInPlaceLR land/韩服原地左右打：独立按键与血蓝维护。
const MapTypeLandInPlaceLR = "land原地左右打"

// MapTypeLandHelleWalk land赫勒地区走路版：x∈[40,70] 走路 + 左右攻击。
const MapTypeLandHelleWalk = "land赫勒地区走路版"

// MapTypeLandHelleTeleport land赫勒地区瞬移版：x∈[20,90] 点击瞬移 + 0～3 次攻击。
const MapTypeLandHelleTeleport = "land赫勒地区瞬移版"

// MapTypeInstituteC1 研究所 C1：原地持续攻击 + 台阶站位维持。
const MapTypeInstituteC1 = "研究所C1"

// MapTypeInstituteC2 研究所 C2：仅定时韩文 OCR 监控（无挂机）。
const MapTypeInstituteC2 = "研究所C2"

// MapTypeInstituteC1LeftPlatform 研究所 C1 左站台：x=9±3 原地攻击 + 掉阶恢复。
const MapTypeInstituteC1LeftPlatform = "研究所C1左站台"

// MapTypeLinear 单层直线图：左右往返巡逻，边走边 YOLO 打怪。
const MapTypeLinear = "直线图"

// MapTypeTreasureIsland 抢夺宝物岛等多层复杂图。
const MapTypeTreasureIsland = "抢夺宝物岛"

// MapTypeYeqiu001 叶秋001：下层瞬移刷怪 + 爬绳上层短刷 + 右跳下层。
const MapTypeYeqiu001 = "叶秋001"

// MapTypeLuffy001 路飞001：同叶秋001，全程无瞬移，走路+攻击 1～2s。
const MapTypeLuffy001 = "路飞001"

// MapTypeLangligelang001 浪里个浪001：下层 x 区间来回瞬移+攻击。
const MapTypeLangligelang001 = "浪里个浪001"

// MapTypeZaozhi001 早吱定制001：上下平台来回瞬移刷怪 + 楼梯爬升。
const MapTypeZaozhi001 = "早吱定制001"

// MapTypeYexiongLingdi 野熊的领地：中下平台瞬移刷怪 + 绳子换层。
const MapTypeYexiongLingdi = "野熊的领地"

// Langligelang001Config 浪里个浪001 双层参数（仅下层挂机）。
type Langligelang001Config struct {
	LowerXMin                    int   `json:"lower_x_min"`
	LowerXMax                    int   `json:"lower_x_max"`
	LowerYMin                    int   `json:"lower_y_min"`
	LowerYMax                    int   `json:"lower_y_max"`
	UpperYMin                    int   `json:"upper_y_min"`
	UpperYMax                    int   `json:"upper_y_max"`
	PatrolTeleportMargin         int   `json:"patrol_teleport_margin"`
	TeleportHoldMinMs            int   `json:"teleport_hold_min_ms"`
	TeleportHoldMaxMs            int   `json:"teleport_hold_max_ms"`
	AfterTeleportWaitMinMs       int   `json:"after_teleport_wait_min_ms"`
	AfterTeleportWaitMaxMs       int   `json:"after_teleport_wait_max_ms"`
	AttackHoldMinMs              int   `json:"attack_hold_min_ms"`
	AttackHoldMaxMs              int   `json:"attack_hold_max_ms"`
	DownJumpCheckMs              int   `json:"down_jump_check_ms"`
	DownJumpMaxRetry             int   `json:"down_jump_max_retry"`
	ClearBagXMin                 int   `json:"clear_bag_x_min"`
	ClearBagXMax                 int   `json:"clear_bag_x_max"`
	ClearBagJumpCheckMs          int   `json:"clear_bag_jump_check_ms"`
	ClearBagUpperWaitMs          int   `json:"clear_bag_upper_wait_ms"`
	ClearBagIntervalMinMin       int   `json:"clear_bag_interval_min_minutes"`
	ClearBagIntervalMaxMin       int   `json:"clear_bag_interval_max_minutes"`
	ClearBagAlignNearWalkDistMax int   `json:"clear_bag_align_near_walk_dist_max"`
	ClearBagAlignWalkMsPerDist   int   `json:"clear_bag_align_walk_ms_per_dist"`
	ClearBagAlignWalkMsMin       int   `json:"clear_bag_align_walk_ms_min"`
	ClearBagAlignWalkMsMax       int   `json:"clear_bag_align_walk_ms_max"`
	ClearBagSellMisc             *bool `json:"clear_bag_sell_misc"`
	DeleteYellow                 []int `json:"delete_yellow"`
}

func (c *Langligelang001Config) normalize() {
	if c == nil {
		return
	}
	if c.LowerXMin == 0 && c.LowerXMax == 0 {
		c.LowerXMin, c.LowerXMax = -85, 110
	}
	if c.LowerYMin == 0 && c.LowerYMax == 0 {
		c.LowerYMin, c.LowerYMax = 145, 155
	}
	if c.UpperYMin == 0 && c.UpperYMax == 0 {
		c.UpperYMin, c.UpperYMax = 117, 127
	}
	if c.PatrolTeleportMargin <= 0 {
		c.PatrolTeleportMargin = 12
	}
	if c.TeleportHoldMinMs <= 0 {
		c.TeleportHoldMinMs = 50
	}
	if c.TeleportHoldMaxMs < c.TeleportHoldMinMs {
		c.TeleportHoldMaxMs = c.TeleportHoldMinMs + 50
	}
	if c.AfterTeleportWaitMinMs <= 0 {
		c.AfterTeleportWaitMinMs = 20
	}
	if c.AfterTeleportWaitMaxMs < c.AfterTeleportWaitMinMs {
		c.AfterTeleportWaitMaxMs = c.AfterTeleportWaitMinMs + 30
	}
	if c.AttackHoldMinMs <= 0 {
		c.AttackHoldMinMs = 100
	}
	if c.AttackHoldMaxMs < c.AttackHoldMinMs {
		c.AttackHoldMaxMs = c.AttackHoldMinMs + 50
	}
	if c.DownJumpCheckMs <= 0 {
		c.DownJumpCheckMs = 200
	}
	if c.DownJumpMaxRetry <= 0 {
		c.DownJumpMaxRetry = 5
	}
	if c.ClearBagXMin == 0 && c.ClearBagXMax == 0 {
		c.ClearBagXMin, c.ClearBagXMax = 86, 88
	}
	if c.ClearBagJumpCheckMs <= 0 {
		c.ClearBagJumpCheckMs = 500
	}
	if c.ClearBagUpperWaitMs <= 0 {
		c.ClearBagUpperWaitMs = 4300
	}
	if c.ClearBagIntervalMinMin <= 0 {
		c.ClearBagIntervalMinMin = 10
	}
	if c.ClearBagIntervalMaxMin < c.ClearBagIntervalMinMin {
		c.ClearBagIntervalMaxMin = c.ClearBagIntervalMinMin
	}
	if c.ClearBagIntervalMaxMin <= 0 {
		c.ClearBagIntervalMaxMin = 15
	}
	if c.ClearBagAlignNearWalkDistMax <= 0 {
		c.ClearBagAlignNearWalkDistMax = 6
	}
	if c.ClearBagAlignWalkMsPerDist <= 0 {
		c.ClearBagAlignWalkMsPerDist = 45
	}
	if c.ClearBagAlignWalkMsMin <= 0 {
		c.ClearBagAlignWalkMsMin = 60
	}
	if c.ClearBagAlignWalkMsMax < c.ClearBagAlignWalkMsMin {
		c.ClearBagAlignWalkMsMax = c.ClearBagAlignWalkMsMin + 200
	}
}

func (c *Langligelang001Config) clearBagSellMisc() bool {
	if c == nil || c.ClearBagSellMisc == nil {
		return true
	}
	return *c.ClearBagSellMisc
}

func (c *Langligelang001Config) applyDeleteYellowRegion() {
	if c == nil {
		return
	}
	applyDeleteYellowRegion(c.DeleteYellow)
}

// Zaozhi001Config 早吱定制001 双层平台与三楼梯爬升参数。
type Zaozhi001Config struct {
	UpperXMin                int   `json:"upper_x_min"`
	UpperXMax                int   `json:"upper_x_max"`
	UpperYMin                int   `json:"upper_y_min"`
	UpperYMax                int   `json:"upper_y_max"`
	UpperYDetectMinOffset    int   `json:"upper_y_detect_min_offset"`
	LowerXMin                int   `json:"lower_x_min"`
	LowerXMax                int   `json:"lower_x_max"`
	LowerYMin                int   `json:"lower_y_min"`
	LowerYMax                int   `json:"lower_y_max"`
	PatrolTeleportMargin     int   `json:"patrol_teleport_margin"`
	AttackHoldMinMs          int   `json:"attack_hold_min_ms"`
	AttackHoldMaxMs          int   `json:"attack_hold_max_ms"`
	UpperLapsMin             int   `json:"upper_laps_min"`
	UpperLapsMax             int   `json:"upper_laps_max"`
	LowerLapsMin             int   `json:"lower_laps_min"`
	LowerLapsMax             int   `json:"lower_laps_max"`
	DownJumpX1Center         int   `json:"down_jump_x1_center"`
	DownJumpX1Tol            int   `json:"down_jump_x1_tolerance"`
	DownJumpX2Center         int   `json:"down_jump_x2_center"`
	DownJumpX2Tol            int   `json:"down_jump_x2_tolerance"`
	DownJumpCheckMs          int   `json:"down_jump_check_ms"`
	DownJumpSuccessYAbove    int   `json:"down_jump_success_y_above"`
	DownJumpMaxRetry         int   `json:"down_jump_max_retry"`
	DescendPollMs            int   `json:"descend_poll_ms"`
	DescendPollMax           int   `json:"descend_poll_max"`
	StairXs                  []int `json:"stair_xs"`
	StairXTolerance          int   `json:"stair_x_tolerance"`
	StairJumpXTolerance      int   `json:"stair_jump_x_tolerance"`
	StairNearWalkDist        int   `json:"stair_near_walk_dist"`
	StairNearWalkMs          int   `json:"stair_near_walk_ms"`
	StairNearStepMs          int   `json:"stair_near_step_ms"`
	StairNearMicroStepMs     int   `json:"stair_near_micro_step_ms"`
	StairAlignWalkMsPerDist  int   `json:"stair_align_walk_ms_per_dist"`
	StairAlignWalkMsMin      int   `json:"stair_align_walk_ms_min"`
	StairAlignWalkMsMax      int   `json:"stair_align_walk_ms_max"`
	AlignMaxPass             int   `json:"align_max_pass"`
	ClimbJumpWaitMs          int   `json:"climb_jump_wait_ms"`
	ClimbAttackJumpWaitMsMin int   `json:"climb_attack_jump_wait_ms_min"`
	ClimbAttackJumpWaitMsMax int   `json:"climb_attack_jump_wait_ms_max"`
	ClimbUpTapMs             int   `json:"climb_up_tap_ms"`
	ClimbLowerRetryMs        int   `json:"climb_lower_retry_ms"`
	ClimbMinSec              int   `json:"climb_min_sec"`
	ClimbMaxSec              int   `json:"climb_max_sec"`
	MidLayerStableSec        int   `json:"mid_layer_stable_sec"`
	ClearBagXCenter          int   `json:"clear_bag_x_center"`
	ClearBagXTol             int   `json:"clear_bag_x_tolerance"`
	ClearBagUpWaitMs         int   `json:"clear_bag_up_wait_ms"`
	ClearBagShopYBelow       int   `json:"clear_bag_shop_y_below"`
	ClearBagIntervalMinMin   int   `json:"clear_bag_interval_min_minutes"`
	ClearBagIntervalMaxMin   int   `json:"clear_bag_interval_max_minutes"`
	ClearBagSellMisc         *bool `json:"clear_bag_sell_misc"`
	DeleteYellow             []int `json:"delete_yellow"`
}

func (z *Zaozhi001Config) normalize() {
	if z == nil {
		return
	}
	if z.UpperXMin == 0 && z.UpperXMax == 0 {
		z.UpperXMin, z.UpperXMax = 6, 88
	}
	if z.UpperYMin == 0 && z.UpperYMax == 0 {
		z.UpperYMin, z.UpperYMax = 128, 136
	}
	if z.UpperYDetectMinOffset <= 0 {
		z.UpperYDetectMinOffset = 2
	}
	if z.LowerXMin == 0 && z.LowerXMax == 0 {
		z.LowerXMin, z.LowerXMax = 6, 88
	}
	if z.LowerYMin == 0 && z.LowerYMax == 0 {
		z.LowerYMin, z.LowerYMax = 155, 161
	}
	if z.PatrolTeleportMargin <= 0 {
		z.PatrolTeleportMargin = 12
	}
	if z.AttackHoldMinMs <= 0 {
		z.AttackHoldMinMs = 200
	}
	if z.AttackHoldMaxMs < z.AttackHoldMinMs {
		z.AttackHoldMaxMs = z.AttackHoldMinMs + 200
	}
	if z.UpperLapsMin <= 0 {
		z.UpperLapsMin = 4
	}
	if z.UpperLapsMax < z.UpperLapsMin {
		z.UpperLapsMax = z.UpperLapsMin + 2
	}
	if z.LowerLapsMin <= 0 {
		z.LowerLapsMin = 1
	}
	if z.LowerLapsMax < z.LowerLapsMin {
		z.LowerLapsMax = z.LowerLapsMin + 1
	}
	if z.DownJumpX1Center == 0 {
		z.DownJumpX1Center = 59
	}
	if z.DownJumpX1Tol <= 0 {
		z.DownJumpX1Tol = 10
	}
	if z.DownJumpX2Center == 0 {
		z.DownJumpX2Center = 23
	}
	if z.DownJumpX2Tol <= 0 {
		z.DownJumpX2Tol = 10
	}
	if z.DownJumpCheckMs <= 0 {
		z.DownJumpCheckMs = 500
	}
	if z.DownJumpSuccessYAbove == 0 {
		z.DownJumpSuccessYAbove = 140
	}
	if z.DownJumpMaxRetry <= 0 {
		z.DownJumpMaxRetry = 5
	}
	if z.DescendPollMs <= 0 {
		z.DescendPollMs = 50
	}
	if z.DescendPollMax <= 0 {
		z.DescendPollMax = 40
	}
	if len(z.StairXs) == 0 {
		z.StairXs = []int{3, 44, 81}
	}
	if z.StairXTolerance <= 0 {
		z.StairXTolerance = 2
	}
	if z.StairJumpXTolerance < 0 {
		z.StairJumpXTolerance = 0
	}
	if z.StairNearWalkDist <= 0 {
		z.StairNearWalkDist = 10
	}
	if z.StairNearWalkMs <= 0 {
		z.StairNearWalkMs = 300
	}
	if z.StairNearStepMs <= 0 {
		z.StairNearStepMs = 55
	}
	if z.StairNearMicroStepMs <= 0 {
		z.StairNearMicroStepMs = 28
	}
	if z.StairAlignWalkMsPerDist <= 0 {
		z.StairAlignWalkMsPerDist = 28
	}
	if z.StairAlignWalkMsMin <= 0 {
		z.StairAlignWalkMsMin = 22
	}
	if z.StairAlignWalkMsMax <= 0 {
		z.StairAlignWalkMsMax = 80
	}
	if z.StairAlignWalkMsMax < z.StairAlignWalkMsMin {
		z.StairAlignWalkMsMax = z.StairAlignWalkMsMin + 40
	}
	if z.AlignMaxPass <= 0 {
		z.AlignMaxPass = 30
	}
	if z.ClimbJumpWaitMs <= 0 {
		z.ClimbJumpWaitMs = 200
	}
	if z.ClimbAttackJumpWaitMsMin <= 0 {
		z.ClimbAttackJumpWaitMsMin = 300
	}
	if z.ClimbAttackJumpWaitMsMax < z.ClimbAttackJumpWaitMsMin {
		z.ClimbAttackJumpWaitMsMax = z.ClimbAttackJumpWaitMsMin + 200
	}
	if z.ClimbUpTapMs <= 0 {
		z.ClimbUpTapMs = 300
	}
	if z.ClimbLowerRetryMs <= 0 {
		z.ClimbLowerRetryMs = 500
	}
	if z.ClimbMinSec <= 0 {
		z.ClimbMinSec = 3
	}
	if z.ClimbMaxSec <= 0 {
		z.ClimbMaxSec = 10
	}
	if z.ClimbMaxSec < z.ClimbMinSec {
		z.ClimbMaxSec = z.ClimbMinSec + 7
	}
	if z.MidLayerStableSec <= 0 {
		z.MidLayerStableSec = 3
	}
	if z.ClearBagXCenter == 0 {
		z.ClearBagXCenter = 57
	}
	if z.ClearBagXTol <= 0 {
		z.ClearBagXTol = 2
	}
	if z.ClearBagUpWaitMs <= 0 {
		z.ClearBagUpWaitMs = 500
	}
	if z.ClearBagIntervalMinMin <= 0 {
		z.ClearBagIntervalMinMin = 50
	}
	if z.ClearBagIntervalMaxMin < z.ClearBagIntervalMinMin {
		z.ClearBagIntervalMaxMin = z.ClearBagIntervalMinMin + 20
	}
	if z.ClearBagIntervalMaxMin <= 0 {
		z.ClearBagIntervalMaxMin = 70
	}
}

func (z *Zaozhi001Config) applyDeleteYellowRegion() {
	if z == nil {
		return
	}
	applyDeleteYellowRegion(z.DeleteYellow)
}

func (z *Zaozhi001Config) downJumpSpot1Min() int {
	return z.DownJumpX1Center - z.DownJumpX1Tol
}

func (z *Zaozhi001Config) downJumpSpot1Max() int {
	return z.DownJumpX1Center + z.DownJumpX1Tol
}

func (z *Zaozhi001Config) downJumpSpot2Min() int {
	return z.DownJumpX2Center - z.DownJumpX2Tol
}

func (z *Zaozhi001Config) downJumpSpot2Max() int {
	return z.DownJumpX2Center + z.DownJumpX2Tol
}

func (z *Zaozhi001Config) nearestDownJumpTarget(relX int) int {
	d1 := relX - z.DownJumpX1Center
	if d1 < 0 {
		d1 = -d1
	}
	d2 := relX - z.DownJumpX2Center
	if d2 < 0 {
		d2 = -d2
	}
	if d1 <= d2 {
		return z.DownJumpX1Center
	}
	return z.DownJumpX2Center
}

func (z *Zaozhi001Config) upperFarmYMin() int {
	off := z.UpperYDetectMinOffset
	if off <= 0 {
		off = 2
	}
	return z.UpperYMin - off
}

// YexiongLingdiConfig 野熊的领地：中下平台巡逻 + 绳子换层（上层平台后续扩展）。
type YexiongLingdiConfig struct {
	LowerXMin                         int   `json:"lower_x_min"`
	LowerXMax                         int   `json:"lower_x_max"`
	LowerYMin                         int   `json:"lower_y_min"`
	LowerYMax                         int   `json:"lower_y_max"`
	LowerLapsMin                      int   `json:"lower_laps_min"`
	LowerLapsMax                      int   `json:"lower_laps_max"`
	LowerPatrolTurnXLeft              int   `json:"lower_patrol_turn_x_left"`
	LowerPatrolTurnXRight             int   `json:"lower_patrol_turn_x_right"`
	MiddleXMin                        int   `json:"middle_x_min"`
	MiddleXMax                        int   `json:"middle_x_max"`
	MiddleAttackXMin                  int   `json:"middle_attack_x_min"`
	MiddleAttackXMax                  int   `json:"middle_attack_x_max"`
	MiddleYMin                        int   `json:"middle_y_min"`
	MiddleYMax                        int   `json:"middle_y_max"`
	MiddleLapsMin                     int   `json:"middle_laps_min"`
	MiddleLapsMax                     int   `json:"middle_laps_max"`
	MiddlePatrolTurnXLeft             int   `json:"middle_patrol_turn_x_left"`
	MiddlePatrolTurnXRight            int   `json:"middle_patrol_turn_x_right"`
	UpperRightYMin                    int   `json:"upper_right_y_min"`
	UpperRightYMax                    int   `json:"upper_right_y_max"`
	UpperRightXMin                    int   `json:"upper_right_x_min"`
	UpperRightXMax                    int   `json:"upper_right_x_max"`
	UpperRightJumpXMin                int   `json:"upper_right_jump_x_min"`
	UpperRightJumpXMax                int   `json:"upper_right_jump_x_max"`
	UpperRightPatrolTurnXLeft         int   `json:"upper_right_patrol_turn_x_left"`
	UpperRightPatrolTurnXRight        int   `json:"upper_right_patrol_turn_x_right"`
	UpperRightDescendXMin             int   `json:"upper_right_descend_x_min"`
	UpperRightDescendXMax             int   `json:"upper_right_descend_x_max"`
	UpperLeftYMin                     int   `json:"upper_left_y_min"`
	UpperLeftYMax                     int   `json:"upper_left_y_max"`
	UpperLeftXMin                     int   `json:"upper_left_x_min"`
	UpperLeftXMax                     int   `json:"upper_left_x_max"`
	UpperLeftJumpXMin                 int   `json:"upper_left_jump_x_min"`
	UpperLeftJumpXMax                 int   `json:"upper_left_jump_x_max"`
	UpperLeftPatrolTurnXLeft          int   `json:"upper_left_patrol_turn_x_left"`
	UpperLeftPatrolTurnXRight         int   `json:"upper_left_patrol_turn_x_right"`
	UpperLeftDescendXMin              int   `json:"upper_left_descend_x_min"`
	UpperLeftDescendXMax              int   `json:"upper_left_descend_x_max"`
	UpperSideClimbYAbove              int   `json:"upper_side_climb_y_above"`
	UpperSideClimbMinMs               int   `json:"upper_side_climb_min_ms"`
	UpperSideAlignApproachOffset      int   `json:"upper_side_align_approach_offset"` // 远距先走到跳点前 N 格（66→64）
	UpperSideAlignCoarseWalkMsPerDist int   `json:"upper_side_align_coarse_walk_ms_per_dist"`
	UpperSideAlignCoarseWalkMsMin     int   `json:"upper_side_align_coarse_walk_ms_min"`
	UpperSideAlignCoarseWalkMsMax     int   `json:"upper_side_align_coarse_walk_ms_max"`
	UpperSideAlignFineWalkMsMax       int   `json:"upper_side_align_fine_walk_ms_max"`
	UpperSideJumpXOvershoot           int   `json:"upper_side_jump_x_overshoot"` // 侧上跳点 x 右侧/绳侧容差，0=右上2左上1
	UpperSideArrivalSettleMs          int   `json:"upper_side_arrival_settle_ms"`
	UpperSideLapsMin                  int   `json:"upper_side_laps_min"`
	UpperSideLapsMax                  int   `json:"upper_side_laps_max"`
	UpperSideClimbMaxAttempts         int   `json:"upper_side_climb_max_attempts"`
	LayerPollMs                       int   `json:"layer_poll_ms"`
	LayerStableCount                  int   `json:"layer_stable_count"`
	DescendXMin                       int   `json:"descend_x_min"`
	DescendXMax                       int   `json:"descend_x_max"`
	RopeJumpRightXMin                 int   `json:"rope_jump_right_x_min"`
	RopeJumpRightXMax                 int   `json:"rope_jump_right_x_max"`
	RopeJumpLeftXMin                  int   `json:"rope_jump_left_x_min"`
	RopeJumpLeftXMax                  int   `json:"rope_jump_left_x_max"`
	RopeJumpFaceMinMs                 int   `json:"rope_jump_face_min_ms"` // 跳绳抓绳前先朝绳方向按键
	RopeJumpFaceMaxMs                 int   `json:"rope_jump_face_max_ms"`
	RopeJumpBeforeMinMs               int   `json:"rope_jump_before_min_ms"` // 朝绳方向按键后、跳跃前等待
	RopeJumpBeforeMaxMs               int   `json:"rope_jump_before_max_ms"`
	RopeAlignXTolerance               int   `json:"rope_align_x_tolerance"`
	RopeJumpSpotXTolerance            int   `json:"rope_jump_spot_x_tolerance"` // 绳跳点 x 容差，近位即视为就位
	RopeAlignWalkMsMin                int   `json:"rope_align_walk_ms_min"`
	RopeAlignWalkMsPerDist            int   `json:"rope_align_walk_ms_per_dist"`
	RopePreUpWaitMinMs                int   `json:"rope_pre_up_wait_min_ms"`
	RopePreUpWaitMaxMs                int   `json:"rope_pre_up_wait_max_ms"`
	RopeUpHoldMs                      int   `json:"rope_up_hold_ms"`
	RopeClimbMaxSec                   int   `json:"rope_climb_max_sec"`
	RopeClimbMaxAttempts              int   `json:"rope_climb_max_attempts"`
	RopeClimbRetryAttackMinMs         int   `json:"rope_climb_retry_attack_min_ms"`
	RopeClimbRetryAttackMaxMs         int   `json:"rope_climb_retry_attack_max_ms"`
	RopeClimbPollMs                   int   `json:"rope_climb_poll_ms"`
	RopeArrivalSettleMs               int   `json:"rope_arrival_settle_ms"`
	RopeAlignMaxPass                  int   `json:"rope_align_max_pass"`
	RopeNearWalkDist                  int   `json:"rope_near_walk_dist"`
	AlignWalkMsPerDist                int   `json:"align_walk_ms_per_dist"`
	AlignWalkMsMin                    int   `json:"align_walk_ms_min"`
	AlignWalkMsMax                    int   `json:"align_walk_ms_max"`
	AlignPreAttackMinMs               int   `json:"align_pre_attack_min_ms"`
	AlignPreAttackMaxMs               int   `json:"align_pre_attack_max_ms"`
	AlignPreWaitMs                    int   `json:"align_pre_wait_ms"`
	AbnormalLayerJumpX                int   `json:"abnormal_layer_jump_x"`
	AbnormalLayerJumpLandMs           int   `json:"abnormal_layer_jump_land_ms"`
	DescendMaxRetry                   int   `json:"descend_max_retry"`
	DownJumpCheckMs                   int   `json:"down_jump_check_ms"`
	DescendPollMs                     int   `json:"descend_poll_ms"`
	DescendPollMax                    int   `json:"descend_poll_max"`
	PatrolTeleportMargin              int   `json:"patrol_teleport_margin"`
	PatrolTeleportSettleMs            int   `json:"patrol_teleport_settle_ms"` // 巡逻瞬移后等小地图黄点刷新再读坐标
	PatrolWalkChancePercent           int   `json:"patrol_walk_chance_percent"`
	PatrolWalkHoldMinMs               int   `json:"patrol_walk_hold_min_ms"`
	PatrolWalkHoldMaxMs               int   `json:"patrol_walk_hold_max_ms"`
	TeleportAttackMinMs               int   `json:"teleport_attack_min_ms"`
	TeleportAttackMaxMs               int   `json:"teleport_attack_max_ms"`
	AttackHoldMinMs                   int   `json:"attack_hold_min_ms"`
	AttackHoldMaxMs                   int   `json:"attack_hold_max_ms"`
	UpperXMin                         int   `json:"upper_x_min"`
	UpperXMax                         int   `json:"upper_x_max"`
	UpperYMin                         int   `json:"upper_y_min"`
	UpperYMax                         int   `json:"upper_y_max"`
	UpperDescendXMin                  int   `json:"upper_descend_x_min"`
	UpperDescendXMax                  int   `json:"upper_descend_x_max"`
	UpperRopeJumpRightXMin            int   `json:"upper_rope_jump_right_x_min"`
	UpperRopeJumpRightXMax            int   `json:"upper_rope_jump_right_x_max"`
	UpperRopeJumpLeftXMin             int   `json:"upper_rope_jump_left_x_min"`
	UpperRopeJumpLeftXMax             int   `json:"upper_rope_jump_left_x_max"`
	UpperClimbMaxAttempts             int   `json:"upper_climb_max_attempts"`
	UpperClimbRetryAttackMs           int   `json:"upper_climb_retry_attack_ms"`
	UpperClimbRetryWaitMs             int   `json:"upper_climb_retry_wait_ms"`
	ClearBagXMin                      int   `json:"clear_bag_x_min"`
	ClearBagXMax                      int   `json:"clear_bag_x_max"`
	ClearBagIntervalMinMin            int   `json:"clear_bag_interval_min_minutes"`
	ClearBagIntervalMaxMin            int   `json:"clear_bag_interval_max_minutes"`
	ClearBagSellMisc                  *bool `json:"clear_bag_sell_misc"`
	DeleteYellow                      []int `json:"delete_yellow"`
}

func (c *YexiongLingdiConfig) normalize() {
	if c == nil {
		return
	}
	if c.LowerXMin == 0 && c.LowerXMax == 0 {
		c.LowerXMin, c.LowerXMax = -42, 118
	}
	if c.LowerYMin == 0 && c.LowerYMax == 0 {
		c.LowerYMin, c.LowerYMax = 168, 172
	}
	if c.LowerLapsMin <= 0 {
		c.LowerLapsMin = 2
	}
	if c.LowerLapsMax < c.LowerLapsMin {
		c.LowerLapsMax = c.LowerLapsMin
	}
	if c.LowerPatrolTurnXLeft == 0 && c.LowerPatrolTurnXRight == 0 {
		c.LowerPatrolTurnXLeft, c.LowerPatrolTurnXRight = -20, 110
	}
	if c.MiddleXMin == 0 && c.MiddleXMax == 0 {
		c.MiddleXMin, c.MiddleXMax = 12, 65
	}
	if c.MiddleAttackXMin == 0 && c.MiddleAttackXMax == 0 {
		c.MiddleAttackXMin, c.MiddleAttackXMax = 23, 55
	}
	if c.MiddleYMin == 0 && c.MiddleYMax == 0 {
		c.MiddleYMin, c.MiddleYMax = 141, 150
	}
	if c.MiddleLapsMin <= 0 {
		c.MiddleLapsMin = 1
	}
	if c.MiddleLapsMax < c.MiddleLapsMin {
		c.MiddleLapsMax = c.MiddleLapsMin
	}
	if c.MiddlePatrolTurnXLeft == 0 && c.MiddlePatrolTurnXRight == 0 {
		c.MiddlePatrolTurnXLeft, c.MiddlePatrolTurnXRight = 25, 45
	}
	if c.UpperRightYMin == 0 && c.UpperRightYMax == 0 {
		c.UpperRightYMin, c.UpperRightYMax = 133, 137
	}
	if c.UpperRightXMin == 0 && c.UpperRightXMax == 0 {
		c.UpperRightXMin, c.UpperRightXMax = 90, 110
	}
	if c.UpperRightJumpXMin == 0 && c.UpperRightJumpXMax == 0 {
		c.UpperRightJumpXMin, c.UpperRightJumpXMax = 64, 66
	}
	if c.UpperRightPatrolTurnXLeft == 0 && c.UpperRightPatrolTurnXRight == 0 {
		c.UpperRightPatrolTurnXLeft, c.UpperRightPatrolTurnXRight = 100, 115
	}
	if c.UpperRightDescendXMin == 0 && c.UpperRightDescendXMax == 0 {
		c.UpperRightDescendXMin, c.UpperRightDescendXMax = 112, 118
	}
	if c.UpperLeftYMin == 0 && c.UpperLeftYMax == 0 {
		c.UpperLeftYMin, c.UpperLeftYMax = 133, 137
	}
	if c.UpperLeftXMin == 0 && c.UpperLeftXMax == 0 {
		c.UpperLeftXMin, c.UpperLeftXMax = -30, -10
	}
	if c.UpperLeftJumpXMin == 0 && c.UpperLeftJumpXMax == 0 {
		c.UpperLeftJumpXMin, c.UpperLeftJumpXMax = 11, 12
	}
	if c.UpperLeftPatrolTurnXLeft == 0 && c.UpperLeftPatrolTurnXRight == 0 {
		c.UpperLeftPatrolTurnXLeft, c.UpperLeftPatrolTurnXRight = -40, -12
	}
	if c.UpperLeftDescendXMin == 0 && c.UpperLeftDescendXMax == 0 {
		c.UpperLeftDescendXMin, c.UpperLeftDescendXMax = -41, -35
	}
	if c.UpperSideClimbYAbove <= 0 {
		c.UpperSideClimbYAbove = 134
	}
	if c.UpperSideClimbMinMs <= 0 {
		c.UpperSideClimbMinMs = 1200
	}
	if c.UpperSideAlignApproachOffset <= 0 {
		c.UpperSideAlignApproachOffset = 2
	}
	if c.UpperSideAlignCoarseWalkMsPerDist <= 0 {
		c.UpperSideAlignCoarseWalkMsPerDist = 55
	}
	if c.UpperSideAlignCoarseWalkMsMin <= 0 {
		c.UpperSideAlignCoarseWalkMsMin = 200
	}
	if c.UpperSideAlignCoarseWalkMsMax <= 0 {
		c.UpperSideAlignCoarseWalkMsMax = 900
	}
	if c.UpperSideAlignFineWalkMsMax <= 0 {
		c.UpperSideAlignFineWalkMsMax = 160
	}
	if c.UpperSideArrivalSettleMs <= 0 {
		c.UpperSideArrivalSettleMs = 700
	}
	if c.UpperSideLapsMin <= 0 {
		c.UpperSideLapsMin = 1
	}
	if c.UpperSideLapsMax < c.UpperSideLapsMin {
		c.UpperSideLapsMax = c.UpperSideLapsMin
	}
	if c.UpperSideClimbMaxAttempts <= 0 {
		c.UpperSideClimbMaxAttempts = 20
	}
	if c.LayerPollMs <= 0 {
		c.LayerPollMs = 200
	}
	if c.LayerStableCount <= 0 {
		c.LayerStableCount = 2
	}
	if c.DescendXMin == 0 && c.DescendXMax == 0 {
		c.DescendXMin, c.DescendXMax = 40, 50
	}
	if c.RopeJumpRightXMin == 0 && c.RopeJumpRightXMax == 0 {
		c.RopeJumpRightXMin, c.RopeJumpRightXMax = 29, 29
	}
	if c.RopeJumpLeftXMin == 0 && c.RopeJumpLeftXMax == 0 {
		c.RopeJumpLeftXMin, c.RopeJumpLeftXMax = 34, 34
	}
	if c.RopeJumpFaceMinMs <= 0 {
		c.RopeJumpFaceMinMs = 20
	}
	if c.RopeJumpFaceMaxMs < c.RopeJumpFaceMinMs {
		c.RopeJumpFaceMaxMs = c.RopeJumpFaceMinMs
	}
	if c.RopeJumpBeforeMinMs <= 0 {
		c.RopeJumpBeforeMinMs = 80
	}
	if c.RopeJumpBeforeMaxMs < c.RopeJumpBeforeMinMs {
		c.RopeJumpBeforeMaxMs = c.RopeJumpBeforeMinMs + 40
	}
	if c.RopeAlignXTolerance <= 0 {
		c.RopeAlignXTolerance = 5
	}
	if c.RopeJumpSpotXTolerance <= 0 {
		c.RopeJumpSpotXTolerance = 2
	}
	if c.RopeAlignWalkMsMin <= 0 {
		c.RopeAlignWalkMsMin = 120
	}
	if c.RopeAlignWalkMsPerDist <= 0 {
		c.RopeAlignWalkMsPerDist = 60
	}
	if c.RopePreUpWaitMinMs <= 0 {
		c.RopePreUpWaitMinMs = 30
	}
	if c.RopePreUpWaitMaxMs < c.RopePreUpWaitMinMs {
		c.RopePreUpWaitMaxMs = c.RopePreUpWaitMinMs + 20
	}
	if c.RopeUpHoldMs <= 0 {
		c.RopeUpHoldMs = 500
	}
	if c.RopeClimbMaxSec <= 0 {
		c.RopeClimbMaxSec = 3
	}
	if c.RopeClimbMaxAttempts <= 0 {
		c.RopeClimbMaxAttempts = 5
	}
	if c.RopeClimbRetryAttackMinMs <= 0 {
		c.RopeClimbRetryAttackMinMs = 400
	}
	if c.RopeClimbRetryAttackMaxMs < c.RopeClimbRetryAttackMinMs {
		c.RopeClimbRetryAttackMaxMs = c.RopeClimbRetryAttackMinMs + 100
	}
	if c.RopeClimbPollMs <= 0 {
		c.RopeClimbPollMs = 100
	}
	if c.RopeArrivalSettleMs <= 0 {
		c.RopeArrivalSettleMs = 500
	}
	if c.RopeAlignMaxPass <= 0 {
		c.RopeAlignMaxPass = 30
	}
	if c.RopeNearWalkDist <= 0 {
		c.RopeNearWalkDist = 8
	}
	if c.AlignWalkMsPerDist <= 0 {
		c.AlignWalkMsPerDist = 40
	}
	if c.AlignWalkMsMin <= 0 {
		c.AlignWalkMsMin = 40
	}
	if c.AlignWalkMsMax <= 0 {
		c.AlignWalkMsMax = 400
	}
	if c.AlignPreAttackMinMs <= 0 {
		c.AlignPreAttackMinMs = 300
	}
	if c.AlignPreAttackMaxMs < c.AlignPreAttackMinMs {
		c.AlignPreAttackMaxMs = c.AlignPreAttackMinMs + 200
	}
	if c.AlignPreWaitMs <= 0 {
		c.AlignPreWaitMs = 300
	}
	if c.AbnormalLayerJumpX <= 0 {
		c.AbnormalLayerJumpX = 35
	}
	if c.AbnormalLayerJumpLandMs <= 0 {
		c.AbnormalLayerJumpLandMs = 500
	}
	if c.DescendMaxRetry <= 0 {
		c.DescendMaxRetry = 5
	}
	if c.DownJumpCheckMs <= 0 {
		c.DownJumpCheckMs = 200
	}
	if c.DescendPollMs <= 0 {
		c.DescendPollMs = 200
	}
	if c.DescendPollMax <= 0 {
		c.DescendPollMax = 40
	}
	if c.PatrolTeleportMargin <= 0 {
		c.PatrolTeleportMargin = 12
	}
	if c.PatrolTeleportSettleMs <= 0 {
		c.PatrolTeleportSettleMs = 200
	}
	if c.PatrolWalkChancePercent <= 0 {
		c.PatrolWalkChancePercent = 10
	}
	if c.PatrolWalkHoldMinMs <= 0 {
		c.PatrolWalkHoldMinMs = 600
	}
	if c.PatrolWalkHoldMaxMs < c.PatrolWalkHoldMinMs {
		c.PatrolWalkHoldMaxMs = c.PatrolWalkHoldMinMs + 600
	}
	if c.TeleportAttackMinMs <= 0 {
		c.TeleportAttackMinMs = 70
	}
	if c.TeleportAttackMaxMs < c.TeleportAttackMinMs {
		c.TeleportAttackMaxMs = c.TeleportAttackMinMs + 50
	}
	if c.AttackHoldMinMs <= 0 {
		c.AttackHoldMinMs = 100
	}
	if c.AttackHoldMaxMs < c.AttackHoldMinMs {
		c.AttackHoldMaxMs = c.AttackHoldMinMs + 100
	}
	if c.UpperXMin == 0 && c.UpperXMax == 0 {
		c.UpperXMin, c.UpperXMax = 0, 75
	}
	if c.UpperYMin == 0 && c.UpperYMax == 0 {
		c.UpperYMin, c.UpperYMax = 118, 122
	}
	if c.UpperDescendXMin == 0 && c.UpperDescendXMax == 0 {
		c.UpperDescendXMin, c.UpperDescendXMax = 30, 43
	}
	if c.UpperRopeJumpRightXMin == 0 && c.UpperRopeJumpRightXMax == 0 {
		c.UpperRopeJumpRightXMin, c.UpperRopeJumpRightXMax = 41, 43
	}
	if c.UpperRopeJumpLeftXMin == 0 && c.UpperRopeJumpLeftXMax == 0 {
		c.UpperRopeJumpLeftXMin, c.UpperRopeJumpLeftXMax = 48, 49
	}
	if c.UpperClimbMaxAttempts <= 0 {
		c.UpperClimbMaxAttempts = 20
	}
	if c.UpperClimbRetryAttackMs <= 0 {
		c.UpperClimbRetryAttackMs = 500
	}
	if c.UpperClimbRetryWaitMs <= 0 {
		c.UpperClimbRetryWaitMs = 300
	}
	if c.ClearBagXMin == 0 && c.ClearBagXMax == 0 {
		c.ClearBagXMin, c.ClearBagXMax = 0, 75
	}
	if c.ClearBagIntervalMinMin <= 0 {
		c.ClearBagIntervalMinMin = 50
	}
	if c.ClearBagIntervalMaxMin < c.ClearBagIntervalMinMin {
		c.ClearBagIntervalMaxMin = c.ClearBagIntervalMinMin + 20
	}
}

// AttackHoldMsFromAPI 读取中控「攻击长按最短/最长毫秒」；未配置时用 defaultMin/defaultMax。
func AttackHoldMsFromAPI(defaultMin, defaultMax int) (minMs, maxMs int) {
	minMs, maxMs = defaultMin, defaultMax
	if core.API == nil {
		return
	}
	if v, err := core.API.GetConfigInt("攻击长按最短毫秒"); err == nil && v > 0 {
		minMs = v
	}
	if v, err := core.API.GetConfigInt("攻击长按最长毫秒"); err == nil && v > 0 {
		maxMs = v
	}
	if maxMs < minMs {
		maxMs = minMs
	}
	return
}

// IntSecRangeFromAPI 读取中控秒数区间配置；未配置时用 defaultMin/defaultMax。
func IntSecRangeFromAPI(minKey, maxKey string, defaultMin, defaultMax int) (minSec, maxSec int) {
	minSec, maxSec = defaultMin, defaultMax
	if core.API == nil {
		return
	}
	if v, err := core.API.GetConfigInt(minKey); err == nil && v > 0 {
		minSec = v
	}
	if v, err := core.API.GetConfigInt(maxKey); err == nil && v > 0 {
		maxSec = v
	}
	if maxSec < minSec {
		maxSec = minSec
	}
	return
}

// PercentPairFromAPI 读取中控两项百分比（如走路/瞬移）；未配置或合计为 0 时用 defaultA/defaultB。
func PercentPairFromAPI(keyA, keyB string, defaultA, defaultB int) (pctA, pctB int) {
	pctA, pctB = defaultA, defaultB
	if core.API == nil {
		return
	}
	if v, err := core.API.GetConfigInt(keyA); err == nil && v >= 0 {
		pctA = v
	}
	if v, err := core.API.GetConfigInt(keyB); err == nil && v >= 0 {
		pctB = v
	}
	if pctA+pctB <= 0 {
		pctA, pctB = defaultA, defaultB
	}
	return
}

// applyAPIConfigOverrides 若中控 API 下发了对应项，覆盖地图 JSON 里的巡逻圈数与攻击长按。
func (c *YexiongLingdiConfig) applyAPIConfigOverrides() bool {
	if c == nil {
		return false
	}
	changed := false
	if v, err := core.API.GetConfigInt("中层巡逻最少次数"); err == nil && v > 0 {
		c.MiddleLapsMin = v
		changed = true
	}
	if v, err := core.API.GetConfigInt("中层巡逻最多次数"); err == nil && v > 0 {
		c.MiddleLapsMax = v
		changed = true
	}
	if v, err := core.API.GetConfigInt("下层巡逻最少次数"); err == nil && v > 0 {
		c.LowerLapsMin = v
		changed = true
	}
	if v, err := core.API.GetConfigInt("下层巡逻最多次数"); err == nil && v > 0 {
		c.LowerLapsMax = v
		changed = true
	}
	if v, err := core.API.GetConfigInt("攻击长按最短毫秒"); err == nil && v > 0 {
		c.AttackHoldMinMs = v
		changed = true
	}
	if v, err := core.API.GetConfigInt("攻击长按最长毫秒"); err == nil && v > 0 {
		c.AttackHoldMaxMs = v
		changed = true
	}
	if c.MiddleLapsMax < c.MiddleLapsMin {
		c.MiddleLapsMax = c.MiddleLapsMin
	}
	if c.LowerLapsMax < c.LowerLapsMin {
		c.LowerLapsMax = c.LowerLapsMin
	}
	if c.AttackHoldMaxMs < c.AttackHoldMinMs {
		c.AttackHoldMaxMs = c.AttackHoldMinMs
	}
	return changed
}

func (c *YexiongLingdiConfig) applyDeleteYellowRegion() {
	if c == nil {
		return
	}
	applyDeleteYellowRegion(c.DeleteYellow)
}

func (c *YexiongLingdiConfig) ropeJumpRightCenter() int {
	return (c.RopeJumpRightXMin + c.RopeJumpRightXMax) / 2
}

func (c *YexiongLingdiConfig) ropeJumpLeftCenter() int {
	return (c.RopeJumpLeftXMin + c.RopeJumpLeftXMax) / 2
}

func (c *YexiongLingdiConfig) middleAttackXMin() int {
	if c.MiddleAttackXMin > 0 {
		return c.MiddleAttackXMin
	}
	return c.MiddleXMin
}

func (c *YexiongLingdiConfig) middleAttackXMax() int {
	if c.MiddleAttackXMax > 0 {
		return c.MiddleAttackXMax
	}
	return c.MiddleXMax
}

func (c *YexiongLingdiConfig) clearBagSellMisc() bool {
	if c.ClearBagSellMisc != nil {
		return *c.ClearBagSellMisc
	}
	return true
}

func (c *YexiongLingdiConfig) upperRopeJumpRightCenter() int {
	return (c.UpperRopeJumpRightXMin + c.UpperRopeJumpRightXMax) / 2
}

func (c *YexiongLingdiConfig) upperRopeJumpLeftCenter() int {
	return (c.UpperRopeJumpLeftXMin + c.UpperRopeJumpLeftXMax) / 2
}

func (z *Zaozhi001Config) nearestStairX(relX int) int {
	if len(z.StairXs) == 0 {
		return 44
	}
	best := z.StairXs[0]
	bestDist := relX - best
	if bestDist < 0 {
		bestDist = -bestDist
	}
	for _, x := range z.StairXs[1:] {
		d := relX - x
		if d < 0 {
			d = -d
		}
		if d < bestDist {
			bestDist = d
			best = x
		}
	}
	return best
}

func (z *Zaozhi001Config) stairJumpXMin(stairX int) int {
	return stairX - z.StairJumpXTolerance
}

func (z *Zaozhi001Config) stairJumpXMax(stairX int) int {
	return stairX + z.StairJumpXTolerance
}

func (z *Zaozhi001Config) clearBagXMin() int {
	return z.ClearBagXCenter - z.ClearBagXTol
}

func (z *Zaozhi001Config) clearBagXMax() int {
	return z.ClearBagXCenter + z.ClearBagXTol
}

func (z *Zaozhi001Config) clearBagShopYBelow() int {
	if z.ClearBagShopYBelow > 0 {
		return z.ClearBagShopYBelow
	}
	return z.UpperYMin
}

func (z *Zaozhi001Config) clearBagSellMisc() bool {
	if z == nil || z.ClearBagSellMisc == nil {
		return true
	}
	return *z.ClearBagSellMisc
}

func (z *Zaozhi001Config) isShopPlatform(relY int) bool {
	return relY < z.clearBagShopYBelow()
}

// Yeqiu001Config 叶秋001 多层与爬绳参数。
type Yeqiu001Config struct {
	UpperXMin                int   `json:"upper_x_min"`
	UpperXMax                int   `json:"upper_x_max"`
	UpperYMin                int   `json:"upper_y_min"`
	UpperYMax                int   `json:"upper_y_max"`
	LowerXMin                int   `json:"lower_x_min"`
	LowerXMax                int   `json:"lower_x_max"`
	LowerYMin                int   `json:"lower_y_min"`
	LowerYMax                int   `json:"lower_y_max"`
	SmallPlatformXCenter     int   `json:"small_platform_x_center"`
	SmallPlatformXTol        int   `json:"small_platform_x_tolerance"`
	SmallPlatformYMin        int   `json:"small_platform_y_min"`
	SmallPlatformYMax        int   `json:"small_platform_y_max"`
	RopeX                    int   `json:"rope_x"`
	RopeLeftXMin             int   `json:"rope_left_x_min"`
	RopeLeftXMax             int   `json:"rope_left_x_max"`
	RopeRightXMin            int   `json:"rope_right_x_min"`
	RopeRightXMax            int   `json:"rope_right_x_max"`
	RopeXTolerance           int   `json:"rope_x_tolerance"`
	RopeJumpXTolerance       int   `json:"rope_jump_x_tolerance"` // 未配 jump_x_min/max 时：rope_x±容差
	RopeJumpXMin             int   `json:"rope_jump_x_min"`       // 爬绳跳抓绳 x 下限
	RopeJumpXMax             int   `json:"rope_jump_x_max"`       // 爬绳跳抓绳 x 上限
	RopeNearWalkDist         int   `json:"rope_near_walk_dist"`
	RopeNearWalkMs           int   `json:"rope_near_walk_ms"`
	RopeNearStepMs           int   `json:"rope_near_step_ms"`       // 距绳较近时小步走
	RopeNearMicroStepMs      int   `json:"rope_near_micro_step_ms"` // 差 1～2 格微调
	LowerFarmSecMin          int   `json:"lower_farm_sec_min"`
	LowerFarmSecMax          int   `json:"lower_farm_sec_max"`
	UpperFarmSecMin          int   `json:"upper_farm_sec_min"`
	UpperFarmSecMax          int   `json:"upper_farm_sec_max"`
	PatrolTeleportMargin     int   `json:"patrol_teleport_margin"` // 近边界提前折返，避免瞬移过冲
	DownJumpCheckMs          int   `json:"down_jump_check_ms"`
	DownJumpMaxRetry         int   `json:"down_jump_max_retry"`
	ClimbFallYAbove          int   `json:"climb_fall_y_above"` // 爬绳中 y 大于此值视为掉到一层
	ClimbNearTopYBelow       int   `json:"climb_near_top_y_below"`
	DescendLeftYAbove        int   `json:"descend_left_y_above"`
	DescendPollMs            int   `json:"descend_poll_ms"`
	ClimbJumpWaitMs          int   `json:"climb_jump_wait_ms"`
	ClimbAttackJumpWaitMsMin int   `json:"climb_attack_jump_wait_ms_min"`
	ClimbAttackJumpWaitMsMax int   `json:"climb_attack_jump_wait_ms_max"`
	ClimbUpTapMs             int   `json:"climb_up_tap_ms"`
	ClimbLowerRetryMs        int   `json:"climb_lower_retry_ms"` // 抓绳后仍在下层则重爬
	ClimbCenterWaitMs        int   `json:"climb_center_wait_ms"`
	ClimbMaxSec              int   `json:"climb_max_sec"`
	ClimbMaxPass             int   `json:"climb_max_pass"`
	AlignMaxPass             int   `json:"align_max_pass"`
	DescendMaxSec            int   `json:"descend_max_sec"`
	UnknownLayerJumpSec      int   `json:"unknown_layer_jump_sec"`
	DisableAutoPotion        bool  `json:"disable_auto_potion"`
	ShowPosToast             bool  `json:"show_pos_toast"`
	PosToastMs               int   `json:"pos_toast_ms"`
	DeleteYellow             []int `json:"delete_yellow"` // 小地图黄点检测区域 [x1,y1,x2,y2]
}

func (y *Yeqiu001Config) normalize() {
	if y == nil {
		return
	}
	if y.UpperXMin == 0 && y.UpperXMax == 0 {
		y.UpperXMin, y.UpperXMax = 33, 51
	}
	if y.UpperYMin == 0 && y.UpperYMax == 0 {
		y.UpperYMin, y.UpperYMax = 140, 152
	}
	if y.LowerXMin == 0 && y.LowerXMax == 0 {
		y.LowerXMin, y.LowerXMax = -20, 80
	}
	if y.LowerYMin == 0 && y.LowerYMax == 0 {
		y.LowerYMin, y.LowerYMax = 187, 193
	}
	if y.SmallPlatformXCenter == 0 {
		y.SmallPlatformXCenter = 50
	}
	if y.SmallPlatformXTol <= 0 {
		y.SmallPlatformXTol = 20
	}
	if y.SmallPlatformYMin == 0 && y.SmallPlatformYMax == 0 {
		y.SmallPlatformYMin, y.SmallPlatformYMax = 170, 174
	}
	if y.RopeX == 0 {
		y.RopeX = 58
	}
	if y.RopeLeftXMin == 0 && y.RopeLeftXMax == 0 {
		y.RopeLeftXMin, y.RopeLeftXMax = 55, 57
	}
	if y.RopeRightXMin == 0 && y.RopeRightXMax == 0 {
		y.RopeRightXMin, y.RopeRightXMax = 59, 62
	}
	if y.RopeXTolerance <= 0 {
		y.RopeXTolerance = 3
	}
	// RopeJumpXTolerance 默认 0：只在 rope_x 跳抓绳
	if y.RopeJumpXMin == 0 && y.RopeJumpXMax == 0 {
		y.RopeJumpXMin = y.RopeX
		y.RopeJumpXMax = y.RopeX
		if y.RopeJumpXTolerance > 0 {
			y.RopeJumpXMin = y.RopeX - y.RopeJumpXTolerance
			y.RopeJumpXMax = y.RopeX + y.RopeJumpXTolerance
		}
	}
	if y.RopeJumpXMax < y.RopeJumpXMin {
		y.RopeJumpXMax = y.RopeJumpXMin
	}
	if y.RopeNearWalkDist <= 0 {
		y.RopeNearWalkDist = 10
	}
	if y.RopeNearWalkMs <= 0 {
		y.RopeNearWalkMs = 300
	}
	if y.RopeNearStepMs <= 0 {
		y.RopeNearStepMs = 120
	}
	if y.RopeNearMicroStepMs <= 0 {
		y.RopeNearMicroStepMs = 65
	}
	if y.LowerFarmSecMin <= 0 {
		y.LowerFarmSecMin = 40
	}
	if y.LowerFarmSecMax < y.LowerFarmSecMin {
		y.LowerFarmSecMax = y.LowerFarmSecMin + 40
	}
	if y.UpperFarmSecMin <= 0 {
		y.UpperFarmSecMin = 7
	}
	if y.UpperFarmSecMax < y.UpperFarmSecMin {
		y.UpperFarmSecMax = y.UpperFarmSecMin + 5
	}
	if y.PatrolTeleportMargin <= 0 {
		y.PatrolTeleportMargin = 12
	}
	if y.DownJumpCheckMs <= 0 {
		y.DownJumpCheckMs = 200
	}
	if y.DownJumpMaxRetry <= 0 {
		y.DownJumpMaxRetry = 5
	}
	if y.ClimbFallYAbove == 0 {
		y.ClimbFallYAbove = 187
	}
	if y.ClimbNearTopYBelow == 0 {
		y.ClimbNearTopYBelow = 153
	}
	if y.DescendLeftYAbove == 0 {
		y.DescendLeftYAbove = 152
	}
	if y.DescendPollMs <= 0 {
		y.DescendPollMs = 200
	}
	if y.ClimbJumpWaitMs <= 0 {
		y.ClimbJumpWaitMs = 200
	}
	if y.ClimbAttackJumpWaitMsMin <= 0 {
		y.ClimbAttackJumpWaitMsMin = 300
	}
	if y.ClimbAttackJumpWaitMsMax < y.ClimbAttackJumpWaitMsMin {
		y.ClimbAttackJumpWaitMsMax = y.ClimbAttackJumpWaitMsMin + 200
	}
	if y.ClimbUpTapMs <= 0 {
		y.ClimbUpTapMs = 300
	}
	if y.ClimbLowerRetryMs <= 0 {
		y.ClimbLowerRetryMs = 500
	}
	if y.ClimbCenterWaitMs <= 0 {
		y.ClimbCenterWaitMs = 500
	}
	if y.ClimbMaxSec <= 0 {
		y.ClimbMaxSec = 25
	}
	if y.ClimbMaxPass <= 0 {
		y.ClimbMaxPass = 80
	}
	if y.AlignMaxPass <= 0 {
		y.AlignMaxPass = 30
	}
	if y.DescendMaxSec <= 0 {
		y.DescendMaxSec = 8
	}
	if y.UnknownLayerJumpSec <= 0 {
		y.UnknownLayerJumpSec = 3
	}
}

func applyDeleteYellowRegion(rect []int) {
	if len(rect) < 4 {
		return
	}
	x1, y1, x2, y2 := rect[0], rect[1], rect[2], rect[3]
	if x2 <= x1 || y2 <= y1 {
		return
	}
	core.SetMinimapYellowRegion(x1, y1, x2, y2)
}

func isLandMapConfig(cfg *mapConfig) bool {
	if cfg == nil {
		return false
	}
	switch cfg.Type {
	case MapTypeLandInPlaceLR, MapTypeLandHelleWalk, MapTypeLandHelleTeleport:
		return true
	}
	return strings.HasPrefix(cfg.Type, "land") || strings.HasPrefix(cfg.Name, "land")
}

func applyMapMinimapRegions(cfg *mapConfig) {
	if cfg == nil {
		return
	}
	if isLandMapConfig(cfg) {
		core.SetMinimapYellowSim(core.MinimapYellowSimLand)
	} else {
		core.SetMinimapYellowSim(core.MinimapYellowSimDefault)
	}
	applyDeleteYellowRegion(cfg.DeleteYellow)
	if len(cfg.WorldProbe) >= 4 {
		x1, y1, x2, y2 := cfg.WorldProbe[0], cfg.WorldProbe[1], cfg.WorldProbe[2], cfg.WorldProbe[3]
		if x2 > x1 && y2 > y1 {
			imgs := strings.TrimSpace(cfg.WorldProbeImgs)
			if imgs == "" {
				imgs = "img/game/左上角.png"
			}
			SetMinimapWorldProbe(x1, y1, x2, y2, imgs)
		}
	}
}

func clearMapMinimapRegions() {
	core.ClearMinimapYellowRegion()
	ResetMinimapWorldProbe()
}

func (y *Yeqiu001Config) applyDeleteYellowRegion() {
	if y == nil {
		return
	}
	applyDeleteYellowRegion(y.DeleteYellow)
}

func (t *TreasureIslandConfig) applyDeleteYellowRegion() {
	if t == nil {
		return
	}
	applyDeleteYellowRegion(t.DeleteYellow)
}

func (y *Yeqiu001Config) smallPlatformXMin() int {
	return y.SmallPlatformXCenter - y.SmallPlatformXTol
}

func (y *Yeqiu001Config) smallPlatformXMax() int {
	return y.SmallPlatformXCenter + y.SmallPlatformXTol
}

func (y *Yeqiu001Config) ropeJumpXMin() int {
	if y.RopeJumpXMin > 0 || y.RopeJumpXMax > 0 {
		return y.RopeJumpXMin
	}
	return y.RopeX
}

func (y *Yeqiu001Config) ropeJumpXMax() int {
	if y.RopeJumpXMin > 0 || y.RopeJumpXMax > 0 {
		return y.RopeJumpXMax
	}
	return y.RopeX
}

func (y *Yeqiu001Config) ropeLeftTargetX() int {
	return (y.RopeLeftXMin + y.RopeLeftXMax) / 2
}

func (y *Yeqiu001Config) ropeRightTargetX() int {
	return (y.RopeRightXMin + y.RopeRightXMax) / 2
}

func (y *Yeqiu001Config) ropeXMin() int {
	return y.RopeX - y.RopeXTolerance
}

func (y *Yeqiu001Config) ropeXMax() int {
	return y.RopeX + y.RopeXTolerance
}

type TreasureIslandLayerConfig struct {
	Layer                  int   `json:"layer"`
	YMin                   int   `json:"y_min"`
	YMax                   int   `json:"y_max"`
	FightXMin              int   `json:"fight_x_min"`
	FightXMax              int   `json:"fight_x_max"`
	PatrolXInset           int   `json:"patrol_x_inset"`    // 巡逻区相对 fight_x 往中间缩进，0=不缩
	FightAttackMin         int   `json:"fight_attack_min"`  // 需攻击数 >= 此值才攻击；0=用全局怪物数阈值
	MinFarmStaySec         int   `json:"min_farm_stay_sec"` // 本层打怪区最少停留秒数；0=用全局
	MaxStaySec             int   `json:"max_stay_sec"`
	AscendX                int   `json:"ascend_x"` // 从本层向上瞬移+右瞬移的 x 中心（0=不适用）
	AscendXTolerance       int   `json:"ascend_x_tolerance"`
	DescendXThreshold      int   `json:"descend_x_threshold"` // x 超过此值后右走掉层
	DescendWalkMsMin       int   `json:"descend_walk_ms_min"`
	DescendWalkMsMax       int   `json:"descend_walk_ms_max"`
	AscendZoneXTolerance   int   `json:"ascend_zone_x_tolerance"`   // 上升点 x±此值内用下方专用 YOLO/攻击区
	AscendZoneYoloRegion   []int `json:"ascend_zone_yolo_region"`   // [x1,y1,x2,y2]
	AscendZoneAttackRegion []int `json:"ascend_zone_attack_region"` // [x1,y1,x2,y2]
}

// TreasureIslandConfig 抢夺宝物岛专用配置。
type TreasureIslandConfig struct {
	DownJumpYMin              int                         `json:"down_jump_y_min"`
	DownJumpYMax              int                         `json:"down_jump_y_max"`
	DownJumpWaitMs            int                         `json:"down_jump_wait_ms"`
	AscendWaitMsMin           int                         `json:"ascend_wait_ms_min"`
	AscendWaitMsMax           int                         `json:"ascend_wait_ms_max"`
	AscendNearWalkDistMax     int                         `json:"ascend_near_walk_dist_max"`    // 距上升点 x 在此距离内用走路对齐
	AscendNearWalkMsPerDist   int                         `json:"ascend_near_walk_ms_per_dist"` // 走路毫秒 ≈ 此值×距目标格数
	AscendNearWalkMsMin       int                         `json:"ascend_near_walk_ms_min"`
	AscendNearWalkMsMax       int                         `json:"ascend_near_walk_ms_max"`
	FightMonsterMin           int                         `json:"fight_monster_min"`            // 打怪阶段：怪物数 > 此值则攻击
	TransitionFightMonsterMin int                         `json:"transition_fight_monster_min"` // 换层瞬移：怪物数 >= 此值则攻击
	UnknownLayerJumpStreak    int                         `json:"unknown_layer_jump_streak"`    // 连续 N 次 0 层则右跳
	MinFarmStaySec            int                         `json:"min_farm_stay_sec"`            // 每层打怪区最少停留秒数
	FightCenterTolerance      int                         `json:"fight_center_tolerance"`       // 打怪 x 居中容差
	FarmJumpSecMin            int                         `json:"farm_jump_sec_min"`            // 打怪随机跳跃间隔
	FarmJumpSecMax            int                         `json:"farm_jump_sec_max"`
	DeleteYellow              []int                       `json:"delete_yellow"` // 小地图黄点检测区域 [x1,y1,x2,y2]
	Layers                    []TreasureIslandLayerConfig `json:"layers"`
	AscendZoneEnabled         bool                        `json:"-"` // 运行时：L3 超过 max_stay_sec 后才启用上升区 YOLO/攻击
}

func (t *TreasureIslandConfig) normalize() {
	if t == nil {
		return
	}
	if t.DownJumpYMin == 0 && t.DownJumpYMax == 0 {
		t.DownJumpYMin, t.DownJumpYMax = 100, 132
	}
	if t.DownJumpWaitMs <= 0 {
		t.DownJumpWaitMs = 400
	}
	if t.AscendWaitMsMin <= 0 {
		t.AscendWaitMsMin = 300
	}
	if t.AscendWaitMsMax < t.AscendWaitMsMin {
		t.AscendWaitMsMax = t.AscendWaitMsMin + 100
	}
	if t.AscendNearWalkDistMax <= 0 {
		t.AscendNearWalkDistMax = 6
	}
	if t.AscendNearWalkMsPerDist <= 0 {
		t.AscendNearWalkMsPerDist = 45
	}
	if t.AscendNearWalkMsMin <= 0 {
		t.AscendNearWalkMsMin = 60
	}
	if t.AscendNearWalkMsMax < t.AscendNearWalkMsMin {
		t.AscendNearWalkMsMax = t.AscendNearWalkMsMin + 200
	}
	if t.FightMonsterMin <= 0 {
		t.FightMonsterMin = 1
	}
	if t.TransitionFightMonsterMin <= 0 {
		t.TransitionFightMonsterMin = 1
	}
	if t.UnknownLayerJumpStreak <= 0 {
		t.UnknownLayerJumpStreak = 2
	}
	if t.MinFarmStaySec <= 0 {
		t.MinFarmStaySec = 15
	}
	if t.FightCenterTolerance <= 0 {
		t.FightCenterTolerance = 8
	}
	if t.FarmJumpSecMin <= 0 {
		t.FarmJumpSecMin = 20
	}
	if t.FarmJumpSecMax < t.FarmJumpSecMin {
		t.FarmJumpSecMax = t.FarmJumpSecMin + 5
	}
	for i := range t.Layers {
		l := &t.Layers[i]
		if l.AscendXTolerance <= 0 {
			l.AscendXTolerance = 4
		}
		if l.DescendWalkMsMin <= 0 {
			l.DescendWalkMsMin = 500
		}
		if l.DescendWalkMsMax < l.DescendWalkMsMin {
			l.DescendWalkMsMax = l.DescendWalkMsMin + 100
		}
		if (l.Layer == 1 || l.Layer == 3) && l.FightAttackMin <= 0 {
			l.FightAttackMin = 2
		}
	}
}

func (t *TreasureIslandConfig) layerDef(n int) *TreasureIslandLayerConfig {
	if t == nil {
		return nil
	}
	for i := range t.Layers {
		if t.Layers[i].Layer == n {
			return &t.Layers[i]
		}
	}
	return nil
}

func (l *TreasureIslandLayerConfig) inAscendZone(relX int) bool {
	if l == nil || len(l.AscendZoneYoloRegion) < 4 {
		return false
	}
	tol := l.AscendZoneXTolerance
	if tol <= 0 {
		tol = 8
	}
	return matchRange(relX, l.AscendX-tol, l.AscendX+tol)
}

func (l *TreasureIslandLayerConfig) patrolXBounds() (xMin, xMax int) {
	if l == nil {
		return 0, 0
	}
	xMin, xMax = l.FightXMin, l.FightXMax
	if l.PatrolXInset <= 0 {
		return xMin, xMax
	}
	xMin += l.PatrolXInset
	xMax -= l.PatrolXInset
	if xMin >= xMax {
		return l.FightXMin, l.FightXMax
	}
	return xMin, xMax
}

func (t *TreasureIslandConfig) yoloScanRegion(lay, relX int) (x1, y1, x2, y2 int, ok bool) {
	if t == nil || !t.AscendZoneEnabled {
		return 0, 0, 0, 0, false
	}
	def := t.layerDef(lay)
	if def == nil || !def.inAscendZone(relX) {
		return 0, 0, 0, 0, false
	}
	r := def.AscendZoneYoloRegion
	return r[0], r[1], r[2], r[3], true
}

func (t *TreasureIslandConfig) attackRegion(lay, relX int) (x1, y1, x2, y2 int, ok bool) {
	if t == nil || !t.AscendZoneEnabled {
		return 0, 0, 0, 0, false
	}
	def := t.layerDef(lay)
	if def == nil || len(def.AscendZoneAttackRegion) < 4 || !def.inAscendZone(relX) {
		return 0, 0, 0, 0, false
	}
	r := def.AscendZoneAttackRegion
	return r[0], r[1], r[2], r[3], true
}

func (t *TreasureIslandConfig) detectLayer(relY int) int {
	if t == nil {
		return 0
	}
	for i := range t.Layers {
		l := &t.Layers[i]
		if l.Layer <= 0 {
			continue
		}
		if matchRange(relY, l.YMin, l.YMax) {
			return l.Layer
		}
	}
	return 0
}

// StairsFarmConfig 台阶图站位与周期调整参数（研究所 C1 等）。
type StairsFarmConfig struct {
	TargetX                 int `json:"target_x"`
	TargetXTolerance        int `json:"target_x_tolerance"`
	YMin                    int `json:"y_min"`
	YMax                    int `json:"y_max"`
	RecoverXMin             int `json:"recover_x_min"`
	RecoverXMax             int `json:"recover_x_max"`
	AdjustIntervalSecMin    int `json:"adjust_interval_sec_min"`
	AdjustIntervalSecMax    int `json:"adjust_interval_sec_max"`
	AdjustWaitMs            int `json:"adjust_wait_ms"`
	AdjustTargetX           int `json:"adjust_target_x"` // 台阶调整前先对齐到此 x
	AdjustTargetXTolerance  int `json:"adjust_target_x_tolerance"`
	DownHoldMsMin           int `json:"down_hold_ms_min"`
	DownHoldMsMax           int `json:"down_hold_ms_max"`
	UpHoldMsMin             int `json:"up_hold_ms_min"`
	UpHoldMsMax             int `json:"up_hold_ms_max"`
	AfterUpLeftJumpWaitMs   int `json:"after_up_left_jump_wait_ms"`   // 先下后上完成后左跳等待
	AfterAdjustNoDownJumpMs int `json:"after_adjust_no_down_jump_ms"` // 台阶调整后禁止下跳冷却
	AfterUpLeftMs           int `json:"after_up_left_ms"`
	SlowWalkMsMin           int `json:"slow_walk_ms_min"`
	SlowWalkMsMax           int `json:"slow_walk_ms_max"`
	AlignFastWalkMs         int `json:"align_fast_walk_ms"` // 左走后首次往右快走时长
	AlignMaxPass            int `json:"align_max_pass"`
	RecoverMaxPass          int `json:"recover_max_pass"`
	FallTeleportWaitMs      int `json:"fall_teleport_wait_ms"`
	StairsDownJumpWaitMs    int `json:"stairs_down_jump_wait_ms"`
	TeleportXDeltaMin       int `json:"teleport_x_delta_min"`    // |relX-目标| 超过此值用左右瞬移
	RecoverWalkMultiplier   int `json:"recover_walk_multiplier"` // 掉阶（台阶下）慢走距离倍数
	FaceSwitchSecMin        int `json:"face_switch_sec_min"`     // 攻击换向间隔（秒）
	FaceSwitchSecMax        int `json:"face_switch_sec_max"`
	FaceSwitchPauseMs       int `json:"face_switch_pause_ms"` // 换向前暂停攻击（毫秒）
}

func (s *StairsFarmConfig) normalize() {
	if s == nil {
		return
	}
	if s.TargetX == 0 {
		s.TargetX = 48
	}
	if s.TargetXTolerance <= 0 {
		s.TargetXTolerance = 1
	}
	if s.YMin == 0 && s.YMax == 0 {
		s.YMin, s.YMax = 123, 127
	}
	if s.RecoverXMin == 0 && s.RecoverXMax == 0 {
		s.RecoverXMin, s.RecoverXMax = 34, 44
	}
	if s.AdjustIntervalSecMin <= 0 {
		s.AdjustIntervalSecMin = 40
	}
	if s.AdjustIntervalSecMax < s.AdjustIntervalSecMin {
		s.AdjustIntervalSecMax = s.AdjustIntervalSecMin + 10
	}
	if s.AdjustWaitMs <= 0 {
		s.AdjustWaitMs = 1000
	}
	if s.AdjustTargetX == 0 {
		s.AdjustTargetX = 48
	}
	if s.AdjustTargetXTolerance <= 0 {
		s.AdjustTargetXTolerance = 1
	}
	if s.DownHoldMsMin <= 0 {
		s.DownHoldMsMin = 200
	}
	if s.DownHoldMsMax < s.DownHoldMsMin {
		s.DownHoldMsMax = s.DownHoldMsMin + 50
	}
	if s.UpHoldMsMin <= 0 {
		s.UpHoldMsMin = 1000
	}
	if s.UpHoldMsMax < s.UpHoldMsMin {
		s.UpHoldMsMax = s.UpHoldMsMin + 500
	}
	if s.AfterUpLeftJumpWaitMs <= 0 {
		s.AfterUpLeftJumpWaitMs = 300
	}
	if s.AfterAdjustNoDownJumpMs <= 0 {
		s.AfterAdjustNoDownJumpMs = 3000
	}
	if s.AfterUpLeftMs <= 0 {
		s.AfterUpLeftMs = 200
	}
	if s.SlowWalkMsMin <= 0 {
		s.SlowWalkMsMin = 35
	}
	if s.SlowWalkMsMax < s.SlowWalkMsMin {
		s.SlowWalkMsMax = s.SlowWalkMsMin + 25
	}
	if s.AlignFastWalkMs <= 0 {
		s.AlignFastWalkMs = 400
	}
	if s.AlignMaxPass <= 0 {
		s.AlignMaxPass = 30
	}
	if s.RecoverMaxPass <= 0 {
		s.RecoverMaxPass = 25
	}
	if s.FallTeleportWaitMs <= 0 {
		s.FallTeleportWaitMs = 500
	}
	if s.StairsDownJumpWaitMs <= 0 {
		s.StairsDownJumpWaitMs = 400
	}
	if s.TeleportXDeltaMin <= 0 {
		s.TeleportXDeltaMin = 10
	}
	if s.RecoverWalkMultiplier <= 0 {
		s.RecoverWalkMultiplier = 3
	}
	if s.FaceSwitchSecMin <= 0 {
		s.FaceSwitchSecMin = 10
	}
	if s.FaceSwitchSecMax < s.FaceSwitchSecMin {
		s.FaceSwitchSecMax = s.FaceSwitchSecMin + 3
	}
	if s.FaceSwitchPauseMs <= 0 {
		s.FaceSwitchPauseMs = 800
	}
}

type RingCorner struct {
	Layer int `json:"layer"`
	XMin  int `json:"x_min"`
	XMax  int `json:"x_max"`
}

type RingGraphConfig struct {
	Layer1YMin    int        `json:"layer1_y_min"`
	Layer1YMax    int        `json:"layer1_y_max"`
	Layer2YMin    int        `json:"layer2_y_min"`
	Layer2YMax    int        `json:"layer2_y_max"`
	TeleportFarPx int        `json:"teleport_far_px"`
	WalkMs        int        `json:"walk_ms"`
	Layer1RelXMax int        `json:"layer1_rel_x_max"`
	A             RingCorner `json:"A"`
	B             RingCorner `json:"B"`
	C             RingCorner `json:"C"`
	D             RingCorner `json:"D"`
}

type mapConfig struct {
	Name                     string                 `json:"name"`
	Type                     string                 `json:"type"`
	Accounts                 []string               `json:"accounts"` // 可选；非空时仅列内 username 可刷此图
	MonsterLabels            string                 `json:"monster_labels"`
	MonsterAllowlist         []string               `json:"monster_allowlist"`
	YoloRegions              map[string][]int       `json:"yolo_regions"`   // YOLO Detect 截图区域（宜用大区域/全屏，裁剪子图易漏检）
	AttackRegions            map[string][]int       `json:"attack_regions"` // 可选；怪 box 与此区域有交集才攻击，未配置则打扫描到的全部
	Ring                     *RingGraphConfig       `json:"ring"`
	AttackLoopPauseMs        int                    `json:"attack_loop_pause_ms"`
	AttackFaceSettleMs       int                    `json:"attack_face_settle_ms"`
	WalkMs                   int                    `json:"walk_ms"`
	RightJumpEverySec        int                    `json:"right_jump_every_sec"`
	MinMonsterScore          float64                `json:"min_monster_score"`             // 低于此置信度不攻击，默认 0.35
	MinMonstersToAttack      int                    `json:"min_monsters_to_attack"`        // 过滤后数量达到此值才攻击，默认 2（≤1 只不打）
	SkipMonsterScanRelXAbove int                    `json:"skip_monster_scan_rel_x_above"` // 小地图 relX 大于此值不 YOLO 扫怪，0 默认 95
	DisableMonsterScan       bool                   `json:"disable_monster_scan"`          // 直线图：不做 YOLO，瞬移后固定随机攻击
	AfterTeleportAttackMin   int                    `json:"after_teleport_attack_min"`
	AfterTeleportAttackMax   int                    `json:"after_teleport_attack_max"`
	DeleteYellow             []int                  `json:"delete_yellow"`    // 小地图黄点 [x1,y1,x2,y2]
	WorldProbe               []int                  `json:"world_probe"`      // 左上角模板搜索 [x1,y1,x2,y2]
	WorldProbeImgs           string                 `json:"world_probe_imgs"` // 默认 img/game/左上角.png
	Rules                    []PositionRule         `json:"rules"`
	Exceptions               []PositionRule         `json:"exceptions"` // 与 rules 同结构，合并判定（坐标命中动作）
	Stairs                   *StairsFarmConfig      `json:"stairs"`
	TreasureIsland           *TreasureIslandConfig  `json:"treasure_island"`
	Yeqiu001                 *Yeqiu001Config        `json:"yeqiu001"`
	Langligelang001          *Langligelang001Config `json:"langligelang001"`
	Zaozhi001                *Zaozhi001Config       `json:"zaozhi001"`
	YexiongLingdi            *YexiongLingdiConfig   `json:"yexiong_lingdi"`
}

func (r *RingGraphConfig) normalize() {
	if r == nil {
		return
	}
	if r.TeleportFarPx <= 0 {
		r.TeleportFarPx = 15
	}
	if r.WalkMs <= 0 {
		r.WalkMs = 400
	}
}
