# 智在记录 CLI — 开发说明

本文面向维护者与贡献者：技术栈、功能架构、本地开发、构建与发包。

面向终端用户的安装与使用说明见仓库根目录 [README.md](../README.md)。

---

## 技术栈

| 层级 | 选型 |
|------|------|
| 语言 / CLI | Go 1.21+、[Cobra](https://github.com/spf13/cobra) |
| 分发 | npm 包 `@zhizai/cli` + `postinstall` 拉取 GitHub Release 二进制 |
| 命令入口 | `zhizai` / 短别名 `zz` |
| 鉴权（当前） | API Key（`Authorization: <key>`，不加 Bearer） |
| 配置目录 | `~/.zhizai/config.json` |
| CI | GitHub Actions：多平台构建 → Release → `npm publish` |
| AI 接入 | `skills/` 原子 Skill + `zhizai setup` |

OpenAPI 基址默认：`https://openapi.zzjilu.com/api/v1`，可用环境变量 `ZHIZAI_API_URL` 覆盖。

---

## 功能架构

```text
用户 / AI Agent
      │
      ▼
 zhizai / zz  (Cobra 命令树)
      │
      ├── auth / doctor / setup / capabilities / version
      ├── notes / note / file / ask / ...
      ▼
 internal/client   ──HTTP──▶  OpenAPI
 internal/config              ~/.zhizai
 internal/output              table / 统一 JSON
 internal/platform            本机 AI 探测
 skills/                      Skill 源码（setup 安装）
```

### 目录结构

```text
zhizai-cli/
├── main.go
├── cmd/                      # Cobra 子命令
├── internal/
│   ├── client/               # OpenAPI HTTP 客户端（重试、resultCode 适配）
│   ├── config/               # ~/.zhizai/config.json
│   ├── output/               # 统一 JSON / 表格输出
│   ├── platform/             # 本机 AI 平台探测
│   ├── ui/
│   └── version/
├── skills/                   # 原子 Skill + 开放平台参考
├── bin/zhizai.js             # npm 启动器
├── scripts/
│   ├── postinstall.js        # npm 安装后下载二进制
│   └── release.sh            # 打 tag 触发发版
├── docs/                     # 开发文档（本文）
└── .github/workflows/release.yml
```

### 能力分层

| 层 | 说明 |
|----|------|
| 命令层 `cmd/` | 参数解析、交互确认、调用 client / config |
| 客户端 `internal/client` | 鉴权头、限流友好错误、瞬时网络重试 |
| 输出 `internal/output` | Agent 可读的 `{success,data,error}` |
| Skill | 意图路由到 CLI，不直接拼 OpenAPI |

字段与接口细节以 `skills/zhiji-open-platform/references/` 为准。

---

## 本地开发

依赖：Go >= 1.21。国内拉模块可设：

```bash
export GOPROXY=https://goproxy.cn,direct
```

常用命令：

```bash
# 构建
make build

# 本地链接到 bin/zhizai（配合 npm 启动器调试）
make dev-link

# 安装到 PATH（同时创建 zz -> zhizai 符号链接）
make install

# 测试 / 静态检查
make test
make lint

# 多平台交叉编译
make build-all
```

构建产物：

| 文件 | 说明 |
|------|------|
| `zhizai-cli` | `make build` 直接产物 |
| `bin/zhizai` | `make dev-link` 复制到 npm 启动器目录 |
| `bin/zz` | `dev-link` / `install` 创建的短别名链接 |

从源码安装示例：

```bash
git clone https://github.com/BoteAI/zhizai-cli.git
cd zhizai-cli
make build
make install
```

---

## 发包流程

### 原理

1. 推送符合 `v*` 的 git tag  
2. GitHub Actions 交叉编译并上传 Release 资产（含 `checksums.txt`）  
3. 同一 workflow 执行 `npm publish --access public`  
4. 用户 `npm install -g @zhizai/cli` 时，`postinstall` 按版本从 Release 下载对应平台包  

**仓库与 Release 须公开**，否则 `postinstall` 下载会 404。

资产命名约定（与 `scripts/postinstall.js` 一致）：

```text
zhizai-cli_{version}_{darwin|linux|windows}_{amd64|arm64}.tar.gz
zhizai-cli_{version}_windows_amd64.zip
```

### 前置条件

| 项 | 说明 |
|----|------|
| npm 组织 | `@zhizai` 作用域有发布权限 |
| GitHub Secret | `NPM_TOKEN`（Actions → Secrets） |
| 本地 `.npmrc` | 仅本机调试用，**勿提交**（见 `.npmrc.example`） |

### 发版命令

```bash
# 使用 package.json 当前版本打 tag 并推送
make release
# 或
npm run release

# 升版本后再发（注意用 V=，不要用 VERSION=；VERSION 留给 go build ldflags）
make release V=patch
make release V=0.0.3
```

脚本会：跑测试 → 构建 →（如有）提交版本变更 → 推送分支与 tag。  
进度：https://github.com/BoteAI/zhizai-cli/actions  

发版后验证：

```bash
npm view @zhizai/cli version
npm install -g @zhizai/cli@latest
zhizai --version
```

---

## 路线图

| 版本 | 内容 |
|------|------|
| **v0.0.x** | 脚手架、auth、doctor、capabilities、notes / note get、setup、npm 发布 |
| **v0.2** | notes / note CRUD、file upload |
| **v0.3** | setup / capabilities 完善 |
| **v0.4** | scene / knowledge / team / msg |
| **v0.5** | ask 动态模版总结管线 |
| **v1.0** | update、稳定性与文档完善 |

---

## 相关链接

- [用户 README](../README.md)
- [智在记录开发者中心](https://www.zzjilu.com/pc/developer)
- [GitHub Releases](https://github.com/BoteAI/zhizai-cli/releases)
- [npm：@zhizai/cli](https://www.npmjs.com/package/@zhizai/cli)
