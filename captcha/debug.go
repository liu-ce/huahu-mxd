package captcha

import (
	"app/assets"
	"app/core"
	"app/play"
	"app/util"
	"fmt"
	"os"
	"runtime/debug"
	"time"

	"github.com/Dasongzi1366/AutoGo/storages"
	"github.com/Dasongzi1366/AutoGo/yolo"
)

// ReportPanicToSLS 作为 main 的 defer 目标，在 main 协程 panic 时收尾。
func ReportPanicToSLS() {
	r := recover()
	if r == nil {
		return
	}
	stack := debug.Stack()
	msg := fmt.Sprintf("[main panic] %v\n--- stack ---\n%s", r, stack)
	fmt.Println(os.Stderr, "%s\n", msg)
	core.SLS(msg)
	os.Exit(2)
}

func TestAttackKey100() {
	const total = 100
	fmt.Printf("[test] 3 秒后开始按住空格 %d 次，每次按住 200-300ms（切到游戏画面）\n", total)
	core.Sleep(3000)
	for i := 0; i < total; i++ {
		fmt.Println(i)
		play.PressAttackHold()
	}
	fmt.Printf("[test][攻击] 完成 %d/%d\n", total, total)
}

func TestTeleportRight() {
	fmt.Println("[test] 3 秒后开始右瞬移，共 5 次，间隔 2s（切到游戏画面）")
	for i := 1; i <= 10; i++ {
		fmt.Printf("[test] 右瞬移 %d/5\n", i)
		play.TestTeleportRight()
		core.Sleep(1000)
	}
	fmt.Println("[test] 右瞬移测试结束")
}

func TestTeleportLeftLoop() {
	fmt.Println("[test] 3 秒后开始左瞬移，每 1s 一次（切到游戏画面，关脚本结束）")
	core.Sleep(3000)
	for i := 1; ; i++ {
		fmt.Printf("[test] 左瞬移 #%d\n", i)
		play.TestTeleportLeft()
		core.Sleep(1000)
	}
}

func TestTeleportLeftAttackLoop() {
	fmt.Println("[test] 3 秒后开始左瞬移+攻击，每 1s 一次（切到游戏画面，关脚本结束）")
	core.Sleep(3000)
	for i := 1; ; i++ {
		fmt.Printf("[test] 左瞬移+攻击 #%d\n", i)
		play.TestTeleportLeftAndAttack()
		core.Sleep(1000)
	}
}

func TestJumpLoop() {
	fmt.Println("[test] 3 秒后开始跳跃，每 1s 一次（切到游戏画面，关脚本结束）")
	core.Sleep(3000)
	for i := 1; ; i++ {
		fmt.Printf("[test] 跳跃 #%d\n", i)
		play.TestJump()
		core.Sleep(1000)
	}
}

func RunMinimapYellowPointLivePrint() {
	for {
		_, _, ok, detail := play.ReadMinimapRelWithDetail()
		if ok {
			fmt.Printf("[yellow] %s\n", detail)
		} else {
			fmt.Printf("[yellow] 小地图未识别: %s\n", detail)
		}
		core.Sleep(250)
	}
}

func YoloRegionMonsterTest() {
	const (
		cropX1, cropY1, cropX2, cropY2 = 336, 223, 876, 685
		scanX1, scanY1, scanX2, scanY2 = 0, 250, 1270, 660
	)
	regions := []play.YoloDebugRegion{
		{Label: "crop_detect", X1: cropX1, Y1: cropY1, X2: cropX2, Y2: cropY2},
		{Label: "scan_detect", X1: scanX1, Y1: scanY1, X2: scanX2, Y2: scanY2},
	}
	scanAttack := &play.YoloScanAttackPair{
		ScanX1: scanX1, ScanY1: scanY1, ScanX2: scanX2, ScanY2: scanY2,
		AttackX1: cropX1, AttackY1: cropY1, AttackX2: cropX2, AttackY2: cropY2,
	}
	start := time.Now()
	if err := play.RunYoloMultiRegionDebugLoop("config/farm_map_时尚大道.json", 500, regions, scanAttack); err != nil {
		fmt.Printf("yolo 区域测试结束: %v (运行 %dms)\n", err, time.Since(start).Milliseconds())
	}
}

func YoloTest() {
	paramPath, binPath, _ := assets.InstallYoloOnDevice()
	detector := yolo.New("v8", 4, paramPath, binPath, "绿蘑菇,混种冰石巨人,月秒,小虎,珍珠奶茶,洛伊德,围巾蜥蜴,大恶魔,地鼠,情报收集机,要塞巨人,螺母,海盗,烧杯怪,哈门库鲁,稻草人")

	for i := 0; i < 10000; i++ {
		results := detector.Detect(0, 0, 0, 0, 0)
		fmt.Println(results)
		for _, result := range results {
			fmt.Printf("检测到 %s，置信度: %.2f\n", result.Label, result.Score)
		}
		core.Sleep(500)
	}
}

func TestDingTalkCaptchaSend() {
	if storages.Get("data", "username") == "" {
		storages.Put("data", "username", "测试账号")
	}
	if storages.Get("data", "windowId") == "" {
		storages.Put("data", "windowId", "测试窗口")
	}
	fmt.Println("[test] 发送钉钉测谎测试消息...")
	if core.NotifyCaptchaDingTalkTest() {
		util.StartCaptchaScreenshotBurst()
	}
	fmt.Println("[test] 已调用，请查看日志与钉钉群")
}
