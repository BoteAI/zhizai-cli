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
| 鉴权（当前） | OAuth 设备授权（默认）+ API Key；OAuth 业务头 `X-OAuth2-Access-Token` |
| 配置目录 | `~/.zhizai/config.json` |
| CI | GitHub Actions：多平台构建 → Release → `npm publish` |
| AI 接入 | `skills/` 原子 Skill + `zhizai setup` |

服务基址集中在 `internal/config/endpoints.go` 的 `EnvPresets`。

### 环境对照

OpenAPI（业务）与 OAuth2（设备授权）**分属不同 base**：OAuth 走 `.../server`，业务走 `.../api/v1`。

| 环境 | `ZHIZAI_ENV` | 官网 | OAuth base | OpenAPI base |
|------|--------------|------|------------|--------------|
| 测试 | `test`（`dev` / `lingxi` 别名） | https://lingxi.iwhalecloud.com/zzjl/ | `https://lingxi.iwhalecloud.com/zzjl/server` | `https://lingxi.iwhalecloud.com/zzjl/api/v1` |
| 灰度 | `gray`（`staging` / `canary`） | https://www.zzjilu.com:9001/pc/home | `https://www.zzjilu.com:9001/server` | `https://www.zzjilu.com:9001/api/v1`（临时；`openapi.zzjilu.com:9001` 待开通） |
| 生产 | `prod`（发布默认） | https://www.zzjilu.com/pc/home | `https://www.zzjilu.com/server` | `https://openapi.zzjilu.com/api/v1` |

示例路径：

- OAuth：`{OAuthBase}/oauth2/device/authorize`
- 业务：`{APIBase}/note/queryNoteList`

**发布包默认锁定生产环境**，忽略 `ZHIZAI_ENV` / `api_url` 等切换项。  
本地开发需显式打开开关后再切环境：

```bash
export ZHIZAI_DEV=1
export ZHIZAI_ENV=test   # 或 gray / prod
# 也可分别覆盖：
# export ZHIZAI_API_URL=...
# export ZHIZAI_OAUTH_URL=...
```

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
 skills/                      Skill 源码（setup 与连接器共用）
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
├── skills/                   # 原子 Skill + 开放平台参考（连接器共用）
├── connector/                # WorkBuddy 连接器元信息
├── bin/zhizai.js             # npm 启动器
├── scripts/
│   ├── postinstall.js        # npm 安装后下载二进制
│   ├── release.sh            # 打 tag 触发发版
│   ├── publish.sh            # 一键发布 CLI
│   └── pack-connector.sh     # 一键打连接器 zip
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

字段与接口细节以 `skills/zhizai-open-platform/references/` 为准。

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

本机只负责「升版本 + 推 tag」；真正的二进制与 npm 发包由 GitHub Actions 完成。

```text
make publish / ./scripts/publish.sh
        │
        ▼
  package.json 升版 → test/build → commit
        │
        ▼
  push master + tag vX.Y.Z
        │
        ▼
  GitHub Actions (.github/workflows/release.yml)
        ├── 交叉编译多平台二进制
        ├── 创建 GitHub Release（含 checksums）
        └── npm publish @zhizai/cli@X.Y.Z
        │
        ▼
  用户: npm install -g @zhizai/cli
        └── postinstall 从 Release 拉对应平台包
```

**仓库与 Release 须公开**，否则 `postinstall` 下载会 404。

资产命名约定（与 `scripts/postinstall.js` 一致）：

```text
zhizai-cli_{version}_{darwin|linux|windows}_{amd64|arm64}.tar.gz
zhizai-cli_{version}_windows_amd64.zip
```

### 前置条件

| 项 | 说明 |
|----|------|
| 分支 | 在 `master` / `main`，工作区干净（未提交改动先 commit） |
| 权限 | 对本仓库有 push 权限 |
| npm 组织 | `@zhizai` 作用域有发布权限 |
| GitHub Secret | `NPM_TOKEN`（Actions → Secrets） |
| 本地 `.npmrc` | 仅本机调试用，**勿提交**（见 `.npmrc.example`） |

### 一键发布（推荐）

```bash
# 默认 patch：0.0.5 → 0.0.6，会先询问确认
make publish
# 或
./scripts/publish.sh
npm run publish:cli

# 跳过确认
make publish YES=1
./scripts/publish.sh -y

# 指定版本 / minor / major
make publish V=0.0.6
make publish V=minor YES=1
./scripts/publish.sh --dry-run          # 只演练，不推送
```

底层仍调用 `scripts/release.sh`（也可用 `make release V=patch`）。

脚本会：跑测试 → 构建 → 提交版本变更 → 推送分支与 tag。  
进度：https://github.com/BoteAI/zhizai-cli/actions  

发版后验证：

```bash
npm view @zhizai/cli version
npm install -g @zhizai/cli@latest
zhizai --version
```

---

## WorkBuddy 连接器

连接器元信息在 `connector/`，Skill **与 CLI 共用** `skills/`（每个 `SKILL.md` 需含 `version` frontmatter）。

`connector-meta.json` 的 `version` 与根目录 `package.json` **对齐**；`make release` / `pack-connector` 会自动同步。

### 一键打 zip

```bash
make connector-zip
# 或
npm run connector:zip
./scripts/pack-connector.sh
```

产物：`dist/zhizai-cli-connector-<version>.zip`（顶层目录名 `zhizai-cli-connector/`），可直接提交 WorkBuddy。

说明见 `connector/README.md`。

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
