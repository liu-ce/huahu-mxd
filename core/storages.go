package core

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Dasongzi1366/AutoGo/storages"
)

var (
	// 设置中国时区
	chinaLocation *time.Location
)

func init() {
	// 初始化中国时区
	var err error
	chinaLocation, err = time.LoadLocation("Asia/Shanghai")
	if err != nil {
		// 如果加载失败，使用UTC+8
		chinaLocation = time.FixedZone("CST", 8*3600)
	}
}

// StoragesHandler 存储功能处理器
type StoragesHandler struct{}

// NewStoragesHandler 创建新的存储功能处理器
func NewStoragesHandler() *StoragesHandler {
	return &StoragesHandler{}
}

// HasRecentActivity 检查指定功能是否在指定时间内执行过
// key: 功能键名
// minutes: 时间间隔（分钟），如果传入0则从配置文件获取
// 返回 true 表示最近执行过，false 表示可以执行
func (s *StoragesHandler) HasRecentActivity(key string, minutes int) bool {
	// 如果传入的时间为0，则从配置文件获取时长
	if minutes == 0 {
		minutes = GetTaskInterval(key, 120) // 默认120分钟（2小时）
	}

	// 获取上次执行时间
	lastTimeStr := storages.Get("功能", key)

	// 如果没有记录，说明从未执行过
	if lastTimeStr == "" {
		// 记录当前时间戳
		currentTime := time.Now().In(chinaLocation).Unix()
		storages.Put("功能", key, strconv.FormatInt(currentTime, 10))
		return false
	}

	// 解析上次执行时间
	lastTime, err := strconv.ParseInt(lastTimeStr, 10, 64)
	if err != nil {
		// 如果解析失败，清除错误数据并重新记录
		storages.Remove("功能", key)
		currentTime := time.Now().In(chinaLocation).Unix()
		storages.Put("功能", key, strconv.FormatInt(currentTime, 10))
		return false
	}

	// 计算时间差（分钟）
	currentTime := time.Now().In(chinaLocation).Unix()
	timeDiff := (currentTime - lastTime) / 60

	// 如果时间差小于指定分钟数，说明最近执行过
	if timeDiff < int64(minutes) {
		return true
	}

	// 超过指定时间，更新执行时间并返回false
	storages.Put("功能", key, strconv.FormatInt(currentTime, 10))
	return false
}

func (s *StoragesHandler) HasRecentActivitySeconds(key string, seconds int) bool {
	// 如果传入的时间为0，则从配置文件获取时长
	if seconds == 0 {
		seconds = GetTaskInterval(key, 60) // 默认60秒
	}

	lastTimeStr := storages.Get("功能", key)

	// 如果没有记录，说明从未执行过
	if lastTimeStr == "" {
		// 记录当前时间戳
		currentTime := time.Now().In(chinaLocation).Unix()
		storages.Put("功能", key, strconv.FormatInt(currentTime, 10))
		return false
	}

	// 解析上次执行时间
	lastTime, err := strconv.ParseInt(lastTimeStr, 10, 64)
	if err != nil {
		// 如果解析失败，清除错误数据并重新记录
		storages.Remove("功能", key)
		currentTime := time.Now().In(chinaLocation).Unix()
		storages.Put("功能", key, strconv.FormatInt(currentTime, 10))
		return false
	}

	// 计算时间差（秒）
	currentTime := time.Now().In(chinaLocation).Unix()
	timeDiff := currentTime - lastTime

	// 如果时间差小于指定秒数，说明最近执行过
	if timeDiff < int64(seconds) {
		return true
	}

	// 超过指定时间，更新执行时间并返回false
	storages.Put("功能", key, strconv.FormatInt(currentTime, 10))
	return false
}

// CheckAndUpdateRecentActivitySeconds 检查指定功能是否在指定秒数内执行过，如果不在则更新时间
// key: 功能键名
// seconds: 时间间隔（秒）
// 返回: 如果在时间间隔内返回 -3，否则返回 0 并更新时间
func (s *StoragesHandler) CheckAndUpdateRecentActivitySeconds(key string, seconds int) int {
	lastTimeStr := storages.Get("功能", key)
	now := time.Now().In(chinaLocation).Unix()

	if lastTimeStr != "" {
		if lastTs, err := strconv.ParseInt(lastTimeStr, 10, 64); err == nil {
			if now-lastTs <= int64(seconds) {
				return -3
			}
		}
	}

	storages.Put("功能", key, strconv.FormatInt(now, 10))
	return 0
}

// GetNextActivityTime 获取下次可执行时间（基于最近活动时间）
// key: 功能键名
// minutes: 时间间隔（分钟），如果传入0则从配置文件获取
// 返回: 格式化的时间字符串，格式为 "2006-01-02 15:04:05"，如果没有记录则返回空字符串
func (s *StoragesHandler) GetNextActivityTime(key string, minutes int) string {
	// 如果传入的时间为0，则从配置文件获取时长
	if minutes == 0 {
		minutes = GetTaskInterval(key, 120) // 默认120分钟（2小时）
	}

	if Role.Level <= 25 {
		minutes = minutes / 2
	}

	// 获取上次执行时间
	lastTimeStr := storages.Get("功能", key)
	if lastTimeStr == "" {
		// 没有记录
		return ""
	}

	// 解析上次执行时间
	lastTime, err := strconv.ParseInt(lastTimeStr, 10, 64)
	if err != nil {
		// 解析失败
		return ""
	}

	// 计算下次可执行时间 = 上次执行时间 + 间隔时间
	nextTime := lastTime + int64(minutes*60)

	// 使用中国时区格式化时间显示
	nextTimeFormatted := time.Unix(nextTime, 0).In(chinaLocation).Format("2006-01-02 15:04:05")
	return nextTimeFormatted
}

// SetTaskDelay 设置任务延迟状态，指定任务在指定时间后才能执行
// key: 任务键名
// delayMinutes: 延迟时间（分钟），如果传入0则从配置文件获取
func (s *StoragesHandler) SetTaskDelay(key string, delayMinutes int) {
	// 如果传入的时间为0，则从配置文件获取时长
	if delayMinutes == 0 {
		delayMinutes = GetTaskInterval(key, 120) // 默认120分钟（2小时）
	}

	// 计算下次可执行时间
	nextTime := time.Now().In(chinaLocation).Add(time.Duration(delayMinutes) * time.Minute).Unix()
	// 使用中国时区格式化时间显示
	nextTimeFormatted := time.Unix(nextTime, 0).In(chinaLocation).Format("2006-01-02 15:04:05")
	fmt.Println("[任务延迟] 设置 " + key + " 延迟 " + fmt.Sprintf("%d", delayMinutes) + " 分钟，下次执行时间: " + nextTimeFormatted)
	storages.Put("任务延迟", key, strconv.FormatInt(nextTime, 10))
}

// IsTaskDelayed 检查任务是否被延迟，是否应该跳过
// key: 任务键名
// 返回 true 表示任务被延迟应该跳过，false 表示可以执行
func (s *StoragesHandler) IsTaskDelayed(key string) bool {
	// 获取下次可执行时间
	nextTimeStr := storages.Get("任务延迟", key)
	if nextTimeStr == "" {
		// 没有延迟记录，可以执行
		return false
	}

	// 解析下次可执行时间
	nextTime, err := strconv.ParseInt(nextTimeStr, 10, 64)
	if err != nil {
		// 如果解析失败，清除错误数据
		SLS_Log2(fmt.Sprintf("[任务延迟] %s 时间解析失败，清除延迟记录\n", key))
		storages.Remove("任务延迟", key)
		return false
	}

	// 检查当前时间是否已经超过下次可执行时间
	currentTime := time.Now().In(chinaLocation).Unix()
	if currentTime < nextTime {
		// 还没到执行时间，应该跳过
		//nextTimeFormatted := time.Unix(nextTime, 0).In(chinaLocation).Format("2006-01-02 15:04:05")
		//fmt.Println(fmt.Sprintf("[任务延迟] %s 还没到执行时间，应该跳过。下次执行时间: %s", key, nextTimeFormatted))
		return true
	}

	// 已经到执行时间，清除延迟记录并允许执行
	//SLS_Log2(fmt.Sprintf("[任务延迟] %s 延迟时间已到，清除延迟记录，允许执行", key))
	storages.Remove("任务延迟", key)
	return false
}

// GetTaskDelayTime 获取任务的延迟时间（格式化后的字符串）
// key: 任务键名
// 返回: 格式化的时间字符串，格式为 "2006-01-02 15:04:05"，如果没有延迟记录则返回空字符串
func (s *StoragesHandler) GetTaskDelayTime(key string) string {
	// 获取下次可执行时间
	nextTimeStr := storages.Get("任务延迟", key)
	if nextTimeStr == "" {
		// 没有延迟记录
		return ""
	}

	// 解析下次可执行时间
	nextTime, err := strconv.ParseInt(nextTimeStr, 10, 64)
	if err != nil {
		// 如果解析失败，返回空字符串
		return ""
	}

	// 使用中国时区格式化时间显示
	nextTimeFormatted := time.Unix(nextTime, 0).In(chinaLocation).Format("2006-01-02 15:04:05")
	return nextTimeFormatted
}

// 如果没做过就做 如果任务没有被延迟，则执行一次并设置延迟
// key: 任务键名
// delayMinutes: 延迟时间（分钟），如果传入0则从配置文件获取
// 返回 true 表示应该执行任务，false 表示任务被延迟应该跳过
func (s *StoragesHandler) A_如果没做过就做(key string, delayMinutes int) bool {
	// 如果传入的时间为0，则从配置文件获取时长
	if delayMinutes == 0 {
		delayMinutes = GetTaskInterval(key, 120) // 默认120分钟（2小时）
	}

	// 检查任务是否被延迟
	if s.IsTaskDelayed(key) {
		return false // 任务被延迟，不执行
	}

	// 任务没有被延迟，设置延迟并返回 true 表示可以执行
	s.SetTaskDelay(key, delayMinutes)
	return true
}

// SetTaskDelaySeconds 设置任务延迟状态（秒级），指定任务在指定时间后才能执行
// key: 任务键名
// delaySeconds: 延迟时间（秒），如果传入0则使用默认值60秒
func (s *StoragesHandler) SetTaskDelaySeconds(key string, delaySeconds int) {
	// 如果传入的时间为0，使用默认值60秒
	if delaySeconds == 0 {
		delaySeconds = 60
	}

	// 计算下次可执行时间
	nextTime := time.Now().In(chinaLocation).Add(time.Duration(delaySeconds) * time.Second).Unix()
	// 使用中国时区格式化时间显示
	nextTimeFormatted := time.Unix(nextTime, 0).In(chinaLocation).Format("2006-01-02 15:04:05")
	fmt.Println("[任务延迟] 设置 " + key + " 延迟 " + fmt.Sprintf("%d", delaySeconds) + " 秒，下次执行时间: " + nextTimeFormatted)
	storages.Put("任务延迟", key, strconv.FormatInt(nextTime, 10))
}

// A_如果没做过就做_秒级 如果任务没有被延迟，则执行一次并设置延迟（秒级）
// key: 任务键名
// delaySeconds: 延迟时间（秒），如果传入0则使用默认值60秒
// 返回 true 表示应该执行任务，false 表示任务被延迟应该跳过
func (s *StoragesHandler) A_如果没做过就做_秒级(key string, delaySeconds int) bool {
	// 如果传入的时间为0，使用默认值60秒
	if delaySeconds == 0 {
		delaySeconds = 60
	}

	// 检查任务是否被延迟
	if s.IsTaskDelayed(key) {
		return false // 任务被延迟，不执行
	}

	// 任务没有被延迟，设置延迟并返回 true 表示可以执行
	s.SetTaskDelaySeconds(key, delaySeconds)
	return true
}

// ClearAllData 清空所有存储数据
func (s *StoragesHandler) ClearAllData() {
	// 清空所有数据
	storages.Clear("任务延迟")
	storages.Clear("功能")

}

// GetCounter 获取计数器当前值，如果不存在则返回0
// key: 计数器键名
func (s *StoragesHandler) GetCounter(key string) int {
	counterStr := storages.Get("计数器", key)
	if counterStr == "" {
		return 0
	}

	counter, err := strconv.Atoi(counterStr)
	if err != nil {
		// 如果解析失败，重置为0
		storages.Put("计数器", key, "0")
		return 0
	}

	return counter
}

// SetCounter 设置计数器值
// key: 计数器键名
// value: 要设置的值
func (s *StoragesHandler) SetCounter(key string, value int) {
	storages.Put("计数器", key, strconv.Itoa(value))
}

// IncrementCounter 增加计数器值并返回新值
// key: 计数器键名
// increment: 增加的值，默认为1
func (s *StoragesHandler) IncrementCounter(key string, increment int) int {
	if increment == 0 {
		increment = 1
	}

	current := s.GetCounter(key)
	newValue := current + increment
	s.SetCounter(key, newValue)

	return newValue
}

// ResetCounter 重置计数器为0
// key: 计数器键名
func (s *StoragesHandler) ResetCounter(key string) {
	s.SetCounter(key, 0)
}

// AddDeath 记录一次死亡
func (s *StoragesHandler) AddDeath() {
	currentTime := time.Now().In(chinaLocation).Unix()

	// 读取现有的时间戳列表
	deathTimestampsStr := storages.Get("死亡统计", "死亡次数")

	var timestamps []int64
	if deathTimestampsStr != "" {
		// 解析现有的时间戳列表
		parts := strings.Split(deathTimestampsStr, ",")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			if ts, err := strconv.ParseInt(part, 10, 64); err == nil {
				timestamps = append(timestamps, ts)
			}
		}
	}

	// 添加当前时间戳
	timestamps = append(timestamps, currentTime)

	// 清理24小时外的时间戳
	timestamps = s.filterTimestampsIn24Hours(timestamps)

	// 保存回存储
	s.saveDeathTimestamps("死亡次数", timestamps)
}

// GetDeathCountIn24Hours 获取24小时内死亡次数
func (s *StoragesHandler) GetDeathCountIn24Hours() int {
	deathTimestampsStr := storages.Get("死亡统计", "死亡次数")
	if deathTimestampsStr == "" {
		return 0
	}

	// 解析时间戳列表
	var timestamps []int64
	parts := strings.Split(deathTimestampsStr, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if ts, err := strconv.ParseInt(part, 10, 64); err == nil {
			timestamps = append(timestamps, ts)
		}
	}

	// 过滤24小时内的记录
	timestamps = s.filterTimestampsIn24Hours(timestamps)

	// 清理并保存（移除过期记录）
	s.saveDeathTimestamps("死亡次数", timestamps)

	return len(timestamps)
}

// filterTimestampsIn24Hours 过滤24小时内的时间戳
func (s *StoragesHandler) filterTimestampsIn24Hours(timestamps []int64) []int64 {
	currentTime := time.Now().In(chinaLocation).Unix()
	cutoffTime := currentTime - 24*60*60 // 24小时前的时间戳

	var validTimestamps []int64
	for _, ts := range timestamps {
		if ts >= cutoffTime {
			validTimestamps = append(validTimestamps, ts)
		}
	}

	return validTimestamps
}

// saveDeathTimestamps 保存时间戳列表
func (s *StoragesHandler) saveDeathTimestamps(key string, timestamps []int64) {
	if len(timestamps) == 0 {
		// 如果没有有效记录，清空存储
		storages.Remove("死亡统计", key)
		return
	}

	// 将时间戳列表转换为逗号分隔的字符串
	var parts []string
	for _, ts := range timestamps {
		parts = append(parts, strconv.FormatInt(ts, 10))
	}

	timestampsStr := strings.Join(parts, ",")
	storages.Put("死亡统计", key, timestampsStr)
}

// SetTeamHeartbeatTime 设置组队心跳时间，将当前时间存储到 key "组队心跳时间"
func (s *StoragesHandler) SetTeamHeartbeatTime() {
	currentTime := time.Now().In(chinaLocation).Unix()
	storages.Put("功能", "组队心跳时间", strconv.FormatInt(currentTime, 10))
}

// GetTeamHeartbeatMinutes 获取上次设置的时间距离现在多少分钟
// 如果没设置值就返回0
func (s *StoragesHandler) GetTeamHeartbeatMinutes() int {
	lastTimeStr := storages.Get("功能", "组队心跳时间")
	if lastTimeStr == "" {
		// 没有设置过，返回0
		return 10
	}

	// 解析上次设置的时间
	lastTime, err := strconv.ParseInt(lastTimeStr, 10, 64)
	if err != nil {
		// 解析失败，返回0
		return 0
	}

	// 计算时间差（分钟）
	currentTime := time.Now().In(chinaLocation).Unix()
	timeDiffMinutes := (currentTime - lastTime) / 60

	return int(timeDiffMinutes)
}

// AddDungeonNumberRecognitionError 记录一次指定类型的异常
// key: 异常类型键名，如 "副本识别不到数字"
func (s *StoragesHandler) AddDungeonNumberRecognitionError(key string) {
	currentTime := time.Now().In(chinaLocation).Unix()

	// 读取现有的时间戳列表
	timestampsStr := storages.Get("异常统计", key)

	var timestamps []int64
	if timestampsStr != "" {
		// 解析现有的时间戳列表
		parts := strings.Split(timestampsStr, ",")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			if ts, err := strconv.ParseInt(part, 10, 64); err == nil {
				timestamps = append(timestamps, ts)
			}
		}
	}

	// 添加当前时间戳
	timestamps = append(timestamps, currentTime)

	// 清理1小时外的时间戳
	timestamps = s.filterTimestampsIn1Hour(timestamps)

	// 保存回存储
	s.saveErrorTimestamps(key, timestamps)
}

// GetDungeonNumberRecognitionErrorCountIn1Hour 获取1小时内指定异常的异常次数
// key: 异常类型键名，如 "副本识别不到数字"
func (s *StoragesHandler) GetDungeonNumberRecognitionErrorCountIn1Hour(key string) int {
	timestampsStr := storages.Get("异常统计", key)
	if timestampsStr == "" {
		return 0
	}

	// 解析时间戳列表
	var timestamps []int64
	parts := strings.Split(timestampsStr, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if ts, err := strconv.ParseInt(part, 10, 64); err == nil {
			timestamps = append(timestamps, ts)
		}
	}

	// 过滤1小时内的记录
	timestamps = s.filterTimestampsIn1Hour(timestamps)

	// 清理并保存（移除过期记录）
	s.saveErrorTimestamps(key, timestamps)

	return len(timestamps)
}

// ClearDungeonNumberRecognitionError 清空指定类型的异常统计记录
// key: 异常类型键名，如 "副本识别不到数字"
func (s *StoragesHandler) ClearDungeonNumberRecognitionError(key string) {
	storages.Remove("异常统计", key)
}

// filterTimestampsIn1Hour 过滤1小时内的时间戳
func (s *StoragesHandler) filterTimestampsIn1Hour(timestamps []int64) []int64 {
	currentTime := time.Now().In(chinaLocation).Unix()
	cutoffTime := currentTime - 60*60 // 1小时前的时间戳

	var validTimestamps []int64
	for _, ts := range timestamps {
		if ts >= cutoffTime {
			validTimestamps = append(validTimestamps, ts)
		}
	}

	return validTimestamps
}

// saveErrorTimestamps 保存异常时间戳列表
func (s *StoragesHandler) saveErrorTimestamps(key string, timestamps []int64) {
	if len(timestamps) == 0 {
		// 如果没有有效记录，清空存储
		storages.Remove("异常统计", key)
		return
	}

	// 将时间戳列表转换为逗号分隔的字符串
	var parts []string
	for _, ts := range timestamps {
		parts = append(parts, strconv.FormatInt(ts, 10))
	}

	timestampsStr := strings.Join(parts, ",")
	storages.Put("异常统计", key, timestampsStr)
}

// GetStoragesInt 从 storages 获取字符串并转换为 int
// category: storages 的类别，如 "data"
// key: storages 的键名，如 "silver"
// 返回值: 如果不存在（空字符串）返回 -1，转换失败返回 0，否则返回转换后的值
// 使用示例: core.GetStoragesInt("data", "silver")
func GetStoragesInt(category, key string) int {
	value := storages.Get(category, key)
	// 如果为空字符串，表示不存在，返回 -1
	if value == "" {
		return -1
	}
	// 否则转换为 int（转换失败返回 0）
	if val, err := strconv.Atoi(strings.TrimSpace(value)); err == nil {
		return val
	}
	return 0
}

// DataPut 写入 storages 的 data 域（与 GetStoragesInt("data", key) 对应）。
func (s *StoragesHandler) DataPut(key, value string) {
	storages.Put("data", key, value)
}

// DataGet 读取 storages 的 data 域。
func (s *StoragesHandler) DataGet(key string) string {
	return storages.Get("data", key)
}
