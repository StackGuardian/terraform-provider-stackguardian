package workflowtemplaterevision

import (
	"strings"
	"testing"
)

func TestWfStepsConfigNotAllowedForTerraformDiagnostics(t *testing.T) {
	cases := []struct {
		name             string
		sourceConfigKind string
		hasWfStepsConfig bool
		wantError        string // substring expected in the diagnostic's detail; "" means no error
	}{
		{
			name:             "TERRAFORM with wf_steps_config is rejected",
			sourceConfigKind: "TERRAFORM",
			hasWfStepsConfig: true,
			wantError:        "wf_steps_config is not allowed when source_config_kind is TERRAFORM or OPENTOFU",
		},
		{
			name:             "OPENTOFU with wf_steps_config is rejected",
			sourceConfigKind: "OPENTOFU",
			hasWfStepsConfig: true,
			wantError:        "wf_steps_config is not allowed when source_config_kind is TERRAFORM or OPENTOFU",
		},
		{
			name:             "TERRAFORM without wf_steps_config is fine",
			sourceConfigKind: "TERRAFORM",
			hasWfStepsConfig: false,
			wantError:        "",
		},
		{
			name:             "OPENTOFU without wf_steps_config is fine",
			sourceConfigKind: "OPENTOFU",
			hasWfStepsConfig: false,
			wantError:        "",
		},
		{
			name:             "CUSTOM with wf_steps_config is fine",
			sourceConfigKind: "CUSTOM",
			hasWfStepsConfig: true,
			wantError:        "",
		},
		{
			name:             "CUSTOM without wf_steps_config is fine",
			sourceConfigKind: "CUSTOM",
			hasWfStepsConfig: false,
			wantError:        "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			diags := wfStepsConfigNotAllowedForTerraformDiagnostics(tc.sourceConfigKind, tc.hasWfStepsConfig)

			if tc.wantError == "" {
				if diags.HasError() {
					t.Fatalf("expected no error, got: %v", diags)
				}
				return
			}

			if !diags.HasError() {
				t.Fatalf("expected an error containing %q, got none", tc.wantError)
			}
			found := false
			for _, d := range diags.Errors() {
				if strings.Contains(d.Detail(), tc.wantError) {
					found = true
					break
				}
			}
			if !found {
				t.Fatalf("expected an error containing %q, got: %v", tc.wantError, diags)
			}
		})
	}
}
