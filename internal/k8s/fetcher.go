package k8s

import (
	"context"

	"golang.org/x/sync/errgroup"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ClusterData struct {
	Deployments  []appsv1.Deployment
	StatefulSets []appsv1.StatefulSet
	DaemonSets   []appsv1.DaemonSet
	ReplicaSets  []appsv1.ReplicaSet
	Pods         []corev1.Pod
	Services     []corev1.Service
	Ingresses    []networkingv1.Ingress
	NetPolicies  []networkingv1.NetworkPolicy
	ConfigMaps   []corev1.ConfigMap
	Secrets      []corev1.Secret
	HPAs         []autoscalingv1.HorizontalPodAutoscaler
	PDBs         []policyv1.PodDisruptionBudget
}

type Fetcher struct {
	client *Client
}

func NewFetcher(client *Client) *Fetcher {
	return &Fetcher{client: client}
}

func (f *Fetcher) FetchAll(ctx context.Context, ns string) (*ClusterData, error) {
	data := &ClusterData{}
	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		list, err := f.client.Clientset.AppsV1().Deployments(ns).List(ctx, metav1.ListOptions{})
		if err == nil {
			data.Deployments = list.Items
		}
		return nil
	})

	g.Go(func() error {
		list, err := f.client.Clientset.AppsV1().StatefulSets(ns).List(ctx, metav1.ListOptions{})
		if err == nil {
			data.StatefulSets = list.Items
		}
		return nil
	})

	g.Go(func() error {
		list, err := f.client.Clientset.AppsV1().DaemonSets(ns).List(ctx, metav1.ListOptions{})
		if err == nil {
			data.DaemonSets = list.Items
		}
		return nil
	})

	g.Go(func() error {
		list, err := f.client.Clientset.AppsV1().ReplicaSets(ns).List(ctx, metav1.ListOptions{})
		if err == nil {
			data.ReplicaSets = list.Items
		}
		return nil
	})

	g.Go(func() error {
		list, err := f.client.Clientset.CoreV1().Pods(ns).List(ctx, metav1.ListOptions{})
		if err == nil {
			data.Pods = list.Items
		}
		return nil
	})

	g.Go(func() error {
		list, err := f.client.Clientset.CoreV1().Services(ns).List(ctx, metav1.ListOptions{})
		if err == nil {
			data.Services = list.Items
		}
		return nil
	})

	g.Go(func() error {
		list, err := f.client.Clientset.NetworkingV1().Ingresses(ns).List(ctx, metav1.ListOptions{})
		if err == nil {
			data.Ingresses = list.Items
		}
		return nil
	})

	g.Go(func() error {
		list, err := f.client.Clientset.NetworkingV1().NetworkPolicies(ns).List(ctx, metav1.ListOptions{})
		if err == nil {
			data.NetPolicies = list.Items
		}
		return nil
	})

	g.Go(func() error {
		list, err := f.client.Clientset.CoreV1().ConfigMaps(ns).List(ctx, metav1.ListOptions{})
		if err == nil {
			data.ConfigMaps = list.Items
		}
		return nil
	})

	g.Go(func() error {
		list, err := f.client.Clientset.CoreV1().Secrets(ns).List(ctx, metav1.ListOptions{})
		if err == nil {
			data.Secrets = list.Items
		}
		return nil
	})

	g.Go(func() error {
		list, err := f.client.Clientset.AutoscalingV1().HorizontalPodAutoscalers(ns).List(ctx, metav1.ListOptions{})
		if err == nil {
			data.HPAs = list.Items
		}
		return nil
	})

	g.Go(func() error {
		list, err := f.client.Clientset.PolicyV1().PodDisruptionBudgets(ns).List(ctx, metav1.ListOptions{})
		if err == nil {
			data.PDBs = list.Items
		}
		return nil
	})

	_ = g.Wait()
	return data, nil
}
