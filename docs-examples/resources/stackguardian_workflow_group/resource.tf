resource "stackguardian_workflow_group" "testing" {
  resource_name = "ONBOARDING-Project01-DevOps"
  description   = "Onboarding example  of terraform-provider-stackguardian for WorkflowGroup"
  tags          = ["tf-provider-example", "onboarding"]
}
