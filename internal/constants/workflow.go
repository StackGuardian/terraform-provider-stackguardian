package constants

// Workflow resource attributes
const (
	WorkflowWorkflowGroupId          = "ID of the parent workflow group."
	WorkflowType                     = `Type of workflow. Options: <span style="background-color: #eff0f0; color: #e53835;">TERRAFORM</span>, <span style="background-color: #eff0f0; color: #e53835;">OPENTOFU</span>, <span style="background-color: #eff0f0; color: #e53835;">ANSIBLE_PLAYBOOK</span>, <span style="background-color: #eff0f0; color: #e53835;">HELM</span>, <span style="background-color: #eff0f0; color: #e53835;">KUBECTL</span>, <span style="background-color: #eff0f0; color: #e53835;">CLOUDFORMATION</span>, <span style="background-color: #eff0f0; color: #e53835;">CUSTOM</span>`
	WorkflowUpgradeMode              = "Controls how the workflow is updated when the underlying template changes. Accepted values: `RESET_TO_TEMPLATE`, `PRESERVE_SETTINGS`."
	WorkflowRunnerConstraints = "Runner constraints to control which runner executes the workflow."
	WorkflowVcsConfig         = "VCS configuration for the workflow."
	WorkflowIacVcsConfig             = "IaC VCS configuration for the workflow."
	WorkflowUseMarketplaceTemplate   = "Whether to use a marketplace template."
	WorkflowIacTemplateId            = "ID of the IaC template from the marketplace."
	WorkflowCustomSource             = "Custom VCS source configuration."
	WorkflowIacInputData             = "IaC input data for the workflow."
	WorkflowIacInputDataSchemaId     = "Schema ID for the input data."
)
