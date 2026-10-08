package workflowtemplate

import (
	"context"
	"fmt"
	"strings"

	"github.com/StackGuardian/terraform-provider-stackguardian/internal/constants"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
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

// validateIdUnchanged errors if id differs between plan and state. The API cannot rename a
// template, and replacing it would delete every revision underneath it, so the change is
// rejected. Unknown plan values are skipped; an id removed from config keeps its state value
// through UseStateForUnknown, so it never reaches here as a change.
func validateIdUnchanged(planId, stateId types.String) diag.Diagnostics {
	var diags diag.Diagnostics

	if planId.IsUnknown() || planId.ValueString() == stateId.ValueString() {
		return diags
	}

	diags.AddAttributeError(
		path.Root("id"),
		"id cannot be changed",
		fmt.Sprintf(
			"id is immutable on an existing workflow template (changed from %q to %q). Create a new stackguardian_workflow_template with the desired id instead.",
			stateId.ValueString(), planId.ValueString(),
		),
	)
	return diags
}

// ValidateSourceConfigKindUnchanged errors if source_config_kind differs between planKind and
// stateKind. Shared by workflow_template and workflow_template_revision's ModifyPlan —
// source_config_kind is immutable on an existing resource: the API has no endpoint to change
// it in place, and replacing the resource would change its identity (a new template, or a new
// revision number), breaking anything that references the old one. resourceLabel names the
// resource in the error message (e.g. "workflow template", "revision") and resourceTypeName
// gives the Terraform type to create instead (e.g. "stackguardian_workflow_template").
func ValidateSourceConfigKindUnchanged(planKind, stateKind types.String, resourceLabel, resourceTypeName string) diag.Diagnostics {
	var diags diag.Diagnostics

	if planKind.IsUnknown() || planKind.ValueString() == stateKind.ValueString() {
		return diags
	}

	diags.AddAttributeError(
		path.Root("source_config_kind"),
		"source_config_kind cannot be changed",
		fmt.Sprintf(
			"source_config_kind is immutable on an existing %s (changed from %q to %q). Create a new %s with the desired source_config_kind instead.",
			resourceLabel, stateKind.ValueString(), planKind.ValueString(), resourceTypeName,
		),
	)
	return diags
}

// runtimeSourceRepoPath is runtime_source.config.repo, the same on workflow_template and
// workflow_template_revision.
var runtimeSourceRepoPath = path.Root("runtime_source").AtName("config").AtName("repo")

// runtimeSourceRepo reads runtime_source.config.repo from data (a tfsdk.Plan or tfsdk.State).
// It returns unknown when repo or either parent is unknown, and null when either parent is
// null. The parents are checked explicitly because GetAttribute (framework v1.19) returns
// null, not unknown, for a child of an unknown parent.
func runtimeSourceRepo(ctx context.Context, data interface {
	GetAttribute(context.Context, path.Path, interface{}) diag.Diagnostics
}) (types.String, diag.Diagnostics) {
	var diags diag.Diagnostics

	for _, parent := range []path.Path{runtimeSourceRepoPath.ParentPath().ParentPath(), runtimeSourceRepoPath.ParentPath()} {
		var obj types.Object
		diags.Append(data.GetAttribute(ctx, parent, &obj)...)
		if diags.HasError() {
			return types.StringNull(), diags
		}
		if obj.IsUnknown() {
			return types.StringUnknown(), diags
		}
		if obj.IsNull() {
			return types.StringNull(), diags
		}
	}

	var repo types.String
	diags.Append(data.GetAttribute(ctx, runtimeSourceRepoPath, &repo)...)
	return repo, diags
}

// ValidateRuntimeSourceRepoUnchanged errors if runtime_source.config.repo differs between
// plan and state. Shared by workflow_template and workflow_template_revision's ModifyPlan —
// repo cannot be changed on an existing template or revision, published or not: the API has
// no endpoint to change it in place, and replacing the resource would change its identity
// (a new template, or a new revision number), breaking anything that references the old one.
// See constants.WorkflowTemplateRuntimeSourceConfigRepo. A plan repo that is unknown — or
// whose runtime_source or config is unknown — is skipped rather than compared as "".
func ValidateRuntimeSourceRepoUnchanged(ctx context.Context, plan tfsdk.Plan, state tfsdk.State) diag.Diagnostics {
	var diags diag.Diagnostics

	planRepo, d := runtimeSourceRepo(ctx, plan)
	diags.Append(d...)
	if diags.HasError() || planRepo.IsUnknown() {
		return diags
	}

	stateRepo, d := runtimeSourceRepo(ctx, state)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	// The update request has no repo field, so any difference would be dropped silently.
	if planRepo.ValueString() != stateRepo.ValueString() {
		diags.AddAttributeError(
			runtimeSourceRepoPath,
			"runtime_source.config.repo cannot be changed",
			"runtime_source.config.repo can only be set when the resource is created; it cannot be added, removed or changed afterwards. Create a new resource with the desired repo instead.",
		)
	}

	return diags
}

// vcsTriggersDestKinds are the runtime_source.source_config_dest_kind values the template
// webhooks/vcs_triggers endpoint can register a webhook for.
var vcsTriggersDestKinds = []string{
	constants.GithubCom,
	constants.GithubAppCustom,
	constants.GitlabCom,
	constants.BitbucketOrg,
	constants.AzureDevops,
	constants.AzureDevopsSp,
}

// validateVcsTriggersRuntimeSource checks that vcs_triggers can be registered: the webhook
// endpoint takes the VCS provider and repository from runtime_source, so runtime_source must
// set a supported source_config_dest_kind and config.repo, and vcs_triggers.type must match the
// dest kind. Unknown values are skipped; they're checked again once known.
func validateVcsTriggersRuntimeSource(ctx context.Context, vcsTriggersObj, runtimeSourceObj types.Object) diag.Diagnostics {
	var diags diag.Diagnostics
	vcsTriggersPath := path.Root("vcs_triggers")

	if vcsTriggersObj.IsNull() || vcsTriggersObj.IsUnknown() || runtimeSourceObj.IsUnknown() {
		return diags
	}
	if runtimeSourceObj.IsNull() {
		diags.AddAttributeError(vcsTriggersPath, "vcs_triggers requires runtime_source",
			"vcs_triggers registers a webhook for the repository in runtime_source; set runtime_source with source_config_dest_kind and config.repo.")
		return diags
	}

	var triggers VCSTriggersModel
	diags.Append(vcsTriggersObj.As(ctx, &triggers, basetypes.ObjectAsOptions{})...)
	var runtimeSource RuntimeSourceModel
	diags.Append(runtimeSourceObj.As(ctx, &runtimeSource, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return diags
	}

	destKindPath := path.Root("runtime_source").AtName("source_config_dest_kind")
	if !runtimeSource.SourceConfigDestKind.IsUnknown() {
		destKind := runtimeSource.SourceConfigDestKind.ValueString()
		supported := false
		for _, k := range vcsTriggersDestKinds {
			if destKind == k {
				supported = true
				break
			}
		}
		if !supported {
			diags.AddAttributeError(destKindPath, "Unsupported source_config_dest_kind for vcs_triggers",
				fmt.Sprintf("vcs_triggers can only be registered when runtime_source.source_config_dest_kind is one of %s, got %q.",
					strings.Join(vcsTriggersDestKinds, ", "), destKind))
		} else if !triggers.Type.IsUnknown() && triggers.Type.ValueString() != destKind {
			diags.AddAttributeError(vcsTriggersPath.AtName("type"), "vcs_triggers.type must match runtime_source",
				fmt.Sprintf("vcs_triggers.type (%q) must equal runtime_source.source_config_dest_kind (%q).", triggers.Type.ValueString(), destKind))
		}
	}

	if !runtimeSource.Config.IsUnknown() {
		repo := types.StringNull()
		if !runtimeSource.Config.IsNull() {
			var cfg RuntimeSourceConfigModel
			diags.Append(runtimeSource.Config.As(ctx, &cfg, basetypes.ObjectAsOptions{})...)
			if diags.HasError() {
				return diags
			}
			repo = cfg.Repo
		}
		if !repo.IsUnknown() && repo.ValueString() == "" {
			diags.AddAttributeError(runtimeSourceRepoPath, "vcs_triggers requires runtime_source.config.repo",
				"vcs_triggers registers a webhook for the repository in runtime_source.config.repo; set it.")
		}
	}

	return diags
}
