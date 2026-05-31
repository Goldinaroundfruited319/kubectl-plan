package main

import (
	"testing"
)

func TestParseArgs(t *testing.T) {
	cases := []struct {
		name      string
		args      []string
		wantKind  string
		wantName  string
		wantError bool
	}{
		{
			name:      "empty args",
			args:      []string{},
			wantError: true,
		},
		{
			name:      "valid single arg",
			args:      []string{"deployment/payment-api"},
			wantKind:  "deployment",
			wantName:  "payment-api",
			wantError: false,
		},
		{
			name:      "invalid single arg",
			args:      []string{"deployment-payment-api"},
			wantError: true,
		},
		{
			name:      "valid two args",
			args:      []string{"deployment", "payment-api"},
			wantKind:  "deployment",
			wantName:  "payment-api",
			wantError: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			kind, name, err := parseArgs(tc.args)
			if (err != nil) != tc.wantError {
				t.Fatalf("parseArgs(%v) error = %v, wantError %v", tc.args, err, tc.wantError)
			}
			if !tc.wantError {
				if kind != tc.wantKind {
					t.Errorf("parseArgs(%v) kind = %v, want %v", tc.args, kind, tc.wantKind)
				}
				if name != tc.wantName {
					t.Errorf("parseArgs(%v) name = %v, want %v", tc.args, name, tc.wantName)
				}
			}
		})
	}
}
