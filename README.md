# 智在记录 CLI

智在记录的命令行工具，让你在终端和 AI Agent 里直接管理笔记、笔记集、场景与团队。

查笔记、写总结、管知识库——一条命令搞定，支持脚本和 AI Agent 调用。

> **当前版本：v0.0.3**  
> 已可用：`auth` / `doctor` / `capabilities` / `version` / `notes` / `note get` / `setup`。  
> 其余业务命令（`note create` / `ask` / `team` 等）子命令已挂载，能力逐步完善中。

参与开发、构建与发包请见 [docs/development.md](./docs/development.md)。

---

## 安装

推荐使用 npm 全局安装（会自动下载对应平台二进制）：

```bash
npm install -g @zhizai/cli@latest
```

安装后可用命令：

- `zhizai` — 主命令
- `zz` — 短别名

验证：

```bash
zhizai version
zz version
```

也可从 [GitHub Releases](https://github.com/BoteAI/zhizai-cli/releases) 下载对应平台压缩包，解压后将 `zhizai` / `zhizai.exe` 放到 `PATH` 中。

---

## 使用要求

- 需要有效的智在记录 **API Key**
- 获取入口：[智在记录开发者](https://www.zzjilu.com/pc/developer)
- 请求头：`Authorization: <api-key>`（**不要**加 `Bearer`）
- 限流：最高约 **2 次/秒**

---

## 开始使用

### 1. 登录

默认走**网页设备授权**（打开浏览器确认）：

```bash
zhizai auth login
```

脚本 / CI 可继续用 API Key：

```bash
zhizai auth login --api-key <your-api-key>
# 或
export ZHIZAI_REC_API_KEY=<your-api-key>
```

登录成功后写入 `~/.zhizai/config.json`。OAuth 与 API Key **互斥**，后登录的会清掉另一种。

刷新 OAuth access_token（业务请求也会在过期时自动刷新）：

```bash
zhizai auth refresh
```

当前默认服务环境为 **dev**（lingxi）。切换方式见下方「服务环境」。

### 2. 检查状态

```bash
zhizai auth status
zhizai doctor
zhizai doctor -o json
zhizai capabilities -o json
```

`doctor` 返回 `ready=true,status=ready` 表示可正常调用业务接口。

### 3. 查看笔记

```bash
zhizai notes
zhizai notes --limit 10 -o json
zhizai notes --all
zhizai note get <id>
zhizai note get <id> --field summary
```

### 4. 接入本机 AI

```bash
# 预览将执行的操作
zhizai setup --dry-run -o json

# 正式安装 Skill
zhizai setup
```

`setup` 会把原子 Skill 安装到 Cursor、Claude Code、Codex 等本机 AI 环境，并引导完成授权。

---

## 功能一览

### 已可用

| 命令 | 说明 |
|------|------|
| `zhizai auth login` | 网页设备授权登录（默认） |
| `zhizai auth login --api-key <key>` | API Key 登录（脚本/CI） |
| `zhizai auth refresh` | 刷新 OAuth access_token |
| `zhizai auth status` | 查看认证状态（凭证掩码） |
| `zhizai auth logout` | 清除本机凭证 |
| `zhizai doctor` | 检查安装、登录与 API 连通性 |
| `zhizai capabilities` | 查看当前版本的稳定能力契约 |
| `zhizai version` | 显示版本 |
| `zhizai notes [--limit\|--page\|--all]` | 笔记列表 |
| `zhizai note get <id> [--field ...]` | 笔记详情 |
| `zhizai setup [--dry-run]` | 为本机 AI 安装原子 Skill 并引导授权 |

### 规划中

| 命令 | 说明 |
|------|------|
| `zhizai note create\|update\|delete\|status` | 笔记写入与状态 |
| `zhizai file upload` | 文件上传 |
| `zhizai ask "<问题>"` | 基于笔记的动态模版问答 / 总结 |
| `zhizai scene` | 场景与知识卡 |
| `zhizai knowledge` | 笔记集 |
| `zhizai team` | 团队与成员 |
| `zhizai msg` | 消息与录音卡 |
| `zhizai update` | 升级 CLI 并同步 Skill |

---

## 服务环境

业务 API 与 OAuth 可使用不同基址。预设在 `internal/config/endpoints.go` 的 `EnvPresets`：

| 环境 | 业务基址 | OAuth 基址 |
|------|----------|------------|
| `dev`（默认） | `https://lingxi.iwhalecloud.com/LCDP-RECORD/api/v1` | `https://lingxi.iwhalecloud.com/zzjl/server` |
| `test` | `https://lingxi.iwhalecloud.com/zzjl/api/v1` | `https://lingxi.iwhalecloud.com/zzjl/server` |
| `prod` | `https://openapi.zzjilu.com/api/v1` | 与业务共用 |

优先级：`ZHIZAI_API_URL` / `ZHIZAI_OAUTH_URL` > `config.json` 的 `api_url` / `oauth_url` > `ZHIZAI_ENV` / `config.env` > 默认 `dev`。

LCDP 直连基址（若手动配置且以 `/note` 结尾）拼接 `/note/...` 时会自动去掉重复的 `/note`。

```bash
# 切到测试环境
export ZHIZAI_ENV=test

# 切到生产
export ZHIZAI_ENV=prod

# 或分别覆盖业务 / OAuth 基址
export ZHIZAI_API_URL=https://lingxi.iwhalecloud.com/zzjl/api/v1
export ZHIZAI_OAUTH_URL=https://lingxi.iwhalecloud.com/zzjl/server
```

---

## 全局参数

| 参数 | 说明 |
|------|------|
| `--api-key <key>` | 临时覆盖 API Key |
| `-o, --output table\|json` | 输出格式（默认 `table`） |

机器 / AI Agent 调用时请加 `-o json`，统一读取：

```json
{
  "success": true,
  "data": {},
  "error": null
}
```

失败时 `success=false`，读取 `error.code` / `error.message` / `error.reason` / `error.retryable`。

---

## 配置

凭证保存在 `~/.zhizai/config.json`：

```json
{
  "api_key": "xxxx",
  "expires_at": "2027-12-31 23:59:59",
  "team_id": ""
}
```

凭证优先级：

```text
--api-key  >  环境变量 ZHIZAI_REC_API_KEY  >  ~/.zhizai/config.json
```

| 环境变量 | 说明 |
|----------|------|
| `ZHIZAI_REC_API_KEY` | API Key |
| `ZHIZAI_API_URL` | 覆盖业务 API 基址 |
| `ZHIZAI_OAUTH_URL` | 覆盖 OAuth 基址（authorize/token/refresh） |

### 鉴权说明

支持两种互斥模式：

1. **OAuth 设备授权（默认）**：`zhizai auth login` → 浏览器确认 → 保存 `access_token` / `refresh_token`；业务请求头 `X-OAuth2-Access-Token: Bearer …`
2. **API Key**：`zhizai auth login --api-key …` 或环境变量 `ZHIZAI_REC_API_KEY`；请求头 `Authorization: <api-key>`（不加 Bearer）

OAuth 过期时 CLI 会自动用 `refresh_token` 刷新；也可手动 `zhizai auth refresh`。

---

## AI Agent 使用

所有命令支持 `-o json`。内置原子 Skill：

| Skill | 职责 |
|-------|------|
| `zhizai-auth` | 安装、登录、诊断、升级 |
| `zhizai-note` | 笔记 CRUD、上传、问答 |
| `zhizai-knowledge` | 笔记集 |
| `zhizai-scene` | 场景与知识卡 |
| `zhizai-team` | 团队与成员 |
| `zhizai-msg` | 消息与录音卡 |

另保留聚合 Skill `zhiji-open-platform`（含完整 OpenAPI 参考文档）。

原则：**Skill 只做意图路由，真实请求一律走 `zhizai` CLI**，不自行拼 OpenAPI。

---

## 相关链接

- [智在记录官网](https://www.zzjilu.com)
- [开发者中心](https://www.zzjilu.com/pc/developer)
- [开发与发包说明](./docs/development.md)
- [问题反馈](https://github.com/BoteAI/zhizai-cli/issues)

## License

[MIT](./LICENSE)
