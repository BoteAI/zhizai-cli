---
name: zhizai-ask
version: 1.0.0
description: |
  用 zhizai CLI 问小智语义问答（ask）：在全部笔记或指定笔记集/目录/笔记范围内提问、追问、带参考文档；动态模版成文（写周报/复盘）；对一段文字流式总结（summarize）；按标题时间查总结和录音（recordings，人资）。
  用户说找找关于、在某某笔记集里问、待办有哪些、追问、写周报、按模版总结本周会议、给这段话出个总结、查这个月经营会的总结和录音时使用。
  列表过滤（最近有哪些 / 标题带…）走 zhizai-note 的 notes；禁止用 notes --title 冒充语义搜索。
---

# 问小智与成文（ask / summarize / recordings）

机器调用加 `-o json`。雪花 ID 当字符串。语义搜索只走 `ask`。

## ask 执行约定

- `zhizai ask "问题" -o json`：对外一次闭环，**默认自动建会话**再流式问答；返回 `chatId` / `contextId` 供追问。
- 追问：已有 `chatId` 时 `zhizai ask "…" --chat-id <id> --context-id <id>`，不必再建会话；建议带同一批范围 ID。
- 负向：`--no-session` 且未提供 `--chat-id` 时失败「请先创建问小智会话」，不要偷偷建会话。
- 空问题：本地失败，不调命令。
- 范围内无内容（空笔记集/空目录）：如实返回错误原文，不编造检索结果。

## 限定范围（名称必须先查 ID）

范围优先级：目录 > 笔记集 > 笔记 > 个人全部笔记。`--dir` 需同时带所属 `--kb`。

| 用户说法 | 顺序 |
|---|---|
| 找找关于支付的笔记（不限范围） | 单步：`zhizai ask "找找关于支付的笔记" -o json`（自动建会话） |
| 在工作笔记集里问… | ① `zhizai kb list` 匹配得 `knowledgeId` ② `zhizai ask "…" --kb <knowledgeId> -o json` |
| 只问某目录 | ① `kb list` 得 `knowledgeId` ② `kb catalog --kb <id>` 匹配得 `directoryId` ③ `ask "…" --kb <knowledgeId> --dir <directoryId>`（检索该目录及子目录） |
| 就那条某某纪要，待办有哪些 | ① `notes --title …` 得 `noteId`（会话已有则跳过）② `ask "待办有哪些" --note <noteId>` |
| 这几条一起问 | 每条先查真实 ID，再 `ask "…" --note <id1> --note <id2>`（或 `--note-ids <id1,id2>` / 多个 `--kb`） |
| 带参考文档 | `ask "…" --with-references`；引用详情用 `zhizai ask refs --doc-ids 1,2 --context-id <id>` |

查不到名称则停；禁止把名称当 `--kb` / `--dir` / `--note`。需要显式分步时：`ask session [--kb/--dir/--note]` 先建会话，再 `ask chat "…" --chat-id <id> --context-id <id>`。

## 动态模版成文（写周报 / 复盘）

「按周报结构总结本周会议」走模版管线，**不是问小智，也不是某条笔记的 summary**：

1. `zhizai ask template "生成周报" --with-notes --start "2026-09-21 00:00:00" --end "2026-09-27 23:59:59" -o json`
2. CLI 只返回模版与笔记 JSON，**不调用 LLM 成文**；由你按模版结构填真实笔记。
3. 模版要求但笔记里没有的字段标「未提及」，禁止编造。

## summarize（文字流式总结）

用户给了一段文字要总结（不是某条笔记、不是问答）：

- `zhizai summarize --content "…" -o json`；点名场景时先查真实 `sceneId`（见 zhizai-scene）再加 `--scene-id`。
- 流式输出结束即完成。

## recordings（人资查询）

「查这个月经营会的总结和录音」：

- `zhizai recordings --title 经营 --start "2026-09-01 00:00:00" --end "2026-09-30 23:59:59" -o json`（`--from` / `--to` 为别名）。
- 返回笔记列表；用户要打开某一条时再 `note get <id>`（变成组合，见 zhizai-note）。
