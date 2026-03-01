# Example 1: Basic workflow template
resource "stackguardian_workflow_template" "basic" {
  template_name      = "my-terraform-template"
  source_config_kind = "TERRAFORM"
  is_public          = "0"
  tags               = ["terraform", "production"]
}

# Example 2: Workflow template with runtime source
resource "stackguardian_workflow_template" "with_runtime" {
  template_name      = "template-with-runtime"
  source_config_kind = "TERRAFORM"
  is_public          = "0"
  tags               = ["terraform", "github"]

  runtime_source = {
    source_config_dest_kind = "GITHUB_COM"
    config = {
      is_private = false
      repo       = "https://github.com/example/terraform-modules.git"
    }
  }
}
