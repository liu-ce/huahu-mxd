package huashu

import (
	"app/TaiBaiYoloV5Ncnn"
	"app/assets"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/Dasongzi1366/AutoGo/files"
	"github.com/Dasongzi1366/AutoGo/images"
)

const defaultDetectThreads = 4

// ModelManager 管理 5 套滑鼠追踪 YOLO 模型（对应懒人精灵 yolo.lua）。
type ModelManager struct {
	mu          sync.Mutex
	detector    *TaiBaiYoloV5Ncnn.Detector
	modelsDir   string
	currentIdx  int
	initialized bool
}

var (
	defaultManager   *ModelManager
	defaultManagerMu sync.Mutex
)

// InitMouseModels 初始化模型目录并预加载 model5（与懒人精灵一致）。
func InitMouseModels() (*ModelManager, error) {
	defaultManagerMu.Lock()
	defer defaultManagerMu.Unlock()
	if defaultManager != nil && defaultManager.initialized {
		return defaultManager, nil
	}

	dir, err := resolveMouseModelsDir()
	if err != nil {
		return nil, err
	}
	m := &ModelManager{
		detector:  TaiBaiYoloV5Ncnn.NewDetector(),
		modelsDir: dir,
	}
	if err := m.LoadModel(5); err != nil {
		return nil, fmt.Errorf("加载 model5: %w", err)
	}
	m.initialized = true
	defaultManager = m
	return m, nil
}

// DefaultManager 返回已初始化的全局 ModelManager。
func DefaultManager() *ModelManager {
	return defaultManager
}

func resolveMouseModelsDir() (string, error) {
	if dir, err := assets.InstallHuashuMouseOnDevice(); err == nil {
		if hasMouseModelSet(dir, 5) {
			return dir, nil
		}
	}
	runtimeDir := filepath.Join(files.Path("./assets"), "huashu/mouse")
	if hasMouseModelSet(runtimeDir, 5) {
		return runtimeDir, nil
	}
	repoDir := filepath.Join("assets", "huashu", "mouse")
	if hasMouseModelSet(repoDir, 5) {
		return repoDir, nil
	}
	return "", errors.New("滑鼠模型未找到：请运行 huashu/copy_models.ps1 将 model1~5 复制到 assets/huashu/mouse/")
}

func hasMouseModelSet(dir string, idx int) bool {
	for _, name := range []string{
		filepath.Join(dir, fmt.Sprintf("model%d.param", idx)),
		filepath.Join(dir, fmt.Sprintf("model%d.bin", idx)),
		filepath.Join(dir, fmt.Sprintf("result%d.txt", idx)),
	} {
		if !files.Exists(name) {
			if strings.Contains(name, "result") {
				alt := strings.Replace(name, "result", "label", 1)
				if files.Exists(alt) {
					continue
				}
			}
			return false
		}
	}
	return true
}

func modelPaths(dir string, idx int) (label, param, bin string, err error) {
	if idx < 1 || idx > assets.HuashuMouseModelCount {
		return "", "", "", fmt.Errorf("无效模型索引 %d", idx)
	}
	param = filepath.Join(dir, fmt.Sprintf("model%d.param", idx))
	bin = filepath.Join(dir, fmt.Sprintf("model%d.bin", idx))
	label = filepath.Join(dir, fmt.Sprintf("result%d.txt", idx))
	if !files.Exists(label) {
		alt := filepath.Join(dir, fmt.Sprintf("label%d.txt", idx))
		if files.Exists(alt) {
			label = alt
		}
	}
	for _, p := range []string{label, param, bin} {
		if !files.Exists(p) {
			return "", "", "", fmt.Errorf("模型文件不存在: %s", p)
		}
	}
	return label, param, bin, nil
}

// LoadModel 加载指定索引模型（1~5）。
func (m *ModelManager) LoadModel(idx int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.currentIdx == idx {
		return nil
	}
	label, param, bin, err := modelPaths(m.modelsDir, idx)
	if err != nil {
		return err
	}
	if err := m.detector.LoadModel(label, param, bin, defaultDetectThreads); err != nil {
		return err
	}
	m.currentIdx = idx
	return nil
}

// DetectRegion 区域截图 YOLO 检测，坐标已换算为屏幕绝对位置。
func (m *ModelManager) DetectRegion(x1, y1, x2, y2 int, threshold, nmsThreshold float32, size int) ([]TaiBaiYoloV5Ncnn.DetectorResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.detector == nil || m.currentIdx == 0 {
		return nil, errors.New("detector 未初始化")
	}
	img := images.CaptureScreen(x1, y1, x2, y2, 0)
	if img == nil {
		return nil, errors.New("截图失败")
	}
	results, err := m.detector.Detect(img, threshold, nmsThreshold, size)
	if err != nil {
		return nil, err
	}
	for i := range results {
		results[i].X += x1
		results[i].Y += y1
	}
	return results, nil
}

// DetectScreen 全屏 YOLO 检测。
func (m *ModelManager) DetectScreen(threshold, nmsThreshold float32, size int) ([]TaiBaiYoloV5Ncnn.DetectorResult, error) {
	return m.DetectRegion(0, 0, 0, 0, threshold, nmsThreshold, size)
}

// Close 释放 detector。
func (m *ModelManager) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.detector != nil {
		m.detector.Close()
		m.detector = nil
	}
	m.initialized = false
	m.currentIdx = 0
}
