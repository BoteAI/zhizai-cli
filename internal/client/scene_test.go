package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestParseInnerSceneList(t *testing.T) {
	raw := json.RawMessage(`{
		"110019": [{"id":"10721","scene_name":"AI外脑辅助分析","scene_desc":"分析风险"}],
		"800015": [
			{"id":"10383","scene_name":"面试自我复盘","scene_desc":"面试复盘"},
			{"id":"10807","scene_name":"晋级答辩自我复盘","scene_desc":"答辩复盘"}
		],
		"groupInfo": {"110019":"分析","800015":"成长"}
	}`)
	list, err := parseInnerSceneList(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 3 {
		t.Fatalf("len=%d", len(list))
	}
	if list[0].ID != "10721" || list[0].CategoryName != "分析" {
		t.Fatalf("first=%+v", list[0])
	}
	found := MatchSceneByName(list, "面试自我复盘")
	if found == nil || found.ID != "10383" {
		t.Fatalf("match=%+v", found)
	}
}

func TestSceneListMyAndInnerAndCards(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/note/queryMySceneList", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("includeSharedFlag") != "true" {
			t.Fatalf("includeSharedFlag=%q", r.URL.Query().Get("includeSharedFlag"))
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"resultCode": "0", "resultMsg": "success",
			"resultObject": []map[string]string{
				{"id": "10192", "scene_name": "通用会议", "scene_desc": "会议记录"},
			},
		})
	})
	mux.HandleFunc("/note/queryInnerSceneList", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"resultCode": "0", "resultMsg": "success",
			"resultObject": map[string]interface{}{
				"400011": []map[string]string{
					{"id": "1", "scene_name": "会议纪要", "scene_desc": "纪要"},
				},
				"groupInfo": map[string]string{"400011": "会议"},
			},
		})
	})
	mux.HandleFunc("/note/queryKnowledgeCardByPage", func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["pageNum"] != float64(2) || body["pageSize"] != float64(5) {
			t.Fatalf("body=%v", body)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"resultCode": "0", "resultMsg": "success",
			"resultObject": map[string]interface{}{
				"pageNum": 2, "pageSize": 5, "total": "1", "pages": 1, "size": 1,
				"hasNextPage": false, "hasPreviousPage": true, "isFirstPage": false, "isLastPage": true,
				"list": []map[string]interface{}{
					{
						"id": "31943", "name": "记账APP优化会议", "summary": "讨论优化",
						"status": "00A", "create_time": "2026-04-22 23:25:54", "creator_id": "1",
						"cards": []map[string]string{{"id": "c1", "title": "问题", "answer": "A", "remarks": "R"}},
					},
				},
			},
		})
	})

	server := httptest.NewServer(mux)
	defer server.Close()
	c := NewWithOptions(server.URL, "test-key", server.Client())

	my, err := c.SceneListMy(true)
	if err != nil || len(my) != 1 || my[0].ID != "10192" {
		t.Fatalf("my=%+v err=%v", my, err)
	}

	inner, err := c.SceneListInner()
	if err != nil || len(inner) != 1 || inner[0].CategoryName != "会议" {
		t.Fatalf("inner=%+v err=%v", inner, err)
	}

	page, err := c.KnowledgeCards(KnowledgeCardsParams{PageNum: 2, PageSize: 5})
	if err != nil || len(page.List) != 1 || len(page.List[0].Cards) != 1 {
		t.Fatalf("cards=%+v err=%v", page, err)
	}
}

func TestKnowledgeWritePaths(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/note/createNoteKnowledge", func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["knowledgeName"] != "工作笔记集" || body["knowledgeAttribute"] != "private" {
			t.Fatalf("create body=%v", body)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"resultCode": "0", "resultMsg": "success", "resultObject": "14192",
		})
	})
	mux.HandleFunc("/note/queryKnowledgeCatalog", func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["knowledgeId"] != float64(10101) || body["catalogId"] != "-1" {
			t.Fatalf("catalog body=%v", body)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"resultCode": "0", "resultMsg": "success",
			"resultObject": []map[string]interface{}{
				{"catalogId": "c1", "knowledgeId": "10101", "catalogName": "周报", "parentCatalogId": "-1", "detailTotal": "2", "catalogTotal": "0"},
			},
		})
	})
	mux.HandleFunc("/note/addKnowledgeCatalog", func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["directoryName"] != "周报" || body["parentDirectoryId"] != "-1" {
			t.Fatalf("mkdir body=%v", body)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"resultCode": "0", "resultMsg": "success",
			"resultObject": map[string]string{
				"catalogId": "c2", "catalogName": "周报", "knowledgeId": "10101", "parentCatalogId": "-1",
			},
		})
	})
	mux.HandleFunc("/note/moveNotesToKnowledge", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("APP-ID") != "app-1" {
			t.Fatalf("APP-ID=%q", r.Header.Get("APP-ID"))
		}
		var body map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["knowledgeId"] != float64(10101) {
			t.Fatalf("move body=%v", body)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"resultCode": "0", "resultMsg": "success", "resultObject": nil,
		})
	})
	mux.HandleFunc("/note/deleteKnowledgeCatalog", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("directoryId") != "c2" {
			t.Fatalf("directoryId=%q", r.URL.Query().Get("directoryId"))
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"resultCode": "0", "resultMsg": "success", "resultObject": nil,
		})
	})

	server := httptest.NewServer(mux)
	defer server.Close()
	c := NewWithOptions(server.URL, "test-key", server.Client())
	c.SetLastAppID("app-1")

	id, err := c.CreateKnowledge(CreateKnowledgeParams{KnowledgeName: "工作笔记集"})
	if err != nil || id != "14192" {
		t.Fatalf("create id=%q err=%v", id, err)
	}

	cats, err := c.KnowledgeCatalogList(KnowledgeCatalogParams{KnowledgeID: "10101"})
	if err != nil || len(cats) != 1 || cats[0].CatalogID != "c1" {
		t.Fatalf("catalog=%+v err=%v", cats, err)
	}

	cat, err := c.AddKnowledgeCatalog(AddKnowledgeCatalogParams{
		KnowledgeID: "10101", DirectoryName: "周报",
	})
	if err != nil || cat.CatalogID != "c2" {
		t.Fatalf("mkdir=%+v err=%v", cat, err)
	}

	if err := c.MoveNotesToKnowledge(MoveNotesToKnowledgeParams{
		KnowledgeID: "10101", NoteID: "31023", DirectoryID: "c2",
	}); err != nil {
		t.Fatal(err)
	}

	if err := c.DeleteKnowledgeCatalog("c2"); err != nil {
		t.Fatal(err)
	}
}

func TestParseIDResultShapes(t *testing.T) {
	cases := []struct {
		raw  string
		want string
	}{
		{`"14192"`, "14192"},
		{`14192`, "14192"},
		{`{"knowledgeId":"14192"}`, "14192"},
		{`{"knowledgeId":14192}`, "14192"},
	}
	for _, tc := range cases {
		got, err := parseIDResult(json.RawMessage(tc.raw), "knowledgeId")
		if err != nil || got != tc.want {
			t.Fatalf("raw=%s got=%q err=%v", tc.raw, got, err)
		}
	}
}

func TestParseKnowledgeCatalogListShapes(t *testing.T) {
	list, err := parseKnowledgeCatalogList(json.RawMessage(`[{"catalogId":"1","catalogName":"a"}]`))
	if err != nil || len(list) != 1 {
		t.Fatalf("array: %+v %v", list, err)
	}
	list, err = parseKnowledgeCatalogList(json.RawMessage(`{"list":[{"catalogId":"2","catalogName":"b"}]}`))
	if err != nil || list[0].CatalogID != "2" {
		t.Fatalf("wrap: %+v %v", list, err)
	}
	list, err = parseKnowledgeCatalogList(json.RawMessage(`{"catalogId":"3","catalogName":"c"}`))
	if err != nil || list[0].CatalogID != "3" {
		t.Fatalf("single: %+v %v", list, err)
	}
	if _, err := parseKnowledgeCatalogList(json.RawMessage(`"bad"`)); err == nil || !strings.Contains(err.Error(), "unexpected shape") {
		t.Fatalf("expected shape error, got %v", err)
	}
}
