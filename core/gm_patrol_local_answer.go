package core

import (
	"fmt"
	"math/rand"
	"strings"
	"unicode/utf8"
)

const (
	gmPatrolDistinctEnglishLength = 12
	gmPatrolDistinctHangulLength  = 12
	gmPatrolCharactersPerLanguage = 3
	gmPatrolMultilingualLanguages = 2
	gmPatrolMultilingualLength    = gmPatrolCharactersPerLanguage * gmPatrolMultilingualLanguages
	gmPatrolLocalAnswerAttempts   = 128
	gmPatrolForbiddenRunLength    = 4
	gmPatrolHangulSyllableFirst   = '\uAC00'
	gmPatrolHangulSyllableLast    = '\uD7A3'
)

// gmPatrolMultilingualCharacterPools 仅使用正常玩家最常输入的韩文和英文。
// 15 字符只是上限，本地答案按题目要求每种语言输出 3 个字符，共 6 个字符。
var gmPatrolMultilingualCharacterPools = [][]rune{
	{'\uBD04', '\uB2EC', '\uBE5B', '\uC0B0', '\uBC14', '\uB2E4'},
	{'Q', 'm', 'T', 'r', 'V', 'x'},
}

// gmPatrolLocalAnswer 仅处理规则完整命中的固定格式随机题。
// 任一关键题干或限制缺失时返回未命中，保留 AI 作答路径，避免错误套用本地规则。
func gmPatrolLocalAnswer(ocrText string) (string, bool, error) {
	if isMultilingualCharactersRequest(ocrText) {
		answer, err := generateMultilingualCharacters()
		return answer, true, err
	}
	if isDistinctMixedCaseEnglishRequest(ocrText) {
		answer, err := generateDistinctMixedCaseEnglish()
		return answer, true, err
	}
	if isDistinctHangulSyllableRequest(ocrText) {
		answer, err := generateDistinctHangulSyllables()
		return answer, true, err
	}
	return "", false, nil
}

func isMultilingualCharactersRequest(ocrText string) bool {
	hasKorean := strings.Contains(ocrText, "\uD55C\uAE00")
	hasEnglish := strings.Contains(ocrText, "\uC601\uC5B4")
	hasChinese := strings.Contains(ocrText, "\uC911\uAD6D\uC5B4")
	hasJapanese := strings.Contains(ocrText, "\uC77C\uBCF8\uC5B4")
	hasRussian := strings.Contains(ocrText, "\uB7EC\uC2DC\uC544")
	hasPerLanguageCount := containsStandaloneNumber(ocrText, "3")
	hasMaximumLength := containsStandaloneNumber(ocrText, "15")

	return hasKorean && hasEnglish && hasChinese && hasJapanese && hasRussian && hasPerLanguageCount && hasMaximumLength &&
		hasCompleteMultilingualCharacterRules(ocrText)
}

func hasCompleteMultilingualCharacterRules(ocrText string) bool {
	return containsAll(ocrText,
		"\uBB38\uC790", "\uC785\uB825", "\uC5B8\uC5B4\uBCC4", "3\uAE00\uC790", "\uCD5C\uB300\uD55C", "\uACF5\uBC31", "\uC5C6\uC774", "\uCD5C\uB300",
		"\uC22B\uC790", "\uD2B9\uC218\uBB38\uC790", "\uADF8 \uC678 \uC5B8\uC5B4")
}

func generateMultilingualCharacters() (string, error) {
	for attempt := 0; attempt < gmPatrolLocalAnswerAttempts; attempt++ {
		answer := make([]rune, 0, gmPatrolMultilingualLength)
		for _, pool := range gmPatrolMultilingualCharacterPools {
			for _, index := range rand.Perm(len(pool))[:gmPatrolCharactersPerLanguage] {
				answer = append(answer, pool[index])
			}
		}
		if isValidMultilingualCharacters(string(answer)) {
			return string(answer), nil
		}
	}

	return "", fmt.Errorf("could not generate a valid multilingual answer")
}

func isDistinctMixedCaseEnglishRequest(ocrText string) bool {
	text := strings.ToLower(ocrText)
	hasUppercase := strings.Contains(text, "대문자") || strings.Contains(text, "uppercase")
	hasLowercase := strings.Contains(text, "소문자") || strings.Contains(text, "lowercase")
	hasEnglish := strings.Contains(text, "영문") || strings.Contains(text, "english")
	hasDistinctRequirement := strings.Contains(text, "서로 다른") ||
		strings.Contains(text, "같은 글자 반복") ||
		strings.Contains(text, "중복") ||
		strings.Contains(text, "unique") ||
		strings.Contains(text, "different")

	return containsStandaloneNumber(text, "12") && hasUppercase && hasLowercase && hasEnglish && hasDistinctRequirement &&
		hasCompleteMixedCaseEnglishRules(text)
}

func hasCompleteMixedCaseEnglishRules(text string) bool {
	return containsAll(text,
		"\uC784\uC758", "\uC785\uB825", "\uAC19\uC740 \uAE00\uC790 \uBC18\uBCF5", "\uC22B\uC790", "\uD55C\uAE00", "\uACF5\uBC31", "\uD2B9\uC218\uBB38\uC790") &&
		containsAny(text, "\uC21C\uCC28", "\uBC30\uC5F4") &&
		containsAny(text, "abcd", "qwer")
}

func generateDistinctMixedCaseEnglish() (string, error) {
	for attempt := 0; attempt < gmPatrolLocalAnswerAttempts; attempt++ {
		letterIndexes := rand.Perm(26)[:gmPatrolDistinctEnglishLength]
		uppercasePositions := make([]bool, gmPatrolDistinctEnglishLength)
		uppercaseCount := rand.Intn(gmPatrolDistinctEnglishLength-1) + 1
		for _, position := range rand.Perm(gmPatrolDistinctEnglishLength)[:uppercaseCount] {
			uppercasePositions[position] = true
		}

		answer := make([]byte, gmPatrolDistinctEnglishLength)
		for i, index := range letterIndexes {
			letter := byte('a' + index)
			if uppercasePositions[i] {
				letter -= 'a' - 'A'
			}
			answer[i] = letter
		}
		if isValidDistinctMixedCaseEnglish(string(answer)) {
			return string(answer), nil
		}
	}

	return "", fmt.Errorf("could not generate a valid distinct mixed-case English answer")
}

func isDistinctHangulSyllableRequest(ocrText string) bool {
	hasHangul := strings.Contains(ocrText, "\uD55C\uAE00")
	hasDistinctRequirement := strings.Contains(ocrText, "\uC11C\uB85C \uB2E4\uB978") ||
		strings.Contains(ocrText, "\uAC19\uC740 \uAE00\uC790 \uBC18\uBCF5") ||
		strings.Contains(ocrText, "\uC911\uBCF5") ||
		strings.Contains(strings.ToLower(ocrText), "unique") ||
		strings.Contains(strings.ToLower(ocrText), "different")

	return containsStandaloneNumber(ocrText, "12") && hasHangul && hasDistinctRequirement &&
		hasCompleteHangulSyllableRules(ocrText)
}

func hasCompleteHangulSyllableRules(ocrText string) bool {
	return containsAll(ocrText,
		"\uC784\uC758", "\uC785\uB825", "\uAC19\uC740 \uAE00\uC790 \uBC18\uBCF5", "\uAC00\uB098\uB2E4", "\uC790\uC74C", "\uBAA8\uC74C", "\uB2E8\uB3C5",
		"\uC22B\uC790", "\uC601\uBB38", "\uACF5\uBC31", "\uD2B9\uC218\uBB38\uC790") &&
		containsAny(ocrText, "\uC21C\uCC28", "\uBC30\uC5F4")
}

func containsAll(text string, terms ...string) bool {
	for _, term := range terms {
		if !strings.Contains(text, term) {
			return false
		}
	}
	return true
}

func containsAny(text string, terms ...string) bool {
	for _, term := range terms {
		if strings.Contains(text, term) {
			return true
		}
	}
	return false
}

func containsStandaloneNumber(text, number string) bool {
	for offset := 0; ; {
		index := strings.Index(text[offset:], number)
		if index < 0 {
			return false
		}
		index += offset
		end := index + len(number)
		if (index == 0 || !isASCIIDigit(text[index-1])) && (end == len(text) || !isASCIIDigit(text[end])) {
			return true
		}
		offset = end
	}
}

func isASCIIDigit(value byte) bool {
	return value >= '0' && value <= '9'
}

func generateDistinctHangulSyllables() (string, error) {
	for attempt := 0; attempt < gmPatrolLocalAnswerAttempts; attempt++ {
		answer := make([]rune, gmPatrolDistinctHangulLength)
		seen := make(map[rune]struct{}, gmPatrolDistinctHangulLength)
		for i := range answer {
			for {
				syllable := gmPatrolHangulSyllableFirst + rune(rand.Intn(int(gmPatrolHangulSyllableLast-gmPatrolHangulSyllableFirst+1)))
				if _, exists := seen[syllable]; exists {
					continue
				}
				seen[syllable] = struct{}{}
				answer[i] = syllable
				break
			}
		}
		if isValidDistinctHangulSyllables(string(answer)) {
			return string(answer), nil
		}
	}

	return "", fmt.Errorf("could not generate valid distinct Hangul syllables")
}

func isValidDistinctMixedCaseEnglish(answer string) bool {
	if len(answer) != gmPatrolDistinctEnglishLength {
		return false
	}

	var seen [26]bool
	uppercaseCount := 0
	lowercaseCount := 0
	for i := 0; i < len(answer); i++ {
		letter := answer[i]
		var index byte
		switch {
		case letter >= 'A' && letter <= 'Z':
			uppercaseCount++
			index = letter - 'A'
		case letter >= 'a' && letter <= 'z':
			lowercaseCount++
			index = letter - 'a'
		default:
			return false
		}
		if seen[index] {
			return false
		}
		seen[index] = true
	}

	return uppercaseCount > 0 && lowercaseCount > 0 && !hasForbiddenLetterRun(answer)
}

func isValidDistinctHangulSyllables(answer string) bool {
	if utf8.RuneCountInString(answer) != gmPatrolDistinctHangulLength {
		return false
	}

	seen := make(map[rune]struct{}, gmPatrolDistinctHangulLength)
	for _, syllable := range answer {
		if syllable < gmPatrolHangulSyllableFirst || syllable > gmPatrolHangulSyllableLast {
			return false
		}
		if _, exists := seen[syllable]; exists {
			return false
		}
		seen[syllable] = struct{}{}
	}

	return !hasForbiddenHangulRun(answer)
}

func isValidMultilingualCharacters(answer string) bool {
	if utf8.RuneCountInString(answer) != gmPatrolMultilingualLength {
		return false
	}

	var languageCounts [gmPatrolMultilingualLanguages]int
	seen := make(map[rune]struct{}, gmPatrolMultilingualLength)
	for _, character := range answer {
		if _, exists := seen[character]; exists {
			return false
		}
		seen[character] = struct{}{}

		switch {
		case character >= gmPatrolHangulSyllableFirst && character <= gmPatrolHangulSyllableLast:
			languageCounts[0]++
		case (character >= 'A' && character <= 'Z') || (character >= 'a' && character <= 'z'):
			languageCounts[1]++
		default:
			return false
		}
	}

	for _, count := range languageCounts {
		if count != gmPatrolCharactersPerLanguage {
			return false
		}
	}
	return true
}

func hasForbiddenLetterRun(answer string) bool {
	lowercase := strings.ToLower(answer)
	if hasAlphabeticalRun(lowercase) {
		return true
	}

	for _, keyboardRow := range []string{"qwertyuiop", "asdfghjkl", "zxcvbnm"} {
		for start := 0; start <= len(keyboardRow)-gmPatrolForbiddenRunLength; start++ {
			run := keyboardRow[start : start+gmPatrolForbiddenRunLength]
			if strings.Contains(lowercase, run) || strings.Contains(lowercase, reverseASCII(run)) {
				return true
			}
		}
	}
	return false
}

func hasAlphabeticalRun(answer string) bool {
	runLength := 1
	direction := 0
	for i := 1; i < len(answer); i++ {
		difference := int(answer[i]) - int(answer[i-1])
		if difference != 1 && difference != -1 {
			runLength = 1
			direction = 0
			continue
		}
		if difference == direction {
			runLength++
		} else {
			direction = difference
			runLength = 2
		}
		if runLength >= gmPatrolForbiddenRunLength {
			return true
		}
	}
	return false
}

func reverseASCII(s string) string {
	result := make([]byte, len(s))
	for i := range s {
		result[len(s)-1-i] = s[i]
	}
	return string(result)
}

func hasForbiddenHangulRun(answer string) bool {
	if hasConsecutiveHangulCodePoints(answer) {
		return true
	}

	for _, sequence := range []string{
		"\uAC00\uB098\uB2E4\uB77C\uB9C8\uBC14\uC0AC\uC544\uC790\uCC28\uCE74\uD0C0\uD30C\uD558",
		"\uBC14\uC790\uB2E4\uAC00\uC0AC",
		"\uB9C8\uB098\uC544\uB77C\uD558",
		"\uCE74\uD0C0\uCC28\uD30C",
	} {
		if containsForbiddenHangulSequence(answer, sequence) {
			return true
		}
	}
	return false
}

func hasConsecutiveHangulCodePoints(answer string) bool {
	runes := []rune(answer)
	runLength := 1
	direction := rune(0)
	for i := 1; i < len(runes); i++ {
		difference := runes[i] - runes[i-1]
		if difference != 1 && difference != -1 {
			runLength = 1
			direction = 0
			continue
		}
		if difference == direction {
			runLength++
		} else {
			direction = difference
			runLength = 2
		}
		if runLength >= gmPatrolForbiddenRunLength {
			return true
		}
	}
	return false
}

func containsForbiddenHangulSequence(answer, sequence string) bool {
	runes := []rune(sequence)
	for start := 0; start <= len(runes)-gmPatrolForbiddenRunLength; start++ {
		run := string(runes[start : start+gmPatrolForbiddenRunLength])
		if strings.Contains(answer, run) || strings.Contains(answer, reverseRunes(run)) {
			return true
		}
	}
	return false
}

func reverseRunes(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}
