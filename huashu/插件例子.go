package huashu

import (
	"app/TaiBaiYoloV5Ncnn"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/Dasongzi1366/AutoGo/files"
	"github.com/Dasongzi1366/AutoGo/images"
)

// 下面代码不要改动

var (
	TaiBaiYoloV5NcnnDetector *TaiBaiYoloV5Ncnn.Detector = nil
)

func 初始化插件(Args ...interface{}) (bool, error) {
	DefaultLabel := "label.txt"  // 参数1
	DefaultParam := "best.param" // 参数2
	DefaultBin := "best.bin"     // 参数3
	DefaultThread := 4           // 参数4
	for I, Arg := range Args {
		switch I {
		case 0:
			if V, OK := Arg.(string); OK {
				DefaultLabel = V
			}
		case 1:
			if V, OK := Arg.(string); OK {
				DefaultParam = V
			}
		case 2:
			if V, OK := Arg.(string); OK {
				DefaultBin = V
			}
		case 3:
			if V, OK := Arg.(int); OK {
				DefaultThread = V
			}
		}
	}
	Lists := map[string]string{
		"Label": DefaultLabel,
		"Param": DefaultParam,
		"Bin":   DefaultBin,
	}
	Assets := files.Path("./assets")
	for Key, Value := range Lists {
		IsRes := false
		if strings.Index(Value, "/") == -1 {
			Lists[Key] = fmt.Sprintf("%s/%s", Assets, Value)
			IsRes = true
		}
		if !files.Exists(Lists[Key]) {
			if IsRes {
				return false, errors.New(fmt.Sprintf("%s 文件不存在, 请将文件添加到resources/assets文件夹中", Value))
			}
			return false, errors.New(fmt.Sprintf("%s 文件不存在, 请确保文件被释放到该目录中", Lists[Key]))
		}
	}
	TaiBaiYoloV5NcnnDetector = TaiBaiYoloV5Ncnn.NewDetector()
	Err := TaiBaiYoloV5NcnnDetector.LoadModel(Lists["Label"], Lists["Param"], Lists["Bin"], DefaultThread)
	if Err != nil {
		return false, Err
	}
	return true, nil
}

func 识别本地文件(File string, Args ...interface{}) []TaiBaiYoloV5Ncnn.DetectorResult {
	var DefaultThreshold float32 = 0.45    // 参数2
	var DefaultNmsThreshold float32 = 0.55 // 参数3
	DefaultSize := 416                     // 参数4
	for I, Arg := range Args {
		switch I {
		case 0:
			if V, OK := Arg.(float32); OK {
				DefaultThreshold = V
			}
		case 1:
			if V, OK := Arg.(float32); OK {
				DefaultNmsThreshold = V
			}
		case 2:
			if V, OK := Arg.(int); OK {
				DefaultSize = V
			}
		}
	}
	Image := images.ReadFromPath(File)
	Value, _ := TaiBaiYoloV5NcnnDetector.Detect(Image, DefaultThreshold, DefaultNmsThreshold, DefaultSize)
	return Value
}

func 识别屏幕坐标(Args ...interface{}) []TaiBaiYoloV5Ncnn.DetectorResult {
	DefaultX1 := 0        // 参数1 左上角x坐标
	DefaultY1 := 0        // 参数2 左上角y坐标
	DefaultX2 := 0        // 参数3 右下角x坐标，0表示屏幕最大宽度
	DefaultY2 := 0        // 参数4 右下角y坐标，0表示屏幕最大高度
	DefaultDisplayId := 0 // 参数5 屏幕ID，0表示主屏幕，其他值表示虚拟屏幕

	var DefaultThreshold float32 = 0.45    // 参数6
	var DefaultNmsThreshold float32 = 0.55 // 参数7
	DefaultSize := 416                     // 参数8
	for I, Arg := range Args {
		switch I {
		case 0:
			if V, OK := Arg.(int); OK {
				DefaultX1 = V
			}
		case 1:
			if V, OK := Arg.(int); OK {
				DefaultY1 = V
			}
		case 2:
			if V, OK := Arg.(int); OK {
				DefaultX2 = V
			}
		case 3:
			if V, OK := Arg.(int); OK {
				DefaultY2 = V
			}
		case 4:
			if V, OK := Arg.(int); OK {
				DefaultDisplayId = V
			}
		case 5:
			if V, OK := Arg.(float32); OK {
				DefaultThreshold = V
			}
		case 6:
			if V, OK := Arg.(float32); OK {
				DefaultNmsThreshold = V
			}
		case 7:
			if V, OK := Arg.(int); OK {
				DefaultSize = V
			}
		}
	}
	Image := images.CaptureScreen(DefaultX1, DefaultY1, DefaultX2, DefaultY2, DefaultDisplayId)
	Value, _ := TaiBaiYoloV5NcnnDetector.Detect(Image, DefaultThreshold, DefaultNmsThreshold, DefaultSize)
	return Value
}

// 上面代码不要改动

func main() {
	// 只需要执行一次
	初始化状态, 是否有错误 := 初始化插件()
	if 初始化状态 == false {
		fmt.Println(是否有错误)
		// 停止脚本或做其他处理
		os.Exit(0)
	}

	图片路径 := fmt.Sprintf("%s/目标检测.png", files.Path("./assets"))
	结果 := 识别本地文件(图片路径)
	fmt.Println(结果[0])

	// 识别屏幕坐标() // 识别全屏
}
