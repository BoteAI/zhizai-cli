package client

import (
	"encoding/json"
	"fmt"
	"net/url"
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
	PageNum        int    `json:"pageNum"`
	PageSize       int    `json:"pageSize"`
	Total          string `json:"total"`
	Pages          int    `json:"pages"`
	Size           int    `json:"size"`
	HasNextPage    bool   `json:"hasNextPage"`
	HasPreviousPage bool  `json:"hasPreviousPage"`
	IsFirstPage    bool   `json:"isFirstPage"`
	IsLastPage     bool   `json:"isLastPage"`
	List           []Note `json:"list"`
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
