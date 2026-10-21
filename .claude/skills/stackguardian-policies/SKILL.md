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

Approvers apply only to `policy_type = "GENERAL"`. Each entry is one of:

| Who | Form |
| --- | --- |
| Local user | `jane@example.com` (also matches an SSO sign-in with the same address) |
| SSO user | `<sso-pool-id>/<sso-provider-name>/<email>` |
| SSO group — anyone in it may approve | `<sso-pool-id>/group/<group-id>` (literal word `group`) |
| Local account only | `<local-pool-id>/local/<email>` |

The pool id depends on the org's region. Write emails in **lowercase**. Entries are not validated
at apply time — a typo applies cleanly and then never matches anyone.

`number_of_approvals_required` and the list should move together: `0` (the default) waits for as
many distinct approvals as there are entries — **everyone**, when the list holds only individual
users; `1` or more means that many different people. Use `1`+ whenever the list contains an SSO
group. A value higher than the number of people who can approve leaves the run waiting forever, and
a single rejection cancels the run regardless.

On a **workflow**, `approval_pre_apply` decides whether runs stop at all, and an **empty**
`approvers` list there means anyone who can reach the run may release it — so set both. A run held
by both a workflow gate and policies must satisfy each one separately.

## Output contract

State which of the three body sources is in use, that both config blocks are present for an inline
policy, what the policy is scoped to, and — for an approval gate — who can release it and whether
runs actually stop.

Full reference: `docs/guides/Policies.md`, and `docs/guides/Approvers.md` for approver formats and
pool ids.
