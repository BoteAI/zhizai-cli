# 智在记录 CLI

智在记录的命令行工具，让你在终端和 AI Agent 里直接管理笔记、笔记集、场景与团队。

查笔记、写总结、管知识库——一条命令搞定，支持脚本和 AI Agent 调用。

> **当前版本：v0.1.0**  
> 已实现：`auth` / `doctor` / `capabilities` / `version` / `notes` / `note get` / `setup`。  
> 其余业务命令（`note create` / `ask` / `team` 等）已挂载子命令树，实现中。

---

## 安装

### 从源码构建（推荐开发期）

需要 Go 1.21+：

```bash
git clone https://github.com/BoteAI/zhizai-cli.git
cd zhizai-cli
make build
make install
```

构建产物：

| 文件 | 说明 |
|------|------|
| `zhizai-cli` | `make build` 直接产物 |
| `bin/zhizai` | `make dev-link` 复制到 npm 启动器目录 |

安装后可用命令：`zhizai` 与短别名 `zz`（`make install` 会安装符号链接；npm 全局安装同样注册两个入口）。

### npm 全局安装（Release 后可用）

```bash
npm install -g @zhizai/cli@latest
zhizai auth login --api-key <your-api-key>
zhizai doctor -o json
```

`postinstall` 会下载对应平台的二进制；发布 Release 前请优先用源码构建。

---

## 使用要求

- 需要有效的智在记录 **API Key**
- 获取入口：[智在记录开发者](https://www.zzjilu.com/pc/developer)
- OpenAPI 基址：`https://openapi.zzjilu.com/api/v1`
- 请求头：`Authorization: <api-key>`（**不要**加 `Bearer`）
- 限流：最高约 **2 次/秒**

---

## 开始使用

### 登录

```bash
# 直接传入 API Key（推荐脚本/CI）
zhizai auth login --api-key <your-api-key>

# 或交互粘贴
zhizai auth login
```

登录成功后会写入 `~/.zhizai/config.json`，并调用 `queryNoteList` 做一次探活。

也可用环境变量（优先级高于配置文件）：

```bash
export ZHIZAI_REC_API_KEY=<your-api-key>
```

### 检查状态

```bash
zhizai auth status
zhizai doctor
zhizai doctor -o json
zhizai capabilities -o json
```

`doctor` 返回 `ready=true,status=ready` 表示可正常调用业务接口。

### 查看笔记

```bash
zhizai notes
zhizai notes --limit 10 -o json
zhizai notes --all
zhizai note get <id>
zhizai note get <id> --field summary
```

### 接入本机 AI

```bash
# 预览将执行的操作
zhizai setup --dry-run -o json

# 正式安装（本地验收可跳过 npm 全局安装）
zhizai setup --skill-source . --skip-cli-install
```

---

## 命令一览

### 已可用

```
zhizai auth login [--api-key <key>]   保存 API Key 并验证连接
zhizai auth status                    查看认证状态（Key 掩码显示）
zhizai auth logout                    清除本机凭证
zhizai doctor                         检查安装、登录与 API 连通性
zhizai capabilities                   查看当前版本的稳定能力契约
zhizai version                        显示版本
zhizai notes [--limit|--page|--all]   笔记列表
zhizai note get <id> [--field ...]    笔记详情
zhizai setup [--dry-run]              为本机 AI 安装原子 Skill 并引导授权
```

### 规划中（子命令已挂载，业务逻辑待实现）

```
zhizai note create|update|delete|status
zhizai file upload                    文件上传
zhizai ask "<问题>"                   基于笔记的动态模版问答/总结
zhizai scene                          场景与知识卡
zhizai knowledge                      笔记集
zhizai team                           团队与成员
zhizai msg                            消息与录音卡
zhizai update                         升级 CLI 并同步 Skill
```

字段与接口细节以 `skills/zhiji-open-platform/references/` 为准。

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
| `GOPROXY` | 构建时 Go 模块代理，国内可用 `https://goproxy.cn,direct` |

---

## 鉴权说明

v0.1 仅支持 **API Key**：

1. 在开发者页获取 Key，或由管理员下发后自行配置
2. `zhizai auth login --api-key ...` 保存并探活
3. 后续请求自动带 `Authorization: <api-key>`

OAuth 浏览器授权计划在后续版本补齐，不影响当前 API Key 流程。

---

## AI Agent 使用

所有命令支持 `-o json`。内置原子 Skill 源在 `skills/`：

| Skill | 职责 |
|-------|------|
| `zhizai-auth` | 安装、登录、诊断、升级 |
| `zhizai-note` | 笔记 CRUD、上传、问答 |
| `zhizai-knowledge` | 笔记集 |
| `zhizai-scene` | 场景与知识卡 |
| `zhizai-team` | 团队与成员 |
| `zhizai-msg` | 消息与录音卡 |

另保留聚合 Skill `zhiji-open-platform`（含完整 OpenAPI 参考文档）。  
`zhizai setup` 就绪后，会把原子 Skill 安装到 Cursor、Claude Code、Codex 等本机 AI。

原则：**Skill 只做意图路由，真实请求一律走 `zhizai` CLI**，不自行拼 OpenAPI。

---

## 从源码开发

```bash
# 依赖
go version   # >= 1.21

# 构建
make build

# 本地链接到 bin/zhizai（配合 npm 启动器调试）
make dev-link

# 安装到 PATH
make install

# 测试 / 静态检查
make test
make lint

# 多平台交叉编译
make build-all
```

目录结构：

```text
zhizai-cli/
├── main.go
├── cmd/                 # Cobra 子命令
├── internal/
│   ├── client/          # OpenAPI HTTP 客户端（限流、resultCode 适配）
│   ├── config/          # ~/.zhizai/config.json
│   ├── output/          # 统一 JSON 输出
│   ├── platform/        # 本机 AI 平台探测
│   ├── ui/
│   └── version/
├── skills/              # 原子 Skill + 开放平台参考
├── bin/zhizai.js        # npm 启动器
├── scripts/postinstall.js
└── .github/workflows/release.yml
```

---

## 路线图

| 版本 | 内容 |
|------|------|
| **v0.1** | 脚手架、auth、doctor、capabilities（当前） |
| **v0.2** | notes / note 详情与 CRUD、file upload |
| **v0.3** | setup 安装 Skill、capabilities 完善 |
| **v0.4** | scene / knowledge / team / msg |
| **v0.5** | ask 动态模版总结管线 |
| **v1.0** | update、Release CI、npm 发布 |

---

## 相关链接

- [智在记录官网](https://www.zzjilu.com)
- [开发者中心](https://www.zzjilu.com/pc/developer)
- [问题反馈](https://github.com/BoteAI/zhizai-cli/issues)

## License

[MIT](./LICENSE)
