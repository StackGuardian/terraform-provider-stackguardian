---
name: stackguardian-access
description: Model access in StackGuardian — roles, permission documents and role assignments. Use when granting someone access to a workflow group, writing allowed_permissions, migrating from stackguardian_role to stackguardian_rolev4, or when a role turns out to grant more or less than intended.
metadata:
  lifecycle-status: active
---

Access is three pieces: a **workflow group** is what access is scoped to, a **role** is a permission
document, and a **role assignment** grants that role to a user or SSO group.

Because a group is the unit of access, the group layout decides the permission model as much as the
org chart does. Design the groups first.

## Use rolev4

`stackguardian_role` is **deprecated**. New configuration uses `stackguardian_rolev4`.

This is not a rename. The two expand permissions differently:

- **`stackguardian_role` (V3)** takes the **cartesian product** of the values in each path
  placeholder — every value combined with every other.
- **`stackguardian_rolev4`** maps them **one to one**.

So migrating is not a matter of changing the resource type. Where V3 relied on the product, V4 needs
the values combined into a single **alternation string**:

```terraform
# V3 behaviour, expressed for V4
paths = { "<wfGrp>" = ["frontend|backend|platform"] }

# NOT this — under V4 this is three separate one-to-one mappings
paths = { "<wfGrp>" = ["frontend", "backend", "platform"] }
```

A naive migration silently narrows a role to its first value, which is worse than an error because
nothing fails. `docs/resources/rolev4.md` documents the migration.

## Permission documents

The key is an **API route**; `name` is the permission's name; `paths` fills the route's
placeholders with **bare names**:

```terraform
resource "stackguardian_rolev4" "frontend_developer" {
  resource_name = "frontend-developer"

  allowed_permissions = {
    "GET/api/v1/orgs/<org>/wfgrps/<wfGrp>/" = {
      name  = "GetWorkflowGroup"
      paths = { "<wfGrp>" = [stackguardian_workflow_group.frontend.resource_name] }
    }
    "POST/api/v1/orgs/<org>/wfgrps/<wfGrp>/wfs/" = {
      name  = "CreateWorkflow"
      paths = { "<wfGrp>" = [stackguardian_workflow_group.frontend.resource_name] }
    }
  }
}
```

Leave `<org>` as the literal placeholder — the platform substitutes it. Values in `paths` are bare
group names, **not** `/wfgrps/...` resource paths. Nested groups use their full path
(`platform/networking`).

## Assignment

```terraform
resource "stackguardian_role_assignment" "alice" {
  resource_name = "alice-frontend-developer"
  user_id       = "alice@example.com"
  role          = stackguardian_rolev4.frontend_developer.resource_name
}
```

`user_id` is an email for a user, or an SSO group identifier to grant the role to everyone in that
group.

## Designing a set of roles

One role per level of access, each scoped to the groups it covers, is easier to reason about than
per-person roles. Centralise the permission sets in `locals` so a change lands in one place:

```terraform
locals {
  developer_permissions = { /* read group, CRU workflows, run, logs, outputs */ }
  elevated_permissions  = { /* + DeleteWorkflow, + ResumeWorkflowRun */ }
}
```

Which permissions belong at which level is a decision about the organisation, not a fact about the
provider. Ask rather than assume — a plausible-looking split can quietly grant teardown rights.

## Output contract

State which groups the role covers, whether V4's one-to-one mapping produces the intended set
(especially when converting from V3), and who ends up with the role. Where permissions were chosen
rather than specified, say they were chosen and on what basis.

Full reference: `docs/guides/AccessManagement.md`, and `docs/guides/TeamOnboarding.md` for a worked
layout.
