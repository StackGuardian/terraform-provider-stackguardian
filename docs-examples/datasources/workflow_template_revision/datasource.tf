data "stackguardian_workflow_template_revision" "example" {
  id = "my-terraform-template:1"
}

output "workflow_template_revision_output" {
  value = data.stackguardian_workflow_template_revision.example.notes
}
