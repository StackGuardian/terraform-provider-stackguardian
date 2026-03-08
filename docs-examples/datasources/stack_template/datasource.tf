data "stackguardian_stack_template" "example" {
  id = "my-stack-template"
}

output "stack_template_output" {
  value = data.stackguardian_stack_template.example.description
}
