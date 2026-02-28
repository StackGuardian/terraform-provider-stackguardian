# Example 1: Basic workflow template revision
resource "stackguardian_workflow_template_revision" "basic" {
  template_id        = "my-terraform-template"
  alias              = "v1"
  source_config_kind = "TERRAFORM"
  is_public          = "0"
  tags               = ["terraform", "revision"]
}

# Example 2: Workflow template revision with detailed configuration
resource "stackguardian_workflow_template_revision" "detailed" {
  template_id                  = "my-terraform-template"
  alias                        = "v3"
  source_config_kind           = "TERRAFORM"
  is_public                    = "0"
  notes                        = "Production-ready revision with approval workflow"
  user_job_cpu                 = 2
  user_job_memory              = 4096
  number_of_approvals_required = 1
  tags                         = ["terraform", "production", "approved"]
  approvers                    = ["user1@example.com", "user2@example.com"]
}
