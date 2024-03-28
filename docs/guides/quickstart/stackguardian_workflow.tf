terraform {
  required_providers {
    stackguardian = {
      source = "terraform/provider/stackguardian"

      # https://developer.hashicorp.com/terraform/language/expressions/version-constraints#version-constraint-behavior
      # NOTE: A prerelease version can be selected only by an exact version constraint.
      version = "0.0.0-rc1" #provider-version
    }
  }
}

/*
The provider configuration should be passed from external environment variables:
```
$ export STACKGUARDIAN_ORG_NAME="YOUR_SG_ORG"
$ export STACKGUARDIAN_API_KEY="YOUR_SG_KEY"
```
*/

provider "stackguardian" {}



resource "stackguardian_workflow_group" "TPS-Quickstart-T000000" {
  data = jsonencode({
    "ResourceName" : "TPS-Quickstart-T000000",
    "Description" : "Example of StackGuardian Workflow Group",
    "Tags" : ["tf-provider-example"],
    "IsActive" : 1,
  })
}


resource "stackguardian_workflow" "TPS-Quickstart-Workflow_WebsiteS3-T000000" {
  wfgrp = stackguardian_workflow_group.TPS-Quickstart-T000000.id

  data = jsonencode({
    "ResourceName": "TPS-Quickstart-Workflow_WebsiteS3-T000000",
    "Description": "Example of StackGuardian Workflow: Deploy a website from AWS S3",
    "Tags": ["tf-provider-example"],
    "EnvironmentVariables": [],
    "DeploymentPlatformConfig": [{
      "kind": "AWS_RBAC",
      "config": {
        "integrationId": "/integrations/aws"
      }
    }],
    "VCSConfig": {
      "iacVCSConfig": {
        "useMarketplaceTemplate": true,
        "iacTemplate": "/stackguardian/aws-s3-demo-website",
        "iacTemplateId": "/stackguardian/aws-s3-demo-website:4"
      },
      "iacInputData": {
        "schemaType": "FORM_JSONSCHEMA",
        "data": {
          "shop_name": "StackGuardian",
          "bucket_region": "eu-central-1"
        }
      }
    },
    "Approvers": [],
    "TerraformConfig": {
      "managedTerraformState": true,
      "terraformVersion": "1.4.6"
    },
    "WfType": "TERRAFORM",
    "UserSchedules": []
  })
}
