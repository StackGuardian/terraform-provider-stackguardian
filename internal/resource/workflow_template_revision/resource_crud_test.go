package workflowtemplaterevision_test

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

func TestAccWorkflowTemplateRevision_CRUD(t *testing.T) {
	acctest.SkipUnlessAcceptance(t)
	templateID := acctest.ResourceName("crud-wf-template-rev")

	// Two revisions: the gone step recreates the first one as :2.
	registerWorkflowTemplateCleanup(t, templateID, 2)
	if err := createWorkflowTemplateFixture(templateID, "TERRAFORM"); err != nil {
		t.Fatalf("creating parent workflow template: %s", err)
	}

	body := func(description, notes, tags string) string {
		return fmt.Sprintf(`
		  alias       = "crud"
		  is_public   = "0"
		  description = %q
		  notes       = %q
		  tags        = %s
		`, description, notes, tags)
	}

	acctest.RunCRUD(t, acctest.CRUD{
		ResourceType: "stackguardian_workflow_template_revision",
		Address:      "stackguardian_workflow_template_revision.test",
		Create:       testAccWorkflowTemplateRevision(templateID, "TERRAFORM", 500, 1024, body("created", "created", `["crud"]`)),
		Update:       testAccWorkflowTemplateRevision(templateID, "TERRAFORM", 1000, 2048, body("updated", "updated", `["crud", "updated"]`)),
		UpdateChecks: []statecheck.StateCheck{
			statecheck.ExpectKnownValue("stackguardian_workflow_template_revision.test",
				tfjsonpath.New("notes"), knownvalue.StringExact("updated")),
			statecheck.ExpectKnownValue("stackguardian_workflow_template_revision.test",
				tfjsonpath.New("user_job_cpu"), knownvalue.Int64Exact(1000)),
		},
		Header: acctest.TemplateHeader(),
		Read: func(ctx context.Context, c *sgclient.Client, org string, rs *terraform.ResourceState) error {
			_, err := c.WorkflowTemplatesRevisions.ReadWorkflowTemplateRevision(ctx, org, rs.Primary.ID)
			return err
		},
		Delete: func(ctx context.Context, c *sgclient.Client, org string, rs *terraform.ResourceState) error {
			return c.WorkflowTemplatesRevisions.DeleteWorkflowTemplateRevision(ctx, org, rs.Primary.ID, true)
		},
	})
}
