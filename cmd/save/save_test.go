package save

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func touch(t *testing.T, dir, name string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestDetectInputVoice(t *testing.T) {
	dir := t.TempDir()
	p := touch(t, dir, "meeting.mp3")
	kind, paths, _, err := detectInput([]string{p}, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	if kind != "voice" || len(paths) != 1 {
		t.Fatalf("got kind=%q paths=%v", kind, paths)
	}
}

func TestDetectInputMultiImage(t *testing.T) {
	dir := t.TempDir()
	a := touch(t, dir, "a.jpg")
	b := touch(t, dir, "b.png")
	kind, paths, _, err := detectInput([]string{a, b}, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	if kind != "image" || len(paths) != 2 {
		t.Fatalf("got kind=%q paths=%v", kind, paths)
	}
}

func TestDetectInputLink(t *testing.T) {
	kind, _, link, err := detectInput([]string{"https://example.com/a"}, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	if kind != "link" || link != "https://example.com/a" {
		t.Fatalf("got kind=%q link=%q", kind, link)
	}
}

func TestDetectInputText(t *testing.T) {
	kind, _, _, err := detectInput(nil, "明天交周报", nil)
	if err != nil {
		t.Fatal(err)
	}
	if kind != "text" {
		t.Fatalf("got kind=%q", kind)
	}
}

func TestDetectInputVideoRejected(t *testing.T) {
	dir := t.TempDir()
	p := touch(t, dir, "meeting.mp4")
	_, _, _, err := detectInput([]string{p}, "", nil)
	if err == nil {
		t.Fatal("expected video error")
	}
	if !strings.Contains(err.Error(), "视频无法保存为笔记") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDetectInputUnknownRejected(t *testing.T) {
	dir := t.TempDir()
	p := touch(t, dir, "a.zip")
	_, _, _, err := detectInput([]string{p}, "", nil)
	if err == nil {
		t.Fatal("expected unknown type error")
	}
	if !strings.Contains(err.Error(), "不支持的文件类型") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDetectInputMissingFile(t *testing.T) {
	_, _, _, err := detectInput([]string{"./no-such-file.mp3"}, "", nil)
	if err == nil || !strings.Contains(err.Error(), "文件不存在") {
		t.Fatalf("expected missing file error, got %v", err)
	}
}

func TestDetectInputVoiceWithPositionalImages(t *testing.T) {
	dir := t.TempDir()
	a := touch(t, dir, "a.mp3")
	img := touch(t, dir, "p1.jpg")
	kind, paths, _, err := detectInput([]string{a, img}, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	if kind != "voice" || len(paths) != 2 {
		t.Fatalf("got kind=%q paths=%v", kind, paths)
	}
}

func TestResolveTextContentExclusive(t *testing.T) {
	_, err := resolveTextContent("a", "b.txt", false, nil)
	if err == nil {
		t.Fatal("expected exclusive error")
	}
}
