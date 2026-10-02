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

So migrating is not a matter of changing the resource type. The difference shows on a route with
**more than one placeholder**. Given

```terraform
paths = {
  "<wfGrp>" = ["wfgrp1", "wfgrp2"]
  "<wf>"    = ["wf1", "wf2"]
}
```

V3 grants all four combinations (`wfgrp1/wf1`, `wfgrp1/wf2`, `wfgrp2/wf1`, `wfgrp2/wf2`); V4 pairs
the lists by position and grants only `wfgrp1/wf1` and `wfgrp2/wf2`. To keep the V3 behaviour,
combine each list into a single **alternation string**:

```terraform
paths = {
  "<wfGrp>" = ["wfgrp1|wfgrp2"]
  "<wf>"    = ["wf1|wf2"]
}
```

Copying a V3 role unchanged silently narrows it, which is worse than an error because nothing
fails. `docs/resources/rolev4.md` documents the migration.

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
  user_id     = "alice@example.com"
  entity_type = "EMAIL"
  roles       = [stackguardian_rolev4.frontend_developer.resource_name]
}

resource "stackguardian_role_assignment" "frontend_team" {
  user_id     = "sg-test-sso/group-devs"   # <sso-provider-name>/<group>
  entity_type = "GROUP"
  roles       = [stackguardian_rolev4.frontend_developer.resource_name]
}
```

There is no `resource_name`; the assignment is identified (and imported) by `user_id`.
`entity_type` is required: `EMAIL` for a user, `GROUP` for an SSO group. `user_id` is a bare email
for a local user, `<sso-provider-name>/<email>` for an SSO user, or `<sso-provider-name>/<group>`
for an SSO group. The provider name must match the org's SSO configuration exactly — the API does
not validate it, so a mismatch yields an assignment that grants nothing. Prefer `roles` (a list);
`role` is deprecated.

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
