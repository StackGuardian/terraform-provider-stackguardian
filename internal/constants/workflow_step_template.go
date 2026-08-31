package constants

// Workflow Step Template - Common documentation
const (
	WorkflowStepTemplateSourceConfigKindCommon = "Where the step's runnable definition comes from. `DOCKER_IMAGE` — a container image, pulled from the registry described by `runtime_source`."

	WorkflowStepTemplateIsPublicCommon = `Whether the workflow step template is publicly available. Valid values:
	<span style="background-color: #eff0f0; color: #e53835;">0</span> (false),
	<span style="background-color: #eff0f0; color: #e53835;">1</span> (true)`

	WorkflowStepTemplateRuntimeSourceDestKindCommon = "Where the step image is hosted. `CONTAINER_REGISTRY` — a Docker-compatible registry; set `config.docker_image` to the image reference, plus `config.auth` and `config.is_private` when the registry needs credentials."

	WorkflowStepTemplateRuntimeSourceConfigIsPrivateCommon = "Indicates whether the container registry or repository is private."

	WorkflowStepTemplateRuntimeSourceConfigAuthCommon = "Authentication credentials or method for accessing the private registry or repository. (Sensitive)"

	WorkflowStepTemplateRuntimeSourceConfigDockerImageCommon = "Docker image URI to be used for template execution. Example: `ubuntu:latest`, `myregistry.azurecr.io/myapp:v1.0`"

	WorkflowStepTemplateRuntimeSourceConfigDockerRegistryUsernameCommon = "Username for authentication with the Docker registry (if using private registries)."

	WorkflowStepTemplateRuntimeSourceConfigLocalWorkspaceDirCommon = "Working directory path inside the workspace, relative to the repository root."
)

// Workflow Step Template Resource documentation
const (
	WorkflowStepTemplateName = "Name of the workflow step template. Must be less than 100 characters."

	WorkflowStepTemplateDescription = "A brief description of the workflow step template. Must be less than 256 characters."

	WorkflowStepTemplateType = "Which family of template this is. Read-only, and always `WORKFLOW_STEP` for this resource. StackGuardian uses the same field across all template types: <ul><li>`WORKFLOW_STEP` — a workflow step template, managed by `stackguardian_workflow_step_template`.</li><li>`IAC` — a workflow template, managed by `stackguardian_workflow_template`.</li><li>`IAC_GROUP` — a stack template, managed by `stackguardian_stack_template`.</li><li>`IAC_POLICY` — a policy template, referenced from `policy_vcs_config.policy_template_id`.</li></ul>"

	WorkflowStepTemplateIsPublic = `Whether the workflow step template is publicly available. Valid values:
	<span style="background-color: #eff0f0; color: #e53835;">0</span> (false),
	<span style="background-color: #eff0f0; color: #e53835;">1</span> (true)`

	WorkflowStepTemplateTags = "A list of tags associated with the workflow step template. A maximum of 10 tags are allowed."

	WorkflowStepTemplateContextTags = "Contextual key-value tags that provide additional context to the main tags."

	WorkflowStepTemplateSharedOrgsList = "List of organization IDs with which this template is shared."

	WorkflowStepTemplateSourceConfigKind = "Where the step's runnable definition comes from. `DOCKER_IMAGE` — a container image, pulled from the registry described by `runtime_source`."

	WorkflowStepTemplateLatestRevision = "Latest revision number of the template."

	WorkflowStepTemplateNextRevision = "Next revision number that will be used for the template."

	WorkflowStepTemplateRuntimeSource = "Runtime source configuration that defines where and how the template code is stored and executed."

	WorkflowStepTemplateRuntimeSourceDestKind = "Where the step image is hosted. `CONTAINER_REGISTRY` — a Docker-compatible registry; set `config.docker_image` to the image reference, plus `config.auth` and `config.is_private` when the registry needs credentials."

	WorkflowStepTemplateRuntimeSourceAdditionalConfig = "Additional configuration settings for the runtime source as key-value pairs."

	WorkflowStepTemplateRuntimeSourceConfig = "Specific configuration settings for the runtime source."

	WorkflowStepTemplateRuntimeSourceConfigIsPrivate = "Indicates whether the container registry or repository is private."

	WorkflowStepTemplateRuntimeSourceConfigAuth = "Authentication credentials or method for accessing the private registry or repository. (Sensitive)"

	WorkflowStepTemplateRuntimeSourceConfigDockerImage = "Docker image URI to be used for template execution. Example: `ubuntu:latest`, `myregistry.azurecr.io/myapp:v1.0`"

	WorkflowStepTemplateRuntimeSourceConfigDockerRegistryUsername = "Username for authentication with the Docker registry (if using private registries)."
)

// Workflow Step Template Revision Resource documentation
const (
	WorkflowStepTemplateRevisionId = "ID of the revision in the format `templateId:revisionNumber`."

	WorkflowStepTemplateRevisionTemplateId = "ID of the parent workflow step template."

	WorkflowStepTemplateRevisionAlias = "Alias for the revision to easily identify it."

	WorkflowStepTemplateRevisionNotes = "Notes or changelog information for this revision."

	WorkflowStepTemplateRevisionDescription = "A brief description of the workflow step template revision. Must be less than 256 characters."

	WorkflowStepTemplateRevisionType = "Which family of template this revision belongs to. Read-only, and always `WORKFLOW_STEP` here. StackGuardian uses the same field across all template types: <ul><li>`WORKFLOW_STEP` — a workflow step template.</li><li>`IAC` — a workflow template.</li><li>`IAC_GROUP` — a stack template.</li><li>`IAC_POLICY` — a policy template.</li></ul>"

	WorkflowStepTemplateRevisionSourceConfigKind = "Where the step's runnable definition comes from. `DOCKER_IMAGE` — a container image, pulled from the registry described by `runtime_source`."

	WorkflowStepTemplateRevisionIsPublic = `Whether the revision is publicly available. Valid values:
	<span style="background-color: #eff0f0; color: #e53835;">0</span> (false),
	<span style="background-color: #eff0f0; color: #e53835;">1</span> (true)`

	WorkflowStepTemplateRevisionTags = "A list of tags associated with the revision. A maximum of 10 tags are allowed."

	WorkflowStepTemplateRevisionContextTags = "Contextual key-value tags that provide additional context to the main tags."

	WorkflowStepTemplateRevisionRuntimeSource = "Runtime source configuration for the revision."

	WorkflowStepTemplateRevisionRuntimeSourceDestKind = "Where the step image is hosted. `CONTAINER_REGISTRY` — a Docker-compatible registry; set `config.docker_image` to the image reference, plus `config.auth` and `config.is_private` when the registry needs credentials."

	WorkflowStepTemplateRevisionRuntimeSourceAdditionalConfig = "Additional configuration settings for the runtime source as key-value pairs."

	WorkflowStepTemplateRevisionRuntimeSourceConfig = "Specific configuration settings for the runtime source."

	WorkflowStepTemplateRevisionRuntimeSourceConfigIsPrivate = "Indicates whether the container registry or repository is private."

	WorkflowStepTemplateRevisionRuntimeSourceConfigAuth = "Authentication credentials or method for accessing the private registry or repository. (Sensitive)"

	WorkflowStepTemplateRevisionRuntimeSourceConfigDockerImage = "Docker image URI to be used for revision execution."

	WorkflowStepTemplateRevisionRuntimeSourceConfigDockerRegistryUsername = "Username for authentication with the Docker registry (if using private registries)."
)
