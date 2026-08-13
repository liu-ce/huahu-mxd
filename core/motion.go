package core

import (
	"fmt"
	"math"
	"math/rand"
	"strings"

	"github.com/Dasongzi1366/AutoGo/motion"
	"github.com/Dasongzi1366/AutoGo/utils"
)

func Log(format string, a ...any) {
	// 格式化消息
	var message string
	if len(a) > 0 {
		// 检查格式字符串是否包含占位符
		if strings.Contains(format, "%") {
			message = fmt.Sprintf(format, a...)
		} else {
			// 如果没有占位符，直接拼接所有参数
			args := make([]interface{}, 0, len(a)+1)
			args = append(args, format)
			args = append(args, a...)
			message = fmt.Sprint(args...)
		}
	} else {
		message = format
	}

	// 暂时注释掉SLS推送，避免编译错误
	// SLS_Log(message)

	// 打印到控制台
	fmt.Println(message)

}

func Click(x, y int) {
	motion.Click(x, y, 1, 0)
}

// 在指定区域内随机点击
func RandomClickInArea(x1, y1, x2, y2 int) (int, int) {
	// 确保坐标顺序正确
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	if y1 > y2 {
		y1, y2 = y2, y1
	}

	// 在区域内生成随机坐标
	randomX := utils.Random(x1, x2)
	randomY := utils.Random(y1, y2)
	// 点击随机位置
	Click(randomX, randomY)
	return randomX, randomY
}

// 在指定区域内随机长按（TouchDown → 等待 → 多次 TouchUp，避免 LongClick 偶发不松开）
func RandomLongClickInArea(x1, y1, x2, y2, minDuration, maxDuration int) (int, int) {
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	if y1 > y2 {
		y1, y2 = y2, y1
	}
	randomX := utils.Random(x1, x2)
	randomY := utils.Random(y1, y2)
	randomDuration := utils.Random(minDuration, maxDuration)
	holdTouchAt(randomX, randomY, randomDuration)
	return randomX, randomY
}

// holdTouchAt 按下 → 保持 durationMs → 松开（defer 再补一次松开）。
func holdTouchAt(x, y, durationMs int) {
	motion.TouchDown(x, y, 0, 0)
	defer releaseTouchAt(x, y)
	if durationMs > 0 {
		Sleep(durationMs)
	}
	releaseTouchAt(x, y)
}

// releaseTouchAt 连发 TouchUp，降低模拟器偶发「按下不松开」概率。
func releaseTouchAt(x, y int) {
	motion.TouchUp(x, y, 0, 0)
	Sleep(50)
	motion.TouchUp(x, y, 0, 0)
	Sleep(50)
	motion.TouchUp(x, y, 0, 0)
}

func Swipe(x1, y1, x2, y2, duration int) {
	motion.Swipe(x1, y1, x2, y2, duration, 0, 0)
}

// Swipe 慢慢从 (x1, y1) 滑动到 (x2, y2)，duration 是总时长（毫秒）

func Swipe2(x1, y1, x2, y2 int, durationMs int, time2 int) {
	steps := 20
	if steps < 1 {
		steps = 1
	}
	if durationMs < steps {
		durationMs = steps
	} // 保证每步至少1ms

	delayMs := durationMs / steps // int，给你的 Sleep 用

	motion.TouchDown(x1, y1, 0, 0)
	for i := 1; i <= steps; i++ {
		t := float64(i) / float64(steps)
		x := int(math.Round(float64(x1) + t*float64(x2-x1)))
		y := int(math.Round(float64(y1) + t*float64(y2-y1)))
		motion.TouchMove(x, y, 0, 0)
		Sleep(delayMs) // ← 这里就是你的 Sleep(int)
	}

	Sleep(time2)
	motion.TouchUp(x2, y2, 0, 0)
	Sleep(50)
	motion.TouchUp(x2, y2, 0, 0)
	Sleep(50)
	motion.TouchUp(x2, y2, 0, 0)
}

func Sleep(ms int) {
	if ms <= 0 {
		return
	}
	const chunk = 50
	rem := ms
	for rem > 0 {
		step := chunk
		if step > rem {
			step = rem
		}
		utils.Sleep(step)
		rem -= step
	}
}

// KeyTap 模拟按一次方向键等（code 见 motion 包 KEYCODE_* 常量），displayId 一般为 0。
func KeyTap(code int) {
	motion.KeyAction(code, 0)
}

// 毫秒
func RandomSleep(min, max int) {
	randNum := utils.Random(min, max)
	Sleep(randNum)
}

// RandomSleepAround 以 centerMs 为中心，在 ±jitterRatio 比例内随机睡眠（如 0.3 → 约 70%～130%）。
func RandomSleepAround(centerMs int, jitterRatio float64) {
	if centerMs <= 0 {
		return
	}
	if jitterRatio < 0 {
		jitterRatio = 0
	}
	low := int(math.Round(float64(centerMs) * (1 - jitterRatio)))
	high := int(math.Round(float64(centerMs) * (1 + jitterRatio)))
	if low < 1 {
		low = 1
	}
	if high <= low {
		high = low + 1
	}
	RandomSleep(low, high)
}

// JitterMs 返回约 center*(1±ratio) 毫秒的随机整数，用于间隔、deadline 等（至少 1，区间宽度至少 2ms）。
func JitterMs(center int, jitterRatio float64) int {
	if center <= 0 {
		return 1
	}
	if jitterRatio < 0 {
		jitterRatio = 0
	}
	low := int(math.Round(float64(center) * (1 - jitterRatio)))
	high := int(math.Round(float64(center) * (1 + jitterRatio)))
	if low < 1 {
		low = 1
	}
	if high <= low {
		high = low + 1
	}
	return low + rand.Intn(high-low+1)
}
