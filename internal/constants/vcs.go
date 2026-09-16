package constants

// SourceConfigDestKind enum values for VCS providers. Used in runtime_source,
// vcs_config, and vcs_triggers across workflow, workflow_template, and
// workflow_template_revision resources.
const (
	GithubCom       = "GITHUB_COM"
	GithubAppCustom = "GITHUB_APP_CUSTOM"
	GitOther        = "GIT_OTHER"
	BitbucketOrg    = "BITBUCKET_ORG"
	GitlabCom       = "GITLAB_COM"
	AzureDevops     = "AZURE_DEVOPS"
	AzureDevopsSp   = "AZURE_DEVOPS_SP"
)
