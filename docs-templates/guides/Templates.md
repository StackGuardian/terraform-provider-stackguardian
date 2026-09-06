---
page_title: "Templates and Revisions"
subcategory: "Concepts"
description: |-
  The template/revision lifecycle, and how a workflow inherits values from a revision.
---

# Templates and Revisions

Use a template when several workflows should share one definition and pick up updates to it. If
you only need one workflow from a repository you own, `stackguardian_workflow_git` is simpler —
see [Object Model](https://registry.terraform.io/providers/StackGuardian/stackguardian/latest/docs/guides/ObjectModel#two-ways-to-define-a-workflow).

## The two resources

A template is a container. It has a name and little else — the content lives in a revision.

```terraform
resource "stackguardian_workflow_template" "example" {
  template_name      = "my-terraform-template"
  source_config_kind = "TERRAFORM"
  is_public          = "0"
}

resource "stackguardian_workflow_template_revision" "v1" {
  template_id        = stackguardian_workflow_template.example.id
  alias              = "v1"
  source_config_kind = "TERRAFORM"
  is_public          = "1" # publish, so workflows can reference it
  user_job_cpu       = 1
  user_job_memory    = 2048
}
```

`is_public = "1"` publishes the revision. Until then nothing can reference it.

~> **`is_public` means two different things.** On a **revision** it controls whether the revision
is published and usable. On the **template** it controls whether the template is shared with other
organizations. Same attribute name, different effect depending on which resource it is set on.

## Creating a workflow from a revision

```terraform
resource "stackguardian_workflow_from_template" "example" {
  workflow_group_id = "my-workflow-group"
  id                = "my-workflow"
  wf_type           = "TERRAFORM"

  vcs_config = {
    iac_vcs_config = {
      iac_template_id = "my-terraform-template:1"
    }
  }

  terraform_config = {
    terraform_version = "1.5.7"
  }
}
```

`iac_template_id` accepts two forms:

| Form | Use it for |
| --- | --- |
| `<template-name>:<revision>` | a template in your own organization |
| `/<org>/<template-name>:<revision>` | a template owned by another organization — one shared with you directly, or published publicly. Templates published by StackGuardian use the `stackguardian` org (`/stackguardian/aws-s3-demo-website:16`) |

A bare id is resolved against your own organization.

### Tracking the latest revision

Use `:latest` instead of a number to follow the most recently published revision:

```terraform
iac_template_id = "my-terraform-template:latest"
```

Convenient for templates you control; pin an explicit revision when the workflow must not change
underneath you.

~> The `stackguardian_workflow_template_revision` **data source** is stricter: its `id` takes only
the bare `<template-name>:<revision>` form, and the fully qualified path returns Unauthorized. See
[Resource IDs](https://registry.terraform.io/providers/StackGuardian/stackguardian/latest/docs/guides/ResourceIDs).

## How inheritance works

The provider fetches the revision, merges it with what you declared, and stores the
fully-resolved result in state. Resolution is **per attribute**:

- **Omit an attribute** → it is inherited from the revision.
- **Declare an attribute** → your value wins, and the revision's default for it is ignored.

In the example above, `terraform_version` is pinned by the workflow; everything else comes from
revision `1`.

### Excluding a value the template sets

`null` will not do this. Terraform cannot distinguish `null` from "omitted", and omitted means
*inherit*. To exclude a value rather than inherit it, declare an **empty value of the right
type**: `""` for a string, `[]` for a list, `{}` for a map or object.

## Upgrading a revision

Publish a new revision, then change the reference:

```terraform
iac_template_id = "my-terraform-template:2"
```

On the next apply:

- attributes you **declared** keep your values,
- attributes you **left unset** adopt revision `2`'s values.

That second point is the one to watch. An attribute you never wrote down can change when the
revision changes, because it was always coming from the template. If a value must not move,
declare it explicitly.

## Deprecating a revision

```terraform
resource "stackguardian_workflow_template_revision" "v1" {
  # ...
  deprecation = {
    effective_date = "2026-12-31"
    message        = "Superseded by v2, which pins Terraform 1.9."
  }
}
```

`deprecation.message` is shown to users still referencing the revision, so name the replacement.

## Stack templates

Stack templates follow the same container/revision split, with one addition: a stack template
revision's `workflows_config` holds several workflows and `actions.order` expresses the
dependencies between them. That ordering is what distinguishes a stack from a workflow group,
where workflows are independent.

## Reading a revision before using it

To inspect a revision's defaults — for example to see what a workflow would inherit:

```terraform
data "stackguardian_workflow_template_revision" "example" {
  id = "my-terraform-template:1"
}
```
