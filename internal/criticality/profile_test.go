package criticality_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/samaasi/kubectl-plan/internal/criticality"
)

func TestGetMultiplier(t *testing.T) {
	cases := []struct {
		level string
		want  float64
	}{
		{"critical", 1.00},
		{"CRITICAL", 1.00},
		{"high", 0.80},
		{"HIGH", 0.80},
		{"medium", 0.60},
		{"low", 0.30},
		{"minimal", 0.10},
		{"unknown", 0.60},
		{"", 0.60},
	}
	for _, tc := range cases {
		got := criticality.GetMultiplier(tc.level)
		if got != tc.want {
			t.Errorf("GetMultiplier(%q) = %v, want %v", tc.level, got, tc.want)
		}
	}
}

func TestEvaluateNamespace_defaults(t *testing.T) {
	cases := []struct {
		ns        string
		wantLevel string
		wantMult  float64
	}{
		{"production", "medium", 0.60},
		{"production-payments", "medium", 0.60},
		{"stage-env", "low", 0.30}, // contains "stage"
		{"test-env", "low", 0.30},  // contains "test"
		{"dev-local", "minimal", 0.10},
		{"default", "medium", 0.60},
	}
	for _, tc := range cases {
		level, mult := criticality.EvaluateNamespace(tc.ns, nil)
		if level != tc.wantLevel || mult != tc.wantMult {
			t.Errorf("EvaluateNamespace(%q, nil) = (%q, %v), want (%q, %v)",
				tc.ns, level, mult, tc.wantLevel, tc.wantMult)
		}
	}
}

func TestEvaluateNamespace_withProfile(t *testing.T) {
	cfg := &criticality.ProfileConfig{
		Profiles: []criticality.ProfileMatch{
			{Match: "production-payments", Criticality: "critical"},
			{Match: "production-*", Criticality: "high"},
			{Match: "*", Criticality: "low"},
		},
	}

	cases := []struct {
		ns        string
		wantLevel string
		wantMult  float64
	}{
		{"production-payments", "critical", 1.00},
		{"production-checkout", "high", 0.80},
		{"anything-else", "low", 0.30},
	}
	for _, tc := range cases {
		level, mult := criticality.EvaluateNamespace(tc.ns, cfg)
		if level != tc.wantLevel || mult != tc.wantMult {
			t.Errorf("EvaluateNamespace(%q, cfg) = (%q, %v), want (%q, %v)",
				tc.ns, level, mult, tc.wantLevel, tc.wantMult)
		}
	}
}

func TestLoadProfile_missingFile(t *testing.T) {
	cfg, err := criticality.LoadProfile()
	// when no file exists in default locations, returns nil, nil
	if err != nil && cfg != nil {
		t.Errorf("expected nil config when no file present, got %+v", cfg)
	}
}

func TestLoadProfile_validFile(t *testing.T) {
	dir := t.TempDir()
	yaml := `version: "1"
profiles:
  - match: "production-*"
    criticality: critical
`
	path := filepath.Join(dir, "criticality.yaml")
	if err := os.WriteFile(path, []byte(yaml), 0o600); err != nil {
		t.Fatal(err)
	}

	// use exported loadFile path via a temp home dir trick
	t.Setenv("HOME", dir)
	t.Setenv("USERPROFILE", dir)
	if err := os.MkdirAll(filepath.Join(dir, ".kubectl-plan"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".kubectl-plan", "criticality.yaml"), []byte(yaml), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := criticality.LoadProfile()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected config, got nil")
	}
	if len(cfg.Profiles) != 1 {
		t.Errorf("expected 1 profile, got %d", len(cfg.Profiles))
	}
}
