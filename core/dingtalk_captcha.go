package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/Dasongzi1366/AutoGo/ime"
	"github.com/Dasongzi1366/AutoGo/storages"
)

const (
	dingTalkCaptchaWebhookDefault = ""

	gmPatrolInputSearchX1      = 479
	gmPatrolInputSearchY1      = 301
	gmPatrolInputSearchX2      = 786
	gmPatrolInputSearchY2      = 528
	gmPatrolInputBoxX1         = 522
	gmPatrolInputBoxW          = 150
	gmPatrolInputBoxH          = 15
	gmPatrolInputBoxColor      = "ffffff"
	gmPatrolInputBoxSim        = float32(0.94)
	gmPatrolInputBoxMinPixels  = 500
	gmPatrolInputSearchStep    = 2
	gmPatrolInputReadyMaxRetry = 5

	gmPatrolSubmitCheckX1        = 514
	gmPatrolSubmitCheckY1        = 503
	gmPatrolSubmitCheckX2        = 626
	gmPatrolSubmitCheckY2        = 515
	gmPatrolSubmitCheckColor     = "434343"
	gmPatrolSubmitCheckSim       = float32(0.95)
	gmPatrolSubmitCheckMinPixels = 2
	gmPatrolPostInputWaitMinMs   = 1000
	gmPatrolPostInputWaitMaxMs   = 1500
	gmPatrolPreConfirmClickX1    = 1185
	gmPatrolPreConfirmClickY1    = 663
	gmPatrolPreConfirmClickX2    = 1206
	gmPatrolPreConfirmClickY2    = 675

	gmPatrolConfirmSearchX1     = 880
	gmPatrolConfirmSearchY1     = 469
	gmPatrolConfirmSearchX2     = 1018
	gmPatrolConfirmSearchY2     = 618
	gmPatrolConfirmBtnColor     = "ee8833"
	gmPatrolConfirmBtnSim       = float32(0.98)
	gmPatrolConfirmBtnW         = 50
	gmPatrolConfirmBtnH         = 9
	gmPatrolConfirmBtnMinPixels = 20 // 50×9 窗口内橙色像素需 >20

	gmPatrolConfirmClickJitter = 2 // 点击按钮中心 x/y ±2

	gmPatrolConfirmMaxRetry = 5

	gmPatrolPostConfirmDelay     = 2 * time.Second
	gmPatrolFollowCheckX1        = 492
	gmPatrolFollowCheckY1        = 176
	gmPatrolFollowCheckX2        = 960
	gmPatrolFollowCheckY2        = 506
	gmPatrolFollowCheckColor     = "f1e5f1"
	gmPatrolFollowCheckSim       = float32(0.94)
	gmPatrolFollowCheckMinPixels = 10000
	gmPatrolFollowClickX1        = 761
	gmPatrolFollowClickY1        = 439
	gmPatrolFollowClickX2        = 791
	gmPatrolFollowClickY2        = 449
	gmPatrolFollowMaxAttempts    = 3
)

const dingTalkCaptchaCooldown = 10 * time.Minute

var (
	dingTalkCaptchaMu       sync.Mutex
	dingTalkCaptchaLastSent time.Time
)

func captchaAccountAndWindow() (username, window string) {
	username = storages.Get("data", "username")
	window = storages.Get("data", "windowId")
	if username == "" {
		if v, ok := Get("username").(string); ok && v != "" {
			username = v
		}
	}
	if username == "" {
		username = "未知账号"
	}
	if window == "" {
		window = "未知"
	}
	return username, window
}

func captchaAccountLabel() string {
	username, window := captchaAccountAndWindow()
	return fmt.Sprintf("%s 窗口%s", username, window)
}

func captchaAtMobiles() []string {
	phone := strings.TrimSpace(API.GetConfigStringValue("configAll.测谎手机号"))
	if phone == "" {
		return nil
	}
	return []string{phone}
}

func captchaDingTalkWebhookURL() string {
	token := strings.TrimSpace(API.GetConfigStringValue("configAll.钉钉.access_token"))
	if token == "" {
		return dingTalkCaptchaWebhookDefault
	}
	if strings.HasPrefix(token, "http://") || strings.HasPrefix(token, "https://") {
		return token
	}
	return "https://oapi.dingtalk.com/robot/send?access_token=" + token
}

// NotifyCaptchaDingTalk 验证码弹出时发钉钉警告；10 分钟内不重复发送。成功发送返回 true。
func NotifyCaptchaDingTalk() bool {
	return sendCaptchaDingTalk("发现测谎："+captchaAccountLabel(), false)
}

// NotifyGMPatrolDingTalkAlert GM巡逻弹出时立即发钉钉（含关键词「测谎」+ 账号+窗口号）。
func NotifyGMPatrolDingTalkAlert() bool {
	username, window := captchaAccountAndWindow()
	alertContent := fmt.Sprintf("测谎\n账号：%s\n窗口号：%s", username, window)
	return sendCaptchaDingTalk(alertContent, false)
}

// GMPatrolBeforeInput 粘贴前钩子（由 captcha 注入 play.ReleaseAllHeldKeys，避免攻击键未松开连打 x）。
var GMPatrolBeforeInput func()

// StartGMPatrolAIAnswer OCR 就绪后后台 AI 答题并输入。
func StartGMPatrolAIAnswer(ocrText string) {
	ocrText = strings.TrimSpace(ocrText)
	if ocrText == "" {
		ocrText = "（未识别）"
	}
	userCopy, windowCopy := captchaAccountAndWindow()
	ocrCopy := ocrText
	go func() {
		reply := "（失败）"
		var elapsed time.Duration
		aiResult, err := AnalyzeGMPatrolKorean(ocrCopy)
		elapsed = aiResult.Elapsed
		modelTag := aiResult.ModelDisplay()
		if aiResult.Provider != "" {
			modelTag = aiResult.Provider + "/" + modelTag
		}
		if err != nil {
			SLS_Log2NoToast(fmt.Sprintf("[GM巡逻] AI 失败 model=%s: %v", modelTag, err))
		} else {
			reply = aiResult.Reply
			SLS_Log2NoToast(fmt.Sprintf("[GM巡逻] AI 作答完成 model=%s reply=%q", modelTag, reply))
		}
		applyGMPatrolReply(reply)
		SubmitGMPatrolAnswerRecord(GMPatrolAnswerSubmitParams{
			Question:     ocrCopy,
			AnswerResult: reply,
			Username:     userCopy,
			WindowID:     windowCopy,
			DurationMs:   elapsed.Milliseconds(),
		})
	}()
}

func applyGMPatrolReply(reply string) {
	reply = strings.TrimSpace(reply)
	if reply == "" || reply == "（失败）" {
		return
	}
	if GMPatrolBeforeInput != nil {
		GMPatrolBeforeInput()
	}
	if !ensureGMPatrolInputReady() {
		SLS_Log2NoToast("[GM巡逻] 输入框未就绪，跳过粘贴")
		return
	}
	if GMPatrolBeforeInput != nil {
		GMPatrolBeforeInput()
	}
	ime.InputText(reply)
	SLS_Log2NoToast("[GM巡逻] 已输入回答: " + reply)
	if !CachedConfigAutoAnswerEnabled() {
		SLS_Log2NoToast("[GM巡逻] auto_answer_enabled=false，跳过自动点确定")
		return
	}
	RandomSleep(gmPatrolPostInputWaitMinMs, gmPatrolPostInputWaitMaxMs)
	cx, cy := RandomClickInArea(
		gmPatrolPreConfirmClickX1, gmPatrolPreConfirmClickY1,
		gmPatrolPreConfirmClickX2, gmPatrolPreConfirmClickY2,
	)
	SLS_Log2NoToast(fmt.Sprintf("[GM巡逻] 输入后预点击 (%d,%d) 区域[%d,%d,%d,%d]",
		cx, cy,
		gmPatrolPreConfirmClickX1, gmPatrolPreConfirmClickY1,
		gmPatrolPreConfirmClickX2, gmPatrolPreConfirmClickY2,
	))
	autoConfirmGMPatrolSubmit()
}

// autoConfirmGMPatrolSubmit 输入后在弹窗内动态找 50×9 橙色确定钮，点击中心±2，最多 5 次；灰条或橙钮消失则停止。
func autoConfirmGMPatrolSubmit() {
	defer gmPatrolPostConfirmFollowUp()
	j := gmPatrolConfirmClickJitter
	for attempt := 0; attempt < gmPatrolConfirmMaxRetry; attempt++ {
		grayN := Color.GetColorCountInRegion(
			gmPatrolSubmitCheckX1, gmPatrolSubmitCheckY1,
			gmPatrolSubmitCheckX2, gmPatrolSubmitCheckY2,
			gmPatrolSubmitCheckColor, gmPatrolSubmitCheckSim,
		)
		btnCX, btnCY, btnOK := findGMPatrolConfirmCenter()
		msg := fmt.Sprintf("[GM巡逻] 确认第%d次 灰条[%d,%d,%d,%d] #%s sim=%.2f pixel=%d 橙钮搜索[%d,%d,%d,%d] win=%dx%d min>%d",
			attempt+1,
			gmPatrolSubmitCheckX1, gmPatrolSubmitCheckY1, gmPatrolSubmitCheckX2, gmPatrolSubmitCheckY2,
			gmPatrolSubmitCheckColor, gmPatrolSubmitCheckSim, grayN,
			gmPatrolConfirmSearchX1, gmPatrolConfirmSearchY1, gmPatrolConfirmSearchX2, gmPatrolConfirmSearchY2,
			gmPatrolConfirmBtnW, gmPatrolConfirmBtnH, gmPatrolConfirmBtnMinPixels,
		)
		if btnOK {
			msg += fmt.Sprintf(" center=(%d,%d)", btnCX, btnCY)
		}
		if grayN <= gmPatrolSubmitCheckMinPixels || !btnOK {
			msg += " → 条件未满足，已确认完成，停止"
			SLS_Log2NoToast(msg)
			fmt.Println(msg)
			return
		}
		clickX, clickY := RandomClickInArea(btnCX-j, btnCY-j, btnCX+j, btnCY+j)
		msg += fmt.Sprintf(" → 点击(%d,%d)", clickX, clickY)
		SLS_Log2NoToast(msg)
		fmt.Println(msg)
		if attempt < gmPatrolConfirmMaxRetry-1 {
			RandomSleep(1000, 2000)
		}
	}
}

// findGMPatrolConfirmCenter 在搜索区内找 50×9 橙色像素最集中的窗口中心（>20 像素）。
func findGMPatrolConfirmCenter() (cx, cy int, ok bool) {
	cx, cy = Color.FindColorWindowPeakCenter(
		gmPatrolConfirmSearchX1, gmPatrolConfirmSearchY1,
		gmPatrolConfirmSearchX2, gmPatrolConfirmSearchY2,
		gmPatrolConfirmBtnColor, gmPatrolConfirmBtnSim,
		gmPatrolConfirmBtnW, gmPatrolConfirmBtnH,
		gmPatrolConfirmBtnMinPixels+1,
	)
	return cx, cy, cx >= 0 && cy >= 0
}

// gmPatrolPostConfirmFollowUp 5 次确定结束后等 2s，检查蓝色弹层并最多跟进点击 3 次。
func gmPatrolPostConfirmFollowUp() {
	time.Sleep(gmPatrolPostConfirmDelay)
	for attempt := 0; attempt < gmPatrolFollowMaxAttempts; attempt++ {
		n := Color.GetColorCountInRegion(
			gmPatrolFollowCheckX1, gmPatrolFollowCheckY1,
			gmPatrolFollowCheckX2, gmPatrolFollowCheckY2,
			gmPatrolFollowCheckColor, gmPatrolFollowCheckSim,
		)
		msg := fmt.Sprintf("[GM巡逻] 提交后跟进第%d次 区域[%d,%d,%d,%d] #%s sim=%.2f pixel=%d 阈值>%d",
			attempt+1,
			gmPatrolFollowCheckX1, gmPatrolFollowCheckY1, gmPatrolFollowCheckX2, gmPatrolFollowCheckY2,
			gmPatrolFollowCheckColor, gmPatrolFollowCheckSim, n, gmPatrolFollowCheckMinPixels,
		)
		if n > gmPatrolFollowCheckMinPixels {
			cx, cy := RandomClickInArea(
				gmPatrolFollowClickX1, gmPatrolFollowClickY1,
				gmPatrolFollowClickX2, gmPatrolFollowClickY2,
			)
			msg += fmt.Sprintf(" → 点击(%d,%d)", cx, cy)
		} else {
			msg += " → 未达阈值，跳过点击"
		}
		SLS_Log2NoToast(msg)
		fmt.Println(msg)
		if attempt < gmPatrolFollowMaxAttempts-1 {
			RandomSleep(800, 1200)
		}
	}
}

func findGMPatrolInputBox() (x1, y1, x2, y2, pixelCount int, ok bool) {
	x1, y1, x2, y2, pixelCount, ok, _ = scanGMPatrolInputBox()
	return x1, y1, x2, y2, pixelCount, ok
}

type gmPatrolInputScanPeak struct {
	y int
	n int
}

func scanGMPatrolInputBox() (x1, y1, x2, y2, pixelCount int, ok bool, peaks []gmPatrolInputScanPeak) {
	sy1, sy2 := gmPatrolInputSearchY1, gmPatrolInputSearchY2
	w, h := gmPatrolInputBoxW, gmPatrolInputBoxH
	x1 = gmPatrolInputBoxX1
	if sy2-sy1+1 < h {
		return 0, 0, 0, 0, 0, false, nil
	}

	bestN := 0
	bestY1 := 0
	step := gmPatrolInputSearchStep
	if step < 1 {
		step = 1
	}
	for y := sy1; y <= sy2-h+1; y += step {
		n := Color.GetColorCountInRegion(
			x1, y, x1+w-1, y+h-1,
			gmPatrolInputBoxColor, gmPatrolInputBoxSim,
		)
		if n > 0 {
			peaks = appendInputScanPeak(peaks, y, n, 5)
		}
		if n > gmPatrolInputBoxMinPixels && n > bestN {
			bestN = n
			bestY1 = y
		}
	}
	if bestN <= gmPatrolInputBoxMinPixels {
		return 0, 0, 0, 0, bestN, false, peaks
	}
	return x1, bestY1, x1 + w - 1, bestY1 + h - 1, bestN, true, peaks
}

func appendInputScanPeak(peaks []gmPatrolInputScanPeak, y, n, max int) []gmPatrolInputScanPeak {
	peaks = append(peaks, gmPatrolInputScanPeak{y: y, n: n})
	for i := len(peaks) - 1; i > 0; i-- {
		if peaks[i].n > peaks[i-1].n {
			peaks[i], peaks[i-1] = peaks[i-1], peaks[i]
		} else {
			break
		}
	}
	if len(peaks) > max {
		peaks = peaks[:max]
	}
	return peaks
}

func logGMPatrolInputBoxDebug(context string, peaks []gmPatrolInputScanPeak, bestN int) {
	var b strings.Builder
	fmt.Fprintf(&b, "[GM巡逻][debug] %s 未找到输入框白条 搜索区[%d,%d,%d,%d] 扫描框x=%d w=%d h=%d #%s sim=%.2f 阈值>%d",
		context,
		gmPatrolInputSearchX1, gmPatrolInputSearchY1, gmPatrolInputSearchX2, gmPatrolInputSearchY2,
		gmPatrolInputBoxX1, gmPatrolInputBoxW, gmPatrolInputBoxH,
		gmPatrolInputBoxColor, gmPatrolInputBoxSim, gmPatrolInputBoxMinPixels,
	)
	if bestN > 0 {
		fmt.Fprintf(&b, " best=%d", bestN)
	} else {
		b.WriteString(" best=0")
	}
	if len(peaks) == 0 {
		b.WriteString(" top=无白色像素命中")
	} else {
		b.WriteString(" top=")
		for i, p := range peaks {
			if i > 0 {
				b.WriteByte(',')
			}
			fmt.Fprintf(&b, "y%d:n%d", p.y, p.n)
		}
	}
	msg := b.String()
	SLS_Log2NoToast(msg)
	fmt.Println(msg)
}

func clickGMPatrolInputBox() bool {
	x1, y1, x2, y2, n, found, peaks := scanGMPatrolInputBox()
	if !found {
		logGMPatrolInputBoxDebug("点击前", peaks, n)
		SLS_Log2NoToast("[GM巡逻] 未找到输入框白条")
		return false
	}
	cy := (y1 + y2) / 2
	cx := gmPatrolInputBoxX1
	Click(cx, cy)
	SLS_Log2NoToast(fmt.Sprintf("[GM巡逻] 点击输入框 (%d,%d) 区域[%d,%d,%d,%d] pixel=%d",
		cx, cy, x1, y1, x2, y2, n))
	return true
}

// ensureGMPatrolInputReady 动态定位输入框白条并点击，未找到则最多再试 5 次。
func ensureGMPatrolInputReady() bool {
	for attempt := 0; attempt <= gmPatrolInputReadyMaxRetry; attempt++ {
		if attempt > 0 {
			SLS_Log2NoToast(fmt.Sprintf("[GM巡逻] 输入框未找到 第%d次重试点击", attempt))
		}
		if !clickGMPatrolInputBox() {
			if attempt < gmPatrolInputReadyMaxRetry {
				RandomSleep(1700, 2100)
			}
			continue
		}
		if attempt == 0 {
			RandomSleep(2000, 3000)
		} else {
			RandomSleep(1700, 2100)
		}
		if _, _, _, _, n, ok := findGMPatrolInputBox(); ok {
			SLS_Log2NoToast(fmt.Sprintf("[GM巡逻] 输入框就绪 pixel=%d", n))
			return true
		}
		_, _, _, _, n, ok, peaks := scanGMPatrolInputBox()
		if !ok {
			logGMPatrolInputBoxDebug(fmt.Sprintf("点击后第%d次复检", attempt), peaks, n)
		}
	}
	_, _, _, _, n, ok, peaks := scanGMPatrolInputBox()
	if !ok {
		logGMPatrolInputBoxDebug("全部重试结束", peaks, n)
	}
	SLS_Log2NoToast("[GM巡逻] 输入框仍未找到")
	return false
}

// NotifyCaptchaDingTalkTest 测试发送，忽略 10 分钟冷却。
func NotifyCaptchaDingTalkTest() bool {
	return sendCaptchaDingTalk("发现测谎："+captchaAccountLabel(), true)
}

func sendCaptchaDingTalk(content string, skipCooldown bool) bool {
	dingTalkCaptchaMu.Lock()
	defer dingTalkCaptchaMu.Unlock()

	if !skipCooldown && !dingTalkCaptchaLastSent.IsZero() && time.Since(dingTalkCaptchaLastSent) < dingTalkCaptchaCooldown {
		SLS_Log2NoToast("[钉钉] 10分钟内已发送，跳过")
		return false
	}

	if strings.TrimSpace(content) == "" {
		content = "发现测谎：" + captchaAccountLabel()
	}
	msg := map[string]interface{}{
		"msgtype": "text",
		"text":    map[string]string{"content": content},
	}
	if atMobiles := captchaAtMobiles(); len(atMobiles) > 0 {
		msg["at"] = map[string]interface{}{
			"atMobiles": atMobiles,
			"isAtAll":   false,
		}
	}
	payload, err := json.Marshal(msg)
	if err != nil {
		SLS_Log2NoToast("[钉钉] 序列化失败: " + err.Error())
		return false
	}

	req, err := http.NewRequest(http.MethodPost, captchaDingTalkWebhookURL(), bytes.NewReader(payload))
	if err != nil {
		SLS_Log2NoToast("[钉钉] 构建请求失败: " + err.Error())
		return false
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := (&http.Client{Timeout: 10 * time.Second}).Do(req)
	if err != nil {
		SLS_Log2NoToast("[钉钉] 发送失败: " + err.Error())
		return false
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		SLS_Log2NoToast(fmt.Sprintf("[钉钉] http=%d body=%s", resp.StatusCode, string(body)))
		return false
	}

	var dingResp struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}
	if err := json.Unmarshal(body, &dingResp); err != nil {
		SLS_Log2NoToast("[钉钉] 响应解析失败: " + err.Error() + " body=" + string(body))
		return false
	}
	if dingResp.ErrCode != 0 {
		SLS_Log2NoToast(fmt.Sprintf("[钉钉] 发送被拒 errcode=%d errmsg=%s content=%q",
			dingResp.ErrCode, dingResp.ErrMsg, content))
		return false
	}

	dingTalkCaptchaLastSent = time.Now()
	SLS_Log2NoToast("[钉钉] 警告已发送: " + strings.ReplaceAll(content, "\n", " "))
	return true
}
