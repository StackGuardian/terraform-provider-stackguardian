---
name: stackguardian-workflows
description: Write or debug StackGuardian workflows and workflow groups — stackguardian_workflow_git, stackguardian_workflow_from_template, terraform_config, drift detection, VCS triggers, environment variables and secret references. Use when creating a workflow, wiring it to a connector or runner group, scheduling runs, gating applies, or when a workflow plan shows an unexpected diff.
metadata:
  lifecycle-status: active
---

A workflow runs IaC. It lives inside a workflow group, deploys through a connector, and takes its
code either straight from git or from a library template.

## Pick the resource

- **`stackguardian_workflow_git`** — the IaC is in a repository you control. The source is
  `vcs_config.iac_vcs_config.custom_source`.
- **`stackguardian_workflow_from_template`** — the IaC comes from a published template revision, via
  `iac_template_id`. See `stackguardian-templates`.

`wf_type` decides the engine: `TERRAFORM`, `OPENTOFU`, or `CUSTOM`. A `CUSTOM` workflow runs the
steps in `wf_steps_config` instead of an engine, and templates of other kinds — Helm, Ansible,
Kubectl, CloudFormation — run as `CUSTOM`.

## Groups first

Every workflow needs a group, and the group is a real resource:

```terraform
resource "stackguardian_workflow_group" "platform" {
  resource_name = "platform"
}

resource "stackguardian_workflow_group" "networking" {
  resource_name = "platform/networking"   # nesting is literal; the parent must already exist
}
```

Reference the group rather than naming it, so ordering is inferred:
`workflow_group_id = stackguardian_workflow_group.networking.resource_name`.

## Nesting modes that are easy to get wrong

These are the ones intuition gets backwards. Check the schema rather than guessing:

| Attribute | Shape |
| --- | --- |
| `deployment_platform_config` | **list** of `{kind, config{integration_id, profile_name}}` |
| `environment_variables` | **list** of `{kind, config{...}}` |
| `user_schedules` | **list** |
| `terraform_config`, `runner_constraints`, `vcs_config`, `mini_steps` | **single** object |
| `mini_steps.notifications.email.<event>` | **list** of `{recipients = [...]}` |
| `vcs_triggers.push` and the other trigger events | **map**, keyed `createWfRun` |

```terraform
mini_steps = {
  notifications = {
    email = {
      completed = [{ recipients = ["team@example.com"] }]   # a list
    }
  }
}

vcs_triggers = {
  tracked_branch = "main"
  push = { createWfRun = { enabled = true } }               # a map
  plan_only      = true
}
```

`deployment_platform_config[].kind` must match the connector's own kind — `AWS_RBAC`, `AWS_OIDC`,
`AZURE_STATIC`, and so on.

## terraform_config

```terraform
terraform_config = {
  terraform_version       = "1.5.7"   # bare; no TERRAFORM-/OPENTOFU- prefix
  managed_terraform_state = true
  approval_pre_apply      = true
  drift_check             = true
  drift_cron              = "0 */6 * * ? *"
}
```

`terraform_version` takes the bare form — StackGuardian stores an engine prefix internally and the
provider strips it, so a value read back can be referenced without producing a perpetual diff. A
patch wildcard such as `1.9.x` is accepted. The engine comes from `wf_type`, not from this value.

`approval_pre_apply` decides **whether** a run stops; `approvers` decides **who may release it**. A
pause with no approvers can be released by anyone, so set both. See `stackguardian-policies`.

The five `*Hooks` lists are shell commands run inside the plan or apply container at a lifecycle
point; the `*WfStepsConfig` lists run marketplace step templates, each in its **own** container, so
files do not carry over except through the shared workspace mount. Hooks are skipped on drift runs
unless the matching `run_*_hooks_on_drift` flag is set.

## Secrets

Two mechanisms, and they are not interchangeable:

- A `VAULT_SECRET` environment variable reads `config.secret_id` at run time.
- A **`${secret::<name>}`** reference inside a plain value — note the **double colon** — is resolved
  by StackGuardian at run time. In Terraform write it `$${secret::<name>}`, or Terraform reads it as
  its own interpolation.

Never put a credential in `text_value`: `PLAIN_TEXT` is visible in configuration and state.

## Output contract

Say which resource and `wf_type`, what must exist first (group, connector, template), which values
are references versus literals the org must already have, and — where a gate is involved — whether
runs actually stop and who can release them.

Deeper: `docs/guides/GettingStarted.md`, and the worked examples on the
`stackguardian_workflow_git` registry page.
