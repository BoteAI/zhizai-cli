package client

import (
	"encoding/json"
	"fmt"
	"strconv"
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
	SearchContent        string      `json:"searchContent,omitempty"`
	PageNum              int         `json:"pageNum,omitempty"`
	PageSize             int         `json:"pageSize,omitempty"`
	NeedSummaryContent   bool        `json:"needSummaryContent"`
	NeedNoteContentTotal bool        `json:"needNoteContentTotal"`
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

func parseKnowledgePage(raw json.RawMessage) (*KnowledgePage, error) {
	var page KnowledgePage
	if err := json.Unmarshal(raw, &page); err != nil {
		return nil, fmt.Errorf("parsing knowledge page: %w", err)
	}
	return &page, nil
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
