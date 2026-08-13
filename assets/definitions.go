package assets

import (
	"embed"
	"github.com/Dasongzi1366/AutoGo/opencv"
)

//go:embed img/*
var ImageFile embed.FS

//go:embed config/*
var ConfigFile embed.FS

//go:embed yolo/*
var YoloFile embed.FS

// 模板图像缓存映射
var TemplateMap = make(map[string]opencv.Mat)

// 遮罩图像缓存映射
var MaskMap = make(map[string]opencv.Mat)
