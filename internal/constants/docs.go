package constants

////////// Resource

// Connector
const (
	SettingsKindMarkdownDoc = `
	The type of connector<br>

	Values with supported config fields:

	**VCS Connectors**
	- <span style="background-color: #eff0f0; color: #e53835;">GITHUB_COM <a href="https://docs.stackguardian.io/docs/connectors/vcs/githubcom/"><span class="fa fa-external-link"></span></span></a>
		- github_com_url
		- github_http_url
	- <span style="background-color: #eff0f0; color: #e53835;">GITHUB_APP_CUSTOM <a href="https://docs.stackguardian.io/docs/connectors/vcs/github_enterprise/"><span class="fa fa-external-link"></span></span></a>
		- github_app_client_id
		- github_app_client_secret
		- github_app_id
		- github_app_pem_file_content
		- github_app_webhook_secret
		- github_app_webhook_url
	- <span style="background-color: #eff0f0; color: #e53835;">BITBUCKET_ORG <a href="https://docs.stackguardian.io/docs/connectors/vcs/bitbucket/"><span class="fa fa-external-link"></span></span></a>
		- bitbucket_creds
	- <span style="background-color: #eff0f0; color: #e53835;">GITLAB_COM <a href="https://docs.stackguardian.io/docs/connectors/vcs/gitlabcom/"><span class="fa fa-external-link"></span></span></a>
		- gitlab_api_url
		- gitlab_creds
		- gitlab_http_url
	- <span style="background-color: #eff0f0; color: #e53835;">AZURE_DEVOPS <a href="https://docs.stackguardian.io/docs/connectors/vcs/azuredevops/"><span class="fa fa-external-link"></span></span></a>
		- azure_devops_api_url
		- azure_devops_http_url
		- azure_creds</br>

	**Cloud Connectors**
	- <span style="background-color: #eff0f0; color: #e53835;">AWS_STATIC <a href="https://docs.stackguardian.io/docs/connectors/csp/aws/#access-keys"><span class="fa fa-external-link"></span></span></a>
		- aws_access_key_id
		- aws_secret_access_key
		- aws_default_region
	- <span style="background-color: #eff0f0; color: #e53835;">AWS_RBAC <a href="https://docs.stackguardian.io/docs/connectors/csp/aws/#roles-or-rbac-recommended"><span class="fa fa-external-link"></span></span></a>
		- role_arn
		- external_id
	- <span style="background-color: #eff0f0; color: #e53835;">AWS_OIDC <a href="https://docs.stackguardian.io/docs/connectors/csp/aws/#using-oidc-identity-provider"><span class="fa fa-external-link"></span></span></a>
		- role_arn
	- <span style="background-color: #eff0f0; color: #e53835;">GCP_STATIC <a href="https://docs.stackguardian.io/docs/connectors/csp/gcp/#using-service-account"><span class="fa fa-external-link"></span></span></a>
		- gcp_config_file_content
	- <span style="background-color: #eff0f0; color: #e53835;">GCP_OIDC <a href="https://docs.stackguardian.io/docs/connectors/csp/gcp/"><span class="fa fa-external-link"></span></span></a>
		- gcp_config_file_content
	- <span style="background-color: #eff0f0; color: #e53835;">AZURE_STATIC <a href="https://docs.stackguardian.io/docs/connectors/csp/azure/#service-principal-with-client-secret"><span class="fa fa-external-link"></span></span></a>
		- arm_client_id
		- arm_client_secret
		- arm_subscription_id
		- arm_tenant_id
	- <span style="background-color: #eff0f0; color: #e53835;">AZURE_OIDC <a href="https://docs.stackguardian.io/docs/connectors/csp/azure/#service-principal-with-workload-identity"><span class="fa fa-external-link"></span></span></a>
		- arm_tenant_id
		- arm_subscription_id
		- arm_client_id
`
	SettingsConfig                        = "Configuration settings for the connector's secrets"
	SettingsConfigRoleArn                 = "The Amazon Resource Name (ARN) of the role that the caller is assuming."
	SettingsConfigExternalId              = `A unique identifier that is used to assume the role in the customers' AWS accounts. Should start with org name followed by ":" and a random string. SG_ORG_NAME:ElfygiFglfldTwnDFpAScQkvgvHTGV`
	SettingsConfigDurationSeconds         = "The duration, in seconds, of the role session. Default is 3600 seconds (1 hour)."
	SettingsConfigInstallationId          = "The installation ID for GitHub applications."
	SettingsConfigGithubAppId             = "The application ID for the GitHub app."
	SettingsConfigGithubAppWebhookSecret  = "Webhook secret for the GitHub app."
	SettingsConfigGithubApiUrl            = "Base URL for the GitHub API."
	SettingsConfigGithubHttpUrl           = "HTTP URL for accessing the GitHub repository."
	SettingsConfigGithubAppClientId       = "Client ID for the GitHub app."
	SettingsConfigGithubAppClientSecret   = "Client secret for the GitHub app."
	SettingsConfigGithubAppPemFileContent = "Content of the PEM file for the GitHub app."
	SettingsConfigGithubAppWebhookUrl     = "Webhook URL for the GitHub app."
	SettingsConfigGitlabCreds             = "Credentials for GitLab integration."
	SettingsConfigGitlabHttpUrl           = "HTTP URL for accessing the GitLab repository."
	SettingsConfigGitlabApiUrl            = "Base URL for the GitLab API."
	SettingsConfigAzureCreds              = "Credentials for Azure integration."
	SettingsConfigAzureDevopsHttpUrl      = "HTTP URL for accessing Azure DevOps services."
	SettingsConfigAzureDevopsApiUrl       = "Base URL for Azure DevOps API."
	SettingsConfigBitbucketCreds          = "Credentials for Bitbucket integration."
	SettingsConfigAwsAccessKeyId          = "AWS access key ID for authentication."
	SettingsConfigAwsSecretAccessKey      = "AWS secret access key for authentication."
	SettingsConfigAwsDefaultRegion        = "Default AWS region for resource operations."
	SettingsConfigArmTenantId             = "Azure Resource Manager tenant ID."
	SettingsConfigArmSubscriptionId       = "Azure Resource Manager subscription ID."
	SettingsConfigArmClientId             = "Client ID for Azure Resource Manager."
	SettingsConfigArmClientSecret         = "Client secret for Azure Resource Manager."
	SettingsConfigGcpConfigFileContent    = "Content of the GCP configuration file."

	DiscoverySettings                      = "Settings for discovery insights related to the connector."
	DiscoverySettingsBenchmarks            = "Statistics for various StackGuardian resources."
	DiscoverySettingsBenchmarksChecks      = "List of checks performed during discovery."
	DiscoverySettingsBenchmarksDescription = "A description of the benchmark. It must be less than 256 characters."
	DiscoverySettingsBenchmarksLabel       = "Label associated with the discovery."

	DiscoverySettingsBenchmarksRuntimeSource                     = "Source configuration type and settings definition"
	DiscoverySettingsBenchmarksRuntimeSourceSourceConfigDestKind = "Kind of the source configuration destination. Valid examples include eg:- AWS_RBAC, AZURE_STATIC."

	DiscoverySettingsBenchmarksRuntimeSourceConfig                 = "Specific configuration settings for runtime source."
	DiscoverySettingsBenchmarksRuntimeSourceConfigIncludeSubModule = "Indicates whether to include sub-modules."
	DiscoverySettingsBenchmarksRuntimeSourceConfigRef              = "Reference identifier for the repository."
	DiscoverySettingsBenchmarksRuntimeSourceConfigGitCoreAutoCRLF  = "Indicates if core.autocrlf should be enabled."
	DiscoverySettingsBenchmarksRuntimeSourceConfigAuth             = "Authentication method for accessing the repository."
	DiscoverySettingsBenchmarksRuntimeSourceConfigWorkingDir       = "Working directory for operations."
	DiscoverySettingsBenchmarksRuntimeSourceConfigRepo             = "Repository name or URL."
	DiscoverySettingsBenchmarksRuntimeSourceConfigIsPrivate        = "Indicates if the repository is private."

	DiscoverySettingsBenchmarksSummaryDescription = "A brief summary of the discovery."
	DiscoverySettingsBenchmarksSummaryTitle       = "Title for the discovery summary."
	DiscoverySettingsBenchmarksDiscoveryInterval  = "Interval for the discovery process."
	DiscoverySettingsBenchmarksIsCustomCheck      = "Indicates if the discovery is a custom check."
	DiscoverySettingsBenchmarksActive             = "Indicates if the discovery is active."

	DiscoverySettingsBenchmarksRegions       = "Regions associated with the discovery."
	DiscoverySettingsBenchmarksRegionsEmails = "List of emails to notify about the discovery."
)

// Role
const (
	AllowedPermissions      = "A map of permissions assigned to the role."
	AllowedPermissionsName  = "The name of the permission."
	AllowedPermissionsPaths = "A map of resource paths to which this permission is scoped."
)

// Role Assignment
const (
	UserId = `Fully qualified user email or group. Examples:
	- Local user: you@example.com
	- SSO user: <SSO-Provider-Name>/you@example.com (e.g., sg-test-sso/you@example.com)
	- SSO group: <SSO-Provider-Name>/group-name (e.g., sg-test-sso/group-devs)`
	EntityType = `Should be one of:
	- <span style="background-color: #eff0f0; color: #e53835;">EMAIL</span>
	- <span style="background-color: #eff0f0; color: #e53835;">GROUP</span>`
	Role = "StackGuardian role name."
)

// Policy
const (
	Approvers                 = "List of stackguardian users"
	NumberOfApprovalsRequired = "Number of approvals required for a policy check to pass"
	EnforcedOn                = "List of Resource path on which this policy is to be applied on"
	PolicyType                = "Type of policy created \"GENERAL\" or \"FILTER.INSIGHT\""

	PolicyConfig       = "Policy configuration"
	PolicyConfigSkip   = "Enable or disable the policy check"
	PolicyConfigOnFail = `Specifies the action to be performed on failure. Options: <span style="background-color: #eff0f0; color: #e53835;">FAIL</span>,
		<span style="background-color: #eff0f0; color: #e53835;">WARN</span>,
		<span style="background-color: #eff0f0; color: #e53835;">PASS</span>,
		<span style="background-color: #eff0f0; color: #e53835;">APPROVAL_REQUIRED</span>`
	PolicyConfigOnPass = `Specifies the action to be performed on pass. Options: <span style="background-color: #eff0f0; color: #e53835;">FAIL</span>,
		<span style="background-color: #eff0f0; color: #e53835;">WARN</span>,
		<span style="background-color: #eff0f0; color: #e53835;">PASS</span>,
		<span style="background-color: #eff0f0; color: #e53835;">APPROVAL_REQUIRED</span>`
	PolicyConfigInputData           = "Policy definition"
	PolicyConfigInputDataSchemaType = `Specifies the schema type of the policy. Options: <span style="background-color: #eff0f0; color: #e53835;">FORM_JSONSCHEMA</span>,
		<span style="background-color: #eff0f0; color: #e53835;">RAW_JSON</span>,
		<span style="background-color: #eff0f0; color: #e53835;">TIRITH_JSON</span>,
		<span style="background-color: #eff0f0; color: #e53835;">NONE</span>`
	PolicyConfigInputDataData = "Policy body"

	PolicyVCSConfig                    = "Configuration to import policy from version control"
	PolicyVCSConfigMarketplaceTemplate = "Name of the template from marketplace"
	PolicyVCSConfigTemplateId          = "ID of the template from marketplace"

	PolicyVCSConfigCustomSource                     = DiscoverySettingsBenchmarksRuntimeSource
	PolicyVCSConfigCustomSourceSourceConfigDestKind = DiscoverySettingsBenchmarksRuntimeSourceSourceConfigDestKind
	PolicyVCSConfigCustomSourceSourceConfigKind     = `Kind of policy. Options: <span style="background-color: #eff0f0; color: #e53835;">OPA_REGO</span>,
		<span style="background-color: #eff0f0; color: #e53835;">SG_POLICY_FRAMEWORK</span>,
	`
	PolicyVCSConfigCustomSourceConfig                  = DiscoverySettingsBenchmarksRuntimeSourceConfig
	PolicyVCSConfigCustomSourceRef                     = DiscoverySettingsBenchmarksRuntimeSourceConfigRef
	PolicyVCSConfigCustomSourceGitCoreAutoCRLF         = DiscoverySettingsBenchmarksRuntimeSourceConfigGitCoreAutoCRLF
	PolicyVCSConfigCustomSourceGitSparseCheckoutConfig = "Configuration for git sparse checkout"
	PolicyVCSConfigCustomSourceAuth                    = DiscoverySettingsBenchmarksRuntimeSourceConfigAuth
	PolicyVCSConfigCustomSourceWorkingDir              = DiscoverySettingsBenchmarksRuntimeSourceConfigWorkingDir
	PolicyVCSConfigCustomSourceRepo                    = DiscoverySettingsBenchmarksRuntimeSourceConfigRepo
	PolicyVCSConfigCustomSourceIsPrivate               = DiscoverySettingsBenchmarksRuntimeSourceConfigIsPrivate

	PolicyVCSConfigAdditionalConfig = "Additional configuration for the policy"
)

// Runner Group
const (
	RunnerToken            = "Private token of the runner group"
	MaxNumberOfRunners     = "Maximum number of runners allowed in a runner group"
	DockerImage            = "Docker image to used to execute workflows"
	DockerRegistryUsername = "Username for docker register"
)

// Workflow Group
const (
	WorkflowGroupResourceName = "Name of the workflow group. Must be less than 100 characters. Allowed characters are ^[a-zA-Z0-9_/]+$"
)

// Role Assignment or User
const (
	SendEmail = "Enable or disable email notification to the user on creation"
)

// Common
const (
	ResourceName         = "Name of the %s. Must be less than 100 characters. Allowed characters are ^[a-zA-Z0-9_]+$"
	Id                   = "ID of the resource — Use this attribute to reference the resource in other resources. The `resource_name` attribute is still available but its use is discouraged and may not work in some cases."
	Description          = "A brief description of the %s. Must be less than 256 characters."
	Tags                 = "A list of tags associated with the %s. A maximum of 10 tags are allowed."
	StorageBackendConfig = "Configuration for storing runner logs"
	RunnerGroupType      = `Platform of the storage:
	- <span style="background-color: #eff0f0; color: #e53835;">aws_s3</span>
	- <span style="background-color: #eff0f0; color: #e53835;">azure_blob_storage</span>
`
	AzureBlobStorageAccountName = "Account of your azure blob storage"
	AzureBlobStorageAccessKey   = "Access key for you blob storage account"
	S3BucketName                = "S3 buckget name"
	AWSRegion                   = "AWS region where the bucket is placed"
	Auth                        = "Authentication required by the runner to access the backend storage. Required only for type \"aws_s3\""
	IntegrationId               = "SG Connector Id. Required only for type \"aws_s3\" eg: /integrations/test-connector"
)

////////////// Data Source

// Common
const (
	DatasourceId                      = "ID of the resource. Should be used to import the resource."
	StackguardianStack                = "Stackguardian stack name"
	StackguardianWorkflow             = "Stackguardian workflow name"
	StackguardianWorkflowGroup        = "Stackguardian workflow group name"
	DataSourceDataJson                = "Raw JSON body"
	DataSourceData                    = "Map of k/v pairs with value as JSON string"
	DatasourceResourceNameDeprecation = " <span style='color: #e53835;'>Deprecated:</span> The `resource_name` attribute is still available but its use is discouraged and may not work in some cases. Use `id`."
)

// api token
const (
	RunnerGroupToken = "Runner Group token"
	RunnerGroupId    = "Runner group ID"
)
