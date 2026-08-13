package TomatoOCR

/*
#include "TomatoOCR.h"
#include <stdlib.h>
#cgo CXXFLAGS: -std=c++11
#cgo arm64 LDFLAGS: -L../resources/libs/arm64-v8a -lTomatoOCR
#cgo amd64 LDFLAGS: -L../resources/libs/x86_64 -lTomatoOCR
#cgo 386 LDFLAGS: -L../resources/libs/x86 -lTomatoOCR
*/
import "C"
import (
	"os"
	"path/filepath"
	"unsafe"

	"github.com/Dasongzi1366/AutoGo/images"
)

type TmoOcr struct {
	pointer *C.TomatoOCR
}

// DetectResult 表示 OCR 检测结果
//type DetectResult struct {
//	X       int     // 左上角 X 坐标
//	Y       int     // 左上角 Y 坐标
//	Width   int     // 矩形宽度
//	Height  int     // 矩形高度
//	CenterX int     // 中心点 X 坐标
//	CenterY int     // 中心点 Y 坐标
//	Score   float32 // 检测结果的置信度
//	Label   string  // 检测到的文字
//}

// New 创建一个新的 TmoOcr 实例，并初始化 OCR 模型。
func New() *TmoOcr {
	// 调用底层 TomatoOCR_new 函数创建 C++ 对象
	pointer := C.TomatoOCR_new()
	if pointer == nil {
		return nil
	}

	// 构建模型文件所在路径
	modelPath := filepath.Dir(os.Args[0]) + "/assets/"

	// 模型文件路径初始化
	cModelPath := C.CString(modelPath)
	defer C.free(unsafe.Pointer(cModelPath))

	// 调用初始化函数并判断是否成功
	if int(C.init(pointer, cModelPath)) == 0 {
		return nil
	}

	return &TmoOcr{
		pointer: pointer,
	}
}

// SetMode 设置模式
func (t *TmoOcr) SetMode(mode string) {
	cMode := C.CString(mode)
	defer C.free(unsafe.Pointer(cMode))

	C.setMode(t.pointer, cMode)
}

// SetHttpIntervalTime 设置间隔时间
func (t *TmoOcr) SetHttpIntervalTime(second int) {
	C.setHttpIntervalTime(t.pointer, C.int(second))
}

// SetLicense 设置授权信息
func (t *TmoOcr) SetLicense(license, remark string) string {
	cLicense := C.CString(license)
	defer C.free(unsafe.Pointer(cLicense))

	cRemark := C.CString(remark)
	defer C.free(unsafe.Pointer(cRemark))

	return C.GoString(C.setLicense(t.pointer, cLicense, cRemark))
}

// SetRecType 设置识别类型
func (t *TmoOcr) SetRecType(recType string) {
	cRecType := C.CString(recType)
	defer C.free(unsafe.Pointer(cRecType))

	C.setRecType(t.pointer, cRecType)
}

// SetDetBoxType 设置检测框类型
func (t *TmoOcr) SetDetBoxType(detBoxType string) {
	cDetBoxType := C.CString(detBoxType)
	defer C.free(unsafe.Pointer(cDetBoxType))

	C.setDetBoxType(t.pointer, cDetBoxType)
}

// SetDetScaleRatio 设置图片缩放比例
func (t *TmoOcr) SetDetScaleRatio(detScaleRatio float32) {
	C.setDetScaleRatio(t.pointer, C.float(detScaleRatio))
}

// SetDetUnclipRatio 设置检测框扩展比例
func (t *TmoOcr) SetDetUnclipRatio(detUnclipRatio float32) {
	C.setDetUnclipRatio(t.pointer, C.float(detUnclipRatio))
}

// SetRecScoreThreshold 设置识别的置信度阈值
func (t *TmoOcr) SetRecScoreThreshold(recScoreThreshold float32) {
	C.setRecScoreThreshold(t.pointer, C.float(recScoreThreshold))
}

// SetReturnType 设置返回结果的格式类型
func (t *TmoOcr) SetReturnType(returnType string) {
	cReturnType := C.CString(returnType)
	defer C.free(unsafe.Pointer(cReturnType))

	C.setReturnType(t.pointer, cReturnType)
}

// SetBinaryThresh 设置二值化阈值
func (t *TmoOcr) SetBinaryThresh(binaryThresh int) {
	C.setBinaryThresh(t.pointer, C.int(binaryThresh))
}

// SetRunMode 设置二值化阈值
func (t *TmoOcr) SetRunMode(runMode string) {
	cRunMode := C.CString(runMode)
	defer C.free(unsafe.Pointer(cRunMode))
	C.setRunMode(t.pointer, cRunMode)
}

//func (t *TmoOcr) SetFilterColor(filterColor string) {
//	cFilterColor := C.CString(filterColor)
//	defer C.free(unsafe.Pointer(cFilterColor))
//	C.setFilterColor(t.pointer, cFilterColor)
//}

// SetFilterColor 设置滤色值和背景色
func (t *TmoOcr) SetFilterColor(filterColor string, backgroundColor string) {
	cFilterColor := C.CString(filterColor)
	cBackgroundColor := C.CString(backgroundColor)
	defer C.free(unsafe.Pointer(cFilterColor))
	defer C.free(unsafe.Pointer(cBackgroundColor))
	C.setFilterColor(t.pointer, cFilterColor, cBackgroundColor)
}

// SetBackgroundColor 设置背景色black/white，默认black
func (t *TmoOcr) SetBackgroundColor(backgroundColor string) {
	cBackgroundColor := C.CString(backgroundColor)
	defer C.free(unsafe.Pointer(cBackgroundColor))
	C.setBackgroundColor(t.pointer, cBackgroundColor)
}

// SetFilterColorPath 设置滤色后的图片路径
func (t *TmoOcr) SetFilterColorPath(filterColorPath string) {
	cFilterColorPath := C.CString(filterColorPath)
	defer C.free(unsafe.Pointer(cFilterColorPath))
	C.setFilterColorPath(t.pointer, cFilterColorPath)
}

// FindTapPoint 找字
func (t *TmoOcr) FindTapPoint(text string) string {
	cText := C.CString(text)
	defer C.free(unsafe.Pointer(cText))
	result := C.findTapPoint(t.pointer, cText)
	return C.GoString(result)
}

// FindTapPoints 找字
func (t *TmoOcr) FindTapPoints(text string) string {
	cText := C.CString(text)
	defer C.free(unsafe.Pointer(cText))
	result := C.findTapPoints(t.pointer, cText)
	return C.GoString(result)
}

// Detect 截取屏幕区域并进行 OCR 检测
func (t *TmoOcr) Detect(x1, y1, x2, y2, type_ int) string {
	// 截取屏幕图片
	img := images.CaptureScreen(x1, y1, x2, y2, 0)
	if img == nil {
		return ""
	}

	// 调用底层 OCR 函数进行检测
	result := C.ocrImageData(
		t.pointer,
		(*C.char)(unsafe.Pointer(&img.Pix[0])),
		C.int(img.Rect.Dx()),
		C.int(img.Rect.Dy()),
		C.int(type_),
	)

	// 将 C 字符串转换为 Go 字符串
	jsonStr := C.GoString(result)
	if jsonStr == "" {
		return ""
	}
	return jsonStr
}

func (t *TmoOcr) InitYolo(key string, yoloSecret string, yoloModelPath string, yoloLabelPath string) int {
	cKey := C.CString(key)
	defer C.free(unsafe.Pointer(cKey))

	cYoloSecret := C.CString(yoloSecret)
	defer C.free(unsafe.Pointer(cYoloSecret))

	cYoloModelPath := C.CString(yoloModelPath)
	defer C.free(unsafe.Pointer(cYoloModelPath))

	cYoloLabelPath := C.CString(yoloLabelPath)
	defer C.free(unsafe.Pointer(cYoloLabelPath))

	result := C.initYolo(t.pointer, cKey, cYoloSecret, cYoloModelPath, cYoloLabelPath)
	return int(C.int(result))
}

func (t *TmoOcr) DetectYolo(key string, x1 int, y1 int, x2 int, y2 int, targetSize int, scoreThreshold float32, nmsScoreThreshold float32) string {
	// 截取屏幕图片
	img := images.CaptureScreen(x1, y1, x2, y2, 0)
	if img == nil {
		return ""
	}

	// 调用底层 yolo 函数进行检测
	result := C.yoloImageData(
		t.pointer,
		C.CString(key),
		(*C.char)(unsafe.Pointer(&img.Pix[0])),
		C.int(img.Rect.Dx()),
		C.int(img.Rect.Dy()),
		C.int(targetSize),
		C.float(scoreThreshold),
		C.float(nmsScoreThreshold),
	)

	// 将 C 字符串转换为 Go 字符串
	jsonStr := C.GoString(result)
	if jsonStr == "" {
		return ""
	}
	return jsonStr
}

func (t *TmoOcr) DetectYoloFile(key string, imagePath string, targetSize int, scoreThreshold float32, nmsScoreThreshold float32) string {

	// 调用底层 yolo 函数进行检测
	result := C.yoloFile(
		t.pointer,
		C.CString(key),
		C.CString(imagePath),
		C.int(targetSize),
		C.float(scoreThreshold),
		C.float(nmsScoreThreshold),
	)

	// 将 C 字符串转换为 Go 字符串
	jsonStr := C.GoString(result)
	if jsonStr == "" {
		return ""
	}
	return jsonStr
}

func (t *TmoOcr) ReleaseYolo(key string) {
	cKey := C.CString(key)
	defer C.free(unsafe.Pointer(cKey))
	C.releaseYolo(t.pointer, cKey)
}

// ----------------------------
