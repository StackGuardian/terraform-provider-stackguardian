package workflowsteptemplate_test

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/StackGuardian/terraform-provider-stackguardian/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

var org = os.Getenv("STACKGUARDIAN_ORG_NAME")

// deleteWorkflowStepTemplateFixture is a safety-net cleanup: Terraform's own destroy step
// tears the template down in the normal case. This exists so a test that fails before
// reaching destroy (e.g. a failed assertion) doesn't leave it behind. Errors are ignored —
// the resource may already be gone. id is the template's id, which defaults to its
// template_name when the config doesn't set an explicit id.
func deleteWorkflowStepTemplateFixture(id string) {
	client := acctest.SGClient()
	client.WorkflowStepTemplate.DeleteWorkflowStepTemplate(context.TODO(), org, id)
}

func testAccWorkflowStepTemplateConfigWithRuntime(name string) string {
	return fmt.Sprintf(`
resource "stackguardian_workflow_step_template" "test" {
  template_name = "%s"
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
	name := acctest.ResourceName("example-workflow-step-template1")

	t.Cleanup(func() { deleteWorkflowStepTemplateFixture(name) })

	var testAccResource = fmt.Sprintf(`
resource "stackguardian_workflow_step_template" "test" {
  template_name = "%s"
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
					resource.TestCheckResourceAttr("stackguardian_workflow_step_template.test", "is_public", "0"),
				),
			},
			// Delete testing automatically occurs
		},
	})
}

func TestAccWorkflowStepTemplate_WithRuntime(t *testing.T) {
	name := acctest.ResourceName("example-workflow-step-template2")

	t.Cleanup(func() { deleteWorkflowStepTemplateFixture(name) })

	customHeader := http.Header{}
	customHeader.Set("x-sg-internal-auth-orgid", "sg-provider-test")
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader),
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
