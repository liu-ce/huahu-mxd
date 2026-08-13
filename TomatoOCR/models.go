package TomatoOCR

// DetectResult 代表单次检测+识别的结果项
type DetectResult struct {
	Location [4][2]int `json:"location"` // 矩形坐标点
	Score    float64   `json:"score"`    // 检测置信度
	Words    string    `json:"words"`    // 检测到的文字
}

// RecognizeResult 代表单次仅识别的结果项
// 注意：根据注释，type=2时返回的是单个对象，而非数组
type RecognizeResult struct {
	Location [4][2]int `json:"location"` // 矩形坐标点
	Score    float64   `json:"score"`    // 检测置信度
	Words    string    `json:"words"`    // 检测到的文字
}

// TapPoint 代表 FindTapPoints 返回的结果项
type TapPoint struct {
	Point [2]int `json:"point"` // 中心坐标点
	Words string `json:"words"` // 检测到的文字
}
