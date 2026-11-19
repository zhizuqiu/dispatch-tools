package service

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// Upload 必须在 multipart 头里只发送文件名，不能发送本地完整路径。
// Java（Netty）服务端不会像 Go 的 multipart.Reader 那样自动取 base name，
// 所以这里直接检查原始 Content-Disposition 头。
func TestUploadSendsBaseName(t *testing.T) {
	src := filepath.Join(t.TempDir(), "some.zip")
	if err := os.WriteFile(src, []byte("payload"), 0644); err != nil {
		t.Fatal(err)
	}

	dispositions := make(chan string, 1)
	contents := make(chan string, 1)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mr, err := r.MultipartReader()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		part, err := mr.NextPart()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer part.Close()
		data, err := io.ReadAll(part)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		dispositions <- part.Header.Get("Content-Disposition")
		contents <- string(data)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	Upload(ts.URL, "/", "", "", src)

	select {
	case disposition := <-dispositions:
		want := `form-data; name="upfile"; filename="some.zip"`
		if disposition != want {
			t.Fatalf("Content-Disposition = %q, want %q", disposition, want)
		}
	default:
		t.Fatalf("server did not receive upload")
	}

	select {
	case content := <-contents:
		if content != "payload" {
			t.Fatalf("uploaded content = %q, want %q", content, "payload")
		}
	default:
		t.Fatalf("server did not receive upload body")
	}
}
