package workflowtemplate

import (
	"context"
	"strings"
	"testing"

	"github.com/StackGuardian/terraform-provider-stackguardian/internal/constants"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
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

func TestValidateIdUnchanged(t *testing.T) {
	cases := []struct {
		name      string
		plan      types.String
		state     types.String
		wantError string // substring expected in the diagnostic's detail; "" means no error
	}{
		{
			name:      "unchanged id is fine",
			plan:      types.StringValue("template-a"),
			state:     types.StringValue("template-a"),
			wantError: "",
		},
		{
			name:      "unknown plan value is skipped",
			plan:      types.StringUnknown(),
			state:     types.StringValue("template-a"),
			wantError: "",
		},
		{
			name:      "changed id is rejected",
			plan:      types.StringValue("template-b"),
			state:     types.StringValue("template-a"),
			wantError: `id is immutable on an existing workflow template (changed from "template-a" to "template-b")`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			diags := validateIdUnchanged(tc.plan, tc.state)

			if tc.wantError == "" {
				if diags.HasError() {
					t.Fatalf("expected no error, got: %v", diags)
				}
				return
			}

			if !diags.HasError() {
				t.Fatalf("expected error containing %q, got none", tc.wantError)
			}
			if detail := diags.Errors()[0].Detail(); !strings.Contains(detail, tc.wantError) {
				t.Fatalf("expected error containing %q, got %q", tc.wantError, detail)
			}
		})
	}
}

// repoTestValue describes how runtime_source is populated in a plan or state for
// TestValidateRuntimeSourceRepoUnchanged.
type repoTestValue int

const (
	repoNullRuntimeSource    repoTestValue = iota // runtime_source = null
	repoUnknownRuntimeSource                      // runtime_source known after apply
	repoUnknownConfig                             // runtime_source.config known after apply
	repoUnknownRepo                               // runtime_source.config.repo known after apply
	repoKnown                                     // runtime_source.config.repo = the given string
)

func repoTestData(t *testing.T, ctx context.Context, kind repoTestValue, repo string) (tfsdk.Plan, tfsdk.State) {
	t.Helper()

	var schemaResp resource.SchemaResponse
	(&workflowTemplateResource{}).Schema(ctx, resource.SchemaRequest{}, &schemaResp)
	sch := schemaResp.Schema

	plan := tfsdk.Plan{Schema: sch, Raw: tftypes.NewValue(sch.Type().TerraformType(ctx), nil)}

	var diags diag.Diagnostics
	runtimeSourcePath := path.Root("runtime_source")
	switch kind {
	case repoNullRuntimeSource:
	case repoUnknownRuntimeSource:
		diags = plan.SetAttribute(ctx, runtimeSourcePath, types.ObjectUnknown(RuntimeSourceModel{}.AttributeTypes()))
	case repoUnknownConfig:
		diags = plan.SetAttribute(ctx, runtimeSourcePath.AtName("config"), types.ObjectUnknown(RuntimeSourceConfigModel{}.AttributeTypes()))
	case repoUnknownRepo:
		diags = plan.SetAttribute(ctx, runtimeSourceRepoPath, types.StringUnknown())
	case repoKnown:
		diags = plan.SetAttribute(ctx, runtimeSourceRepoPath, types.StringValue(repo))
	}
	if diags.HasError() {
		t.Fatalf("building test data: %v", diags)
	}

	return plan, tfsdk.State{Schema: plan.Schema, Raw: plan.Raw}
}

func TestValidateRuntimeSourceRepoUnchanged(t *testing.T) {
	const repoA = "https://github.com/StackGuardian/tf-null-resource.git"
	const repoB = "https://github.com/StackGuardian/terraform-provider-stackguardian.git"

	cases := []struct {
		name      string
		plan      repoTestValue
		planRepo  string
		state     repoTestValue
		stateRepo string
		wantError string // substring expected in the diagnostic's detail; "" means no error
	}{
		{name: "unchanged repo is fine", plan: repoKnown, planRepo: repoA, state: repoKnown, stateRepo: repoA},
		{name: "unknown runtime_source is skipped", plan: repoUnknownRuntimeSource, state: repoKnown, stateRepo: repoA},
		{name: "unknown runtime_source.config is skipped", plan: repoUnknownConfig, state: repoKnown, stateRepo: repoA},
		{name: "unknown repo is skipped", plan: repoUnknownRepo, state: repoKnown, stateRepo: repoA},
		{name: "null runtime_source on both sides is fine", plan: repoNullRuntimeSource, state: repoNullRuntimeSource},
		{
			name: "changed repo is rejected", plan: repoKnown, planRepo: repoB, state: repoKnown, stateRepo: repoA,
			wantError: `runtime_source.config.repo is immutable on an existing resource (changed from "` + repoA + `" to "` + repoB + `")`,
		},
		{
			name: "removed runtime_source is rejected", plan: repoNullRuntimeSource, state: repoKnown, stateRepo: repoA,
			wantError: `(changed from "` + repoA + `" to "")`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			plan, _ := repoTestData(t, ctx, tc.plan, tc.planRepo)
			_, state := repoTestData(t, ctx, tc.state, tc.stateRepo)

			diags := ValidateRuntimeSourceRepoUnchanged(ctx, plan, state)

			if tc.wantError == "" {
				if diags.HasError() {
					t.Fatalf("expected no error, got: %v", diags)
				}
				return
			}

			if !diags.HasError() {
				t.Fatalf("expected error containing %q, got none", tc.wantError)
			}
			if detail := diags.Errors()[0].Detail(); !strings.Contains(detail, tc.wantError) {
				t.Fatalf("expected error containing %q, got %q", tc.wantError, detail)
			}
		})
	}
}
