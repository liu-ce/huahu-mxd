package main

import (
	"app/captcha"
	"app/core"
	"app/play"
	"app/util"
	"fmt"
	"os"

	"github.com/Dasongzi1366/AutoGo/storages"
	"github.com/gin-gonic/gin"
)

func init() {
	core.RegisterGMPatrolAnswerScreenshotUploader(util.UploadAnswerScreenshot)
}

// UI回调函数，当UI访问路由时触发
func uiCallback(c *gin.Context) bool {
	urlPath := c.Request.URL.Path

	if urlPath == "/login" {
		// 获取所有表单数据
		username := c.Query("username")
		password := c.Query("password")
		windowId := c.Query("windowId")

		storages.Put("data", "username", username)
		storages.Put("data", "password", password)
		storages.Put("data", "windowId", windowId)

		return true
	}

	if urlPath == "/getFormData" {
		username := storages.Get("data", "username")
		password := storages.Get("data", "password")
		windowId := storages.Get("data", "windowId")

		c.Header("Content-Type", "application/json")
		c.JSON(200, gin.H{
			"username": username,
			"password": password,
			"windowId": windowId,
		})
	}

	if urlPath == "/close" {
		os.Exit(0)
	}

	return false
}

func main() {

	username := storages.Get("data", "username")
	password := storages.Get("data", "password")
	windowId := storages.Get("data", "windowId")

	err := core.API.LoginAndSetup(username, password, windowId)
	if err != nil {
		core.SLS_Log2(err.Error())
		core.Sleep(3000)
		return
	}

	// 地图名写死在 UI / 代码里，对应 assets/config/farm_map_<名>.json
	// 挂机地图 := core.API.GetConfigStringValue("挂机地图") // 旧：从中控读取
	挂机地图 := "韩服抢夺宝物岛" // 入门测试地图，对应 farm_map_韩服抢夺宝物岛.json
	fmt.Println(挂机地图)
	core.ExitOnGameExceptionIfNeeded = ExitOnGameExceptionIfNeeded

	go captcha.Run()

	if err := play.RunAutoFarm(fmt.Sprintf("config/farm_map_%s.json", 挂机地图)); err != nil {
		fmt.Println("挂机流程结束:", err)
	}

	// 定时下线测试：游戏内角色界面时调用；参数=离线等待分钟（<=0 则 1 分钟）
	//play.TestScheduledLogoutOnce(0)

	//if err := play.RunAutoFarm(fmt.Sprintf("config/farm_map_%s.json", "land赫勒地区瞬移版")); err != nil {
	//	fmt.Println("挂机流程结束:", err)
	//}

	//captcha.YoloRegionMonsterTest()

	//captcha.RunMinimapYellowPointLivePrint()

	//captcha.YoloTest()

	//select {}

}
