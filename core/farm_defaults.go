package core

// DefaultFarmMonsterLabels 全地图共用的 YOLO 怪物标签列表；地图 JSON 不写 monster_labels 时使用。
const DefaultFarmMonsterLabels = "绿蘑菇,混种冰石巨人,月秒,小虎,珍珠奶茶,洛伊德,围巾蜥蜴,大恶魔,地鼠,情报收集机,要塞巨人,螺母,海盗,烧杯怪,哈门库鲁,稻草人"

// 以下为挂机相关矩形，格式均为 {x1,y1,x2,y2}（左上→右下），点击时在框内 RandomClickInArea。

// 跳跃按键：{1166,543,1197,571}
var FarmJumpKeyRect = [4]int{1167, 541, 1193, 569}

// 瞬移辅助：{1053,623,1069,639}
var FarmTeleportAssistRect = [4]int{1049, 625, 1073, 645}

// 自动技能 A：{1116,389,1146,418}
// 自动技能 B：{1004,386,1037,411}
// 自动技能 C：{924,469,947,491}
// 自动技能 D：{932,567,954,592}
var FarmAutoSkillRects = map[string][4]int{
	"A": {1122, 388, 1147, 409},
	"B": {1007, 389, 1029, 409},
	"C": {928, 469, 939, 483},
	"D": {937, 585, 953, 591},
}

// 跳跃间隔毫秒：min=70 max=130（RandomSleep 用）
const (
	FarmJumpGapMinMs = 70
	FarmJumpGapMaxMs = 130
)

// FarmGoRandomClickRect 对矩形 {x1,y1,x2,y2} 随机点一次。
// 挂机高频路径下以前用 goroutine 会与截图/YOLO 抢 JNI，易堆积协程与触发 native 压力；改为同步点击并在点后短睡让出。
func FarmGoRandomClickRect(r [4]int) {
	RandomClickInArea(r[0], r[1], r[2], r[3])
	Sleep(JitterMs(12, 0.35))
}
