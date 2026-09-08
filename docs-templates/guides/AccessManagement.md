---
page_title: "Access Management"
subcategory: "Concepts"
description: |-
  Roles and role assignments — which resource does what, and how they compose.
---

# Access Management

Three resources are involved, and they do different jobs:

| Resource                            | Job                                                          |
| ----------------------------------- | ------------------------------------------------------------ |
| `stackguardian_rolev4`              | A named set of permissions over resource paths.               |
| `stackguardian_role`                | The V3 form of the same thing. **Deprecated** — see below.    |
| `stackguardian_role_assignment`     | Grants roles to a user or an SSO group.                       |

Roles are about people. A role controls what someone may do in StackGuardian; whether a given
run is allowed to proceed is a separate question, answered by a policy — see
[Policies](https://registry.terraform.io/providers/StackGuardian/stackguardian/latest/docs/guides/Policies).

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

~> The values under `paths` are **bare resource names**, not `/wfgrps/...` paths — unlike a
policy's `enforced_on`, which does take paths. See
[Resource IDs](https://registry.terraform.io/providers/StackGuardian/stackguardian/latest/docs/guides/ResourceIDs)
and [Policies](https://registry.terraform.io/providers/StackGuardian/stackguardian/latest/docs/guides/Policies).

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

## A practical pattern

Scope by workflow group, and let the group hierarchy carry the structure:

1. One workflow group per team or environment, nested where that helps.
2. One role per level of access, scoped to those groups.
3. Role assignments mapping people or SSO groups to roles.
4. [Policies](https://registry.terraform.io/providers/StackGuardian/stackguardian/latest/docs/guides/Policies) enforced on the same groups, for rules that must hold at run time.

## Building this with AI

<!-- AI-SKILLS:START -->
Generating configuration from this guide? Load the **`stackguardian-access`** skill, which turns the guidance here into rules an agent can follow.

**Worth knowing either way:** Permission map keys are API routes, and `paths` fills their placeholders with bare names — not `/wfgrps/...` resource paths.

The skills live in [the provider repository](https://github.com/StackGuardian/terraform-provider-stackguardian/tree/main/.claude/skills) and work with Claude Code, Cursor, Copilot, Windsurf and any agent that reads [`AGENTS.md`](https://github.com/StackGuardian/terraform-provider-stackguardian/blob/main/AGENTS.md).
<!-- AI-SKILLS:END -->
