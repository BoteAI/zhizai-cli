package client

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

// XiaozhiSessionParams is POST /note/createXiaozhiSession.
type XiaozhiSessionParams struct {
	Query                   string   `json:"query"`
	KnowledgeID             string   `json:"knowledgeId,omitempty"`
	KnowledgeIDList         []string `json:"knowledgeIdList,omitempty"`
	NoteID                  string   `json:"noteId,omitempty"`
	NoteIDList              []string `json:"noteIdList,omitempty"`
	DirectoryID             string   `json:"directoryId,omitempty"`
	DirectoryIDList         []string `json:"directoryIdList,omitempty"`
	CatalogID               string   `json:"catalogId,omitempty"`
	KnowledgeCatalogIDList  []string `json:"knowledgeCatalogIdList,omitempty"`
	ChatType                string   `json:"chatType,omitempty"`
	AgentCode               string   `json:"agentCode,omitempty"`
	BotID                   string   `json:"botId,omitempty"`
	SceneID                 string   `json:"sceneId,omitempty"`
	SkillID                 string   `json:"skillId,omitempty"`
}

// XiaozhiSession is createXiaozhiSession resultObject.
type XiaozhiSession struct {
	SessionID    string `json:"sessionId"`
	ChatID       string `json:"chatId"`
	ContextID    string `json:"contextId"`
	SessionTitle string `json:"sessionTitle"`
	ChatTitle    string `json:"chatTitle"`
	Prologue     string `json:"prologue,omitempty"`
}

// XiaozhiChatParams is POST /note/xiaozhiChat.
type XiaozhiChatParams struct {
	Query           string   `json:"query"`
	ChatID          string   `json:"chatId,omitempty"`
	SessionID       string   `json:"sessionId,omitempty"`
	ContextID       string   `json:"contextId"`
	WithReferences  bool     `json:"withReferences,omitempty"`
	LastChatID      string   `json:"lastChatId,omitempty"`
	ChatVersion     string   `json:"chatVersion,omitempty"`
	KnowledgeID     string   `json:"knowledgeId,omitempty"`
	KnowledgeIDList []string `json:"knowledgeIdList,omitempty"`
	NoteID          string   `json:"noteId,omitempty"`
	NoteIDList      []string `json:"noteIdList,omitempty"`
	DirectoryID     string   `json:"directoryId,omitempty"`
	DirectoryIDList []string `json:"directoryIdList,omitempty"`
}

// XiaozhiChatResult aggregates an SSE xiaozhiChat stream.
type XiaozhiChatResult struct {
	Text          string   `json:"text"`
	ReferenceIDs  []string `json:"referenceIds,omitempty"`
	LastEventID   string   `json:"lastEventId,omitempty"`
	Done          bool     `json:"done"`
}

// XiaozhiReference is one item from qryXiaozhiReferences.
type XiaozhiReference struct {
	NoteID  string `json:"noteId"`
	Name    string `json:"name"`
	Summary string `json:"summary"`
	ID      string `json:"id,omitempty"`
}

// CreateXiaozhiSession creates a xiaozhi chat session (stream=false on server).
func (c *Client) CreateXiaozhiSession(params XiaozhiSessionParams) (*XiaozhiSession, error) {
	params.Query = strings.TrimSpace(params.Query)
	if params.Query == "" {
		return nil, fmt.Errorf("问题不能为空")
	}
	normalizeXiaozhiScope(
		&params.KnowledgeID, &params.KnowledgeIDList,
		&params.NoteID, &params.NoteIDList,
		&params.DirectoryID, &params.DirectoryIDList,
	)
	raw, err := doPost(c, "/note/createXiaozhiSession", params)
	if err != nil {
		return nil, err
	}
	var sess XiaozhiSession
	if err := json.Unmarshal(raw, &sess); err != nil {
		return nil, fmt.Errorf("parsing xiaozhi session: %w", err)
	}
	if sess.ChatID == "" {
		sess.ChatID = sess.SessionID
	}
	if sess.SessionID == "" {
		sess.SessionID = sess.ChatID
	}
	return &sess, nil
}

// XiaozhiChat streams POST /note/xiaozhiChat and collects text / reference ids.
// notData and error events become RequestError with the event payload as Message.
func (c *Client) XiaozhiChat(params XiaozhiChatParams) (*XiaozhiChatResult, error) {
	params.Query = strings.TrimSpace(params.Query)
	if params.Query == "" {
		return nil, fmt.Errorf("问题不能为空")
	}
	params.ChatID = strings.TrimSpace(params.ChatID)
	params.SessionID = strings.TrimSpace(params.SessionID)
	params.ContextID = strings.TrimSpace(params.ContextID)
	if params.ChatID == "" && params.SessionID == "" {
		return nil, fmt.Errorf("请先创建问小智会话（缺少 chatId）")
	}
	if params.ContextID == "" {
		return nil, fmt.Errorf("上下文不能为空（缺少 contextId）")
	}
	normalizeXiaozhiScope(
		&params.KnowledgeID, &params.KnowledgeIDList,
		&params.NoteID, &params.NoteIDList,
		&params.DirectoryID, &params.DirectoryIDList,
	)
	if params.ChatVersion == "" {
		params.ChatVersion = "1"
	}

	result := &XiaozhiChatResult{}
	var textBuilder strings.Builder
	err := c.doSSEPOST("/note/xiaozhiChat", params, func(event, data string) error {
		ev := strings.TrimSpace(event)
		switch ev {
		case "text", "":
			if ev == "" && (data == "[DONE]" || data == "") {
				if data == "[DONE]" {
					result.Done = true
				}
				return nil
			}
			textBuilder.WriteString(data)
		case "references":
			ids := parseReferenceIDs(data)
			result.ReferenceIDs = append(result.ReferenceIDs, ids...)
		case "done":
			result.Done = true
		case "notData":
			msg := strings.TrimSpace(data)
			if msg == "" {
				msg = "notData"
			}
			return &RequestError{
				APIError: APIError{
					Code:      "notData",
					Message:   msg,
					Reason:    "xiaozhi_not_data",
					Retryable: false,
				},
			}
		case "error":
			msg := strings.TrimSpace(data)
			if msg == "" {
				msg = "xiaozhiChat error"
			}
			return &RequestError{
				APIError: APIError{
					Code:      "error",
					Message:   msg,
					Reason:    "xiaozhi_error",
					Retryable: false,
				},
			}
		}
		return nil
	})
	result.Text = textBuilder.String()
	if err != nil {
		return result, err
	}
	return result, nil
}

// XiaozhiRefs queries GET /note/qryXiaozhiReferences.
func (c *Client) XiaozhiRefs(docIDs []string) ([]XiaozhiReference, error) {
	cleaned := make([]string, 0, len(docIDs))
	for _, id := range docIDs {
		id = strings.TrimSpace(id)
		if id != "" {
			cleaned = append(cleaned, id)
		}
	}
	if len(cleaned) == 0 {
		return nil, fmt.Errorf("docIdList 不能为空")
	}
	q := url.Values{"docIdList": {strings.Join(cleaned, ",")}}
	raw, err := doGet(c, "/note/qryXiaozhiReferences?"+q.Encode())
	if err != nil {
		return nil, err
	}
	if string(raw) == "null" || len(raw) == 0 {
		return []XiaozhiReference{}, nil
	}
	var list []XiaozhiReference
	if err := json.Unmarshal(raw, &list); err != nil {
		return nil, fmt.Errorf("parsing xiaozhi references: %w", err)
	}
	return list, nil
}

// ClearXiaozhiCache calls GET /note/clearXiaozhiCache.
func (c *Client) ClearXiaozhiCache(contextID string) error {
	contextID = strings.TrimSpace(contextID)
	if contextID == "" {
		return fmt.Errorf("contextId 不能为空")
	}
	q := url.Values{"contextId": {contextID}}
	_, err := doGet(c, "/note/clearXiaozhiCache?"+q.Encode())
	return err
}

func normalizeXiaozhiScope(
	knowledgeID *string, knowledgeIDList *[]string,
	noteID *string, noteIDList *[]string,
	directoryID *string, directoryIDList *[]string,
) {
	*knowledgeIDList = cleanStringList(*knowledgeIDList)
	*noteIDList = cleanStringList(*noteIDList)
	*directoryIDList = cleanStringList(*directoryIDList)

	*knowledgeID = strings.TrimSpace(*knowledgeID)
	*noteID = strings.TrimSpace(*noteID)
	*directoryID = strings.TrimSpace(*directoryID)

	if *knowledgeID == "" && len(*knowledgeIDList) == 1 {
		*knowledgeID = (*knowledgeIDList)[0]
		*knowledgeIDList = nil
	}
	if *noteID == "" && len(*noteIDList) == 1 {
		*noteID = (*noteIDList)[0]
		*noteIDList = nil
	}
	if *directoryID == "" && len(*directoryIDList) == 1 {
		*directoryID = (*directoryIDList)[0]
		*directoryIDList = nil
	}
}

func cleanStringList(in []string) []string {
	if len(in) == 0 {
		return nil
	}
	out := make([]string, 0, len(in))
	for _, s := range in {
		s = strings.TrimSpace(s)
		if s != "" {
			out = append(out, s)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func parseReferenceIDs(data string) []string {
	data = strings.TrimSpace(data)
	if data == "" || data == "null" {
		return nil
	}
	// Prefer JSON array of objects with id, or array of strings / numbers.
	var anyJSON any
	if err := json.Unmarshal([]byte(data), &anyJSON); err == nil {
		switch v := anyJSON.(type) {
		case []any:
			ids := make([]string, 0, len(v))
			for _, item := range v {
				switch it := item.(type) {
				case string:
					if strings.TrimSpace(it) != "" {
						ids = append(ids, strings.TrimSpace(it))
					}
				case float64:
					ids = append(ids, fmt.Sprintf("%.0f", it))
				case map[string]any:
					if id, ok := it["id"]; ok {
						ids = append(ids, fmt.Sprint(id))
					}
				}
			}
			return ids
		}
	}
	// Fallback: comma-separated ids.
	parts := strings.Split(data, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
