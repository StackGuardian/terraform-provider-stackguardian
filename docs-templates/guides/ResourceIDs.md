---
page_title: "Resource IDs"
subcategory: "Concepts"
description: |-
  What the `id` attribute holds, which attributes take path-like IDs such as /integrations/<id>, and how to build them.
---

# Resource IDs

Two things are easy to conflate:

- The **`id` attribute** every resource exposes. It is a **bare slug** — `production-aws`,
  `platform`, `my-template` — never a path.
- The **path-form IDs** the platform expects in *references*: `/integrations/production-aws`,
  `/wfgrps/platform`, `/stackguardian/aws-s3-demo-website:16`. The leading `/` is part of the value.

A reference is the `id` plus the prefix the attribute expects. Let Terraform build it:

```terraform
integration_id = "/integrations/${stackguardian_connector.aws.id}"
enforced_on    = ["/wfgrps/${stackguardian_workflow_group.platform.id}"]
```

This keeps the reference correct through renames and orders resource creation for you.

## What `id` holds

- Set `id` yourself and it is stored verbatim.
- Leave it out and the platform derives it from `resource_name`: unchanged when the name is
  already slug-shaped (letters, digits, `_`, `-`); otherwise lowercased, spaces replaced by `-`,
  **with a random suffix appended** — `"AWS Prod Connector"` becomes something like
  `aws-prod-connector-x7k2lm9q4n1d`.
- Workflow groups always have `id == resource_name`, nested ones included: `platform/networking`.
- The `GITHUB_COM` connector is a per-organization singleton whose `id` is always `github_com`,
  whatever its `resource_name`.

So `resource_name` is only a safe stand-in for `id` when you know the name is slug-shaped.
Reference `id`.

## The forms

| Form | Example | Used for |
| --- | --- | --- |
| `<slug>` | `production-aws` | the `id` attribute of every resource; most import IDs |
| `<parent>/<child>` | `platform/networking` | the `id` of a nested workflow group |
| `/integrations/<id>` | `/integrations/production-aws` | referencing a connector |
| `/secrets/<name>` | `/secrets/db-password` | referencing a secret |
| `/wfgrps/<id>` | `/wfgrps/platform` | policy scope |
| `*` | `*` | policy scope: the whole organization |
| `<name>:<revision>` | `my-template:1` | a template revision in your own org; also the `id` of a revision resource |
| `/<org>/<name>:<revision>` | `/stackguardian/aws-s3-demo-website:16` | a template revision owned by another org, including marketplace |
| `/policies/<name>:<revision>` | `/policies/aws-all:1` | a policy template created in your own org |
| `/wfgrps/<group>/wfs/<workflow>` | `/wfgrps/platform/wfs/deploy-vpc` | how the API addresses a workflow; not used by any provider attribute |

## Which attribute takes which

### Path-form — build with the prefix

| Attribute | Form |
| --- | --- |
| `deployment_platform_config.config.integration_id` | `"/integrations/${stackguardian_connector.x.id}"` |
| `storage_backend_config.auth.integration_id` (runner group) | `"/integrations/${stackguardian_connector.x.id}"` |
| `custom_source.config.auth` | `"/integrations/${stackguardian_connector.x.id}"` for `GITHUB_COM`, `GITHUB_APP_CUSTOM`, `GITLAB_COM`, `BITBUCKET_ORG`, `AZURE_DEVOPS*`; `/secrets/<secret-name>` for any source, and the **only** form `GIT_OTHER` accepts |
| `environment_variables.config.secret_id` | `/secrets/<secret-name>` |
| `wf_steps_config.wf_step_template_id` | `/<org>/<step-template-name>:<revision>` |
| `terraform_config.wf_step_template_revision_id` | `/<org>/<name>:<revision>` |
| `policy_vcs_config.policy_template_id` | `/policies/<name>:<rev>`, or `/<org>/<name>:<rev>` |
| `policy.enforced_on` | `"/wfgrps/${stackguardian_workflow_group.x.id}"` (no trailing slash), or `*` for the whole org |

### Bare `id`, no prefix

| Attribute | Value |
| --- | --- |
| `workflow_group_id` on a workflow | `stackguardian_workflow_group.x.id` — the full path when nested |
| `template_id` on a revision | the parent template's `template_name`, which is its `id` |
| `template_id` / `iac_template_id` inside a stack template revision | the bare `template_name` of a template in your own org; the provider adds `/<org>/` |
| `runner_group_id` on the token data source | the runner group's `id`; the provider adds `/runnergroups/` |
| `runner_constraints.names` | runner group `resource_name` values |
| `role_assignment.role` / `.roles` | role `resource_name` values |
| `allowed_permissions.*.paths` values | bare resource names — see below |
| `mini_steps.wf_chaining.*` ids | bare workflow / stack / group names |

### Either form

`vcs_config.iac_vcs_config.iac_template_id` on `stackguardian_workflow_from_template` accepts both:

- `my-template:1` — your own organization.
- `/stackguardian/aws-s3-demo-website:16` — a template owned by another organization. Any organization can
  publish its own templates, make them public, or share them one-to-one with another org; public
  templates show up for other orgs as external templates. Templates published by StackGuardian
  itself live under the `stackguardian` org.

A bare id is resolved against your own organization.

In place of a revision number you can use `:latest`, which tracks the most recently published
revision — `my-terraform-template:latest`. Pin an explicit revision when the workflow must not move.

## Where it goes wrong

### A bare connector id applies, then every run fails

`integration_id = stackguardian_connector.aws.id` passes `terraform apply` — the API stores the
string exactly as given — and every run then fails in its preparation step (`pre_0_step`) with
`Connector aws does not exist`, because the platform looks connectors up by `/integrations/<id>`.
Runner groups are stricter: the same mistake in `storage_backend_config.auth.integration_id` is
rejected on apply with `Invalid integration id used in auth for storage backend config`.

### `auth` is checked on apply

`custom_source.config.auth` must start with `/secrets/` or `/integrations/`; anything else is a
validation error. With `source_config_dest_kind = "GIT_OTHER"` only the `/secrets/` form is
allowed — `only secrets supported for GIT_OTHER`.

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
`/wfgrps/<id>`, with **no trailing slash**.

## Import IDs are a separate question

Importing does **not** use path-form IDs. Almost every resource imports by bare `resource_name`;
workflows use `<workflow_group_id>/<workflow_id>`. The full table is in
[Importing Existing Resources](https://registry.terraform.io/providers/StackGuardian/stackguardian/latest/docs/guides/ImportingResources).
