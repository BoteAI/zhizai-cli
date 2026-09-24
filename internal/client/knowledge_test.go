package client

import (
	"encoding/json"
	"testing"
)

func TestParseKnowledgePage(t *testing.T) {
	raw := json.RawMessage(`{
		"pageNum": 1,
		"pageSize": "10",
		"total": "1",
		"pages": 1,
		"size": 1,
		"hasNextPage": false,
		"hasPreviousPage": false,
		"isFirstPage": true,
		"isLastPage": true,
		"list": [{
			"knowledgeId": "10101",
			"knowledgeName": "项目周报笔记集",
			"knowledgeAttribute": "私密",
			"contentTotal": "12",
			"createTime": "2026-05-01 10:00:00"
		}]
	}`)
	page, err := parseKnowledgePage(raw)
	if err != nil {
		t.Fatal(err)
	}
	if int(page.Total) != 1 || len(page.List) != 1 {
		t.Fatalf("unexpected page: %+v", page)
	}
	if page.List[0].KnowledgeID != "10101" {
		t.Fatalf("knowledgeId = %q", page.List[0].KnowledgeID)
	}
	if int(page.List[0].ContentTotal) != 12 {
		t.Fatalf("contentTotal = %d", page.List[0].ContentTotal)
	}
}

func TestCoerceKnowledgeID(t *testing.T) {
	if got := coerceKnowledgeID("10101"); got != int64(10101) {
		t.Fatalf("string id -> %v (%T)", got, got)
	}
	if got := coerceKnowledgeID("abc"); got != "abc" {
		t.Fatalf("non-numeric -> %v", got)
	}
}

func TestParseKnowledgeCatalogList(t *testing.T) {
	list, err := parseKnowledgeCatalogList(json.RawMessage(`[
		{"catalogId":"c1","catalogName":"A","parentCatalogId":"-1","detailTotal":"3"}
	]`))
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 || list[0].CatalogID != "c1" || int(list[0].DetailTotal) != 3 {
		t.Fatalf("%+v", list)
	}

	wrapped, err := parseKnowledgeCatalogList(json.RawMessage(`{"list":[{"catalogId":"c2","catalogName":"B"}]}`))
	if err != nil {
		t.Fatal(err)
	}
	if len(wrapped) != 1 || wrapped[0].CatalogID != "c2" {
		t.Fatalf("%+v", wrapped)
	}
}

