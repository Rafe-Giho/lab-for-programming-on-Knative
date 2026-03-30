package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	appconfig "github.com/giho/python-runner-portal/internal/config"
	apphttp "github.com/giho/python-runner-portal/internal/http"
	"github.com/giho/python-runner-portal/internal/runtimecatalog"
)

func main() {
	cfg := appconfig.MustLoad()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	catalog := runtimecatalog.New(cfg.RuntimeImages())
	executor, err := apphttp.BuildExecutor(cfg, catalog, logger)
	if err != nil {
		logger.Error("failed to build executor", "error", err)
		os.Exit(1)
	}

	handler, err := apphttp.NewServer(cfg, catalog, executor, logger)
	if err != nil {
		logger.Error("failed to initialize server", "error", err)
		os.Exit(1)
	}

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		logger.Info("starting portal server", "port", cfg.Port, "executor_mode", cfg.ExecutorMode)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server failed", "error", err)
			stop()
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = server.Shutdown(shutdownCtx)
}
