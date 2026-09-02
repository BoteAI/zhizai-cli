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

```bash
# 直接传入 API Key（推荐脚本 / CI）
zhizai auth login --api-key <your-api-key>

# 或交互粘贴
zhizai auth login
```

登录成功后会写入 `~/.zhizai/config.json`，并做一次接口探活。

也可用环境变量（优先级高于配置文件）：

```bash
export ZHIZAI_REC_API_KEY=<your-api-key>
```

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
| `zhizai auth login [--api-key <key>]` | 保存 API Key 并验证连接 |
| `zhizai auth status` | 查看认证状态（Key 掩码显示） |
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
| `ZHIZAI_API_URL` | 覆盖 API 基址（默认 `https://openapi.zzjilu.com/api/v1`） |

### 鉴权说明

当前仅支持 **API Key**：

1. 在[开发者中心](https://www.zzjilu.com/pc/developer)获取 Key，或由管理员下发
2. `zhizai auth login --api-key ...` 保存并探活
3. 后续请求自动带 `Authorization: <api-key>`

OAuth 浏览器授权计划在后续版本补齐，不影响当前 API Key 流程。

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
