package core

import (
	"fmt"
	"strings"
	"time"
)

// GMPatrolAnswerSubmitParams GM 巡逻答题记录提交参数。
type GMPatrolAnswerSubmitParams struct {
	Question     string
	AnswerResult string
	Username     string
	WindowID     string
	DurationMs   int64
}

// AnswerRecordSubmitRequest POST /api/answer-records/submit 请求体。
type AnswerRecordSubmitRequest struct {
	Question     string `json:"question"`
	AnswerResult string `json:"answer_result"`
	WindowID     string `json:"window_id"`
	AnswerTime   string `json:"answer_time"`
	DurationMs   int64  `json:"duration_ms"`
	OSSFileName  string `json:"oss_file_name"`
	Username     string `json:"username"`
}

var gmPatrolAnswerScreenshotUploader func() (string, error)

// RegisterGMPatrolAnswerScreenshotUploader 注册 GM 巡逻答题截图上传（由 main 注入 util 实现，避免循环依赖）。
func RegisterGMPatrolAnswerScreenshotUploader(fn func() (string, error)) {
	gmPatrolAnswerScreenshotUploader = fn
}

// SubmitAnswerRecord 提交答题记录（需登录 JWT）。
func (c *APIClient) SubmitAnswerRecord(req AnswerRecordSubmitRequest) error {
	tokenInterface := Get("token")
	if tokenInterface == nil {
		return fmt.Errorf("未找到token，请先登录")
	}
	token, ok := tokenInterface.(string)
	if !ok || strings.TrimSpace(token) == "" {
		return fmt.Errorf("token格式错误")
	}

	resp, err := c.PostWithToken("/api/answer-records/submit", req, token)
	if err != nil {
		return err
	}
	if resp.Code != 0 && resp.Code != 200 {
		return fmt.Errorf("提交答题记录失败: %s", resp.Message)
	}
	return nil
}

// SubmitGMPatrolAnswerRecord 上传截图并提交 GM 巡逻答题记录。
func SubmitGMPatrolAnswerRecord(params GMPatrolAnswerSubmitParams) {
	ossFileName := ""
	if gmPatrolAnswerScreenshotUploader != nil {
		key, err := gmPatrolAnswerScreenshotUploader()
		if err != nil {
			SLS_Log2NoToast("[GM巡逻] 答题截图上传失败: " + err.Error())
		} else {
			ossFileName = key
			SLS_Log2NoToast("[GM巡逻] 答题截图已上传: " + key)
		}
	} else {
		SLS_Log2NoToast("[GM巡逻] 未注册答题截图上传器，跳过截图")
	}

	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		loc = time.FixedZone("CST", 8*3600)
	}
	answerTime := time.Now().In(loc).Format("2006-01-02 15:04:05")

	req := AnswerRecordSubmitRequest{
		Question:     strings.TrimSpace(params.Question),
		AnswerResult: strings.TrimSpace(params.AnswerResult),
		WindowID:     strings.TrimSpace(params.WindowID),
		AnswerTime:   answerTime,
		DurationMs:   params.DurationMs,
		OSSFileName:  ossFileName,
		Username:     strings.TrimSpace(params.Username),
	}

	if API == nil {
		SLS_Log2NoToast("[GM巡逻] API 未初始化，跳过答题记录提交")
		return
	}
	if err := API.SubmitAnswerRecord(req); err != nil {
		SLS_Log2NoToast("[GM巡逻] 答题记录提交失败: " + err.Error())
		return
	}
	SLS_Log2NoToast(fmt.Sprintf("[GM巡逻] 答题记录已提交 question=%q answer=%q duration=%dms oss=%q",
		req.Question, req.AnswerResult, req.DurationMs, req.OSSFileName))
}
