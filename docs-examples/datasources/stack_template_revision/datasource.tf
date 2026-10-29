# Read a specific revision of a stack template, including the workflows it deploys and the
# order they run in. The revision id is `<template-id>:<revision>`.
data "stackguardian_stack_template_revision" "example" {
  id = "my-stack-template:1"
}

output "revision_notes" {
  value = data.stackguardian_stack_template_revision.example.notes
}
