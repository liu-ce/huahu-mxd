// GM 巡逻 AI 压测（Lemon）：对 Lemon 各模型各调用 N 次，输出每次结果/耗时及平均耗时。
//
// 用法（项目根目录）:
//
//	go run ./cmd/bench_gm_patrol_ai
//	go run ./cmd/bench_gm_patrol_ai -rounds 20
//	go run ./cmd/bench_gm_patrol_ai -model "[L]claude-opus-4-8"
//	go run ./cmd/bench_gm_patrol_ai -models "[L]claude-opus-4-8,[L]claude-sonnet-4-8"
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const sampleOCRText = "안녕하세요. 메이플 플래닛입니다.정상적인 게임 플레이 상태를 확인 진행중입니다.현재 화면을 확인하고 조작 중이시라면,아래의 요청 사항을제한 시간 내에 정확하게 입력해 주시기 바랍니다.아래 칸에 자신의 캐릭터 이름과 직업을 입력하세요.※현재 모험가님이 접속중인 캐릭터 이름과 직업입니다.※제한 시간 내 답변을 꼭 입력해 주세요"

const (
	myRoleName  = "절망의집행자"   // 我的游戏名
	myVocation  = "메이지 (썬콜)" // 我的职业
	sampleLevel = 135
)

const gmPatrolSystemPrompt = `你是韩文信息抽取助手。从原文判断「玩家应填写什么」，只输出最终答案。
严禁输出分析、解释、说明、总结、前缀、示例或任何与答案无关的文字。
你的完整回复就是答案本身，单行输出不要使用换行；若问题中指定了分隔符则优先使用该分隔符，未指定时多个答案可用空格分隔，无任何其他字符。`

const gmPatrolTestScenarioNote = `请从韩文原文提取应填写的内容。只输出答案，不要分析过程。`

const gmPatrolPersonaRules = `身份与人设（永远以一个真实韩国玩家视角作答）：
- 若问题询问你是否是 AI/机器人/ChatGPT/模型等，回答：나는 AI가 아니고 어떤 모델도 없어
- 若问题询问你来自哪里/国籍/是哪国人，回答：한국 사람이야
- 除下方硬性规则外，其他开放性问题均按韩国人日常口吻用韩文回答`

const gmPatrolOutputRules = `【输出格式 - 必须严格遵守】
- 只输出最终应填写的内容，一个字都不要多
- 禁止：分析、解释、说明、总结、前缀、后缀、示例、括号备注
- 答案必须单行输出，不要使用换行
- 多个答案：若原文/问题中指定了分隔符（如 /、, 等），优先使用该分隔符；未指定时可用空格分隔`

const gmPatrolOCRNotes = `【OCR 易错识别 - 勿按字面理解或翻译】
- 「메이플 프레넛」是 OCR 对「메이플스토리」(冒险岛/MapleStory) 的误识，不是「花生/peanut」
- 「메이플 프레넛의 최대 레벨은?」= 冒险岛的最高等级是多少？（韩服当前满级 300）
- 「메이플 프레넛의 최대 잠재 등급은?」= 冒险岛的最高潜能等级是什么？（最高为 레전더리）
- 「창작무민」可能是 OCR 对「장착한」(装备的) 的误识，勿按字面理解为「创作木民」`

var defaultLemonModels = []string{
	"[L]claude-opus-4-8",
}

var sampleQuestionBank = []questionBankItem{
	{QuestionKo: "메이플 프레넛의 최대 레벨은?", Meaning: "冒险岛的最高等级是多少？", Answer: "300"},
	{QuestionKo: "메이플 프레넛의 최대 잠재 등급은?", Meaning: "冒险岛的最高潜能等级是什么？", Answer: "레전더리"},
	{QuestionKo: "지금 생각나는 영웅 직업군은?", Meaning: "随便说一个英雄职业群（开放题，任一合法英雄职业均可）", Answer: "아란"},
	{QuestionKo: "현재 사용중인 언어는?", Meaning: "当前使用的语言是？", Answer: "한국어"},
	{QuestionKo: "장착한 무기 이름은?", Meaning: "当前装备的武器名称", Answer: "피닉스 완드"},
	{QuestionKo: "현재 창작무민 무기 아이템 이름은?", Meaning: "当前装备的武器名称（창작무민=장착한 OCR误识）", Answer: "피닉스 완드"},
	{QuestionKo: "현재 생각나는 물약 아이템 이름은?", Meaning: "随便说一种药水道具名（开放题）", Answer: "마나 엘릭서"},
}

type benchRound struct {
	Index   int
	Elapsed time.Duration
	Reply   string
	Err     error
}

type benchSummary struct {
	Model   string
	Rounds  []benchRound
	Success int
	Fail    int
	Avg     time.Duration
	Min     time.Duration
	Max     time.Duration
}

type questionBankItem struct {
	QuestionKo string
	Meaning    string
	Answer     string
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

type benchPromptParams struct {
	Text     string
	Role     string
	Vocation string
	Level    int
}

func main() {
	configPath := flag.String("config", "assets/config/default.json", "配置文件")
	rounds := flag.Int("rounds", 20, "调用次数")
	timeoutSec := flag.Int("timeout", 45, "单次请求超时（秒）")
	modelFlag := flag.String("model", "", "只测单个 Lemon 模型")
	modelsFlag := flag.String("models", "", "Lemon 模型列表，逗号分隔（默认内置 5 个模型）")
	text := flag.String("text", sampleOCRText, "OCR 韩文原文")
	roleName := flag.String("role", myRoleName, "role_name（我的游戏名）")
	vocation := flag.String("vocation", myVocation, "vocation_name（我的职业）")
	level := flag.Int("level", sampleLevel, "等级")
	flag.Parse()

	cfg, err := loadConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "加载配置失败: %v\n", err)
		os.Exit(1)
	}
	if cfg.APIKey == "" {
		fmt.Fprintln(os.Stderr, "未配置 lemon_api.api_key")
		os.Exit(1)
	}
	models := defaultLemonModels
	if m := strings.TrimSpace(*modelsFlag); m != "" {
		models = splitCSV(m)
	}
	if m := strings.TrimSpace(*modelFlag); m != "" {
		models = []string{m}
	}

	promptParams := benchPromptParams{
		Text: *text, Role: *roleName, Vocation: *vocation, Level: *level,
	}
	timeout := time.Duration(*timeoutSec) * time.Second

	fmt.Printf("=== GM 巡逻 AI 压测 (Lemon) === models=%d rounds=%d timeout=%s\n\n", len(models), *rounds, timeout)

	var summaries []benchSummary
	for _, model := range models {
		summaries = append(summaries, runBench(cfg.BaseURL, cfg.APIKey, model, promptParams, *rounds, timeout))
	}

	fmt.Println("========== 汇总 ==========")
	for _, s := range summaries {
		fmt.Printf("[lemon] %-40s 成功 %2d/%2d  平均 %.2fs  最短 %.2fs  最长 %.2fs\n",
			s.Model,
			s.Success, len(s.Rounds),
			s.Avg.Seconds(), s.Min.Seconds(), s.Max.Seconds())
	}
}

func runBench(baseURL, apiKey, model string, params benchPromptParams, rounds int, timeout time.Duration) benchSummary {
	summary := benchSummary{Model: model}
	fmt.Printf("========== [lemon] %s x%d ==========\n", model, rounds)

	for i := 1; i <= rounds; i++ {
		userPrompt := buildGMPatrolUserPrompt(params.Text, params.Role, params.Vocation, params.Level, sampleQuestionBank)
		messages := []chatCompletionMsg{
			{Role: "system", Content: gmPatrolSystemPrompt},
			{Role: "user", Content: userPrompt},
		}

		start := time.Now()
		reply, err := callChatCompletion(baseURL, apiKey, model, messages, timeout)
		elapsed := time.Since(start)

		round := benchRound{Index: i, Elapsed: elapsed, Reply: reply, Err: err}
		summary.Rounds = append(summary.Rounds, round)

		if err != nil {
			summary.Fail++
			fmt.Printf("%02d  %.2fs  FAIL  %v\n", i, elapsed.Seconds(), err)
			continue
		}
		reply = normalizeReply(reply)
		round.Reply = reply
		summary.Rounds[len(summary.Rounds)-1] = round
		summary.Success++
		fmt.Printf("%02d  %.2fs  OK    %s\n", i, elapsed.Seconds(), reply)
	}

	var total time.Duration
	summary.Min = time.Duration(1<<63 - 1)
	for _, r := range summary.Rounds {
		if r.Err != nil {
			continue
		}
		total += r.Elapsed
		if r.Elapsed < summary.Min {
			summary.Min = r.Elapsed
		}
		if r.Elapsed > summary.Max {
			summary.Max = r.Elapsed
		}
	}
	if summary.Success > 0 {
		summary.Avg = total / time.Duration(summary.Success)
	} else {
		summary.Min = 0
	}
	fmt.Printf("----------------------------------------\n")
	fmt.Printf("成功 %d/%d  平均 %.2fs  最短 %.2fs  最长 %.2fs\n\n",
		summary.Success, len(summary.Rounds),
		summary.Avg.Seconds(), summary.Min.Seconds(), summary.Max.Seconds())
	return summary
}

type lemonConfig struct {
	BaseURL string
	APIKey  string
	Model   string
}

func loadConfig(path string) (lemonConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return lemonConfig{}, err
	}
	var root map[string]interface{}
	if err := json.Unmarshal(data, &root); err != nil {
		return lemonConfig{}, err
	}
	cfg := lemonConfig{
		BaseURL: "https://www.lemonapi.ai/v1",
		Model:   "[L]claude-opus-4-8",
	}
	if lemon, _ := root["lemon_api"].(map[string]interface{}); lemon != nil {
		if v, ok := lemon["base_url"].(string); ok && strings.TrimSpace(v) != "" {
			cfg.BaseURL = strings.TrimRight(strings.TrimSpace(v), "/")
		}
		if v, ok := lemon["api_key"].(string); ok {
			cfg.APIKey = strings.TrimSpace(v)
		}
		if v, ok := lemon["model"].(string); ok && strings.TrimSpace(v) != "" {
			cfg.Model = strings.TrimSpace(v)
		}
	}
	if key := strings.TrimSpace(os.Getenv("LEMON_API_KEY")); key != "" {
		cfg.APIKey = key
	}
	return cfg, nil
}

func callChatCompletion(baseURL, apiKey, model string, messages []chatCompletionMsg, timeout time.Duration) (string, error) {
	reqBody := chatCompletionRequest{Model: model, Stream: false, Messages: messages}
	payload, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}
	url := strings.TrimRight(baseURL, "/") + "/chat/completions"
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(payload))
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
		return "", fmt.Errorf("parse: %v", err)
	}
	if result.Error != nil && result.Error.Message != "" {
		return "", fmt.Errorf("%s", result.Error.Message)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("http=%d %s", resp.StatusCode, string(body))
	}
	if len(result.Choices) == 0 {
		return "", fmt.Errorf("empty choices")
	}
	content := strings.TrimSpace(result.Choices[0].Message.Content)
	if content == "" {
		return "", fmt.Errorf("empty content")
	}
	return content, nil
}

func buildGMPatrolUserPrompt(ocrText, roleName, vocationName string, level int, bank []questionBankItem) string {
	levelStr := gmPatrolLevelStr(level)
	var b strings.Builder
	for i, q := range bank {
		if q.Meaning != "" {
			fmt.Fprintf(&b, "%d. %s（%s）→ %s\n", i+1, q.QuestionKo, q.Meaning, q.Answer)
		} else {
			fmt.Fprintf(&b, "%d. %s → %s\n", i+1, q.QuestionKo, q.Answer)
		}
	}
	bankText := strings.TrimRight(b.String(), "\n")
	if bankText == "" {
		bankText = "（无）"
	}
	return fmt.Sprintf(`%s

玩家角色信息：
角色名（role_name）：%s
职业（vocation_name）：%s
等级（level）：%s

%s

%s

题库参考：
%s

【本次请求 #%d】

韩文原文如下：
%s

请完成这件事：
确定玩家应输入的回复内容（只给出应输入的内容本身）：

%s

你的回复有且仅有答案本身，不要任何其他文字！！！`,
		gmPatrolTestScenarioNote, roleName, vocationName, levelStr,
		gmPatrolPersonaRules, gmPatrolOCRNotes, bankText,
		gmPatrolRequestNonce(), strings.TrimSpace(ocrText), gmPatrolOutputRules)
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

func normalizeReply(raw string) string {
	for _, prefix := range []string{"回答：", "回答:", "答案：", "答案:", "答：", "答:"} {
		if strings.HasPrefix(raw, prefix) {
			raw = strings.TrimSpace(strings.TrimPrefix(raw, prefix))
		}
	}
	raw = strings.Trim(raw, "`\"' \t")
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

func splitCSV(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func init() {
	if _, err := os.Stat("assets/config/default.json"); err == nil {
		return
	}
	exe, err := os.Executable()
	if err != nil {
		return
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(exe), "..", ".."))
	candidate := filepath.Join(root, "assets", "config", "default.json")
	if _, err := os.Stat(candidate); err == nil {
		_ = os.Chdir(root)
	}
}
