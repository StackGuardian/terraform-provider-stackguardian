data "stackguardian_workflow_template" "example" {
  template_name = "my-terraform-template"
}

output "workflow_template_output" {
  value = data.stackguardian_workflow_template.example.description
}
