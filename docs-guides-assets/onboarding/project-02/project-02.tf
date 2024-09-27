// Project-02

terraform {
  required_providers {
    stackguardian = {
      source  = "terraform/provider/stackguardian"
      version = "0.0.1"
    }
  }
}

provider "stackguardian" {
  api_key  = ""
  org_name = ""
  api_uri  = ""
}


resource "stackguardian_role" "ONBOARDING-Project02-Manager-Frontend" {
    resource_name = "ONBOARDING-Project02-Manager-Frontend"
    description = "Onboarding example of terraform-provider-stackguardian for Role Manager of Frontend team" 
    tags = ["tf-provider-example", "onboarding"]
    allowed_permissions = {
      "GET/api/v1/orgs/<org>/wfgrps/<wfGrp>/" : {
        "name" : "GetWorkflowGroup",
        "paths" : {
          "<wfGrp>" : [
            resource.stackguardian_workflow_group.ONBOARDING-Project02-Frontend.resource_name,
          ]
        }
      },
    }
}

resource "stackguardian_role" "ONBOARDING-Project02-Developer-Frontend" {
    resource_name = "ONBOARDING-Project02-Developer-Frontend"
    description = "Onboarding example of terraform-provider-stackguardian for Role Developer of Frontend team" 
    tags = ["tf-provider-example", "onboarding"]
    allowed_permissions = {
      "GET/api/v1/orgs/<org>/wfgrps/<wfGrp>/" : {
        "name" : "GetWorkflowGroup",
        "paths" : {
          "<wfGrp>" : [
            resource.stackguardian_workflow_group.ONBOARDING-Project02-Frontend.resource_name,
          ]
        }
      },
    }
}

resource "stackguardian_role" "ONBOARDING-Project02-Manager-Backend" {
    resource_name = "ONBOARDING-Project02-Manager-Backend"
    description = "Onboarding example of terraform-provider-stackguardian for Role Manager of Backend team" 
    tags = ["tf-provider-example", "onboarding"]
    allowed_permissions = {
      "GET/api/v1/orgs/<org>/wfgrps/<wfGrp>/" : {
        "name" : "GetWorkflowGroup",
        "paths" : {
          "<wfGrp>" : [
            resource.stackguardian_workflow_group.ONBOARDING-Project02-Backend.resource_name,
          ]
        }
      },
    }
}

resource "stackguardian_role" "ONBOARDING-Project02-Developer-Backend" {
    resource_name = "ONBOARDING-Project02-Developer-Backend"
    description = "Onboarding example of terraform-provider-stackguardian for Role Developer of Backend team" 
    tags = ["tf-provider-example", "onboarding"]
    allowed_permissions = {
      "GET/api/v1/orgs/<org>/wfgrps/<wfGrp>/" : {
        "name" : "GetWorkflowGroup",
        "paths" : {
          "<wfGrp>" : [
            resource.stackguardian_workflow_group.ONBOARDING-Project02-Backend.resource_name,
          ]
        }
      },
    }
}

resource "stackguardian_role" "ONBOARDING-Project02-Developer-DevOps" {
    resource_name = "ONBOARDING-Project02-Developer-DevOps"
    description = "Onboarding example of terraform-provider-stackguardian for Role Developer of DevOps team" 
    tags = ["tf-provider-example", "onboarding"]
    allowed_permissions = {
      "GET/api/v1/orgs/<org>/wfgrps/<wfGrp>/" : {
        "name" : "GetWorkflowGroup",
        "paths" : {
          "<wfGrp>" : [
            resource.stackguardian_workflow_group.ONBOARDING-Project02-DevOps.resource_name,
          ]
        }
      },
    }
}

resource "stackguardian_workflow_group" "ONBOARDING-Project02-Backend" {
    resource_name = "ONBOARDING-Project02-Backend"
    description = "Onboarding example of terraform-provider-stackguardian for WorkflowGroup for Backend team"
    tags = ["tf-provider-example", "onboarding"]  
}

resource "stackguardian_workflow_group" "ONBOARDING-Project02-Frontend" {
    resource_name = "ONBOARDING-Project02-Frontend"
    description = "Onboarding example of terraform-provider-stackguardian for WorkflowGroup for Frontend team"
    tags = ["tf-provider-example", "onboarding"]
}

resource "stackguardian_workflow_group" "ONBOARDING-Project02-DevOps" {
    resource_name = "ONBOARDING-Project02-DevOps"
    description = "Onboarding example of terraform-provider-stackguardian for WorkflowGroup for DevOps team"
    tags = ["tf-provider-example", "onboarding"]
}

resource "stackguardian_role_assignment" "ONBOARDING-Project02-Frontend-Manager" {
  user_id = "frontend.manager.p02@dummy.com"
  entity_type = "EMAIL"
  role = resource.stackguardian_role.ONBOARDING-Project02-Manager-Frontend.resource_name
}

resource "stackguardian_role_assignment" "ONBOARDING-Project02-Backend-Developer" {
  user_id = "backend.developer.p02@dummy.com"
  entity_type = "EMAIL"
  role = resource.stackguardian_role.ONBOARDING-Project02-Developer-Backend.resource_name
}

resource "stackguardian_role_assignment" "ONBOARDING-Project02-DevOps-Developer" {
  user_id = "devops.developer.p02@dummy.com"
  entity_type = "EMAIL"
  role = resource.stackguardian_role.ONBOARDING-Project02-Developer-DevOps.resource_name
}

#Commented until connectors is ready for testing

# resource "stackguardian_connector" "ONBOARDING-Project02-Cloud-Connector" {
#     resource_name = "ONBOARDING-Project02"
#     tags = ["tf-provider-example", "onboarding"]
#     description = "Onboarding example  of terraform-provider-stackguardian for ConnectorCloud"
#     settings = {
#       kind = "AWS_STATIC",
#       config = jsonencode([
#         {
#           "awsAccessKeyId" : "REPLACEME-aws-key",
#           "awsSecretAccessKey" : "REPLACEME-aws-key",
#           "awsDefaultRegion" : "REPLACEME-us-west-2"
#         }
#       ])
#     }
# }

# resource "stackguardian_connector" "ONBOARDING-Project02-VCS-Connector" {
#     resource_name = "ONBOARDING-Project02"
#     tags = ["tf-provider-example", "onboarding"]
#     description = "Onboarding example of terraform-provider-stackguardian for ConnectorVcs"
#     settings = {
#       kind = "GITHUB_COM",
#       config = jsonencode([
#         {
#           "gitlabCreds" : "REPLACEME-example-user:REPLACEME-example-token"
#         }
#       ])
#     }
# }
