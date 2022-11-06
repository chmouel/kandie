package kandie

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type KubeClient struct {
	clientset      *kubernetes.Clientset
	kubeConfigPath string
	kubeContext    string
	namespace      string
}

func (k *KubeClient) podList(ctx context.Context) (*corev1.PodList, error) {
	return k.clientset.CoreV1().Pods(k.namespace).List(ctx, metav1.ListOptions{})
}

func (k *KubeClient) create() error {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()

	if k.kubeConfigPath != "" {
		loadingRules.ExplicitPath = k.kubeConfigPath
	}
	configOverrides := &clientcmd.ConfigOverrides{}
	if k.kubeContext != "" {
		configOverrides.CurrentContext = k.kubeContext
	}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	if k.namespace == "" {
		namespace, _, err := kubeConfig.Namespace()
		if err != nil {
			return fmt.Errorf("Couldn't get kubeConfiguration namespace: %w", err)
		}
		k.namespace = namespace
	}
	config, err := kubeConfig.ClientConfig()
	if err != nil {
		return fmt.Errorf("Couldn't get kubeConfiguration: %w", err)
	}
	k.clientset, err = kubernetes.NewForConfig(config)
	return err
}
