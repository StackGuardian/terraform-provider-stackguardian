package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const testAccCheckConfig_ResourceSgConnectorVcs = `
resource "stackguardian_connector_vcs" "TPS-Test-ConnectorVcs" {
	// integrationgroup = "TPS-Test"
	data = jsonencode({
	"ResourceName": "TPS-Test-ConnectorVcs",
	"ResourceType": "INTEGRATION.GITLAB_COM",
	"Tags" : ["tf-provider-test"]
	"Description": "Test of terraform-provider-stackguardian for ConnectorVcs",
	"Settings": {
		"kind": "GITLAB_COM",
		"config": [
			{
				"gitlabCreds": "test-user:test-token"
			}
		]
	},
	})
}
`

func TestAcc_ResourceSgConnectorVcs(t *testing.T) {
	t.Skipf("TODO: Fix DELETE: deletion of ConnectorVcs resource is not possible with API Key")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckConfig_ResourceSgConnectorVcs,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"stackguardian_connector_vcs.TPS-Test-ConnectorVcs",
						"id",
						"TPS-Test-ConnectorVcs",
					),
				),
			},
		},
	})
}
