---
page_title: "Access Control"
subcategory: "Concepts"
description: |-
  Roles, role assignments and policies — which resource does what, and how they compose.
---

# Access Control

Four resources are involved, and they do different jobs:

| Resource                            | Job                                                          |
| ----------------------------------- | ------------------------------------------------------------ |
| `stackguardian_rolev4`              | A named set of permissions over resource paths.               |
| `stackguardian_role`                | The V3 form of the same thing. **Deprecated** — see below.    |
| `stackguardian_role_assignment`     | Grants roles to a user or an SSO group.                       |
| `stackguardian_policy`              | A guardrail evaluated during workflow runs.                   |

Roles and policies are independent. A role controls what a person may do in StackGuardian; a
policy controls whether a workflow run is allowed to proceed.

## Use `stackguardian_rolev4`

`stackguardian_role` expands permissions by combining **every path with every other path**.
`stackguardian_rolev4` maps paths **one to one**, which is what most configurations intend. Prefer
`rolev4` for anything new; `stackguardian_role` is deprecated and Terraform will warn at plan
time when you use it.

The [`stackguardian_rolev4` page](https://registry.terraform.io/providers/StackGuardian/stackguardian/latest/docs/resources/rolev4)
carries a side-by-side migration example.

## Scoping a role

A permission's key is an HTTP method joined to an API path containing placeholders. The values
you substitute into those placeholders are what scope the role — usually workflow group names,
which is how a role ends up scoped to part of the organization:

```terraform
resource "stackguardian_workflow_group" "frontend" {
  resource_name = "frontend"
}

resource "stackguardian_rolev4" "frontend_developer" {
  resource_name = "frontend-developer"
  description   = "Read access to the frontend workflow group"

  allowed_permissions = {
    "GET/api/v1/orgs/<org>/wfgrps/<wfGrp>/" = {
      name = "GetWorkflowGroup"
      paths = {
        "<wfGrp>" = [
          resource.stackguardian_workflow_group.frontend.resource_name,
        ]
      }
    },
  }
}
```

Reference the workflow group rather than typing its name, so the role follows any rename.

~> The values under `paths` are **bare resource names**, not `/wfgrps/...` paths — unlike
`policy.enforced_on`, which does take paths. See
[Resource IDs](https://registry.terraform.io/providers/StackGuardian/stackguardian/latest/docs/guides/ResourceIDs).

~> Watch the quoting. A path value written as `"resource.stackguardian_workflow_group.x.resource_name"`
is a literal string, not a reference — it applies cleanly and silently grants access to a group
that does not exist. Leave interpolated references unquoted.

## Granting a role

A role does nothing until it is assigned.

```terraform
resource "stackguardian_role_assignment" "frontend_dev" {
  user_id     = "developer@example.com"
  entity_type = "EMAIL"
  role        = stackguardian_rolev4.frontend_developer.resource_name
}
```

`role` and `roles` take a role's `resource_name`.

### SSO assignments

For SSO users and groups the provider name must match your organization's SSO configuration
**exactly**. The API does not validate it at apply time, so a mismatch applies successfully and
then produces a user with no permissions — a failure that looks like a login problem rather than
a configuration one. Check the spelling against your SSO setup before applying.

## Policies

A policy evaluates during workflow runs. `enforced_on` controls where it applies — either the
whole organization, or any combination of workflow groups, workflows and connectors:

- `["*"]` — organization-wide. Used on its own, not combined with other entries.
- `["/wfgrps/frontend"]` — one workflow group and everything in it. **No trailing slash.**
- Workflows and connectors follow the same resource-path convention and can be listed alongside
  workflow groups.

~> This is a third convention: not the bare names a role's `paths` uses, and not quite the same as
the IDs used elsewhere. Confirm an unfamiliar form against an existing policy before relying on it.

```terraform
resource "stackguardian_policy" "require_tags" {
  resource_name = "require-tags"
  policy_type   = "GENERAL"

  # "*" would enforce this organization-wide.
  enforced_on = ["/wfgrps/frontend"]

  policies_config = [{
    name    = "require-tags"
    on_fail = "FAIL"

    policy_input_data = {
      schema_type = "FORM_JSONSCHEMA"
      data        = jsonencode({})
    }
  }]
}
```

`on_fail` decides what happens when evaluation fails. The policy body can be written inline with
`policy_input_data`, or pulled from version control or the marketplace with `policy_vcs_config`.

## A practical pattern

Scope by workflow group, and let the group hierarchy carry the structure:

1. One workflow group per team or environment, nested where that helps.
2. One role per level of access, scoped to those groups.
3. Role assignments mapping people or SSO groups to roles.
4. Policies enforced on the same groups, for rules that must hold at run time.
