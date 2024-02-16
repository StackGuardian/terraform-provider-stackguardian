package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const testAccCheckConfig_ResourceSgStack = `
resource "stackguardian_stack" "TPS-Test-Stack" {
	wfgrp = "TPS-Test"
	data = jsonencode({
	  "ResourceName": "TPS-Test-Stack",
	  "TemplatesConfig": {
		  "templateGroupId": "/stackguardian/terraform-aws-vpc-ec2:3",
		  "templates": [{
			  "id": 0,
			  "WfType": "TERRAFORM",
			  "ResourceName": "terraform-aws-vpc-stripped-vciK",
			  "Description": "",
			  "EnvironmentVariables": [],
			  "DeploymentPlatformConfig": [],
			  "RunnerConstraint": {
				  "type": "shared"
			  },
			  "TerraformConfig": {
				  "terraformVersion": "1.3.6",
				  "managedTerraformState": true,
				  "approvalPreApply": false,
				  "driftCheck": false
			  },
			  "VCSConfig": {
				  "iacVCSConfig": {
					  "useMarketplaceTemplate": true,
					  "iacTemplate": "/stackguardian/terraform-aws-vpc-stripped",
					  "iacTemplateId": "/stackguardian/terraform-aws-vpc-stripped:2"
				  },
				  "iacInputData": {
					  "schemaType": "FORM_JSONSCHEMA",
					  "data": {
						  "name": "NewVPC",
						  "public_subnets": ["10.0.1.0/24", "10.0.2.0/24"],
						  "cidr": "10.0.0.0/16",
						  "azs": ["eu-central-1a", "eu-central-1b"]
					  }
				  }
			  },
			  "MiniSteps": {
				  "wfChaining": {
					  "ERRORED": [],
					  "COMPLETED": []
				  },
				  "notifications": {
					  "email": {
						  "ERRORED": [],
						  "COMPLETED": [],
						  "APPROVAL_REQUIRED": [],
						  "CANCELLED": []
					  }
				  }
			  },
			  "Approvers": [],
			  "GitHubComSync": {
				  "pull_request_opened": {
					  "createWfRun": {
						  "enabled": false
					  }
				  }
			  },
			  "UserSchedules": []
		  }, {
			  "id": 1,
			  "WfType": "TERRAFORM",
			  "ResourceName": "terraform-azure-aks-stripped-oFa5",
			  "Description": "",
			  "EnvironmentVariables": [],
			  "DeploymentPlatformConfig": [],
			  "RunnerConstraint": {
				  "type": "shared"
			  },
			  "TerraformConfig": {
				  "terraformVersion": "1.3.6",
				  "managedTerraformState": true,
				  "approvalPreApply": false,
				  "driftCheck": false
			  },
			  "VCSConfig": {
				  "iacVCSConfig": {
					  "useMarketplaceTemplate": true,
					  "iacTemplate": "/stackguardian/terraform-azure-aks-stripped",
					  "iacTemplateId": "/stackguardian/terraform-azure-aks-stripped:5"
				  }
			  },
			  "MiniSteps": {
				  "wfChaining": {
					  "ERRORED": [],
					  "COMPLETED": []
				  },
				  "notifications": {
					  "email": {
						  "ERRORED": [],
						  "COMPLETED": [],
						  "APPROVAL_REQUIRED": [],
						  "CANCELLED": []
					  }
				  }
			  },
			  "Approvers": [],
			  "GitHubComSync": {
				  "pull_request_opened": {
					  "createWfRun": {
						  "enabled": false
					  }
				  }
			  },
			  "UserSchedules": []
		  }, {
			  "id": 2,
			  "WfType": "TERRAFORM",
			  "ResourceName": "terraform-aws-vpc-stripped-6Q7Y",
			  "Description": "",
			  "EnvironmentVariables": [],
			  "DeploymentPlatformConfig": [],
			  "RunnerConstraint": {
				  "type": "shared"
			  },
			  "TerraformConfig": {
				  "terraformVersion": "1.3.6",
				  "managedTerraformState": true,
				  "approvalPreApply": false,
				  "driftCheck": false
			  },
			  "VCSConfig": {
				  "iacVCSConfig": {
					  "useMarketplaceTemplate": true,
					  "iacTemplate": "/stackguardian/terraform-aws-vpc-stripped",
					  "iacTemplateId": "/stackguardian/terraform-aws-vpc-stripped:16"
				  }
			  },
			  "MiniSteps": {
				  "wfChaining": {
					  "ERRORED": [],
					  "COMPLETED": []
				  },
				  "notifications": {
					  "email": {
						  "ERRORED": [],
						  "COMPLETED": [],
						  "APPROVAL_REQUIRED": [],
						  "CANCELLED": []
					  }
				  }
			  },
			  "Approvers": [],
			  "GitHubComSync": {
				  "pull_request_opened": {
					  "createWfRun": {
						  "enabled": false
					  }
				  }
			  },
			  "UserSchedules": []
		  }]
	  }
  }
	)
}
`
const testAccCheckConfig_ResourceSgStack_Deps = `
resource "stackguardian_workflow" "terraform-aws-vpc-stripped-vciK" {
	wfgrp = "TPS-Test"
	stack = "TPS-Test-Stack"
	data = ""
}

resource "stackguardian_workflow" "terraform-azure-aks-stripped-oFa5" {
	wfgrp = "TPS-Test"
	stack = "TPS-Test-Stack"
	data = ""
}

resource "stackguardian_workflow" "terraform-aws-vpc-stripped-6Q7Y" {
	wfgrp = "TPS-Test"
	stack = "TPS-Test-Stack"
	data = ""
}
`

func TestAcc_ResourceSgStack(t *testing.T) {
	t.Skipf("TODO: Fix test: Dependencies must be deleted before")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckConfig_ResourceSgStack,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"stackguardian_stack.TPS-Test-Stack",
						"id",
						"TPS-Test-Stack",
					),
					resource.TestCheckResourceAttr(
						"stackguardian_stack.TPS-Test-Stack",
						"wfgrp",
						"TPS-Test",
					),
				),
				Destroy: false,
			},
			{
				Config: testAccCheckConfig_ResourceSgStack_Deps,
				//PlanOnly:           true,
				//ExpectNonEmptyPlan: true,
				Destroy: true,
			},
			// {
			// 	Config:             testAccCheckConfig_ResourceSgStack,
			// 	PlanOnly:           true,
			// 	ExpectNonEmptyPlan: true,
			// 	Destroy:            true,
			// },
		},
	})
}
