package runnergroup_test

import (
	"context"
	"fmt"
	"testing"

	sgsdkgo "github.com/StackGuardian/sg-sdk-go"
	sgclient "github.com/StackGuardian/sg-sdk-go/client"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccRunnerGroup_CRUD(t *testing.T) {
	name := acctest.ResourceName("crud-runnergroup")
	t.Cleanup(func() { deleteRunnerGroupFixture(name) })

	// testAccResourceUpdate adds a tag. The S3 backend uses the org's
	// pre-existing /integrations/taher-aws connector.
	acctest.RunCRUD(t, acctest.CRUD{
		ResourceType: "stackguardian_runner_group",
		Address:      "stackguardian_runner_group.this",
		Create:       fmt.Sprintf(testAccResource, "this", name),
		Update:       fmt.Sprintf(testAccResourceUpdate, "this", name),
		UpdateChecks: []statecheck.StateCheck{
			statecheck.ExpectKnownValue("stackguardian_runner_group.this",
				tfjsonpath.New("tags"), knownvalue.ListSizeExact(3)),
		},
		Read: func(ctx context.Context, c *sgclient.Client, org string, rs *terraform.ResourceState) error {
			no := false
			_, err := c.RunnerGroups.ReadRunnerGroup(ctx, org, rs.Primary.ID, &sgsdkgo.ReadRunnerGroupRequest{
				GetActiveWorkflows:        &no,
				GetActiveWorkflowsDetails: &no,
			})
			return err
		},
		Delete: func(ctx context.Context, c *sgclient.Client, org string, rs *terraform.ResourceState) error {
			_, err := c.RunnerGroups.DeleteRunnerGroup(ctx, org, rs.Primary.ID)
			return err
		},
	})
}
