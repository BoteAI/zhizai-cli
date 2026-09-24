package client

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

// TextNoteSummaryParams is POST /note/createTextNoteSummary.
type TextNoteSummaryParams struct {
	Content string `json:"content"`
	SceneID string `json:"sceneId,omitempty"`
}

// QueryTemplateResult is GET /know/queryStandardInputOutputByCommand resultObject.
type QueryTemplateResult struct {
	Input  string `json:"input"`
	Output string `json:"output"`
}

// SummaryAndRecordingParams is GET /note/querySummaryAndRecording.
type SummaryAndRecordingParams struct {
	Title     string
	StartTime string
	EndTime   string
}

// SummaryScene is one scene summary under a note.
type SummaryScene struct {
	SceneID        string `json:"scene_id"`
	SceneName      string `json:"scene_name"`
	SummaryContent string `json:"summary_content"`
	CreateTime     string `json:"create_time"`
}

// RecordingTranscript is one transcript segment.
type RecordingTranscript struct {
	RawText string `json:"raw_text"`
	Start   any    `json:"start"`
	End     any    `json:"end"`
	Spk     string `json:"spk"`
	Text    string `json:"text,omitempty"`
	TnText  string `json:"tn_text,omitempty"`
}

// RecordingItem is one recording under querySummaryAndRecording.
type RecordingItem struct {
	RecordingID string                `json:"recording_id"`
	Duration    string                `json:"duration"`
	StartTime   string                `json:"start_time"`
	CreateTime  string                `json:"create_time"`
	Transcript  []RecordingTranscript `json:"transcript"`
}

// SummaryAndRecording is one resultObject entry.
type SummaryAndRecording struct {
	NoteID         string          `json:"noteId"`
	NoteTitle      string          `json:"noteTitle"`
	NoteCreateTime string          `json:"noteCreateTime"`
	NoteStatus     string          `json:"noteStatus"`
	SceneList      []SummaryScene  `json:"sceneList"`
	RecordingList  []RecordingItem `json:"recordingList"`
}

// CreateTextNoteSummary streams POST /note/createTextNoteSummary and concatenates data chunks.
func (c *Client) CreateTextNoteSummary(params TextNoteSummaryParams) (string, error) {
	params.Content = strings.TrimSpace(params.Content)
	if params.Content == "" {
		return "", fmt.Errorf("content 不能为空")
	}
	params.SceneID = strings.TrimSpace(params.SceneID)

	var b strings.Builder
	err := c.doSSEPOST("/note/createTextNoteSummary", params, func(event, data string) error {
		ev := strings.TrimSpace(event)
		switch ev {
		case "error":
			msg := strings.TrimSpace(data)
			if msg == "" {
				msg = "createTextNoteSummary error"
			}
			return &RequestError{
				APIError: APIError{
					Code:      "error",
					Message:   msg,
					Reason:    "summary_error",
					Retryable: false,
				},
			}
		case "done":
			return nil
		default:
			if data == "[DONE]" {
				return nil
			}
			b.WriteString(data)
		}
		return nil
	})
	return b.String(), err
}

// QueryTemplate queries GET /know/queryStandardInputOutputByCommand.
// If output is empty, retries once with the same command (caller may pass keyword then full sentence).
func (c *Client) QueryTemplate(command string) (*QueryTemplateResult, error) {
	command = strings.TrimSpace(command)
	if command == "" {
		return nil, fmt.Errorf("command 不能为空")
	}
	q := url.Values{"command": {command}}
	raw, err := doGet(c, "/know/queryStandardInputOutputByCommand?"+q.Encode())
	if err != nil {
		return nil, err
	}
	var out QueryTemplateResult
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("parsing template: %w", err)
	}
	return &out, nil
}

// QuerySummaryAndRecording queries GET /note/querySummaryAndRecording.
func (c *Client) QuerySummaryAndRecording(params SummaryAndRecordingParams) ([]SummaryAndRecording, error) {
	q := url.Values{}
	if t := strings.TrimSpace(params.Title); t != "" {
		q.Set("title", t)
	}
	if t := strings.TrimSpace(params.StartTime); t != "" {
		q.Set("startTime", t)
	}
	if t := strings.TrimSpace(params.EndTime); t != "" {
		q.Set("endTime", t)
	}
	path := "/note/querySummaryAndRecording"
	if enc := q.Encode(); enc != "" {
		path += "?" + enc
	}
	raw, err := doGet(c, path)
	if err != nil {
		return nil, err
	}
	if string(raw) == "null" || len(raw) == 0 {
		return []SummaryAndRecording{}, nil
	}
	var list []SummaryAndRecording
	if err := json.Unmarshal(raw, &list); err != nil {
		return nil, fmt.Errorf("parsing summary and recording: %w", err)
	}
	return list, nil
}
