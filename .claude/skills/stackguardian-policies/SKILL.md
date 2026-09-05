---
name: stackguardian-policies
description: Write or debug StackGuardian policies — guardrails evaluated during runs, approval gates, and the three ways a policy body is supplied (inline Tirith, marketplace template, or Rego from git). Use when adding a policy, scoping one with enforced_on, configuring approvers, or when a policy resource shows a permanent diff after apply.
metadata:
  lifecycle-status: active
---

A policy is a guardrail evaluated during a workflow or stack run. It can fail a run, warn, or hold
it for human approval.

## Where the body comes from

Three cases, and the two config blocks are **not** alternatives:

| Case | `policy_input_data` | `policy_vcs_config` |
| --- | --- | --- |
| Inline Tirith policy | the body | `custom_source` = `SG_POLICY_FRAMEWORK` / `INLINE` |
| Marketplace template | omitted — the template carries it | `use_marketplace_template = true` + `policy_template_id` |
| Rego from your git | omitted | `custom_source` = `OPA_REGO` + repo config |

An **inline policy sets both**. Writing only `policy_input_data` is the most common mistake here.

## schema_type

For an inline SG_POLICY_FRAMEWORK policy the value is **`TIRITH_JSON`**.

`RAW_JSON` is a legal value in the SDK enum, so it applies — but the platform stores and returns
`TIRITH_JSON`, which leaves a permanent configuration/state diff. Use `TIRITH_JSON`.

```terraform
policies_config = [{
  name    = "require-environment-tag"
  skip    = false
  on_fail = "FAIL"
  on_pass = "PASS"

  policy_input_data = {
    schema_type = "TIRITH_JSON"
    data = jsonencode({
      meta            = { version = "v1", required_provider = "stackguardian/terraform_plan" }
      eval_expression = "has-env-tag"
      evaluators = [{
        id            = "has-env-tag"
        provider_args = {
          operation_type               = "attribute"
          terraform_resource_type      = "aws_instance"
          terraform_resource_attribute = "tags"
        }
        condition = { type = "Contains", value = { Environment = "Production" } }
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
```

`data` is a JSON **string**, so `jsonencode` keeps it readable.

## on_fail

- `FAIL` — stop the run
- `WARN` — record a warning, continue
- `PASS` — treat the failure as acceptable
- `APPROVAL_REQUIRED` — hold the run until an approver signs off

## Scope

```terraform
enforced_on = ["/wfgrps/production"]   # a group and everything in it; no trailing slash
enforced_on = ["*"]                    # the whole organization; used alone
```

`["*"]` is organization-wide — never reach for it to make a test policy inert. Scope a throwaway
policy to a group that does not exist instead, so it gates nothing.

## Approvers

An entry is a user's **email address**, or an **SSO group name** to allow anyone in that group. The
fully qualified `<user-pool-id>/local/<email>` form is also accepted. Bare emails are the common
case.

`number_of_approvals_required` interacts with the list and the pair should move together: `1` means
any single approver unblocks; `0` with a non-empty list means **unanimous**, while `0` with an empty
list means anyone can approve. Adding approvers to a policy whose count is `0` therefore turns
"anyone" into "everyone".

Approvals on a **workflow** work the same way, and there `approval_pre_apply` decides whether runs
stop at all. A gate with no approvers can be released by anyone, so set both.

## Output contract

State which of the three body sources is in use, that both config blocks are present for an inline
policy, what the policy is scoped to, and — for an approval gate — who can release it and whether
runs actually stop.

Full reference: `docs/guides/Policies.md`.
