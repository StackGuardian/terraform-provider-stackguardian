package constants

// Workflow resource attributes
const (
	WorkflowWorkflowGroupId = "ID of the parent workflow group."
	// WorkflowFromTemplateWorkflowGroupId is workflow_group_id for the workflow_from_template
	// resource, where it is immutable (changing it forces recreation).
	WorkflowFromTemplateWorkflowGroupId = "ID of the parent workflow group. Immutable — changing this forces the workflow to be recreated (destroy and create), as the platform has no operation to move a workflow between groups."
	WorkflowType                        = "What this workflow runs. <ul><li>`TERRAFORM` / `OPENTOFU` — Terraform or OpenTofu configuration.</li><li>`ANSIBLE_PLAYBOOK` — an Ansible playbook.</li><li>`HELM` — a Helm chart.</li><li>`KUBECTL` — Kubernetes manifests applied with kubectl.</li><li>`CLOUDFORMATION` — an AWS CloudFormation stack.</li><li>`CUSTOM` — anything else, typically a public repository run with your own steps.</li></ul>"
	WorkflowRunnerConstraints           = "Runner constraints to control which runner executes the workflow."
	WorkflowVcsConfig                   = "VCS configuration for the workflow."
	WorkflowIacVcsConfig                = "IaC VCS configuration for the workflow."
	WorkflowUseMarketplaceTemplate      = "Whether to use a marketplace template."
	WorkflowIacTemplateId               = "Workflow template revision this workflow is created from. <ul><li>`&lt;template-name&gt;:&lt;revision&gt;` — a template in your own organization.</li><li>`/&lt;org&gt;/&lt;template-name&gt;:&lt;revision&gt;` — a template owned by another organization: one shared with you, or published publicly. StackGuardian's own templates use the `stackguardian` org, for example `/stackguardian/terraform:11`.</li></ul>A bare id is resolved against your own organization. Use `:latest` in place of a revision number to track the most recently published revision; pin an explicit revision when the workflow must not move."
	WorkflowCustomSource                = "Custom VCS source configuration."
	WorkflowIacInputData                = "IaC input data for the workflow."
	WorkflowIacInputDataSchemaId        = "Schema ID for the input data."
	WorkflowIacInputDataSchemaType      = "How the value in `data` is formatted. <ul><li>`FORM_JSONSCHEMA` — a StackGuardian NoCode form; `data` holds the values that form collects.</li><li>`RAW_HCL` — HCL-formatted input, as you would write in a `.tfvars` file.</li><li>`RAW_JSON` — the same input expressed as JSON.</li><li>`NONE` — the workflow takes no inputs.</li></ul>"
	WorkflowIacInputDataData            = "Input data as a JSON string."
)
