# huahu-mxd

冒险岛韩服自动化脚本。

## 配置

复制 `assets/config/default.json.example` 为 `assets/config/default.json`，填入 OCR、OSS、SLS、AI 等密钥。

模板里**不包含**服务端地址和下载地址——本项目默认不依赖中控下发配置或 APK 热更新。

## 服务端代码说明（可自行裁剪）

仓库里仍保留了一部分历史服务端交互代码。**实际有用的通常只有两类：**

1. **卡密验证** — 登录鉴权（`core/user.go` 的 `Login` / `LoginAndSetup`）
2. **统计数据** — GM 巡逻答题上报（`core/answer_record.go`）、运行日志（`core/sls.go`，可选）

其余服务端能力可按需删除或停用，例如：

| 可删 / 可关 | 作用 |
|-------------|------|
| `util/apk.go` | APK 检查更新、下载安装 |
| `LoginAndSetup` 里的 `RefreshConfig`、`RefreshQuestionBank` | 从服务端拉角色配置、题库 |
| `core/api.go` 的 `GetConfig*` 系列 | 读服务端下发的挂机参数 |
| `play/farm_role_update.go`、`job/读取数据.go` | 中控通讯、`RoleUpdate` 上报 |
| `job/role_silver_loop.go` 中的 `RoleUpdate` | 定时向中控同步角色数据 |
| `util/config_checker.go` | 远程配置版本检查 |

若保留卡密验证，API 基址请在本地 `default.json` 自行添加，或直接写死在 `core/api.go` 的 `NewAPIClient`（原 `server.host` 字段已移出模板）。

挂机参数（攻击间隔、自动清包等）建议改为读本地 `farm_map_*.json` 或代码常量，不再走服务端配置。

## 地图选择

**建议在 UI 层写死地图**，不要从服务端读 `挂机地图`。

`main.go` 中把地图名写死即可，文件名规则：`assets/config/farm_map_<地图名>.json`：

```go
// 示例：UI 固定选「韩服研究所C1」
挂机地图 := "韩服研究所C1"
play.RunAutoFarm(fmt.Sprintf("config/farm_map_%s.json", 挂机地图))
```

现有地图配置见 `assets/config/farm_map_*.json`。

## 构建

依赖 AutoGo 本地 replace，见 `go.mod`。
