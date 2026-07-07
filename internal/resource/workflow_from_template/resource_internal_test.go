package workflowfromtemplate

import (
	"context"
	"testing"

	sgsdkgo "github.com/StackGuardian/sg-sdk-go"
)

func TestSplitTemplateOrg(t *testing.T) {
	cases := []struct {
		name        string
		id          string
		callerOrg   string
		wantOrg     string
		wantRevisID string
	}{
		{
			name:        "own template fully-qualified",
			id:          "/demo-org/my-template:3",
			callerOrg:   "demo-org",
			wantOrg:     "demo-org",
			wantRevisID: "my-template:3",
		},
		{
			name:        "shared template from another org",
			id:          "/other-org/their-template:2",
			callerOrg:   "demo-org",
			wantOrg:     "other-org",
			wantRevisID: "their-template:2",
		},
		{
			name:        "bare id without org prefix falls back to caller org",
			id:          "my-template:1",
			callerOrg:   "demo-org",
			wantOrg:     "demo-org",
			wantRevisID: "my-template:1",
		},
		{
			name:        "empty id falls back to caller org",
			id:          "",
			callerOrg:   "demo-org",
			wantOrg:     "demo-org",
			wantRevisID: "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gotOrg, gotRevID := splitTemplateOrg(tc.id, tc.callerOrg)
			if gotOrg != tc.wantOrg || gotRevID != tc.wantRevisID {
				t.Errorf("splitTemplateOrg(%q, %q) = (%q, %q), want (%q, %q)",
					tc.id, tc.callerOrg, gotOrg, gotRevID, tc.wantOrg, tc.wantRevisID)
			}
		})
	}
}

// TestConvertEnvironmentVariablesFromAPI_NilConfig guards the fix for the nil-Config deref:
// an EnvVars whose Config pointer is nil must not panic when flattened (the API can return
// such an entry, and the pointer-slice-to-value-slice conversion could also leave a zero-value
// EnvVars behind). The nil-Config entry flattens to an empty config object rather than crashing.
func TestConvertEnvironmentVariablesFromAPI_NilConfig(t *testing.T) {
	textValue := "v1"
	envVars := []sgsdkgo.EnvVars{
		{Kind: sgsdkgo.EnvVarsKindEnumPlainText, Config: nil}, // must not panic
		{
			Kind:   sgsdkgo.EnvVarsKindEnumPlainText,
			Config: &sgsdkgo.EnvVarConfig{VarName: "GOOD", TextValue: &textValue},
		},
	}

	list, diags := convertEnvironmentVariablesFromAPI(context.Background(), envVars)
	if diags.HasError() {
		t.Fatalf("unexpected diags: %v", diags)
	}
	if list.IsNull() || list.IsUnknown() {
		t.Fatalf("expected a known list, got null/unknown")
	}
	if got := len(list.Elements()); got != 2 {
		t.Fatalf("expected 2 elements (nil-Config entry preserved as empty), got %d", got)
	}
}
