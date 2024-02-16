terraform {
  required_providers {
    stackguardian = {
      source = "terraform/provider/stackguardian"
      version = "0.1.0-beta1"
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


resource "stackguardian_tf_provider_workflow" "Workflow_DeployWebsiteS3" {
  wfgrp = "WorkflowGroup_DeployWebsiteS3"

  data = jsonencode({
    "ResourceName": "Workflow_DeployWebsiteS3",
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
        "iacTemplateId": "/stackguardian/aws-s3-demo-website:11"
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
