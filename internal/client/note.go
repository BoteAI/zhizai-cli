package client

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"
)

// DefaultNoteWaitTimeout is used when NoteWait timeout <= 0.
const DefaultNoteWaitTimeout = 10 * time.Minute

// Note is a 智在记录 note item.
type Note struct {
	ID           string          `json:"id"`
	Title        string          `json:"title"`
	Summary      string          `json:"summary"`
	Abstract     string          `json:"abstract"`
	Content      json.RawMessage `json:"content,omitempty"`
	Status       string          `json:"status"`
	NoteType     string          `json:"note_type"`
	NoteState    string          `json:"note_state"`
	CreateTime   string          `json:"create_time"`
	CreatorID    string          `json:"creator_id"`
	SceneName    string          `json:"scene_name"`
	SceneID      string          `json:"scene_id"`
	SourceNoteID string          `json:"source_note_id,omitempty"`
	NoteCategory *int            `json:"note_category,omitempty"`
	DeviceSN     string          `json:"device_sn,omitempty"`
	Latitude     string          `json:"latitude,omitempty"`
	Longitude    string          `json:"longitude,omitempty"`
	RecEndTime   string          `json:"rec_end_time,omitempty"`
	AccountNum   string          `json:"account_num,omitempty"`
	ShortURL     string          `json:"short_url,omitempty"`
}

// NoteListData is the paginated note list payload.
type NoteListData struct {
	PageNum         int    `json:"pageNum"`
	PageSize        int    `json:"pageSize"`
	Total           string `json:"total"`
	Pages           int    `json:"pages"`
	Size            int    `json:"size"`
	HasNextPage     bool   `json:"hasNextPage"`
	HasPreviousPage bool   `json:"hasPreviousPage"`
	IsFirstPage     bool   `json:"isFirstPage"`
	IsLastPage      bool   `json:"isLastPage"`
	List            []Note `json:"list"`
}

// NoteListParams matches POST /note/queryNoteList.
type NoteListParams struct {
	Title           string `json:"title,omitempty"`
	AbstractContent string `json:"abstractContent,omitempty"`
	Summary         string `json:"summary,omitempty"`
	Content         string `json:"content,omitempty"`
	NoteType        string `json:"noteType,omitempty"`
	StartTime       string `json:"startTime,omitempty"`
	EndTime         string `json:"endTime,omitempty"`
	PageNum         int    `json:"pageNum,omitempty"`
	PageSize        int    `json:"pageSize,omitempty"`
	WithContent     string `json:"withContent,omitempty"`
	WithShortUrl    string `json:"withShortUrl,omitempty"`
}

// NoteTextContent is textContent for createNote.
type NoteTextContent struct {
	Title   string `json:"title,omitempty"`
	Content string `json:"content,omitempty"`
}

// NoteVoiceContent is voiceContent for createNote.
type NoteVoiceContent struct {
	VoiceFileID     string   `json:"voiceFileId,omitempty"`
	Title           string   `json:"title,omitempty"`
	Text            string   `json:"text,omitempty"`
	RecStartTime    string   `json:"recStartTime,omitempty"`
	RecEndTime      string   `json:"recEndTime,omitempty"`
	Duration        string   `json:"duration,omitempty"`
	ImageFileIds    []string `json:"imageFileIds,omitempty"`
	AppendNoteID    string   `json:"appendNoteId,omitempty"`
	DeviceSN        string   `json:"deviceSn,omitempty"`
	Latitude        string   `json:"latitude,omitempty"`
	Longitude       string   `json:"longitude,omitempty"`
	RecordingSource string   `json:"recordingSource,omitempty"`
	KnowledgeID     string   `json:"knowledgeId,omitempty"`
	DirectoryID     string   `json:"directoryId,omitempty"`
}

// NoteImageFile is one entry in imageContent.fileIds.
type NoteImageFile struct {
	FileID string `json:"fileId,omitempty"`
	Remark string `json:"remark,omitempty"`
}

// NoteImageContent is imageContent for createNote.
type NoteImageContent struct {
	FileIds []NoteImageFile `json:"fileIds,omitempty"`
}

// NoteDocumentContent is documentContent for createNote.
type NoteDocumentContent struct {
	FileID   string `json:"fileId,omitempty"`
	FileName string `json:"fileName,omitempty"`
	Title    string `json:"title,omitempty"`
}

// NoteLinkContent is linkContent for createNote.
type NoteLinkContent struct {
	URL string `json:"url,omitempty"`
}

// NoteCreateParams matches POST /note/createNote.
type NoteCreateParams struct {
	NoteType        string               `json:"noteType"`
	SceneID         interface{}          `json:"sceneId,omitempty"`
	TextContent     *NoteTextContent     `json:"textContent,omitempty"`
	VoiceContent    *NoteVoiceContent    `json:"voiceContent,omitempty"`
	ImageContent    *NoteImageContent    `json:"imageContent,omitempty"`
	DocumentContent *NoteDocumentContent `json:"documentContent,omitempty"`
	LinkContent     *NoteLinkContent     `json:"linkContent,omitempty"`
}

// NoteUpdateParams matches POST /note/updateNoteInfo.
type NoteUpdateParams struct {
	NoteID          string `json:"noteId"`
	Title           string `json:"title,omitempty"`
	AbstractContent string `json:"abstractContent,omitempty"`
	Summary         string `json:"summary,omitempty"`
}

// NoteStatus is GET /note/queryNoteStatus resultObject.
type NoteStatus struct {
	NoteState string `json:"noteState"`
}

// NoteGetOptions controls optional query params for note detail.
type NoteGetOptions struct {
	WithShortUrl string
}

// NoteAppendDetail is resultObject of GET /note/qryNoteDetailInfoAndAppend.
type NoteAppendDetail struct {
	QueryMainNoteInfo  *Note  `json:"queryMainNoteInfo"`
	QueryRecordingNote []Note `json:"queryRecordingNote,omitempty"`
}

// NoteList queries notes with pagination.
func (c *Client) NoteList(params NoteListParams) (*NoteListData, error) {
	if params.PageNum <= 0 {
		params.PageNum = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 20
	}
	raw, err := doPost(c, "/note/queryNoteList", params)
	if err != nil {
		return nil, err
	}
	var data NoteListData
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, fmt.Errorf("parsing note list: %w", err)
	}
	return &data, nil
}

// NoteGet queries a single note by ID.
func (c *Client) NoteGet(noteID string) (*Note, error) {
	return c.NoteGetWithOptions(noteID, NoteGetOptions{})
}

// NoteGetWithOptions queries a single note with optional withShortUrl.
func (c *Client) NoteGetWithOptions(noteID string, opts NoteGetOptions) (*Note, error) {
	noteID = strings.TrimSpace(noteID)
	if noteID == "" {
		return nil, fmt.Errorf("noteId 不能为空")
	}
	q := url.Values{"noteId": {noteID}}
	if strings.TrimSpace(opts.WithShortUrl) != "" {
		q.Set("withShortUrl", strings.TrimSpace(opts.WithShortUrl))
	}
	raw, err := doGet(c, "/note/querySingleNoteDetail?"+q.Encode())
	if err != nil {
		return nil, err
	}
	var note Note
	if err := json.Unmarshal(raw, &note); err != nil {
		return nil, fmt.Errorf("parsing note detail: %w", err)
	}
	return &note, nil
}

// NoteGetWithAppend queries main note plus append segments.
func (c *Client) NoteGetWithAppend(noteID string) (*NoteAppendDetail, error) {
	noteID = strings.TrimSpace(noteID)
	if noteID == "" {
		return nil, fmt.Errorf("noteId 不能为空")
	}
	q := url.Values{"noteId": {noteID}}
	raw, err := doGet(c, "/note/qryNoteDetailInfoAndAppend?"+q.Encode())
	if err != nil {
		return nil, err
	}
	var detail NoteAppendDetail
	if err := json.Unmarshal(raw, &detail); err != nil {
		return nil, fmt.Errorf("parsing note append detail: %w", err)
	}
	return &detail, nil
}

// NoteCreate creates a note (text / voice / image / document / link).
func (c *Client) NoteCreate(params NoteCreateParams) (*Note, error) {
	params.NoteType = strings.TrimSpace(params.NoteType)
	if params.NoteType == "" {
		params.NoteType = "text"
	}
	switch params.NoteType {
	case "text":
		if params.TextContent == nil {
			return nil, fmt.Errorf("缺少 textContent")
		}
		if strings.TrimSpace(params.TextContent.Content) == "" && strings.TrimSpace(params.TextContent.Title) == "" {
			return nil, fmt.Errorf("title 与 content 不能同时为空")
		}
	case "voice":
		if params.VoiceContent == nil {
			return nil, fmt.Errorf("缺少 voiceContent")
		}
		if strings.TrimSpace(params.VoiceContent.VoiceFileID) == "" {
			return nil, fmt.Errorf("voiceContent.voiceFileId 不能为空")
		}
		if strings.TrimSpace(params.VoiceContent.RecordingSource) == "" {
			params.VoiceContent.RecordingSource = "offlineImport"
		}
	case "image":
		if params.ImageContent == nil || len(params.ImageContent.FileIds) == 0 {
			return nil, fmt.Errorf("缺少 imageContent.fileIds")
		}
		for i, f := range params.ImageContent.FileIds {
			if strings.TrimSpace(f.FileID) == "" {
				return nil, fmt.Errorf("imageContent.fileIds[%d].fileId 不能为空", i)
			}
		}
	case "document":
		if params.DocumentContent == nil {
			return nil, fmt.Errorf("缺少 documentContent")
		}
		if strings.TrimSpace(params.DocumentContent.FileID) == "" {
			return nil, fmt.Errorf("documentContent.fileId 不能为空")
		}
	case "link":
		if params.LinkContent == nil {
			return nil, fmt.Errorf("缺少 linkContent")
		}
		if strings.TrimSpace(params.LinkContent.URL) == "" {
			return nil, fmt.Errorf("linkContent.url 不能为空")
		}
	default:
		return nil, fmt.Errorf("不支持的笔记类型 %q（支持 text/voice/image/document/link）", params.NoteType)
	}
	if params.SceneID != nil {
		params.SceneID = coerceKnowledgeID(params.SceneID)
	}
	raw, err := doPost(c, "/note/createNote", params)
	if err != nil {
		return nil, err
	}
	var note Note
	if err := json.Unmarshal(raw, &note); err != nil {
		return nil, fmt.Errorf("parsing create note response: %w", err)
	}
	return &note, nil
}

// NoteUpdate updates title / abstract / summary. Blank fields are omitted (server keeps unchanged).
func (c *Client) NoteUpdate(params NoteUpdateParams) error {
	params.NoteID = strings.TrimSpace(params.NoteID)
	if params.NoteID == "" {
		return fmt.Errorf("noteId 不能为空")
	}
	params.Title = strings.TrimSpace(params.Title)
	params.AbstractContent = strings.TrimSpace(params.AbstractContent)
	params.Summary = strings.TrimSpace(params.Summary)
	if params.Title == "" && params.AbstractContent == "" && params.Summary == "" {
		return fmt.Errorf("请至少提供 --title / --abstract / --summary 之一")
	}
	_, err := doPost(c, "/note/updateNoteInfo", params)
	return err
}

// NoteDelete deletes a note by ID.
func (c *Client) NoteDelete(noteID string) error {
	noteID = strings.TrimSpace(noteID)
	if noteID == "" {
		return fmt.Errorf("noteId 不能为空")
	}
	q := url.Values{"noteId": {noteID}}
	_, err := doGet(c, "/note/deleteNote?"+q.Encode())
	return err
}

// NoteStatus queries processing state for a note.
func (c *Client) NoteStatus(noteID string) (*NoteStatus, error) {
	noteID = strings.TrimSpace(noteID)
	if noteID == "" {
		return nil, fmt.Errorf("noteId 不能为空")
	}
	q := url.Values{"noteId": {noteID}}
	raw, err := doGet(c, "/note/queryNoteStatus?"+q.Encode())
	if err != nil {
		return nil, err
	}
	var st NoteStatus
	if err := json.Unmarshal(raw, &st); err != nil {
		return nil, fmt.Errorf("parsing note status: %w", err)
	}
	return &st, nil
}

// isTerminalNoteState reports whether note processing has reached a final state.
func isTerminalNoteState(state string) bool {
	switch state {
	case "completed", "failed", "recognizing_failed", "analyzing_failed":
		return true
	default:
		return false
	}
}

// NoteWait polls NoteStatus until a terminal state or timeout (default 10 minutes).
// Polling is paced by the client's >=500ms rate limiter.
func (c *Client) NoteWait(noteID string, timeout time.Duration) (*NoteStatus, error) {
	noteID = strings.TrimSpace(noteID)
	if noteID == "" {
		return nil, fmt.Errorf("noteId 不能为空")
	}
	if timeout <= 0 {
		timeout = DefaultNoteWaitTimeout
	}
	deadline := time.Now().Add(timeout)
	var last *NoteStatus
	for {
		st, err := c.NoteStatus(noteID)
		if err != nil {
			return last, err
		}
		last = st
		if isTerminalNoteState(st.NoteState) {
			return st, nil
		}
		if time.Now().After(deadline) {
			return last, &RequestError{
				APIError: APIError{
					Code:      "timeout",
					Message:   fmt.Sprintf("等待笔记处理超时（当前状态 %s）", st.NoteState),
					Reason:    "note_wait_timeout",
					Retryable: true,
				},
			}
		}
		// Next NoteStatus call is paced by waitRateLimit (>=500ms).
	}
}

// DownloadNoteAudio downloads note audio via GET /note/downloadNoteAudio.
// Sets APP-ID from ZHIZAI_APP_ID or last upload appId when available.
func (c *Client) DownloadNoteAudio(noteID, destPath string) error {
	noteID = strings.TrimSpace(noteID)
	if noteID == "" {
		return fmt.Errorf("noteId 不能为空")
	}
	destPath = strings.TrimSpace(destPath)
	if destPath == "" {
		return fmt.Errorf("目标路径不能为空")
	}
	q := url.Values{"noteId": {noteID}}
	headers := map[string]string{}
	if appID := c.ResolveAppID(); appID != "" {
		headers["APP-ID"] = appID
	}
	return c.doBinaryGET("/note/downloadNoteAudio?"+q.Encode(), destPath, headers)
}

// NoteTypeLabel returns the Chinese label for a note type.
func NoteTypeLabel(noteType string) string {
	switch noteType {
	case "text":
		return "文本"
	case "voice":
		return "录音"
	case "document":
		return "文档"
	case "link":
		return "链接"
	case "image":
		return "图片"
	case "video":
		return "视频"
	case "knowCard":
		return "知识卡片"
	default:
		if noteType == "" {
			return "-"
		}
		return noteType
	}
}

// NoteStateLabel returns a short Chinese label for note processing state.
func NoteStateLabel(state string) string {
	switch state {
	case "completed":
		return "已完成"
	case "pending":
		return "处理中"
	case "recognizing":
		return "转写中"
	case "analyzing":
		return "总结中"
	case "failed":
		return "失败"
	case "recognizing_failed":
		return "转写失败"
	case "analyzing_failed":
		return "总结失败"
	default:
		if state == "" {
			return "-"
		}
		return state
	}
}
