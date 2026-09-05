// Project-02 -- a hierarchical team.
//
// Two teams, frontend and backend, each with a manager and developers, plus a
// cross-functional devops role contributing to both. The access model:
//
//   Developer  -- read the group, build and run workflows in it. No delete.
//   Manager    -- everything a developer can do in the same group, plus delete a
//                 workflow and resume a run that a policy has held for approval.
//   DevOps     -- read and run across all three groups, plus delete, because this
//                 role is responsible for tearing down what is no longer needed.
//
// Roles use stackguardian_rolev4. In v4 a placeholder takes ONE entry per list
// slot, so several allowed values are combined into a single alternation string
// with "|". stackguardian_role (v3) is deprecated.

terraform {
  required_providers {
    stackguardian = {
      source  = "StackGuardian/stackguardian"
      version = "~> 1.12"
    }
  }
}

provider "stackguardian" {
  api_key  = "<YOUR-API-KEY>"                   # Replace this with your API key
  org_name = "<YOUR-ORG-NAME>"                  # Replace this with your organization name
  api_uri  = "https://api.app.stackguardian.io" # Use https://api.us.stackguardian.io for the US region
}

# --- Workflow groups -----------------------------------------------------------

resource "stackguardian_workflow_group" "frontend" {
  resource_name = "ONBOARDING-Project02-Frontend"
  description   = "Onboarding example of terraform-provider-stackguardian for WorkflowGroup"
  tags          = ["tf-provider-example", "onboarding"]
}

resource "stackguardian_workflow_group" "backend" {
  resource_name = "ONBOARDING-Project02-Backend"
  description   = "Onboarding example of terraform-provider-stackguardian for WorkflowGroup"
  tags          = ["tf-provider-example", "onboarding"]
}

resource "stackguardian_workflow_group" "devops" {
  resource_name = "ONBOARDING-Project02-DevOps"
  description   = "Onboarding example of terraform-provider-stackguardian for WorkflowGroup"
  tags          = ["tf-provider-example", "onboarding"]
}

# --- Permission sets -----------------------------------------------------------
# Built once as locals and reused, so the difference between a manager and a
# developer is visible in one place rather than buried in five near-identical
# blocks. Each takes the group(s) the role is scoped to.

locals {
  all_groups = join("|", [
    stackguardian_workflow_group.frontend.resource_name,
    stackguardian_workflow_group.backend.resource_name,
    stackguardian_workflow_group.devops.resource_name,
  ])

  # Read the group, build and run workflows in it, read the results.
  developer_permissions = {
    for grp in ["FRONTEND", "BACKEND", "DEVOPS", "ALL"] : grp => {
      "GET/api/v1/orgs/<org>/wfgrps/<wfGrp>/" = {
        name  = "GetWorkflowGroup"
        paths = { "<wfGrp>" = [local.group_scope[grp]] }
      },
      "GET/api/v1/orgs/<org>/wfgrps/<wfGrp>/wfs/<wf>/" = {
        name  = "GetWorkflow"
        paths = { "<wfGrp>" = [local.group_scope[grp]], "<wf>" = [".*"] }
      },
      "POST/api/v1/orgs/<org>/wfgrps/<wfGrp>/wfs/" = {
        name  = "CreateWorkflow"
        paths = { "<wfGrp>" = [local.group_scope[grp]] }
      },
      "PATCH/api/v1/orgs/<org>/wfgrps/<wfGrp>/wfs/<wf>/" = {
        name  = "UpdateWorkflow"
        paths = { "<wfGrp>" = [local.group_scope[grp]], "<wf>" = [".*"] }
      },
      "POST/api/v1/orgs/<org>/wfgrps/<wfGrp>/wfs/<wf>/wfruns/" = {
        name  = "CreateWorkflowRun"
        paths = { "<wfGrp>" = [local.group_scope[grp]], "<wf>" = [".*"] }
      },
      "GET/api/v1/orgs/<org>/wfgrps/<wfGrp>/wfs/<wf>/wfruns/<wfRun>/" = {
        name  = "GetWorkflowRun"
        paths = { "<wfGrp>" = [local.group_scope[grp]], "<wf>" = [".*"], "<wfRun>" = [".*"] }
      },
      "GET/api/v1/orgs/<org>/wfgrps/<wfGrp>/wfs/<wf>/wfruns/<wfRun>/logs/" = {
        name  = "GetWorkflowRunLogs"
        paths = { "<wfGrp>" = [local.group_scope[grp]], "<wf>" = [".*"], "<wfRun>" = [".*"] }
      },
      "GET/api/v1/orgs/<org>/wfgrps/<wfGrp>/wfs/<wf>/outputs/" = {
        name  = "GetWorkflowOutputs"
        paths = { "<wfGrp>" = [local.group_scope[grp]], "<wf>" = [".*"] }
      },
      "GET/api/v1/orgs/<org>/integrationgroups/<integrationgroup>/integrations/<integration>/" = {
        name  = "GetIntegrationGroupChild"
        paths = { "<integrationgroup>" = [".*"], "<integration>" = [".*"] }
      },
    }
  }

  # What a manager can do that a developer cannot: remove a workflow, and release
  # a run that an APPROVAL_REQUIRED policy is holding.
  elevated_permissions = {
    for grp in ["FRONTEND", "BACKEND", "DEVOPS", "ALL"] : grp => {
      "DELETE/api/v1/orgs/<org>/wfgrps/<wfGrp>/wfs/<wf>/" = {
        name  = "DeleteWorkflow"
        paths = { "<wfGrp>" = [local.group_scope[grp]], "<wf>" = [".*"] }
      },
      "POST/api/v1/orgs/<org>/wfgrps/<wfGrp>/wfs/<wf>/wfruns/<wfRun>/resume/" = {
        name  = "ResumeWorkflowRun"
        paths = { "<wfGrp>" = [local.group_scope[grp]], "<wf>" = [".*"], "<wfRun>" = [".*"] }
      },
    }
  }

  group_scope = {
    FRONTEND = stackguardian_workflow_group.frontend.resource_name
    BACKEND  = stackguardian_workflow_group.backend.resource_name
    DEVOPS   = stackguardian_workflow_group.devops.resource_name
    ALL      = local.all_groups
  }
}

# --- Roles ---------------------------------------------------------------------

resource "stackguardian_rolev4" "manager_frontend" {
  resource_name       = "ONBOARDING-Project02-Manager-Frontend"
  description         = "Onboarding example of terraform-provider-stackguardian for Role Manager of Frontend team"
  tags                = ["tf-provider-example", "onboarding"]
  allowed_permissions = merge(local.developer_permissions.FRONTEND, local.elevated_permissions.FRONTEND)
}

resource "stackguardian_rolev4" "developer_frontend" {
  resource_name       = "ONBOARDING-Project02-Developer-Frontend"
  description         = "Onboarding example of terraform-provider-stackguardian for Role Developer of Frontend team"
  tags                = ["tf-provider-example", "onboarding"]
  allowed_permissions = local.developer_permissions.FRONTEND
}

resource "stackguardian_rolev4" "manager_backend" {
  resource_name       = "ONBOARDING-Project02-Manager-Backend"
  description         = "Onboarding example of terraform-provider-stackguardian for Role Manager of Backend team"
  tags                = ["tf-provider-example", "onboarding"]
  allowed_permissions = merge(local.developer_permissions.BACKEND, local.elevated_permissions.BACKEND)
}

resource "stackguardian_rolev4" "developer_backend" {
  resource_name       = "ONBOARDING-Project02-Developer-Backend"
  description         = "Onboarding example of terraform-provider-stackguardian for Role Developer of Backend team"
  tags                = ["tf-provider-example", "onboarding"]
  allowed_permissions = local.developer_permissions.BACKEND
}

# Cross-functional: contributes to both teams, and owns teardown, so it gets the
# elevated set across every group.
resource "stackguardian_rolev4" "developer_devops" {
  resource_name       = "ONBOARDING-Project02-Developer-DevOps"
  description         = "Onboarding example of terraform-provider-stackguardian for cross-functional DevOps Role"
  tags                = ["tf-provider-example", "onboarding"]
  allowed_permissions = merge(local.developer_permissions.ALL, local.elevated_permissions.ALL)
}

# --- Role assignments ----------------------------------------------------------
# Every role declared above is assigned; a role nobody holds grants nothing.

resource "stackguardian_role_assignment" "manager_frontend" {
  user_id     = "frontend.manager.p02@dummy.com"
  entity_type = "EMAIL"
  role        = stackguardian_rolev4.manager_frontend.resource_name
}

resource "stackguardian_role_assignment" "developer_frontend" {
  user_id     = "frontend.developer.p02@dummy.com"
  entity_type = "EMAIL"
  role        = stackguardian_rolev4.developer_frontend.resource_name
}

resource "stackguardian_role_assignment" "manager_backend" {
  user_id     = "backend.manager.p02@dummy.com"
  entity_type = "EMAIL"
  role        = stackguardian_rolev4.manager_backend.resource_name
}

resource "stackguardian_role_assignment" "developer_backend" {
  user_id     = "backend.developer.p02@dummy.com"
  entity_type = "EMAIL"
  role        = stackguardian_rolev4.developer_backend.resource_name
}

resource "stackguardian_role_assignment" "developer_devops" {
  user_id     = "devops.developer.p02@dummy.com"
  entity_type = "EMAIL"
  role        = stackguardian_rolev4.developer_devops.resource_name
}

# --- Connectors ----------------------------------------------------------------

resource "stackguardian_connector" "cloud" {
  resource_name = "ONBOARDING-Project02-Cloud-Connector"
  description   = "Onboarding example of terraform-provider-stackguardian for ConnectorCloud"

  settings = {
    kind = "AWS_RBAC"
    config = [{
      role_arn         = "arn:aws:iam::000000000000:role/StackGuardian"
      external_id      = "REPLACE-ME"
      duration_seconds = "3600"
    }]
  }
}

resource "stackguardian_connector" "vcs" {
  resource_name = "ONBOARDING-Project02-VCS-Connector"
  description   = "Onboarding example of terraform-provider-stackguardian for ConnectorVcs"

  settings = {
    kind = "GITLAB_COM"
    config = [{
      gitlab_creds = "REPLACE-ME-user:REPLACE-ME-token"
    }]
  }
}

# --- A workflow ----------------------------------------------------------------

resource "stackguardian_workflow_git" "backend_deploy" {
  workflow_group_id = stackguardian_workflow_group.backend.resource_name
  id                = "ONBOARDING-Project02-Backend-Deploy"
  wf_type           = "TERRAFORM"

  description = "Deploys the Terraform in the referenced repository"
  tags        = ["tf-provider-example", "onboarding"]

  vcs_config = {
    iac_vcs_config = {
      custom_source = {
        source_config_dest_kind = "GIT_OTHER"
        config = {
          is_private = false
          repo       = "https://github.com/my-org/my-repo.git"
        }
      }
    }
  }

  deployment_platform_config = [{
    kind = "AWS_RBAC"
    config = {
      integration_id = stackguardian_connector.cloud.id
    }
  }]
}
