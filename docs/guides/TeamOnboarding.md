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
[Access Control](https://registry.terraform.io/providers/StackGuardian/stackguardian/latest/docs/guides/AccessControl)
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
  enforced_on   = ["/wfgrps/production/"]

  policies_config = [{
    name    = "production-guardrail"
    on_fail = "FAIL"
  }]
}
```

## Worked examples

Two complete configurations are maintained in the repository:

| Example                                                                                                                          | Shows                                                              |
| -------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------ |
| [`project-01`](https://github.com/StackGuardian/terraform-provider-stackguardian/tree/main/docs-guides-assets/onboarding/project-01) | A flat team: three workflow groups, one developer role, connectors, and an import script. |
| [`project-02`](https://github.com/StackGuardian/terraform-provider-stackguardian/tree/main/docs-guides-assets/onboarding/project-02) | A hierarchical team: frontend and backend groups with separate manager and developer roles. |

Both use placeholder credentials and organization names — replace them before applying.

## Bringing an existing organization under Terraform

If the team already has workflow groups and roles, import them rather than recreating them. See
[Importing Existing Resources](https://registry.terraform.io/providers/StackGuardian/stackguardian/latest/docs/guides/ImportingResources);
`project-01` includes an import script covering every resource it declares.
