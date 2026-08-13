package captcha

import (
	"app/core"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

const (
	gmPatrolOCRWaitMin          = 6 * time.Second
	gmPatrolOCRWaitMax          = 10 * time.Second
	gmPatrolOCRPollInterval     = 2 * time.Second
	gmPatrolScrollPhrase1       = "내려주세요"
	gmPatrolScrollPhrase2       = "아래로"
	gmPatrolScrollClickX1       = 945
	gmPatrolScrollClickY1       = 484
	gmPatrolScrollClickX2       = 949
	gmPatrolScrollClickY2       = 489
	gmPatrolScrollCheckX1       = 937
	gmPatrolScrollCheckY1       = 449
	gmPatrolScrollCheckX2       = 959
	gmPatrolScrollCheckY2       = 478
	gmPatrolScrollCheckColor    = "43678a"
	gmPatrolScrollCheckSim      = float32(0.95)
	gmPatrolScrollCheckMinPixel = 5
	gmPatrolScrollHintX1        = 937
	gmPatrolScrollHintY1        = 478
	gmPatrolScrollHintX2        = 957
	gmPatrolScrollHintY2        = 492
	gmPatrolScrollHintColor     = "266295"
	gmPatrolScrollHintSim       = float32(0.95)
	gmPatrolScrollHintMinPixel  = 10
	gmPatrolScrollClickMin      = 18
	gmPatrolScrollClickMax      = 23
	gmPatrolStableOCRMaxAttempt = 5
	gmPatrolStableOCRGap        = 2 * time.Second
)

func (rule popupRule) ocrGMText() string {
	return detectAreaOCR(
		rule.ocrOriginalX1, rule.ocrOriginalY1,
		rule.ocrOriginalX2, rule.ocrOriginalY2,
		rule.ocrRecType, true,
	)
}

func lastFourChars(s string) string {
	rs := []rune(strings.TrimSpace(s))
	if len(rs) == 0 {
		return ""
	}
	if len(rs) <= 4 {
		return string(rs)
	}
	return string(rs[len(rs)-4:])
}

func gmPatrolOCRStable(prev, curr string) bool {
	if strings.TrimSpace(curr) == "" {
		return false
	}
	if prev == "" {
		return false
	}
	return lastFourChars(prev) == lastFourChars(curr)
}

func gmPatrolScrollHintVisible() bool {
	n := core.Color.GetColorCountInRegion(
		gmPatrolScrollHintX1, gmPatrolScrollHintY1,
		gmPatrolScrollHintX2, gmPatrolScrollHintY2,
		gmPatrolScrollHintColor, gmPatrolScrollHintSim,
	)
	return n > gmPatrolScrollHintMinPixel
}

func gmPatrolNeedsScrollDown(ocrText string) bool {
	if strings.Contains(ocrText, gmPatrolScrollPhrase1) || strings.Contains(ocrText, gmPatrolScrollPhrase2) {
		return true
	}
	return gmPatrolScrollHintVisible()
}

func gmPatrolScrollReason(ocrText string) string {
	var reasons []string
	if strings.Contains(ocrText, gmPatrolScrollPhrase1) {
		reasons = append(reasons, gmPatrolScrollPhrase1)
	}
	if strings.Contains(ocrText, gmPatrolScrollPhrase2) {
		reasons = append(reasons, gmPatrolScrollPhrase2)
	}
	if gmPatrolScrollHintVisible() {
		reasons = append(reasons, "color266295")
	}
	if len(reasons) == 0 {
		return "unknown"
	}
	return strings.Join(reasons, "+")
}

// waitGMPatrolOCRStable 最少等 6s、最多 10s，每 2s OCR；末 4 字连续相同视为展示完成。
func waitGMPatrolOCRStable(rule popupRule) string {
	start := time.Now()
	minReady := start.Add(gmPatrolOCRWaitMin)
	deadline := start.Add(gmPatrolOCRWaitMax)

	if wait := time.Until(minReady); wait > 0 {
		time.Sleep(wait)
	}

	var prev string
	for {
		curr := rule.ocrGMText()
		elapsed := int(time.Since(start).Seconds())
		core.SLS_Log2NoToast(fmt.Sprintf("[GM巡逻] OCR等待 %ds last4=%q len=%d",
			elapsed, lastFourChars(curr), len([]rune(curr))))

		if gmPatrolNeedsScrollDown(curr) {
			core.SLS_Log2NoToast("[GM巡逻] 检测到需下拉(" + gmPatrolScrollReason(curr) + ") 执行滚动")
			if text := handleGMPatrolScrollDown(rule); strings.TrimSpace(text) != "" {
				return text
			}
			curr = rule.ocrGMText()
		}

		if gmPatrolOCRStable(prev, curr) {
			core.SLS_Log2NoToast("[GM巡逻] 文字已稳定")
			return curr
		}
		prev = curr

		if time.Now().After(deadline) {
			core.SLS_Log2NoToast("[GM巡逻] 已达最大等待 10s")
			if strings.TrimSpace(curr) != "" {
				return curr
			}
			return prev
		}

		sleep := gmPatrolOCRPollInterval
		if remain := time.Until(deadline); remain < sleep {
			sleep = remain
		}
		if sleep > 0 {
			time.Sleep(sleep)
		}
	}
}

func handleGMPatrolScrollDown(rule popupRule) string {
	maxClicks := gmPatrolScrollClickMin + rand.Intn(gmPatrolScrollClickMax-gmPatrolScrollClickMin+1)
	core.SLS_Log2NoToast(fmt.Sprintf("[GM巡逻] 滚动最多 %d 次", maxClicks))
	for i := 0; i < maxClicks; i++ {
		core.RandomClickInArea(
			gmPatrolScrollClickX1, gmPatrolScrollClickY1,
			gmPatrolScrollClickX2, gmPatrolScrollClickY2,
		)
		core.RandomSleep(50, 100)
		n := core.Color.GetColorCountInRegion(
			gmPatrolScrollCheckX1, gmPatrolScrollCheckY1,
			gmPatrolScrollCheckX2, gmPatrolScrollCheckY2,
			gmPatrolScrollCheckColor, gmPatrolScrollCheckSim,
		)
		core.SLS_Log2NoToast(fmt.Sprintf("[GM巡逻] 滚动点击 #%d 检测区 pixel=%d", i+1, n))
		if n > gmPatrolScrollCheckMinPixel {
			break
		}
	}

	var last string
	for attempt := 0; attempt < gmPatrolStableOCRMaxAttempt; attempt++ {
		ocr1 := rule.ocrGMText()
		time.Sleep(gmPatrolStableOCRGap)
		ocr2 := rule.ocrGMText()
		last = ocr2
		core.SLS_Log2NoToast(fmt.Sprintf("[GM巡逻] 滚动后稳定 #%d last4=%q vs %q",
			attempt+1, lastFourChars(ocr1), lastFourChars(ocr2)))
		if gmPatrolOCRStable(ocr1, ocr2) {
			core.SLS_Log2NoToast("[GM巡逻] 滚动后文字已稳定")
			return ocr2
		}
	}
	return last
}
