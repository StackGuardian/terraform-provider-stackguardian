# Example 1: Basic stack template
resource "stackguardian_stack_template" "basic" {
  template_name      = "my-stack-template"
  source_config_kind = "TERRAFORM"
  is_public          = "0"
  tags               = ["terraform", "production"]
}
