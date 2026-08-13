// GM 巡逻 AI 汇总压测：按预设/自定义 OCR 题目逐题压测，输出每题完整结果。
//
// 用法（项目根目录）:
//
//	go run ./cmd/bench_gm_patrol_suite
//	go run ./cmd/bench_gm_patrol_suite -rounds 20
//	go run ./cmd/bench_gm_patrol_suite -model "[L]claude-opus-4-6" -rounds 10
//	go run ./cmd/bench_gm_patrol_suite -cases assets/config/gm_patrol_suite_cases.json
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

const (
	myRoleName  = "절망의집행자"
	myVocation  = "메이지 (썬콜)"
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

// var defaultSuiteCases = []suiteCase{
// 	{
// 		Name: "题型1",
// 		OCR:  "숫자와 영문이 섞인 9자리 임의 랜덤 문자를 입력하세요.",
// 	},
// 	{
// 		Name: "题型2",
// 		OCR:  "숫자, 대문자 영문, 소문자 영문을 각각 1개 이상 포함한 17자리 임의 랜덤 문자를 입력하세요.",
// 	},
// 	{
// 		Name: "题型3",
// 		OCR:  "숫자, 대문자 영문, 소문자 영문을 각각 1개 이상 포함한 17자리 임의 랜덤 문자를 입력하세요.",
// 	},
// 	{
// 		Name: "题型4",
// 		OCR:  "숫자, 대문자 영문, 소문자 영문을 각각 1개 이상 포함한 17자리 임의 랜덤 문자를 입력하세요.",
// 	},
// 	{
// 		Name: "题型5",
// 		OCR:  "서로 다른 한글 17글자를 임의로 입력해 주세요.",
// 	},
// 	{
// 		Name: "题型6",
// 		OCR:  "서로 다른 한글 17글자를 직접 입력해 주세요.",
// 	},
// }

var defaultSuiteCases = []suiteCase{
	{
		Name: "题型1",
		OCR:  "아래 요청 사항을 확인 후, 요청된 답만을 입력하세요.[요청 사항]대문자와 소문자가 섞인 서로 다른 영문 12글자를 임의로 입력해 주세요[규척 안내]※ 대문자와 소문자는 각각 1개 이상 포함※ 같은 글자 반복, ABCD/abcd 등 순차 입력 제외※ 숫자, 한글,공백, 특수문자 제외",
	},
}

var sampleQuestionBank = []questionBankItem{
	{QuestionKo: "메이플 프레넛의 최대 레벨은?", Meaning: "冒险岛的最高等级是多少？", Answer: "300"},
	{QuestionKo: "메이플 프레넛의 최대 잠재 등급은?", Meaning: "冒险岛的最高潜能等级是什么？", Answer: "레전더리"},
	{QuestionKo: "지금 생각나는 영웅 직업군은?", Meaning: "随便说一个英雄职业群（开放题）", Answer: "아란"},
	{QuestionKo: "현재 사용중인 언어는?", Meaning: "当前使用的语言是？", Answer: "한국어"},
	{QuestionKo: "장착한 무기 이름은?", Meaning: "当前装备的武器名称", Answer: "피닉스 완드"},
	{QuestionKo: "현재 창작무민 무기 아이템 이름은?", Meaning: "当前装备的武器名称", Answer: "피닉스 완드"},
	{QuestionKo: "현재 생각나는 물약 아이템 이름은?", Meaning: "随便说一种药水道具名（开放题）", Answer: "마나 엘릭서"},
}

type suiteCase struct {
	Name string `json:"name"`
	OCR  string `json:"ocr"`
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

type benchPromptParams struct {
	Text     string
	Role     string
	Vocation string
	Level    int
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

type lemonConfig struct {
	BaseURL string
	APIKey  string
	Model   string
}

func main() {
	configPath := flag.String("config", "assets/config/default.json", "配置文件")
	casesPath := flag.String("cases", "", "题目 JSON 文件（默认内置 3 题）")
	rounds := flag.Int("rounds", 20, "每题调用次数")
	timeoutSec := flag.Int("timeout", 45, "单次请求超时（秒）")
	modelFlag := flag.String("model", "", "Lemon 模型（默认读配置 lemon_api.model）")
	roleName := flag.String("role", myRoleName, "role_name")
	vocation := flag.String("vocation", myVocation, "vocation_name")
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
	model := cfg.Model
	if m := strings.TrimSpace(*modelFlag); m != "" {
		model = m
	}

	cases := defaultSuiteCases
	if p := strings.TrimSpace(*casesPath); p != "" {
		cases, err = loadSuiteCases(p)
		if err != nil {
			fmt.Fprintf(os.Stderr, "加载题目失败: %v\n", err)
			os.Exit(1)
		}
	}
	if len(cases) == 0 {
		fmt.Fprintln(os.Stderr, "没有可压测的题目")
		os.Exit(1)
	}

	params := benchPromptParams{
		Role: *roleName, Vocation: *vocation, Level: *level,
	}
	timeout := time.Duration(*timeoutSec) * time.Second
	modelName := modelDisplay(model)

	fmt.Printf("=== GM 巡逻 AI 汇总压测 === model=%s cases=%d rounds=%d timeout=%s\n\n",
		model, len(cases), *rounds, timeout)

	var allSummaries []benchSummary
	for i, c := range cases {
		if i > 0 {
			fmt.Println()
		}
		fmt.Printf("%s：%s\n\n", c.Name, c.OCR)
		fmt.Println("模型返回结果和耗时：")
		params.Text = c.OCR
		summary := runBench(cfg.BaseURL, cfg.APIKey, model, modelName, params, *rounds, timeout)
		allSummaries = append(allSummaries, summary)
	}

	fmt.Println("\n========== 总汇总 ==========")
	for i, c := range cases {
		s := allSummaries[i]
		fmt.Printf("%s  成功 %2d/%2d  平均 %.2fs  最短 %.2fs  最长 %.2fs\n",
			c.Name, s.Success, len(s.Rounds),
			s.Avg.Seconds(), s.Min.Seconds(), s.Max.Seconds())
	}
}

func runBench(baseURL, apiKey, model, modelName string, params benchPromptParams, rounds int, timeout time.Duration) benchSummary {
	summary := benchSummary{Model: model}
	fmt.Printf("========== %s x%d ==========\n", modelName, rounds)

	for i := 1; i <= rounds; i++ {
		userPrompt := buildGMPatrolUserPrompt(params.Text, params.Role, params.Vocation, params.Level, sampleQuestionBank)
		messages := []chatCompletionMsg{
			{Role: "system", Content: gmPatrolSystemPrompt},
			{Role: "user", Content: userPrompt},
		}

		start := time.Now()
		reply, err := callChatCompletion(baseURL, apiKey, model, messages, timeout)
		elapsed := time.Since(start)
		if err == nil {
			reply = normalizeReply(reply)
		}

		round := benchRound{Index: i, Elapsed: elapsed, Reply: reply, Err: err}
		summary.Rounds = append(summary.Rounds, round)

		if err != nil {
			summary.Fail++
			fmt.Printf("%02d  %.2fs  FAIL  %v\n", i, elapsed.Seconds(), err)
			continue
		}
		summary.Success++
		fmt.Printf("%02d  %.2fs  OK    %s\n", i, elapsed.Seconds(), reply)
	}

	var total time.Duration
	summary.Min = time.Duration(1<<63 - 1)
	for _, r := range summary.Rounds {
		if r.Err != nil || r.Reply == "" {
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
	fmt.Printf("成功 %d/%d  平均 %.2fs  最短 %.2fs  最长 %.2fs\n",
		summary.Success, len(summary.Rounds),
		summary.Avg.Seconds(), summary.Min.Seconds(), summary.Max.Seconds())
	return summary
}

func loadSuiteCases(path string) ([]suiteCase, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cases []suiteCase
	if err := json.Unmarshal(data, &cases); err != nil {
		return nil, err
	}
	for i := range cases {
		cases[i].Name = strings.TrimSpace(cases[i].Name)
		cases[i].OCR = strings.TrimSpace(cases[i].OCR)
		if cases[i].Name == "" {
			cases[i].Name = fmt.Sprintf("问题%d", i+1)
		}
		if cases[i].OCR == "" {
			return nil, fmt.Errorf("题目 %s 缺少 ocr", cases[i].Name)
		}
	}
	return cases, nil
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
若要求输入角色名，使用上方 role_name（韩文原样）
若涉及职业，参考 vocation_name

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

func modelDisplay(model string) string {
	m := strings.TrimSpace(model)
	if strings.HasPrefix(m, "[L][V]") {
		return strings.TrimPrefix(m, "[L][V]")
	}
	if strings.HasPrefix(m, "[L]") {
		return strings.TrimPrefix(m, "[L]")
	}
	return m
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
