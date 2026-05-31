package k8s_test

import (
	"context"
	"testing"

	"github.com/samaasi/kubectl-plan/internal/k8s"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestGetClusterUID(t *testing.T) {
	fakeClient := fake.NewSimpleClientset(
		&corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{Name: "kube-system", UID: "12345-abcde"},
		},
	)

	client := &k8s.Client{
		Clientset: fakeClient,
	}

	uid, err := client.GetClusterUID(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if uid != "12345-abcde" {
		t.Errorf("expected uid '12345-abcde', got %q", uid)
	}
}

func TestGetClusterUID_NotFound(t *testing.T) {
	fakeClient := fake.NewSimpleClientset() // no namespaces

	client := &k8s.Client{
		Clientset: fakeClient,
	}

	uid, err := client.GetClusterUID(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if uid != "unknown-cluster-uid" {
		t.Errorf("expected uid 'unknown-cluster-uid', got %q", uid)
	}
}
