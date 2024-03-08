package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const testAccCheckConfig_ResourceSgConnectorCloud = `
resource "stackguardian_connector_cloud" "TPS-Test-ConnectorCloud" {
	// integrationgroup = "TPS-Test"
	data = jsonencode({
	"ResourceName": "TPS-Test-ConnectorCloud",
	"Tags" : ["tf-provider-test"]
	"Description": "Test of terraform-provider-stackguardian for ConnectorCloud",
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

func TestAcc_ResourceSgConnectorCloud(t *testing.T) {
	t.Skipf("TODO: Fix DELETE: deletion of ConnectorCloud resource is not possible with API Key")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckConfig_ResourceSgConnectorCloud,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"stackguardian_connector_cloud.TPS-Test-ConnectorCloud",
						"id",
						"TPS-Test-ConnectorCloud",
					),
				),
			},
		},
	})
}
