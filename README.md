# huahu-mxd

冒险岛韩服自动化脚本。**纯本地运行**：挂机逻辑、地图参数都在仓库里的 JSON 和 Go 代码中，不依赖云控、中控或远程下发配置。

## 快速开始

1. 复制 `assets/config/default.json.example` → `assets/config/default.json`
2. 去 [番茄 OCR 官网](https://52tomato.com/) 购买 **1000 口** 套餐（不贵），把 `license_key` 填进 `default.json` 的 `ocr` 段
3. 准备 **AutoGo 1.14.2**（见下方说明），在 `go.mod` 里 `replace` 到本地 `./AutoGo`
4. 准备 `resources/` 原生库（本地编译用，不上传 GitHub，需自行从构建环境拷贝）
5. 改 `main.go` 里的默认地图，或按 UI 写死地图名
6. 编译运行，建议先用 **抢夺宝物岛** 跑通流程

默认测试地图（`main.go` 已写死）：

```go
挂机地图 := "韩服抢夺宝物岛"
// 对应 assets/config/farm_map_韩服抢夺宝物岛.json
```

## 番茄 OCR（必配）

OCR 使用 **番茄 OCR（TomatoOCR）**，代码在 `TomatoOCR/`、`core/ocr.go`。

- Token 即 `default.json` → `ocr.license_key`
- 自行去官网注册购买，**1000 口** 够用且便宜
- 未配置或过期会导致读字、读等级等功能失败

## AutoGo 版本

本项目按 **AutoGo 1.14.2** 开发和调试。

`go.mod` 示例：

```go
require github.com/Dasongzi1366/AutoGo v0.0.0-...
replace github.com/Dasongzi1366/AutoGo => ./AutoGo
```

`./AutoGo` 目录需自行放置对应版本源码（已在 `.gitignore`，不进仓库）。

可以用更新版本的 AutoGo，但 API、打包方式可能有变动，**需要你自己会改**（路径、依赖、JNI 库等）。不熟的话先用 **1.14.2** 最省事。

## 一切用本地脚本

| 本地已有 | 说明 |
|----------|------|
| `assets/config/farm_map_*.json` | 各地图挂机参数 |
| `assets/config/scene_map.json` | 场景相关 |
| `play/` | 各地图玩法实现 |
| `assets/img/` | 图色、模板图 |

**不需要云端**：不必接中控、不必拉远程配置、不必 APK 热更新。地图在 UI 或 `main.go` 写死名字即可。

挂机细节（攻击间隔、清包、巡逻等）优先改对应 `farm_map_*.json`，或改 `play/` 里具体地图文件。

## 云控 / 服务端代码（建议删掉）

仓库里还留着历史云控、中控相关代码，**纯本地用法用不上**。可以让 AI 帮你整体删除或 stub 掉，例如：

| 文件 / 模块 | 作用（可删） |
|-------------|--------------|
| `main.go` 里 `core.API.LoginAndSetup(...)` | 登录云控、拉远程配置 |
| `core/user.go` | 登录、`RefreshConfig`、`RoleUpdate` |
| `core/api.go` | HTTP 客户端、`GetConfig*` 读云端配置 |
| `core/question_bank.go` | 远程题库 |
| `core/answer_record.go` | 答题记录上报云端 |
| `core/sls.go` | 阿里云日志上报 |
| `util/oss.go`、`util/apk.go` | OSS 截图、APK 更新下载 |
| `util/config_checker.go` | 远程配置版本检查 |
| `play/farm_role_update.go` | 定时 `RoleUpdate` 推中控 |
| `job/读取数据.go`、`job/role_silver_loop.go` | 中控通讯、金币同步 |

**本地最小跑通** 大致只需：

- `play/` + `assets/config/farm_map_*.json` — 挂机
- `captcha/` — 验证码 / GM 巡逻（若需要）
- `TomatoOCR` + `ocr.license_key` — 识字
- `core/` 里与图色、 motion、opencv 相关的部分

删掉登录后，`main.go` 可直接进挂机，例如：

```go
func main() {
    // 本地模式：跳过 LoginAndSetup
    挂机地图 := "韩服抢夺宝物岛"
    go captcha.Run() // 不需要验证码检测可注释
    play.RunAutoFarm(fmt.Sprintf("config/farm_map_%s.json", 挂机地图))
}
```

`default.json` 里 **OSS、SLS、lemon_api** 等也是云端/第三方服务，纯本地可留空或让 AI 一并移除相关调用。

## 地图选择

地图名写死在 UI 或 `main.go`，规则：

```
assets/config/farm_map_<地图名>.json
```

现有配置见 `assets/config/farm_map_*.json`。换地图只改字符串，例如 `"韩服研究所C1"`、`"land赫勒地区走路版"`。

## 构建

- Go 版本见 `go.mod`
- AutoGo 本地 `replace` 到 `./AutoGo`（**1.14.2** 推荐）
- `resources/` 需本地准备（`.gitignore` 已排除）

## 配置模板说明

`default.json.example` **不含** `server`、`download` 等云控字段。按需保留：

- **必填（OCR）**：`ocr.license_key`
- **可选**：`lemon_api`（GM 巡逻 AI）、`oss` / `sls`（上报类，纯本地可删代码后忽略）
