---
name: stackguardian-import
description: Bring existing StackGuardian resources under Terraform management, and resolve 409 "already exists" errors. Use when adopting resources created in the UI, when a create fails because the name is taken, when writing import blocks or terraform import commands, or when deciding whether a resource is safe to destroy.
metadata:
  lifecycle-status: active
---

Two situations lead here: deliberately adopting something created outside Terraform, and a create
that failed with `409 already exists`.

## 409 already exists

The name is taken. Three causes, and they need different responses:

1. **A real resource exists** that should be managed — import it (below).
2. **A previous run left it behind.** Common in test and CI organizations where a run died between
   create and destroy. Delete the leftover, or import and destroy it.
3. **The provider adopted it silently.** `stackguardian_workflow_group` does *not* fail when a group
   of the same name exists — it **adopts** the group into state. Convenient on a re-run, but it
   means Terraform can end up managing a group it did not create, and a later `terraform destroy`
   will delete it. Review the plan when the name might already be taken.

## Importing

Terraform 1.5 and later, declarative and reviewable in a plan:

```terraform
import {
  to = stackguardian_workflow_group.example
  id = "platform"
}
```

Or the CLI, for any version:

```bash
terraform import stackguardian_workflow_group.example platform
```

The resource must already be declared in configuration — an `import` block whose `to` names a
resource that is not in the config is an error.

## Import ID formats

The import id is generally the resource's own identifier, not a full API path:

| Resource | Import id |
| --- | --- |
| `stackguardian_workflow_group` | the group name, nested as `parent/child` |
| `stackguardian_workflow_git`, `stackguardian_workflow_from_template` | see the resource page — a workflow is identified within its group |
| `stackguardian_connector` | the connector name |
| `stackguardian_policy` | the policy name |
| `stackguardian_rolev4`, `stackguardian_role` | the role name |
| `stackguardian_runner_group` | the runner group name |
| templates and revisions | the template name, and `<template>:<revision>` for a revision |

Each resource page on the registry carries a worked `import` block and CLI line under **Import** —
use that rather than inferring the shape, since a few resources need the group as well as the name.

## After importing

Run a plan immediately. The usual outcome is a diff, because the configuration you wrote does not
yet match what the platform holds. Close it by writing the attributes the platform reports, not by
removing them from state.

Two recurring causes of a diff that will not settle:

- An attribute the platform **computes and returns** but the configuration omits. Declare it.
- A value written in a form the platform normalises. `terraform_version` is the example: the
  platform stores an engine prefix (`TERRAFORM-`) which the provider strips, so write the bare
  `1.5.7`.

## Before destroying

`terraform destroy` on an imported resource deletes it for real. Two things to check first:

- Was the resource **adopted** rather than created by this configuration? Deleting a workflow group
  takes its workflows with it.
- Does anything else reference it — a policy scoped to the group, a role granting access to it, a
  workflow chained to it? Those references do not block the delete.

## Output contract

State what will be imported and its id form, that a plan should be run before anything else, and —
where the diff is expected — which attributes are likely to appear and why. If a destroy is in
scope, say explicitly what else goes with it.

Full reference: `docs/guides/ImportingResources.md`, and `docs/guides/Troubleshooting.md` for
post-import diffs.
