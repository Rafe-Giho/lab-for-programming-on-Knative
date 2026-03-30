package execute

import (
	"context"
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/giho/python-runner-portal/internal/config"
	"github.com/giho/python-runner-portal/internal/runtimecatalog"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/client-go/kubernetes"
)

type KubernetesExecutor struct {
	client  kubernetes.Interface
	config  config.Config
	catalog runtimecatalog.Catalog
}

func NewKubernetesExecutor(client kubernetes.Interface, cfg config.Config, catalog runtimecatalog.Catalog) *KubernetesExecutor {
	return &KubernetesExecutor{
		client:  client,
		config:  cfg,
		catalog: catalog,
	}
}

func (e *KubernetesExecutor) Run(ctx context.Context, req RunRequest) (RunResult, error) {
	runtime, err := e.catalog.Find(req.Language, req.Version)
	if err != nil {
		return RunResult{}, err
	}

	started := time.Now().UTC()
	runID := buildRunID(req.Language, req.Version)
	job := e.buildJob(runID, runtime.Image, req.Code)

	created, err := e.client.BatchV1().Jobs(e.config.RunnerNamespace).Create(ctx, job, metav1.CreateOptions{})
	if err != nil {
		return RunResult{}, fmt.Errorf("create job: %w", err)
	}

	_, exitCode, logs, status, err := e.waitForResult(ctx, created.Name)
	if err != nil {
		return RunResult{}, fmt.Errorf("wait for job result: %w", err)
	}

	finished := time.Now().UTC()
	return RunResult{
		RunID:      created.Name,
		Status:     status,
		Backend:    "kubernetes",
		Image:      runtime.Image,
		Stdout:     logs,
		Stderr:     "",
		ExitCode:   exitCode,
		StartedAt:  started,
		FinishedAt: finished,
		DurationMs: finished.Sub(started).Milliseconds(),
	}, nil
}

func (e *KubernetesExecutor) buildJob(runID, image, code string) *batchv1.Job {
	timeout := e.config.ExecutionTimeoutSeconds
	ttl := e.config.JobTTLSeconds
	codeB64 := base64.StdEncoding.EncodeToString([]byte(code))

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      runID,
			Namespace: e.config.RunnerNamespace,
			Labels: map[string]string{
				"app.kubernetes.io/name": "code-runner-job",
				"runner.giho.dev/run-id": runID,
			},
		},
		Spec: batchv1.JobSpec{
			BackoffLimit:            int32Ptr(0),
			TTLSecondsAfterFinished: &ttl,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"runner.giho.dev/run-id": runID,
					},
				},
				Spec: corev1.PodSpec{
					RestartPolicy:               corev1.RestartPolicyNever,
					ActiveDeadlineSeconds:       &timeout,
					AutomountServiceAccountToken: boolPtr(false),
					Containers: []corev1.Container{
						{
							Name:            "runner",
							Image:           image,
							ImagePullPolicy: corev1.PullPolicy(e.config.RunnerImagePullPolicy),
							WorkingDir:      "/workspace",
							Env: []corev1.EnvVar{
								{Name: "CODE_B64", Value: codeB64},
								{Name: "EXEC_TIMEOUT_SECONDS", Value: fmt.Sprintf("%d", e.config.ExecutionTimeoutSeconds)},
								{Name: "TMPDIR", Value: "/workspace/tmp"},
							},
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse(e.config.RunnerCPURequest),
									corev1.ResourceMemory: resource.MustParse(e.config.RunnerMemoryRequest),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse(e.config.RunnerCPULimit),
									corev1.ResourceMemory: resource.MustParse(e.config.RunnerMemoryLimit),
								},
							},
							SecurityContext: &corev1.SecurityContext{
								AllowPrivilegeEscalation: boolPtr(false),
								ReadOnlyRootFilesystem:   boolPtr(true),
								RunAsNonRoot:             boolPtr(true),
								Capabilities: &corev1.Capabilities{
									Drop: []corev1.Capability{"ALL"},
								},
								SeccompProfile: &corev1.SeccompProfile{
									Type: corev1.SeccompProfileTypeRuntimeDefault,
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "workspace",
									MountPath: "/workspace",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "workspace",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
					},
				},
			},
		},
	}

	if e.config.RunnerServiceAccountName != "" {
		job.Spec.Template.Spec.ServiceAccountName = e.config.RunnerServiceAccountName
	}

	return job
}

func (e *KubernetesExecutor) waitForResult(ctx context.Context, jobName string) (podName string, exitCode int32, logs string, status string, err error) {
	ticker := time.NewTicker(time.Duration(e.config.JobPollIntervalMillis) * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return "", 0, "", "timeout", ctx.Err()
		case <-ticker.C:
			pods, listErr := e.client.CoreV1().Pods(e.config.RunnerNamespace).List(ctx, metav1.ListOptions{
				LabelSelector: "job-name=" + jobName,
			})
			if listErr != nil {
				return "", 0, "", "", listErr
			}
			if len(pods.Items) == 0 {
				continue
			}

			pod := pods.Items[0]
			switch pod.Status.Phase {
			case corev1.PodSucceeded:
				status = "succeeded"
			case corev1.PodFailed:
				status = "failed"
			default:
				continue
			}

			logBytes, logErr := e.client.CoreV1().Pods(e.config.RunnerNamespace).GetLogs(pod.Name, &corev1.PodLogOptions{}).DoRaw(ctx)
			if logErr != nil {
				return "", 0, "", "", logErr
			}

			exitCode = 1
			for _, c := range pod.Status.ContainerStatuses {
				if c.Name == "runner" && c.State.Terminated != nil {
					exitCode = c.State.Terminated.ExitCode
					break
				}
			}

			return pod.Name, exitCode, string(logBytes), status, nil
		}
	}
}

func boolPtr(value bool) *bool {
	return &value
}

func int32Ptr(value int32) *int32 {
	return &value
}

var invalidNameChars = regexp.MustCompile(`[^a-z0-9-]+`)

func buildRunID(language, version string) string {
	base := sanitizeName(language) + "-" + sanitizeName(version)
	base = strings.Trim(base, "-")
	if base == "" {
		base = "run"
	}
	return base + "-" + rand.String(5)
}

func sanitizeName(value string) string {
	sanitized := strings.ToLower(value)
	sanitized = strings.ReplaceAll(sanitized, ".", "-")
	sanitized = strings.ReplaceAll(sanitized, "_", "-")
	sanitized = invalidNameChars.ReplaceAllString(sanitized, "-")
	sanitized = strings.Trim(sanitized, "-")
	for strings.Contains(sanitized, "--") {
		sanitized = strings.ReplaceAll(sanitized, "--", "-")
	}
	return sanitized
}
