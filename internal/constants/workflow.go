package constants

// SecretReferenceSyntax documents the `${secret::<name>}` form the platform
// resolves at run time. The separator is a double colon, not a dot -- verified
// against live workflow payloads.
const SecretReferenceSyntax = "The value may contain a `${secret::<secret-name>}` reference, which StackGuardian resolves at run time; write it as `$${secret::<secret-name>}` in Terraform so the `$` is not read as an interpolation."

// Workflow resource attributes
const (
	WorkflowWorkflowGroupId = "ID of the parent workflow group."
	// WorkflowFromTemplateWorkflowGroupId is workflow_group_id for the workflow_from_template
	// resource, where it is immutable (changing it forces recreation).
	WorkflowFromTemplateWorkflowGroupId = "ID of the parent workflow group. Immutable — changing this forces the workflow to be recreated (destroy and create), as the platform has no operation to move a workflow between groups."
	WorkflowType                        = "How this workflow is executed. <ul><li>`TERRAFORM` — run with Terraform.</li><li>`OPENTOFU` — run with OpenTofu.</li><li>`CUSTOM` — run the steps in `wf_steps_config` yourself, rather than a built-in engine. Templates of other kinds (Helm, Ansible, Kubectl, CloudFormation) run as `CUSTOM` workflows.</li></ul>This is a smaller set than a template's `source_config_kind`, which describes what the template contains rather than how the workflow runs."
	WorkflowRunnerConstraints           = "Runner constraints to control which runner executes the workflow."
	WorkflowVcsConfig                   = "VCS configuration for the workflow."
	WorkflowIacVcsConfig                = "IaC VCS configuration for the workflow."
	WorkflowUseMarketplaceTemplate      = "Whether to use a marketplace template."
	WorkflowIacTemplateId               = "Workflow template revision this workflow is created from. <ul><li>`&lt;template-name&gt;:&lt;revision&gt;` — a template in your own organization.</li><li>`/&lt;org&gt;/&lt;template-name&gt;:&lt;revision&gt;` — a template owned by another organization: one shared with you, or published publicly. StackGuardian's own templates use the `stackguardian` org, for example `/stackguardian/aws-s3-demo-website:16`.</li></ul>A bare id is resolved against your own organization. Use `:latest` in place of a revision number to track the most recently published revision; pin an explicit revision when the workflow must not move."
	WorkflowCustomSource                = "Custom VCS source configuration."
	WorkflowIacInputData                = "IaC input data for the workflow."
	WorkflowIacInputDataSchemaId        = "Schema ID for the input data."
	WorkflowIacInputDataSchemaType      = "How the value in `data` is formatted. <ul><li>`FORM_JSONSCHEMA` — a StackGuardian NoCode form; `data` holds the values that form collects.</li><li>`RAW_HCL` — HCL-formatted input, as you would write in a `.tfvars` file.</li><li>`RAW_JSON` — the same input expressed as JSON.</li><li>`NONE` — the workflow takes no inputs.</li></ul>The platform also returns `NO_CODE_JSON` on some existing workflows. The provider passes that value through unchanged, but it is outside the set above and should not be written by hand."
	WorkflowIacInputDataData            = "Input data as a JSON string. " + SecretReferenceSyntax
)
