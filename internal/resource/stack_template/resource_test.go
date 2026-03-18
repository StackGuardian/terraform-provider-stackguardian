package stacktemplate_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/StackGuardian/terraform-provider-stackguardian/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

var sourceConfigKind = "TERRAFORM"

func TestAccStackTemplate_Basic(t *testing.T) {
	templateName := "tf-provider-stack-template-1"

	customHeader := http.Header{}
	customHeader.Set("x-sg-internal-auth-orgid", "sg-provider-test")

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader),
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccStackTemplateConfig(templateName, sourceConfigKind),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_stack_template.test", "template_name", templateName),
					resource.TestCheckResourceAttr("stackguardian_stack_template.test", "is_public", "0"),
					resource.TestCheckResourceAttrSet("stackguardian_stack_template.test", "id"),
				),
			},
			// Update and Read testing
			{
				Config: testAccStackTemplateConfigUpdated(templateName, sourceConfigKind),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_stack_template.test", "template_name", templateName),
					resource.TestCheckResourceAttr("stackguardian_stack_template.test", "description", "Updated description"),
				),
			},
			// Delete testing automatically occurs
		},
	})
}

func testAccStackTemplateConfig(name, sourceConfigKind string) string {
	return fmt.Sprintf(`
resource "stackguardian_stack_template" "test" {
  template_name    = "%s"
  source_config_kind = "%s"
  is_public        = "0"
  tags             = ["test", "terraform"]
}
`, name, sourceConfigKind)
}

func testAccStackTemplateConfigUpdated(name, sourceConfigKind string) string {
	return fmt.Sprintf(`
resource "stackguardian_stack_template" "test" {
  template_name    = "%s"
  source_config_kind = "%s"
  is_public        = "1"
  description      = "Updated description"
  tags             = ["test", "terraform", "updated"]
}
`, name, sourceConfigKind)
}
