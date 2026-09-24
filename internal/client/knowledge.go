package client

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// Knowledge is a note collection (笔记集) list item.
type Knowledge struct {
	KnowledgeID        string  `json:"knowledgeId"`
	KnowledgeName      string  `json:"knowledgeName"`
	KnowledgeDetail    string  `json:"knowledgeDetail"`
	KnowledgeAttribute string  `json:"knowledgeAttribute"`
	KnowledgePublic    string  `json:"knowledgePublic"`
	KnowledgeType      string  `json:"knowledgeType"`
	ContentTotal       flexInt `json:"contentTotal"`
	SubscribeTotal     flexInt `json:"subscribeTotal"`
	UserName           string  `json:"userName"`
	CreateTime         string  `json:"createTime"`
	Status             string  `json:"status"`
	StickyFlag         bool    `json:"stickyFlag"`
	// Received-list fields
	KnowledgeEmpowerID       string `json:"knowledgeEmpowerId,omitempty"`
	KnowledgeEmpowerRole     string `json:"knowledgeEmpowerRole,omitempty"`
	KnowledgeEmpowerEditable string `json:"knowledgeEmpowerEditable,omitempty"`
}

// KnowledgePage is a paginated knowledge list.
type KnowledgePage struct {
	PageNum         flexInt     `json:"pageNum"`
	PageSize        flexInt     `json:"pageSize"`
	Total           flexInt     `json:"total"`
	Pages           flexInt     `json:"pages"`
	Size            flexInt     `json:"size"`
	HasNextPage     bool        `json:"hasNextPage"`
	HasPreviousPage bool        `json:"hasPreviousPage"`
	IsFirstPage     bool        `json:"isFirstPage"`
	IsLastPage      bool        `json:"isLastPage"`
	List            []Knowledge `json:"list"`
}

// KnowledgeDetail is POST /note/queryNoteKnowledgeDetail resultObject.
type KnowledgeDetail struct {
	KnowledgeID        string          `json:"knowledgeId"`
	KnowledgeName      string          `json:"knowledgeName"`
	KnowledgeDetail    string          `json:"knowledgeDetail"`
	KnowledgeAttribute string          `json:"knowledgeAttribute"`
	KnowledgePublic    string          `json:"knowledgePublic"`
	KnowledgeType      string          `json:"knowledgeType"`
	ContentTotal       flexInt         `json:"contentTotal"`
	NoteTotal          flexInt         `json:"noteTotal"`
	FileTotal          flexInt         `json:"fileTotal"`
	UserName           string          `json:"userName"`
	CreateTime         string          `json:"createTime"`
	Editable           bool            `json:"editable"`
	Role               string          `json:"role"`
	Status             string          `json:"status"`
	PageInfo           *KnowledgeNotes `json:"pageInfo"`
}

// KnowledgeNoteItem is a note entry inside a knowledge set.
type KnowledgeNoteItem struct {
	KnowledgeDetailID string `json:"knowledgeDetailId"`
	KnowledgeID       string `json:"knowledgeId"`
	NoteID            string `json:"noteId"`
	Name              string `json:"name"`
	Summary           string `json:"summary"`
	SummaryContent    string `json:"summaryContent,omitempty"`
	NoteType          string `json:"noteType"`
	CatalogName       string `json:"catalogName,omitempty"`
	NoteUpdateTime    string `json:"noteUpdateTime"`
	UserName          string `json:"userName"`
	CreateTime        string `json:"createTime"`
	AllowDelete       bool   `json:"allowDelete"`
}

// KnowledgeNotes is the pageInfo block of knowledge detail.
type KnowledgeNotes struct {
	PageNum         flexInt             `json:"pageNum"`
	PageSize        flexInt             `json:"pageSize"`
	Total           flexInt             `json:"total"`
	Pages           flexInt             `json:"pages"`
	Size            flexInt             `json:"size"`
	HasNextPage     bool                `json:"hasNextPage"`
	HasPreviousPage bool                `json:"hasPreviousPage"`
	List            []KnowledgeNoteItem `json:"list"`
}

// KnowledgeCatalog is a directory node under a knowledge set.
type KnowledgeCatalog struct {
	CatalogID       string             `json:"catalogId"`
	KnowledgeID     string             `json:"knowledgeId"`
	CatalogName     string             `json:"catalogName"`
	ParentCatalogID string             `json:"parentCatalogId"`
	DetailTotal     flexInt            `json:"detailTotal"`
	CatalogTotal    flexInt            `json:"catalogTotal"`
	Children        []KnowledgeCatalog `json:"children,omitempty"`
}

// KnowledgeListParams is POST /note/queryNoteKnowledge.
type KnowledgeListParams struct {
	QryType     string `json:"qryType"`
	QryContent  string `json:"qryContent,omitempty"`
	KnowledgeID string `json:"knowledgeId,omitempty"`
	PageNum     int    `json:"pageNum,omitempty"`
	PageSize    int    `json:"pageSize,omitempty"`
}

// KnowledgeReceivedParams is POST /note/queryNoteKnowledgeEmpower.
type KnowledgeReceivedParams struct {
	QryEmpowerToMe   bool   `json:"qryEmpowerToMe"`
	QryEmpowerFromMe bool   `json:"qryEmpowerFromMe"`
	QryContent       string `json:"qryContent,omitempty"`
	PageNum          int    `json:"pageNum,omitempty"`
	PageSize         int    `json:"pageSize,omitempty"`
}

// KnowledgeDetailParams is POST /note/queryNoteKnowledgeDetail.
type KnowledgeDetailParams struct {
	KnowledgeID          interface{} `json:"knowledgeId"`
	DirectoryID          string      `json:"directoryId,omitempty"`
	SearchContent        string      `json:"searchContent,omitempty"`
	PageNum              int         `json:"pageNum,omitempty"`
	PageSize             int         `json:"pageSize,omitempty"`
	NeedSummaryContent   bool        `json:"needSummaryContent"`
	NeedNoteContentTotal bool        `json:"needNoteContentTotal"`
}

// CreateKnowledgeParams is POST /note/createNoteKnowledge.
type CreateKnowledgeParams struct {
	KnowledgeName      string `json:"knowledgeName"`
	KnowledgeDetail    string `json:"knowledgeDetail,omitempty"`
	KnowledgeAttribute string `json:"knowledgeAttribute,omitempty"`
}

// KnowledgeCatalogParams is POST /note/queryKnowledgeCatalog.
type KnowledgeCatalogParams struct {
	KnowledgeID interface{} `json:"knowledgeId"`
	CatalogID   string      `json:"catalogId,omitempty"`
}

// AddKnowledgeCatalogParams is POST /note/addKnowledgeCatalog.
type AddKnowledgeCatalogParams struct {
	KnowledgeID       interface{} `json:"knowledgeId"`
	DirectoryName     string      `json:"directoryName"`
	ParentDirectoryID string      `json:"parentDirectoryId,omitempty"`
}

// MoveNotesToKnowledgeParams is POST /note/moveNotesToKnowledge.
type MoveNotesToKnowledgeParams struct {
	KnowledgeID interface{} `json:"knowledgeId,omitempty"`
	NoteID      string      `json:"noteId,omitempty"`
	NoteIDList  []string    `json:"noteIdList,omitempty"`
	DirectoryID string      `json:"directoryId,omitempty"`
}

// KnowledgeList queries note collections I created.
func (c *Client) KnowledgeList(params KnowledgeListParams) (*KnowledgePage, error) {
	if params.QryType == "" {
		params.QryType = "myCreate"
	}
	if params.PageNum <= 0 {
		params.PageNum = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 10
	}
	raw, err := doPost(c, "/note/queryNoteKnowledge", params)
	if err != nil {
		return nil, err
	}
	return parseKnowledgePage(raw)
}

// KnowledgeReceived queries note collections shared with me.
func (c *Client) KnowledgeReceived(params KnowledgeReceivedParams) (*KnowledgePage, error) {
	params.QryEmpowerToMe = true
	params.QryEmpowerFromMe = false
	if params.PageNum <= 0 {
		params.PageNum = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 10
	}
	raw, err := doPost(c, "/note/queryNoteKnowledgeEmpower", params)
	if err != nil {
		return nil, err
	}
	return parseKnowledgePage(raw)
}

// KnowledgeGet queries knowledge detail (and optionally embedded notes page).
func (c *Client) KnowledgeGet(params KnowledgeDetailParams) (*KnowledgeDetail, error) {
	if params.PageNum <= 0 {
		params.PageNum = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 10
	}
	params.KnowledgeID = coerceKnowledgeID(params.KnowledgeID)
	raw, err := doPost(c, "/note/queryNoteKnowledgeDetail", params)
	if err != nil {
		return nil, err
	}
	var detail KnowledgeDetail
	if err := json.Unmarshal(raw, &detail); err != nil {
		return nil, fmt.Errorf("parsing knowledge detail: %w", err)
	}
	return &detail, nil
}

// CreateKnowledge creates a note collection via POST /note/createNoteKnowledge.
func (c *Client) CreateKnowledge(params CreateKnowledgeParams) (string, error) {
	name := strings.TrimSpace(params.KnowledgeName)
	if name == "" {
		return "", fmt.Errorf("knowledgeName 不能为空")
	}
	params.KnowledgeName = name
	if strings.TrimSpace(params.KnowledgeAttribute) == "" {
		params.KnowledgeAttribute = "private"
	}
	raw, err := doPost(c, "/note/createNoteKnowledge", params)
	if err != nil {
		return "", err
	}
	return parseIDResult(raw, "knowledgeId")
}

// KnowledgeCatalogList lists directories under a knowledge set.
func (c *Client) KnowledgeCatalogList(params KnowledgeCatalogParams) ([]KnowledgeCatalog, error) {
	params.KnowledgeID = coerceKnowledgeID(params.KnowledgeID)
	if strings.TrimSpace(params.CatalogID) == "" {
		params.CatalogID = "-1"
	}
	raw, err := doPost(c, "/note/queryKnowledgeCatalog", params)
	if err != nil {
		return nil, err
	}
	return parseKnowledgeCatalogList(raw)
}

// AddKnowledgeCatalog creates a directory under a knowledge set.
func (c *Client) AddKnowledgeCatalog(params AddKnowledgeCatalogParams) (*KnowledgeCatalog, error) {
	params.KnowledgeID = coerceKnowledgeID(params.KnowledgeID)
	params.DirectoryName = strings.TrimSpace(params.DirectoryName)
	if params.DirectoryName == "" {
		return nil, fmt.Errorf("directoryName 不能为空")
	}
	if strings.TrimSpace(params.ParentDirectoryID) == "" {
		params.ParentDirectoryID = "-1"
	}
	raw, err := doPost(c, "/note/addKnowledgeCatalog", params)
	if err != nil {
		return nil, err
	}
	var cat KnowledgeCatalog
	if err := json.Unmarshal(raw, &cat); err != nil {
		return nil, fmt.Errorf("parsing add catalog result: %w", err)
	}
	return &cat, nil
}

// MoveNotesToKnowledge adds or moves notes into a knowledge set directory.
// Sends APP-ID when ResolveAppID is available; otherwise hints ZHIZAI_APP_ID on missing-app errors.
func (c *Client) MoveNotesToKnowledge(params MoveNotesToKnowledgeParams) error {
	params.KnowledgeID = coerceKnowledgeID(params.KnowledgeID)
	headers := map[string]string{}
	if appID := c.ResolveAppID(); appID != "" {
		headers["APP-ID"] = appID
	}
	_, err := doPostWithHeaders(c, "/note/moveNotesToKnowledge", params, headers)
	if err == nil {
		return nil
	}
	if c.ResolveAppID() != "" {
		return err
	}
	if !strings.Contains(err.Error(), "缺失应用 ID") && !strings.Contains(err.Error(), "应用ID") {
		return err
	}
	return fmt.Errorf("%w；请设置环境变量 ZHIZAI_APP_ID 或先执行 zhizai file upload 以获取 appId", err)
}

// DeleteKnowledgeCatalog deletes a directory (cascades associations, not note bodies).
func (c *Client) DeleteKnowledgeCatalog(directoryID string) error {
	id := strings.TrimSpace(directoryID)
	if id == "" {
		return fmt.Errorf("directoryId 不能为空")
	}
	q := url.Values{}
	q.Set("directoryId", id)
	_, err := doGet(c, "/note/deleteKnowledgeCatalog?"+q.Encode())
	return err
}

func parseKnowledgePage(raw json.RawMessage) (*KnowledgePage, error) {
	var page KnowledgePage
	if err := json.Unmarshal(raw, &page); err != nil {
		return nil, fmt.Errorf("parsing knowledge page: %w", err)
	}
	return &page, nil
}

func parseKnowledgeCatalogList(raw json.RawMessage) ([]KnowledgeCatalog, error) {
	if string(raw) == "null" || len(raw) == 0 {
		return []KnowledgeCatalog{}, nil
	}
	trimmed := strings.TrimSpace(string(raw))
	if strings.HasPrefix(trimmed, "[") {
		var list []KnowledgeCatalog
		if err := json.Unmarshal(raw, &list); err != nil {
			return nil, fmt.Errorf("parsing knowledge catalog list: %w", err)
		}
		return list, nil
	}
	// Some backends wrap as { list: [...] } or a single object.
	var wrap struct {
		List []KnowledgeCatalog `json:"list"`
	}
	if err := json.Unmarshal(raw, &wrap); err == nil && wrap.List != nil {
		return wrap.List, nil
	}
	var single KnowledgeCatalog
	if err := json.Unmarshal(raw, &single); err == nil && single.CatalogID != "" {
		return []KnowledgeCatalog{single}, nil
	}
	return nil, fmt.Errorf("parsing knowledge catalog list: unexpected shape")
}

func parseIDResult(raw json.RawMessage, field string) (string, error) {
	if string(raw) == "null" || len(raw) == 0 {
		return "", fmt.Errorf("empty %s in response", field)
	}
	trimmed := strings.TrimSpace(string(raw))
	if strings.HasPrefix(trimmed, "\"") {
		var s string
		if err := json.Unmarshal(raw, &s); err != nil {
			return "", fmt.Errorf("parsing %s: %w", field, err)
		}
		return strings.TrimSpace(s), nil
	}
	if trimmed != "" && ((trimmed[0] >= '0' && trimmed[0] <= '9') || trimmed[0] == '-') {
		var n json.Number
		if err := json.Unmarshal(raw, &n); err == nil {
			return n.String(), nil
		}
	}
	var obj map[string]interface{}
	if err := json.Unmarshal(raw, &obj); err == nil {
		if v, ok := obj[field]; ok {
			switch t := v.(type) {
			case string:
				return strings.TrimSpace(t), nil
			case float64:
				return strconv.FormatInt(int64(t), 10), nil
			case json.Number:
				return t.String(), nil
			default:
				return fmt.Sprint(t), nil
			}
		}
	}
	return "", fmt.Errorf("parsing %s: unexpected resultObject %s", field, trimmed)
}

func coerceKnowledgeID(id interface{}) interface{} {
	switch v := id.(type) {
	case string:
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			return n
		}
		return v
	default:
		return id
	}
}
