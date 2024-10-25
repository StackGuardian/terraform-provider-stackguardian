data "stackguardian_workflow_group" "example" {
  resource_name = "workflow-group-name"
}

output "workflow-group-output" {
  value = stackguardian_workflow_group.example.description
}
