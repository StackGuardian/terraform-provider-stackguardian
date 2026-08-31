# Look up a connector that is managed elsewhere -- by another team, or created in the
# StackGuardian UI -- so a workflow can deploy with it without this configuration owning it.
data "stackguardian_connector" "shared_aws" {
  resource_name = "shared-aws-production"
}

resource "stackguardian_workflow_git" "example" {
  workflow_group_id = "my-workflow-group"
  id                = "deploy-network"
  wf_type           = "TERRAFORM"

  vcs_config = {
    iac_vcs_config = {
      custom_source = {
        source_config_dest_kind = "GIT_OTHER"
        config = {
          is_private = false
          repo       = "https://github.com/my-org/network.git"
        }
      }
    }
  }

  # The connector supplies the cloud credentials the workflow deploys with.
  deployment_platform_config = [{
    kind = "AWS_RBAC"
    config = {
      integration_id = data.stackguardian_connector.shared_aws.id
    }
  }]
}
