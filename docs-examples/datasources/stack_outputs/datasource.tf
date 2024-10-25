data "stackguardian_stack_outputs" "example-stack-outputs" {
  stack          = "stack-name"
  workflow_group = "workflow-group-name"
}

output "stack-outputs-output" {
  value = data.stackguardian_stack_outputs.example-stack-outputs.data_json
}


