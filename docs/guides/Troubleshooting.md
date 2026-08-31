---
page_title: "Troubleshooting"
subcategory: "Operations"
description: |-
  Errors and surprising behaviour you are likely to hit, and what causes them.
---

# Troubleshooting

## Authentication

### Unauthorized, with a key you know is valid

Check `api_uri` against your organization's region. A key issued in one region does not
authenticate against the other:

| Region | `api_uri`                          |
| ------ | ---------------------------------- |
| EU     | `https://api.app.stackguardian.io` |
| US     | `https://api.us.stackguardian.io`  |

Also confirm `org_name` matches the organization the key belongs to.

### Unauthorized when referencing a template revision

Use the bare `<template-name>:<revision>` form for `iac_template_id` and for the
`stackguardian_workflow_template_revision` data source. The fully qualified
`/<org>/<template-name>:<revision>` path returns Unauthorized.

## Plan and apply

### A plan is never empty — the same attributes change every time

Usually an attribute the API normalizes or fills in differently from what you wrote. Compare the
planned value against the real resource in the StackGuardian UI. For a workflow created from a
template, an attribute you never declared is coming from the template revision, so it will move
whenever the revision does — declare it explicitly if it must not.

### `drift_cron` disappears when I disable drift checking

Expected. Setting `terraform_config.drift_check = false` clears `drift_cron`, because a cron is
only meaningful when drift checking is on. The provider predicts this during plan so that plan
and apply agree. Set `drift_check = true` to keep a cron.

### Setting an attribute to `null` does not remove the template's value

For `stackguardian_workflow_from_template`, `null` is indistinguishable from "omitted", and
omitted means *inherit from the template*. To exclude a value rather than inherit it, declare an
empty value of the right type: `""`, `[]` or `{}`.

### Changing `id` replaces the workflow

`id` is user-chosen on `stackguardian_workflow_git` and
`stackguardian_workflow_from_template`, and it identifies the workflow — so changing it destroys
and recreates. Rename with `terraform state mv` if you want to keep the existing workflow.

### Changing `workflow_group_id` also replaces the workflow

There is no move API, so a workflow cannot change groups in place.

## Creating resources

### Creating a workflow group succeeded, but I did not expect it to

If a group with that name already exists, the provider adopts it into state rather than failing.
Terraform now manages a group it did not create, and `terraform destroy` will delete it. Check
`terraform plan` output before applying when a name may already be in use.

### Creating a nested workflow group fails

The parent must exist first. `platform/networking` requires `platform`. Create the parent in the
same configuration and Terraform will order them correctly.

### My workflow disappeared after a failed apply

If VCS trigger registration fails while creating a `stackguardian_workflow_git`, the provider
deletes the workflow it just created rather than leaving one behind with no working triggers.
Fix the underlying problem — usually connector permissions or repository access — and apply again.

## VCS triggers

### Triggers are rejected when I enable them

`vcs_triggers` registers a webhook, so the repository must be reachable: set
`vcs_config.iac_vcs_config.custom_source.config.is_private` to `true` and point `auth` at a
`stackguardian_connector` with access.

Triggers are supported for `GITHUB_COM`, `GITHUB_APP_CUSTOM` and `GITLAB_COM`.

### Webhooks still fire after I removed `vcs_triggers`

Removing the block from your configuration does not unregister webhooks that a previous apply
created. Remove them in your VCS provider or in the StackGuardian UI.

### `file_trigger_patterns` has no effect

It is only evaluated when `file_triggers_enabled = true`, and file-based filtering is only valid
for `GITLAB_COM`.

## Access control

### A user authenticates but has no permissions

For SSO assignments, the provider name must match your organization's SSO configuration exactly.
The API does not validate it, so a mismatch applies cleanly and produces exactly this symptom.

### A role grants more than intended

`stackguardian_role` combines every permission path with every other path. Use
`stackguardian_rolev4`, which maps paths one to one. See
[Access Control](https://registry.terraform.io/providers/StackGuardian/stackguardian/latest/docs/guides/AccessControl).

### A role grants nothing, and the permission path looks right

Check for a quoted interpolation. A path value written as
`"resource.stackguardian_workflow_group.x.resource_name"` is a literal string, not a reference —
valid HCL that silently points at a group that does not exist.

## Importing

An import that fails with a not-found error is usually the wrong ID format. Most resources import
by bare `resource_name`, not by an API path; workflows need `<workflow_group_id>/<workflow_id>`.
The full table is in
[Importing Existing Resources](https://registry.terraform.io/providers/StackGuardian/stackguardian/latest/docs/guides/ImportingResources).

## Getting more detail

Set `TF_LOG=DEBUG` to see the provider's own logging, including the API calls it makes:

```bash
TF_LOG=DEBUG terraform apply 2> terraform-debug.log
```

The log contains your API key. Do not attach it to a ticket without redacting.

If the behaviour still looks wrong, open an issue at
[github.com/StackGuardian/terraform-provider-stackguardian/issues](https://github.com/StackGuardian/terraform-provider-stackguardian/issues)
with the resource, the plan output, and the provider version.
