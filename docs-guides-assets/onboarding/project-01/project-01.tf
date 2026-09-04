// Project-01 -- a flat team.
//
// One team with no management layer: everyone works across frontend, backend and
// devops, so a single role covers everybody. The team is also responsible for
// tearing down workflows it no longer needs, so the role includes delete.
//
// Roles use stackguardian_rolev4. In v4 a placeholder takes ONE entry per list
// slot, so several allowed values are combined into a single alternation string
// with "|". stackguardian_role (v3) instead took a list and combined every path
// with every other path; it is deprecated.

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
# Declared first: the role below scopes itself by referencing them.

resource "stackguardian_workflow_group" "frontend" {
  resource_name = "ONBOARDING-Project01-Frontend"
  description   = "Onboarding example of terraform-provider-stackguardian for WorkflowGroup"
  tags          = ["tf-provider-example", "onboarding"]
}

resource "stackguardian_workflow_group" "backend" {
  resource_name = "ONBOARDING-Project01-Backend"
  description   = "Onboarding example of terraform-provider-stackguardian for WorkflowGroup"
  tags          = ["tf-provider-example", "onboarding"]
}

resource "stackguardian_workflow_group" "devops" {
  resource_name = "ONBOARDING-Project01-DevOps"
  description   = "Onboarding example of terraform-provider-stackguardian for WorkflowGroup"
  tags          = ["tf-provider-example", "onboarding"]
}

# All three groups, as one alternation string. Referencing the resources rather
# than typing names means the role follows any rename.
locals {
  all_groups = join("|", [
    stackguardian_workflow_group.frontend.resource_name,
    stackguardian_workflow_group.backend.resource_name,
    stackguardian_workflow_group.devops.resource_name,
  ])
}

# --- Role ----------------------------------------------------------------------
# One full-stack Developer role over all three groups. Replace <org> with your
# organization name -- it is part of the permission key, not just documentation.

resource "stackguardian_rolev4" "developer" {
  resource_name = "ONBOARDING-Project01-Developer"
  description   = "Onboarding example of terraform-provider-stackguardian for Role Developer"
  tags          = ["tf-provider-example", "onboarding"]

  allowed_permissions = {
    # Workflow groups
    "GET/api/v1/orgs/<org>/wfgrps/<wfGrp>/" = {
      name  = "GetWorkflowGroup"
      paths = { "<wfGrp>" = [local.all_groups] }
    },

    # Workflows -- full lifecycle, including delete for teardown
    "GET/api/v1/orgs/<org>/wfgrps/<wfGrp>/wfs/<wf>/" = {
      name  = "GetWorkflow"
      paths = { "<wfGrp>" = [local.all_groups], "<wf>" = [".*"] }
    },
    "POST/api/v1/orgs/<org>/wfgrps/<wfGrp>/wfs/" = {
      name  = "CreateWorkflow"
      paths = { "<wfGrp>" = [local.all_groups] }
    },
    "PATCH/api/v1/orgs/<org>/wfgrps/<wfGrp>/wfs/<wf>/" = {
      name  = "UpdateWorkflow"
      paths = { "<wfGrp>" = [local.all_groups], "<wf>" = [".*"] }
    },
    "DELETE/api/v1/orgs/<org>/wfgrps/<wfGrp>/wfs/<wf>/" = {
      name  = "DeleteWorkflow"
      paths = { "<wfGrp>" = [local.all_groups], "<wf>" = [".*"] }
    },

    # Runs
    "GET/api/v1/orgs/<org>/wfgrps/<wfGrp>/wfs/<wf>/wfruns/<wfRun>/" = {
      name  = "GetWorkflowRun"
      paths = { "<wfGrp>" = [local.all_groups], "<wf>" = [".*"], "<wfRun>" = [".*"] }
    },
    "POST/api/v1/orgs/<org>/wfgrps/<wfGrp>/wfs/<wf>/wfruns/" = {
      name  = "CreateWorkflowRun"
      paths = { "<wfGrp>" = [local.all_groups], "<wf>" = [".*"] }
    },
    "POST/api/v1/orgs/<org>/wfgrps/<wfGrp>/wfs/<wf>/wfruns/<wfRun>/resume/" = {
      name  = "ResumeWorkflowRun"
      paths = { "<wfGrp>" = [local.all_groups], "<wf>" = [".*"], "<wfRun>" = [".*"] }
    },
    "GET/api/v1/orgs/<org>/wfgrps/<wfGrp>/wfs/<wf>/wfruns/<wfRun>/logs/" = {
      name  = "GetWorkflowRunLogs"
      paths = { "<wfGrp>" = [local.all_groups], "<wf>" = [".*"], "<wfRun>" = [".*"] }
    },

    # Outputs and artifacts
    "GET/api/v1/orgs/<org>/wfgrps/<wfGrp>/wfs/<wf>/outputs/" = {
      name  = "GetWorkflowOutputs"
      paths = { "<wfGrp>" = [local.all_groups], "<wf>" = [".*"] }
    },
    "GET/api/v1/orgs/<org>/wfgrps/<wfGrp>/wfs/<wf>/listall_artifacts/" = {
      name  = "ListWorkflowArtifacts"
      paths = { "<wfGrp>" = [local.all_groups], "<wf>" = [".*"] }
    },

    # Connectors -- read only, so workflows can reference them
    "GET/api/v1/orgs/<org>/integrationgroups/<integrationgroup>/" = {
      name  = "GetIntegrationGroup"
      paths = { "<integrationgroup>" = [".*"] }
    },
    "GET/api/v1/orgs/<org>/integrationgroups/<integrationgroup>/integrations/<integration>/" = {
      name  = "GetIntegrationGroupChild"
      paths = { "<integrationgroup>" = [".*"], "<integration>" = [".*"] }
    },
  }
}

# --- Role assignment -----------------------------------------------------------
# Prefer assigning to an SSO group over an individual where you can; it keeps this
# file stable as people join and leave.

resource "stackguardian_role_assignment" "developer" {
  user_id     = "frontend.developer.p01@dummy.com"
  entity_type = "EMAIL"
  role        = stackguardian_rolev4.developer.resource_name
}

# --- Connectors ----------------------------------------------------------------
# AWS_RBAC assumes a role rather than storing keys. Prefer it, and the other
# role-assumption kinds, over the static variants.

resource "stackguardian_connector" "cloud" {
  resource_name = "ONBOARDING-Project01-Cloud-Connector"
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
  resource_name = "ONBOARDING-Project01-VCS-Connector"
  description   = "Onboarding example of terraform-provider-stackguardian for ConnectorVcs"

  settings = {
    kind = "GITLAB_COM"
    config = [{
      gitlab_creds = "REPLACE-ME-user:REPLACE-ME-token"
    }]
  }
}

# --- A workflow ----------------------------------------------------------------
# Without this the groups stay empty and the cloud connector is never used.
# deployment_platform_config is what wires the connector in.

resource "stackguardian_workflow_git" "frontend_deploy" {
  workflow_group_id = stackguardian_workflow_group.frontend.resource_name
  id                = "ONBOARDING-Project01-Frontend-Deploy"
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
