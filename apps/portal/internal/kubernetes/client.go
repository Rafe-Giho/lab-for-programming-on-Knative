package kubernetes

import (
	"fmt"
	"os"

	"github.com/giho/python-runner-portal/internal/config"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func NewClient(cfg config.Config) (kubernetes.Interface, error) {
	restConfig, err := buildConfig(cfg)
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(restConfig)
}

func buildConfig(cfg config.Config) (*rest.Config, error) {
	if os.Getenv("KUBERNETES_SERVICE_HOST") != "" {
		return rest.InClusterConfig()
	}
	if cfg.KubeconfigPath != "" {
		return clientcmd.BuildConfigFromFlags("", cfg.KubeconfigPath)
	}
	return nil, fmt.Errorf("kubernetes configuration not available")
}
