package execute

import (
	"context"
	"fmt"
	"time"

	"github.com/giho/python-runner-portal/internal/runtimecatalog"
)

type MockExecutor struct {
	catalog runtimecatalog.Catalog
}

func NewMockExecutor(catalog runtimecatalog.Catalog) *MockExecutor {
	return &MockExecutor{catalog: catalog}
}

func (e *MockExecutor) Run(_ context.Context, req RunRequest) (RunResult, error) {
	runtime, err := e.catalog.Find(req.Language, req.Version)
	if err != nil {
		return RunResult{}, err
	}

	started := time.Now().UTC()
	finished := time.Now().UTC()

	return RunResult{
		RunID:      fmt.Sprintf("mock-%d", started.UnixNano()),
		Status:     "succeeded",
		Backend:    "mock",
		Image:      runtime.Image,
		Stdout:     fmt.Sprintf("mock mode\nselected runtime: %s %s\nimage: %s\n\nsubmitted code:\n%s\n", req.Language, req.Version, runtime.Image, req.Code),
		Stderr:     "EXECUTOR_MODE=mock is enabled. Set EXECUTOR_MODE=kubernetes for real Job execution.\n",
		ExitCode:   0,
		StartedAt:  started,
		FinishedAt: finished,
		DurationMs: finished.Sub(started).Milliseconds(),
	}, nil
}
