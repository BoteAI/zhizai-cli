---
name: zhizai-note
description: 使用智在记录 CLI 查询、创建、更新和删除笔记，上传文件并查看处理进度。
---

# 智在记录笔记

通过官方 `zhizai` CLI 完成真实操作。参数与字段说明见仓库内 `skills/zhiji-open-platform/references/note.md`。

## 路由

| 意图 | 命令 |
|---|---|
| 笔记列表 | `zhizai notes -o json` |
| 笔记详情 | `zhizai note get <id> -o json` |
| 创建笔记 | `zhizai note create` |
| 更新笔记 | `zhizai note update` |
| 删除笔记 | `zhizai note delete` |
| 处理进度 | `zhizai note status <id> -o json` |
| 上传文件 | `zhizai file upload <path>` |
| 问答总结 | `zhizai ask "<问题>" -o json` |
