package workflowsteptemplaterevision_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/StackGuardian/terraform-provider-stackguardian/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func testAccWorkflowStepTemplateRevisionConfig(templateName string, revisionAlias string) string {
	return fmt.Sprintf(`
resource "stackguardian_workflow_step_template" "test" {
  template_name = "%s"
  is_active     = "0"
  is_public     = "0"
  description   = "Test template for revision"

  source_config_kind = "DOCKER_IMAGE"

  runtime_source = {
    source_config_dest_kind = "CONTAINER_REGISTRY"
    config = {
      docker_image = "ubuntu:latest"
      is_private   = false
    }
  }
}

resource "stackguardian_workflow_step_template_revision" "test" {
  template_id = stackguardian_workflow_step_template.test.id
  alias       = "%s"
  notes       = "Test revision notes"

  source_config_kind = "DOCKER_IMAGE"

  runtime_source = {
    source_config_dest_kind = "CONTAINER_REGISTRY"
    config = {
      docker_image = "ubuntu:20.04"
      is_private   = false
    }
  }
}
`, templateName, revisionAlias)
}

func TestAccWorkflowStepTemplateRevision_Basic(t *testing.T) {
	templateName := "provider-test-workflow-step-template1"
	revisionAlias := "v1"

	testAccResource := testAccWorkflowStepTemplateRevisionConfig(templateName, revisionAlias)
	customHeader := http.Header{}
	customHeader.Set("x-sg-internal-auth-orgid", "sg-provider-test")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader),
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccResource,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_step_template_revision.test", "alias", revisionAlias),
					resource.TestCheckResourceAttr("stackguardian_workflow_step_template_revision.test", "notes", "Test revision notes"),
					resource.TestCheckResourceAttr("stackguardian_workflow_step_template_revision.test", "source_config_kind", "DOCKER_IMAGE"),
					resource.TestCheckResourceAttrSet("stackguardian_workflow_step_template_revision.test", "id"),
					resource.TestCheckResourceAttrSet("stackguardian_workflow_step_template_revision.test", "template_id"),
				),
			},
		},
	})
}
