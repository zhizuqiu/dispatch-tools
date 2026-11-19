package service

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type listItem struct {
	Id           int
	IsDir        bool
	LastModified string
	Length       int64
	Name         string
	Path         string
}

func startTestServer(t *testing.T, cfg ServerConfig) *httptest.Server {
	t.Helper()
	s, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	ts := httptest.NewServer(s.Handler())
	t.Cleanup(ts.Close)
	return ts
}

func doUpload(t *testing.T, ts *httptest.Server, path, field, filename string, content []byte, user, password string) *http.Response {
	t.Helper()
	body := new(bytes.Buffer)
	mw := multipart.NewWriter(body)
	part, err := mw.CreateFormFile(field, filename)
	if err != nil {
		t.Fatalf("CreateFormFile: %v", err)
	}
	if _, err := part.Write(content); err != nil {
		t.Fatalf("write part: %v", err)
	}
	if err := mw.Close(); err != nil {
		t.Fatalf("close multipart: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, ts.URL+"/file/UploadFile?Path="+url.QueryEscape(path), body)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())
	if user != "" {
		req.SetBasicAuth(user, password)
	}
	resp, err := ts.Client().Do(req)
	if err != nil {
		t.Fatalf("upload: %v", err)
	}
	resp.Body.Close()
	return resp
}

func doRequest(t *testing.T, ts *httptest.Server, method, rawurl, user, password string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(method, ts.URL+rawurl, http.NoBody)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	if user != "" {
		req.SetBasicAuth(user, password)
	}
	resp, err := ts.Client().Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, rawurl, err)
	}
	return resp
}

func listFiles(t *testing.T, ts *httptest.Server, path string) []listItem {
	t.Helper()
	resp := doRequest(t, ts, http.MethodGet, "/file/List?Path="+url.QueryEscape(path), "", "")
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("list status = %d, want 200", resp.StatusCode)
	}
	var items []listItem
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	return items
}

func TestUploadListDownloadRoundTrip(t *testing.T) {
	root := t.TempDir()
	ts := startTestServer(t, ServerConfig{Root: root})

	if resp := doUpload(t, ts, "/docs", "upfile", "a.txt", []byte("hello"), "", ""); resp.StatusCode != http.StatusOK {
		t.Fatalf("upload status = %d, want 200", resp.StatusCode)
	}

	got, err := os.ReadFile(filepath.Join(root, "docs", "a.txt"))
	if err != nil {
		t.Fatalf("uploaded file not on disk: %v", err)
	}
	if string(got) != "hello" {
		t.Fatalf("file content = %q, want %q", got, "hello")
	}

	items := listFiles(t, ts, "/docs/")
	if len(items) != 1 {
		t.Fatalf("list returned %d items, want 1", len(items))
	}
	item := items[0]
	if item.Name != "a.txt" || item.IsDir || item.Path != "/docs/a.txt" || item.Length != 5 {
		t.Fatalf("unexpected list item: %+v", item)
	}
	if _, err := time.ParseInLocation("2006-01-02 15:04:05", item.LastModified, time.Local); err != nil {
		t.Fatalf("LastModified %q not in expected format: %v", item.LastModified, err)
	}

	resp := doRequest(t, ts, http.MethodGet, "/data/docs/a.txt", "", "")
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("download status = %d, want 200", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read download: %v", err)
	}
	if string(body) != "hello" {
		t.Fatalf("download body = %q, want %q", body, "hello")
	}
	if resp.ContentLength != 5 {
		t.Fatalf("Content-Length = %d, want 5", resp.ContentLength)
	}
}

func TestUploadAcceptsAnyFileFieldName(t *testing.T) {
	root := t.TempDir()
	ts := startTestServer(t, ServerConfig{Root: root})

	if resp := doUpload(t, ts, "/", "upload", "b.bin", []byte("xyz"), "", ""); resp.StatusCode != http.StatusOK {
		t.Fatalf("upload status = %d, want 200", resp.StatusCode)
	}
	if _, err := os.Stat(filepath.Join(root, "b.bin")); err != nil {
		t.Fatalf("file not saved: %v", err)
	}
}

func TestListSortsDirsFirstThenByName(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "b.txt"), []byte("b"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "a.txt"), []byte("a"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(root, "sub"), 0755); err != nil {
		t.Fatal(err)
	}
	ts := startTestServer(t, ServerConfig{Root: root})

	items := listFiles(t, ts, "/")
	names := make([]string, 0, len(items))
	for _, it := range items {
		names = append(names, it.Name)
	}
	want := []string{"sub/", "a.txt", "b.txt"}
	if len(items) != len(want) {
		t.Fatalf("items = %v, want %v", names, want)
	}
	for i := range want {
		if names[i] != want[i] {
			t.Fatalf("items = %v, want %v", names, want)
		}
	}
	if items[0].Path != "/sub/" {
		t.Fatalf("dir Path = %q, want /sub/", items[0].Path)
	}
}

func TestUploadRejectsTraversal(t *testing.T) {
	root := t.TempDir()
	ts := startTestServer(t, ServerConfig{Root: root})
	outside := filepath.Join(filepath.Dir(root), "outside")

	resp := doUpload(t, ts, "/../outside/", "upfile", "evil.txt", []byte("x"), "", "")
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("upload traversal status = %d, want 400", resp.StatusCode)
	}
	if _, err := os.Stat(filepath.Join(outside, "evil.txt")); err == nil {
		t.Fatalf("traversal upload escaped root to %s", outside)
	}

	resp = doRequest(t, ts, http.MethodGet, "/file/List?Path="+url.QueryEscape("/../"), "", "")
	resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("list traversal status = %d, want 400", resp.StatusCode)
	}
}

func TestUploadUsesBaseName(t *testing.T) {
	root := t.TempDir()
	ts := startTestServer(t, ServerConfig{Root: root})

	resp := doUpload(t, ts, "/", "upfile", "../../evil.txt", []byte("x"), "", "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("upload status = %d, want 200", resp.StatusCode)
	}
	if _, err := os.Stat(filepath.Join(root, "evil.txt")); err != nil {
		t.Fatalf("file not saved under base name: %v", err)
	}
	if _, err := os.Stat(filepath.Join(filepath.Dir(root), "evil.txt")); err == nil {
		t.Fatalf("file escaped root")
	}

	resp = doUpload(t, ts, "/", "upfile", "/tmp/abs.txt", []byte("x"), "", "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("absolute filename upload status = %d, want 200", resp.StatusCode)
	}
	if _, err := os.Stat(filepath.Join(root, "abs.txt")); err != nil {
		t.Fatalf("absolute filename not reduced to base name: %v", err)
	}
}

func TestAuth(t *testing.T) {
	root := t.TempDir()
	ts := startTestServer(t, ServerConfig{Root: root, User: "admin", Password: "secret"})

	if err := os.WriteFile(filepath.Join(root, "f.txt"), []byte("data"), 0644); err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name     string
		user     string
		password string
		want     int
	}{
		{"anonymous", "", "", http.StatusUnauthorized},
		{"wrong password", "admin", "bad", http.StatusUnauthorized},
		{"correct", "admin", "secret", http.StatusOK},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			resp := doRequest(t, ts, http.MethodGet, "/file/List?Path=/", c.user, c.password)
			resp.Body.Close()
			if resp.StatusCode != c.want {
				t.Fatalf("status = %d, want %d", resp.StatusCode, c.want)
			}
		})
	}

	t.Run("download requires auth", func(t *testing.T) {
		resp := doRequest(t, ts, http.MethodGet, "/data/f.txt", "", "")
		resp.Body.Close()
		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("status = %d, want 401", resp.StatusCode)
		}
	})

	t.Run("upload requires auth", func(t *testing.T) {
		resp := doUpload(t, ts, "/", "upfile", "anon.txt", []byte("x"), "", "")
		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("status = %d, want 401", resp.StatusCode)
		}
		if _, err := os.Stat(filepath.Join(root, "anon.txt")); err == nil {
			t.Fatalf("unauthenticated upload was saved")
		}
	})
}

func TestUploadNoOverwrite(t *testing.T) {
	root := t.TempDir()
	ts := startTestServer(t, ServerConfig{Root: root, NoOverwrite: true})

	if resp := doUpload(t, ts, "/", "upfile", "a.txt", []byte("one"), "", ""); resp.StatusCode != http.StatusOK {
		t.Fatalf("first upload status = %d, want 200", resp.StatusCode)
	}
	if resp := doUpload(t, ts, "/", "upfile", "a.txt", []byte("two"), "", ""); resp.StatusCode != http.StatusConflict {
		t.Fatalf("second upload status = %d, want 409", resp.StatusCode)
	}
	got, err := os.ReadFile(filepath.Join(root, "a.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "one" {
		t.Fatalf("content = %q, want %q", got, "one")
	}
}

func TestUploadOverwritesByDefault(t *testing.T) {
	root := t.TempDir()
	ts := startTestServer(t, ServerConfig{Root: root})

	if resp := doUpload(t, ts, "/", "upfile", "a.txt", []byte("one"), "", ""); resp.StatusCode != http.StatusOK {
		t.Fatalf("first upload status = %d, want 200", resp.StatusCode)
	}
	if resp := doUpload(t, ts, "/", "upfile", "a.txt", []byte("two"), "", ""); resp.StatusCode != http.StatusOK {
		t.Fatalf("second upload status = %d, want 200", resp.StatusCode)
	}
	got, err := os.ReadFile(filepath.Join(root, "a.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "two" {
		t.Fatalf("content = %q, want %q", got, "two")
	}
}

func TestNewDirAndDelete(t *testing.T) {
	root := t.TempDir()
	ts := startTestServer(t, ServerConfig{Root: root})

	resp := doRequest(t, ts, http.MethodPut, "/file/newDir?Path="+url.QueryEscape("/newdir"), "", "")
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("newDir status = %d, want 200", resp.StatusCode)
	}
	if info, err := os.Stat(filepath.Join(root, "newdir")); err != nil || !info.IsDir() {
		t.Fatalf("newdir not created: %v", err)
	}

	resp = doRequest(t, ts, http.MethodDelete, "/file/Delete?Path=/newdir&IsDir=true", "", "")
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("delete dir status = %d, want 200", resp.StatusCode)
	}
	if _, err := os.Stat(filepath.Join(root, "newdir")); !os.IsNotExist(err) {
		t.Fatalf("newdir still exists: %v", err)
	}

	if err := os.WriteFile(filepath.Join(root, "f.txt"), []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	resp = doRequest(t, ts, http.MethodDelete, "/file/Delete?Path=/f.txt&IsDir=false", "", "")
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("delete file status = %d, want 200", resp.StatusCode)
	}
	if _, err := os.Stat(filepath.Join(root, "f.txt")); !os.IsNotExist(err) {
		t.Fatalf("file still exists: %v", err)
	}

	resp = doRequest(t, ts, http.MethodPut, "/file/newDir?Path="+url.QueryEscape("/../escaped"), "", "")
	resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("newDir traversal status = %d, want 400", resp.StatusCode)
	}
	if _, err := os.Stat(filepath.Join(filepath.Dir(root), "escaped")); err == nil {
		t.Fatalf("newDir escaped root")
	}

	resp = doRequest(t, ts, http.MethodDelete, "/file/Delete?Path=/&IsDir=true", "", "")
	resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("delete root status = %d, want 400", resp.StatusCode)
	}
	if _, err := os.Stat(root); err != nil {
		t.Fatalf("data root was deleted: %v", err)
	}
}

func TestDownloadRejectsDirectoryAndMissingFile(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "sub"), 0755); err != nil {
		t.Fatal(err)
	}
	ts := startTestServer(t, ServerConfig{Root: root})

	for _, p := range []string{"/data/", "/data/sub/", "/data/nope.txt"} {
		resp := doRequest(t, ts, http.MethodGet, p, "", "")
		resp.Body.Close()
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("%s status = %d, want 404", p, resp.StatusCode)
		}
	}
}

func TestRoutingGuardAndIndex(t *testing.T) {
	root := t.TempDir()
	ts := startTestServer(t, ServerConfig{Root: root})

	resp := doRequest(t, ts, http.MethodGet, "/", "", "")
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		t.Fatalf("read index: %v", err)
	}
	if resp.StatusCode != http.StatusOK || !bytes.Contains(body, []byte("dispatch server")) {
		t.Fatalf("index status = %d, body = %q", resp.StatusCode, body)
	}

	resp = doRequest(t, ts, http.MethodGet, "/unknown", "", "")
	resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("unknown path status = %d, want 404", resp.StatusCode)
	}

	resp = doRequest(t, ts, http.MethodPost, "/file/List?Path=/", "", "")
	resp.Body.Close()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("POST list status = %d, want 405", resp.StatusCode)
	}

	resp = doRequest(t, ts, http.MethodGet, "/file/UploadFile?Path=/", "", "")
	resp.Body.Close()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("GET upload status = %d, want 405", resp.StatusCode)
	}
}
