package workflowtemplate

import (
	"context"
	"strings"

	"github.com/StackGuardian/terraform-provider-stackguardian/internal/constants"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// runtimeSourceDestKindsRequiringAuthRules is the set of source_config_dest_kind values the
// isPrivate/auth rules below apply to. Other kinds (e.g. CONTAINER_REGISTRY, INLINE) don't have
// this isPrivate/auth/repo shape, so validation is a no-op for them.
var runtimeSourceDestKindsRequiringAuthRules = map[string]bool{
	constants.GithubCom:       true,
	constants.GithubAppCustom: true,
	constants.GitOther:        true,
	constants.BitbucketOrg:    true,
	constants.GitlabCom:       true,
	constants.AzureDevops:     true,
	constants.AzureDevopsSp:   true,
}

// ValidateRuntimeSourceAuth validates the is_private/auth relationship on a git-based
// runtime_source, at the attrPath it was read from (e.g. path.Root("runtime_source")).
//
// Rules:
//   - is_private=true requires auth to be set.
//   - Any source_config_dest_kind other than GIT_OTHER requires auth to be set,
//     regardless of is_private — GIT_OTHER is the only kind that can be public/authless.
//   - When auth is set: GIT_OTHER requires a "/secrets/" reference; every other kind
//     requires a reference starting with "/integration" (matches "/integrations/..." and
//     any other value beginning with that substring). For GIT_OTHER, auth may be set with
//     is_private=false too — an access token for a public repository is valid.
func ValidateRuntimeSourceAuth(ctx context.Context, runtimeSourceObj types.Object, attrPath path.Path) diag.Diagnostics {
	var diags diag.Diagnostics

	if runtimeSourceObj.IsNull() || runtimeSourceObj.IsUnknown() {
		return diags
	}

	var rs RuntimeSourceModel
	d := runtimeSourceObj.As(ctx, &rs, basetypes.ObjectAsOptions{
		UnhandledNullAsEmpty:    true,
		UnhandledUnknownAsEmpty: true,
	})
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	if rs.SourceConfigDestKind.IsUnknown() || rs.SourceConfigDestKind.IsNull() {
		return diags
	}

	destKind := rs.SourceConfigDestKind.ValueString()
	if !runtimeSourceDestKindsRequiringAuthRules[destKind] {
		return diags
	}

	var cfg RuntimeSourceConfigModel
	if !rs.Config.IsNull() && !rs.Config.IsUnknown() {
		d := rs.Config.As(ctx, &cfg, basetypes.ObjectAsOptions{
			UnhandledNullAsEmpty:    true,
			UnhandledUnknownAsEmpty: true,
		})
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
	}

	if cfg.IsPrivate.IsUnknown() || cfg.Auth.IsUnknown() {
		return diags
	}

	// ValueBool()/ValueString() return the zero value (false / "") for Null too, matching
	// Python's None-is-falsy semantics for the isPrivate/auth checks below.
	isPrivate := cfg.IsPrivate.ValueBool()
	auth := cfg.Auth.ValueString()
	hasAuth := auth != ""

	configPath := attrPath.AtName("config")

	if (isPrivate || destKind != constants.GitOther) && !hasAuth {
		diags.AddAttributeError(
			configPath.AtName("auth"),
			"Invalid runtime_source Configuration",
			"auth is required for this runtime_source.",
		)
		return diags
	}

	if hasAuth {
		if destKind == constants.GitOther {
			if !strings.HasPrefix(auth, "/secrets/") {
				diags.AddAttributeError(
					configPath.AtName("auth"),
					"Invalid runtime_source Configuration",
					"auth must start with /secrets/ for "+constants.GitOther+".",
				)
			}
		} else if !strings.HasPrefix(auth, "/integration") {
			diags.AddAttributeError(
				configPath.AtName("auth"),
				"Invalid runtime_source Configuration",
				"auth must start with /integration for this source_config_dest_kind.",
			)
		}
	}

	return diags
}
