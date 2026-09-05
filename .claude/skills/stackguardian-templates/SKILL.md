---
name: stackguardian-templates
description: Work with StackGuardian templates and revisions — workflow templates, workflow step templates, stack templates, and the revision resources under each. Use when publishing a template, pinning or upgrading a revision, wiring a workflow to a template, or when a workflow's settings change unexpectedly after a revision moves.
metadata:
  lifecycle-status: active
---

A template is a container; a **revision** holds the actual content. Every template resource has a
revision twin, and almost everything meaningful lives on the revision.

| Container | Revision |
| --- | --- |
| `stackguardian_workflow_template` | `stackguardian_workflow_template_revision` |
| `stackguardian_workflow_step_template` | `stackguardian_workflow_step_template_revision` |
| `stackguardian_stack_template` | `stackguardian_stack_template_revision` |

Create the container first, then a revision against it. A revision id is `<template>:<n>`, and
revision numbers are assigned by the platform, not chosen.

## The three kinds are not interchangeable

- **IAC templates** back a workflow — `iac_template_id`.
- **WORKFLOW_STEP templates** back a step — `wf_step_template_id`. These are container images.
- **IAC_GROUP / stack templates** back a stack — `template_group_id`.

`/stackguardian/terraform` is a step template, not an IaC one. Putting it in `iac_template_id` is a
category error that passes `terraform validate` and fails at apply. If unsure which kind an id is,
check before using it rather than inferring from the name.

## Pinning, and what a revision change moves

```terraform
iac_template_id = "my-template:3"        # pinned
iac_template_id = "my-template:latest"   # tracks the newest published revision
```

This is the part that surprises people. When a workflow moves to a new revision:

- attributes you **declared** keep your values;
- attributes you **left unset** adopt the new revision's values.

An attribute you never wrote down can therefore change when the revision changes, because it was
always coming from the template. **If a value must not move, declare it explicitly** — and pin the
revision when the workflow must not move at all.

## Referencing across the split

```terraform
resource "stackguardian_workflow_template" "example" { ... }

resource "stackguardian_workflow_template_revision" "v1" {
  template_id = stackguardian_workflow_template.example.template_name
  # ...
}

resource "stackguardian_workflow_from_template" "app" {
  vcs_config = {
    iac_vcs_config = {
      iac_template_id = "${stackguardian_workflow_template.example.template_name}:1"
    }
  }
}
```

Reference the container rather than repeating its name, so Terraform orders creation correctly.

For an id form that belongs to another organization, and the one place the data source is stricter
than the resource, load `stackguardian-resource-ids`.

## Deprecating a revision

```terraform
deprecation = {
  effective_date = "2026-12-31"
  message        = "Superseded by v2, which pins Terraform 1.9."
}
```

`message` is shown to anyone still referencing the revision, so name the replacement in it.

## Stack templates

A stack template revision's `workflows_config` holds several workflows, and `actions.order`
expresses the dependencies between them. That ordering is what distinguishes a stack from a workflow
group, where workflows are independent. Objects nested inside `workflows_config.workflows[]` are
sub-objects of the revision — they are created and destroyed with it and do not collide with
org-level names.

## Output contract

State which template kind is in play, whether the revision is pinned or floating, and — if floating
— which attributes are being left to the template and could therefore move. Where a template must
already exist in the organization, say so; `terraform validate` cannot confirm it.

Full reference: `docs/guides/Templates.md`.
