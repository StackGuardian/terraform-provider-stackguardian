---
page_title: "Getting Started"
subcategory: "Getting Started"
description: |-
  Build a working StackGuardian deployment from nothing, one resource at a time.
---

# Getting Started

This guide builds one working deployment: a workflow that deploys Terraform from your repository
using credentials you control. It assumes you have completed
[Installation](https://registry.terraform.io/providers/StackGuardian/stackguardian/latest/docs/guides/Installation)
and exported `STACKGUARDIAN_API_KEY` and `STACKGUARDIAN_ORG_NAME`.

The finished configuration is three resources:

```
workflow_group   →   connector   →   workflow_git
   the folder        the creds       the workflow
```

## 1. Create a workflow group

Everything lives in a workflow group, so start there.

```terraform
resource "stackguardian_workflow_group" "quickstart" {
  resource_name = "quickstart"
  description   = "Created while following the getting started guide"
  tags          = ["quickstart"]
}
```

Apply this on its own if you like — `terraform apply` — and you will see the group appear in the
StackGuardian UI.

~> A `/` in `resource_name` creates a nested group rather than a name containing a slash.
`platform/networking` puts `networking` inside an existing `platform` group.

## 2. Add a connector

A workflow needs credentials to deploy with. That is a connector.

```terraform
resource "stackguardian_connector" "aws" {
  resource_name = "quickstart-aws"
  description   = "AWS credentials for the quickstart workflow"

  settings = {
    kind = "AWS_RBAC"
    config = [{
      role_arn         = "arn:aws:iam::000000000000:role/StackGuardian"
      external_id      = "REPLACE-ME"
      duration_seconds = "3600"
    }]
  }
}
```

`kind` selects the connector type, and each kind takes a different set of `config` fields — the
[connector resource page](https://registry.terraform.io/providers/StackGuardian/stackguardian/latest/docs/resources/connector)
lists them all. `AWS_RBAC` assumes an IAM role, which is preferable to static keys.

## 3. Create the workflow

Now wire the two together.

```terraform
resource "stackguardian_workflow_git" "quickstart" {
  workflow_group_id = stackguardian_workflow_group.quickstart.resource_name
  id                = "quickstart-workflow"
  wf_type           = "TERRAFORM"

  description = "Deploys the Terraform in the referenced repository"

  vcs_config = {
    iac_vcs_config = {
      custom_source = {
        source_config_dest_kind = "GIT_OTHER"
        config = {
          is_private  = false
          repo        = "https://github.com/my-org/my-terraform-repo.git"
          working_dir = "."
          ref         = "main"
        }
      }
    }
  }

  deployment_platform_config = [{
    kind = "AWS_RBAC"
    config = {
      integration_id = stackguardian_connector.aws.id
    }
  }]
}
```

Two things worth noticing:

- `id` is **required** and chosen by you. Changing it later replaces the workflow.
- `integration_id` references the connector resource rather than a hard-coded string, so
  Terraform creates them in the right order and keeps them in step.

Run `terraform apply`. The workflow appears in the group, ready to run.

## 4. Use the results

Once the workflow has run, read its outputs back:

```terraform
data "stackguardian_workflow_outputs" "quickstart" {
  workflow       = stackguardian_workflow_git.quickstart.id
  workflow_group = stackguardian_workflow_group.quickstart.resource_name
}

locals {
  outputs = jsondecode(data.stackguardian_workflow_outputs.quickstart.data_json)
}
```

Outputs arrive as a JSON string, so decode before use.

## Cloning a private repository

The example above uses a public repository. For a private one, set `is_private` and point `auth`
at a VCS connector:

```terraform
config = {
  is_private = true
  auth       = stackguardian_connector.github.id
  repo       = "https://github.com/my-org/private-repo.git"
}
```

## Where to go next

- [Object Model](https://registry.terraform.io/providers/StackGuardian/stackguardian/latest/docs/guides/ObjectModel) — how everything fits together.
- [Resource IDs](https://registry.terraform.io/providers/StackGuardian/stackguardian/latest/docs/guides/ResourceIDs) — why some values look like `/integrations/my-connector`.
- [Templates](https://registry.terraform.io/providers/StackGuardian/stackguardian/latest/docs/guides/Templates) — when several workflows should share one definition.
- [Access Management](https://registry.terraform.io/providers/StackGuardian/stackguardian/latest/docs/guides/AccessManagement) — granting your team access to what you just built.
- [Importing Resources](https://registry.terraform.io/providers/StackGuardian/stackguardian/latest/docs/guides/ImportingResources) — bringing existing StackGuardian objects under Terraform.

A complete, runnable version of this configuration is in
[`docs-guides-assets/quickstart`](https://github.com/StackGuardian/terraform-provider-stackguardian/tree/main/docs-guides-assets/quickstart).

## Building this with AI

<!-- AI-SKILLS:START -->
Generating configuration from this guide? Load the **`stackguardian-provider`** skill, which turns the guidance here into rules an agent can follow.

**Worth knowing either way:** Order matters: the workflow group and connector must exist before the workflow that references them.

The skills live in [the provider repository](https://github.com/StackGuardian/terraform-provider-stackguardian/tree/main/.claude/skills) and work with Claude Code, Cursor, Copilot, Windsurf and any agent that reads [`AGENTS.md`](https://github.com/StackGuardian/terraform-provider-stackguardian/blob/main/AGENTS.md).
<!-- AI-SKILLS:END -->
