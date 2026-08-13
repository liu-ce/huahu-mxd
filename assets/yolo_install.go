package assets

import (
	"fmt"
	"os"
	"path/filepath"
)

// 与 yolo.New 入参一致：手机里常用可写目录下的固定文件名
const (
	// 只改这里一个名字，下面路径会自动同步
	YoloModelBaseName = "best_pnnx.py.ncnn"

	YoloDeviceParamPath = "/data/local/tmp/assets/" + YoloModelBaseName + ".param"
	YoloDeviceBinPath   = "/data/local/tmp/assets/" + YoloModelBaseName + ".bin"

	YoloEmbeddedParamPath = "yolo/" + YoloModelBaseName + ".param"
	YoloEmbeddedBinPath   = "yolo/" + YoloModelBaseName + ".bin"
)

// InstallYoloOnDevice 将 go:embed 的 assets/yolo/* 中的 param/bin 解压到 /data/local/tmp/，
// 返回传给 yolo.New 的 paramPath、binPath。标签由调用方写死传入 New，不再写文件。
func InstallYoloOnDevice() (paramPath, binPath string, err error) {
	paramPath = YoloDeviceParamPath
	binPath = YoloDeviceBinPath

	dir := filepath.Dir(paramPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", "", fmt.Errorf("创建目录 %s: %w", dir, err)
	}

	// 与 assets/yolo 目录下 ncnn 导出文件名一致（内容仍解压为 /data/local/tmp/param、bin）
	if err := writeOneEmbedded(YoloEmbeddedParamPath, paramPath); err != nil {
		return "", "", err
	}
	if err := writeOneEmbedded(YoloEmbeddedBinPath, binPath); err != nil {
		return "", "", err
	}

	return paramPath, binPath, nil
}

func writeOneEmbedded(name, dst string) error {
	b, err := YoloFile.ReadFile(name)
	if err != nil {
		return fmt.Errorf("读取嵌入 %s: %w（请确认已放入 assets/yolo/ 对应文件）", name, err)
	}
	if len(b) == 0 {
		return fmt.Errorf("嵌入文件 %s 长度为 0，编译时 assets/yolo 里是否缺文件或用空文件占位？", name)
	}
	if err := os.WriteFile(dst, b, 0o644); err != nil {
		return fmt.Errorf("写入 %s: %w", dst, err)
	}
	st, err := os.Stat(dst)
	if err != nil {
		return fmt.Errorf("写入后 stat %s: %w", dst, err)
	}
	if st.Size() == 0 {
		return fmt.Errorf("写入后 %s 仍为 0 字节", dst)
	}
	return nil
}
