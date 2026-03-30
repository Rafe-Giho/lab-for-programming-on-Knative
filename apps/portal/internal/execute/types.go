package execute

import (
	"context"
	"time"
)

type RunRequest struct {
	Language string `json:"language"`
	Version  string `json:"version"`
	Code     string `json:"code"`
}

type RunResult struct {
	RunID      string    `json:"runId"`
	Status     string    `json:"status"`
	Backend    string    `json:"backend"`
	Image      string    `json:"image"`
	Stdout     string    `json:"stdout"`
	Stderr     string    `json:"stderr"`
	ExitCode   int32     `json:"exitCode"`
	StartedAt  time.Time `json:"startedAt"`
	FinishedAt time.Time `json:"finishedAt"`
	DurationMs int64     `json:"durationMs"`
}

type Executor interface {
	Run(ctx context.Context, req RunRequest) (RunResult, error)
}
