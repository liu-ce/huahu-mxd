package job

import (
	"app/core"
	"app/util"
	"fmt"
	"strconv"
)

// DO_读取金币：先在道具区找 item 并点一下等 2 秒，再按枫币锚点 OCR；最后尽量关掉道具框（close1）。
func DO_读取金币() {
	ix, iy := core.OpenCV.FindImage(476, 633, 771, 711, "img/game/item.png,img/game/item2.png", false, 1.0, 0.8)
	fmt.Printf("[job][金币] item锚点 x=%d y=%d\n", ix, iy)
	if ix > 0 && iy > 0 {
		core.RandomClickInArea(ix, iy, ix+5, iy+5)
		core.RandomSleep(3000, 4000)
	}

	x, y := core.OpenCV.FindImage(0, 0, 813, 716, "img/game/枫币.png,img/game/枫币2.png", false, 1.0, 0.7)
	fmt.Printf("[job][金币] 枫币锚点 x=%d y=%d\n", x, y)
	if x > 0 && y > 0 {
		num, ok := core.OCR.DetectNumber(x+61, y-2, x+174, y+19)
		fmt.Printf("[job][金币] num=%d ok=%v\n", num, ok)
		if ok {
			core.Storages.DataPut("silver", strconv.Itoa(num))
		}
	}

	CloseTopBarCloseButtons("[job][金币]")
}

func DO_使用回城卷轴() bool {
	ix, iy := core.OpenCV.FindImage(537, 637, 827, 715, "img/game/item.png,img/game/item2.png", false, 1.0, 0.8)
	fmt.Printf("[job][金币] item锚点 x=%d y=%d\n", ix, iy)
	if ix > 0 && iy > 0 {
		core.RandomClickInArea(ix, iy, ix+5, iy+5)
		core.RandomSleep(3000, 4000)
	} else {
		return false
	}

	ix, iy = core.OpenCV.FindImage(0, 0, 1099, 601, "img/game/消耗1.png,img/game/消耗2.png,img/game/消耗3.png,img/game/消耗4.png", false, 1.0, 0.7)
	fmt.Printf("[job][消耗] item锚点 x=%d y=%d\n", ix, iy)
	if ix > 0 && iy > 0 {
		core.SLS_Log2("点击消耗品")
		core.RandomClickInArea(ix, iy, ix+5, iy+5)
		core.RandomSleep(3000, 4000)
	} else {
		core.SLS_Log2("未找到消耗菜单")
		return false
	}

	// 双击回城卷轴
	ix, iy = core.OpenCV.FindImage(0, 0, 1099, 601, "img/game/回城卷轴.png,img/game/回城卷轴2.png,img/game/回城卷轴3.png,img/game/回城卷轴4.png", false, 1.0, 0.7)
	fmt.Printf("[job][卷轴] item锚点 x=%d y=%d\n", ix, iy)
	if ix > 0 && iy > 0 {
		core.SLS_Log2("点击回城卷轴")
		for i := 0; i < 5; i++ {
			core.RandomClickInArea(ix+10, iy+10, ix+13, iy+13)
			core.RandomSleep(100, 120)
		}
		core.RandomSleep(2000, 3000)
	} else {
		core.SLS_Log2("未找到回城卷轴")
		return false
	}

	return true

}

func DO_关闭所有框() {
	for i := 0; i < 5; i++ {
		text := core.OCR.DetectText(242, 391, 242, 391)
		if text == "取消" {
			core.RandomClickInArea(813, 401, 836, 413)
			core.RandomSleep(1200, 1500)
		}

		text = core.OCR.DetectText(756, 398, 806, 422)
		if util.CalculateTextSimilarity(text, "取消") >= 0.4 {
			core.RandomClickInArea(768, 402, 791, 411)
		}

		cx, cy := core.OpenCV.FindImage(closeTopBarX1, closeTopBarY1, closeTopBarX2, closeTopBarY2, closeTopBarTemplates, false, 1.0, closeTopBarSim)
		if cx <= 0 || cy <= 0 {
			break
		}
		fmt.Printf("[job][金币] 关闭道具框 close1 x=%d y=%d i=%d\n", cx, cy, i)
		core.RandomClickInArea(cx, cy, cx+2, cy+2)
		core.Sleep(closeTopBarBetweenMs)
	}

}

func DO_点击NPC() {
	core.RandomClickInArea(931, 567, 956, 595)
	core.RandomSleep(3000, 4000)
}

func DO_等待界面加载出来() bool {
	for i := 0; i < 100; i++ {
		ix, iy := core.OpenCV.FindImage(476, 633, 771, 711, "img/game/item.png,img/game/item2.png", false, 1.0, 0.7)
		if ix > 0 && iy > 0 {
			core.SLS_Log2("界面已经加载")
			core.RandomSleep(3000, 4000)
			return true
		}
		core.RandomSleep(1000, 1200)
	}

	return false
}
