package k8s

import (
	"context"
	"path/filepath"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type Client struct {
	Clientset kubernetes.Interface
	Namespace string
	Context   string
}

func NewClient(contextName, namespace string) (*Client, error) {
	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	if rules.ExplicitPath == "" {
		if home := homedir.HomeDir(); home != "" {
			rules.ExplicitPath = filepath.Join(home, ".kube", "config")
		}
	}

	configOverrides := &clientcmd.ConfigOverrides{}
	if contextName != "" {
		configOverrides.CurrentContext = contextName
	}
	if namespace != "" {
		configOverrides.Context.Namespace = namespace
	}

	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, configOverrides).RawConfig()
	if err != nil {
		return nil, err
	}

	activeContext := config.CurrentContext
	if contextName != "" {
		activeContext = contextName
	}

	activeNamespace := "default"
	if namespace != "" {
		activeNamespace = namespace
	} else if ctx, ok := config.Contexts[activeContext]; ok && ctx.Namespace != "" {
		activeNamespace = ctx.Namespace
	}

	restConfig, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, configOverrides).ClientConfig()
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	return &Client{
		Clientset: clientset,
		Namespace: activeNamespace,
		Context:   activeContext,
	}, nil
}

func (c *Client) GetClusterUID(ctx context.Context) (string, error) {
	ns, err := c.Clientset.CoreV1().Namespaces().Get(ctx, "kube-system", meta.GetOptions{})
	if err == nil && ns.UID != "" {
		return string(ns.UID), nil
	}
	return "unknown-cluster-uid", nil
}
