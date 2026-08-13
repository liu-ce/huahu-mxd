package assets

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
)

const (
	HuashuMouseModelCount   = 5
	HuashuMouseEmbedPrefix  = "huashu/mouse"
	HuashuMouseDeviceSubdir = "huashu/mouse"
)

//go:embed huashu/mouse/*
var HuashuMouseFile embed.FS

// InstallHuashuMouseOnDevice 将嵌入的 assets/huashu/mouse/* 解压到 /data/local/tmp/assets/huashu/mouse/。
func InstallHuashuMouseOnDevice() (dir string, err error) {
	dir = filepath.Join("/data/local/tmp/assets", HuashuMouseDeviceSubdir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("创建目录 %s: %w", dir, err)
	}
	for i := 1; i <= HuashuMouseModelCount; i++ {
		for _, ext := range []struct {
			src, dst string
		}{
			{fmt.Sprintf("%s/model%d.param", HuashuMouseEmbedPrefix, i), fmt.Sprintf("model%d.param", i)},
			{fmt.Sprintf("%s/model%d.bin", HuashuMouseEmbedPrefix, i), fmt.Sprintf("model%d.bin", i)},
			{fmt.Sprintf("%s/result%d.txt", HuashuMouseEmbedPrefix, i), fmt.Sprintf("result%d.txt", i)},
		} {
			if err := writeHuashuMouseEmbedded(ext.src, filepath.Join(dir, ext.dst)); err != nil {
				return "", err
			}
		}
	}
	return dir, nil
}

func writeHuashuMouseEmbedded(name, dst string) error {
	b, err := HuashuMouseFile.ReadFile(name)
	if err != nil {
		return fmt.Errorf("读取嵌入 %s: %w（请先复制模型到 assets/huashu/mouse/）", name, err)
	}
	if len(b) == 0 {
		return fmt.Errorf("嵌入文件 %s 为空", name)
	}
	if err := os.WriteFile(dst, b, 0o644); err != nil {
		return fmt.Errorf("写入 %s: %w", dst, err)
	}
	return nil
}
