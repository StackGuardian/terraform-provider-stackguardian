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
	WorkflowTemplateRevisionTemplateId   = "Resource ID of the parent workflow template."
	WorkflowTemplateRevisionInputSchemas = "JSONSchema Form representation of input JSON data"
)

// Runtime Source attributes (shared)
const (
	RuntimeSource                       = "Runtime source configuration for the %s."
	RuntimeSourceDestKind               = `VCS provider kind. Options: <span style="background-color: #eff0f0; color: #e53835;">GITHUB_COM</span>, <span style="background-color: #eff0f0; color: #e53835;">GITHUB_APP_CUSTOM</span>, <span style="background-color: #eff0f0; color: #e53835;">GIT_OTHER</span>, <span style="background-color: #eff0f0; color: #e53835;">BITBUCKET_ORG</span>, <span style="background-color: #eff0f0; color: #e53835;">GITLAB_COM</span>, <span style="background-color: #eff0f0; color: #e53835;">AZURE_DEVOPS</span>, <span style="background-color: #eff0f0; color: #e53835;">AZURE_DEVOPS_SP</span>`
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
	VCSTriggersType                     = `The VCS platform type. Determines which webhook integration is used. Supported values: <span style="background-color: #eff0f0; color: #e53835;">GITHUB_COM</span>, <span style="background-color: #eff0f0; color: #e53835;">GITHUB_APP_CUSTOM</span>, <span style="background-color: #eff0f0; color: #e53835;">GITLAB_COM</span>,`
	VCSTriggersCreateTag                = "Trigger configuration on tag creation in VCS"
	VCSTriggersCreateTagRevision        = "Create new revision on tag creation"
	VCSTriggersCreateTagRevisionEnabled = "Whether to create revision when tag is created."

	VCSTriggersTrackedBranch       = "The branch that push and pull request events must target to trigger a workflow run. For push events, the pushed-to branch must equal this value. For pull request events, the PR's base (target) branch must equal this value — unless `all_pull_requests.createWfRun.enabled` is `true`, which bypasses this check entirely. If omitted, falls back to the branch set in the workflow's VCS config, then to the repository's default branch."
	VCSTriggersApprovalPreApply    = "When `true`, workflow runs triggered by push or tag events run `apply` but require manual approval before the apply executes. Has no effect on pull request events — those always run `plan` regardless. Ignored when `plan_only` is `true`; `plan_only` takes precedence."
	VCSTriggersPlanOnly            = "When `true`, all workflow runs triggered by push or tag events execute `plan` instead of `apply`. Takes precedence over `approval_pre_apply` — setting both to `true` results in `plan` only, with no apply or approval step. Has no effect on pull request events — those always run `plan` regardless."
	VCSTriggersFileTriggersEnabled = "When `true`, activates file-based filtering using the patterns in `file_trigger_patterns`. A webhook event only triggers a workflow run if at least one changed file matches a pattern. Must be `true` for `file_trigger_patterns` to have any effect; setting patterns without enabling this flag is a no-op."
	VCSTriggersFileTriggerPatterns = "List of [fnmatch](https://docs.python.org/3/library/fnmatch.html) glob patterns matched against the files changed in the event (e.g. `[\"*.tf\", \"infra/**/*.json\"]`). A workflow run is triggered only if at least one changed file matches at least one pattern. Only evaluated when `file_triggers_enabled` is `true`; has no effect otherwise."
	VCSTriggersGlHookId            = "The GitLab webhook ID created by StackGuardian when the VCS trigger is registered. Populated automatically on first apply. Read-only."
	VCSTriggersBbHookId            = "The Bitbucket webhook ID created by StackGuardian when the VCS trigger is registered. Populated automatically on first apply. Read-only."
	VCSTriggersGhWebhookUrl        = "The StackGuardian webhook URL registered to receive GitHub events for this workflow. Populated automatically on first apply. Read-only."
	VCSTriggersAdoHooksId          = "Map of Azure DevOps service hook subscription IDs created by StackGuardian, keyed by event type (e.g. `git.push`, `git.pullrequest.created`). Populated automatically on first apply. Read-only."
	VCSTriggersAllPullRequests     = "Actions to trigger on StackGuardian for all pull request events, regardless of target branch. Supported action key: `createWfRun`. When `createWfRun.enabled` is `true`, this overrides `pull_request_opened`, `pull_request_modified`, and `tracked_branch` — any PR event fires a workflow run without branch filtering. When absent or disabled, `pull_request_opened` and `pull_request_modified` are evaluated individually, each subject to `tracked_branch`."
	VCSTriggersPullRequestOpened   = "Actions to trigger on StackGuardian when a pull request is opened. Supported action key: `createWfRun`. Only evaluated when `all_pull_requests.createWfRun.enabled` is `false` or absent. When `createWfRun.enabled` is `true`, a workflow run is created if the PR's target branch equals `tracked_branch`. The triggered run always executes `plan`, regardless of `plan_only` or `approval_pre_apply`."
	VCSTriggersPullRequestModified = "Actions to trigger on StackGuardian when new commits are pushed to an open pull request. Supported action key: `createWfRun`. Only evaluated when `all_pull_requests.createWfRun.enabled` is `false` or absent. When `createWfRun.enabled` is `true`, a workflow run is created if the PR's target branch equals `tracked_branch`. The triggered run always executes `plan`, regardless of `plan_only` or `approval_pre_apply`."
	VCSTriggersCreateTagAction     = "Actions to trigger on StackGuardian when a git tag is created. Supported action key: `createWfRun`. When `createWfRun.enabled` is `true`, a workflow run is created with the tag set as the VCS ref. The Terraform action follows `plan_only` / `approval_pre_apply` — unlike pull request events, tag events are not hardcoded to `plan`."
	VCSTriggersPush                = "Actions to trigger on StackGuardian on a push event. Supported action key: `createWfRun`. When `createWfRun.enabled` is `true`, a workflow run is created only when the pushed branch equals `tracked_branch`. The Terraform action is `plan` if `plan_only` is `true`, `apply` with a manual approval gate if `approval_pre_apply` is `true`, or `apply` by default."
)

// Environment Variables attributes
const (
	EnvVarConfig          = "Configuration for the environment variable."
	EnvVarConfigVarName   = "Name of the variable."
	EnvVarConfigSecretId  = `ID of the secret (if using vault secret). Only if type is <span style="background-color: #eff0f0; color: #e53835;">SECRET_REF</span>`
	EnvVarConfigTextValue = `Text value (if using plain text). Only if type is <span style="background-color: #eff0f0; color: #e53835;">TEXT</span>`
	EnvVarKind            = `Kind of the environment variable. Options: <span style="background-color: #eff0f0; color: #e53835;">PLAIN_TEXT</span>, <span style="background-color: #eff0f0; color: #e53835;">SECRET_VALUE</span>`
)

// Input Schemas attributes
const (
	InputSchemaName         = "Name of the input schema."
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
	RunnerConstraintsType  = `Type of runner. Valid options: <span style="background-color: #eff0f0; color: #e53835;">shared</span> or <span style="background-color: #eff0f0; color: #e53835;">private</span>`
	RunnerConstraintsNames = "Id of the runner group. Allowed only if type is external."
)

// User Schedules attributes
const (
	UserScheduleCron  = `Cron expression defining the schedule. Use [AWS cron](https://docs.aws.amazon.com/eventbridge/latest/userguide/eb-scheduled-rule-pattern.html) expression format.`
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
	TerraformConfig                  = "Terraform configuration. Valid only for terraform type template"
	TerraformVersion                 = "Terraform version to use."
	TerraformDriftCheck              = "Enable drift check."
	TerraformDriftCron               = "Cron expression for drift check."
	TerraformManagedState            = "Enable stackguardian managed terraform state."
	TerraformApprovalPreApply        = "Require approval before apply."
	TerraformPlanOptions             = "Additional options for terraform plan."
	TerraformInitOptions             = "Additional options for terraform init."
	TerraformBinPath                 = "Mount points for terraform binary."
	TerraformTimeout                 = "Timeout for terraform operations in seconds."
	TerraformPostApplyWfSteps        = "Workflow steps configuration to run after apply."
	TerraformPreApplyWfSteps         = "Workflow steps configuration to run before apply."
	TerraformPrePlanWfSteps          = "Workflow steps configuration to run before plan."
	TerraformPostPlanWfSteps         = "Workflow steps configuration to run after plan."
	TerraformPreInitHooks            = "Hooks to run before init."
	TerraformPrePlanHooks            = "Hooks to run before plan."
	TerraformPostPlanHooks           = "Hooks to run after plan."
	TerraformPreApplyHooks           = "Hooks to run before apply."
	TerraformPostApplyHooks          = "Hooks to run after apply."
	TerraformRunPreInitHooksOnDrift  = "Run pre-init hooks on drift detection."
	TerraformRunPrePlanHooksOnDrift  = "Run pre-plan hooks on drift detection."
	TerraformRunPostPlanHooksOnDrift = "Run post-plan hooks on drift detection."
)
