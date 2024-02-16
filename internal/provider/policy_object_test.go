package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const testAccCheckConfig_ResourceSgPolicy = `
resource "stackguardian_policy" "TPS-Test-Policy" {
	data = jsonencode({
		"ResourceName" : "TPS-Test-Policy",
		"Description" : "Test of terraform-provider-stackguardian for Policy",
		"Tags" : ["tf-provider-test", "test", "policy"]
	})
}
`

func TestAcc_ResourceSgPolicy(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckConfig_ResourceSgPolicy,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"stackguardian_policy.TPS-Test-Policy",
						"id",
						"TPS-Test-Policy",
					),
				),
			},
		},
	})
}
