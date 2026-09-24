---
name: zhizai-note
version: 2.0.0
description: |
  用 zhizai CLI 做笔记：把录音/文档/图片/链接/文字做成笔记（save 自动判型）、只上传文件、列表过滤、打开详情/纪要/原文、改标题/摘要/总结、删除、等处理完成、下载音频。
  用户说做成笔记、记一条、收藏链接、上传文件、最近有哪些、打开这条、只要转写/纪要、改标题、删掉、等它处理完、下载音频时使用。
  语义搜索「找找关于…」走 zhizai-ask；笔记集归档走 zhizai-knowledge；场景名查 ID 走 zhizai-scene。
---

# 智在记录笔记（save / file / notes / note）

只通过 `zhizai` CLI 操作，禁止直调开放 API。机器调用一律加 `-o json`。雪花 ID 全程当字符串传递。

## 执行约定（必须先读）

| 标记 | 怎么做 |
|---|---|
| 单步 | 一条 CLI 闭环（如 `notes`、`file upload`）。 |
| 单步（内部多接口） | 对外仍一条 `save`：内部上传+创建，`--wait` 再轮询。**不要**拆成 `file upload` + 创建 + `note wait` 三步。 |
| 组合 N 步 | 必须按 ①②③ 做完，缺一步即错；中途失败即停，不跳步、不猜 ID。 |
| 附加参数 | 不能单独执行，必须并进同一条 `save`。 |

**会话里已有的 ID 才能跳过查询。** 用户只说名称/标题而会话没有对应 ID，必须先查。禁止把名称当 ID，禁止编造雪花 ID。

名称 → ID：

| 用户说的 | 先执行 | 得到 |
|---|---|---|
| 笔记 / 「这条纪要」/ 标题 | `zhizai notes --title … -o json`，零条停、多条让用户选 | `noteId` |
| 笔记集名称 | `zhizai kb list`（别人分享的再 `kb shared`），按名称匹配 | `knowledgeId`（见 zhizai-knowledge） |
| 目录名称 | 先有 `knowledgeId`，再 `zhizai kb catalog --kb <id>` 按名称匹配 | `directoryId` |
| 总结场景名称 | `zhizai scene list` 或 `scene builtins`，按名称匹配 | `sceneId`（见 zhizai-scene） |

创建成功 ≠ 转写/总结完成。`save` 成功最低标准是非空 `noteId`；只有 `completed`（或用户明确说不用等）才能宣称完成。

## 意图分流（不要混用）

| 用户说法 | 走哪条 | 不要走 |
|---|---|---|
| 最近有哪些 / 我的录音 / 标题带经营分析 / 这个月的 | `notes` 列表过滤 | ask（列表过滤不是语义搜索） |
| 找找关于… / 待办有哪些（语义问答） | **zhizai-ask** 的 `ask` | 用 `notes --title` 冒充语义搜索 |
| 做成笔记 / 记一条 / 收藏链接 / 把文件生成笔记 | `save` 一条闭环 | 只 `file upload` 就停；追问「这是录音还是文档」 |
| 只上传、不做成笔记 | `file upload` | `save` |
| 打开这条 / 看纪要 / 只要原文 | 无 ID 先 `notes --title` 选定，再 `note get` | 把 summary 当原文 |
| 改标题/摘要/总结、删除、下载音频 | 无 ID 先查，再 `note update/delete/audio` | 没有 `noteId` 就调用 |
| 等它处理完 | `note wait`，仅 `completed` 后再 `note get` | 失败自动再 `save` |
| 给这段话出个总结（用户给了文字） | `zhizai summarize`（见 zhizai-ask） | ask；某条笔记的 `--field summary` |

## save 判型（用户没点名类型时必须用这套，不要追问）

| 用户给了什么 | 判定 | 命令 |
|---|---|---|
| 音频 mp3/wav/m4a/aac/ogg/flac/amr/opus | 录音 | `save <file> --wait`；途径默认 offlineImport，不要追问 |
| 图片 jpg/jpeg/png/gif/webp/bmp/heic | 图片 | 多图合成一条：`save a.jpg b.png --wait` |
| 文档 pdf/doc/docx/ppt/pptx/xls/xlsx/txt/md | 文档 | `save <file> --wait`；`--title` 优先于文件名 |
| `http(s)://` URL | 链接 | `save <url>` |
| 只有文字 | 文字 | `save --title … --content …`；长文用 `--content-file` / `--stdin` |
| 一份音频 + 若干图片且说「附图/一起保存」 | 录音带附图 | `save a.mp3 --image p1.jpg --image p2.png --wait` |
| 视频 mp4/mov/avi/mkv/webm/m4v | **不能做成笔记** | 停并说明；用户只要「上传」则 `file upload`。禁止当成录音 |
| 无法识别的扩展名 | 不创建 | 列出可支持格式（录音/文档/图片/链接/文字），不猜测 |

录音/图片/文档默认 `--wait`（已默认开启）；用户说不用等则 `--no-wait`，立即返回 `noteId` 与 `note_state`，之后要结果再 `note wait`。

## save 附加参数（并进同一条 save，不单独执行）

| 用户说法 | 加进同一条 save |
|---|---|
| 从 14:00 录到 14:30，时长 30 分钟 | `--start "2026-09-18 14:00:00" --end "2026-09-18 14:30:00" --duration 1800`（**录制**时间，不是会议排期） |
| 手机内录 / 实时录音 / 录音卡 | `--source phoneInternal\|realtime\|recordingCard`（仅用户明确指定才传，默认 offlineImport） |
| 定位在广州 | `--lat 23.1291 --lng 113.2644` |
| 再记一句备注 | `--text "会后跟进渠道下沉"`（随手备注，不等于 AI 总结） |
| 图片备注 | `--remark "白板"` |
| 已有 fileId 跳过上传 | `--file-id <fileId>`（单文件类型） |
| 用某某场景 | 先查 `sceneId`，再加 `--scene-id <id>` |
| 放到某笔记集/目录 | 先查 ID，再加 `--kb <knowledgeId> [--dir <directoryId>]` |
| 追加到刚才那条会 | 先定 `noteId`，再加 `--append-to <noteId>` |

## 哪些必须组合（缺步即错）

| 用户说法 | 顺序 |
|---|---|
| 用会议纪要场景做笔记 | ① `scene list`/`scene builtins` 查 `sceneId` ② `save … --scene-id --wait` |
| 放到工作笔记集再保存 | ① `kb list` 查 `knowledgeId`（② 点目录再 `kb catalog` 查 `directoryId`）③ `save … --kb [--dir] --wait` |
| 打开/纪要/原文/短链/改/删/下音频（没给 ID） | ① `notes --title` 选定 ② `note get/update/delete/audio …` |
| 改标题后再确认 | ① `note update <id> --title …` ② `note get <id>` 读回 |
| 追加到刚才那条会（无 noteId） | ① `notes --title` 选定 ② `save … --append-to <id> --wait`（不要另存新笔记） |
| 上次那条还在转，好了叫我 | ① `note wait <id>` ② 仅 `completed` 后 `note get <id>` |

## 命令速查

| 意图 | 命令 |
|---|---|
| 做成笔记 | `zhizai save <文件\|URL>... [--title …] [--wait] -o json` |
| 文字笔记 | `zhizai save --title … --content …`（或 `note create --title … --content …`，仅 text） |
| 只上传 | `zhizai file upload <path> [--compress] -o json`（音频不建议 `--compress`） |
| 按 fileId 下载 | `zhizai file get <fileId> --out <path>`（`-o` 是输出格式，下载路径用 `--out`） |
| 列表 | `zhizai notes [--title/--type/--abstract/--summary/--content …] [--from … --to …] [--page --limit \| --all] [--with-content] [--with-short-url] -o json` |
| 详情 | `zhizai note get <id> [--field content\|summary\|abstract\|…] [--short-url] [--with-append]` |
| 进度 / 等待 | `zhizai note status <id>` / `zhizai note wait <id>` |
| 改 | `zhizai note update <id> [--title --abstract --summary]`（只改传入字段） |
| 删 | `zhizai note delete <id> --yes` |
| 下载录音 | `zhizai note audio <id> --out ./a.mp3`（非录音笔记会失败；只有 fileId 走 `file get`） |

## 错误边界

- 本地路径不存在：不执行 `save` / `file upload`。
- 空列表是成功：「没有符合条件的笔记」。
- 删除、覆盖性修改先确认；未 `--yes` 不删。无权限/不存在不能说已删。
- `note wait` 得到 `failed`：展示状态与原因，**禁止自动再 `save`**。
- `--kb` 无效：笔记可能已创建但未归档，必须拆开说「已创建但未归档」。
- 处理中（`recognizing`/`analyzing`）如实说明；不把 summary 冒充原文，不假装完成。
- 用户没要分享时，不主动加 `--short-url` / `--with-short-url`。
