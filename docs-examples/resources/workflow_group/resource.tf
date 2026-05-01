# Example for creating a Workflow Group in StackGuardian
resource "stackguardian_workflow_group" "example_workflow_group" {
  resource_name = "example-workflow-group"
  description   = "Example of terraform-provider-stackguardian for Workflow Group"
  tags          = ["example-tag"]
}

# Example for creating a nested Workflow Group
resource "stackguardian_workflow_group" "parent" {
  resource_name = "parent-workflow-group"
  description   = "Parent workflow group"
  tags          = ["parent"]
}

resource "stackguardian_workflow_group" "nested" {
  id            = "${stackguardian_workflow_group.parent.id}/nested-child"
  resource_name = "${stackguardian_workflow_group.parent.id}/nested-child"
  description   = "Nested workflow group under parent"
  tags          = ["nested", "child"]
}