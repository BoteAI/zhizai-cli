---
name: zhizai-knowledge
version: 2.0.0
description: |
  用 zhizai CLI 管理笔记集（knowledge，别名 kb）：列出我创建的/收到的笔记集、打开笔记集与目录、新建笔记集、建目录、把笔记移入目录、删目录。
  用户说我的笔记集、我收到的笔记集、打开某某笔记集、新建笔记集、建目录、把这条笔记放到某某笔记集、删掉目录时使用。
  名称必须先查 ID 再操作；保存时归档（save --kb）见 zhizai-note。
---

# 智在记录笔记集（kb）

`knowledge` 别名 `kb`，以下统一写 `kb`。机器调用加 `-o json`。雪花 ID 当字符串。

## 名称 → ID（写操作前必须完成）

| 用户说的 | 先执行 | 得到 |
|---|---|---|
| 笔记集名称 | `zhizai kb list -o json`（别人分享的再 `kb shared`），按名称匹配 | `knowledgeId` |
| 目录名称 | 先有 `knowledgeId`，再 `zhizai kb catalog --kb <knowledgeId> -o json`，按目录名匹配 | `directoryId`（即 `catalogId`） |

零条停、多条让用户选。禁止把名称当 ID、禁止编造 ID。会话里本轮已拿到的 ID 可直接用。

## 命令与组合

| 用户说法 | 执行 | 顺序 |
|---|---|---|
| 我的笔记集 | 单步 | `zhizai kb list -o json`。空列表是成功 |
| 我收到的笔记集 | 单步 | `zhizai kb shared -o json`（等价 `kb list --received`） |
| 打开某某笔记集 | 组合 2 步 | ① `kb list` 匹配名称 ② `zhizai kb get <knowledgeId>` 或 `kb notes <knowledgeId>` |
| 看某目录里的笔记 | 组合 3 步 | ① `kb list` ② `kb catalog --kb <id>` 匹配目录 ③ `kb notes <id> --dir <directoryId>` |
| 新建笔记集 | 单步 | `zhizai kb create --name "工作笔记集" [--detail "…"]`，返回新 `knowledgeId` |
| 建目录 | 组合 2 步 | ① `kb list` 匹配笔记集 ② `zhizai kb mkdir --kb <knowledgeId> --name "三季度会议" [--parent <父目录Id>]`；不传 `--parent` 或传 `-1` 表示根层。父目录也是名称时先 `kb catalog` 匹配 |
| 移笔记到目录 | 组合 3～4 步 | ① 查 `noteId`（会话没有则 `notes --title`）② `kb list` 得 `knowledgeId` ③ 点目录则 `kb catalog` 得 `directoryId` ④ `zhizai kb add --kb <knowledgeId> --note-id <noteId> [--dir <directoryId>]` |
| 批量移笔记 | 组合 3～4 步 | 每条笔记都先查到真实 `noteId`，再一次 `kb add --kb <id> --note-ids <id1,id2> [--dir …]`；指定目录时只能对着一个笔记集 |
| 移到笔记集根层 | 组合 2～3 步 | ① 查 `noteId` ② 查 `knowledgeId` ③ `kb add --kb <knowledgeId> --note-id <noteId>`（不传 `--dir`） |
| 删目录 | 组合 2～3 步 | ① 不知在哪个集则 `kb list` 再 `kb catalog --kb <id>` 匹配目录 ② 用户确认 ③ `zhizai kb rmdir <directoryId> --yes` |

## 边界

- `kb add`：未入集则新增关联；已在集内则改目录；不删笔记正文。已在目标目录可能失败「没有需要添加的笔记」。
- `kb rmdir`：必须 `--yes`；级联去掉子目录与笔记**关联**，不删笔记正文。未确认不删。
- 目录不存在、父目录无效：如实报「目录不存在」类错误，不假装已建/已删。
- 保存时归档用 `save --kb [--dir]`（见 zhizai-note）；`--kb` 无效时笔记可能已创建但未归档，须拆开说「已创建但未归档」。
