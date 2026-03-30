package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log/slog"
	nethttp "net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/giho/python-runner-portal/internal/config"
	"github.com/giho/python-runner-portal/internal/execute"
	appkube "github.com/giho/python-runner-portal/internal/kubernetes"
	"github.com/giho/python-runner-portal/internal/runtimecatalog"
)

type Server struct {
	config   config.Config
	catalog  runtimecatalog.Catalog
	executor execute.Executor
	logger   *slog.Logger
	tmpl     *template.Template
}

type ViewData struct {
	Versions          []string
	Languages         []string
	RuntimeOptions    template.JS
	RuntimeSummaries  []runtimecatalog.LanguageSummary
	ExecutorMode      string
	CurrentPath       string
}

func NewServer(cfg config.Config, catalog runtimecatalog.Catalog, executor execute.Executor, logger *slog.Logger) (nethttp.Handler, error) {
	tmpl, err := template.ParseGlob(filepath.Join("web", "templates", "*.html"))
	if err != nil {
		return nil, err
	}

	s := &Server{
		config:   cfg,
		catalog:  catalog,
		executor: executor,
		logger:   logger,
		tmpl:     tmpl,
	}

	mux := nethttp.NewServeMux()
	mux.HandleFunc("GET /", s.handleIndex)
	mux.HandleFunc("GET /architecture", s.handleArchitecture)
	mux.HandleFunc("GET /healthz", s.handleHealth)
	mux.HandleFunc("GET /api/runtimes", s.handleRuntimes)
	mux.HandleFunc("POST /api/run", s.handleRun)
	mux.Handle("GET /static/", nethttp.StripPrefix("/static/", nethttp.FileServer(nethttp.Dir(filepath.Join("web", "static")))))
	return s.withLogging(mux), nil
}

func BuildExecutor(cfg config.Config, catalog runtimecatalog.Catalog, logger *slog.Logger) (execute.Executor, error) {
	if cfg.ExecutorMode == "mock" {
		logger.Warn("starting in mock executor mode")
		return execute.NewMockExecutor(catalog), nil
	}

	client, err := appkube.NewClient(cfg)
	if err != nil {
		return nil, err
	}
	return execute.NewKubernetesExecutor(client, cfg, catalog), nil
}

func (s *Server) handleIndex(w nethttp.ResponseWriter, _ *nethttp.Request) {
	data := s.newViewData("/")
	if err := s.tmpl.ExecuteTemplate(w, "index.html", data); err != nil {
		nethttp.Error(w, err.Error(), nethttp.StatusInternalServerError)
	}
}

func (s *Server) handleArchitecture(w nethttp.ResponseWriter, _ *nethttp.Request) {
	data := s.newViewData("/architecture")
	if err := s.tmpl.ExecuteTemplate(w, "architecture.html", data); err != nil {
		nethttp.Error(w, err.Error(), nethttp.StatusInternalServerError)
	}
}

func (s *Server) handleHealth(w nethttp.ResponseWriter, _ *nethttp.Request) {
	writeJSON(w, nethttp.StatusOK, map[string]string{
		"status":          "ok",
		"executorMode":    s.config.ExecutorMode,
		"runnerNamespace": s.config.RunnerNamespace,
	})
}

func (s *Server) handleRuntimes(w nethttp.ResponseWriter, _ *nethttp.Request) {
	writeJSON(w, nethttp.StatusOK, map[string]any{
		"languages": s.catalog.Supported(),
	})
}

func (s *Server) newViewData(currentPath string) ViewData {
	runtimeOptions, err := json.Marshal(s.catalog.Supported())
	if err != nil {
		runtimeOptions = []byte("{}")
	}

	return ViewData{
		Versions:         s.catalog.PythonVersions(),
		Languages:        s.catalog.Languages(),
		RuntimeOptions:   template.JS(runtimeOptions),
		RuntimeSummaries: s.catalog.Summaries(),
		ExecutorMode:     s.config.ExecutorMode,
		CurrentPath:      currentPath,
	}
}

func (s *Server) handleRun(w nethttp.ResponseWriter, r *nethttp.Request) {
	req, err := s.parseRunRequest(r)
	if err != nil {
		writeJSON(w, nethttp.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), time.Duration(s.config.JobWaitTimeoutSeconds)*time.Second)
	defer cancel()

	result, err := s.executor.Run(ctx, req)
	if err != nil {
		s.logger.Error("execution failed", "error", err)
		writeJSON(w, nethttp.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, nethttp.StatusOK, result)
}

func (s *Server) parseRunRequest(r *nethttp.Request) (execute.RunRequest, error) {
	var req execute.RunRequest
	contentType := r.Header.Get("Content-Type")

	if strings.Contains(contentType, "application/json") {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			return execute.RunRequest{}, err
		}
	} else {
		if err := r.ParseForm(); err != nil {
			return execute.RunRequest{}, err
		}
		req = execute.RunRequest{
			Language: r.FormValue("language"),
			Version:  r.FormValue("version"),
			Code:     r.FormValue("code"),
		}
	}

	req.Language = strings.TrimSpace(strings.ToLower(req.Language))
	req.Version = strings.TrimSpace(req.Version)
	req.Code = strings.TrimSpace(req.Code)

	if req.Language == "" {
		req.Language = "python"
	}
	if req.Version == "" {
		return execute.RunRequest{}, errors.New("version is required")
	}
	if req.Code == "" {
		return execute.RunRequest{}, errors.New("code is required")
	}
	if len(req.Code) > s.config.MaxCodeBytes {
		return execute.RunRequest{}, fmt.Errorf("code exceeds max size of %d bytes", s.config.MaxCodeBytes)
	}
	if _, err := s.catalog.Find(req.Language, req.Version); err != nil {
		return execute.RunRequest{}, err
	}

	return req, nil
}

func (s *Server) withLogging(next nethttp.Handler) nethttp.Handler {
	return nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		started := time.Now()
		writer := &statusCapturingResponseWriter{ResponseWriter: w, statusCode: nethttp.StatusOK}
		next.ServeHTTP(writer, r)
		s.logger.Info("http request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", writer.statusCode,
			"duration_ms", time.Since(started).Milliseconds(),
		)
	})
}

type statusCapturingResponseWriter struct {
	nethttp.ResponseWriter
	statusCode int
}

func (w *statusCapturingResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func writeJSON(w nethttp.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}
