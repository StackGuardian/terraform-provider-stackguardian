---
name: stackguardian-provider
description: Write, review, or debug Terraform configuration for the StackGuardian provider — workflows, workflow groups, stacks, templates and revisions, connectors, policies, roles and runner groups. Use when a config declares `stackguardian_*` resources, when the provider returns 401/403/404/409, when a plan shows a permanent diff, or when choosing which StackGuardian resource fits a task. Routes to the specific skill for the area in play.
metadata:
  lifecycle-status: active
---

StackGuardian orchestrates infrastructure-as-code runs. This provider manages the orchestration —
the workflows, the templates they come from, the credentials they use, and the guardrails around
them — not the cloud resources those workflows deploy.

## Diagnose first, then route

Identify what is actually in play before writing HCL. Most failures here are one of five kinds:

| Symptom | Likely cause | Load |
| --- | --- | --- |
| `404`, `Unauthorized`, "not found" on a reference | Wrong ID form for the position it is used in | `stackguardian-resource-ids` |
| `409 already exists` | Name is taken — a real resource, or one a failed run left behind | `stackguardian-import` |
| Permanent diff after apply | An attribute the platform computes is being written, or one it returns is omitted | the area skill, then `docs/guides/Troubleshooting.md` |
| Permissions apply to more than intended | `stackguardian_role` expands paths as a cartesian product | `stackguardian-access` |
| A run is not gated, or is gated by the wrong people | `approval_pre_apply` and `approvers` do different jobs | `stackguardian-policies` |

Area skills: `stackguardian-workflows`, `stackguardian-templates`, `stackguardian-policies`,
`stackguardian-access`, `stackguardian-import`, `stackguardian-resource-ids`.

## Provider configuration

```terraform
provider "stackguardian" {
  api_key  = "<key>"                            # or STACKGUARDIAN_API_KEY
  org_name = "<org>"                            # or STACKGUARDIAN_ORG_NAME
  api_uri  = "https://api.app.stackguardian.io" # or STACKGUARDIAN_API_URI
}
```

Prefer the environment variables; every argument falls back to one. Use
`https://api.us.stackguardian.io` for the US region. A `401` almost always means the key or the
region is wrong, not the configuration.

## The object model, briefly

A **workflow group** is a folder *and* the unit access is scoped to, so the group layout decides the
permission model. Groups nest with `/`, and a nested group is a resource of its own — creating
`platform/networking` does not create `platform`.

A **workflow** lives inside a group and runs IaC either straight from git
(`stackguardian_workflow_git`) or from a library template
(`stackguardian_workflow_from_template`). It deploys through a **connector**, may be pinned to a
**runner group**, and is governed by **policies**. **Roles** grant people access to groups.

Full picture: `docs/guides/ObjectModel.md`.

## Choosing the resource

- IaC lives in a git repo you control → `stackguardian_workflow_git`
- IaC comes from a published template → `stackguardian_workflow_from_template`
- Several workflows deployed together with ordering → a stack, via
  `stackguardian_stack_template` + `stackguardian_stack_template_revision`
- Granting access → `stackguardian_rolev4` plus `stackguardian_role_assignment`.
  **Not** `stackguardian_role`, which is deprecated.

There are 15 resources and 17 data sources; `internal/provider/provider.go` is the authoritative
list. A type not in it does not exist, however plausible the name looks.

## Output contract

When proposing configuration, state:

1. **Which resources** and why those rather than the alternatives.
2. **Ordering and dependencies** — what must exist first, and whether Terraform infers it or needs
   an explicit reference.
3. **What is not verifiable statically** — a connector, secret, or template ID that must already
   exist in the organization. Say so rather than implying the config is self-contained.
4. **Validation**: `terraform validate` catches schema errors, but not whether a referenced
   connector or template exists. Only a plan against the real org does that.

## Enforce quality

- Never invent a resource type, attribute, or enum value. Check the schema.
- Reference resources rather than hardcoding IDs: `stackguardian_connector.aws.id` over
  `"/integrations/aws"`, so Terraform orders the graph.
- Keep credentials out of `text_value` — use a `VAULT_SECRET` variable or a `${secret::<name>}`
  reference.
- Pin template revisions explicitly when a workflow must not move.
