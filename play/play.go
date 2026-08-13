package play

import (
	"app/core"
	"errors"
	"fmt"
	"strings"
)

// ErrNotImplemented 未实现的地图类型。
var ErrNotImplemented = errors.New("play: 该地图挂机逻辑尚未实现")

// RunAutoFarm 挂机入口：按地图 type/name 分发到对应 Play_*。
func RunAutoFarm(mapAssetPath string) error {
	cfg, err := loadMapConfig(mapAssetPath)
	if err != nil {
		return err
	}
	if cfg.Type == MapTypeTreasureIsland || cfg.Name == "抢夺宝物岛" || strings.Contains(mapAssetPath, "抢夺宝物岛") {
		return Play_抢夺宝物岛(mapAssetPath)
	}
	if cfg.Type == MapTypeInstituteC1LeftPlatform || cfg.Name == "研究所C1左站台" || strings.Contains(mapAssetPath, "研究所C1左站台") {
		return Play_研究所C1左站台(mapAssetPath)
	}
	if cfg.Type == MapTypeInstituteC2 || cfg.Name == "研究所C2" || strings.Contains(mapAssetPath, "研究所C2") {
		return Play_研究所C2(mapAssetPath)
	}
	if cfg.Type == MapTypeInstituteC1 || cfg.Name == "研究所C1" || strings.Contains(mapAssetPath, "研究所C1") {
		return Play_研究所C1(mapAssetPath)
	}
	skipStartupClearBag := cfg.Type == MapTypeLangligelang001 || cfg.Name == "浪里个浪001" || cfg.Name == "时尚大道CD" ||
		strings.Contains(mapAssetPath, "浪里个浪001") || strings.Contains(mapAssetPath, "时尚大道CD") ||
		cfg.Type == MapTypeZaozhi001 || cfg.Name == "早吱定制001" ||
		strings.Contains(mapAssetPath, "早吱001") ||
		cfg.Type == MapTypeYexiongLingdi || cfg.Name == "野熊的领地" ||
		strings.Contains(mapAssetPath, "野熊的领地") ||
		cfg.Type == MapTypeLandHelleWalk || cfg.Name == "land赫勒地区走路版" ||
		strings.Contains(mapAssetPath, "land赫勒地区走路版") ||
		cfg.Type == MapTypeLandHelleTeleport || cfg.Name == "land赫勒地区瞬移版" ||
		strings.Contains(mapAssetPath, "land赫勒地区瞬移版") ||
		cfg.Type == MapTypeLandInPlaceLR || cfg.Name == "land原地左右打" ||
		strings.Contains(mapAssetPath, "land原地左右打")
	if core.API.GetConfigBoolValue("自动清包") && !skipStartupClearBag {
		RunStartupAutoClearBag(fmt.Sprintf("[%s]", cfg.Name))
	}
	if cfg.Type == MapTypeLandHelleTeleport || cfg.Name == "land赫勒地区瞬移版" || strings.Contains(mapAssetPath, "land赫勒地区瞬移版") {
		return Play_land赫勒地区瞬移版(mapAssetPath)
	}
	if cfg.Type == MapTypeLandHelleWalk || cfg.Name == "land赫勒地区走路版" || strings.Contains(mapAssetPath, "land赫勒地区走路版") {
		return Play_land赫勒地区走路版(mapAssetPath)
	}
	if cfg.Type == MapTypeLandInPlaceLR || cfg.Name == "land原地左右打" || strings.Contains(mapAssetPath, "land原地左右打") {
		return Play_land原地左右打(mapAssetPath)
	}
	if cfg.Type == MapTypeLinear || cfg.Name == "赫勒地区" || strings.Contains(mapAssetPath, "赫勒地区") || strings.Contains(mapAssetPath, "直线图") {
		return Play_直线图(mapAssetPath)
	}
	if cfg.Type == MapTypeInPlaceLR || cfg.Name == "原地左右打" || strings.Contains(mapAssetPath, "原地左右打") {
		return Play_原地左右打(mapAssetPath)
	}
	// 时尚大道CD 实为浪里个浪001（仅改名），须在「时尚大道」路径匹配之前判断。
	if cfg.Type == MapTypeLangligelang001 || cfg.Name == "浪里个浪001" || cfg.Name == "时尚大道CD" ||
		strings.Contains(mapAssetPath, "浪里个浪001") || strings.Contains(mapAssetPath, "时尚大道CD") {
		return Play_浪里个浪001(mapAssetPath)
	}
	if cfg.Type == MapTypeFashionAvenue || cfg.Name == "时尚大道" || strings.Contains(mapAssetPath, "时尚大道") {
		return Play_定制_时尚大道(mapAssetPath)
	}
	if cfg.Type == MapTypeLuffy001 || cfg.Name == "路飞001" || strings.Contains(mapAssetPath, "路飞001") {
		return Play_路飞001(mapAssetPath)
	}
	if cfg.Type == MapTypeYeqiu001 || cfg.Name == "叶秋001" || strings.Contains(mapAssetPath, "叶秋001") {
		return Play_叶秋001(mapAssetPath)
	}
	if cfg.Type == MapTypeZaozhi001 || cfg.Name == "早吱定制001" || strings.Contains(mapAssetPath, "早吱定制001") ||
		strings.Contains(mapAssetPath, "早吱001") {
		return Play_早吱定制001(mapAssetPath)
	}
	if cfg.Type == MapTypeYexiongLingdi || cfg.Name == "野熊的领地" || strings.Contains(mapAssetPath, "野熊的领地") {
		return Play_野熊的领地(mapAssetPath)
	}
	return fmt.Errorf("%w: %s", ErrNotImplemented, cfg.Name)
}

// RunRingGraphFarm 兼容旧入口。
func RunRingGraphFarm(mapAssetPath string) error {
	return RunAutoFarm(mapAssetPath)
}
