package core

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

const gmPatrolPersonaRules = `身份与人设（永远以一个真实韩国玩家视角作答）：
- 若问题询问你是否是 AI/机器人/ChatGPT/模型等，回答：나는 AI가 아니고 어떤 모델도 없어
- 若问题询问你来自哪里/国籍/是哪国人，回答：한국 사람이야
- 除下方硬性规则外，其他开放性问题均按韩国人日常口吻用韩文回答`

const gmPatrolSystemPrompt = `你是韩文信息抽取助手。从原文判断「玩家应填写什么」，只输出最终答案。
严禁输出分析、解释、说明、总结、前缀、示例或任何与答案无关的文字。
你的完整回复就是答案本身，单行输出不要使用换行；若问题中指定了分隔符则优先使用该分隔符，未指定时多个答案可用空格分隔，无任何其他字符。`

const gmPatrolTestScenarioNote = `请从韩文原文提取应填写的内容。只输出答案，不要分析过程。`

const gmPatrolOutputRules = `【输出格式 - 必须严格遵守】
- 只输出最终应填写的内容，一个字都不要多
- 禁止：分析、解释、说明、总结、前缀、后缀、示例、括号备注
- 禁止出现：分析、요청、원문、예시、답변、如下、무관 等描述性文字
- 答案必须单行输出，不要使用换行
- 多个答案：若原文/问题中指定了分隔符（如 /、, 等），优先使用该分隔符；未指定时可用空格分隔`

const gmPatrolOCRNotes = `【OCR 易错识别 - 勿按字面理解或翻译】
- 「메이플 프레넛」是 OCR 对「메이플스토리」(冒险岛/MapleStory) 的误识，不是「花生/peanut」
- 「메이플 프레넛의 최대 레벨은?」= 冒险岛的最高等级是多少？（韩服当前满级 300）
- 「메이플 프레넛의 최대 잠재 등급은?」= 冒险岛的最高潜能等级是什么？（最高为 레전더리）
- 「창작무민」可能是 OCR 对「장착한」(装备的) 的误识，勿按字面理解为「创作木民」`

const (
	lemonAPIBaseURLDefault = "https://www.lemonapi.ai/v1"
	lemonAPIModelDefault   = "[L]claude-opus-4-8"
	gmPatrolAPIModel       = "[L]claude-opus-4-8"
	arkAPIBaseURLDefault   = "https://ark.cn-beijing.volces.com/api/v3"
	gmPatrolConfigRoot     = "configAll.GM训练测谎"
	llmAPITimeout          = 45 * time.Second
	gmPatrolAPITimeout     = 10 * time.Second
	gmPatrolAIMaxAttempts  = 3
	gmPatrolAIDirectBudget = 22 * time.Second
)

const (
	gmPatrolProviderArk    = "ark"
	gmPatrolProviderClaude = "claude"
	gmPatrolProviderLemon  = "lemon"
)

var lemonAPIModelFallbacks = []string{
	"[L]claude-opus-4-8",
	"[L]claude-sonnet-4-8",
}

// GMPatrolAIResult GM 巡逻大模型分析结果。
type GMPatrolAIResult struct {
	Translation string
	Reply       string
	Provider    string
	Model       string
	Elapsed     time.Duration
}

func (r GMPatrolAIResult) ModelDisplay() string {
	model := strings.TrimSpace(r.Model)
	if strings.HasPrefix(model, "[L][V]") {
		model = strings.TrimPrefix(model, "[L][V]")
	} else if strings.HasPrefix(model, "[L]") {
		model = strings.TrimPrefix(model, "[L]")
	}
	if model == "" {
		return "（失败）"
	}
	return model
}

type chatCompletionRequest struct {
	Model    string              `json:"model"`
	Stream   bool                `json:"stream"`
	Messages []chatCompletionMsg `json:"messages"`
}

type chatCompletionMsg struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatCompletionResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func lemonAPIBaseURL() string {
	if v := strings.TrimSpace(Config.GetString("lemon_api.base_url")); v != "" {
		return strings.TrimRight(v, "/")
	}
	return lemonAPIBaseURLDefault
}

func lemonAPIKey() string {
	return strings.TrimSpace(Config.GetString("lemon_api.api_key"))
}

func lemonAPIModel() string {
	if v := strings.TrimSpace(Config.GetString("lemon_api.model")); v != "" {
		return v
	}
	return lemonAPIModelDefault
}

func lemonAPIModelsToTry() []string {
	primary := lemonAPIModel()
	seen := map[string]bool{primary: true}
	models := []string{primary}
	for _, m := range lemonAPIModelFallbacks {
		if !seen[m] {
			seen[m] = true
			models = append(models, m)
		}
	}
	return models
}

// TranslateKoreanToChinese 调用大模型将韩文翻译为简体中文。
func TranslateKoreanToChinese(text string) (string, error) {
	text = strings.TrimSpace(text)
	if text == "" || text == "（未识别）" {
		return "", fmt.Errorf("无可翻译原文")
	}

	apiKey := lemonAPIKey()
	if apiKey == "" {
		return "", fmt.Errorf("未配置 lemon_api.api_key")
	}

	messages := []chatCompletionMsg{
		{
			Role:    "system",
			Content: "你是专业翻译。将用户输入的韩文准确翻译为简体中文，只输出翻译结果，不要解释、不要加引号。",
		},
		{Role: "user", Content: text},
	}

	var lastErr error
	for _, model := range lemonAPIModelsToTry() {
		translation, err := translateWithModel(apiKey, model, messages)
		if err == nil {
			if model != lemonAPIModel() {
				SLS_Log2NoToast("[翻译] 模型 " + lemonAPIModel() + " 不可用，已切换 " + model)
			}
			return translation, nil
		}
		lastErr = err
		SLS_Log2NoToast(fmt.Sprintf("[翻译] 模型 %s 失败: %v", model, err))
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("所有模型均不可用")
	}
	return "", lastErr
}

type gmPatrolAIRuntime struct {
	Provider string
	APIKey   string
	Model    string
	BaseURL  string
	Prompt   string // 非空表示使用配置模板
}

// AnalyzeGMPatrolKorean 分析 GM 巡逻韩文，统一走 AI 作答。
func AnalyzeGMPatrolKorean(ocrText string) (GMPatrolAIResult, error) {
	ocrText = strings.TrimSpace(ocrText)
	if ocrText == "" || ocrText == "（未识别）" {
		return GMPatrolAIResult{}, fmt.Errorf("无可翻译原文")
	}

	start := time.Now()
	if reply, matched, err := gmPatrolLocalAnswer(ocrText); matched {
		if err != nil {
			return GMPatrolAIResult{}, err
		}
		SLS_Log2NoToast(fmt.Sprintf("[GM patrol] local rule answer=%q", reply))
		return GMPatrolAIResult{
			Translation: "（无）",
			Reply:       reply,
			Provider:    "local",
			Model:       "rule-based",
			Elapsed:     time.Since(start),
		}, nil
	}
	rt := resolveGMPatrolAIRuntime()

	roleName := CachedConfigRoleName()
	vocationName := CachedConfigVocationName()
	level := CachedRoleLevel()
	var questionBank []QuestionBankItem
	if API != nil {
		items, err := API.GetQuestionBankAll()
		if err != nil {
			SLS_Log2NoToast("[GM巡逻] 获取题库失败: " + err.Error())
		} else {
			questionBank = items
			SLS_Log2NoToast(fmt.Sprintf("[GM巡逻] 已加载题库 %d 题 role_name=%q vocation_name=%q level=%d",
				len(items), roleName, vocationName, level))
		}
	}

	userPrompt := buildGMPatrolUserPrompt(rt.Prompt, ocrText, roleName, vocationName, level, questionBank)
	SLS_Log2NoToast(fmt.Sprintf("[GM巡逻] 使用 AI provider=%s model=%s key=%q keyLen=%d base=%s",
		rt.Provider, rt.Model, rt.APIKey, len(rt.APIKey), rt.BaseURL))
	logGMPatrolAIPrompt([]chatCompletionMsg{
		{Role: "system", Content: gmPatrolSystemPrompt},
		{Role: "user", Content: userPrompt},
	})

	ctx, cancel := context.WithTimeout(context.Background(), gmPatrolAIDirectBudget)
	defer cancel()

	result, err := tryGMPatrolAI(ctx, rt, userPrompt)
	result.Elapsed = time.Since(start)
	if err == nil {
		SLS_Log2NoToast(fmt.Sprintf("[GM巡逻] AI 成功 provider=%s model=%s reply=%q", result.Provider, result.ModelDisplay(), result.Reply))
	}
	return result, err
}

func resolveGMPatrolAIRuntime() gmPatrolAIRuntime {
	provider := strings.ToLower(strings.TrimSpace(gmPatrolConfigString("ai_provider")))
	prompt := strings.TrimSpace(gmPatrolConfigString("prompt"))

	switch provider {
	case gmPatrolProviderArk:
		apiKey := strings.TrimSpace(gmPatrolConfigString("ark.api_key"))
		model := strings.TrimSpace(gmPatrolConfigString("ark.model"))
		if apiKey != "" && model != "" {
			return gmPatrolAIRuntime{
				Provider: gmPatrolProviderArk,
				APIKey:   apiKey,
				Model:    model,
				BaseURL:  arkAPIBaseURLDefault,
				Prompt:   prompt,
			}
		}
		SLS_Log2NoToast("[GM巡逻] 配置 ark 不完整，回退默认 lemon")
	case gmPatrolProviderClaude:
		apiKey := strings.TrimSpace(gmPatrolConfigString("claude.api_key"))
		model := strings.TrimSpace(gmPatrolConfigString("claude.model"))
		if apiKey != "" && model != "" {
			return gmPatrolAIRuntime{
				Provider: gmPatrolProviderClaude,
				APIKey:   apiKey,
				Model:    model,
				BaseURL:  lemonAPIBaseURL(),
				Prompt:   prompt,
			}
		}
		SLS_Log2NoToast("[GM巡逻] 配置 claude 不完整，回退默认 lemon")
	case gmPatrolProviderLemon:
		// 显式 lemon：优先配置块，再回退本地 lemon_api
		apiKey := strings.TrimSpace(gmPatrolConfigString("claude.api_key"))
		model := strings.TrimSpace(gmPatrolConfigString("claude.model"))
		if apiKey == "" {
			apiKey = lemonAPIKey()
		}
		if model == "" {
			model = gmPatrolAPIModel
		}
		if apiKey != "" {
			return gmPatrolAIRuntime{
				Provider: gmPatrolProviderLemon,
				APIKey:   apiKey,
				Model:    model,
				BaseURL:  lemonAPIBaseURL(),
				Prompt:   prompt,
			}
		}
	}

	// 默认：本地 lemon_api + 代码内置 prompt
	return gmPatrolAIRuntime{
		Provider: gmPatrolProviderLemon,
		APIKey:   lemonAPIKey(),
		Model:    gmPatrolAPIModel,
		BaseURL:  lemonAPIBaseURL(),
		Prompt:   "",
	}
}

func gmPatrolConfigString(subPath string) string {
	if API == nil {
		return ""
	}
	path := gmPatrolConfigRoot
	if subPath != "" {
		path = gmPatrolConfigRoot + "." + subPath
	}
	// 必须用 GetConfigString：GetConfigStringValue 会按 "-" 截断（挂机地图线号用），
	// 会把 sk-xxx / [L]claude-opus-4-8 截成 sk / [L]claude。
	v, err := API.GetConfigString(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(v)
}

func tryGMPatrolAI(ctx context.Context, rt gmPatrolAIRuntime, userPrompt string) (GMPatrolAIResult, error) {
	if strings.TrimSpace(rt.APIKey) == "" {
		return GMPatrolAIResult{}, fmt.Errorf("未配置 %s api_key", rt.Provider)
	}
	if strings.TrimSpace(rt.Model) == "" {
		return GMPatrolAIResult{}, fmt.Errorf("未配置 %s model", rt.Provider)
	}

	var lastErr error
	for attempt := 1; attempt <= gmPatrolAIMaxAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return GMPatrolAIResult{}, err
		}
		var (
			raw string
			err error
		)
		switch rt.Provider {
		case gmPatrolProviderArk:
			raw, err = arkResponsesCompletion(ctx, rt.BaseURL, rt.APIKey, rt.Model, userPrompt, gmPatrolAPITimeout)
		default:
			messages := []chatCompletionMsg{
				{Role: "system", Content: gmPatrolSystemPrompt},
				{Role: "user", Content: userPrompt},
			}
			raw, err = chatCompletion(ctx, rt.BaseURL, rt.APIKey, rt.Model, messages, gmPatrolAPITimeout)
		}
		if err != nil {
			lastErr = err
		} else {
			result, parseErr := parseGMPatrolAIResult(raw)
			if parseErr == nil {
				result.Provider = rt.Provider
				result.Model = rt.Model
				SLS_Log2NoToast(fmt.Sprintf("[GM巡逻] %s %s 第 %d 次成功", rt.Provider, rt.Model, attempt))
				return result, nil
			}
			lastErr = parseErr
		}
		if attempt < gmPatrolAIMaxAttempts {
			SLS_Log2NoToast(fmt.Sprintf("[GM巡逻] %s %s 第 %d 次失败，重试: %v", rt.Provider, rt.Model, attempt, lastErr))
		}
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("所有重试均失败")
	}
	return GMPatrolAIResult{Provider: rt.Provider, Model: rt.Model}, lastErr
}

func buildGMPatrolUserPrompt(template, ocrText, roleName, vocationName string, level int, questionBank []QuestionBankItem) string {
	if roleName == "" {
		roleName = "（未知）"
	}
	if vocationName == "" {
		vocationName = "（未知）"
	}
	levelStr := gmPatrolLevelStr(level)
	bankText := formatQuestionBankForPrompt(questionBank)

	if tpl := strings.TrimSpace(template); tpl != "" {
		return renderGMPatrolPromptTemplate(tpl, ocrText, roleName, vocationName, levelStr, bankText)
	}

	return fmt.Sprintf(`%s

玩家角色信息：
角色名（role_name）：%s
职业（vocation_name）：%s
等级（level）：%s

%s

%s

题库参考（韩文题目与标准答案；若为知识问答/选择题，请优先在题库中匹配并采用对应 answer）：
%s

【本次请求 #%d】

韩文原文如下：
%s

请完成这件事：
确定玩家应输入的回复内容（只给出应输入的内容本身）：
若要求输入角色名，使用上方 role_name（韩文原样）
若涉及职业，参考 vocation_name
若要求输入等级，使用上方 level 的数字
若为题库题目，使用题库中的 answer（如 A/B/C/D 或指定韩文）
其他按对话要求，韩文原样输出应输入的内容

%s

你的回复有且仅有答案本身，不要任何其他文字！！！`, gmPatrolTestScenarioNote, roleName, vocationName, levelStr, gmPatrolPersonaRules, gmPatrolOCRNotes, bankText, gmPatrolRequestNonce(), ocrText, gmPatrolOutputRules)
}

func renderGMPatrolPromptTemplate(tpl, ocrText, roleName, vocationName, levelStr, bankText string) string {
	replacer := strings.NewReplacer(
		"${role_name}", roleName,
		"${vocation_name}", vocationName,
		"${level}", levelStr,
		"${问题}", ocrText,
		"{题库按顺序拼接}", bankText,
		"${角色相关问题}", "",
	)
	return replacer.Replace(tpl)
}

func arkResponsesCompletion(ctx context.Context, baseURL, apiKey, model, userPrompt string, timeout time.Duration) (string, error) {
	reqBody := map[string]interface{}{
		"model": model,
		"input": []map[string]interface{}{
			{
				"role": "user",
				"content": []map[string]interface{}{
					{
						"type": "input_text",
						"text": userPrompt,
					},
				},
			},
		},
	}
	payload, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	url := strings.TrimRight(baseURL, "/") + "/responses"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := (&http.Client{Timeout: timeout}).Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("http=%d body=%s", resp.StatusCode, string(body))
	}

	text, err := extractArkResponsesText(body)
	if err != nil {
		return "", err
	}
	return text, nil
}

func extractArkResponsesText(body []byte) (string, error) {
	var root map[string]interface{}
	if err := json.Unmarshal(body, &root); err != nil {
		return "", fmt.Errorf("解析响应失败: %v", err)
	}
	if errObj, ok := root["error"].(map[string]interface{}); ok {
		if msg, _ := errObj["message"].(string); strings.TrimSpace(msg) != "" {
			return "", fmt.Errorf("%s", msg)
		}
	}

	if output, ok := root["output"].([]interface{}); ok {
		var parts []string
		for _, item := range output {
			msg, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			content, _ := msg["content"].([]interface{})
			for _, c := range content {
				part, ok := c.(map[string]interface{})
				if !ok {
					continue
				}
				typ, _ := part["type"].(string)
				if typ == "output_text" || typ == "text" {
					if text, _ := part["text"].(string); strings.TrimSpace(text) != "" {
						parts = append(parts, strings.TrimSpace(text))
					}
				}
			}
		}
		if joined := strings.TrimSpace(strings.Join(parts, "\n")); joined != "" {
			return joined, nil
		}
	}

	// 兼容少数网关把文本放在 output_text / text 顶层
	for _, key := range []string{"output_text", "text"} {
		if text, _ := root[key].(string); strings.TrimSpace(text) != "" {
			return strings.TrimSpace(text), nil
		}
	}
	return "", fmt.Errorf("响应无文本内容: %s", string(body))
}

func gmPatrolRequestNonce() int {
	return rand.Intn(900000) + 100000
}

func gmPatrolLevelStr(level int) string {
	if level >= 10 && level <= 250 {
		return fmt.Sprintf("%d", level)
	}
	return fmt.Sprintf("%d", rand.Intn(41)+80)
}

func logGMPatrolAIPrompt(messages []chatCompletionMsg) {
	fmt.Println("[GM巡逻] ========== 发给 AI 的内容 START ==========")
	for _, m := range messages {
		fmt.Printf("--- role: %s ---\n%s\n", m.Role, m.Content)
	}
	fmt.Println("[GM巡逻] ========== 发给 AI 的内容 END ==========")
}

func parseGMPatrolAIResult(raw string) (GMPatrolAIResult, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return GMPatrolAIResult{}, fmt.Errorf("响应为空")
	}
	reply := trimGMPatrolAnswer(raw)
	if reply == "" {
		return GMPatrolAIResult{}, fmt.Errorf("回答为空")
	}
	return GMPatrolAIResult{Translation: "（无）", Reply: reply}, nil
}

func trimGMPatrolAnswer(raw string) string {
	raw = strings.TrimSpace(raw)
	for _, prefix := range []string{"回答：", "回答:", "答案：", "答案:", "答：", "答:"} {
		if strings.HasPrefix(raw, prefix) {
			raw = strings.TrimSpace(strings.TrimPrefix(raw, prefix))
		}
	}
	raw = strings.Trim(raw, "`\"' \t")
	return normalizeGMPatrolAnswerLine(raw)
}

func normalizeGMPatrolAnswerLine(raw string) string {
	raw = strings.ReplaceAll(raw, "\r\n", "\n")
	raw = strings.ReplaceAll(raw, "\r", "\n")
	var parts []string
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			parts = append(parts, line)
		}
	}
	if len(parts) == 0 {
		return ""
	}
	if len(parts) == 1 {
		return parts[0]
	}
	return strings.Join(parts, " ")
}

func translateWithModel(apiKey, model string, messages []chatCompletionMsg) (string, error) {
	return chatCompletion(context.Background(), lemonAPIBaseURL(), apiKey, model, messages, llmAPITimeout)
}

func chatCompletion(ctx context.Context, baseURL, apiKey, model string, messages []chatCompletionMsg, timeout time.Duration) (string, error) {
	reqBody := chatCompletionRequest{
		Model:    model,
		Stream:   false,
		Messages: messages,
	}
	payload, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	url := strings.TrimRight(baseURL, "/") + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := (&http.Client{Timeout: timeout}).Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result chatCompletionResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("解析响应失败: %v", err)
	}
	if result.Error != nil && result.Error.Message != "" {
		return "", fmt.Errorf("%s", result.Error.Message)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("http=%d body=%s", resp.StatusCode, string(body))
	}
	if len(result.Choices) == 0 {
		return "", fmt.Errorf("响应无内容")
	}

	content := strings.TrimSpace(result.Choices[0].Message.Content)
	if content == "" {
		return "", fmt.Errorf("响应内容为空")
	}
	return content, nil
}
