package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const testAccCheckConfig_ResourceSgWorkflow = `
resource "stackguardian_workflow" "TPS-Test-Workflow" {
	wfgrp = "TPS-Test"

	data = jsonencode({
	  "ResourceName": "TPS-Test-Workflow",
	  "Description": "Test of terraform-provider-stackguardian for Workflow: Deploy a website from AWS S3",
	  "Tags": ["tf-provider-test"],
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
`

func TestAcc_ResourceSgWorkflow(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckConfig_ResourceSgWorkflow,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"stackguardian_workflow.TPS-Test-Workflow",
						"id",
						"TPS-Test-Workflow",
					),
					resource.TestCheckResourceAttr(
						"stackguardian_workflow.TPS-Test-Workflow",
						"wfgrp",
						"TPS-Test",
					),
				),
			},
		},
	})
}
