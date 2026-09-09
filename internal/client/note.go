package client

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

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
}

// NoteTextContent is textContent for createNote.
type NoteTextContent struct {
	Title   string `json:"title,omitempty"`
	Content string `json:"content,omitempty"`
}

// NoteCreateParams matches POST /note/createNote (phase-1: text notes).
type NoteCreateParams struct {
	NoteType    string           `json:"noteType"`
	SceneID     interface{}      `json:"sceneId,omitempty"`
	TextContent *NoteTextContent `json:"textContent,omitempty"`
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
	q := url.Values{"noteId": {noteID}}
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

// NoteCreate creates a note (phase-1: text).
func (c *Client) NoteCreate(params NoteCreateParams) (*Note, error) {
	params.NoteType = strings.TrimSpace(params.NoteType)
	if params.NoteType == "" {
		params.NoteType = "text"
	}
	if params.NoteType != "text" {
		return nil, fmt.Errorf("当前仅支持创建文字笔记（--type text），其他类型待 file upload 就绪后开放")
	}
	if params.TextContent == nil {
		return nil, fmt.Errorf("缺少 textContent")
	}
	if strings.TrimSpace(params.TextContent.Content) == "" && strings.TrimSpace(params.TextContent.Title) == "" {
		return nil, fmt.Errorf("title 与 content 不能同时为空")
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
	default:
		if state == "" {
			return "-"
		}
		return state
	}
}
