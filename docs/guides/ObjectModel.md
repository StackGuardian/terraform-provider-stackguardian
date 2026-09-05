---
page_title: "Object Model"
subcategory: "Concepts"
description: |-
  How StackGuardian's objects relate, and which Terraform resource manages each one.
---

# Object Model

Most confusion with this provider comes from one distinction: **templates are containers, and
revisions hold the content**. Everything else follows from that.

## The shape of an organization

```
Organization
│
├── Workflow Group ................ stackguardian_workflow_group
│   │                              a folder; nests with "/"
│   ├── Workflow .................. stackguardian_workflow_git
│   │                              or stackguardian_workflow_from_template
│   └── Stack ..................... created from a stack template revision
│
├── Connector ..................... stackguardian_connector
│                                  credentials for a cloud or VCS provider
│
├── Runner Group .................. stackguardian_runner_group
│                                  self-hosted runners + log storage
│
└── Access management
    ├── Role .................... stackguardian_rolev4 (or stackguardian_role)
    ├── Role Assignment ......... stackguardian_role_assignment
    └── Policy .................. stackguardian_policy
```

## Workflow groups

A workflow group is a folder. It organizes workflows and stacks, and it is the thing roles are
usually scoped to — which makes it the main unit of access control as well as of organization.

Groups nest by putting `/` in `resource_name`: `platform/networking` creates `networking` inside
the existing `platform` group. Refer to a nested group by its full path everywhere.

## Templates and revisions

Three template families work identically:

| Container                             | Revision                                       | What the revision holds                          |
| ------------------------------------- | ---------------------------------------------- | ------------------------------------------------ |
| `stackguardian_workflow_template`     | `stackguardian_workflow_template_revision`     | IaC source, inputs, execution settings           |
| `stackguardian_stack_template`        | `stackguardian_stack_template_revision`        | ordered workflows and their configuration        |
| `stackguardian_workflow_step_template`| `stackguardian_workflow_step_template_revision`| the container image or source, and its inputs    |

The container on its own does nothing. Creating a usable template always means **two resources**:
the container, and at least one revision referencing it through `template_id`. Publish the
revision with `is_public` before anything can reference it.

Revisions are versions: to change what a template produces, create a new revision rather than
editing the existing one. Mark superseded revisions with `deprecation` so users are steered to
the replacement.

See [Templates](https://registry.terraform.io/providers/StackGuardian/stackguardian/latest/docs/guides/Templates)
for the full lifecycle.

## Two ways to define a workflow

| | `stackguardian_workflow_git` | `stackguardian_workflow_from_template` |
| --- | --- | --- |
| Where the definition lives | in the resource | in a template revision |
| Good for | one-off workflows, IaC you own | many workflows sharing a definition |
| Inherits template updates | no | yes, for attributes you leave unset |

Both live inside a workflow group and both take a user-chosen `id`.

## What a workflow references

- **A connector**, through `deployment_platform_config.integration_id` — the credentials it
  deploys with. Also through `custom_source.config.auth` when cloning a private repository.
  Connector IDs take the form `/integrations/<resource_name>`; see
  [Resource IDs](https://registry.terraform.io/providers/StackGuardian/stackguardian/latest/docs/guides/ResourceIDs)
  for every ID format the provider uses.
- **A runner group**, through `runner_constraints` — set `type = "private"` and list the group's
  `resource_name` in `names`. With `type = "shared"` it uses StackGuardian's shared runners.

## Stacks versus workflow groups

Both hold multiple workflows, but they mean different things:

- A **workflow group** is a folder of independent workflows with no ordering between them.
- A **stack** is a unit of infrastructure whose workflows run in a defined order, with the
  outputs of one feeding the next. That order lives in the stack template revision's
  `actions.order`.

Reach for a stack when the workflows depend on each other; a workflow group when they don't.

## Access management

A **role** is a set of API permissions, scoped by substituting bare resource names — usually
workflow group names — into the placeholders in each permission. A role does nothing until a
**role assignment** grants it to a user or an SSO group.

A **policy** is separate: it evaluates during workflow runs and can block or flag a run. It is
scoped by `enforced_on`: `["*"]` for the whole organization, or any combination of workflow
groups, workflows and connectors — a different value form again from a role's bare names.

See [Access Management](https://registry.terraform.io/providers/StackGuardian/stackguardian/latest/docs/guides/AccessManagement) for roles, and [Policies](https://registry.terraform.io/providers/StackGuardian/stackguardian/latest/docs/guides/Policies) for policies.

## Reading values back out

Three data sources return values produced by runs, as JSON strings you `jsondecode`:

| Data source                              | Returns                                   |
| ---------------------------------------- | ----------------------------------------- |
| `stackguardian_workflow_outputs`         | outputs of a workflow's latest run        |
| `stackguardian_stack_outputs`            | outputs of a stack's latest run           |
| `stackguardian_stack_workflow_outputs`   | outputs of one workflow inside a stack    |

## Building this with AI

<!-- AI-SKILLS:START -->
Generating configuration from this guide? Load the **`stackguardian-provider`** skill, which turns the guidance here into rules an agent can follow.

**Worth knowing either way:** A workflow group is both a folder and the unit access is scoped to, so the group layout decides the permission model.

The skills live in [the provider repository](https://github.com/StackGuardian/terraform-provider-stackguardian/tree/main/.claude/skills) and work with Claude Code, Cursor, Copilot, Windsurf and any agent that reads [`AGENTS.md`](https://github.com/StackGuardian/terraform-provider-stackguardian/blob/main/AGENTS.md).
<!-- AI-SKILLS:END -->
