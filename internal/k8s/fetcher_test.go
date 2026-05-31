package k8s_test

import (
	"context"
	"testing"

	"github.com/samaasi/kubectl-plan/internal/k8s"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestFetchAll(t *testing.T) {
	fakeClient := fake.NewSimpleClientset(
		&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{Name: "payment-api", Namespace: "default"},
		},
	)

	client := &k8s.Client{
		Clientset: fakeClient,
	}

	fetcher := k8s.NewFetcher(client)
	data, err := fetcher.FetchAll(context.Background(), "default")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(data.Deployments) != 1 {
		t.Errorf("expected 1 deployment, got %d", len(data.Deployments))
	}
	if data.Deployments[0].Name != "payment-api" {
		t.Errorf("expected deployment 'payment-api', got %q", data.Deployments[0].Name)
	}
}
