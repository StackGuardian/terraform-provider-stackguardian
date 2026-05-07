resource "stackguardian_workflow_group" "example" {
  resource_name = "example-workflow-group"
  description   = "Example workflow group"
}

# Example 1: TERRAFORM workflow using a StackGuardian marketplace template
resource "stackguardian_workflow" "marketplace" {
  workflow_group_id = stackguardian_workflow_group.example.id
  resource_name     = "example-marketplace-workflow"
  description       = "Deploy an S3 website from the StackGuardian marketplace"
  wf_type           = "TERRAFORM"
  tags              = ["example", "terraform", "marketplace"]

  deployment_platform_config = [{
    kind = "AWS_RBAC"
    config = {
      integration_id = "/integrations/aws-connector"
    }
  }]

  vcs_config = {
    iac_vcs_config = {
      use_marketplace_template = true
      iac_template_id          = "/stackguardian/aws-s3-demo-website:4"
    }
    iac_input_data = {
      schema_type = "FORM_JSONSCHEMA"
      data        = jsonencode({ shop_name = "StackGuardian", bucket_region = "eu-central-1" })
    }
  }

  terraform_config = {
    managed_terraform_state = true
    terraform_version       = "1.5.7"
  }
}

# Example 2: TERRAFORM workflow using a custom git source
resource "stackguardian_workflow" "custom_vcs" {
  workflow_group_id            = stackguardian_workflow_group.example.id
  resource_name                = "example-custom-vcs-workflow"
  description                  = "Deploy infrastructure from a private GitHub repository"
  wf_type                      = "TERRAFORM"
  tags                         = ["example", "terraform", "custom-vcs"]
  approvers                    = ["engineer@example.com"]
  number_of_approvals_required = 1

  environment_variables = [{
    config = {
      text_value = "production"
      var_name   = "ENVIRONMENT"
    }
    kind = "TEXT"
  }]

  deployment_platform_config = [{
    kind = "AWS_RBAC"
    config = {
      integration_id = "/integrations/aws-connector"
    }
  }]

  vcs_config = {
    iac_vcs_config = {
      use_marketplace_template = false
      custom_source = {
        source_config_dest_kind = "GITHUB_COM"
        config = {
          is_private = true
          auth       = "/integrations/github-connector"
          repo       = "https://github.com/my-org/my-infra-repo.git"
          ref        = "main"
          working_dir = "terraform/envs/prod"
        }
      }
    }
  }

  terraform_config = {
    managed_terraform_state = true
    terraform_version       = "1.5.7"
    approval_pre_apply      = true
  }

  mini_steps = {
    notifications = {
      email = {
        errored = [{
          recipients = ["oncall@example.com"]
        }]
        completed = [{
          recipients = ["team@example.com"]
        }]
      }
    }
  }
}
