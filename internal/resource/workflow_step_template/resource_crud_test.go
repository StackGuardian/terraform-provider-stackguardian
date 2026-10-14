package workflowsteptemplate_test

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
resource "stackguardian_workflow_step_template" "this" {
  template_name      = %q
  is_public          = "0"
  description        = %q
  tags               = %s
  source_config_kind = "DOCKER_IMAGE"

  runtime_source = {
    source_config_dest_kind = "CONTAINER_REGISTRY"
    config = {
      docker_image = %q
      is_private   = false
    }
  }
}`

func TestAccWorkflowStepTemplate_CRUD(t *testing.T) {
	name := acctest.ResourceName("crud-step-template")
	t.Cleanup(func() { deleteWorkflowStepTemplateFixture(name) })

	acctest.RunCRUD(t, acctest.CRUD{
		ResourceType: "stackguardian_workflow_step_template",
		Address:      "stackguardian_workflow_step_template.this",
		Create:       fmt.Sprintf(testAccCRUDConfig, name, "created", `["crud"]`, "ubuntu:22.04"),
		Update:       fmt.Sprintf(testAccCRUDConfig, name, "updated", `["crud", "updated"]`, "ubuntu:24.04"),
		UpdateChecks: []statecheck.StateCheck{
			statecheck.ExpectKnownValue("stackguardian_workflow_step_template.this",
				tfjsonpath.New("description"), knownvalue.StringExact("updated")),
			statecheck.ExpectKnownValue("stackguardian_workflow_step_template.this",
				tfjsonpath.New("runtime_source").AtMapKey("config").AtMapKey("docker_image"),
				knownvalue.StringExact("ubuntu:24.04")),
		},
		Header: acctest.TemplateHeader(),
		Read: func(ctx context.Context, c *sgclient.Client, org string, rs *terraform.ResourceState) error {
			_, err := c.WorkflowStepTemplate.ReadWorkflowStepTemplate(ctx, org, rs.Primary.ID)
			return err
		},
		Delete: func(ctx context.Context, c *sgclient.Client, org string, rs *terraform.ResourceState) error {
			return c.WorkflowStepTemplate.DeleteWorkflowStepTemplate(ctx, org, rs.Primary.ID)
		},
	})
}
