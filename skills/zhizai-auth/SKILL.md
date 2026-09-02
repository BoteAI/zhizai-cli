---
name: zhizai-auth
description: 安装和连接智在记录，完成 API Key 配置、环境诊断与 CLI 升级。用户说安装、连接、登录、检查为什么不能用时使用。
---

# 智在记录连接与诊断

通过官方 `zhizai` CLI 完成真实操作。机器调用优先使用 `-o json`。

## 路由

| 意图 | 命令 |
|---|---|
| 登录 | `zhizai auth login --api-key <key>` |
| 查看状态 | `zhizai auth status` |
| 退出 | `zhizai auth logout` |
| 诊断 | `zhizai doctor -o json` |
| 能力契约 | `zhizai capabilities -o json` |
| 安装 Skill | `zhizai setup` |
