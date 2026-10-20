package validators_test

import (
	"context"
	"testing"

	"github.com/StackGuardian/terraform-provider-stackguardian/internal/validators"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestHasTemplateRevision(t *testing.T) {
	for id, want := range map[string]bool{
		"my-template:1":          true,
		"my-template:latest":     true,
		"my-template:v1.0":       true,
		"/my-org/my-template:3":  true,
		"my-template":            false,
		"/my-org/my-template":    false,
		"my-template:":           false,
		"/org.with:colon/name":   false,
		"/org/name.with.dots:12": true,
	} {
		if got := validators.HasTemplateRevision(id); got != want {
			t.Errorf("HasTemplateRevision(%q) = %v, want %v", id, got, want)
		}
	}
}

func TestWarnUnpinnedTemplateRevision(t *testing.T) {
	for _, tc := range []struct {
		name     string
		value    types.String
		wantWarn bool
	}{
		{"pinned", types.StringValue("my-template:1"), false},
		{"latest", types.StringValue("my-template:latest"), false},
		{"bare", types.StringValue("my-template"), true},
		{"null", types.StringNull(), false},
		{"unknown", types.StringUnknown(), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var resp validator.StringResponse
			validators.WarnUnpinnedTemplateRevision().ValidateString(context.Background(), validator.StringRequest{
				Path:        path.Root("iac_template_id"),
				ConfigValue: tc.value,
			}, &resp)
			if resp.Diagnostics.HasError() {
				t.Fatalf("unexpected error: %v", resp.Diagnostics)
			}
			if got := resp.Diagnostics.WarningsCount() > 0; got != tc.wantWarn {
				t.Errorf("warning = %v, want %v", got, tc.wantWarn)
			}
		})
	}
}
