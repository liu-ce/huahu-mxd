package core

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
)

// QuestionBankItem 题库条目。
type QuestionBankItem struct {
	ID         int    `json:"id"`
	QuestionKo string `json:"question_ko"`
	Answer     string `json:"answer"`
}

type questionBankResponse struct {
	Code    int                `json:"code"`
	Message string             `json:"message"`
	Data    []QuestionBankItem `json:"data"`
}

var questionBankCache struct {
	mu   sync.RWMutex
	Data []QuestionBankItem
}

// CachedConfigRoleName 从配置缓存读取服务端 role_name。
func CachedConfigRoleName() string {
	configCache.mu.RLock()
	defer configCache.mu.RUnlock()
	if configCache.Data == nil {
		return ""
	}
	return configCache.Data.RoleName
}

// CachedConfigVocationName 从配置缓存读取服务端 vocation_name。
func CachedConfigVocationName() string {
	configCache.mu.RLock()
	defer configCache.mu.RUnlock()
	if configCache.Data == nil {
		return ""
	}
	return configCache.Data.VocationName
}

// CachedConfigAIAnswerCustom 从配置缓存读取 ai_answer_custom。
func CachedConfigAIAnswerCustom() string {
	configCache.mu.RLock()
	defer configCache.mu.RUnlock()
	if configCache.Data == nil {
		return ""
	}
	return strings.TrimSpace(configCache.Data.AIAnswerCustom)
}

// CachedConfigAutoAnswerEnabled GM 巡逻是否自动点确定提交（未配置时默认 true）。
// false 时仍走 AI 回答、钉钉、输入、答题记录，仅不自动点确定。
func CachedConfigAutoAnswerEnabled() bool {
	configCache.mu.RLock()
	defer configCache.mu.RUnlock()
	if configCache.Data == nil || configCache.Data.AutoAnswerEnabled == nil {
		return true
	}
	return *configCache.Data.AutoAnswerEnabled
}

// CachedRoleLevel 从 core.Role 读取当前等级（OCR/RoleUpdate 写入）。
func CachedRoleLevel() int {
	if Role == nil {
		return 0
	}
	return Role.Level
}

// GetQuestionBankAll 获取题库（优先读缓存）。
func (c *APIClient) GetQuestionBankAll() ([]QuestionBankItem, error) {
	questionBankCache.mu.RLock()
	if questionBankCache.Data != nil {
		out := append([]QuestionBankItem(nil), questionBankCache.Data...)
		questionBankCache.mu.RUnlock()
		return out, nil
	}
	questionBankCache.mu.RUnlock()
	return c.fetchQuestionBankAll()
}

// RefreshQuestionBank 强制拉取并更新题库缓存。
func (c *APIClient) RefreshQuestionBank() ([]QuestionBankItem, error) {
	return c.fetchQuestionBankAll()
}

func (c *APIClient) fetchQuestionBankAll() ([]QuestionBankItem, error) {
	tokenInterface := Get("token")
	if tokenInterface == nil {
		return nil, fmt.Errorf("未找到token，请先登录")
	}
	token, ok := tokenInterface.(string)
	if !ok {
		return nil, fmt.Errorf("token格式错误")
	}

	url := c.baseURL + "/api/question-bank/all"
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送GET请求失败: %v", err)
	}
	defer resp.Body.Close()

	var bankResp questionBankResponse
	if err := json.NewDecoder(resp.Body).Decode(&bankResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("请求失败，状态码: %d, 消息: %s", resp.StatusCode, bankResp.Message)
	}
	if bankResp.Code != 200 {
		return nil, fmt.Errorf("获取题库失败: %s", bankResp.Message)
	}

	items := bankResp.Data
	if items == nil {
		items = []QuestionBankItem{}
	}

	questionBankCache.mu.Lock()
	questionBankCache.Data = append([]QuestionBankItem(nil), items...)
	questionBankCache.mu.Unlock()

	return items, nil
}

func formatQuestionBankForPrompt(items []QuestionBankItem) string {
	if len(items) == 0 {
		return "（无）"
	}
	var b strings.Builder
	for i, q := range items {
		b.WriteString(fmt.Sprintf("%d. %s → %s\n", i+1, q.QuestionKo, q.Answer))
	}
	return strings.TrimRight(b.String(), "\n")
}
