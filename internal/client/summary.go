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
