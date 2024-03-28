package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const testAccCheckConfig_ResourceSgRole = `
resource "stackguardian_role" "TPS-Test-Role" {

	data = jsonencode({
		"ResourceName": "TPS-Test-Role",
		//"Description": "Test of terraform-provider-stackguardian for Role", // TODO: Uncomment after fix in Frontend
		"Tags": ["tf-provider-test"],
		"Actions": [
			"Action-1"
		],
		"AllowedPermissions": {
			"Permission-key-1": "Permission-val-1",
			"Permission-key-2": "Permission-val-2"
		}
	})
}
`

func TestAcc_ResourceSgRole(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckConfig_ResourceSgRole,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"stackguardian_role.TPS-Test-Role",
						"id",
						"TPS-Test-Role",
					),
				),
			},
		},
	})
}
