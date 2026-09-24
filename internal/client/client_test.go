package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
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
	if NoteStateLabel("recognizing_failed") != "转写失败" {
		t.Fatal("NoteStateLabel recognizing_failed")
	}
}

func TestUploadDownloadAndNoteCreateTypes(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "sample.m4a")
	if err := os.WriteFile(src, []byte("fake-audio"), 0o644); err != nil {
		t.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/file/uploadSingleFile", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "test-key" {
			t.Fatalf("auth=%q", r.Header.Get("Authorization"))
		}
		ct := r.Header.Get("Content-Type")
		if !strings.HasPrefix(ct, "multipart/form-data") {
			t.Fatalf("Content-Type=%q", ct)
		}
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Fatal(err)
		}
		if r.FormValue("compressFile") != "true" {
			t.Fatalf("compressFile=%q", r.FormValue("compressFile"))
		}
		file, hdr, err := r.FormFile("file")
		if err != nil {
			t.Fatal(err)
		}
		defer file.Close()
		if hdr.Filename != "sample.m4a" {
			t.Fatalf("filename=%q", hdr.Filename)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"resultCode": "0",
			"resultMsg":  "success",
			"resultObject": map[string]interface{}{
				"fileId": "fid-1", "fileName": "sample.m4a", "appId": "app-9",
				"storeType": "MINIO", "createDate": "2026-09-24", "statusCd": "00A", "statusDate": "2026-09-24",
			},
		})
	})
	mux.HandleFunc("/file/getFile/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/file/getFile/")
		if id != "fid-1" {
			t.Fatalf("fileId path=%q", r.URL.Path)
		}
		http.Redirect(w, r, "/cdn/fid-1.bin", http.StatusFound)
	})
	mux.HandleFunc("/cdn/fid-1.bin", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		_, _ = w.Write([]byte("downloaded-bytes"))
	})
	mux.HandleFunc("/note/createNote", func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&body)
		switch body["noteType"] {
		case "voice":
			vc, _ := body["voiceContent"].(map[string]interface{})
			if vc["voiceFileId"] != "fid-1" || vc["recordingSource"] != "offlineImport" {
				t.Fatalf("voiceContent=%v", vc)
			}
		case "document":
			dc, _ := body["documentContent"].(map[string]interface{})
			if dc["fileId"] != "fid-1" {
				t.Fatalf("documentContent=%v", dc)
			}
		case "link":
			lc, _ := body["linkContent"].(map[string]interface{})
			if lc["url"] != "https://example.com" {
				t.Fatalf("linkContent=%v", lc)
			}
		case "image":
			ic, _ := body["imageContent"].(map[string]interface{})
			files, _ := ic["fileIds"].([]interface{})
			if len(files) != 1 {
				t.Fatalf("imageContent=%v", ic)
			}
		default:
			t.Fatalf("unexpected noteType=%v", body["noteType"])
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"resultCode": "0", "resultMsg": "success",
			"resultObject": map[string]interface{}{
				"id": "n1", "title": "x", "note_type": body["noteType"], "note_state": "pending",
			},
		})
	})
	mux.HandleFunc("/note/queryNoteList", func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["withShortUrl"] != "true" {
			t.Fatalf("withShortUrl=%v", body["withShortUrl"])
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"resultCode": "0", "resultMsg": "success",
			"resultObject": map[string]interface{}{
				"pageNum": 1, "pageSize": 20, "total": "0", "list": []any{},
			},
		})
	})
	mux.HandleFunc("/note/querySingleNoteDetail", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("withShortUrl") != "true" {
			t.Fatalf("withShortUrl=%q", r.URL.Query().Get("withShortUrl"))
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"resultCode": "0", "resultMsg": "success",
			"resultObject": map[string]interface{}{
				"id": "30480", "title": "t", "short_url": "https://s.example/a",
				"note_type": "voice", "note_state": "completed",
			},
		})
	})
	mux.HandleFunc("/note/qryNoteDetailInfoAndAppend", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("noteId") != "30480" {
			t.Fatalf("noteId=%q", r.URL.Query().Get("noteId"))
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"resultCode": "0", "resultMsg": "success",
			"resultObject": map[string]interface{}{
				"queryMainNoteInfo":  map[string]interface{}{"id": "30480", "title": "main"},
				"queryRecordingNote": []map[string]interface{}{{"id": "30481", "title": "append"}},
			},
		})
	})
	mux.HandleFunc("/note/downloadNoteAudio", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("noteId") != "30480" {
			t.Fatalf("noteId=%q", r.URL.Query().Get("noteId"))
		}
		if r.Header.Get("APP-ID") != "app-9" {
			t.Fatalf("APP-ID=%q", r.Header.Get("APP-ID"))
		}
		w.Header().Set("Content-Type", "audio/mpeg")
		_, _ = w.Write([]byte("mp3-bytes"))
	})

	server := httptest.NewServer(mux)
	defer server.Close()
	c := NewWithOptions(server.URL, "test-key", server.Client())

	up, err := c.UploadFile(src, true)
	if err != nil || up.FileID != "fid-1" || up.AppID != "app-9" {
		t.Fatalf("upload=%+v err=%v", up, err)
	}
	if c.LastAppID() != "app-9" {
		t.Fatalf("lastAppID=%q", c.LastAppID())
	}

	dest := filepath.Join(tmp, "out.bin")
	if err := c.DownloadFile("fid-1", dest); err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(dest)
	if string(got) != "downloaded-bytes" {
		t.Fatalf("downloaded=%q", got)
	}

	if _, err := c.NoteCreate(NoteCreateParams{
		NoteType:     "voice",
		VoiceContent: &NoteVoiceContent{VoiceFileID: "fid-1", Title: "v"},
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := c.NoteCreate(NoteCreateParams{
		NoteType:        "document",
		DocumentContent: &NoteDocumentContent{FileID: "fid-1", FileName: "a.docx"},
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := c.NoteCreate(NoteCreateParams{
		NoteType:    "link",
		LinkContent: &NoteLinkContent{URL: "https://example.com"},
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := c.NoteCreate(NoteCreateParams{
		NoteType: "image",
		ImageContent: &NoteImageContent{FileIds: []NoteImageFile{
			{FileID: "fid-1", Remark: "r"},
		}},
	}); err != nil {
		t.Fatal(err)
	}

	if _, err := c.NoteList(NoteListParams{WithShortUrl: "true"}); err != nil {
		t.Fatal(err)
	}
	note, err := c.NoteGetWithOptions("30480", NoteGetOptions{WithShortUrl: "true"})
	if err != nil || note.ShortURL == "" {
		t.Fatalf("note=%+v err=%v", note, err)
	}
	appendDetail, err := c.NoteGetWithAppend("30480")
	if err != nil || appendDetail.QueryMainNoteInfo == nil || len(appendDetail.QueryRecordingNote) != 1 {
		t.Fatalf("append=%+v err=%v", appendDetail, err)
	}
	audioDest := filepath.Join(tmp, "note.mp3")
	if err := c.DownloadNoteAudio("30480", audioDest); err != nil {
		t.Fatal(err)
	}
	audio, _ := os.ReadFile(audioDest)
	if string(audio) != "mp3-bytes" {
		t.Fatalf("audio=%q", audio)
	}
}

func TestNoteWaitAndDownloadJSONError(t *testing.T) {
	polls := 0
	mux := http.NewServeMux()
	mux.HandleFunc("/note/queryNoteStatus", func(w http.ResponseWriter, r *http.Request) {
		polls++
		state := "pending"
		if polls >= 2 {
			state = "completed"
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"resultCode": "0", "resultMsg": "success",
			"resultObject": map[string]interface{}{"noteState": state},
		})
	})
	mux.HandleFunc("/file/getFile/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"resultCode": "1", "resultMsg": "文件不存在",
		})
	})

	server := httptest.NewServer(mux)
	defer server.Close()
	c := NewWithOptions(server.URL, "test-key", server.Client())

	st, err := c.NoteWait("1", 5*time.Second)
	if err != nil || st.NoteState != "completed" {
		t.Fatalf("wait=%+v err=%v polls=%d", st, err, polls)
	}
	if polls < 2 {
		t.Fatalf("polls=%d", polls)
	}

	err = c.DownloadFile("missing", filepath.Join(t.TempDir(), "x.bin"))
	if err == nil {
		t.Fatal("expected JSON download error")
	}
	var reqErr *RequestError
	if !errors.As(err, &reqErr) {
		t.Fatalf("err type=%T %v", err, err)
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
