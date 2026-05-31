package output_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/samaasi/kubectl-plan/internal/analysis"
	"github.com/samaasi/kubectl-plan/internal/dependency"
	"github.com/samaasi/kubectl-plan/internal/output"
	"github.com/samaasi/kubectl-plan/internal/risk"
)

func TestRenderer_RenderDoctor(t *testing.T) {
	var buf bytes.Buffer
	renderer := output.NewRenderer("terminal", &buf, true) // ascii true for simpler matching

	res := &output.DoctorResult{
		Namespace:           "production",
		K8sAPIReachable:     true,
		EstimatedConfidence: 0.65,
	}

	err := renderer.RenderDoctor(res)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "ANALYSIS READINESS") {
		t.Errorf("expected output to contain 'ANALYSIS READINESS', got:\n%s", out)
	}
	if !strings.Contains(out, "reachable") {
		t.Errorf("expected output to contain 'reachable', got:\n%s", out)
	}
}

func TestRenderer_RenderTerminal(t *testing.T) {
	var buf bytes.Buffer
	renderer := output.NewRenderer("terminal", &buf, true)

	res := &analysis.AnalysisResult{
		Action: "scale",
		Target: dependency.Node{Kind: "Deployment", Name: "app", Namespace: "default"},
		Risk: risk.RiskScore{
			Score: 5.0,
			Level: risk.RiskMedium,
		},
		Confidence: analysis.OverallConfidence{
			Overall: 0.8,
			Sources: []string{"test"},
		},
		Uncertainty: risk.UncertaintyScore{
			Level:   risk.UncertaintyMedium,
			Reasons: []string{"test reason"},
		},
		Graph: dependency.DependencyGraph{
			Target: dependency.Node{Kind: "Deployment", Name: "app", Namespace: "default"},
		},
	}

	err := renderer.RenderTerminal(res)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "RISK SCORE:") {
		t.Errorf("expected output to contain 'RISK SCORE:', got:\n%s", out)
	}
}
