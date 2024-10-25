data "stackguardian_workflow_outputs" "example" {
  workflow       = "workflow-name"
  workflow_group = "workflow-group-name"
}

output "workflow-output-json" {
  value = data.stackguardian_workflow_outputs.example.data_json
}
