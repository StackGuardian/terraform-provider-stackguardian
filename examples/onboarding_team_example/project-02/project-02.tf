// Project-02

terraform {
  required_providers {
    stackguardian = {
      source  = "terraform/provider/stackguardian"
      version = "0.0.0-dev"
    }
  }
}

provider "stackguardian" {}


resource "stackguardian_role" "ONBOARDING-Project02-Manager-Frontend" {
  data = jsonencode({
    "ResourceName" : "ONBOARDING-Project02-Manager-Frontend",
    "Description" : "Onboarding example of terraform-provider-stackguardian for Role Manager of Frontend team" ,
    "Tags" : ["tf-provider-example", "onboarding"],
    "Actions" : [
      "REPLACEME-Action-1"
    ],
    "AllowedPermissions" : {
      "REPLACEME-Permission-key-1" : "REPLACEME-Permission-val-1",
      "REPLACEME-Permission-key-2" : "REPLACEME-Permission-val-2"
    }
  })
}

resource "stackguardian_role" "ONBOARDING-Project02-Developer-Frontend" {
  data = jsonencode({
    "ResourceName" : "ONBOARDING-Project02-Developer-Frontend",
    "Description" : "Onboarding example of terraform-provider-stackguardian for Role Developer of Frontend team" ,
    "Tags" : ["tf-provider-example", "onboarding"],
    "Actions" : [
      "REPLACEME-Action-1"
    ],
    "AllowedPermissions" : {
      "REPLACEME-Permission-key-1" : "REPLACEME-Permission-val-1",
      "REPLACEME-Permission-key-2" : "REPLACEME-Permission-val-2"
    }
  })
}

resource "stackguardian_role" "ONBOARDING-Project02-Manager-Backend" {
  data = jsonencode({
    "ResourceName" : "ONBOARDING-Project02-Manager-Backend",
    "Description" : "Onboarding example of terraform-provider-stackguardian for Role Manager of Backend team" ,
    "Tags" : ["tf-provider-example", "onboarding"],
    "Actions" : [
      "REPLACEME-Action-1"
    ],
    "AllowedPermissions" : {
      "REPLACEME-Permission-key-1" : "REPLACEME-Permission-val-1",
      "REPLACEME-Permission-key-2" : "REPLACEME-Permission-val-2"
    }
  })
}

resource "stackguardian_role" "ONBOARDING-Project02-Developer-Backend" {
  data = jsonencode({
    "ResourceName" : "ONBOARDING-Project02-Developer-Backend",
    "Description" : "Onboarding example of terraform-provider-stackguardian for Role Developer of Backend team" ,
    "Tags" : ["tf-provider-example", "onboarding"],
    "Actions" : [
      "REPLACEME-Action-1"
    ],
    "AllowedPermissions" : {
      "REPLACEME-Permission-key-1" : "REPLACEME-Permission-val-1",
      "REPLACEME-Permission-key-2" : "REPLACEME-Permission-val-2"
    }
  })
}

resource "stackguardian_role" "ONBOARDING-Project02-Developer-DevOps" {
  data = jsonencode({
    "ResourceName" : "ONBOARDING-Project02-Developer-DevOps",
    "Description" : "Onboarding example of terraform-provider-stackguardian for Role Developer of DevOps team" ,
    "Tags" : ["tf-provider-example", "onboarding"],
    "Actions" : [
      "REPLACEME-Action-1"
    ],
    "AllowedPermissions" : {
      "REPLACEME-Permission-key-1" : "REPLACEME-Permission-val-1",
      "REPLACEME-Permission-key-2" : "REPLACEME-Permission-val-2"
    }
  })
}

resource "stackguardian_workflow_group" "ONBOARDING-Project02-Backend" {
  data = jsonencode({
    "ResourceName" : "ONBOARDING-Project02-Backend",
    "Description" : "Onboarding example of terraform-provider-stackguardian for WorkflowGroup for Backend team",
    "Tags" : ["tf-provider-example", "onboarding"],
    "IsActive" : 1,
  })
}

resource "stackguardian_workflow_group" "ONBOARDING-Project02-Frontend" {
  data = jsonencode({
    "ResourceName" : "ONBOARDING-Project02-Frontend",
    "Description" : "Onboarding example of terraform-provider-stackguardian for WorkflowGroup for Frontend team",
    "Tags" : ["tf-provider-example", "onboarding"],
    "IsActive" : 1,
  })
}

resource "stackguardian_workflow_group" "ONBOARDING-Project02-DevOps" {
  data = jsonencode({
    "ResourceName" : "ONBOARDING-Project02-DevOps",
    "Description" : "Onboarding example of terraform-provider-stackguardian for WorkflowGroup for DevOps team",
    "Tags" : ["tf-provider-example", "onboarding"],
    "IsActive" : 1,
  })
}

resource "stackguardian_policy" "ONBOARDING-Project02" {
  data = jsonencode({
    "ResourceName" : "ONBOARDING-Project02-Policy",
    "Description" : "Onboarding example  of terraform-provider-stackguardian for Policy",
    "Tags" : ["tf-provider-example", "onboarding"]
  })
}

resource "stackguardian_connector_cloud" "ONBOARDING-Project02" {
  data = jsonencode({
    "ResourceName" : "ONBOARDING-Project02",
    "Tags" : ["tf-provider-example", "onboarding"]
    "Description" : "Onboarding example  of terraform-provider-stackguardian for ConnectorCloud",
    "Settings" : {
      "kind" : "AWS_STATIC",
      "config" : [
        {
          "awsAccessKeyId" : "REPLACEME-aws-key",
          "awsSecretAccessKey" : "REPLACEME-aws-key",
          "awsDefaultRegion" : "REPLACEME-us-west-2"
        }
      ]
    }
  })
}

resource "stackguardian_connector_vcs" "ONBOARDING-Project02" {
  data = jsonencode({
    "ResourceName" : "ONBOARDING-Project02",
    "ResourceType" : "INTEGRATION.GITLAB_COM",
    "Tags" : ["tf-provider-example", "onboarding"]
    "Description" : "Onboarding example of terraform-provider-stackguardian for ConnectorVcs",
    "Settings" : {
      "kind" : "GITLAB_COM",
      "config" : [
        {
          "gitlabCreds" : "REPLACEME-example-user:REPLACEME-example-token"
        }
      ]
    },
  })
}
