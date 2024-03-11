package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const testAccCheckConfig_ResourceSgSecret = `
resource "stackguardian_secret" "TPS-Test-Secret-Name" {
	data = jsonencode({
		"ResourceName":  "TPS-Test-Secret-Name",
		"ResourceValue": "TPS-Test-Secret-Value"
	})
}
`

func TestAcc_ResourceSgSecret(t *testing.T) {
	//t.Skipf("TODO: Fix DELETE: deletion of Secret resource is not possible with API Key")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckConfig_ResourceSgSecret,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"stackguardian_secret.TPS-Test-Secret-Name",
						"id",
						"TPS-Test-Secret-Name",
					),
				),
				//Destroy: true,
				//PreventPostDestroyRefresh: true,
			},
		},
	})
}
