resource "stackguardian_stack_template" "example" {
  template_name      = "my-stack-template"
  source_config_kind = "TERRAFORM"
  is_public          = "0"
  tags               = ["terraform", "production"]
}

# Example 1: Basic stack template revision
resource "stackguardian_stack_template_revision" "basic" {
  parent_template_id = stackguardian_stack_template.example.id
  alias              = "v1"
  notes              = "Initial revision"
  description        = "First revision of the stack template"
  source_config_kind = "TERRAFORM"

  workflows_config = {
    workflows = [
      {
        id            = "d8dfaf15-2ad9-da29-8af0-c6b288b12089"
        template_id   = "my-workflow-template"
        resource_name = "wf-1"

        terraform_config = {
          managed_terraform_state = true
          terraform_version       = "1.5.7"
        }
      }
    ]
  }
}

# Example 2: Stack template revision with VCS config
resource "stackguardian_stack_template_revision" "with_vcs" {
  parent_template_id = stackguardian_stack_template.example.id
  alias              = "v2"
  notes              = "Revision with VCS configuration"
  description        = "Stack template revision with VCS and input data"
  source_config_kind = "TERRAFORM"

  workflows_config = {
    workflows = [
      {
        id            = "d8dfaf15-2ad9-da29-8af0-c6b288b12089"
        template_id   = "my-workflow-template"
        resource_name = "wf-1"

        vcs_config = {
          iac_vcs_config = {
            use_marketplace_template = true
            iac_template_id          = "my-workflow-template"
          }
          iac_input_data = {
            schema_type = "RAW_JSON"
            data = jsonencode({
              bucket_region = "eu-central-1"
            })
          }
        }

        terraform_config = {
          managed_terraform_state = true
          terraform_version       = "1.5.7"
        }
      }
    ]
  }
}
