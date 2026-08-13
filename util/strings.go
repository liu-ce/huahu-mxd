package util

import (
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func hasParenthesesPair(text string) bool {
	// 匹配 () 或 （）
	re := regexp.MustCompile(`(\(|（).*(\)|）)`)
	return re.MatchString(text)
}

// CalculateTextSimilarity 计算两个文本的相似度，使用编辑距离算法
func CalculateTextSimilarity(text1, text2 string) float32 {
	// 去除空格和转换为小写，提高匹配准确性
	s1 := strings.ReplaceAll(strings.ToLower(text1), " ", "")
	s2 := strings.ReplaceAll(strings.ToLower(text2), " ", "")

	if s1 == s2 {
		return 1.0
	}

	if len(s1) == 0 || len(s2) == 0 {
		return 0.0
	}

	// 使用编辑距离算法计算相似度
	distance := levenshteinDistance(s1, s2)
	maxLen := max(len(s1), len(s2))

	similarity := 1.0 - float32(distance)/float32(maxLen)
	if similarity < 0 {
		similarity = 0
	}

	return similarity
}

// levenshteinDistance 计算两个字符串的编辑距离
func levenshteinDistance(s1, s2 string) int {
	len1, len2 := len(s1), len(s2)
	if len1 == 0 {
		return len2
	}
	if len2 == 0 {
		return len1
	}

	// 创建距离矩阵
	matrix := make([][]int, len1+1)
	for i := range matrix {
		matrix[i] = make([]int, len2+1)
	}

	// 初始化第一行和第一列
	for i := 0; i <= len1; i++ {
		matrix[i][0] = i
	}
	for j := 0; j <= len2; j++ {
		matrix[0][j] = j
	}

	// 填充矩阵
	for i := 1; i <= len1; i++ {
		for j := 1; j <= len2; j++ {
			cost := 0
			if s1[i-1] != s2[j-1] {
				cost = 1
			}

			matrix[i][j] = min(
				min(matrix[i-1][j]+1, matrix[i][j-1]+1), // 删除和插入的最小值
				matrix[i-1][j-1]+cost,                   // 替换
			)
		}
	}

	return matrix[len1][len2]
}

// min 返回两个整数中的较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// max 返回两个整数中的较大值
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// ExtractRightNumber 检查文本是否存在数字/数字结构，如果存在返回右边的数字，否则返回-1
func ExtractRightNumber(text string) int {
	// 移除空格
	text = strings.ReplaceAll(text, " ", "")

	// 匹配各种数字结构模式：x/y, x-y, x:y, x|y, x,y 等
	patterns := []string{
		`(\d+)[/\-:|,](\d+)`, // 匹配 x/y, x-y, x:y, x|y, x,y
		`(\d+).*?(\d+)`,      // 匹配两个数字，中间可能有其他字符
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(text)
		if len(matches) >= 3 {
			// 返回第二个数字（右边的数字）
			rightNum, err := strconv.Atoi(matches[2])
			if err == nil {
				return rightNum
			}
		}
	}

	// 如果没有找到数字结构，检查是否只有一个数字
	re := regexp.MustCompile(`\d+`)
	matches := re.FindAllString(text, -1)
	if len(matches) == 1 {
		// 只有一个数字，返回这个数字
		num, err := strconv.Atoi(matches[0])
		if err == nil {
			return num
		}
	}

	// 没有找到数字或数字结构
	return -1
}

// IsNumberEqual 判断OCR识别的字符串数值是否等于目标数值
// 会去除所有非数字字符（如逗号、空格等），只保留数字进行比较
// text: OCR识别的文本
// target: 目标数值
// 返回: true表示数值相等，false表示不相等
func IsNumberEqual(text string, target int) bool {
	// 去除所有非数字字符，只保留数字
	var numStr strings.Builder
	for _, char := range text {
		if char >= '0' && char <= '9' {
			numStr.WriteRune(char)
		}
	}

	// 如果没有数字，返回false
	cleanText := numStr.String()
	if cleanText == "" {
		return false
	}

	// 转换为整数并比较
	num := 0
	for _, char := range cleanText {
		if char >= '0' && char <= '9' {
			num = num*10 + int(char-'0')
		}
	}
	return num == target
}

// ExtractNumberFromText 从文本中提取纯数字
// 会去除所有非数字字符，只保留数字并转换为整数
// 例如: "161.270" -> 161270, "50,000" -> 50000, "$1,234.56" -> 123456
func ExtractNumberFromText(text string) int {
	var numStr strings.Builder
	for _, char := range text {
		if char >= '0' && char <= '9' {
			numStr.WriteRune(char)
		}
	}

	cleanText := numStr.String()
	if cleanText == "" {
		return 0
	}

	// 手动转换字符串为整数
	num := 0
	for _, char := range cleanText {
		if char >= '0' && char <= '9' {
			num = num*10 + int(char-'0')
		}
	}
	return num
}

// FilterNumbersAndSlash 过滤文本，只保留数字和斜杠
// 例如: "1abc/5def" -> "1/5", "3人/10人" -> "3/10"
func FilterNumbersAndSlash(text string) string {
	var result strings.Builder
	for _, char := range text {
		if (char >= '0' && char <= '9') || char == '/' {
			result.WriteRune(char)
		}
	}
	return result.String()
}

// Game name generator vocabulary - optimized for 6-10 character names
var (
	// Random prefix generator - generates 3-5 character random English strings
	// This will be generated dynamically instead of using a fixed array

	// Core words (300+ words) - optimized for 6-10 character names (short words only)
	nameCores = []string{
		// Animals
		"Dragon", "Tiger", "Wolf", "Eagle", "Lion", "Bear", "Deer", "Fox", "Cat", "Dog",
		"Snake", "Hawk", "Falcon", "Owl", "Raven", "Crow", "Bat", "Rat", "Mouse", "Rabbit",
		"Bull", "Ox", "Ram", "Goat", "Sheep", "Pig", "Boar", "Elk", "Moose", "Antelope",
		"Shark", "Whale", "Dolphin", "Seal", "Penguin", "Swan", "Duck", "Goose", "Crane", "Heron",
		"Spider", "Scorpion", "Wasp", "Bee", "Ant", "Beetle", "Moth", "Butterfly", "Dragonfly", "Cricket",

		// Weapons & Tools
		"Sword", "Blade", "Spear", "Bow", "Shield", "Armor", "Helm", "Robe", "Staff", "Ring",
		"Axe", "Hammer", "Mace", "Club", "Dagger", "Knife", "Arrow", "Bolt", "Crossbow", "Sling",
		"Whip", "Chain", "Rope", "Net", "Hook", "Claw", "Fang", "Tusk", "Horn", "Spike",
		"Gun", "Rifle", "Pistol", "Cannon", "Bomb", "Grenade", "Mine", "Trap", "Cage", "Lock",

		// Fantasy & Magic
		"Heart", "Soul", "Spirit", "God", "Demon", "Fairy", "Saint", "King", "Lord", "Duke",
		"Angel", "Devil", "Ghost", "Witch", "Wizard", "Mage", "Priest", "Monk", "Nun", "Prophet",
		"Oracle", "Seer", "Shaman", "Druid", "Paladin", "Cleric", "Bard", "Rogue", "Assassin", "Thief",
		"Magic", "Spell", "Curse", "Blessing", "Charm", "Hex", "Rune", "Sigil", "Symbol", "Mark",

		// Nature & Elements
		"Star", "Moon", "Sun", "Cloud", "Wind", "Rain", "Snow", "Fire", "Ice", "Storm",
		"Thunder", "Lightning", "Frost", "Flame", "Spark", "Ember", "Ash", "Smoke", "Mist", "Fog",
		"Sea", "River", "Lake", "Wood", "Tree", "Stone", "Rock", "Hill", "Cave", "Gate",
		"Mountain", "Valley", "Desert", "Forest", "Jungle", "Swamp", "Marsh", "Meadow", "Field", "Garden",
		"Flower", "Rose", "Lily", "Tulip", "Daisy", "Violet", "Orchid", "Lotus", "Petal", "Thorn",

		// Materials & Gems
		"Gold", "Silver", "Iron", "Jade", "Pearl", "Gem", "Crystal", "Metal", "Steel", "Bronze",
		"Copper", "Tin", "Lead", "Zinc", "Nickel", "Platinum", "Titanium", "Aluminum", "Mercury", "Cobalt",
		"Diamond", "Ruby", "Sapphire", "Emerald", "Topaz", "Amethyst", "Garnet", "Opal", "Turquoise", "Coral",
		"Marble", "Granite", "Quartz", "Obsidian", "Basalt", "Limestone", "Sandstone", "Clay", "Mud", "Sand",

		// Colors & Appearance
		"Light", "Shadow", "Dark", "Bright", "White", "Black", "Red", "Blue", "Green", "Purple",
		"Yellow", "Orange", "Pink", "Brown", "Gray", "Silver", "Gold", "Bronze", "Copper", "Crimson",
		"Azure", "Violet", "Indigo", "Turquoise", "Magenta", "Cyan", "Lime", "Olive", "Navy", "Maroon",
		"Shiny", "Dull", "Glossy", "Matte", "Rough", "Smooth", "Sharp", "Blunt", "Curved", "Straight",

		// Body & Anatomy
		"Sky", "Earth", "Human", "Ghost", "Beast", "Mind", "Body", "Hand", "Eye", "Face",
		"Head", "Foot", "Arm", "Leg", "Chest", "Back", "Neck", "Shoulder", "Elbow", "Knee",
		"Finger", "Toe", "Thumb", "Palm", "Heel", "Ankle", "Wrist", "Hip", "Waist", "Belly",
		"Hair", "Beard", "Mustache", "Eyebrow", "Eyelash", "Cheek", "Chin", "Forehead", "Temple", "Jaw",

		// Emotions & Feelings
		"War", "Fight", "Battle", "Life", "Death", "Love", "Hate", "Joy", "Pain", "Hope",
		"Fear", "Anger", "Rage", "Fury", "Wrath", "Pride", "Shame", "Guilt", "Regret", "Sorrow",
		"Grief", "Melancholy", "Despair", "Desperation", "Loneliness", "Isolation", "Solitude", "Peace", "Calm", "Serenity",
		"Excitement", "Thrill", "Adventure", "Wonder", "Awe", "Amazement", "Surprise", "Shock", "Stun", "Daze",

		// Concepts & Ideas
		"Dream", "Wish", "Faith", "Trust", "Truth", "Law", "Rule", "Order", "Peace", "War",
		"Justice", "Mercy", "Grace", "Honor", "Glory", "Victory", "Defeat", "Triumph", "Success", "Failure",
		"Freedom", "Liberty", "Independence", "Autonomy", "Sovereignty", "Authority", "Power", "Control", "Command", "Leadership",
		"Wisdom", "Knowledge", "Understanding", "Insight", "Intuition", "Instinct", "Memory", "Remembrance", "Forgetting", "Oblivion",

		// Forces & Energy
		"Force", "Power", "Strength", "Might", "Aura", "Breath", "Voice", "Sound", "Echo", "Song",
		"Energy", "Vitality", "Vigor", "Stamina", "Endurance", "Resilience", "Toughness", "Hardiness", "Robustness", "Sturdiness",
		"Speed", "Velocity", "Momentum", "Acceleration", "Deceleration", "Rhythm", "Beat", "Pulse", "Throb", "Vibration",
		"Wave", "Ripple", "Surge", "Rush", "Flow", "Stream", "Current", "Tide", "Flood", "Torrent",

		// Objects & Items
		"Way", "Path", "Road", "Door", "Key", "Lock", "Chain", "Rope", "Wire", "Line",
		"Book", "Scroll", "Parchment", "Paper", "Ink", "Pen", "Quill", "Brush", "Canvas", "Paint",
		"Crown", "Throne", "Scepter", "Orb", "Globe", "Sphere", "Cube", "Pyramid", "Triangle", "Circle",
		"Box", "Chest", "Trunk", "Bag", "Pouch", "Sack", "Basket", "Vessel", "Container", "Receptacle",

		// Qualities & States
		"Real", "True", "False", "Good", "Evil", "Pure", "Clean", "Dirty", "Fresh", "Old",
		"New", "Young", "Ancient", "Modern", "Contemporary", "Current", "Present", "Past", "Future", "Eternal",
		"Perfect", "Flawless", "Flawed", "Broken", "Damaged", "Repaired", "Fixed", "Mended", "Healed", "Cured",
		"Complete", "Incomplete", "Partial", "Whole", "Entire", "Total", "Full", "Empty", "Void", "Null",

		// Sizes & Dimensions
		"Big", "Small", "High", "Low", "Deep", "Wide", "Long", "Short", "Fast", "Slow",
		"Tall", "Short", "Thick", "Thin", "Fat", "Slim", "Narrow", "Broad", "Wide", "Tight",
		"Loose", "Firm", "Solid", "Liquid", "Gas", "Vapor", "Steam", "Smoke", "Dust", "Particle",
		"Giant", "Tiny", "Huge", "Massive", "Enormous", "Minute", "Microscopic", "Invisible", "Hidden", "Visible",

		// Temperatures & Conditions
		"Hot", "Cold", "Warm", "Cool", "Dry", "Wet", "Soft", "Hard", "Smooth", "Rough",
		"Sharp", "Blunt", "Pointed", "Rounded", "Square", "Round", "Oval", "Rectangular", "Triangular", "Hexagonal",
		"Frozen", "Melted", "Solidified", "Liquefied", "Evaporated", "Condensed", "Crystallized", "Dissolved", "Mixed", "Separated",
		"Burning", "Freezing", "Boiling", "Steaming", "Smoking", "Glowing", "Shining", "Dull", "Bright", "Dim",

		// Directions & Positions
		"Top", "Bottom", "Left", "Right", "Front", "Back", "Side", "Edge", "Corner", "Center",
		"Middle", "Interior", "Exterior", "Inside", "Outside", "Surface", "Depth", "Height", "Width", "Length",
		"North", "South", "East", "West", "Up", "Down", "Above", "Below", "Over", "Under",
		"Near", "Far", "Close", "Distant", "Remote", "Adjacent", "Neighboring", "Surrounding", "Central", "Peripheral",

		// Actions & Movements
		"Start", "End", "Open", "Close", "Rise", "Fall", "Up", "Down", "In", "Out",
		"Come", "Go", "Run", "Walk", "Jump", "Fly", "Swim", "Climb", "Crawl", "Stand",
		"Move", "Stop", "Pause", "Rest", "Sleep", "Wake", "Dream", "Think", "Feel", "Sense",
		"Touch", "Hold", "Grab", "Release", "Push", "Pull", "Lift", "Drop", "Throw", "Catch",

		// Creation & Destruction
		"Break", "Build", "Make", "Form", "Shape", "Size", "Weight", "Height", "Width", "Depth",
		"Create", "Destroy", "Generate", "Produce", "Manufacture", "Craft", "Forge", "Mold", "Carve", "Sculpt",
		"Assemble", "Disassemble", "Construct", "Deconstruct", "Erect", "Demolish", "Raise", "Lower", "Lift", "Drop",
		"Grow", "Shrink", "Expand", "Contract", "Stretch", "Compress", "Extend", "Retract", "Lengthen", "Shorten",
	}

	// Number suffixes (50 words) - optimized for 6-10 character names
	nameNumbers = []string{
		"1", "2", "3", "4", "5", "6", "7", "8", "9", "0",
		"01", "02", "03", "04", "05", "06", "07", "08", "09", "10",
		"11", "12", "13", "14", "15", "16", "17", "18", "19", "20",
		"21", "22", "23", "24", "25", "26", "27", "28", "29", "30",
		"31", "32", "33", "34", "35", "36", "37", "38", "39", "40",
	}
)

// GenerateRandomGameName generates random game names optimized for 6-10 characters
// Supports multiple combination patterns with tens of thousands of possible combinations
// pattern: combination pattern (1-5)
//
//	1: prefix+core+number (e.g.: AbcDragon23)
//	2: core+number (e.g.: Dragon23)
//	3: prefix+core+number (e.g.: XyZSword1)
//	4: core+core (e.g.: DragonTiger)
//	5: random combination pattern
func GenerateRandomGameName(pattern int) string {
	rand.Seed(time.Now().UnixNano())

	for attempts := 0; attempts < 100; attempts++ {
		var name string

		switch pattern {
		case 1:
			// prefix+core+number (3-5 + 3-6 + 1-2 = 7-13 chars)
			name = generateRandomPrefix() +
				nameCores[rand.Intn(len(nameCores))] +
				nameNumbers[rand.Intn(len(nameNumbers))]

		case 2:
			// core+number (3-6 + 1-2 = 4-8 chars)
			name = nameCores[rand.Intn(len(nameCores))] +
				nameNumbers[rand.Intn(len(nameNumbers))]

		case 3:
			// prefix+core+number (same as case 1)
			name = generateRandomPrefix() +
				nameCores[rand.Intn(len(nameCores))] +
				nameNumbers[rand.Intn(len(nameNumbers))]

		case 4:
			// core+core (3-6 + 3-6 = 6-12 chars)
			name = nameCores[rand.Intn(len(nameCores))] +
				nameCores[rand.Intn(len(nameCores))]

		case 5:
			// random combination pattern
			patterns := []int{1, 2, 3, 4}
			return GenerateRandomGameName(patterns[rand.Intn(len(patterns))])

		default:
			// default to pattern 2
			return GenerateRandomGameName(2)
		}

		// Check if name length is within 6-10 characters
		if len(name) >= 6 && len(name) <= 10 {
			return name
		}
	}

	// If 100 attempts don't find a suitable length name, return a default name
	return GenerateRandomGameName(2)
}

// GenerateRandomGameNameWithLength generates random game names with specified length
// minLength: minimum length
// maxLength: maximum length
func GenerateRandomGameNameWithLength(minLength, maxLength int) string {
	rand.Seed(time.Now().UnixNano())

	for attempts := 0; attempts < 100; attempts++ {
		name := GenerateRandomGameName(6) // use random pattern
		if len(name) >= minLength && len(name) <= maxLength {
			return name
		}
	}

	// if 100 attempts don't find a suitable length name, return a default name
	return GenerateRandomGameName(1)
}

// GenerateRandomGameNameWithNumbers generates random game names with numbers (6-10 characters)
// This function ensures the generated names contain both letters and numbers
func GenerateRandomGameNameWithNumbers() string {
	rand.Seed(time.Now().UnixNano())

	for attempts := 0; attempts < 100; attempts++ {
		var name string

		// Use pattern 4 (prefix+core+number) as base, but with more variations
		patterns := []int{4, 4, 4, 4, 4} // Favor pattern 4 for numbers
		pattern := patterns[rand.Intn(len(patterns))]

		switch pattern {
		case 4:
			// prefix+core+number suffix
			name = generateRandomPrefix() +
				nameCores[rand.Intn(len(nameCores))] +
				nameNumbers[rand.Intn(len(nameNumbers))]
		default:
			// Fallback to pattern 4
			name = generateRandomPrefix() +
				nameCores[rand.Intn(len(nameCores))] +
				nameNumbers[rand.Intn(len(nameNumbers))]
		}

		// Check if name length is within 6-10 characters and contains numbers
		if len(name) >= 6 && len(name) <= 10 && containsNumber(name) {
			return name
		}
	}

	// If 100 attempts don't find a suitable name, return a default name with numbers
	return generateRandomPrefix() + nameCores[rand.Intn(len(nameCores))] + "1"
}

// containsNumber checks if a string contains at least one digit
func containsNumber(s string) bool {
	for _, char := range s {
		if char >= '0' && char <= '9' {
			return true
		}
	}
	return false
}

// generateRandomPrefix generates a random 3-5 character lowercase English string
func generateRandomPrefix() string {
	rand.Seed(time.Now().UnixNano())

	// Random length between 3-5 characters
	length := rand.Intn(3) + 3 // 3, 4, or 5

	// English letters (lowercase only)
	letters := "abcdefghijklmnopqrstuvwxyz"

	var prefix strings.Builder
	for i := 0; i < length; i++ {
		prefix.WriteByte(letters[rand.Intn(len(letters))])
	}

	return prefix.String()
}

// RemoveNearbyValues removes nearby values (within 5 difference) and keeps only one, then sorts
// input: slice of integers
// threshold: maximum difference to consider as "nearby" (default: 5)
// returns: sorted slice with nearby values removed
func RemoveNearbyValues(values []int, threshold int) []int {
	if len(values) == 0 {
		return values
	}

	if threshold <= 0 {
		threshold = 5
	}

	// Sort the input slice first
	sorted := make([]int, len(values))
	copy(sorted, values)

	// Simple bubble sort
	for i := 0; i < len(sorted)-1; i++ {
		for j := 0; j < len(sorted)-i-1; j++ {
			if sorted[j] > sorted[j+1] {
				sorted[j], sorted[j+1] = sorted[j+1], sorted[j]
			}
		}
	}

	// Remove nearby values
	result := []int{}
	used := make([]bool, len(sorted))

	for i := 0; i < len(sorted); i++ {
		if used[i] {
			continue
		}

		// Add current value to result
		result = append(result, sorted[i])
		used[i] = true

		// Mark nearby values as used
		for j := i + 1; j < len(sorted); j++ {
			if !used[j] && abs(sorted[j]-sorted[i]) <= threshold {
				used[j] = true
			}
		}
	}

	return result
}

// abs returns the absolute value of an integer
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// ExtractFirstDigitInRange 从文本中提取从左到右第一个数字，如果在指定范围内返回该数字，否则返回0
// text: 要提取数字的文本
// min: 最小值（包含）
// max: 最大值（包含）
// 返回: 如果在范围内返回数字，否则返回0
func ExtractFirstDigitInRange(text string, min, max int) int {
	re := regexp.MustCompile(`\d`)
	firstDigitStr := re.FindString(text)
	if firstDigitStr == "" {
		return 0
	}

	digit, err := strconv.Atoi(firstDigitStr)
	if err != nil {
		return 0
	}

	if digit >= min && digit <= max {
		return digit
	}

	return 0
}

// ExtractChannelNumber 从频道文本中提取频道号
// text: OCR识别的频道文本（如"Ch7"、"ch3"等）
// 返回: (频道号, 是否成功提取)
// 如果文本不包含"Ch"或"ch"，或者无法提取数字，返回 (0, false)
func ExtractChannelNumber(text string) (int, bool) {
	// 检查是否包含频道标识
	if !strings.Contains(text, "Ch") && !strings.Contains(text, "ch") {
		return 0, false
	}

	// 提取数字
	re := regexp.MustCompile(`\d+`)
	channelDigits := re.FindString(text)
	if channelDigits == "" {
		return 0, false
	}

	// 转换为整数
	channelNum, err := strconv.Atoi(channelDigits)
	if err != nil {
		return 0, false
	}

	return channelNum, true
}

// IsChannelEqual 检查文本中的频道号是否等于目标频道
// text: OCR识别的频道文本
// targetChannel: 目标频道号
// 返回: true表示频道号相等，false表示不相等或无法提取
func IsChannelEqual(text string, targetChannel int) bool {
	channelNum, ok := ExtractChannelNumber(text)
	if !ok {
		return false
	}
	return channelNum == targetChannel
}

// ExtractRightPart extracts the right part after the last occurrence of a separator
// input: the full string
// separator: the separator character/string
// returns: the right part after the separator, or the original string if separator not found
func ExtractRightPart(input, separator string) string {
	if input == "" || separator == "" {
		return input
	}

	// Find the last occurrence of the separator
	lastIndex := strings.LastIndex(input, separator)
	if lastIndex == -1 {
		return input
	}

	// Return the part after the separator
	return input[lastIndex+len(separator):]
}

// RemoveSpaces removes all spaces from a string
// input: the string to process
// returns: the string with all spaces removed
func RemoveSpaces(input string) string {
	return strings.ReplaceAll(input, " ", "")
}

// ParseDigitInRange 识别OCR文本是否为数字，如果是数字且在指定范围内，返回该数字
// 参数: text - OCR识别的文本, min - 最小值, max - 最大值
// 返回: (数字值, 是否有效)
// 如果文本是数字且在min-max范围内，返回 (数字, true)
// 否则返回 (0, false)
func ParseDigitInRange(text string, min, max int) (int, bool) {
	// 提取所有数字
	digits := ""
	for _, char := range text {
		if char >= '0' && char <= '9' {
			digits += string(char)
		}
	}

	// 如果没有数字，返回false
	if digits == "" {
		return 0, false
	}

	// 转换为整数
	num, err := strconv.Atoi(digits)
	if err != nil {
		return 0, false
	}

	// 判断是否在指定范围内
	if num >= min && num <= max {
		return num, true
	}

	return 0, false
}

// GenerateMultipleGameNames generates multiple unique random game names
// count: number of names to generate
// pattern: combination pattern (1-6)
func GenerateMultipleGameNames(count, pattern int) []string {
	names := make([]string, 0, count)
	nameSet := make(map[string]bool)

	for len(names) < count {
		name := GenerateRandomGameName(pattern)
		if !nameSet[name] {
			nameSet[name] = true
			names = append(names, name)
		}
	}

	return names
}

// GetGameNamePatterns returns all available game name generation patterns
func GetGameNamePatterns() map[int]string {
	return map[int]string{
		1: "prefix+core+number",
		2: "core+number",
		3: "prefix+core+number",
		4: "core+core",
		5: "random combination pattern",
	}
}

// GenerateRandomGameNameAnyPattern generates random game names using any pattern (1-5)
// This function randomly selects from all available patterns and ensures 6-10 character length
func GenerateRandomGameNameAnyPattern() string {
	rand.Seed(time.Now().UnixNano())

	// Randomly select from patterns 1-5
	patterns := []int{1, 2, 3, 4, 5}
	selectedPattern := patterns[rand.Intn(len(patterns))]

	return GenerateRandomGameName(selectedPattern)
}

// GenerateRandomGameNameAnyPatternWithNumbers generates random game names using any pattern (1-5)
// but ensures the result contains numbers
func GenerateRandomGameNameAnyPatternWithNumbers() string {
	rand.Seed(time.Now().UnixNano())

	// Randomly select from patterns that include numbers (1, 2, 3)
	patterns := []int{1, 2, 3}
	selectedPattern := patterns[rand.Intn(len(patterns))]

	return GenerateRandomGameName(selectedPattern)
}

// GenerateRandomGameNameAnyPatternPureText generates random game names using any pattern (1-5)
// but ensures the result contains only letters (no numbers)
func GenerateRandomGameNameAnyPatternPureText() string {
	rand.Seed(time.Now().UnixNano())

	// Randomly select from patterns that don't include numbers (4)
	patterns := []int{4}
	selectedPattern := patterns[rand.Intn(len(patterns))]

	return GenerateRandomGameName(selectedPattern)
}

func NormalizeString(s string) string {
	// 1. 替换全角字符为半角字符
	s = strings.ReplaceAll(s, "（", "(")
	s = strings.ReplaceAll(s, "）", ")")
	s = strings.ReplaceAll(s, "：", ":") // 全角冒号转半角冒号

	// 2. 移除特定词汇
	s = strings.ReplaceAll(s, "归属", "")

	// 3. 移除所有括号、空格和标点符号，以处理格式差异
	s = strings.ReplaceAll(s, "(", "")
	s = strings.ReplaceAll(s, ")", "")
	s = strings.ReplaceAll(s, " ", "") // 移除空格
	s = strings.ReplaceAll(s, ":", "") // 移除冒号

	// 4. 可以根据需要添加其他规则，例如转换为小写等
	// s = strings.ToLower(s)

	return s
}
