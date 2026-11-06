package workflowtemplate_test

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

func TestAccWorkflowTemplate_CRUD(t *testing.T) {
	name := acctest.ResourceName("crud-wf-template")
	t.Cleanup(func() { deleteWorkflowTemplateFixture(name) })

	body := func(isPublic, description, tags string) string {
		return fmt.Sprintf(`
		  is_public   = %q
		  description = %q
		  tags        = %s
		`, isPublic, description, tags)
	}

	acctest.RunCRUD(t, acctest.CRUD{
		ResourceType: "stackguardian_workflow_template",
		Address:      "stackguardian_workflow_template.test",
		Create:       testAccWorkflowTemplate(name, sourceConfigKind, body("0", "created", `["crud"]`)),
		Update:       testAccWorkflowTemplate(name, sourceConfigKind, body("1", "updated", `["crud", "updated"]`)),
		UpdateChecks: []statecheck.StateCheck{
			statecheck.ExpectKnownValue("stackguardian_workflow_template.test",
				tfjsonpath.New("description"), knownvalue.StringExact("updated")),
			statecheck.ExpectKnownValue("stackguardian_workflow_template.test",
				tfjsonpath.New("is_public"), knownvalue.StringExact("1")),
		},
		Header: acctest.TemplateHeader(),
		Read: func(ctx context.Context, c *sgclient.Client, org string, rs *terraform.ResourceState) error {
			_, err := c.WorkflowTemplates.ReadWorkflowTemplate(ctx, org, rs.Primary.ID)
			return err
		},
		Delete: func(ctx context.Context, c *sgclient.Client, org string, rs *terraform.ResourceState) error {
			return c.WorkflowTemplates.DeleteWorkflowTemplate(ctx, org, rs.Primary.ID)
		},
	})
}
