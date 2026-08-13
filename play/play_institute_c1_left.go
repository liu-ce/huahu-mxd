package play

import (
	"app/core"
	"fmt"
	"time"
)

func normalizeInstituteC1LeftPlatform(st *StairsFarmConfig) {
	if st == nil {
		return
	}
	if st.RecoverXMin == 0 && st.RecoverXMax == 0 {
		st.RecoverXMin, st.RecoverXMax = -2, 9
	}
	if st.TargetXTolerance <= 0 {
		st.TargetXTolerance = 3
	}
	if st.FaceSwitchSecMin <= 0 {
		st.FaceSwitchSecMin = 10
	}
	if st.FaceSwitchSecMax <= 0 {
		st.FaceSwitchSecMax = 20
	}
	st.normalize()
}

func handleLeftPlatformPosition(tag string, st *StairsFarmConfig, relX, relY int) bool {
	if relY > st.YMax {
		recoverFromFall(tag, st, relX, relY)
		return true
	}
	if relY < st.YMin {
		recoverFromStairs(tag, st, relX, relY)
		return true
	}
	return nudgeXIfNeeded(tag, st, relX)
}

func onLeftPlatformFightSpot(st *StairsFarmConfig, relX, relY int) bool {
	return relY >= st.YMin && relY <= st.YMax && stairsXInTarget(st, relX)
}

// Play_研究所C1左站台 x=target±3、y=[y_min,y_max] 原地攻击；掉阶后 x∈[recover] 上瞬移回位。
func Play_研究所C1左站台(mapAssetPath string) error {
	cfg, err := loadMapConfig(mapAssetPath)
	if err != nil {
		return err
	}
	if cfg.Stairs == nil {
		return fmt.Errorf("%s: 缺少 stairs 配置", cfg.Name)
	}
	st := cfg.Stairs
	normalizeInstituteC1LeftPlatform(st)

	logTag := instituteC1LogTag(cfg.Name)
	SetFarmLogTag(logTag)

	StartFarmMaintainLoop(logTag)
	defer StopFarmMaintainLoop()

	facingRight := true
	nextFaceSwitch := time.Now().Add(faceSwitchInterval(st))
	faceRight()
	instituteC1Log(logTag, "开始挂机 x=%d±%d y=[%d,%d] 掉阶recover=[%d,%d] 换向%d～%ds",
		st.TargetX, st.TargetXTolerance, st.YMin, st.YMax,
		st.RecoverXMin, st.RecoverXMax, st.FaceSwitchSecMin, st.FaceSwitchSecMax)

	for {
		core.BlockWhileCaptchaHold()
		TickFarmMainThreadTasks()

		if relX, relY, ok := readMinimapRel(); ok {
			if handleLeftPlatformPosition(logTag, st, relX, relY) {
				continue
			}
			if !onLeftPlatformFightSpot(st, relX, relY) {
				continue
			}
		}

		maybeSwitchAttackFace(logTag, st, &facingRight, &nextFaceSwitch)
		keyHoldPress(attackKeyCode(), keyHoldShortMin, keyHoldShortMax)
	}
}
