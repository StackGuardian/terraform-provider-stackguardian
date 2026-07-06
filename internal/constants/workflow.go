package constants

// Workflow resource attributes
const (
	WorkflowWorkflowGroupId          = "ID of the parent workflow group."
	// WorkflowFromTemplateWorkflowGroupId is workflow_group_id for the workflow_from_template
	// resource, where it is immutable (changing it forces recreation).
	WorkflowFromTemplateWorkflowGroupId = "ID of the parent workflow group. Immutable — changing this forces the workflow to be recreated (destroy and create), as the platform has no operation to move a workflow between groups."
	WorkflowType                     = `Type of workflow. Options: <span style="background-color: #eff0f0; color: #e53835;">TERRAFORM</span>, <span style="background-color: #eff0f0; color: #e53835;">OPENTOFU</span>, <span style="background-color: #eff0f0; color: #e53835;">ANSIBLE_PLAYBOOK</span>, <span style="background-color: #eff0f0; color: #e53835;">HELM</span>, <span style="background-color: #eff0f0; color: #e53835;">KUBECTL</span>, <span style="background-color: #eff0f0; color: #e53835;">CLOUDFORMATION</span>, <span style="background-color: #eff0f0; color: #e53835;">CUSTOM</span>`
	WorkflowRunnerConstraints = "Runner constraints to control which runner executes the workflow."
	WorkflowVcsConfig         = "VCS configuration for the workflow."
	WorkflowIacVcsConfig             = "IaC VCS configuration for the workflow."
	WorkflowUseMarketplaceTemplate   = "Whether to use a marketplace template."
	WorkflowIacTemplateId            = "ID of the IaC template from the marketplace."
	WorkflowCustomSource             = "Custom VCS source configuration."
	WorkflowIacInputData             = "IaC input data for the workflow."
	WorkflowIacInputDataSchemaId     = "Schema ID for the input data."
	WorkflowIacInputDataSchemaType   = "Schema type for the input data. Allowed values are `FORM_JSONSCHEMA`, `RAW_HCL`, `RAW_JSON`, `NO_CODE_JSON`, `NONE`."
	WorkflowIacInputDataData         = "Input data as a JSON string."
)
