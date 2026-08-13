package core

import (
	"app/assets"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Dasongzi1366/AutoGo/utils"
	"github.com/golang/protobuf/proto"

	sls "github.com/aliyun/aliyun-log-go-sdk"
)

var client sls.ClientInterface
var project string
var logstore string

func init() {
	// 读取SLS配置
	configData, err := assets.ConfigFile.ReadFile("config/default.json")
	if err != nil {
		SLS_Log2(fmt.Sprintf("读取SLS配置失败: %v", err))
		return
	}
	var config map[string]interface{}
	if err := json.Unmarshal(configData, &config); err != nil {
		SLS_Log2(fmt.Sprintf("解析SLS配置失败: %v", err))
		return
	}
	slsConfig := config["sls"].(map[string]interface{})

	endpoint := slsConfig["endpoint"].(string)
	accessKeyID := slsConfig["access_key_id"].(string)
	accessKeySecret := slsConfig["access_key_secret"].(string)
	project = slsConfig["project"].(string)
	logstore = slsConfig["logstore"].(string)
	securityToken := ""

	provider := sls.NewStaticCredentialsProvider(accessKeyID, accessKeySecret, securityToken)
	client = sls.CreateNormalInterfaceV2(endpoint, provider)
}

// SLS_Log 写入一条日志到阿里云SLS，包含内容、设备号、用户名
func SLS_Log(message string) {
	utils.Toast(message, 300, 0, 1000)
	SLS(message)
}

// ErrorDetectionConfig 异常检测配置表（公开变量，供其他包使用）
// key: 异常消息关键词
// value: 1小时内的异常次数阈值
var ErrorDetectionConfig = map[string]int{
	"副本识别不到数字": 8,
}

// SLS_ERR 记录异常日志并检测是否需要触发处理
// message: 异常消息内容
func SLS_ERR(message string) {
	utils.Toast(message, 300, 0, 10000)
	SLS_Log2("发现异常:  " + message)

	// 遍历异常检测配置，记录异常
	for errorKey := range ErrorDetectionConfig {
		// 检查异常消息是否包含当前配置的关键词
		if strings.Contains(message, errorKey) {
			// 记录此次异常
			Storages.AddDungeonNumberRecognitionError(errorKey)
		}
	}
}

func Debug(message string) {
	websocketInfo := API.GetWebsocketInfo()
	if websocketInfo == "1" {
		SLS_Log2(message)
	} else {
		fmt.Println(message)
	}
}

func SLS_Log2(message string) {

	fmt.Println(message)
	utils.Toast(message, 300, 0, 1500)
	//if tokenInterface == nil {
	//	return
	//}
	//
	//if client == nil {
	//	fmt.Println("SLS客户端未初始化")
	//	return
	//}
	//
	//// 构造日志内容
	//logItem := &sls.Log{
	//	Time: proto.Uint32(uint32(time.Now().Unix())),
	//	Contents: []*sls.LogContent{
	//
	//		{
	//			Key:   proto.String("username"),
	//			Value: proto.String(Get("username").(string)),
	//		},
	//		{
	//			Key:   proto.String("seq"),
	//			Value: proto.String(Get("seq").(string)),
	//		},
	//		{
	//			Key:   proto.String("content"),
	//			Value: proto.String(message),
	//		},
	//		{
	//			Key:   proto.String("version"),
	//			Value: proto.String(Role.Version),
	//		},
	//	},
	//}
	//logGroup := &sls.LogGroup{
	//	Topic:  proto.String(""),
	//	Source: proto.String(""),
	//	Logs:   []*sls.Log{logItem},
	//}
	//
	//err := client.PutLogs(project, logstore, logGroup)
	//if err != nil {
	//	fmt.Printf("PutLogs failed %v\n", err)
	//}

}

// SLS_Log2NoToast 仅打印日志，不弹 Toast（避免覆盖 GM 巡逻等长 Toast）。
func SLS_Log2NoToast(message string) {
	fmt.Println(message)
}

func slsPutLogContent(message string) {
	tokenInterface := Get("token")
	if tokenInterface == nil {
		return
	}

	if client == nil {
		fmt.Println("SLS客户端未初始化")
		return
	}

	versionVal := ""
	if Role != nil {
		versionVal = Role.Version
	}

	// 构造日志内容
	logItem := &sls.Log{
		Time: proto.Uint32(uint32(time.Now().Unix())),
		Contents: []*sls.LogContent{

			{
				Key:   proto.String("username"),
				Value: proto.String(Get("username").(string)),
			},
			{
				Key:   proto.String("seq"),
				Value: proto.String(Get("seq").(string)),
			},
			{
				Key:   proto.String("content"),
				Value: proto.String(message),
			},
			{
				Key:   proto.String("version"),
				Value: proto.String(versionVal),
			},
		},
	}
	logGroup := &sls.LogGroup{
		Topic:  proto.String(""),
		Source: proto.String(""),
		Logs:   []*sls.Log{logItem},
	}

	err := client.PutLogs(project, logstore, logGroup)
	if err != nil {
		fmt.Printf("PutLogs failed %v\n", err)
	}

}

// SLS_ContentNoToast 仅写入 SLS（不 Toast），用于定时内存/GC 等指标。
func SLS_ContentNoToast(message string) {
	slsPutLogContent(message)
}

func SLS(message string) {

	slsPutLogContent(message)

}
