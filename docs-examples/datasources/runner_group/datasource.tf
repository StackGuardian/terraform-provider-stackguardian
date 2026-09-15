# Look up a runner group so a workflow can be pinned to it without this configuration
# managing the group itself.
data "stackguardian_runner_group" "private_runners" {
  resource_name = "private-runners"
}

resource "stackguardian_workflow_git" "example" {
  workflow_group_id = "my-workflow-group"
  id                = "deploy-on-private-runners"
  wf_type           = "TERRAFORM"

  # type = "private" pins the workflow to the named groups. With "shared" it would use
  # StackGuardian's shared runners and `names` would not apply.
  runner_constraints = {
    type  = "private"
    names = [data.stackguardian_runner_group.private_runners.resource_name]
  }

  vcs_config = {
    iac_vcs_config = {
      custom_source = {
        source_config_dest_kind = "GIT_OTHER"
        config = {
          is_private = false
          repo       = "https://github.com/my-org/infra.git"
        }
      }
    }
  }
}
