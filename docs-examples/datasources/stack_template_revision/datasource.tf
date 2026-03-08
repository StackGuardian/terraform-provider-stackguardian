data "stackguardian_stack_template_revision" "example" {
  id = "my-stack-template:1"
}

output "stack_template_revision_output" {
  value = data.stackguardian_stack_template_revision.example.notes
}
