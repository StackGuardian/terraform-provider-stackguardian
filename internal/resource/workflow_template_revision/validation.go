package workflowtemplaterevision

import (
	"github.com/StackGuardian/sg-sdk-go/workflowtemplates"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

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
