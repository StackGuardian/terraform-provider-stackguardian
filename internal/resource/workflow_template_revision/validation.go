package workflowtemplaterevision

import (
	"fmt"

	"github.com/StackGuardian/sg-sdk-go/workflowtemplates"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// validateTemplateIdUnchanged returns a diagnostic if template_id differs between plan and
// state. A revision cannot be moved to another template, and replacing it would give it a
// new identity, so the change is rejected. Unknown plan values are skipped.
func validateTemplateIdUnchanged(planTemplateId, stateTemplateId types.String) diag.Diagnostics {
	var diags diag.Diagnostics

	if planTemplateId.IsUnknown() || planTemplateId.ValueString() == stateTemplateId.ValueString() {
		return diags
	}

	diags.AddAttributeError(
		path.Root("template_id"),
		"template_id cannot be changed",
		fmt.Sprintf(
			"template_id is immutable on an existing revision (changed from %q to %q). Create a new stackguardian_workflow_template_revision under the desired template instead.",
			stateTemplateId.ValueString(), planTemplateId.ValueString(),
		),
	)
	return diags
}

// wfStepsConfigNotAllowedForTerraformDiagnostics returns a diagnostic if wf_steps_config is
// set while source_config_kind is TERRAFORM or OPENTOFU — those kinds use fixed, built-in
// run steps instead. Returns an empty diag.Diagnostics when the combination is fine.
func wfStepsConfigNotAllowedForTerraformDiagnostics(sourceConfigKind string, hasWfStepsConfig bool) diag.Diagnostics {
	var diags diag.Diagnostics

	isTerraformOrOpentofu := sourceConfigKind == string(workflowtemplates.WorkflowTemplateSourceConfigKindTerraform) ||
		sourceConfigKind == string(workflowtemplates.WorkflowTemplateSourceConfigKindOpentofu)

	if isTerraformOrOpentofu && hasWfStepsConfig {
		diags.AddAttributeError(
			path.Root("wf_steps_config"),
			"Invalid Attribute Combination",
			"wf_steps_config is not allowed when source_config_kind is TERRAFORM or OPENTOFU; those workflow types use fixed, built-in run steps instead.",
		)
	}

	return diags
}
