data "stackguardian_workflow_template" "example" {
  id = "my-terraform-template"
}

output "workflow_template_output" {
  value = data.stackguardian_workflow_template.example.description
}
