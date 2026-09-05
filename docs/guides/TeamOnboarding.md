---
page_title: "Team Onboarding"
subcategory: "Operations"
description: |-
  Structuring an organization for a team, with worked examples.
---

# Team Onboarding

Onboarding a team means deciding three things: how work is organized, who can reach what, and
which credentials the workflows run with. This guide covers the shape; the worked examples build
it out.

## 1. Decide the workflow group layout

Workflow groups are folders *and* the unit access is scoped to, so the layout decides your
permission model as much as your organization. Nest with `/`:

```terraform
resource "stackguardian_workflow_group" "frontend" {
  resource_name = "frontend"
  description   = "Frontend workflows, owned by the frontend team"
}

resource "stackguardian_workflow_group" "platform" {
  resource_name = "platform"
}

resource "stackguardian_workflow_group" "networking" {
  resource_name = "platform/networking"
  description   = "Networking workflows, owned by the platform team"
}
```

Two layouts cover most cases:

- **By team** — `frontend`, `backend`, `platform`. Roles map to teams.
- **By environment** — `production`, `staging`. Roles map to who may touch production.

Nest them when you need both: `platform/production`.

## 2. Define roles, then assign them

One role per level of access, scoped to the groups it covers:

```terraform
resource "stackguardian_rolev4" "frontend_developer" {
  resource_name = "frontend-developer"

  allowed_permissions = {
    "GET/api/v1/orgs/<org>/wfgrps/<wfGrp>/" = {
      name = "GetWorkflowGroup"
      paths = {
        "<wfGrp>" = [
          resource.stackguardian_workflow_group.frontend.resource_name,
        ]
      }
    },
  }
}

resource "stackguardian_role_assignment" "alice" {
  user_id     = "alice@example.com"
  entity_type = "EMAIL"
  role        = stackguardian_rolev4.frontend_developer.resource_name
}
```

Assign to SSO groups rather than individuals where you can — it keeps the Terraform
configuration stable as people join and leave. See
[Access Management](https://registry.terraform.io/providers/StackGuardian/stackguardian/latest/docs/guides/AccessManagement)
for the SSO naming pitfall.

## 3. Add connectors

Connectors hold the credentials workflows deploy with. Create them once, per environment, and
reference them from workflows rather than repeating credentials:

```terraform
resource "stackguardian_connector" "production_aws" {
  resource_name = "production-aws"

  settings = {
    kind = "AWS_RBAC"
    config = [{
      role_arn         = "arn:aws:iam::000000000000:role/StackGuardianProduction"
      external_id      = "REPLACE-ME"
      duration_seconds = "3600"
    }]
  }
}
```

Prefer role-assumption kinds (`AWS_RBAC`, `AWS_OIDC`, `AZURE_OIDC`, `GCP_OIDC`) over static keys.

## 4. Add guardrails

Policies enforce rules at run time, independently of who has access:

```terraform
resource "stackguardian_policy" "production_guardrail" {
  resource_name = "production-guardrail"
  policy_type   = "GENERAL"

  # No trailing slash on a workflow group path.
  enforced_on = ["/wfgrps/production"]

  policies_config = [{
    name    = "production-guardrail"
    on_fail = "FAIL"
    on_pass = "PASS"

    # A policy needs a body. Written inline, that means policy_input_data
    # carrying the definition and a policy_vcs_config saying it is inline.
    policy_input_data = {
      schema_type = "TIRITH_JSON"
      data = jsonencode({
        meta = {
          version           = "v1"
          required_provider = "stackguardian/terraform_plan"
        }
        eval_expression = "no-public-buckets"
        evaluators = [{
          id          = "no-public-buckets"
          description = "S3 buckets must not be public"
          provider_args = {
            operation_type               = "attribute"
            terraform_resource_type      = "aws_s3_bucket"
            terraform_resource_attribute = "acl"
          }
          condition = {
            type            = "NotEquals"
            value           = "public-read"
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

See [Policies](https://registry.terraform.io/providers/StackGuardian/stackguardian/latest/docs/guides/Policies)
for the other ways to supply a body, and for approval gating.

## Worked examples

Two complete configurations are maintained in the repository:

| Example                                                                                                                          | Shows                                                              |
| -------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------ |
| [`project-01`](https://github.com/StackGuardian/terraform-provider-stackguardian/tree/main/docs-guides-assets/onboarding/project-01) | A flat team: three workflow groups, one developer role, connectors, and an import script. |
| [`project-02`](https://github.com/StackGuardian/terraform-provider-stackguardian/tree/main/docs-guides-assets/onboarding/project-02) | A hierarchical team: manager and developer roles per team, differing in who may delete a workflow and release a held run. |

Both use placeholder credentials and organization names — replace them before applying.

## Bringing an existing organization under Terraform

If the team already has workflow groups and roles, import them rather than recreating them. See
[Importing Existing Resources](https://registry.terraform.io/providers/StackGuardian/stackguardian/latest/docs/guides/ImportingResources);
`project-01` includes an import script covering every resource it declares.

## Building this with AI

<!-- AI-SKILLS:START -->
Generating configuration from this guide? Load the **`stackguardian-access`** skill, which turns the guidance here into rules an agent can follow.

**Worth knowing either way:** Which permissions belong at which level is a decision about your organization, not a fact about the provider.

The skills live in [the provider repository](https://github.com/StackGuardian/terraform-provider-stackguardian/tree/main/.claude/skills) and work with Claude Code, Cursor, Copilot, Windsurf and any agent that reads [`AGENTS.md`](https://github.com/StackGuardian/terraform-provider-stackguardian/blob/main/AGENTS.md).
<!-- AI-SKILLS:END -->
