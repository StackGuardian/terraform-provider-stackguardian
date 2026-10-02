---
page_title: "Resource IDs"
subcategory: "Concepts"
description: |-
  StackGuardian identifies resources with path-like IDs. This is what each one looks like.
---

# Resource IDs

StackGuardian identifies most resources with a **path-like ID** rather than a plain name or a
UUID — `/integrations/production-aws`, `/stackguardian/aws-s3-demo-website:16`, `/wfgrps/platform/wfs/deploy-vpc`.
The leading `/` is part of the value, not a typo, and these strings appear as ordinary attribute
values throughout the provider.

Not every attribute takes a path, though, and the difference is not guessable. This page is the
reference.

## The forms

| Form | Example | Used for |
| --- | --- | --- |
| `/integrations/<name>` | `/integrations/production-aws` | connectors |
| `/secrets/<name>` | `/secrets/db-password` | secrets |
| `/<org>/<name>:<revision>` | `/stackguardian/aws-s3-demo-website:16` | template revisions, including marketplace |
| `<name>:<revision>` | `my-template:1` | a template revision in your own org |
| `/policies/<name>:<revision>` | `/policies/aws-all:1` | policies created in your own org |
| `/wfgrps/<group>` | `/wfgrps/platform` | policy scope |
| `*` | `*` | policy scope: the whole organization |
| `/wfgrps/<group>/wfs/<workflow>` | `/wfgrps/platform/wfs/deploy-vpc` | how the API addresses a workflow |
| `<name>` | `production-aws` | most `resource_name` values, and most import IDs |
| `<parent>/<child>` | `platform/networking` | a nested workflow group |

## Which attribute takes which

### Path-form IDs

| Attribute | Form |
| --- | --- |
| `deployment_platform_config.config.integration_id` | `/integrations/<connector-name>` |
| `custom_source.config.auth` | `/integrations/<connector-name>` |
| `storage_backend_config.auth.integration_id` (runner group) | `/integrations/<connector-name>` |
| `environment_variables.config.secret_id` | `/secrets/<secret-name>` |
| `wf_steps_config.wf_step_template_id` | `/<org>/<step-template-name>:<revision>` |
| `terraform_config.wf_step_template_revision_id` | `/<org>/<name>:<revision>` |
| `policy_vcs_config.policy_template_id` | `/policies/<name>:<rev>`, or `/<org>/<name>:<rev>` |
| `policy.enforced_on` | `/wfgrps/<group>` (no trailing slash), or `*` for the whole org |

### Bare names, not paths

| Attribute | Value |
| --- | --- |
| `resource_name` on every resource | the name itself |
| `workflow_group_id` on a workflow | the group's name; full path if nested |
| `template_id` on a revision | the parent template's `template_name` |
| `runner_constraints.names` | runner group `resource_name` values |
| `role_assignment.role` / `.roles` | role `resource_name` values |
| `allowed_permissions.*.paths` values | bare resource names — see below |
| `mini_steps.wf_chaining.*` ids | bare workflow / stack / group names |

### Either form

`vcs_config.iac_vcs_config.iac_template_id` accepts both:

- `my-template:1` — your own organization.
- `/stackguardian/aws-s3-demo-website:16` — a template owned by another organization. Any organization can
  publish its own templates, make them public, or share them one-to-one with another org; public
  templates show up for other orgs as external templates. Templates published by StackGuardian
  itself live under the `stackguardian` org.

A bare id is resolved against your own organization.

In place of a revision number you can use `:latest`, which tracks the most recently published
revision — `my-terraform-template:latest`. Pin an explicit revision when the workflow must not move.

## Two places that surprise people

### The template revision data source is stricter than the resource

`iac_template_id` on `stackguardian_workflow_from_template` accepts the fully qualified
`/<org>/<name>:<revision>` form. The `id` on the `stackguardian_workflow_template_revision`
**data source** does not — it takes only the bare `<name>:<revision>` form, and the qualified
path returns `Unauthorized`.

```terraform
# resource: both forms work
iac_template_id = "/stackguardian/aws-s3-demo-website:16"

# data source: bare form only
data "stackguardian_workflow_template_revision" "example" {
  id = "my-template:1"        # correct
  # id = "/my-org/my-template:1"   # Unauthorized
}
```

### Role permission paths are not paths

In `allowed_permissions`, the **key** is an HTTP method concatenated with an API path containing
placeholders, and `paths` maps each placeholder to **bare resource names** — not to
`/wfgrps/…` values:

```terraform
allowed_permissions = {
  "GET/api/v1/orgs/<org>/wfgrps/<wfGrp>/" = {
    name = "GetWorkflowGroup"
    paths = {
      "<wfGrp>" = ["frontend"]        # bare name
      # "<wfGrp>" = ["/wfgrps/frontend"]   # wrong
    }
  },
}
```

`policy.enforced_on` uses a third convention: `["*"]` for the whole organization — on its own,
not combined — or a list mixing workflow groups, workflows and connectors. A workflow group is
`/wfgrps/<group>`, with **no trailing slash**.

## Prefer references over literals

Every path-form ID is derivable from a resource attribute, so let Terraform build it:

```terraform
deployment_platform_config = [{
  kind   = "AWS_RBAC"
  config = { integration_id = stackguardian_connector.aws.id }
}]
```

This keeps the ID correct through renames, and orders resource creation for you. Hard-coding
`"/integrations/aws"` works until someone renames the connector.

## Import IDs are a separate question

Importing does **not** use path-form IDs. Almost every resource imports by bare `resource_name`;
workflows use `<workflow_group_id>/<workflow_id>`. The full table is in
[Importing Existing Resources](https://registry.terraform.io/providers/StackGuardian/stackguardian/latest/docs/guides/ImportingResources).

## Building this with AI

<!-- AI-SKILLS:START -->
Generating configuration from this guide? Load the **`stackguardian-resource-ids`** skill, which turns the guidance here into rules an agent can follow.

**Worth knowing either way:** IDs are position-sensitive: the same value can be valid in one attribute and rejected in another, and the rejection reads as `Unauthorized` rather than a not-found.

The skills live in [the provider repository](https://github.com/StackGuardian/terraform-provider-stackguardian/tree/main/.claude/skills) and work with Claude Code, Cursor, Copilot, Windsurf and any agent that reads [`AGENTS.md`](https://github.com/StackGuardian/terraform-provider-stackguardian/blob/main/AGENTS.md).
<!-- AI-SKILLS:END -->
