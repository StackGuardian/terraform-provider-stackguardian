package constants

// common template attributes
const (
	TemplateRevisionAlias                    string = "Alias for the template revision"
	TemplateRevisionNotes                    string = "Notes for the revision"
	TemplateRevisionIsPublic                 string = `Whether a revision is published to be used. Options: <span style="background-color: #eff0f0; color: #e53835;">"1"</span>, <span style="background-color: #eff0f0; color: #e53835;">"0"</span>`
	TemplateRevisionDeprecation              string = "Marking a template revision for deprecation"
	TemplateRevisionDeprecationEffectiveDate string = "Effective date for after which revision will be deprecated"
	TemplateRevisionDeprecationMessage       string = "Deprecation message"
	DeprecationMessage                       string = "Deprecation message"
)

// Common attributes shared between workflow template and revision
const (
	SourceConfigKind string = `Source configuration kind. Options: <span style="background-color: #eff0f0; color: #e53835;">TERRAFORM</span>, <span style="background-color: #eff0f0; color: #e53835;">OPENTOFU</span>, <span style="background-color: #eff0f0; color: #e53835;">ANSIBLE_PLAYBOOK</span>, <span style="background-color: #eff0f0; color: #e53835;">HELM</span>, <span style="background-color: #eff0f0; color: #e53835;">KUBECTL</span>, <span style="background-color: #eff0f0; color: #e53835;">CLOUDFORMATION</span>, <span style="background-color: #eff0f0; color: #e53835;">CUSTOM</span>`
	ContextTags      string = "Context tags for %s"
)

// Workflow Template attributes
const (
	WorkflowTemplateName       = "Name of the workflow template."
	WorkflowTemplateOwnerOrg   = "Organization the template belongs to"
	WorkflowTemplateIsPublic   = `Make template available to other organisations. Available values: <span style="background-color: #eff0f0; color: #e53835;">"0"</span> or <span style="background-color: #eff0f0; color: #e53835;">"1"</span>`
	WorkflowTemplateSharedOrgs = "List of organizations the template is shared with."
)

// Workflow Template Revision attributes
const (
	WorkflowTemplateRevisionTemplateId               = "Resource ID of the parent workflow template."
	WorkflowTemplateRevisionApprovers                = "List of approvers for approvals during workflow execution."
	WorkflowTemplateRevisionNumberOfApprovals        = "Number of approvals required."
	WorkflowTemplateRevisionUserJobCPU               = "Limits to set user job CPU."
	WorkflowTemplateRevisionUserJobMemory            = "Limits to set user job memory."
	WorkflowTemplateRevisionEnvironmentVariables     = "List of environment variables for the revision."
	WorkflowTemplateRevisionInputSchemas             = "JSONSchema Form representation of input JSON data"
	WorkflowTemplateRevisionMiniSteps                = "Actions that are required to be performed once workflow execution is complete"
	WorkflowTemplateRevisionUserSchedules            = "Configuration for scheduling runs for the workflows."
	WorkflowTemplateRevisionDeploymentPlatformConfig = "Deployment platform configuration for the revision."
	WorkflowTemplateRevisionWfStepsConfig            = "Workflow steps configuration. Valid only for source_config_kind *CUSTOM*."
)

// Runtime Source attributes (shared)
const (
	RuntimeSource                       = "Runtime source configuration for the %s."
	RuntimeSourceDestKind               = `Destination kind for the source configuration. Options: <span style="background-color: #eff0f0; color: #e53835;">GITHUB_COM</span>, <span style="background-color: #eff0f0; color: #e53835;">GITHUB_APP_CUSTOM</span>, <span style="background-color: #eff0f0; color: #e53835;">GITLAB_OAUTH_SSH</span>, <span style="background-color: #eff0f0; color: #e53835;">GITLAB_COM</span>, <span style="background-color: #eff0f0; color: #e53835;">AZURE_DEVOPS</span>`
	RuntimeSourceConfig                 = "Configuration for the runtime environment."
	RuntimeSourceConfigAuth             = "Connector id to access private git repository"
	RuntimeSourceConfigGitCoreCRLF      = "Whether to automatically handle CRLF line endings."
	RuntimeSourceConfigGitSparse        = "Git sparse checkout command line git cli options."
	RuntimeSourceConfigIncludeSubmodule = "Whether to include git submodules."
	RuntimeSourceConfigIsPrivate        = "Whether the repository is private. Auth is required if the repository is private"
	RuntimeSourceConfigRef              = "Git reference (branch, tag, or commit hash)."
	RuntimeSourceConfigRepo             = "Git repository URL."
	RuntimeSourceConfigWorkingDir       = "Working directory within the repository."
)

// VCS Triggers attributes
const (
	VCSTriggers                         = "VCS trigger configuration for the workflow."
	VCSTriggersType                     = `VCS provider type. Options: <span style="background-color: #eff0f0; color: #e53835;">GITHUB_COM</span>, <span style="background-color: #eff0f0; color: #e53835;">GITHUB_APP_CUSTOM</span>, <span style="background-color: #eff0f0; color: #e53835;">GITLAB_OAUTH_SSH</span>, <span style="background-color: #eff0f0; color: #e53835;">GITLAB_COM</span>`
	VCSTriggersCreateTag                = "Trigger configuration on tag creation in VCS"
	VCSTriggersCreateTagRevision        = "Create new revision on tag creation"
	VCSTriggersCreateTagRevisionEnabled = "Whether to create revision when tag is created."
)

// Environment Variables attributes
const (
	EnvVarConfig          = "Configuration for the environment variable."
	EnvVarConfigVarName   = "Name of the variable."
	EnvVarConfigSecretId  = `ID of the secret (if using vault secret). Only if type is <span style="background-color: #eff0f0; color: #e53835;">SECRET_REF</span>`
	EnvVarConfigTextValue = `Text value (if using plain text). Only if type is <span style="background-color: #eff0f0; color: #e53835;">TEXT</span>`
	EnvVarKind            = `Kind of the environment variable. Options: <span style="background-color: #eff0f0; color: #e53835;">TEXT</span>, <span style="background-color: #eff0f0; color: #e53835;">SECRET_REF</span>`
)

// Input Schemas attributes
const (
	InputSchemaType         = "Type of the schema."
	InputSchemaEncodedData  = "JSON schema for the Form in templates. The schema needs to be base64 encoded."
	InputSchemaUISchemaData = "Schema for how the JSON schema is to be visualized. The schema needs to be base64 encoded."
)

// Mini Steps attributes
const (
	MiniStepsNotifications             = "Configuration for notifications to be sent on workflow completion"
	MiniStepsNotificationsEmail        = `Configuration for email notifications to be sent on completion. Statuses on which notifications can be sent: <span style="background-color: #eff0f0; color: #e53835;">approval_required</span>, <span style="background-color: #eff0f0; color: #e53835;">cancelled</span>, <span style="background-color: #eff0f0; color: #e53835;">completed</span>, <span style="background-color: #eff0f0; color: #e53835;">drift_detected</span>, <span style="background-color: #eff0f0; color: #e53835;">errored</span>`
	MiniStepsNotificationsRecipients   = "List of emails"
	MiniStepsWebhooks                  = `Configuration for webhooks to be triggered on completion. Statuses on which webhooks can be sent: <span style="background-color: #eff0f0; color: #e53835;">approval_required</span>, <span style="background-color: #eff0f0; color: #e53835;">cancelled</span>, <span style="background-color: #eff0f0; color: #e53835;">completed</span>, <span style="background-color: #eff0f0; color: #e53835;">drift_detected</span>, <span style="background-color: #eff0f0; color: #e53835;">errored</span>`
	MiniStepsWebhookName               = "Webhook name"
	MiniStepsWebhookURL                = "Webhook URL"
	MiniStepsWebhookSecret             = "Secret to be sent with API request to webhook url"
	MiniStepsWorkflowChaining          = `Configuration for other workflows to be triggered on completion. Statuses on which workflows can be chained: <span style="background-color: #eff0f0; color: #e53835;">completed</span>, <span style="background-color: #eff0f0; color: #e53835;">errored</span>`
	MiniStepsWfChainingWorkflowGroupId = "Workflow group id for the workflow."
	MiniStepsWfChainingStackId         = "Stack id for the stack to be triggered."
	MiniStepsWfChainingStackPayload    = "JSON string specifying overrides for the stack to be triggered"
	MiniStepsWfChainingWorkflowId      = "Workflow id for the workflow to be triggered"
	MiniStepsWfChainingWorkflowPayload = "JSON string specifying overrides for the workflow to be triggered"
)

// Runner Constraints attributes
const (
	RunnerConstraintsType  = `Type of runner. Valid options: <span style="background-color: #eff0f0; color: #e53835;">shared</span> or <span style="background-color: #eff0f0; color: #e53835;">external</span>`
	RunnerConstraintsNames = "Id of the runner group"
)

// User Schedules attributes
const (
	UserScheduleCron  = "Cron expression defining the schedule."
	UserScheduleState = `State of the schedule. Options: <span style="background-color: #eff0f0; color: #e53835;">ENABLED</span>, <span style="background-color: #eff0f0; color: #e53835;">DISABLED</span>`
	UserScheduleDesc  = "Description of the schedule."
	UserScheduleName  = "Name of the schedule."
)

// Deployment Platform Config attributes
const (
	DeploymentPlatformKind          = `Deployment platform kind. Options: <span style="background-color: #eff0f0; color: #e53835;">AWS_STATIC</span>, <span style="background-color: #eff0f0; color: #e53835;">AWS_RBAC</span>, <span style="background-color: #eff0f0; color: #e53835;">AWS_OIDC</span>, <span style="background-color: #eff0f0; color: #e53835;">AZURE_STATIC</span>, <span style="background-color: #eff0f0; color: #e53835;">AZURE_OIDC</span>, <span style="background-color: #eff0f0; color: #e53835;">GCP_STATIC</span>, <span style="background-color: #eff0f0; color: #e53835;">GCP_OIDC</span>`
	DeploymentPlatformConfigDetails = "Deployment platform configuration details."
	DeploymentPlatformIntegrationId = "Integration ID for the deployment platform."
	DeploymentPlatformProfileName   = "Profile name for the deployment platform."
)

// Mount Point attributes
const (
	MountPointSource   = "Source path for mount point."
	MountPointTarget   = "Target path for mount point."
	MountPointReadOnly = "If the directory is to be mounted as read only or not"
)

// Workflow Steps Config attributes
const (
	WfStepName                = "Step name."
	WfStepEnvVars             = "Environment variables for the workflow steps."
	WfStepApproval            = "Enable approval for the workflow step."
	WfStepTimeout             = "Workflow step execution timeout in seconds."
	WfStepCmdOverride         = "Override command for the step."
	WfStepMountPoints         = "Mount points for the step."
	WfStepTemplateId          = "Workflow step template ID."
	WfStepInputData           = "Workflow step input data (JSON string)"
	WfStepInputDataSchemaType = `Schema type for the input data. Options: <span style="background-color: #eff0f0; color: #e53835;">FORM_JSONSCHEMA</span>`
	WfStepInputDataData       = "Input data (JSON)."
)

// Terraform Config attributes
const (
	TerraformConfig                 = "Terraform configuration. Valid only for terraform type template"
	TerraformVersion                = "Terraform version to use."
	TerraformDriftCheck             = "Enable drift check."
	TerraformDriftCron              = "Cron expression for drift check."
	TerraformManagedState           = "Enable stackguardian managed terraform state."
	TerraformApprovalPreApply       = "Require approval before apply."
	TerraformPlanOptions            = "Additional options for terraform plan."
	TerraformInitOptions            = "Additional options for terraform init."
	TerraformBinPath                = "Mount points for terraform binary."
	TerraformTimeout                = "Timeout for terraform operations in seconds."
	TerraformPostApplyWfSteps       = "Workflow steps configuration to run after apply."
	TerraformPreApplyWfSteps        = "Workflow steps configuration to run before apply."
	TerraformPrePlanWfSteps         = "Workflow steps configuration to run before plan."
	TerraformPostPlanWfSteps        = "Workflow steps configuration to run after plan."
	TerraformPreInitHooks           = "Hooks to run before init."
	TerraformPrePlanHooks           = "Hooks to run before plan."
	TerraformPostPlanHooks          = "Hooks to run after plan."
	TerraformPreApplyHooks          = "Hooks to run before apply."
	TerraformPostApplyHooks         = "Hooks to run after apply."
	TerraformRunPreInitHooksOnDrift = "Run pre-init hooks on drift detection."
)
