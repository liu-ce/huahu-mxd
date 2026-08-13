# 从懒人精灵 longwang 项目复制 YOLO 模型到 assets/huashu/mouse/
param(
    [Parameter(Mandatory = $true)]
    [string]$Source
)

$ErrorActionPreference = "Stop"
$dest = Join-Path $PSScriptRoot "..\assets\huashu\mouse"
New-Item -ItemType Directory -Force -Path $dest | Out-Null

for ($i = 1; $i -le 5; $i++) {
    foreach ($ext in @("param", "bin")) {
        $name = "model$i.$ext"
        $srcFile = Join-Path $Source $name
        if (-not (Test-Path $srcFile)) {
            Write-Error "缺少文件: $srcFile"
        }
        Copy-Item -Force $srcFile (Join-Path $dest $name)
        Write-Host "OK $name"
    }
    foreach ($label in @("result$i.txt", "label$i.txt", "result$i")) {
        $srcFile = Join-Path $Source $label
        if (Test-Path $srcFile) {
            Copy-Item -Force $srcFile (Join-Path $dest "result$i.txt")
            Write-Host "OK result$i.txt (from $label)"
            break
        }
    }
    if (-not (Test-Path (Join-Path $dest "result$i.txt"))) {
        Write-Warning "未找到 result$i.txt，请手动从 longwang.rc 解压目录复制"
    }
}

Write-Host "完成 → $dest"
