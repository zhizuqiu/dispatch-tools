package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// PathInfo 是 /file/List 的响应元素，字段名与 Java 服务端保持一致
type PathInfo struct {
	Id           int
	IsDir        bool
	LastModified string
	Length       int64
	Name         string
	Path         string
}

type apiResult struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

const timeLayout = "2006-01-02 15:04:05"

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintf(w, "dispatch server\n\n数据目录: %s\n\n", s.root)
	fmt.Fprintf(w, "上传: dispatch up -a http://<host>:<port>/ ./some.zip\n")
	fmt.Fprintf(w, "列表: dispatch list -a http://<host>:<port>/ -d /some/dir/\n")
	fmt.Fprintf(w, "下载: http://<host>:<port>/data/some/dir/some.zip\n")
}

func (s *Server) handleList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.methodNotAllowed(w, http.MethodGet)
		return
	}

	dirParam := normalizeDir(r.URL.Query().Get("Path"))
	dir, err := s.resolve(dirParam)
	if err != nil {
		s.badRequest(w, err)
		return
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		// 目录不存在或不是目录时返回空数组，与 Java 服务端行为一致
		s.writeJSON(w, http.StatusOK, []PathInfo{})
		return
	}

	items := make([]PathInfo, 0, len(entries))
	for i, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}
		name := entry.Name()
		isDir := entry.IsDir()
		itemPath := dirParam + name
		if isDir {
			name += "/"
			itemPath += "/"
		}
		items = append(items, PathInfo{
			Id:           i,
			IsDir:        isDir,
			LastModified: info.ModTime().Format(timeLayout),
			Length:       info.Size(),
			Name:         name,
			Path:         itemPath,
		})
	}

	// 目录在前，其余按名称排序
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].IsDir != items[j].IsDir {
			return items[i].IsDir
		}
		return items[i].Name < items[j].Name
	})

	s.writeJSON(w, http.StatusOK, items)
}

func (s *Server) handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.methodNotAllowed(w, http.MethodPost)
		return
	}

	dir, err := s.resolve(normalizeDir(r.URL.Query().Get("Path")))
	if err != nil {
		s.badRequest(w, err)
		return
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		s.serverError(w, err)
		return
	}

	reader, err := r.MultipartReader()
	if err != nil {
		s.badRequest(w, errors.New("请求不是 multipart/form-data"))
		return
	}

	saved := 0
	for {
		part, err := reader.NextPart()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			s.serverError(w, err)
			return
		}
		if part.FileName() == "" {
			_ = part.Close()
			continue
		}
		if err := s.savePart(dir, part); err != nil {
			_ = part.Close()
			if errors.Is(err, errFileExists) {
				s.writeJSON(w, http.StatusConflict, apiResult{Code: http.StatusConflict, Message: err.Error()})
				return
			}
			if errors.Is(err, errInvalidPath) {
				s.badRequest(w, err)
				return
			}
			s.serverError(w, err)
			return
		}
		_ = part.Close()
		saved++
	}

	if saved == 0 {
		s.badRequest(w, errors.New("没有上传文件"))
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = io.WriteString(w, "上传成功")
}

// savePart 把 multipart 中的文件流式写入目标目录
func (s *Server) savePart(dir string, part *multipart.Part) error {
	base := filepath.Base(filepath.FromSlash(part.FileName()))
	if base == "" || base == "." || base == ".." || base == string(filepath.Separator) {
		return errInvalidPath
	}
	dst := filepath.Join(dir, base)

	if s.noOverwrite {
		if _, err := os.Stat(dst); err == nil {
			return errFileExists
		}
	}

	tmp, err := os.CreateTemp(dir, ".dispatch-upload-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	_, copyErr := io.Copy(tmp, part)
	closeErr := tmp.Close()
	if copyErr != nil {
		_ = os.Remove(tmpName)
		return copyErr
	}
	if closeErr != nil {
		_ = os.Remove(tmpName)
		return closeErr
	}
	if err := os.Rename(tmpName, dst); err != nil {
		_ = os.Remove(tmpName)
		return err
	}
	return nil
}

func (s *Server) handleNewDir(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		s.methodNotAllowed(w, http.MethodPut)
		return
	}

	p := r.URL.Query().Get("Path")
	if p == "" {
		s.badRequest(w, errors.New("缺少 Path 参数"))
		return
	}
	full, err := s.resolve(p)
	if err != nil {
		s.badRequest(w, err)
		return
	}
	if err := os.MkdirAll(full, 0755); err != nil {
		s.serverError(w, err)
		return
	}
	s.writeJSON(w, http.StatusOK, apiResult{Code: http.StatusOK})
}

func (s *Server) handleDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		s.methodNotAllowed(w, http.MethodDelete)
		return
	}

	p := r.URL.Query().Get("Path")
	if p == "" {
		s.badRequest(w, errors.New("缺少 Path 参数"))
		return
	}
	full, err := s.resolve(p)
	if err != nil {
		s.badRequest(w, err)
		return
	}
	if full == s.root {
		s.badRequest(w, errors.New("不能删除数据根目录"))
		return
	}

	isDir := r.URL.Query().Get("IsDir") == "true"
	if isDir {
		err = os.RemoveAll(full)
	} else {
		err = os.Remove(full)
	}
	if err != nil && !os.IsNotExist(err) {
		s.serverError(w, err)
		return
	}
	s.writeJSON(w, http.StatusOK, apiResult{Code: http.StatusOK})
}

func (s *Server) handleDownload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		s.methodNotAllowed(w, "GET, HEAD")
		return
	}

	rel := strings.TrimPrefix(r.URL.Path, "/data")
	if rel == "" || strings.HasSuffix(rel, "/") {
		http.NotFound(w, r)
		return
	}
	full, err := s.resolve(rel)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	info, err := os.Stat(full)
	if err != nil || info.IsDir() {
		http.NotFound(w, r)
		return
	}
	f, err := os.Open(full)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer func() { _ = f.Close() }()

	// ServeContent 负责 Content-Length、Last-Modified、Range 和 HEAD
	http.ServeContent(w, r, info.Name(), info.ModTime(), f)
}

func (s *Server) writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func (s *Server) badRequest(w http.ResponseWriter, err error) {
	s.writeJSON(w, http.StatusBadRequest, apiResult{Code: http.StatusBadRequest, Message: err.Error()})
}

func (s *Server) serverError(w http.ResponseWriter, err error) {
	s.logger.Printf("error: %v", err)
	s.writeJSON(w, http.StatusInternalServerError, apiResult{Code: http.StatusInternalServerError, Message: err.Error()})
}

func (s *Server) methodNotAllowed(w http.ResponseWriter, allow string) {
	w.Header().Set("Allow", allow)
	s.writeJSON(w, http.StatusMethodNotAllowed, apiResult{Code: http.StatusMethodNotAllowed, Message: "方法不允许"})
}
