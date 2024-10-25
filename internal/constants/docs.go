package constants

////////// Resource

// Connector
const (
	SettingsKindMarkdownDoc = `
	The type of connector<br>
	Values with supported config fields:
	- <span style="background-color: #eff0f0; color: #e53835;">GITHUB_COM</span>
		- github_com_url
		- github_http_url
	- <span style="background-color: #eff0f0; color: #e53835;">GITHUB_APP_CUSTOM</span>
		- github_app_client_id
		- github_app_client_secret
		- github_app_id
		- github_app_pem_file_content
		- github_app_webhook_secret
		- github_app_webhook_url
	- <span style="background-color: #eff0f0; color: #e53835;">AWS_STATIC</span>
		- aws_access_key_id
		- aws_secret_access_key
		- aws_default_region
	- <span style="background-color: #eff0f0; color: #e53835;">AWS_RBAC</span>
		- role_arn
		- external_id
		- arm_client_id
	- <span style="background-color: #eff0f0; color: #e53835;">AWS_OIDC</span>
		- role_arn
	- <span style="background-color: #eff0f0; color: #e53835;">GCP_STATIC</span>
		- gcp_config_file_content
	- <span style="background-color: #eff0f0; color: #e53835;">AZURE_STATIC</span>
		- arm_client_id
		- arm_client_secret
		- arm_subscription_id
		- arm_tenant_id
	- <span style="background-color: #eff0f0; color: #e53835;">AZURE_OIDC</span>
		- arm_tenant_id
		- arm_subscription_id
		- arm_client_id
	- <span style="background-color: #eff0f0; color: #e53835;">BITBUCKET_ORG</span>
		- bitbucket_creds
	- <span style="background-color: #eff0f0; color: #e53835;">GITLAB_COM</span>
		- gitlab_api_url
		- gitlab_creds
		- gitlab_http_url
	- <span style="background-color: #eff0f0; color: #e53835;">AZURE_DEVOPS</span>
		- azure_devops_api_url
		- azure_devops_http_url
		- azure_creds
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

	DiscoverySettingsBenchmarksRuntimeSource                     = ""
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
	UserId     = "Fully qualified user email or group. Example: you@example.com for a local user, <SSO Login Method Identifier>/you@example.com for a SSO email when entity_type in EMAIL. <SSO Login Method Identifier>/group-devs when entity_type in GROUP."
	EntityType = `Should be one of:
	- <span style="background-color: #eff0f0; color: #e53835;">EMAIL</span>
	- <span style="background-color: #eff0f0; color: #e53835;">GROUP</span>`
	Role = "StackGuardian role name."
)

// Common
const (
	ResourceName = "The name of the %s. Must be less than 100 characters. Allowed characters are ^[a-zA-Z0-9_]+$"
	Description  = "A brief description of the %s. Must be less than 256 characters."
	Tags         = "A list of tags associated with the %s. A maximum of 10 tags are allowed."
)

////////////// Data Source

// Common
const (
	StackguardianStack         = "Stackguardian stack name"
	StackguardianWorkflow      = "Stackguardian workflow name"
	StackguardianWorkflowGroup = "Stackguardian workflow group name"
	DataSourceDataJson         = "Raw JSON body"
	DataSourceData             = "Map of k/v pairs with value as JSON string"
)
