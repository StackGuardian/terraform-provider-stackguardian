# Read an existing git-based workflow -- for example to confirm which repository and
# revision it deploys from, without managing the workflow in this configuration.
#
# A workflow is identified by its own id plus the group that contains it. For a nested
# group, give the full path.
data "stackguardian_workflow_git" "example" {
  workflow_group_id = "my-workflow-group"
  id                = "my-workflow"
}

output "workflow_git_wf_type" {
  value = data.stackguardian_workflow_git.example.wf_type
}

output "workflow_git_repo" {
  description = "Repository the workflow deploys from."
  value       = data.stackguardian_workflow_git.example.vcs_config.iac_vcs_config.custom_source.config.repo
}
