package constants

// Workflow Step Template - Common documentation
const (
	WorkflowStepTemplateSourceConfigKindCommon = `Source configuration kind that defines how the template is deployed. Valid values:
	<span style="background-color: #eff0f0; color: #e53835;">DOCKER_IMAGE</span>,
	<span style="background-color: #eff0f0; color: #e53835;">GIT_REPO</span>,
	<span style="background-color: #eff0f0; color: #e53835;">S3</span>`

	WorkflowStepTemplateIsActiveCommon = `Whether the workflow step template is active. Valid values:
	<span style="background-color: #eff0f0; color: #e53835;">0</span> (false),
	<span style="background-color: #eff0f0; color: #e53835;">1</span> (true)`

	WorkflowStepTemplateIsPublicCommon = `Whether the workflow step template is publicly available. Valid values:
	<span style="background-color: #eff0f0; color: #e53835;">0</span> (false),
	<span style="background-color: #eff0f0; color: #e53835;">1</span> (true)`

	WorkflowStepTemplateRuntimeSourceDestKindCommon = "Destination kind for the runtime source configuration. Examples:" +
		"\n<span style=\"background-color: #eff0f0; color: #e53835;\">CONTAINER_REGISTRY</span>," +
		"\n<span style=\"background-color: #eff0f0; color: #e53835;\">GIT</span>," +
		"\n<span style=\"background-color: #eff0f0; color: #e53835;\">S3</span>"

	WorkflowStepTemplateRuntimeSourceConfigIsPrivateCommon = "Indicates whether the container registry or repository is private."

	WorkflowStepTemplateRuntimeSourceConfigAuthCommon = "Authentication credentials or method for accessing the private registry or repository. (Sensitive)"

	WorkflowStepTemplateRuntimeSourceConfigDockerImageCommon = "Docker image URI to be used for template execution. Example: `ubuntu:latest`, `myregistry.azurecr.io/myapp:v1.0`"

	WorkflowStepTemplateRuntimeSourceConfigDockerRegistryUsernameCommon = "Username for authentication with the Docker registry (if using private registries)."

	WorkflowStepTemplateRuntimeSourceConfigLocalWorkspaceDirCommon = "Workfing directory path."
)

// Workflow Step Template Resource documentation
const (
	WorkflowStepTemplateName = "Name of the workflow step template. Must be less than 100 characters."

	WorkflowStepTemplateDescription = "A brief description of the workflow step template. Must be less than 256 characters."

	WorkflowStepTemplateType = `Type of the template. Valid values:
	<span style="background-color: #eff0f0; color: #e53835;">WORKFLOW_STEP</span>,
	<span style="background-color: #eff0f0; color: #e53835;">IAC</span>,
	<span style="background-color: #eff0f0; color: #e53835;">IAC_GROUP</span>,
	<span style="background-color: #eff0f0; color: #e53835;">IAC_POLICY</span>`

	WorkflowStepTemplateIsActive = `Whether the workflow step template is active. Valid values:
	<span style="background-color: #eff0f0; color: #e53835;">0</span> (false),
	<span style="background-color: #eff0f0; color: #e53835;">1</span> (true)`

	WorkflowStepTemplateIsPublic = `Whether the workflow step template is publicly available. Valid values:
	<span style="background-color: #eff0f0; color: #e53835;">0</span> (false),
	<span style="background-color: #eff0f0; color: #e53835;">1</span> (true)`

	WorkflowStepTemplateTags = "A list of tags associated with the workflow step template. A maximum of 10 tags are allowed."

	WorkflowStepTemplateContextTags = "Contextual key-value tags that provide additional context to the main tags."

	WorkflowStepTemplateSharedOrgsList = "List of organization IDs with which this template is shared."

	WorkflowStepTemplateSourceConfigKind = `Source configuration kind that defines how the template is deployed. Valid values:
	<span style="background-color: #eff0f0; color: #e53835;">DOCKER_IMAGE</span>,
	<span style="background-color: #eff0f0; color: #e53835;">GIT_REPO</span>,
	<span style="background-color: #eff0f0; color: #e53835;">S3</span>`

	WorkflowStepTemplateLatestRevision = "Latest revision number of the template."

	WorkflowStepTemplateNextRevision = "Next revision number that will be used for the template."

	WorkflowStepTemplateRuntimeSource = "Runtime source configuration that defines where and how the template code is stored and executed."

	WorkflowStepTemplateRuntimeSourceDestKind = "Destination kind for the runtime source configuration. Examples:" +
		"\n<span style=\"background-color: #eff0f0; color: #e53835;\">CONTAINER_REGISTRY</span>," +
		"\n<span style=\"background-color: #eff0f0; color: #e53835;\">GIT</span>," +
		"\n<span style=\"background-color: #eff0f0; color: #e53835;\">S3</span>"

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

	WorkflowStepTemplateRevisionType = `Type of the template. Valid values:
	<span style="background-color: #eff0f0; color: #e53835;">WORKFLOW_STEP</span>,
	<span style="background-color: #eff0f0; color: #e53835;">IAC</span>,
	<span style="background-color: #eff0f0; color: #e53835;">IAC_GROUP</span>,
	<span style="background-color: #eff0f0; color: #e53835;">IAC_POLICY</span>`

	WorkflowStepTemplateRevisionSourceConfigKind = `Source configuration kind. Valid values:
	<span style="background-color: #eff0f0; color: #e53835;">DOCKER_IMAGE</span>,
	<span style="background-color: #eff0f0; color: #e53835;">GIT_REPO</span>,
	<span style="background-color: #eff0f0; color: #e53835;">S3</span>`

	WorkflowStepTemplateRevisionIsActive = `Whether the revision is active. Valid values:
	<span style="background-color: #eff0f0; color: #e53835;">0</span> (false),
	<span style="background-color: #eff0f0; color: #e53835;">1</span> (true)`

	WorkflowStepTemplateRevisionIsPublic = `Whether the revision is publicly available. Valid values:
	<span style="background-color: #eff0f0; color: #e53835;">0</span> (false),
	<span style="background-color: #eff0f0; color: #e53835;">1</span> (true)`

	WorkflowStepTemplateRevisionTags = "A list of tags associated with the revision. A maximum of 10 tags are allowed."

	WorkflowStepTemplateRevisionContextTags = "Contextual key-value tags that provide additional context to the main tags."

	WorkflowStepTemplateRevisionRuntimeSource = "Runtime source configuration for the revision."

	WorkflowStepTemplateRevisionRuntimeSourceDestKind = "Destination kind for the runtime source configuration."

	WorkflowStepTemplateRevisionRuntimeSourceAdditionalConfig = "Additional configuration settings for the runtime source as key-value pairs."

	WorkflowStepTemplateRevisionRuntimeSourceConfig = "Specific configuration settings for the runtime source."

	WorkflowStepTemplateRevisionRuntimeSourceConfigIsPrivate = "Indicates whether the container registry or repository is private."

	WorkflowStepTemplateRevisionRuntimeSourceConfigAuth = "Authentication credentials or method for accessing the private registry or repository. (Sensitive)"

	WorkflowStepTemplateRevisionRuntimeSourceConfigDockerImage = "Docker image URI to be used for revision execution."

	WorkflowStepTemplateRevisionRuntimeSourceConfigDockerRegistryUsername = "Username for authentication with the Docker registry (if using private registries)."
)
