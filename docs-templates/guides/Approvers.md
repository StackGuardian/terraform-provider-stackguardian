---
page_title: "Approvers"
subcategory: "Concepts"
description: |-
  How to write an `approvers` entry for a local user, an SSO user or an SSO group, and which user-pool prefix SSO entries take.
---

# Approvers

An `approvers` list names who may release a run that is waiting for approval. The same list, with
the same entry format, appears on:

- `stackguardian_policy` — for rules that end in `APPROVAL_REQUIRED`.
- `stackguardian_workflow_git` and `stackguardian_workflow_from_template` — for the workflow's own
  approval gates, such as `terraform_config.approval_pre_apply`.
- `stackguardian_workflow_template_revision` and `stackguardian_stack_template_revision` — as the
  default for workflows created from the revision.

## How to write an approver

| Who | Write it as | Example (EU region) |
|-----|-------------|---------------------|
| **Local user** — an account created directly in StackGuardian | the email address | `jane@example.com` |
| **SSO user** — signs in through your identity provider | `<sso-pool-id>/<sso-provider-name>/<email>` | `eu-central-1_xut85XJiL/sg-test-sso/jane@example.com` |
| **SSO group** — anyone in the group may approve | `<sso-pool-id>/group/<group-id>` | `eu-central-1_xut85XJiL/group/platform-admins` |

For a group the middle part is the literal word `group`, not the SSO provider name.

## User pool IDs

SSO users and groups take a prefix that depends on your StackGuardian region.

| Region | `api_uri` | SSO users and groups | Local users (optional) |
|--------|-----------|----------------------|------------------------|
| EU | `https://api.app.stackguardian.io` | `eu-central-1_xut85XJiL` | `eu-central-1_srEmUITJM` |
| US | `https://api.us.stackguardian.io` | `us-east-2_LKSJfYtpl` | `us-east-2_B8artiaXu` |

On any other StackGuardian installation the IDs are different —
[read them from an existing resource](#checking-a-value) instead.

The local pool ID is only needed for the [local-account-only form](#allowing-the-local-account-only).

## SSO provider name and group ID

Both are the values you already use in `stackguardian_role_assignment`. If someone has a role
assignment, their approver entry follows from it:

| `user_id` in the role assignment | `entity_type` | Approver entry |
|----------------------------------|---------------|----------------|
| `jane@example.com` | `EMAIL` | `jane@example.com` |
| `sg-test-sso/jane@example.com` | `EMAIL` | `<sso-pool-id>/sg-test-sso/jane@example.com` |
| `sg-test-sso/group-devs` | `GROUP` | `<sso-pool-id>/group/group-devs` |

- **SSO provider name** — the name of your organization's SSO configuration (`sg-test-sso` above).
- **Group ID** — the group exactly as your identity provider sends it to StackGuardian. For some
  providers that is a readable name, for others (Microsoft Entra ID, for example) an object ID.

## Example

Keep the SSO prefix in a `local` so each entry stays readable:

```terraform
locals {
  sso_pool = "eu-central-1_xut85XJiL" # US region: us-east-2_LKSJfYtpl
}

resource "stackguardian_policy" "approval_on_apply" {
  resource_name = "approval-on-apply"
  policy_type   = "GENERAL"

  approvers = [
    "platform-lead@example.com",                            # local user
    "${local.sso_pool}/sg-test-sso/sre-oncall@example.com", # SSO user
    "${local.sso_pool}/group/platform-admins",              # SSO group
  ]
  number_of_approvals_required = 1
}
```

Workflows and template revisions take exactly the same entries.

## Rules worth knowing

- **Write emails in lowercase.** Workflows store the email part in lowercase, so mixed case in
  your configuration will not match what the platform returns. The user pool ID and the SSO
  provider name are case-sensitive and must match exactly.
- **An SSO entry only matches an SSO sign-in.** `<sso-pool-id>/sg-test-sso/jane@example.com`
  cannot approve from a local account with the same address.
- **With a group in the list, set `number_of_approvals_required` to `1` or more.** `0` means
  "every listed approver must approve", which cannot be counted for a group.
- **One person counts once.** Someone listed as an SSO user who is also in a listed SSO group
  still gives a single approval.
- **Entries are not checked at apply time.** A wrong prefix or a typo applies cleanly, and that
  entry then never matches anyone.
- **An empty list on a workflow means anyone can approve.** A workflow gate with no approvers
  can be released by any user who can reach the run.

## Allowing the local account only

A bare email address matches that address however the person signs in — with their local account,
or through SSO if they have both. That is usually what you want. To accept the local account and
nothing else, write the local user in full, with the local pool ID from the
[table above](#user-pool-ids):

```terraform
locals {
  local_pool = "eu-central-1_srEmUITJM" # US region: us-east-2_B8artiaXu
}

resource "stackguardian_policy" "approval_on_apply" {
  resource_name = "approval-on-apply"
  policy_type   = "GENERAL"

  approvers                    = ["${local.local_pool}/local/platform-lead@example.com"]
  number_of_approvals_required = 1
}
```

The StackGuardian dashboard always saves this full form, so a local user added there reads back
as `<local-pool-id>/local/<email>`. Change an approvers list that Terraform manages in Terraform,
not in the dashboard, to keep the two from disagreeing.

## Checking a value

Not sure about a prefix, a provider name or a group ID? Add the approver once in the StackGuardian
dashboard, then read it back:

```terraform
data "stackguardian_policy" "existing" {
  resource_name = "approval-on-apply"
}

output "approvers" {
  value = data.stackguardian_policy.existing.approvers
}
```

The output shows each approver exactly as the platform stores it — copy the prefix, provider
name or group ID from there.

See [Review and approve Workflow Runs](https://docs.stackguardian.io/docs/deploy/workflows/workflow_components/approvals_config/)
in the platform documentation for how approvals are counted and what approvers see.
