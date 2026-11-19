package service

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

// ServerConfig 是文件服务端的配置
type ServerConfig struct {
	// Root 是本地数据根目录，URL 中的 /temp/a.zip 对应 Root/temp/a.zip
	Root string
	// Listen 是监听地址，例如 127.0.0.1:8080
	Listen string
	// User/Password 非空时启用 Basic Auth
	User     string
	Password string
	// NoOverwrite 为 true 时，上传同名文件返回 409，而不是覆盖
	NoOverwrite bool
	// Logger 为空时输出到标准输出
	Logger *log.Logger
}

// Server 是对外提供 dispatch 协议的文件服务端
type Server struct {
	root        string
	listen      string
	user        string
	password    string
	noOverwrite bool
	logger      *log.Logger
}

var (
	errInvalidPath = errors.New("非法的路径参数")
	errFileExists  = errors.New("文件已存在")
)

// NewServer 校验并准备数据目录，返回可直接使用的服务端
func NewServer(cfg ServerConfig) (*Server, error) {
	if cfg.Root == "" {
		cfg.Root = "."
	}
	root, err := filepath.Abs(cfg.Root)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(root, 0755); err != nil {
		return nil, fmt.Errorf("创建数据目录 %s 失败: %w", root, err)
	}

	logger := cfg.Logger
	if logger == nil {
		logger = log.New(os.Stdout, "", log.LstdFlags)
	}

	listen := cfg.Listen
	if listen == "" {
		listen = "127.0.0.1:8080"
	}

	return &Server{
		root:        root,
		listen:      listen,
		user:        cfg.User,
		password:    cfg.Password,
		noOverwrite: cfg.NoOverwrite,
		logger:      logger,
	}, nil
}

// Handler 返回带认证和访问日志的 HTTP handler
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/file/List", s.handleList)
	mux.HandleFunc("/file/UploadFile", s.handleUpload)
	mux.HandleFunc("/file/newDir", s.handleNewDir)
	mux.HandleFunc("/file/Delete", s.handleDelete)
	mux.HandleFunc("/data/", s.handleDownload)
	mux.HandleFunc("/", s.handleIndex)
	return s.withLogging(s.withAuth(mux))
}

// Run 启动服务，收到信号或 ctx 取消时优雅退出
func (s *Server) Run(ctx context.Context) error {
	srv := &http.Server{
		Addr:              s.listen,
		Handler:           s.Handler(),
		ReadHeaderTimeout: 30 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		auth := "关闭"
		if s.authEnabled() {
			auth = "开启"
		}
		s.logger.Printf("dispatch server 已启动: http://%s  数据目录: %s  认证: %s", s.listen, s.root, auth)
		errCh <- srv.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		s.logger.Println("正在关闭 dispatch server ...")
		return srv.Shutdown(shutdownCtx)
	}
}

// RunServer 创建服务端并阻塞运行，直到收到 SIGINT/SIGTERM
func RunServer(cfg ServerConfig) error {
	s, err := NewServer(cfg)
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	return s.Run(ctx)
}

func (s *Server) authEnabled() bool {
	return s.user != "" || s.password != ""
}

func (s *Server) withAuth(next http.Handler) http.Handler {
	if !s.authEnabled() {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, password, ok := r.BasicAuth()
		if !ok || !secureEqual(user, s.user) || !secureEqual(password, s.password) {
			w.Header().Set("WWW-Authenticate", `Basic realm="dispatch"`)
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func secureEqual(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

type statusRecorder struct {
	http.ResponseWriter
	status int
	bytes  int64
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *statusRecorder) Write(p []byte) (int, error) {
	n, err := r.ResponseWriter.Write(p)
	r.bytes += int64(n)
	return n, err
}

func (s *Server) withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		s.logger.Printf("%s %s %d %s %s", r.Method, r.URL.RequestURI(), rec.status,
			humanBytes(rec.bytes), time.Since(start).Round(time.Millisecond))
	})
}

// resolve 把 URL 中的路径转换为本地路径，并保证不会逃出数据根目录
func (s *Server) resolve(urlPath string) (string, error) {
	if urlPath == "" {
		urlPath = "/"
	}
	if !strings.HasPrefix(urlPath, "/") {
		urlPath = "/" + urlPath
	}
	for _, seg := range strings.Split(urlPath, "/") {
		if seg == ".." {
			return "", errInvalidPath
		}
	}

	full := filepath.Join(s.root, filepath.FromSlash(path.Clean(urlPath)))
	rel, err := filepath.Rel(s.root, full)
	if err != nil {
		return "", errInvalidPath
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", errInvalidPath
	}
	return full, nil
}

func normalizeDir(dir string) string {
	if dir == "" {
		dir = "/"
	}
	if !strings.HasPrefix(dir, "/") {
		dir = "/" + dir
	}
	if !strings.HasSuffix(dir, "/") {
		dir = dir + "/"
	}
	return dir
}

func humanBytes(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%dB", n)
	}
	div, exp := int64(unit), 0
	for m := n / unit; m >= unit; m /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%cB", float64(n)/float64(div), "KMGTPE"[exp])
}
