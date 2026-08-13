package core

import (
	"os"
	"strconv"
	"testing"
)

const defaultGeneratedAnswerTestRounds = 100

// generatedAnswerTestRounds 允许本地通过环境变量调整生成轮数，未配置时保持 100 轮回归验证。
func generatedAnswerTestRounds(t *testing.T) int {
	t.Helper()
	raw := os.Getenv("GM_PATROL_GENERATION_ROUNDS")
	if raw == "" {
		return defaultGeneratedAnswerTestRounds
	}
	rounds, err := strconv.Atoi(raw)
	if err != nil || rounds < 1 {
		t.Fatalf("GM_PATROL_GENERATION_ROUNDS must be a positive integer, got %q", raw)
	}
	return rounds
}

func TestGMPatrolLocalAnswerRecognizesDistinctMixedCaseEnglishRequest(t *testing.T) {
	prompt := "아래 요청 사항을 확인 후, 요청된 답만을 입력하세요. 대문자와 소문자가 섞인 서로 다른 영문 12글자를 임의로 입력해 주세요. 같은 글자 반복, qwer/abcd 등의 키 배열 및 순차 입력은 정상 답변에서 제외됩니다."

	prompt = "\uB300\uBB38\uC790\uC640 \uC18C\uBB38\uC790\uAC00 \uC11E\uC778 \uC11C\uB85C \uB2E4\uB978 \uC601\uBB38 12\uAE00\uC790\uB97C \uC784\uC758\uB85C \uC785\uB825\uD558\uC138\uC694. \uAC19\uC740 \uAE00\uC790 \uBC18\uBCF5, qwer/abcd \uB4F1 \uD0A4 \uBC30\uC5F4 \uBC0F \uC21C\uCC28 \uC785\uB825 \uC81C\uC678. \uC22B\uC790, \uD55C\uAE00, \uACF5\uBC31, \uD2B9\uC218\uBB38\uC790 \uC81C\uC678"
	answer, matched, err := gmPatrolLocalAnswer(prompt)
	if err != nil {
		t.Fatalf("gmPatrolLocalAnswer returned an error: %v", err)
	}
	if !matched {
		t.Fatal("expected prompt to be handled locally")
	}
	if !isValidDistinctMixedCaseEnglish(answer) {
		t.Fatalf("generated invalid local answer: %q", answer)
	}
}

func TestGMPatrolLocalAnswerDoesNotMatchUnrelatedQuestion(t *testing.T) {
	answer, matched, err := gmPatrolLocalAnswer("현재 캐릭터의 레벨을 입력하세요.")
	if err != nil {
		t.Fatalf("gmPatrolLocalAnswer returned an error: %v", err)
	}
	if matched || answer != "" {
		t.Fatalf("unexpected local answer: matched=%v answer=%q", matched, answer)
	}
}

func TestGMPatrolLocalAnswerFallsBackWhenRulesAreIncomplete(t *testing.T) {
	tests := []struct {
		name   string
		prompt string
	}{
		{
			name:   "English missing special-character exclusion",
			prompt: "\uB300\uBB38\uC790\uC640 \uC18C\uBB38\uC790\uAC00 \uC11E\uC778 \uC11C\uB85C \uB2E4\uB978 \uC601\uBB38 12\uAE00\uC790\uB97C \uC785\uB825\uD558\uC138\uC694. \uAC19\uC740 \uAE00\uC790 \uBC18\uBCF5, qwer/abcd \uB4F1 \uC21C\uCC28/\uBC30\uC5F4 \uC81C\uC678. \uC22B\uC790, \uD55C\uAE00, \uACF5\uBC31 \uC81C\uC678",
		},
		{
			name:   "Hangul missing special-character exclusion",
			prompt: "\uC11C\uB85C \uB2E4\uB978 \uD55C\uAE00 12\uAE00\uC790\uB97C \uC785\uB825\uD558\uC138\uC694. \uAC19\uC740 \uAE00\uC790 \uBC18\uBCF5, \uAC00\uB098\uB2E4\uB77C \uB4F1 \uC21C\uCC28/\uBC30\uC5F4 \uC81C\uC678, \uC790\uC74C \uBAA8\uC74C \uB2E8\uB3C5 \uC81C\uC678. \uC22B\uC790, \uC601\uBB38, \uACF5\uBC31 \uC81C\uC678",
		},
		{
			name:   "Multilingual missing special-character exclusion",
			prompt: "\uD55C\uAE00/\uC601\uC5B4/\uC911\uAD6D\uC5B4/\uC77C\uBCF8\uC5B4/\uB7EC\uC2DC\uC544\uC5B4/\uADF8 \uC678 \uC5B8\uC5B4\uB97C \uC5B8\uC5B4\uBCC4\uB85C 3\uAE00\uC790\uC529 \uCD5C\uB300\uD55C \uB9CE\uC774 \uACF5\uBC31 \uC5C6\uC774 \uC785\uB825\uD558\uC138\uC694. \uCD5C\uB300 15\uAE00\uC790, \uC22B\uC790\uC640 \uACF5\uBC31 \uC81C\uC678",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			answer, matched, err := gmPatrolLocalAnswer(test.prompt)
			if err != nil {
				t.Fatalf("gmPatrolLocalAnswer returned an error: %v", err)
			}
			if matched || answer != "" {
				t.Fatalf("incomplete rules unexpectedly matched locally: matched=%v answer=%q", matched, answer)
			}
		})
	}
}

func TestContainsStandaloneNumber(t *testing.T) {
	tests := []struct {
		text   string
		number string
		want   bool
	}{
		{text: "12\uAE00\uC790", number: "12", want: true},
		{text: "112\uAE00\uC790", number: "12", want: false},
		{text: "13\uAE00\uC790", number: "3", want: false},
		{text: "15\uAE00\uC790", number: "15", want: true},
		{text: "150\uAE00\uC790", number: "15", want: false},
	}

	for _, test := range tests {
		t.Run(test.text, func(t *testing.T) {
			if got := containsStandaloneNumber(test.text, test.number); got != test.want {
				t.Fatalf("containsStandaloneNumber(%q, %q) = %v, want %v", test.text, test.number, got, test.want)
			}
		})
	}
}

func TestDistinctMixedCaseEnglishValidation(t *testing.T) {
	tests := []struct {
		name   string
		answer string
		valid  bool
	}{
		{name: "valid", answer: "QmTzLpRaVfNx", valid: true},
		{name: "ascending sequence", answer: "aBcDeFgHiJkL", valid: false},
		{name: "keyboard row", answer: "QwErTyUiOpAs", valid: false},
		{name: "duplicate ignoring case", answer: "QmTzLpRaVfNq", valid: false},
		{name: "non-letter", answer: "QmTzLpRaVfN9", valid: false},
		{name: "single case", answer: "QMTZLPRAVFNX", valid: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := isValidDistinctMixedCaseEnglish(test.answer); got != test.valid {
				t.Fatalf("isValidDistinctMixedCaseEnglish(%q) = %v, want %v", test.answer, got, test.valid)
			}
		})
	}
}

func TestGenerateDistinctMixedCaseEnglish(t *testing.T) {
	for i := 0; i < generatedAnswerTestRounds(t); i++ {
		answer, err := generateDistinctMixedCaseEnglish()
		if err != nil {
			t.Fatalf("generateDistinctMixedCaseEnglish returned an error: %v", err)
		}
		if !isValidDistinctMixedCaseEnglish(answer) {
			t.Fatalf("generated invalid answer: %q", answer)
		}
		if i < 10 {
			t.Logf("generated answer %d: %s", i+1, answer)
		}
	}
}

func TestGMPatrolLocalAnswerRecognizesDistinctHangulRequest(t *testing.T) {
	prompt := "\uC544\uB798 \uC694\uCCAD \uC0AC\uD56D\uC744 \uD655\uC778 \uD6C4, \uC694\uCCAD\uB41C \uB2F5\uB9CC\uC744 \uC785\uB825\uD558\uC138\uC694. \uC11C\uB85C \uB2E4\uB978 \uD55C\uAE00 12\uAE00\uC790\uB97C \uC784\uC758\uB85C \uC785\uB825\uD574 \uC8FC\uC138\uC694. \uAC19\uC740 \uAE00\uC790 \uBC18\uBCF5, \uAC00\uB098\uB2E4\uB77C \uB4F1 \uC21C\uCC28/\uBC30\uC5F4 \uC785\uB825, \uC790\uC74C \uBAA8\uC74C\uB2E8\uB3C5 \uC785\uB825\uC740 \uC81C\uC678"
	prompt += " \uC22B\uC790, \uC601\uBB38, \uACF5\uBC31, \uD2B9\uC218\uBB38\uC790 \uC81C\uC678"

	answer, matched, err := gmPatrolLocalAnswer(prompt)
	if err != nil {
		t.Fatalf("gmPatrolLocalAnswer returned an error: %v", err)
	}
	if !matched {
		t.Fatal("expected Hangul prompt to be handled locally")
	}
	if !isValidDistinctHangulSyllables(answer) {
		t.Fatalf("generated invalid local Hangul answer: %q", answer)
	}
}

func TestDistinctHangulSyllableValidation(t *testing.T) {
	tests := []struct {
		name   string
		answer string
		valid  bool
	}{
		{name: "valid", answer: "\uAC01\uB023\uB245\uB467\uB689\uB8AB\uBACD\uBCEF\uBF11\uC133\uC355\uC577", valid: true},
		{name: "ga-na-da sequence", answer: "\uAC00\uB098\uB2E4\uB77C\uB9C8\uBC14\uC0AC\uC544\uC790\uCC28\uCE74\uD0C0", valid: false},
		{name: "Korean keyboard row", answer: "\uBC14\uC790\uB2E4\uAC00\uC0AC\uB245\uB467\uB689\uB8AB\uBACD\uBCEF\uBF11", valid: false},
		{name: "consecutive code points", answer: "\uAC00\uAC01\uAC02\uAC03\uB467\uB689\uB8AB\uBACD\uBCEF\uBF11\uC133\uC355", valid: false},
		{name: "duplicate", answer: "\uAC01\uAC01\uB245\uB467\uB689\uB8AB\uBACD\uBCEF\uBF11\uC133\uC355\uC577", valid: false},
		{name: "standalone jamo", answer: "\u3131\u3132\u3133\u3134\u3135\u3136\u3137\u3138\u3139\u313A\u313B\u313C", valid: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := isValidDistinctHangulSyllables(test.answer); got != test.valid {
				t.Fatalf("isValidDistinctHangulSyllables(%q) = %v, want %v", test.answer, got, test.valid)
			}
		})
	}
}

func TestGenerateDistinctHangulSyllables(t *testing.T) {
	for i := 0; i < generatedAnswerTestRounds(t); i++ {
		answer, err := generateDistinctHangulSyllables()
		if err != nil {
			t.Fatalf("generateDistinctHangulSyllables returned an error: %v", err)
		}
		if !isValidDistinctHangulSyllables(answer) {
			t.Fatalf("generated invalid Hangul answer: %q", answer)
		}
		if i < 10 {
			t.Logf("generated Hangul answer %d: %s", i+1, answer)
		}
	}
}

func TestGMPatrolLocalAnswerRecognizesMultilingualCharacterRequest(t *testing.T) {
	prompt := "\uC544\uB798 \uC694\uCCAD \uC0AC\uD56D\uC744 \uD655\uC778 \uD6C4, \uC694\uCCAD\uB41C \uB2F5\uB9CC \uC785\uB825\uD558\uC138\uC694. \uD604\uC7AC \uC785\uB825 \uAC00\uB2A5\uD55C \uBAA8\uB4E0 \uC5B8\uC5B4\uC758 \uBB38\uC790\uB97C \uC5B8\uC5B4\uBCC4\uB85C 3\uAE00\uC790\uC529 \uCD5C\uB300\uD55C \uB9CE\uC774 \uACF5\uBC31 \uC5C6\uC774 \uC785\uB825\uD574 \uC8FC\uC138\uC694. \uD55C\uAE00/\uC601\uC5B4/\uC911\uAD6D\uC5B4/\uC77C\uBCF8\uC5B4/\uB7EC\uC2DC\uC544\uC5B4/\uADF8 \uC678 \uC5B8\uC5B4 \uB4F1. \uCD5C\uB300 15\uAE00\uC790\uAE4C\uC9C0 \uC785\uB825\uC774 \uAC00\uB2A5\uD569\uB2C8\uB2E4. \uC22B\uC790, \uACF5\uBC31, \uD2B9\uC218\uBB38\uC790 \uC81C\uC678"

	answer, matched, err := gmPatrolLocalAnswer(prompt)
	if err != nil {
		t.Fatalf("gmPatrolLocalAnswer returned an error: %v", err)
	}
	if !matched {
		t.Fatal("expected multilingual prompt to be handled locally")
	}
	if !isValidMultilingualCharacters(answer) {
		t.Fatalf("generated invalid multilingual answer: %q", answer)
	}
}

func TestMultilingualCharacterValidation(t *testing.T) {
	tests := []struct {
		name   string
		answer string
		valid  bool
	}{
		{name: "valid", answer: "\uBD04\uB2EC\uBE5BQmT", valid: true},
		{name: "too short", answer: "\uBD04\uB2EC\uBE5BQm", valid: false},
		{name: "wrong language count", answer: "\uBD04\uB2EC\uBE5B\uC0B0Qm", valid: false},
		{name: "duplicate", answer: "\uBD04\uBD04\uBE5BQmT", valid: false},
		{name: "space", answer: "\uBD04\uB2EC\uBE5B Qm", valid: false},
		{name: "digit", answer: "\uBD04\uB2EC\uBE5BQm1", valid: false},
		{name: "Chinese", answer: "\uBD04\uB2EC\uBE5BQm\u4E2D", valid: false},
		{name: "Japanese", answer: "\uBD04\uB2EC\uBE5BQm\u3042", valid: false},
		{name: "Cyrillic", answer: "\uBD04\uB2EC\uBE5BQm\u0416", valid: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := isValidMultilingualCharacters(test.answer); got != test.valid {
				t.Fatalf("isValidMultilingualCharacters(%q) = %v, want %v", test.answer, got, test.valid)
			}
		})
	}
}

func TestGenerateMultilingualCharacters(t *testing.T) {
	for i := 0; i < generatedAnswerTestRounds(t); i++ {
		answer, err := generateMultilingualCharacters()
		if err != nil {
			t.Fatalf("generateMultilingualCharacters returned an error: %v", err)
		}
		if !isValidMultilingualCharacters(answer) {
			t.Fatalf("generated invalid multilingual answer: %q", answer)
		}
		if i < 10 {
			t.Logf("generated multilingual answer %d: %s", i+1, answer)
		}
	}
}
