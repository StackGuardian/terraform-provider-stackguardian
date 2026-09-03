---
page_title: "Policies"
subcategory: "Concepts"
description: |-
  Guardrails evaluated during a run — where a policy applies, where its body comes from, and how to gate a run on approval.
---

# Policies

A policy is a guardrail. It evaluates during workflow and stack runs and can let the run
proceed, flag it, block it, or hold it for human approval.

Policies are independent of roles. A role controls what a person may do in StackGuardian; a
policy controls whether a run is allowed to proceed, regardless of who started it. See
[Access Management](https://registry.terraform.io/providers/StackGuardian/stackguardian/latest/docs/guides/AccessManagement)
for roles and assignments.

## Where a policy applies

`enforced_on` scopes the policy — either the whole organization, or any combination of workflow
groups, workflows and connectors:

- `["*"]` — organization-wide. Used on its own, not combined with other entries.
- `["/wfgrps/frontend"]` — one workflow group and everything in it. **No trailing slash.**
- Workflows and connectors follow the same resource-path convention and can be listed alongside
  workflow groups.

~> This is a third value convention: not the bare names a role's `paths` uses, and not quite the
same as the IDs used elsewhere. Confirm an unfamiliar form against an existing policy before
relying on it. See
[Resource IDs](https://registry.terraform.io/providers/StackGuardian/stackguardian/latest/docs/guides/ResourceIDs).

`enforced_on`, `approvers` and `number_of_approvals_required` apply only to
`policy_type = "GENERAL"`. A `FILTER.INSIGHT` policy filters Insight findings rather than gating
a run, so it takes neither scope nor approval settings.

## Where the policy body comes from

`policy_vcs_config` says where the body comes from, and `policy_input_data` carries it when the
policy is written inline. For an inline policy you set **both** — they are not alternatives.

| Source | What to set |
| ------ | ----------- |
| Written inline | `policy_input_data`, plus a `custom_source` of `SG_POLICY_FRAMEWORK` / `INLINE` |
| Marketplace | `use_marketplace_template = true` and `policy_template_id` |
| Your own repository | `custom_source` set to the policy kind (`OPA_REGO`, `SG_POLICY_FRAMEWORK`) with its `config` pointing at the repo |

~> Omitting `policy_vcs_config` is valid — the backend applies the inline configuration at
runtime, and a policy read back from the platform always carries it. The provider does not fill
it in for you, so writing it out explicitly is what keeps configuration and state describing the
same thing.

An inline policy body uses `schema_type = "TIRITH_JSON"`. That is what the platform stores and
returns for a `SG_POLICY_FRAMEWORK` policy, so using `RAW_JSON` instead leaves a permanent diff
between your configuration and state.

```terraform
resource "stackguardian_workflow_group" "frontend" {
  resource_name = "frontend"
}

resource "stackguardian_policy" "require_tags" {
  resource_name = "require-tags"
  description   = "EC2 instances must carry an Environment tag"
  policy_type   = "GENERAL"

  # "*" would enforce this organization-wide.
  enforced_on = ["/wfgrps/${stackguardian_workflow_group.frontend.resource_name}"]

  policies_config = [{
    name    = "require-tags"
    on_fail = "FAIL"
    on_pass = "PASS"

    policy_input_data = {
      schema_type = "TIRITH_JSON"
      data = jsonencode({
        meta = {
          version           = "v1"
          required_provider = "stackguardian/terraform_plan"
        }
        eval_expression = "has-environment-tag"
        evaluators = [{
          id          = "has-environment-tag"
          description = "EC2 instances must be tagged Environment=Production"
          provider_args = {
            operation_type               = "attribute"
            terraform_resource_type      = "aws_instance"
            terraform_resource_attribute = "tags"
          }
          condition = {
            type            = "Contains"
            value           = { Environment = "Production" }
            error_tolerance = 0
          }
        }]
      })
    }

    policy_vcs_config = {
      use_marketplace_template = false
      custom_source = {
        source_config_kind      = "SG_POLICY_FRAMEWORK"
        source_config_dest_kind = "INLINE"
      }
    }
  }]
}
```

## What happens on pass and fail

`on_fail` and `on_pass` take the same four values:

- `FAIL` — stop the run.
- `WARN` — record a warning and let the run continue.
- `PASS` — continue.
- `APPROVAL_REQUIRED` — hold the run until an approver signs off.

`on_pass` is usually `PASS`, but the other values are available when a passing evaluation should
still draw attention.

## Gating a run on approval

Set `on_fail = "APPROVAL_REQUIRED"` and list who may approve. The run pauses instead of failing.

```terraform
resource "stackguardian_policy" "approval_on_apply" {
  resource_name = "approval-on-apply"
  policy_type   = "GENERAL"

  enforced_on = ["/wfgrps/frontend"]

  approvers = [
    "eu-central-1_EXAMPLEPOOL/local/platform-lead@example.com",
  ]
  number_of_approvals_required = 1

  policies_config = [{
    name    = "approval-on-apply"
    on_fail = "APPROVAL_REQUIRED"
    on_pass = "PASS"

    policy_input_data = {
      schema_type = "TIRITH_JSON"
      data = jsonencode({
        meta = {
          version           = "v1"
          required_provider = "stackguardian/terraform_plan"
        }
        eval_expression = "create-operation"
        evaluators = [{
          id          = "create-operation"
          description = "Apply detected -- stop and request approval"
          provider_args = {
            operation_type               = "action"
            terraform_resource_type      = "*"
            terraform_resource_attribute = ""
          }
          condition = {
            type            = "NotEquals"
            value           = "apply"
            error_tolerance = 0
          }
        }]
      })
    }
  }]
}
```

~> `approvers` are fully qualified user IDs, not bare email addresses — the form is
`<user-pool-id>/local/<email>`. Read an existing policy with the `stackguardian_policy` data
source to see the prefix your organization uses.

## Where to go next

The [`stackguardian_policy` resource page](https://registry.terraform.io/providers/StackGuardian/stackguardian/latest/docs/resources/policy)
carries worked examples of all three body sources, including the marketplace and git variants.
