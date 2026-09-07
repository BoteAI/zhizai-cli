# 设计：OAuth 设备授权登录

日期：2026-09-06  
状态：已实现（方案 1）

## 决策摘要

- 默认 `zhizai auth login` = 设备授权；`--api-key` 保留给脚本/CI
- 不传 `scope`，使用服务端客户端默认权限
- OAuth 与 API Key 凭证互斥
- 业务请求自动 refresh + `zhizai auth refresh`
- 服务基址集中在 `internal/config/endpoints.go`，默认 `dev`（lingxi）

## 关键文件

- `internal/config/endpoints.go` — `ServiceBaseURLs` / `ResolveAPIBaseURL`
- `internal/config/config.go` — OAuth 字段与互斥写入
- `internal/client/oauth.go` — authorize / poll / refresh
- `internal/client/client.go` — 双鉴权头与自动刷新
- `cmd/auth/auth.go` — 登录 UX

## 接口约定

见 `docs/zhizai-cli-设备授权认证接口文档.md`。
