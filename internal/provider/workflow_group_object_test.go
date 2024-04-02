package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const testAccCheckConfig_ResourceSgWorkflowGroup = `
resource "stackguardian_workflow_group" "TPS-Test-WorkflowGroup" {

	data = jsonencode({
	  "ResourceName": "TPS-Test-WorkflowGroup",
	  "Description": "Test of terraform-provider-stackguardian for WorkflowGroup",
	  "Tags": ["tf-provider-test"],
	  "IsActive": 1,
	})
  }
`

func TestAcc_ResourceSgWorkflowGroup(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckConfig_ResourceSgWorkflowGroup,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"stackguardian_workflow_group.TPS-Test-WorkflowGroup",
						"id",
						"TPS-Test-WorkflowGroup",
					),
				),
			},
		},
	})
}
