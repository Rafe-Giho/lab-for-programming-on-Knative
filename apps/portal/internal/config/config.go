package config

import (
	"log"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port                     string
	ExecutorMode             string
	RunnerNamespace          string
	RunnerServiceAccountName string
	RunnerImagePullPolicy    string
	MaxCodeBytes             int
	ExecutionTimeoutSeconds  int64
	JobTTLSeconds            int32
	JobPollIntervalMillis    int
	JobWaitTimeoutSeconds    int
	RunnerCPULimit           string
	RunnerMemoryLimit        string
	RunnerCPURequest         string
	RunnerMemoryRequest      string
	KubeconfigPath           string
}

func MustLoad() Config {
	cfg := Config{
		Port:                     getEnv("PORT", "8080"),
		ExecutorMode:             strings.ToLower(getEnv("EXECUTOR_MODE", "mock")),
		RunnerNamespace:          getEnv("RUNNER_NAMESPACE", "code-runner-exec"),
		RunnerServiceAccountName: getEnv("RUNNER_SERVICE_ACCOUNT_NAME", ""),
		RunnerImagePullPolicy:    getEnv("RUNNER_IMAGE_PULL_POLICY", "IfNotPresent"),
		MaxCodeBytes:             getEnvInt("MAX_CODE_BYTES", 20000),
		ExecutionTimeoutSeconds:  int64(getEnvInt("EXECUTION_TIMEOUT_SECONDS", 10)),
		JobTTLSeconds:            int32(getEnvInt("JOB_TTL_SECONDS", 300)),
		JobPollIntervalMillis:    getEnvInt("JOB_POLL_INTERVAL_MILLIS", 1000),
		JobWaitTimeoutSeconds:    getEnvInt("JOB_WAIT_TIMEOUT_SECONDS", 20),
		RunnerCPULimit:           getEnv("RUNNER_CPU_LIMIT", "500m"),
		RunnerMemoryLimit:        getEnv("RUNNER_MEMORY_LIMIT", "512Mi"),
		RunnerCPURequest:         getEnv("RUNNER_CPU_REQUEST", "100m"),
		RunnerMemoryRequest:      getEnv("RUNNER_MEMORY_REQUEST", "192Mi"),
		KubeconfigPath:           getEnv("KUBECONFIG", ""),
	}

	switch cfg.ExecutorMode {
	case "mock", "kubernetes":
	default:
		log.Fatalf("unsupported EXECUTOR_MODE: %s", cfg.ExecutorMode)
	}

	return cfg
}

func (c Config) RuntimeImages() map[string]string {
	return map[string]string{
		"python:3.11": getEnv("PYTHON_IMAGE_3_11", "shinkiho/runner-python:3.11"),
		"java:17":     getEnv("JAVA_IMAGE_17", "shinkiho/runner-java:17"),
		"c:gcc-14":    getEnv("C_IMAGE_GCC_14", "shinkiho/runner-c:gcc-14"),
		"cpp:g++-14":  getEnv("CPP_IMAGE_GXX_14", "shinkiho/runner-cpp:gxx-14"),
	}
}

func getEnv(key, fallback string) string {
	value, ok := os.LookupEnv(key)
	if !ok || strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func getEnvInt(key string, fallback int) int {
	value := getEnv(key, "")
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		log.Fatalf("invalid integer for %s: %v", key, err)
	}
	return parsed
}
