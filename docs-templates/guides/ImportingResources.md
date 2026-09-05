---
page_title: "Importing Existing Resources"
subcategory: "Operations"
description: |-
  Bring StackGuardian objects that already exist under Terraform management.
---

# Importing Existing Resources

Most StackGuardian resources import by their **bare name** — the value of `resource_name` — not
by an API path or a URL.

## Import ID by resource

| Resource                                       | Import ID                              |
| ---------------------------------------------- | -------------------------------------- |
| `stackguardian_connector`                       | `resource_name`                        |
| `stackguardian_policy`                          | `resource_name`                        |
| `stackguardian_role`                            | `resource_name`                        |
| `stackguardian_rolev4`                          | `resource_name`                        |
| `stackguardian_runner_group`                    | `resource_name`                        |
| `stackguardian_workflow_group`                  | `resource_name` (full path if nested)  |
| `stackguardian_stack_template`                  | `id`                                   |
| `stackguardian_stack_template_revision`         | `id` — `<template-name>:<revision>`    |
| `stackguardian_workflow_template`               | `id`                                   |
| `stackguardian_workflow_template_revision`      | `id` — `<template-name>:<revision>`    |
| `stackguardian_workflow_step_template`          | `id`                                   |
| `stackguardian_workflow_step_template_revision` | `id` — `<template-id>:<revision>`      |
| `stackguardian_role_assignment`                 | `user_id`                              |
| `stackguardian_workflow_git`                    | `<workflow_group_id>/<workflow_id>`    |
| `stackguardian_workflow_from_template`          | `<workflow_group_id>/<workflow_id>`    |

## Using an import block (Terraform 1.5+)

```terraform
import {
  to = stackguardian_workflow_group.frontend
  id = "frontend"
}

resource "stackguardian_workflow_group" "frontend" {
  resource_name = "frontend"
}
```

Run `terraform plan` to see what would be imported before committing to it, then `terraform apply`.

## Using the CLI

```bash
terraform import stackguardian_workflow_group.frontend frontend
```

The resource block must already exist in your configuration.

## Nested workflow groups

Give the full path, since that is the group's identity:

```bash
terraform import stackguardian_workflow_group.networking platform/networking
```

## Workflows

Workflows need both the group and the workflow, separated by `/`:

```bash
terraform import stackguardian_workflow_git.deploy my-workflow-group/my-workflow
```

If the group is itself nested, the group half contains slashes too — everything before the last
`/` is the group, and the final segment is the workflow:

```bash
terraform import stackguardian_workflow_git.deploy platform/networking/deploy-vpc
```

## Workflow groups import themselves without being asked

Creating a `stackguardian_workflow_group` whose name already exists does **not** fail. The
provider reads the existing group and adopts it into state.

That is convenient when re-running a configuration, but it has a consequence worth knowing: you
can end up managing a group you did not create, and a later `terraform destroy` will delete it.
Review the plan before applying when a name might already be taken.

## After importing

Run `terraform plan` immediately. An empty plan means your configuration matches what exists. A
non-empty plan means the two differ — read it carefully rather than applying, because applying
will change the real resource to match your configuration, not the other way round.

Computed attributes are filled from the API on the first refresh, so you do not need to write
them out.

## A worked example

The onboarding examples include import scripts alongside the configurations they match:
[`docs-guides-assets/onboarding`](https://github.com/StackGuardian/terraform-provider-stackguardian/tree/main/docs-guides-assets/onboarding).

## Building this with AI

<!-- AI-SKILLS:START -->
Generating configuration from this guide? Load the **`stackguardian-import`** skill, which turns the guidance here into rules an agent can follow.

**Worth knowing either way:** Run a plan straight after importing. A diff is expected, and it is closed by declaring what the platform reports, not by removing it from state.

The skills live in [the provider repository](https://github.com/StackGuardian/terraform-provider-stackguardian/tree/main/.claude/skills) and work with Claude Code, Cursor, Copilot, Windsurf and any agent that reads [`AGENTS.md`](https://github.com/StackGuardian/terraform-provider-stackguardian/blob/main/AGENTS.md).
<!-- AI-SKILLS:END -->
