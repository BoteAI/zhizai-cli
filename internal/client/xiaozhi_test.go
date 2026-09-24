package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCreateXiaozhiSessionAndChat(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/note/createXiaozhiSession", func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["query"] != "找找关于支付的笔记" {
			t.Fatalf("query=%v", body["query"])
		}
		if body["knowledgeId"] != "kb1" {
			t.Fatalf("knowledgeId=%v", body["knowledgeId"])
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"resultCode": "0",
			"resultMsg":  "success",
			"resultObject": map[string]interface{}{
				"sessionId": "s1", "chatId": "s1", "contextId": "c1", "sessionTitle": "找找关于支付的笔记",
			},
		})
	})
	mux.HandleFunc("/note/xiaozhiChat", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("event: text\ndata: 你好\n\n"))
		_, _ = w.Write([]byte("event: text\ndata: 世界\n\n"))
		_, _ = w.Write([]byte("event: references\ndata: [{\"id\":\"1\"},{\"id\":\"2\"}]\n\n"))
		_, _ = w.Write([]byte("event: done\ndata: [DONE]\n\n"))
	})
	mux.HandleFunc("/note/qryXiaozhiReferences", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("docIdList") != "1,2" {
			t.Fatalf("docIdList=%q", r.URL.Query().Get("docIdList"))
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"resultCode": "0",
			"resultMsg":  "success",
			"resultObject": []map[string]interface{}{
				{"noteId": "n1", "name": "支付", "summary": "摘要"},
			},
		})
	})
	mux.HandleFunc("/note/clearXiaozhiCache", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("contextId") != "c1" {
			t.Fatalf("contextId=%q", r.URL.Query().Get("contextId"))
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"resultCode": "0", "resultMsg": "success", "resultObject": nil,
		})
	})

	server := httptest.NewServer(mux)
	defer server.Close()
	c := NewWithOptions(server.URL, "test-key", server.Client())

	sess, err := c.CreateXiaozhiSession(XiaozhiSessionParams{Query: "找找关于支付的笔记", KnowledgeID: "kb1"})
	if err != nil {
		t.Fatal(err)
	}
	if sess.ChatID != "s1" || sess.ContextID != "c1" {
		t.Fatalf("sess=%+v", sess)
	}

	chat, err := c.XiaozhiChat(XiaozhiChatParams{
		Query: "找找关于支付的笔记", ChatID: "s1", ContextID: "c1", WithReferences: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if chat.Text != "你好世界" {
		t.Fatalf("text=%q", chat.Text)
	}
	if len(chat.ReferenceIDs) != 2 {
		t.Fatalf("refs=%v", chat.ReferenceIDs)
	}

	refs, err := c.XiaozhiRefs(chat.ReferenceIDs)
	if err != nil || len(refs) != 1 || refs[0].NoteID != "n1" {
		t.Fatalf("refs=%v err=%v", refs, err)
	}
	if err := c.ClearXiaozhiCache("c1"); err != nil {
		t.Fatal(err)
	}
}

func TestXiaozhiChatNotData(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/note/xiaozhiChat", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("event: notData\ndata: 范围内没有可用笔记\n\n"))
	})
	server := httptest.NewServer(mux)
	defer server.Close()
	c := NewWithOptions(server.URL, "test-key", server.Client())

	_, err := c.XiaozhiChat(XiaozhiChatParams{Query: "q", ChatID: "s", ContextID: "c"})
	if err == nil {
		t.Fatal("expected error")
	}
	re, ok := err.(*RequestError)
	if !ok || re.Code != "notData" || !strings.Contains(re.Message, "范围内没有可用笔记") {
		t.Fatalf("err=%v", err)
	}
}

func TestCreateTextNoteSummaryAndTemplateAndRecordings(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/note/createTextNoteSummary", func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["content"] != "一段话" || body["sceneId"] != "sc1" {
			t.Fatalf("body=%v", body)
		}
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("data: ## 总结\n\n"))
		_, _ = w.Write([]byte("data: - 要点\n\n"))
	})
	mux.HandleFunc("/know/queryStandardInputOutputByCommand", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("command") != "生成周报" {
			t.Fatalf("command=%q", r.URL.Query().Get("command"))
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"resultCode": "0",
			"resultMsg":  "success",
			"resultObject": map[string]interface{}{
				"input": "本周笔记", "output": "# 周报\n## 概览",
			},
		})
	})
	mux.HandleFunc("/note/querySummaryAndRecording", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("title") != "经营" {
			t.Fatalf("title=%q", r.URL.Query().Get("title"))
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"resultCode": "0",
			"resultMsg":  "success",
			"resultObject": []map[string]interface{}{
				{"noteId": "n9", "noteTitle": "经营会", "noteCreateTime": "2026-09-01", "noteStatus": "completed"},
			},
		})
	})

	server := httptest.NewServer(mux)
	defer server.Close()
	c := NewWithOptions(server.URL, "test-key", server.Client())

	sum, err := c.CreateTextNoteSummary(TextNoteSummaryParams{Content: "一段话", SceneID: "sc1"})
	if err != nil {
		t.Fatal(err)
	}
	if sum != "## 总结- 要点" {
		t.Fatalf("sum=%q", sum)
	}

	tpl, err := c.QueryTemplate("生成周报")
	if err != nil || tpl.Output == "" {
		t.Fatalf("tpl=%v err=%v", tpl, err)
	}

	list, err := c.QuerySummaryAndRecording(SummaryAndRecordingParams{Title: "经营"})
	if err != nil || len(list) != 1 || list[0].NoteID != "n9" {
		t.Fatalf("list=%v err=%v", list, err)
	}
}

func TestCreateXiaozhiSessionEmptyQuery(t *testing.T) {
	c := NewWithOptions("http://example.invalid", "k", nil)
	_, err := c.CreateXiaozhiSession(XiaozhiSessionParams{Query: "  "})
	if err == nil || !strings.Contains(err.Error(), "问题不能为空") {
		t.Fatalf("err=%v", err)
	}
}
