data "stackguardian_workflow_step_template_revision" "example" {
  id = "12345678901234567890:1"
}

output "workflow_step_template_revision_info" {
  value = {
    template_id = data.stackguardian_workflow_step_template_revision.example.template_id
    alias       = data.stackguardian_workflow_step_template_revision.example.alias
    notes       = data.stackguardian_workflow_step_template_revision.example.notes
    tags        = data.stackguardian_workflow_step_template_revision.example.tags
  }
}
