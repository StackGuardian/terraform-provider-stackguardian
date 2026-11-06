package workflowsteptemplaterevision_test

import (
	"context"
	"fmt"
	"testing"

	sgclient "github.com/StackGuardian/sg-sdk-go/client"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

const testAccCRUDConfig = `
resource "stackguardian_workflow_step_template" "parent" {
  template_name      = %q
  is_public          = "0"
  source_config_kind = "DOCKER_IMAGE"

  runtime_source = {
    source_config_dest_kind = "CONTAINER_REGISTRY"
    config = {
      docker_image = "ubuntu:latest"
      is_private   = false
    }
  }
}

resource "stackguardian_workflow_step_template_revision" "this" {
  template_id        = stackguardian_workflow_step_template.parent.id
  alias              = "crud"
  notes              = %q
  source_config_kind = "DOCKER_IMAGE"

  runtime_source = {
    source_config_dest_kind = "CONTAINER_REGISTRY"
    config = {
      docker_image = %q
      is_private   = false
    }
  }
}`

func TestAccWorkflowStepTemplateRevision_CRUD(t *testing.T) {
	templateName := acctest.ResourceName("crud-step-template-rev")
	// Two revisions: the gone step recreates the first one as :2.
	t.Cleanup(func() {
		for i := 1; i <= 2; i++ {
			deprecateStepTemplateRevisionFixture(fmt.Sprintf("%s:%d", templateName, i))
			deleteStepTemplateRevisionFixture(fmt.Sprintf("%s:%d", templateName, i))
		}
		deleteStepTemplateFixture(templateName)
	})

	acctest.RunCRUD(t, acctest.CRUD{
		ResourceType: "stackguardian_workflow_step_template_revision",
		Address:      "stackguardian_workflow_step_template_revision.this",
		Create:       fmt.Sprintf(testAccCRUDConfig, templateName, "created", "ubuntu:22.04"),
		Update:       fmt.Sprintf(testAccCRUDConfig, templateName, "updated", "ubuntu:24.04"),
		UpdateChecks: []statecheck.StateCheck{
			statecheck.ExpectKnownValue("stackguardian_workflow_step_template_revision.this",
				tfjsonpath.New("notes"), knownvalue.StringExact("updated")),
			statecheck.ExpectKnownValue("stackguardian_workflow_step_template_revision.this",
				tfjsonpath.New("runtime_source").AtMapKey("config").AtMapKey("docker_image"),
				knownvalue.StringExact("ubuntu:24.04")),
		},
		Header: acctest.TemplateHeader(),
		Read: func(ctx context.Context, c *sgclient.Client, org string, rs *terraform.ResourceState) error {
			_, err := c.WorkflowStepTemplateRevision.ReadWorkflowStepTemplateRevision(ctx, org, rs.Primary.ID)
			return err
		},
		Delete: func(ctx context.Context, c *sgclient.Client, org string, rs *terraform.ResourceState) error {
			return c.WorkflowStepTemplateRevision.DeleteWorkflowStepTemplateRevision(ctx, org, rs.Primary.ID, true)
		},
	})
}
