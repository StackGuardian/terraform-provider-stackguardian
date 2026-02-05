package workflowsteptemplate_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/StackGuardian/terraform-provider-stackguardian/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func testAccWorkflowStepTemplateConfigWithRuntime(name string) string {
	return fmt.Sprintf(`
resource "stackguardian_workflow_step_template" "test" {
  template_name = "%s"
  is_active     = "0"
  is_public     = "0"
  description   = "Test with runtime source"
  
  source_config_kind = "DOCKER_IMAGE"

  runtime_source = {
    source_config_dest_kind = "CONTAINER_REGISTRY"
    config = {
      docker_image = "ubuntu:latest"
      is_private   = false
    }
  }
}
`, name)
}

func TestAccWorkflowStepTemplate_Basic(t *testing.T) {
	name := "example-workflow-step-template1"
	var testAccResource = fmt.Sprintf(`
resource "stackguardian_workflow_step_template" "test" {
  template_name = "%s"
  is_active     = "0"
  is_public     = "0"
  description   = "Test with runtime source"
  
  source_config_kind = "DOCKER_IMAGE"

  runtime_source = {
    source_config_dest_kind = "CONTAINER_REGISTRY"
    config = {
      docker_image = "ubuntu:latest"
      is_private   = false
    }
  }
}
`, name)
	customHeader := http.Header{}
	customHeader.Set("x-sg-internal-auth-orgid", "sg-provider-test")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader),
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccResource,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_step_template.test", "template_name", name),
					resource.TestCheckResourceAttr("stackguardian_workflow_step_template.test", "is_public", "0"),
					resource.TestCheckResourceAttrSet("stackguardian_workflow_step_template.test", "id"),
				),
			},
			// Update and Read testing
			{
				Config: testAccResource,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_step_template.test", "template_name", name),
					resource.TestCheckResourceAttr("stackguardian_workflow_step_template.test", "is_active", "0"),
				),
			},
			// Delete testing automatically occurs
		},
	})
}

func TestAccWorkflowStepTemplate_WithRuntime(t *testing.T) {
	name := "example-workflow-step-template2"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProviderFactories(http.Header{}),
		Steps: []resource.TestStep{
			// Create and Read testing with runtime source
			{
				Config: testAccWorkflowStepTemplateConfigWithRuntime(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_step_template.test", "template_name", name),
					resource.TestCheckResourceAttr("stackguardian_workflow_step_template.test", "runtime_source.source_config_dest_kind", "CONTAINER_REGISTRY"),
					resource.TestCheckResourceAttr("stackguardian_workflow_step_template.test", "runtime_source.config.docker_image", "ubuntu:latest"),
				),
			},
			// Delete testing automatically occurs
		},
	})
}
