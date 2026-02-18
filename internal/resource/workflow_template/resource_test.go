package workflowtemplate_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/StackGuardian/terraform-provider-stackguardian/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

var sourceConfigKind = "TERRAFORM"

func TestAccWorkflowTemplate_Basic(t *testing.T) {
	templateName := "tf-provider-workflow-template-1"

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
				Config: testAccWorkflowTemplateConfig(templateName, sourceConfigKind),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_template.test", "template_name", templateName),
					resource.TestCheckResourceAttr("stackguardian_workflow_template.test", "is_public", "0"),
					resource.TestCheckResourceAttrSet("stackguardian_workflow_template.test", "id"),
				),
			},
			// Update and Read testing
			{
				Config: testAccWorkflowTemplateConfigUpdated(templateName, sourceConfigKind),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_template.test", "template_name", templateName),
					resource.TestCheckResourceAttr("stackguardian_workflow_template.test", "description", "Updated description"),
				),
			},
			// Delete testing automatically occurs
		},
	})
}

func TestAccWorkflowTemplate_WithRuntime(t *testing.T) {
	templateName := "tf-provider-workflow-template-2"

	customHeader := http.Header{}
	customHeader.Set("x-sg-internal-auth-orgid", "sg-provider-test")

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader),
		Steps: []resource.TestStep{
			// Create and Read testing with runtime source
			{
				Config: testAccWorkflowTemplateConfigWithRuntime(templateName, sourceConfigKind),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_template.test", "template_name", templateName),
					resource.TestCheckResourceAttrSet("stackguardian_workflow_template.test", "runtime_source.source_config_dest_kind"),
				),
			},
			// Delete testing automatically occurs
		},
	})
}

func testAccWorkflowTemplateConfig(name, sourceConfigKind string) string {
	return fmt.Sprintf(`
resource "stackguardian_workflow_template" "test" {
  template_name = "%s"
  source_config_kind = "%s"
  is_public     = "0"
  tags          = ["test", "terraform"]
}
`, name, sourceConfigKind)
}

func testAccWorkflowTemplateConfigUpdated(name, sourceConfigKind string) string {
	return fmt.Sprintf(`
resource "stackguardian_workflow_template" "test" {
  template_name = "%s"
  source_config_kind = "%s"
  is_public     = "1"
  description   = "Updated description"
  tags          = ["test", "terraform", "updated"]
}
`, name, sourceConfigKind)
}

func testAccWorkflowTemplateConfigWithRuntime(name, sourceConfigKind string) string {
	return fmt.Sprintf(`
resource "stackguardian_workflow_template" "test" {
  template_name = "%s"
  source_config_kind = "%s"
  is_public     = "0"
  tags          = ["test", "terraform"]

  runtime_source = {
    source_config_dest_kind = "GITHUB_COM"
    config = {
      is_private   = false
	  repo = "https://github.com/taherkk/taher-null-resource.git"
    }
  }
}
`, name, sourceConfigKind)
}
