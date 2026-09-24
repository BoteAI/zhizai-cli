---
name: zhizai-scene
version: 2.0.0
description: |
  用 zhizai CLI 查询总结场景与知识卡：内置场景、我的场景（含共享）、我创建的知识卡。
  用户说有哪些总结场景、我的场景、我的知识卡，或「用某某场景」做笔记/总结前需要查 sceneId 时使用。
---

# 智在记录场景与知识卡（scene / cards）

机器调用加 `-o json`。`sceneId` 当字符串。

## 命令

| 用户说法 | 命令 | 说明 |
|---|---|---|
| 有哪些总结场景 | `zhizai scene builtins -o json` | 内置/平台场景。**返回是按分类分组的 Map**，必须遍历各分类数组取 `id` / `scene_name`，不要当普通 list（否则得到 0 条） |
| 我的场景 | `zhizai scene list -o json` | 默认含共享场景（`--include-shared=false` 关闭）。空列表是成功 |
| 我的知识卡 | `zhizai cards [--page 1 --limit 10] -o json` | 等价 `zhizai scene cards`。分页列表，空列表是成功 |

## 名称 → sceneId

用户说「用会议纪要场景…」时，场景名不能直接用：

1. `zhizai scene builtins -o json`（或 `scene list`）按名称匹配，得真实 `sceneId`。
2. 再把 `sceneId` 传给 `save --scene-id`（做成笔记，见 zhizai-note）或 `summarize --scene-id`（文字总结，见 zhizai-ask）。

禁止把场景名当 `--scene-id`；查不到则停，不要空传。无效场景会导致创建失败。
