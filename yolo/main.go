package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// 直接在这里填写怪物名称；留空则启动时手动输入。
const presetMonsterName = "测试"

// 截图保存根目录：其下为 image/<怪物名>/...
const imageOutputRoot = "C:\\Users\\70511\\Downloads\\训练图片"

// adbSerial 用于指定要连接的目标设备串号（例如 192.168.31.48:5555）
var adbSerial string

func main() {
	projectRoot, err := resolveProjectRoot()
	if err != nil {
		fmt.Printf("定位项目目录失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("==== YOLO 采集工具（Go版） ====")
	fmt.Printf("项目目录: %s\n", projectRoot)
	fmt.Printf("截图根目录: %s\n", imageOutputRoot)
	fmt.Println("按 Ctrl+C 停止采集")
	fmt.Println()

	monsterName, err := getMonsterName()
	if err != nil {
		fmt.Printf("读取怪物名称失败: %v\n", err)
		os.Exit(1)
	}

	if err := ensureADBReady(); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	targetDir := filepath.Join(imageOutputRoot, "image", monsterName)
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		fmt.Printf("创建目录失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n开始采集，保存目录: %s\n", targetDir)

	for {
		fileName := fmt.Sprintf("%s.png", time.Now().Format("20060102_150405"))
		filePath := filepath.Join(targetDir, fileName)

		if err := screenshotToFile(filePath); err != nil {
			fmt.Printf("[%s] 截图失败: %v，1秒后重试...\n", time.Now().Format("15:04:05"), err)
			time.Sleep(1 * time.Second)
			continue
		}

		fmt.Printf("[%s] 已保存: %s\n", time.Now().Format("15:04:05"), filePath)
		time.Sleep(10 * time.Second)
	}
}

func resolveProjectRoot() (string, error) {
	exePath, err := os.Executable()
	if err != nil {
		return "", err
	}
	exeDir := filepath.Dir(exePath)

	// 支持 go run 与 go build 后执行。
	if strings.Contains(exeDir, os.TempDir()) {
		cwd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		return cwd, nil
	}
	return filepath.Clean(filepath.Join(exeDir, "..", "..")), nil
}

func getMonsterName() (string, error) {
	if strings.TrimSpace(presetMonsterName) != "" {
		return strings.TrimSpace(presetMonsterName), nil
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("请输入怪物名称（例如：绿蘑菇）: ")
	name, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return "", fmt.Errorf("怪物名称不能为空")
	}
	return name, nil
}

func ensureADBReady() error {
	if _, err := exec.LookPath("adb"); err != nil {
		return fmt.Errorf("未检测到 adb，请先安装并配置到 PATH")
	}

	// 你只希望使用的设备（通常 tcp 设备会显示为 "192.168.31.48:5555"）
	const preferredHost = "192.168.31.48"

	serial, devicesList, err := pickADBSerial(preferredHost)
	if err != nil {
		return err
	}
	adbSerial = serial

	cmd := exec.Command("adb", "-s", adbSerial, "get-state")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("未能与目标设备 %q 通信: %s（adb devices=%s）", adbSerial, strings.TrimSpace(string(out)), devicesList)
	}
	return nil
}

// pickADBSerial 从 adb devices 中选择一个最合适的串号。
// preferredHost 用来匹配期望的设备（例如 "192.168.31.48" 会匹配 "192.168.31.48:5555"）。
func pickADBSerial(preferredHost string) (serial string, devicesList string, err error) {
	cmd := exec.Command("adb", "devices")
	out, cmdErr := cmd.CombinedOutput()
	if cmdErr != nil {
		return "", "", fmt.Errorf("执行 adb devices 失败: %s", strings.TrimSpace(string(out)))
	}

	devicesList = strings.TrimSpace(string(out))

	lines := strings.Split(devicesList, "\n")
	type dev struct {
		serial string
		state  string
	}
	var devices []dev
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// 跳过 header：List of devices attached
		if strings.HasPrefix(line, "List of devices attached") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		serial = fields[0]
		state := fields[1]
		devices = append(devices, dev{serial: serial, state: state})
	}

	if len(devices) == 0 {
		return "", devicesList, fmt.Errorf("未检测到设备，请先连接并开启 USB 调试: %s", devicesList)
	}

	// 先匹配 preferredHost 且在线（state=device）的设备
	var matches []dev
	for _, d := range devices {
		if preferredHost != "" && strings.Contains(d.serial, preferredHost) && d.state == "device" {
			matches = append(matches, d)
		}
	}
	if len(matches) == 0 {
		// 如果 preferredHost 没匹配到，就退回到“仅有一个在线设备”的情况
		var online []dev
		for _, d := range devices {
			if d.state == "device" {
				online = append(online, d)
			}
		}
		if len(online) == 1 && preferredHost != "" {
			return online[0].serial, devicesList, nil
		}
		return "", devicesList, fmt.Errorf("未找到可用的目标设备（preferred=%q）。当前 adb devices=%s", preferredHost, devicesList)
	}

	// 如果匹配到多个，优先精确包含 ':5555' 的（一般 tcpip），否则取第一个
	for _, d := range matches {
		if strings.Contains(d.serial, ":5555") {
			return d.serial, devicesList, nil
		}
	}
	return matches[0].serial, devicesList, nil
}

func screenshotToFile(filePath string) error {
	if strings.TrimSpace(adbSerial) == "" {
		return fmt.Errorf("adbSerial 为空：未调用 ensureADBReady() 或未选中目标设备")
	}
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	cmd := exec.Command("adb", "-s", adbSerial, "exec-out", "screencap", "-p")
	cmd.Stdout = file
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
