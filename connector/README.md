# WorkBuddy 连接器（zhizai-cli-connector）

基于 [WorkBuddy CLI + Skill 接入规范](https://open.workbuddy.cn/docs/connector#cli-skill-%E6%8E%A5%E5%85%A5)，将官方 [`@zhizai/cli`](https://www.npmjs.com/package/@zhizai/cli) 封装为连接器。

本目录位于 **zhizai-cli 仓库内**，与 CLI **共用** `../skills/`（打包时拷入 zip，不在此目录重复维护）。

- 连接器标识（`source`）：`zhizai-cli-connector`
- 连接器版本：与根目录 `package.json` 对齐（打包脚本会同步写入 `connector-meta.json`）
- 市场图标：`icon.svg`

## 目录（仓库内）

```text
zhizai-cli/
├── connector/                 # 本目录：连接器元信息
│   ├── connector-meta.json
│   ├── cli.json
│   ├── icon.svg
│   └── README.md
├── skills/                    # 与 CLI / setup 共用的 Skill 源
└── scripts/pack-connector.sh  # 一键打 zip
```

## 一键打包

在仓库根目录：

```bash
make connector-zip
# 或
./scripts/pack-connector.sh
```

产物：`dist/zhizai-cli-connector-<version>.zip`，解压后顶层为 `zhizai-cli-connector/`，可直接提交 WorkBuddy 审核。

## 环境与端点（默认生产）

| 用途 | 生产环境（默认） | 灰度环境 |
|---|---|---|
| 业务 API | `https://openapi.zzjilu.com/api/v1` | `https://www.zzjilu.com:9001/api/v1` |
| OAuth | `https://www.zzjilu.com/server` | `https://www.zzjilu.com:9001/server` |
| 授权页 | `https://www.zzjilu.com/oauth/device?user_code=...` | `https://www.zzjilu.com:9001/oauth/device?...` |
| 切换 | 默认 | `ZHIZAI_DEV=1 ZHIZAI_ENV=gray` |

详见 `docs/development.md` 与 `internal/config/endpoints.go`。

## 认证设计（Device Flow）

需 WorkBuddy ≥ 5.0.0（`minWorkbuddyVersion` / `authDeviceFlow`）：

| 字段 | 值 |
|---|---|
| `auth` | `zhizai auth login --no-open` |
| `authWaitForExit` | `true` |
| `authSuppressBrowser` | `false` |
| `authUrlDomain` | `www.zzjilu.com` |
| `statusMatch` | `认证方式:` |

CLI `--no-open` 输出示例：

```text
[device_code] XXXX-XXXX-XXXX
[verify_url]  https://www.zzjilu.com/oauth/device?user_code=XXXX-XXXX-XXXX
[expires_in]  300
[interval]    5
```

正则见 `cli.json` 的 `authDeviceFlow`。凭证在本机 `~/.zhizai/config.json`。

## 提交前检查

- [x] 仅 CLI 方案，未混用 MCP
- [x] `source` kebab-case：`zhizai-cli-connector`
- [x] `type: "cli"`，中英文名称 / 描述 / 示例齐全
- [x] `cli.json` 含各平台 `init` / `auth` / `unAuth` / `status`
- [x] Skill 来自仓库 `skills/`，且含 `version` frontmatter
- [x] `minWorkbuddyVersion: "5.0.0"`

## 依赖的 CLI

- npm：`@zhizai/cli@latest`（与本仓库发版版本一致时体验最佳）
- 需支持 `auth login --no-open` 与机器可读 `[verify_url]` 输出
