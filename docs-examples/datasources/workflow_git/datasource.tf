data "stackguardian_workflow_git" "example" {
  workflow_group_id = "my-workflow-group"
  id                = "my-workflow"
}

output "workflow_git_wf_type" {
  value = data.stackguardian_workflow_git.example.wf_type
}

output "workflow_git_description" {
  value = data.stackguardian_workflow_git.example.description
}

output "workflow_git_repo" {
  value = data.stackguardian_workflow_git.example.vcs_config.iac_vcs_config.custom_source.config.repo
}
