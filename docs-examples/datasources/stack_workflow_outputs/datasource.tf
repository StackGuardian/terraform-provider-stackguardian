data "stackguardian_stack_workflow_outputs" "example" {
  stack          = "stack-name"
  workflow       = "workflow-name"
  workflow_group = "workflow-group-name"
}

output "stack-workflow-output-json" {
  value = data.stackguardian_stack_workflow_outputs.example.data_json
}
