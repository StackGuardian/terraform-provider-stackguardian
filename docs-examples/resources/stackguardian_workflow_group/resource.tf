<<<<<<< HEAD
resource "stackguardian_workflow_group" "testing" {
  resource_name = "ONBOARDING-Project01-DevOps"
  description   = "Onboarding example  of terraform-provider-stackguardian for WorkflowGroup"
  tags          = ["tf-provider-example", "onboarding"]
}
=======
# Example for creating a Workflow Group in StackGuardian
resource "stackguardian_workflow_group" "example_workflow_group" {
  resource_name = "example-workflow-group"
  description   = "Example of terraform-provider-stackguardian for Workflow Group"
  tags          = ["example-tag"]
}
>>>>>>> main
