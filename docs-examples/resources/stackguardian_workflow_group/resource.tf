resource "stackguardian_workflow_group" "TPS-Example-WorkflowGroup" {
  data = jsonencode({
    "ResourceName" : "TPS-Example",
    "Description" : "Example of terraform-provider-stackguardian for WorkflowGroup",
    "Tags" : ["tf-provider-example"],
    "IsActive" : 1,
  })
}
