package workflowtemplate

import (
	"context"
	"strings"
	"testing"

	"github.com/StackGuardian/terraform-provider-stackguardian/internal/constants"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// mustRuntimeSourceObject builds a runtime_source types.Object with the given
// source_config_dest_kind, config.is_private, and config.auth (authNull=true leaves auth
// unset, matching a practitioner who never wrote the attribute). Every other config field is
// left Null, mirroring a minimal practitioner config.
func mustRuntimeSourceObject(t *testing.T, destKind string, isPrivate bool, auth string, authNull bool) types.Object {
	t.Helper()
	ctx := context.Background()

	authValue := types.StringValue(auth)
	if authNull {
		authValue = types.StringNull()
	}

	cfg := RuntimeSourceConfigModel{
		IsPrivate:               types.BoolValue(isPrivate),
		Auth:                    authValue,
		GitCoreAutoCrlf:         types.BoolNull(),
		GitSparseCheckoutConfig: types.StringNull(),
		IncludeSubModule:        types.BoolNull(),
		Ref:                     types.StringNull(),
		Repo:                    types.StringValue("https://example.com/repo.git"),
		WorkingDir:              types.StringNull(),
	}
	cfgObj, diags := types.ObjectValueFrom(ctx, RuntimeSourceConfigModel{}.AttributeTypes(), cfg)
	if diags.HasError() {
		t.Fatalf("building config object: %v", diags)
	}

	rs := RuntimeSourceModel{
		SourceConfigDestKind: types.StringValue(destKind),
		Config:               cfgObj,
	}
	rsObj, diags := types.ObjectValueFrom(ctx, RuntimeSourceModel{}.AttributeTypes(), rs)
	if diags.HasError() {
		t.Fatalf("building runtime_source object: %v", diags)
	}
	return rsObj
}

func TestValidateRuntimeSourceAuth(t *testing.T) {
	ctx := context.Background()
	attrPath := path.Root("runtime_source")

	cases := []struct {
		name      string
		obj       types.Object
		wantError string // substring expected in the single diagnostic's detail; "" means no error
	}{
		{
			name:      "null runtime_source is a no-op",
			obj:       types.ObjectNull(RuntimeSourceModel{}.AttributeTypes()),
			wantError: "",
		},
		{
			name:      "dest kind outside the rule set is a no-op regardless of auth",
			obj:       mustRuntimeSourceObject(t, "CONTAINER_REGISTRY", true, "", true),
			wantError: "",
		},
		{
			name:      "is_private true requires auth",
			obj:       mustRuntimeSourceObject(t, constants.GitOther, true, "", true),
			wantError: "auth is required for this runtime_source",
		},
		{
			name:      "non-GIT_OTHER requires auth even when is_private is false",
			obj:       mustRuntimeSourceObject(t, constants.GithubCom, false, "", true),
			wantError: "auth is required for this runtime_source",
		},
		{
			name:      "GIT_OTHER auth must start with /secrets/",
			obj:       mustRuntimeSourceObject(t, constants.GitOther, true, "/integrations/oops", false),
			wantError: "auth must start with /secrets/ for " + constants.GitOther,
		},
		{
			name:      "non-GIT_OTHER auth must start with /integration",
			obj:       mustRuntimeSourceObject(t, constants.GithubCom, true, "/secrets/oops", false),
			wantError: "auth must start with /integration for this source_config_dest_kind",
		},
		{
			name:      "GIT_OTHER private repo with a correctly-prefixed secret is valid",
			obj:       mustRuntimeSourceObject(t, constants.GitOther, true, "/secrets/my-token", false),
			wantError: "",
		},
		{
			name:      "GIT_OTHER public repo with no auth is valid",
			obj:       mustRuntimeSourceObject(t, constants.GitOther, false, "", true),
			wantError: "",
		},
		{
			// A public repo can still use an access token — not the is_private/auth
			// contradiction the validator rejects; GIT_OTHER allows auth regardless of
			// is_private.
			name:      "GIT_OTHER public repo with auth is valid",
			obj:       mustRuntimeSourceObject(t, constants.GitOther, false, "/secrets/my-token", false),
			wantError: "",
		},
		{
			name:      "non-GIT_OTHER private repo with a correctly-prefixed integration is valid",
			obj:       mustRuntimeSourceObject(t, constants.GithubCom, true, "/integrations/my-conn", false),
			wantError: "",
		},
		{
			// The prefix check is "/integration", not the exact literal "/integrations/" —
			// this value has no trailing "s" or slash and must still pass.
			name:      "non-GIT_OTHER auth prefix check is loose, not the exact literal /integrations/",
			obj:       mustRuntimeSourceObject(t, constants.GithubCom, true, "/integration-hub/my-conn", false),
			wantError: "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			diags := ValidateRuntimeSourceAuth(ctx, tc.obj, attrPath)

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
