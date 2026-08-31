package constants

// common template attributes
const (
	TemplateRevisionAlias                    string = "Alias for the template revision"
	TemplateRevisionNotes                    string = "Notes for the revision"
	TemplateRevisionIsPublic                 string = `Whether this **revision** is published and available to be referenced. Distinct from ` + "`is_public`" + ` on the parent template, which controls cross-organization sharing. Options: <span style="background-color: #eff0f0; color: #e53835;">"1"</span>, <span style="background-color: #eff0f0; color: #e53835;">"0"</span>`
	TemplateRevisionDeprecation              string = "Marking a template revision for deprecation"
	TemplateRevisionDeprecationEffectiveDate string = "Effective date for after which revision will be deprecated"
	TemplateRevisionDeprecationMessage       string = "Message shown to users who reference this revision after it has been deprecated. Use it to point them at the replacement revision."
	DeprecationMessage                       string = "Message shown to users who reference this revision after it has been deprecated. Use it to point them at the replacement revision."
)

// Common attributes shared between workflow template and revision
const (
	SourceConfigKind string = "What this template deploys, which decides how StackGuardian runs it. <ul><li>`TERRAFORM` / `OPENTOFU` — Terraform or OpenTofu configuration.</li><li>`ANSIBLE_PLAYBOOK` — an Ansible playbook.</li><li>`HELM` — a Helm chart.</li><li>`KUBECTL` — Kubernetes manifests applied with kubectl.</li><li>`CLOUDFORMATION` — an AWS CloudFormation stack.</li><li>`CUSTOM` — anything else, typically a public repository run with your own steps.</li></ul>"
	ContextTags      string = "Context tags for %s"
)

// Workflow Template attributes
const (
	WorkflowTemplateName       = "Name of the workflow template."
	WorkflowTemplateOwnerOrg   = "Organization the template belongs to"
	WorkflowTemplateIsPublic   = `Whether this **template** is shared with other organizations. Distinct from ` + "`is_public`" + ` on a revision, which controls whether that revision is published for use. Available values: <span style="background-color: #eff0f0; color: #e53835;">"0"</span> or <span style="background-color: #eff0f0; color: #e53835;">"1"</span>`
	WorkflowTemplateSharedOrgs = "List of organizations the template is shared with."
)

// Workflow Template Revision attributes
const (
	WorkflowTemplateRevisionTemplateId   = "Parent workflow template, as its bare `template_name` (e.g. `my-terraform-template`) — not a path. Reference the `stackguardian_workflow_template` resource rather than typing it."
	WorkflowTemplateRevisionInputSchemas = "JSONSchema Form representation of input JSON data"
)

// Runtime Source attributes (shared)
const (
	RuntimeSource                       = "Runtime source configuration for the %s."
	RuntimeSourceDestKind               = "Which VCS provider hosts the repository. This decides how StackGuardian authenticates and, for `vcs_triggers`, which webhook integration is used. <ul><li>`GITHUB_COM` — github.com. See the [GitHub connector docs](https://docs.stackguardian.io/docs/connectors/vcs/githubcom/).</li><li>`GITHUB_APP_CUSTOM` — GitHub Enterprise, or a GitHub App you manage yourself. See the [GitHub Enterprise docs](https://docs.stackguardian.io/docs/connectors/vcs/github_enterprise/).</li><li>`GITLAB_COM` — gitlab.com. See the [GitLab connector docs](https://docs.stackguardian.io/docs/connectors/vcs/gitlabcom/).</li><li>`BITBUCKET_ORG` — Bitbucket Cloud. See the [Bitbucket connector docs](https://docs.stackguardian.io/docs/connectors/vcs/bitbucket/).</li><li>`AZURE_DEVOPS` — Azure DevOps. See the [Azure DevOps connector docs](https://docs.stackguardian.io/docs/connectors/vcs/azuredevops/).</li><li>`AZURE_DEVOPS_SP` — Azure DevOps authenticated with a service principal.</li><li>`GIT_OTHER` — any other Git host, including public repositories that need no authentication.</li></ul>"
	RuntimeSourceConfig                 = "Configuration for the runtime environment."
	RuntimeSourceConfigAuth             = "Connector used to clone a private repository, as a path-form ID: `/integrations/<connector-name>` (e.g. `/integrations/github-connector`). Required when `is_private` is `true`."
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
	VCSTriggers     = "Webhook triggers for this workflow. Supported when the repository is on `GITHUB_COM`, `GITHUB_APP_CUSTOM` or `GITLAB_COM`.<br><br>**Requires** `vcs_config.iac_vcs_config.custom_source.config.is_private` to be `true` and `...config.auth` to name a `stackguardian_connector` with access — StackGuardian has to authenticate to register the webhook. See the [VCS connector docs](https://docs.stackguardian.io/docs/connectors/vcs/)." + "`source_config_dest_kind`" + ` is <span style="background-color: #eff0f0; color: #e53835;">GITHUB_COM</span>, <span style="background-color: #eff0f0; color: #e53835;">GITHUB_APP_CUSTOM</span>, or <span style="background-color: #eff0f0; color: #e53835;">GITLAB_COM</span>. **Requires** ` + "`vcs_config.iac_vcs_config.custom_source.config.is_private`" + ` to be ` + "`true`" + ` and ` + "`vcs_config.iac_vcs_config.custom_source.config.auth`" + ` to be set with a valid connector ID.`
	VCSTriggersType = "Which VCS platform the webhook is registered with. <ul><li>`GITHUB_COM` — github.com.</li><li>`GITHUB_APP_CUSTOM` — GitHub Enterprise, or a GitHub App you manage yourself.</li><li>`GITLAB_COM` — gitlab.com.</li></ul>Must match the provider hosting the repository in `runtime_source`."
	// TemplateVCSTriggers documents the `vcs_triggers` block on `stackguardian_workflow_template`.
	// It is deliberately separate from VCSTriggers above: a workflow template has no `vcs_config`
	// attribute (it uses `runtime_source`), and its only trigger creates a new *template revision*
	// on tag push rather than starting a workflow run.
	TemplateVCSTriggers                 = `VCS trigger configuration for the template. On a tag push, StackGuardian creates a new ` + "`stackguardian_workflow_template_revision`" + ` from the tagged commit — it does not start a workflow run. The repository is taken from ` + "`runtime_source`" + `.`
	VCSTriggersCreateTag                = "Trigger configuration on tag creation in VCS"
	VCSTriggersCreateTagRevision        = "Create new revision on tag creation"
	VCSTriggersCreateTagRevisionEnabled = "Whether to create revision when tag is created."

	VCSTriggersTrackedBranch       = "The branch that push and pull request events must target to trigger a workflow run. For push events, the pushed-to branch must equal this value. For pull request events, the PR's base (target) branch must equal this value — unless `all_pull_requests.createWfRun.enabled` is `true`, which bypasses this check entirely. If omitted, falls back to the branch set in the workflow's VCS config, then to the repository's default branch."
	VCSTriggersApprovalPreApply    = "When `true`, workflow runs triggered by push or tag events run `apply` but require manual approval before the apply executes. Has no effect on pull request events — those always run `plan` regardless. Ignored when `plan_only` is `true`; `plan_only` takes precedence."
	VCSTriggersPlanOnly            = "When `true`, all workflow runs triggered by push or tag events execute `plan` instead of `apply`. Takes precedence over `approval_pre_apply` — setting both to `true` results in `plan` only, with no apply or approval step. Has no effect on pull request events — those always run `plan` regardless."
	VCSTriggersFileTriggersEnabled = "When `true`, activates file-based filtering using the patterns in `file_trigger_patterns`. A webhook event only triggers a workflow run if at least one changed file matches a pattern. Must be `true` for `file_trigger_patterns` to have any effect; setting patterns without enabling this flag is a no-op. **Only valid when `source_config_dest_kind` is `GITLAB_COM`.**"
	VCSTriggersFileTriggerPatterns = "List of [fnmatch](https://docs.python.org/3/library/fnmatch.html) glob patterns matched against the files changed in the event (e.g. `[\"*.tf\", \"infra/**/*.json\"]`). A workflow run is triggered only if at least one changed file matches at least one pattern. Only evaluated when `file_triggers_enabled` is `true`; has no effect otherwise."
	VCSTriggersGlHookId            = "The GitLab webhook ID created by StackGuardian when the VCS trigger is registered. Populated automatically on first apply. Read-only."
	VCSTriggersBbHookId            = "The Bitbucket webhook ID created by StackGuardian when the VCS trigger is registered. Populated automatically on first apply. Read-only."
	VCSTriggersGhWebhookUrl        = "The StackGuardian webhook URL registered to receive GitHub events for this workflow. Populated automatically on first apply. Read-only."
	VCSTriggersAdoHooksId          = "Map of Azure DevOps service hook subscription IDs created by StackGuardian, keyed by event type (e.g. `git.push`, `git.pullrequest.created`). Populated automatically on first apply. Read-only."
	VCSTriggersAllPullRequests     = "Actions to trigger on StackGuardian for all pull request events, regardless of target branch. Supported action key: `createWfRun`. When `createWfRun.enabled` is `true`, this overrides `pull_request_opened`, `pull_request_modified`, and `tracked_branch` — any PR event fires a workflow run without branch filtering. When absent or disabled, `pull_request_opened` and `pull_request_modified` are evaluated individually, each subject to `tracked_branch`."
	VCSTriggersPullRequestOpened   = "Actions to trigger on StackGuardian when a pull request is opened. Supported action key: `createWfRun`. Only evaluated when `all_pull_requests.createWfRun.enabled` is `false` or absent. When `createWfRun.enabled` is `true`, a workflow run is created if the PR's target branch equals `tracked_branch`. The triggered run always executes `plan`, regardless of `plan_only` or `approval_pre_apply`."
	VCSTriggersPullRequestModified = "Actions to trigger on StackGuardian when new commits are pushed to an open pull request. Supported action key: `createWfRun`. Only evaluated when `all_pull_requests.createWfRun.enabled` is `false` or absent. When `createWfRun.enabled` is `true`, a workflow run is created if the PR's target branch equals `tracked_branch`. The triggered run always executes `plan`, regardless of `plan_only` or `approval_pre_apply`."
	VCSTriggersCreateTagAction     = "Actions to trigger on StackGuardian when a git tag is created. Supported action key: `createWfRun`. When `createWfRun.enabled` is `true`, a workflow run is created with the tag set as the VCS ref. The Terraform action follows `plan_only` / `approval_pre_apply` — unlike pull request events, tag events are not hardcoded to `plan`."
	VCSTriggersActionEnabled       = "Whether this trigger action is active. When `false` the event is received but no workflow run is created."
	VCSTriggersPush                = "Actions to trigger on StackGuardian on a push event. Supported action key: `createWfRun`. When `createWfRun.enabled` is `true`, a workflow run is created only when the pushed branch equals `tracked_branch`. The Terraform action is `plan` if `plan_only` is `true`, `apply` with a manual approval gate if `approval_pre_apply` is `true`, or `apply` by default."
)

// Environment Variables attributes
const (
	EnvVarConfig          = "Configuration for the environment variable."
	EnvVarConfigVarName   = "Name of the variable."
	EnvVarConfigSecretId  = "Secret to read the value from, as a path-form ID: `/secrets/<secret-name>` (e.g. `/secrets/db-password`). Only used when `kind` is `VAULT_SECRET`."
	EnvVarConfigTextValue = `Text value (if using plain text). Only if type is <span style="background-color: #eff0f0; color: #e53835;">TEXT</span>`
	EnvVarKind            = "Where the variable's value comes from. <ul><li>`PLAIN_TEXT` — the value is written inline in `config.text_value`. It is visible in configuration and state, so do not use it for credentials.</li><li>`VAULT_SECRET` — the value is read at run time from the secret named by `config.secret_id`, so it never appears in your configuration or state.</li></ul>"
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
	MiniStepsWfChainingWorkflowGroupId = "Workflow group containing the workflow to trigger, as a bare name. For a nested group give the full path (e.g. `platform/networking`)."
	MiniStepsWfChainingStackId         = "Stack to trigger, as a bare name within `workflow_group_id`."
	MiniStepsWfChainingStackPayload    = "JSON string specifying overrides for the stack to be triggered"
	MiniStepsWfChainingWorkflowId      = "Workflow to trigger, as a bare name within `workflow_group_id`."
	MiniStepsWfChainingWorkflowPayload = "JSON string specifying overrides for the workflow to be triggered"
)

// Runner Constraints attributes
const (
	RunnerConstraintsType  = `Type of runner. Valid options: <span style="background-color: #eff0f0; color: #e53835;">shared</span> or <span style="background-color: #eff0f0; color: #e53835;">private</span>`
	RunnerConstraintsNames = "Runner groups the workflow is pinned to, given as a list of `stackguardian_runner_group` `resource_name` values. Set this when `type` is `private`; with `type = \"shared\"` the workflow uses StackGuardian's shared runners and this field is not applicable."
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
	DeploymentPlatformKind          = "Which cloud this workflow deploys to and how it authenticates. Must match the `kind` of the connector named in `integration_id`. <ul><li>`AWS_STATIC` — AWS with a static access key pair.</li><li>`AWS_RBAC` — AWS by assuming an IAM role. Preferred over static keys.</li><li>`AWS_OIDC` — AWS through an OIDC identity provider, with no stored credentials.</li><li>`AZURE_STATIC` — Azure with a service principal and client secret.</li><li>`AZURE_OIDC` — Azure through workload identity federation.</li><li>`GCP_STATIC` — GCP with a service account key file.</li><li>`GCP_OIDC` — GCP through workload identity federation.</li></ul>See the [cloud connector docs](https://docs.stackguardian.io/docs/connectors/csp/)."
	DeploymentPlatformConfigDetails = "Deployment platform configuration details."
	DeploymentPlatformIntegrationId = "Connector supplying the credentials this workflow deploys with, as a path-form ID: `/integrations/<connector-name>` (e.g. `/integrations/production-aws`). Reference a `stackguardian_connector` rather than typing the string."
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
	WfStepName                        = "Step name."
	WfStepEnvVars                     = "Environment variables for the workflow steps."
	WfStepApproval                    = "Enable approval for the workflow step."
	WfStepTimeout                     = "Workflow step execution timeout in seconds."
	WfStepCmdOverride                 = "Override command for the step."
	WfStepMountPoints                 = "Mount points for the step."
	WfStepTemplateId                  = "Workflow step template revision, as a path-form ID: `/<org>/<step-template-name>:<revision>` (e.g. `/my-org/ansible:6`). Steps published by StackGuardian live under the `stackguardian` org — for example `/stackguardian/terraform:11`."
	TerraformWfStepTemplateRevisionId = "Fully-qualified workflow step template revision id pinned for this terraform config (e.g. \"/<org>/<name>:<rev>\")."
	WfStepInputData                   = "Workflow step input data (JSON string)"
	WfStepInputDataSchemaType         = "How the value in `data` is formatted. `FORM_JSONSCHEMA` is a StackGuardian NoCode form; `data` holds the values that form collects."
	WfStepInputDataData               = "Input data (JSON)."
)

// Terraform Config attributes
const (
	TerraformConfig                  = "Terraform configuration. Valid only for terraform type template"
	TerraformVersion                 = "Terraform or OpenTofu version, in bare form (e.g. `1.5.7`). StackGuardian stores an engine prefix (`TERRAFORM-` / `OPENTOFU-`) internally; the provider strips it, so the value can be referenced directly into another resource without producing a perpetual diff. Which engine runs is decided by `wf_type`, not by this value."
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
