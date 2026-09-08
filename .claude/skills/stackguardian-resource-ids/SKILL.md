---
name: stackguardian-resource-ids
description: Get StackGuardian's path-form resource IDs right — connectors, secrets, template revisions, workflow groups, policies and runner groups. Use when a reference returns 404 or Unauthorized, when writing any attribute that names another resource, when a config mixes bare and organization-qualified template IDs, or when a role's permission paths do not match what was intended.
metadata:
  lifecycle-status: active
---

StackGuardian identifies most things with a **path-like ID**, not a plain name or a UUID. The
leading `/` is part of the value. Getting the form wrong is the most common failure in this
provider, and the error it produces — `Unauthorized`, or a plain `404` — rarely names the real
problem.

## The forms

| Form | Example | Used for |
| --- | --- | --- |
| `/integrations/<name>` | `/integrations/production-aws` | connectors, both cloud and VCS |
| `/secrets/<name>` | `/secrets/db-password` | vault secrets |
| `<name>:<revision>` | `my-template:1` | a template revision in **your own** org |
| `/<org>/<name>:<revision>` | `/stackguardian/aws-s3-demo-website:16` | a template revision in **another** org |
| `/wfgrps/<group>` | `/wfgrps/platform` | policy scope |
| `/policies/<name>:<revision>` | `/policies/aws-all:1` | policies created in your org |
| `/runnergroups/<name>` | `/runnergroups/private-eu` | runner groups |

A bare id resolves against your own organization. Templates published by StackGuardian live under
the `stackguardian` org, so they always take the qualified form.

## Three places this goes wrong

**1. The data source is stricter than the resource.**
`iac_template_id` on `stackguardian_workflow_from_template` accepts both the bare and the qualified
form. The `id` on the `stackguardian_workflow_template_revision` **data source** accepts only the
bare `<name>:<revision>` form — the qualified path returns `Unauthorized`, not "not found", which
sends people looking for a permissions problem that does not exist.

**2. Workflow group paths take no trailing slash, and nesting is literal.**
`enforced_on = ["/wfgrps/platform"]`, not `.../platform/`. A nested group is referenced by its full
path everywhere: `platform/networking`, including in `workflow_group_id` and in role permission
paths.

**3. A template's kind is not interchangeable with another's.**
`/stackguardian/terraform` is a **WORKFLOW_STEP** template. Using it as an `iac_template_id` is a
category error — the form is right and the value is wrong, so it fails only at apply. Step templates
belong in `wf_step_template_id`; IaC templates in `iac_template_id`.

## Role permission paths are not paths

In `allowed_permissions`, the map key is an **API route** and `paths` fills its placeholders with
**bare names**:

```terraform
allowed_permissions = {
  "GET/api/v1/orgs/<org>/wfgrps/<wfGrp>/" = {
    name  = "GetWorkflowGroup"
    paths = {
      "<wfGrp>" = ["frontend"]           # bare name
      # "<wfGrp>" = ["/wfgrps/frontend"] # wrong — this is not a resource path
    }
  }
}
```

Leave `<org>` as the literal placeholder; the platform substitutes it.

## Revisions

`:latest` tracks the most recently published revision. Pin an explicit number when the workflow must
not move on its own — a revision change can alter attributes you never declared. See
`stackguardian-templates`.

## Practice

Prefer a reference to a literal, so Terraform builds the dependency edge and the value cannot drift:

```terraform
integration_id = stackguardian_connector.production_aws.id
workflow_group_id = stackguardian_workflow_group.platform.resource_name
```

Use a literal only for something this configuration does not manage — and say so, because
`terraform validate` cannot tell you whether it exists.

## Output contract

When an ID is involved, state which form the attribute takes, whether the target must already
exist, and — if the config is being written blind — that only a plan against the real organization
can confirm the reference resolves.

Full reference: `docs/guides/ResourceIDs.md`.
