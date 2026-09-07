package constants

// RuntimeReferencesGuide is the guide documenting every `${...}` form the platform
// resolves at run time -- secrets, external secrets, workflow outputs and stack
// template outputs -- along with the `$$` escaping Terraform requires.
const RuntimeReferencesGuide = "https://registry.terraform.io/providers/StackGuardian/stackguardian/latest/docs/guides/RuntimeReferences"

// RuntimeReferenceNote is appended to every attribute whose value the platform
// resolves at run time. The forms themselves are stated once, in the guide, so an
// attribute description cannot drift from them.
const RuntimeReferenceNote = "May contain a reference the platform resolves at run time. See the [Runtime References guide](" + RuntimeReferencesGuide + ")."

// Workflow resource attributes
const (
	WorkflowWorkflowGroupId = "Workflow group the workflow lives in, as its bare `id` (e.g. `platform`; the full path `platform/networking` for a nested group) — not `/wfgrps/…`."
	// WorkflowFromTemplateWorkflowGroupId is workflow_group_id for the workflow_from_template
	// resource, where it is immutable (changing it forces recreation).
	WorkflowFromTemplateWorkflowGroupId = "Workflow group the workflow lives in, as its bare `id` (e.g. `platform`; the full path `platform/networking` for a nested group) — not `/wfgrps/…`. Immutable — changing this forces the workflow to be recreated (destroy and create), as the platform has no operation to move a workflow between groups."
	WorkflowType                        = "How this workflow is executed. <ul><li>`TERRAFORM` — run with Terraform.</li><li>`OPENTOFU` — run with OpenTofu.</li><li>`CUSTOM` — run the steps in `wf_steps_config` yourself, rather than a built-in engine. Templates of other kinds (Helm, Ansible, Kubectl, CloudFormation) run as `CUSTOM` workflows.</li></ul>This is a smaller set than a template's `source_config_kind`, which describes what the template contains rather than how the workflow runs."
	WorkflowRunnerConstraints           = "Runner constraints to control which runner executes the workflow."
	WorkflowVcsConfig                   = "VCS configuration for the workflow."
	WorkflowIacVcsConfig                = "IaC VCS configuration for the workflow."
	WorkflowUseMarketplaceTemplate      = "Whether to use a marketplace template."
	WorkflowIacTemplateId               = "Workflow template revision this workflow is created from. <ul><li>`<template-name>:<revision>` — a template in your own organization.</li><li>`/<org>/<template-name>:<revision>` — a template owned by another organization: one shared with you, or published publicly. StackGuardian's own templates use the `stackguardian` org, for example `/stackguardian/aws-s3-demo-website:16`.</li></ul>A bare id is resolved against your own organization. Use `:latest` in place of a revision number to track the most recently published revision; pin an explicit revision when the workflow must not move."
	WorkflowCustomSource                = "Custom VCS source configuration."
	WorkflowIacInputData                = "IaC input data for the workflow."
	WorkflowIacInputDataSchemaId        = "Schema ID for the input data."
	WorkflowIacInputDataSchemaType      = "How the value in `data` is formatted. <ul><li>`FORM_JSONSCHEMA` — a StackGuardian NoCode form; `data` holds the values that form collects.</li><li>`RAW_HCL` — HCL-formatted input, as you would write in a `.tfvars` file.</li><li>`RAW_JSON` — the same input expressed as JSON.</li><li>`NONE` — the workflow takes no inputs.</li></ul>The platform also returns `NO_CODE_JSON` on some existing workflows. The provider passes that value through unchanged, but it is outside the set above and should not be written by hand."
	WorkflowIacInputDataData            = "Input data as a JSON string. " + RuntimeReferenceNote
)
