# Create a workflow from a workflow template revision. Any attribute the user does not
# declare is inherited from the template revision referenced by iac_template_id.
resource "stackguardian_workflow_from_template" "example" {
  workflow_group_id = "my-workflow-group"
  id                = "my-workflow"
  wf_type           = "TERRAFORM"

  vcs_config = {
    iac_vcs_config = {
      iac_template_id = "/my-org/my-template:1"
    }
  }

  # Declared fields win over the template default; omitted fields are inherited.
  terraform_config = {
    terraform_version = "1.5.0"
  }
}
