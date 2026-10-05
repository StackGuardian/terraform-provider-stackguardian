# Read an existing stack -- for example to see which stack template revision it runs and the
# workflows it contains, without managing the stack in this configuration.
#
# A stack is identified by its own id plus the workflow group that contains it. For a nested
# group, give the full path.
data "stackguardian_stack" "example" {
  workflow_group_id = "my-workflow-group"
  id                = "my-stack"
}

output "stack_template_revision" {
  value = data.stackguardian_stack.example.template_group_id
}

output "stack_workflow_ids" {
  description = "Resource ids of the workflows the stack created."
  value       = [for wf in data.stackguardian_stack.example.workflows_config.workflows : wf.workflow_id]
}
