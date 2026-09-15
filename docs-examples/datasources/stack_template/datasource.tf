# Read an existing stack template. The template is a container only -- its workflows and
# configuration live in a revision, so use stackguardian_stack_template_revision for those.
data "stackguardian_stack_template" "example" {
  id = "my-stack-template"
}

output "stack_template_description" {
  value = data.stackguardian_stack_template.example.description
}
