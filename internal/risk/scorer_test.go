package risk

import (
	"testing"

	"github.com/samaasi/kubectl-plan/internal/criticality"
	"github.com/samaasi/kubectl-plan/internal/dependency"
)

func TestRiskScorer(t *testing.T) {
	profile := &criticality.ProfileConfig{
		Version: "1",
		Profiles: []criticality.ProfileMatch{
			{Match: "prod-payments", Criticality: "critical"},
			{Match: "prod-marketing", Criticality: "medium"},
			{Match: "dev-*", Criticality: "minimal"},
		},
	}

	scorer := NewScorer(profile)

	target := dependency.Node{
		Kind:      "Deployment",
		Name:      "payment-api",
		Namespace: "prod-payments",
	}

	graph := dependency.NewDependencyGraph(target)

	graph.AddNode(dependency.Node{
		Kind:      "Service",
		Name:      "payment-svc",
		Namespace: "prod-payments",
	})
	graph.AddEdge(dependency.Edge{
		From:         "prod-payments/Service/payment-svc",
		To:           "prod-payments/Deployment/payment-api",
		Relationship: dependency.RelSelects,
		Depth:        1,
		Confidence:   0.95,
	})

	graph.AddNode(dependency.Node{
		Kind:      "Ingress",
		Name:      "payment-ingress",
		Namespace: "prod-payments",
	})
	graph.AddEdge(dependency.Edge{
		From:         "prod-payments/Ingress/payment-ingress",
		To:           "prod-payments/Service/payment-svc",
		Relationship: dependency.RelRoutes,
		Depth:        2,
		Confidence:   0.95,
	})

	riskScore, uncertainty := scorer.Score(graph)

	if riskScore.Score == 0.0 {
		t.Errorf("expected risk score to be greater than 0, got %.1f", riskScore.Score)
	}

	if uncertainty.Level != UncertaintyHigh {
		t.Errorf("expected uncertainty to be HIGH for topology-only, got %s", uncertainty.Level)
	}
}
