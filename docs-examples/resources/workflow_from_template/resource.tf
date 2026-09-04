# The full chain: a template, a published revision of it, and a workflow created from that
# revision. Most configurations only create the workflow and reference a revision someone
# else published -- the first two resources are shown here so the relationship is visible.
resource "stackguardian_workflow_template" "example" {
  template_name      = "my-terraform-template"
  source_config_kind = "TERRAFORM"
  is_public          = "0"
}

resource "stackguardian_workflow_template_revision" "v1" {
  template_id        = stackguardian_workflow_template.example.id
  alias              = "v1"
  source_config_kind = "TERRAFORM"
  is_public          = "1"
  user_job_cpu       = 1
  user_job_memory    = 2048
}

resource "stackguardian_workflow_from_template" "example" {
  workflow_group_id = "my-workflow-group"
  id                = "my-workflow"
  wf_type           = "TERRAFORM"

  vcs_config = {
    iac_vcs_config = {
      # `<template-name>:<revision>`. Anything this workflow does not declare below is
      # inherited from this revision.
      iac_template_id = "${stackguardian_workflow_template.example.template_name}:1"
    }
  }

  # Declared attributes win over the template default; omitted ones are inherited --
  # including after the revision is upgraded.
  terraform_config = {
    terraform_version = "1.5.7"
  }

  depends_on = [stackguardian_workflow_template_revision.v1]
}
