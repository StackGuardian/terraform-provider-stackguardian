# A workflow template is looked up by its ID, which is the template name.
# The template itself is a container -- use stackguardian_workflow_template_revision
# to read the configuration that workflows actually inherit.
data "stackguardian_workflow_template" "example" {
  id = "my-terraform-template"
}

output "workflow_template_output" {
  value = data.stackguardian_workflow_template.example.description
}
