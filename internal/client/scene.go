package client

import (
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strings"
)

// Scene is a summary scene item from my-scene or inner-scene lists.
type Scene struct {
	ID           string `json:"id"`
	SceneName    string `json:"scene_name"`
	SceneDesc    string `json:"scene_desc"`
	SceneRequire string `json:"scene_require,omitempty"`
	CategoryID   string `json:"category_id,omitempty"`
	CategoryName string `json:"category_name,omitempty"`
}

// KnowledgeCardItem is one card inside a knowledge-card note.
type KnowledgeCardItem struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Summary string `json:"summary"`
	Answer  string `json:"answer"`
	Remarks string `json:"remarks"`
}

// KnowledgeCardNote is a note that contains knowledge cards.
type KnowledgeCardNote struct {
	ID              string              `json:"id"`
	Name            string              `json:"name"`
	Summary         string              `json:"summary"`
	SummaryContent  string              `json:"summary_content,omitempty"`
	Status          string              `json:"status"`
	CreateTime      string              `json:"create_time"`
	CreatorID       string              `json:"creator_id"`
	Cards           []KnowledgeCardItem `json:"cards"`
}

// KnowledgeCardsPage is POST /note/queryKnowledgeCardByPage resultObject.
type KnowledgeCardsPage struct {
	PageNum         flexInt             `json:"pageNum"`
	PageSize        flexInt             `json:"pageSize"`
	Total           flexInt             `json:"total"`
	Pages           flexInt             `json:"pages"`
	Size            flexInt             `json:"size"`
	HasNextPage     bool                `json:"hasNextPage"`
	HasPreviousPage bool                `json:"hasPreviousPage"`
	IsFirstPage     bool                `json:"isFirstPage"`
	IsLastPage      bool                `json:"isLastPage"`
	List            []KnowledgeCardNote `json:"list"`
}

// KnowledgeCardsParams is POST /note/queryKnowledgeCardByPage.
type KnowledgeCardsParams struct {
	PageNum  int `json:"pageNum,omitempty"`
	PageSize int `json:"pageSize,omitempty"`
}

// SceneListMy queries GET /note/queryMySceneList.
func (c *Client) SceneListMy(includeShared bool) ([]Scene, error) {
	q := url.Values{}
	if includeShared {
		q.Set("includeSharedFlag", "true")
	} else {
		q.Set("includeSharedFlag", "false")
	}
	raw, err := doGet(c, "/note/queryMySceneList?"+q.Encode())
	if err != nil {
		return nil, err
	}
	if string(raw) == "null" || len(raw) == 0 {
		return []Scene{}, nil
	}
	var list []Scene
	if err := json.Unmarshal(raw, &list); err != nil {
		return nil, fmt.Errorf("parsing my scene list: %w", err)
	}
	return list, nil
}

// SceneListInner queries GET /note/queryInnerSceneList and flattens the category map.
func (c *Client) SceneListInner() ([]Scene, error) {
	raw, err := doGet(c, "/note/queryInnerSceneList")
	if err != nil {
		return nil, err
	}
	return parseInnerSceneList(raw)
}

// KnowledgeCards queries POST /note/queryKnowledgeCardByPage.
func (c *Client) KnowledgeCards(params KnowledgeCardsParams) (*KnowledgeCardsPage, error) {
	if params.PageNum <= 0 {
		params.PageNum = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 10
	}
	raw, err := doPost(c, "/note/queryKnowledgeCardByPage", params)
	if err != nil {
		return nil, err
	}
	var page KnowledgeCardsPage
	if err := json.Unmarshal(raw, &page); err != nil {
		return nil, fmt.Errorf("parsing knowledge cards page: %w", err)
	}
	return &page, nil
}

func parseInnerSceneList(raw json.RawMessage) ([]Scene, error) {
	if string(raw) == "null" || len(raw) == 0 {
		return []Scene{}, nil
	}
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(raw, &obj); err != nil {
		return nil, fmt.Errorf("parsing inner scene list: %w", err)
	}

	groupInfo := map[string]string{}
	if giRaw, ok := obj["groupInfo"]; ok && string(giRaw) != "null" {
		_ = json.Unmarshal(giRaw, &groupInfo)
	}

	categoryIDs := make([]string, 0, len(obj))
	for key := range obj {
		if key == "groupInfo" {
			continue
		}
		categoryIDs = append(categoryIDs, key)
	}
	sort.Strings(categoryIDs)

	out := make([]Scene, 0)
	for _, catID := range categoryIDs {
		var scenes []Scene
		if err := json.Unmarshal(obj[catID], &scenes); err != nil {
			// skip non-array keys silently
			continue
		}
		catName := groupInfo[catID]
		for _, s := range scenes {
			s.CategoryID = catID
			s.CategoryName = catName
			out = append(out, s)
		}
	}
	return out, nil
}

// MatchSceneByName finds a scene by exact scene_name (trimmed).
func MatchSceneByName(scenes []Scene, name string) *Scene {
	want := strings.TrimSpace(name)
	if want == "" {
		return nil
	}
	for i := range scenes {
		if strings.TrimSpace(scenes[i].SceneName) == want {
			return &scenes[i]
		}
	}
	return nil
}
