package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const testAccCheckConfig_ResourceSgIntegration = `
resource "stackguardian_integration" "TPS-Test-Integration" {
	data = jsonencode({
	"ResourceName": "TPS-Test-Integration",
	// "Tags" : ["tf-provider-test"]
	"Description": "Test of terraform-provider-stackguardian for Integration",
	"Settings": {
		"kind": "AWS_STATIC",
		"config": [
			{
				"awsAccessKeyId": "test-aws-key",
				"awsSecretAccessKey": "test-aws-key",
				"awsDefaultRegion": "us-west-2"
			}
		]
	}
	})
}
`

func TestAcc_ResourceSgIntegration(t *testing.T) {
	t.Skipf("TODO: Fix DELETE: deletion of Integration resource is not possible with API Key")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckConfig_ResourceSgIntegration,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"stackguardian_integration.TPS-Test-Integration",
						"id",
						"TPS-Test-Integration",
					),
				),
			},
		},
	})
}
