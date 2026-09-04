########################################################################
# StackGuardian provider quickstart
#
# Creates one working deployment, end to end:
#
#   workflow_group   the folder the workflow lives in
#        |
#   connector        the AWS credentials the workflow deploys with
#        |
#   workflow_git     the workflow itself, pointing at your Terraform repo
#
# Before you apply:
#   1. export STACKGUARDIAN_API_KEY=...   (Organization Settings -> API Keys)
#   2. export STACKGUARDIAN_ORG_NAME=...
#   3. Replace the role_arn and external_id below with your own.
#   4. Point `repo` at a repository containing Terraform.
#
# Then:  terraform init && terraform plan && terraform apply
########################################################################

terraform {
  required_providers {
    stackguardian = {
      source  = "StackGuardian/stackguardian"
      version = "~> 1.12" #provider-version
    }
  }
}

provider "stackguardian" {
  # api_key and org_name are read from STACKGUARDIAN_API_KEY and
  # STACKGUARDIAN_ORG_NAME. Use https://api.us.stackguardian.io for the US region.
  api_uri = "https://api.app.stackguardian.io"
}

# 1. A workflow group is the folder that organizes workflows, and the unit that
#    roles are scoped to. Nest one by putting "/" in resource_name.
resource "stackguardian_workflow_group" "quickstart" {
  resource_name = "quickstart"
  description   = "Created by the StackGuardian provider quickstart"
  tags          = ["quickstart"]
}

# 2. A connector holds the credentials a workflow deploys with. This one uses an
#    AWS IAM role; see the connector docs for the other supported kinds.
resource "stackguardian_connector" "aws" {
  resource_name = "quickstart-aws"
  description   = "AWS credentials used by the quickstart workflow"

  settings = {
    kind = "AWS_RBAC"

    config = [{
      role_arn         = "arn:aws:iam::000000000000:role/StackGuardian"
      external_id      = "REPLACE-ME"
      duration_seconds = "3600"
    }]
  }
}

# 3. The workflow. deployment_platform_config wires in the connector created
#    above, so the two resources are genuinely connected rather than hardcoded.
resource "stackguardian_workflow_git" "quickstart" {
  workflow_group_id = stackguardian_workflow_group.quickstart.resource_name
  id                = "quickstart-workflow"
  wf_type           = "TERRAFORM"

  description = "Deploys the Terraform in the referenced repository"
  tags        = ["quickstart"]

  vcs_config = {
    iac_vcs_config = {
      custom_source = {
        source_config_dest_kind = "GIT_OTHER"
        config = {
          is_private  = false
          repo        = "https://github.com/my-org/my-terraform-repo.git"
          working_dir = "."
          ref         = "main"
        }
      }
    }
  }

  deployment_platform_config = [{
    kind = "AWS_RBAC"
    config = {
      integration_id = stackguardian_connector.aws.id
    }
  }]
}

output "workflow_group" {
  description = "Name of the workflow group that was created."
  value       = stackguardian_workflow_group.quickstart.resource_name
}

output "workflow_id" {
  description = "ID of the workflow that was created."
  value       = stackguardian_workflow_git.quickstart.id
}
