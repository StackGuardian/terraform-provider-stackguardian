resource "stackguardian_workflow_step_template" "example" {
  template_name = "example-workflow-step-template"
  is_active     = "1"
  is_public     = "0"
  description   = "Example workflow step template with runtime source"

  source_config_kind = "DOCKER_IMAGE"

  runtime_source = {
    source_config_dest_kind = "CONTAINER_REGISTRY"
    config = {
      docker_image = "ubuntu:latest"
      is_private   = false
    }
  }

  tags = ["terraform", "example"]
}
