package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNoteTypeLabel(t *testing.T) {
	if got := NoteTypeLabel("voice"); got != "录音" {
		t.Fatalf("got %q", got)
	}
	if got := NoteTypeLabel("unknown"); got != "unknown" {
		t.Fatalf("got %q", got)
	}
}

func TestNoteListAndGet(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/note/queryNoteList", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "test-key" {
			t.Fatalf("missing auth header")
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"resultCode": "0",
			"resultMsg":  "success",
			"resultObject": map[string]interface{}{
				"pageNum":     1,
				"pageSize":    20,
				"total":       "1",
				"hasNextPage": false,
				"list": []map[string]interface{}{
					{"id": "30480", "title": "通话测试", "note_type": "voice", "note_state": "completed", "create_time": "2026-01-26 10:52:40"},
				},
			},
		})
	})
	mux.HandleFunc("/note/querySingleNoteDetail", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("noteId") != "30480" {
			t.Fatalf("noteId = %q", r.URL.Query().Get("noteId"))
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"resultCode": "0",
			"resultMsg":  "success",
			"resultObject": map[string]interface{}{
				"id": "30480", "title": "通话测试确认", "note_type": "voice",
				"note_state": "completed", "summary": "## 会议目标", "abstract": "设备测试",
				"create_time": "2026-01-26 10:52:40",
			},
		})
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	c := NewWithOptions(server.URL, "test-key", server.Client())
	list, err := c.NoteList(NoteListParams{PageNum: 1, PageSize: 20})
	if err != nil {
		t.Fatal(err)
	}
	if len(list.List) != 1 || list.List[0].ID != "30480" {
		t.Fatalf("list = %+v", list)
	}

	note, err := c.NoteGet("30480")
	if err != nil {
		t.Fatal(err)
	}
	if note.Title != "通话测试确认" || note.NoteType != "voice" {
		t.Fatalf("note = %+v", note)
	}
}

func TestNoteCreateUpdateDeleteStatus(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/note/createNote", func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["noteType"] != "text" {
			t.Fatalf("noteType=%v", body["noteType"])
		}
		tc, _ := body["textContent"].(map[string]interface{})
		if tc["content"] != "hello" {
			t.Fatalf("textContent=%v", tc)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"resultCode": "0",
			"resultMsg":  "success",
			"resultObject": map[string]interface{}{
				"id": "9001", "title": "t", "note_type": "text", "note_state": "pending",
			},
		})
	})
	mux.HandleFunc("/note/updateNoteInfo", func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["noteId"] != "9001" || body["title"] != "新标题" {
			t.Fatalf("update body=%v", body)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"resultCode": "0", "resultMsg": "success", "resultObject": nil,
		})
	})
	mux.HandleFunc("/note/deleteNote", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("noteId") != "9001" {
			t.Fatalf("noteId=%q", r.URL.Query().Get("noteId"))
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"resultCode": "0", "resultMsg": "success", "resultObject": nil,
		})
	})
	mux.HandleFunc("/note/queryNoteStatus", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"resultCode": "0",
			"resultMsg":  "success",
			"resultObject": map[string]interface{}{
				"noteState": "completed",
			},
		})
	})

	server := httptest.NewServer(mux)
	defer server.Close()
	c := NewWithOptions(server.URL, "test-key", server.Client())

	created, err := c.NoteCreate(NoteCreateParams{
		NoteType:    "text",
		TextContent: &NoteTextContent{Title: "t", Content: "hello"},
	})
	if err != nil || created.ID != "9001" {
		t.Fatalf("create: %+v err=%v", created, err)
	}
	if err := c.NoteUpdate(NoteUpdateParams{NoteID: "9001", Title: "新标题"}); err != nil {
		t.Fatal(err)
	}
	st, err := c.NoteStatus("9001")
	if err != nil || st.NoteState != "completed" {
		t.Fatalf("status=%+v err=%v", st, err)
	}
	if err := c.NoteDelete("9001"); err != nil {
		t.Fatal(err)
	}
	if NoteStateLabel("completed") != "已完成" {
		t.Fatal("NoteStateLabel")
	}
}

func TestPingUnauthorized(t *testing.T) {
	c := NewWithOptions("http://example.invalid", "", nil)
	if err := c.Ping(); err == nil {
		t.Fatal("expected missing api key error")
	}
}

func TestIsRetryableNetworkError(t *testing.T) {
	if !isRetryableNetworkError(fmt.Errorf(`read tcp 10.0.0.1:1->1.1.1.1:443: read: connection reset by peer`)) {
		t.Fatal("connection reset should be retryable")
	}
	if isRetryableNetworkError(fmt.Errorf("invalid character")) {
		t.Fatal("parse error should not be retryable")
	}
}

func TestDoRetriesConnectionReset(t *testing.T) {
	attempts := 0
	mux := http.NewServeMux()
	mux.HandleFunc("/note/queryNoteList", func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			hj, ok := w.(http.Hijacker)
			if !ok {
				t.Fatal("hijack unsupported")
			}
			conn, _, err := hj.Hijack()
			if err != nil {
				t.Fatal(err)
			}
			_ = conn.Close()
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"resultCode": "0",
			"resultMsg":  "success",
			"resultObject": map[string]interface{}{
				"pageNum": 1, "pageSize": 1, "total": "0", "list": []any{},
			},
		})
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	c := NewWithOptions(server.URL, "test-key", server.Client())
	if _, err := c.NoteList(NoteListParams{PageNum: 1, PageSize: 1}); err != nil {
		t.Fatalf("expected retry success, got %v", err)
	}
	if attempts != 3 {
		t.Fatalf("attempts = %d", attempts)
	}
}

func TestNoteListLCDPBaseStripsDuplicateNote(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/app/note/queryNoteList", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"resultCode": "0",
			"resultMsg":  "success",
			"resultObject": map[string]interface{}{
				"pageNum": 1, "pageSize": 1, "total": "0", "list": []any{},
			},
		})
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	c := NewWithOptions(server.URL+"/app/note", "test-key", server.Client())
	if _, err := c.NoteList(NoteListParams{PageNum: 1, PageSize: 1}); err != nil {
		t.Fatalf("LCDP-style base failed: %v", err)
	}
}
