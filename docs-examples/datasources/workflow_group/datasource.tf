# Look up an existing workflow group so workflows can be created inside it without this
# configuration owning the group.
#
# For a nested group, give the full path -- for example "platform/networking".
data "stackguardian_workflow_group" "platform" {
  resource_name = "platform"
}

resource "stackguardian_workflow_git" "example" {
  workflow_group_id = data.stackguardian_workflow_group.platform.resource_name
  id                = "deploy-vpc"
  wf_type           = "TERRAFORM"

  vcs_config = {
    iac_vcs_config = {
      custom_source = {
        source_config_dest_kind = "GIT_OTHER"
        config = {
          is_private = false
          repo       = "https://github.com/my-org/vpc.git"
        }
      }
    }
  }
}
