data "stackguardian_workflow_step_template" "example" {
  id = "12345678901234567890"
}

output "workflow_step_template_info" {
  value = {
    name        = data.stackguardian_workflow_step_template.example.template_name
    type        = data.stackguardian_workflow_step_template.example.template_type
    description = data.stackguardian_workflow_step_template.example.description
    tags        = data.stackguardian_workflow_step_template.example.tags
  }
}
