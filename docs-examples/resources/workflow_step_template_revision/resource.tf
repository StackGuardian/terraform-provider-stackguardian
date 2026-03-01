resource "stackguardian_workflow_step_template" "example" {
  template_name = "example-workflow-step-template"
  is_active     = "1"
  is_public     = "0"
  description   = "Example workflow step template"

  source_config_kind = "DOCKER_IMAGE"

  runtime_source = {
    source_config_dest_kind = "CONTAINER_REGISTRY"
    config = {
      docker_image = "ubuntu:latest"
      is_private   = false
    }
  }
}

resource "stackguardian_workflow_step_template_revision" "example" {
  template_id = stackguardian_workflow_step_template.example.id
  alias       = "v1"
  notes       = "Initial revision"

  source_config_kind = "DOCKER_IMAGE"

  runtime_source = {
    source_config_dest_kind = "CONTAINER_REGISTRY"
    config = {
      docker_image = "ubuntu:20.04"
      is_private   = false
    }
  }

  tags = ["terraform", "example"]
}
