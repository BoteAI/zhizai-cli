package save

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/BoteAI/zhizai-cli/internal/client"
	"github.com/BoteAI/zhizai-cli/internal/output"
	"github.com/BoteAI/zhizai-cli/internal/ui"
	"github.com/spf13/cobra"
)

var (
	audioExts = map[string]struct{}{
		".mp3": {}, ".wav": {}, ".m4a": {}, ".aac": {}, ".ogg": {},
		".flac": {}, ".amr": {}, ".opus": {},
	}
	imageExts = map[string]struct{}{
		".jpg": {}, ".jpeg": {}, ".png": {}, ".gif": {},
		".webp": {}, ".bmp": {}, ".heic": {},
	}
	docExts = map[string]struct{}{
		".pdf": {}, ".doc": {}, ".docx": {}, ".ppt": {}, ".pptx": {},
		".xls": {}, ".xlsx": {}, ".txt": {}, ".md": {},
	}
	videoExts = map[string]struct{}{
		".mp4": {}, ".mov": {}, ".avi": {}, ".mkv": {}, ".webm": {}, ".m4v": {},
	}
)

const supportedFormatsHint = "支持：录音 mp3/wav/m4a/aac/ogg/flac/amr/opus；图片 jpg/jpeg/png/gif/webp/bmp/heic；文档 pdf/doc/docx/ppt/pptx/xls/xlsx/txt/md；链接 http(s) URL；文字 --content/--content-file/--stdin。视频无法保存为笔记，请使用 zhizai file upload"

// NewSaveCmd returns the save command (auto-detect note type).
func NewSaveCmd() *cobra.Command {
	var (
		title       string
		content     string
		contentFile string
		stdin       bool
		waitFlag    bool
		noWait      bool
		sceneID     string
		kb          string
		dir         string
		appendTo    string
		source      string
		start       string
		end         string
		duration    string
		lat         string
		lng         string
		text        string
		images      []string
		remark      string
		fileID      string
	)

	cmd := &cobra.Command{
		Use:   "save [文件|URL]...",
		Short: "自动判型并做成笔记（内部上传+创建，异步类型默认等待）",
		Long: `按输入自动判断笔记类型，一条命令完成上传与创建。
录音/图片/文档默认等待处理完成；可用 --no-wait 立即返回 noteId。
视频不能做成笔记，请改用 zhizai file upload。`,
		Example: `  zhizai save ./meeting.mp3 --title "三季度经营分析会"
  zhizai save ./meeting.mp3 --no-wait -o json
  zhizai save ./a.jpg ./b.png --remark "白板" --wait
  zhizai save ./方案.pdf --title "三季度方案"
  zhizai save https://example.com/a
  zhizai save --title "周报提醒" --content "明天交周报"
  zhizai save ./a.mp3 --image ./p1.jpg --kb <knowledgeId>`,
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			textBody, err := resolveTextContent(content, contentFile, stdin, cmd.InOrStdin())
			if err != nil {
				return err
			}

			kind, paths, linkURL, err := detectInput(args, textBody, images)
			if err != nil {
				return err
			}

			shouldWait := kind == "voice" || kind == "image" || kind == "document"
			if noWait {
				shouldWait = false
			} else if cmd.Flags().Changed("wait") {
				shouldWait = waitFlag
			}

			c := client.New()
			params, archiveViaCreate, err := buildCreateParams(c, cmd, kind, paths, linkURL, textBody, saveOpts{
				title:    title,
				sceneID:  sceneID,
				kb:       kb,
				dir:      dir,
				appendTo: appendTo,
				source:   source,
				start:    start,
				end:      end,
				duration: duration,
				lat:      lat,
				lng:      lng,
				text:     text,
				images:   images,
				remark:   remark,
				fileID:   fileID,
			})
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.ErrOrStderr(), "请求接口: %s\n", c.APIEndpoint("/note/createNote"))
			note, err := c.NoteCreate(params)
			if err != nil {
				return err
			}
			if note == nil || strings.TrimSpace(note.ID) == "" {
				return fmt.Errorf("创建成功但未返回 noteId")
			}

			warning := ""
			kb = strings.TrimSpace(kb)
			if kb != "" && !archiveViaCreate {
				warning = "已创建但未归档"
			}

			state := note.NoteState
			if shouldWait {
				fmt.Fprintf(cmd.ErrOrStderr(), "请求接口: %s\n", c.APIEndpoint("/note/queryNoteStatus"))
				st, waitErr := c.NoteWait(note.ID, 0)
				if waitErr != nil {
					return waitErr
				}
				if st != nil {
					state = st.NoteState
				}
				if state == "completed" {
					detail, getErr := c.NoteGet(note.ID)
					if getErr == nil && detail != nil {
						note = detail
						state = detail.NoteState
					}
				}
			}

			if kb != "" && archiveViaCreate && warning == "" {
				// 录音 knowledgeId 写入创建体；无效 ID 时服务端仍可能创建成功但未归档
				warning = "笔记已创建；若笔记集无效则可能已创建但未归档"
			}

			return writeSaveResult(cmd, note, state, warning)
		},
	}

	cmd.Flags().StringVar(&title, "title", "", "标题")
	cmd.Flags().StringVar(&content, "content", "", "文字笔记正文")
	cmd.Flags().StringVar(&contentFile, "content-file", "", "从文件读取文字正文")
	cmd.Flags().BoolVar(&stdin, "stdin", false, "从标准输入读取文字正文")
	cmd.Flags().BoolVar(&waitFlag, "wait", true, "等待异步处理完成（录音/图片/文档默认开启）")
	cmd.Flags().BoolVar(&noWait, "no-wait", false, "不等待，立即返回 noteId 与 note_state")
	cmd.Flags().StringVar(&sceneID, "scene-id", "", "总结场景 ID")
	cmd.Flags().StringVar(&kb, "kb", "", "归档到笔记集 knowledgeId（录音写入创建体）")
	cmd.Flags().StringVar(&dir, "dir", "", "笔记集目录 ID（与 --kb 配合）")
	cmd.Flags().StringVar(&appendTo, "append-to", "", "追加到已有录音笔记 ID")
	cmd.Flags().StringVar(&source, "source", "", "录音途径: realtime/phoneInternal/offlineImport/recordingCard")
	cmd.Flags().StringVar(&start, "start", "", "录制开始时间")
	cmd.Flags().StringVar(&end, "end", "", "录制结束时间")
	cmd.Flags().StringVar(&duration, "duration", "", "录音时长（秒）")
	cmd.Flags().StringVar(&lat, "lat", "", "纬度")
	cmd.Flags().StringVar(&lng, "lng", "", "经度")
	cmd.Flags().StringVar(&text, "text", "", "录音随手记备注")
	cmd.Flags().StringArrayVar(&images, "image", nil, "录音附图（可重复）")
	cmd.Flags().StringVar(&remark, "remark", "", "图片备注")
	cmd.Flags().StringVar(&fileID, "file-id", "", "已有 fileId，跳过上传（单文件类型）")

	return cmd
}

type saveOpts struct {
	title    string
	sceneID  string
	kb       string
	dir      string
	appendTo string
	source   string
	start    string
	end      string
	duration string
	lat      string
	lng      string
	text     string
	images   []string
	remark   string
	fileID   string
}

func resolveTextContent(content, contentFile string, useStdin bool, in io.Reader) (string, error) {
	n := 0
	if strings.TrimSpace(content) != "" {
		n++
	}
	if strings.TrimSpace(contentFile) != "" {
		n++
	}
	if useStdin {
		n++
	}
	if n > 1 {
		return "", fmt.Errorf("--content、--content-file、--stdin 只能选用其一")
	}
	if strings.TrimSpace(content) != "" {
		return content, nil
	}
	if strings.TrimSpace(contentFile) != "" {
		b, err := os.ReadFile(contentFile)
		if err != nil {
			return "", fmt.Errorf("读取内容文件失败: %w", err)
		}
		return string(b), nil
	}
	if useStdin {
		b, err := io.ReadAll(in)
		if err != nil {
			return "", fmt.Errorf("读取标准输入失败: %w", err)
		}
		return string(b), nil
	}
	return "", nil
}

func detectInput(args []string, textBody string, imageFlags []string) (kind string, paths []string, linkURL string, err error) {
	var files []string
	var urls []string
	for _, a := range args {
		a = strings.TrimSpace(a)
		if a == "" {
			continue
		}
		if isHTTPURL(a) {
			urls = append(urls, a)
			continue
		}
		files = append(files, a)
	}

	hasText := strings.TrimSpace(textBody) != ""
	if len(urls) > 0 && (len(files) > 0 || hasText || len(imageFlags) > 0) {
		return "", nil, "", fmt.Errorf("链接与文件/文字不能混用")
	}
	if len(urls) > 1 {
		return "", nil, "", fmt.Errorf("一次只能保存一个链接")
	}
	if len(urls) == 1 {
		return "link", nil, urls[0], nil
	}

	if len(files) == 0 {
		if hasText {
			return "text", nil, "", nil
		}
		if len(imageFlags) > 0 {
			return "", nil, "", fmt.Errorf("附图需配合录音主文件使用，例如: zhizai save ./a.mp3 --image ./p1.jpg")
		}
		return "", nil, "", fmt.Errorf("请提供文件、URL，或使用 --content/--content-file/--stdin。%s", supportedFormatsHint)
	}

	for _, p := range files {
		if _, err := os.Stat(p); err != nil {
			if os.IsNotExist(err) {
				return "", nil, "", fmt.Errorf("文件不存在: %s", p)
			}
			return "", nil, "", fmt.Errorf("无法访问文件 %s: %w", p, err)
		}
	}
	for _, p := range imageFlags {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if _, err := os.Stat(p); err != nil {
			if os.IsNotExist(err) {
				return "", nil, "", fmt.Errorf("附图不存在: %s", p)
			}
			return "", nil, "", fmt.Errorf("无法访问附图 %s: %w", p, err)
		}
	}

	// Classify each file
	var audio, images, docs, videos, unknown []string
	for _, p := range files {
		ext := strings.ToLower(filepath.Ext(p))
		switch {
		case hasExt(audioExts, ext):
			audio = append(audio, p)
		case hasExt(imageExts, ext):
			images = append(images, p)
		case hasExt(docExts, ext):
			docs = append(docs, p)
		case hasExt(videoExts, ext):
			videos = append(videos, p)
		default:
			unknown = append(unknown, p)
		}
	}

	if len(videos) > 0 {
		return "", nil, "", fmt.Errorf("视频无法保存为笔记（不支持做成笔记）。如只需上传，请使用: zhizai file upload <视频路径>")
	}
	if len(unknown) > 0 {
		return "", nil, "", fmt.Errorf("不支持的文件类型 %s。%s", strings.Join(unknown, ", "), supportedFormatsHint)
	}

	nKinds := 0
	if len(audio) > 0 {
		nKinds++
	}
	if len(images) > 0 {
		nKinds++
	}
	if len(docs) > 0 {
		nKinds++
	}
	if nKinds > 1 {
		// 允许：一份音频 + 位置参数里的图片（等同附图）—— 若同时有 --image，一并视为录音附图
		if len(audio) == 1 && len(docs) == 0 && len(images) > 0 {
			return "voice", append([]string{audio[0]}, images...), "", nil
		}
		return "", nil, "", fmt.Errorf("不能混用不同类型的文件。%s", supportedFormatsHint)
	}

	if len(audio) > 0 {
		if len(audio) > 1 {
			return "", nil, "", fmt.Errorf("一次只能保存一份录音文件")
		}
		return "voice", audio, "", nil
	}
	if len(docs) > 0 {
		if len(docs) > 1 {
			return "", nil, "", fmt.Errorf("一次只能保存一份文档")
		}
		if len(imageFlags) > 0 {
			return "", nil, "", fmt.Errorf("--image 仅用于录音附图")
		}
		return "document", docs, "", nil
	}
	if len(images) > 0 {
		if len(imageFlags) > 0 {
			return "", nil, "", fmt.Errorf("图片笔记请直接传图片路径，不要再用 --image")
		}
		return "image", images, "", nil
	}

	return "", nil, "", fmt.Errorf("无法识别输入。%s", supportedFormatsHint)
}

func hasExt(set map[string]struct{}, ext string) bool {
	_, ok := set[ext]
	return ok
}

func isHTTPURL(s string) bool {
	u, err := url.Parse(s)
	if err != nil {
		return false
	}
	scheme := strings.ToLower(u.Scheme)
	return (scheme == "http" || scheme == "https") && u.Host != ""
}

func buildCreateParams(c *client.Client, cmd *cobra.Command, kind string, paths []string, linkURL, textBody string, opts saveOpts) (client.NoteCreateParams, bool, error) {
	params := client.NoteCreateParams{NoteType: kind}
	if sid := strings.TrimSpace(opts.sceneID); sid != "" {
		params.SceneID = sid
	}
	archiveViaCreate := false

	switch kind {
	case "text":
		params.TextContent = &client.NoteTextContent{
			Title:   strings.TrimSpace(opts.title),
			Content: textBody,
		}
		if strings.TrimSpace(opts.kb) != "" {
			// 非录音类型创建体不带 knowledgeId；Batch 1 无 move 接口
			archiveViaCreate = false
		}

	case "link":
		params.LinkContent = &client.NoteLinkContent{URL: linkURL}

	case "voice":
		voicePath := paths[0]
		extraImages := opts.images
		// paths 可能已含位置参数中的附图
		var attachPaths []string
		if len(paths) > 1 {
			attachPaths = append(attachPaths, paths[1:]...)
		}
		attachPaths = append(attachPaths, extraImages...)

		vfID := strings.TrimSpace(opts.fileID)
		if vfID == "" {
			fmt.Fprintf(cmd.ErrOrStderr(), "请求接口: %s\n", c.APIEndpoint("/file/uploadSingleFile"))
			up, err := c.UploadFile(voicePath, false)
			if err != nil {
				return params, false, err
			}
			vfID = up.FileID
		}

		imageIDs := make([]string, 0, len(attachPaths))
		for _, p := range attachPaths {
			fmt.Fprintf(cmd.ErrOrStderr(), "请求接口: %s\n", c.APIEndpoint("/file/uploadSingleFile"))
			up, err := c.UploadFile(p, false)
			if err != nil {
				return params, false, fmt.Errorf("上传附图失败 %s: %w", p, err)
			}
			imageIDs = append(imageIDs, up.FileID)
		}

		src := strings.TrimSpace(opts.source)
		if src == "" {
			src = "offlineImport"
		}
		vc := &client.NoteVoiceContent{
			VoiceFileID:     vfID,
			Title:           strings.TrimSpace(opts.title),
			Text:            strings.TrimSpace(opts.text),
			RecStartTime:    strings.TrimSpace(opts.start),
			RecEndTime:      strings.TrimSpace(opts.end),
			Duration:        strings.TrimSpace(opts.duration),
			ImageFileIds:    imageIDs,
			AppendNoteID:    strings.TrimSpace(opts.appendTo),
			Latitude:        strings.TrimSpace(opts.lat),
			Longitude:       strings.TrimSpace(opts.lng),
			RecordingSource: src,
		}
		if kb := strings.TrimSpace(opts.kb); kb != "" {
			vc.KnowledgeID = kb
			archiveViaCreate = true
			if d := strings.TrimSpace(opts.dir); d != "" && d != "-1" {
				vc.DirectoryID = d
			}
		}
		params.VoiceContent = vc

	case "image":
		files := make([]client.NoteImageFile, 0, len(paths))
		fid := strings.TrimSpace(opts.fileID)
		if fid != "" && len(paths) == 1 {
			files = append(files, client.NoteImageFile{
				FileID: fid,
				Remark: strings.TrimSpace(opts.remark),
			})
		} else {
			if fid != "" && len(paths) > 1 {
				return params, false, fmt.Errorf("--file-id 仅适用于单张图片；多图请直接传路径由 CLI 上传")
			}
			for i, p := range paths {
				fmt.Fprintf(cmd.ErrOrStderr(), "请求接口: %s\n", c.APIEndpoint("/file/uploadSingleFile"))
				up, err := c.UploadFile(p, false)
				if err != nil {
					return params, false, err
				}
				item := client.NoteImageFile{FileID: up.FileID}
				if i == 0 {
					item.Remark = strings.TrimSpace(opts.remark)
				}
				files = append(files, item)
			}
		}
		params.ImageContent = &client.NoteImageContent{FileIds: files}

	case "document":
		docPath := paths[0]
		fid := strings.TrimSpace(opts.fileID)
		fileName := filepath.Base(docPath)
		if fid == "" {
			fmt.Fprintf(cmd.ErrOrStderr(), "请求接口: %s\n", c.APIEndpoint("/file/uploadSingleFile"))
			up, err := c.UploadFile(docPath, false)
			if err != nil {
				return params, false, err
			}
			fid = up.FileID
			if up.FileName != "" {
				fileName = up.FileName
			}
		}
		params.DocumentContent = &client.NoteDocumentContent{
			FileID:   fid,
			FileName: fileName,
			Title:    strings.TrimSpace(opts.title),
		}

	default:
		return params, false, fmt.Errorf("内部错误：未知类型 %s", kind)
	}

	return params, archiveViaCreate, nil
}

func writeSaveResult(cmd *cobra.Command, note *client.Note, state, warning string) error {
	data := map[string]interface{}{
		"noteId":    note.ID,
		"title":     note.Title,
		"note_type": note.NoteType,
		"note_state": state,
	}
	if note.Abstract != "" {
		data["abstract"] = note.Abstract
	}
	if note.Summary != "" {
		data["summary"] = note.Summary
	}
	if warning != "" {
		data["warning"] = warning
	}

	if output.Format() == "json" {
		return output.WriteSuccessJSON(cmd.OutOrStdout(), data)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "已创建笔记 %s\n", note.ID)
	fmt.Fprint(cmd.OutOrStdout(), ui.FormatFieldPair("Title", note.Title))
	fmt.Fprint(cmd.OutOrStdout(), ui.FormatFieldPair("Type", client.NoteTypeLabel(note.NoteType)))
	fmt.Fprint(cmd.OutOrStdout(), ui.FormatFieldPair("State", state))
	if note.Abstract != "" {
		fmt.Fprint(cmd.OutOrStdout(), ui.FormatFieldPair("Abstract", ui.Truncate(note.Abstract, 160)))
	}
	if note.Summary != "" {
		fmt.Fprint(cmd.OutOrStdout(), ui.FormatFieldPair("Summary", ui.Truncate(note.Summary, 200)))
	}
	if warning != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "警告: %s\n", warning)
	}
	return nil
}
