# Read a published template revision, then create a workflow from it.
#
# The revision id is `<template-name>:<revision>`. Use the bare form shown here -- the fully
# qualified `/<org>/<template-name>:<revision>` path returns an Unauthorized error.
data "stackguardian_workflow_template_revision" "example" {
  id = "my-terraform-template:1"
}

output "revision_notes" {
  description = "Release notes recorded on this revision."
  value       = data.stackguardian_workflow_template_revision.example.notes
}

output "revision_template_id" {
  description = "The parent workflow template."
  value       = data.stackguardian_workflow_template_revision.example.template_id
}
