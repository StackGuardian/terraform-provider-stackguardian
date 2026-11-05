package stacktemplaterevision_test

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

func TestAccStackTemplateRevision_CRUD(t *testing.T) {
	acctest.SkipUnlessAcceptance(t)
	stackTemplateID := acctest.ResourceName("crud-stack-template-rev")
	wfTemplateID := acctest.ResourceName("crud-wft-for-stack-rev")

	// Two revisions: the gone step recreates the first one as :2.
	t.Cleanup(func() {
		for i := 1; i <= 2; i++ {
			deprecateStackTemplateRevisionFixture(fmt.Sprintf("%s:%d", stackTemplateID, i))
			deleteStackTemplateRevisionFixture(fmt.Sprintf("%s:%d", stackTemplateID, i))
		}
		deleteStackTemplateFixture(stackTemplateID)
		deleteWorkflowTemplateFixture(wfTemplateID)
	})
	if err := createWorkflowTemplateFixture(wfTemplateID, "TERRAFORM"); err != nil {
		t.Fatalf("creating workflow template fixture: %s", err)
	}
	if err := createStackTemplateFixture(stackTemplateID, "TERRAFORM"); err != nil {
		t.Fatalf("creating stack template fixture: %s", err)
	}

	acctest.RunCRUD(t, acctest.CRUD{
		ResourceType: "stackguardian_stack_template_revision",
		Address:      "stackguardian_stack_template_revision.test",
		Create:       testAccStackTemplateRevisionConfig(stackTemplateID, wfTemplateID, "crud", "created", "created"),
		Update:       testAccStackTemplateRevisionConfig(stackTemplateID, wfTemplateID, "crud", "updated", "updated"),
		UpdateChecks: []statecheck.StateCheck{
			statecheck.ExpectKnownValue("stackguardian_stack_template_revision.test",
				tfjsonpath.New("notes"), knownvalue.StringExact("updated")),
			statecheck.ExpectKnownValue("stackguardian_stack_template_revision.test",
				tfjsonpath.New("description"), knownvalue.StringExact("updated")),
		},
		Header: acctest.TemplateHeader(),
		Read: func(ctx context.Context, c *sgclient.Client, org string, rs *terraform.ResourceState) error {
			_, err := c.StackTemplateRevisions.ReadStackTemplateRevision(ctx, org, rs.Primary.ID)
			return err
		},
		Delete: func(ctx context.Context, c *sgclient.Client, org string, rs *terraform.ResourceState) error {
			return c.StackTemplateRevisions.DeleteStackTemplateRevision(ctx, org, rs.Primary.ID, true)
		},
	})
}
