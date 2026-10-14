package policy_test

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

func TestAccPolicy_CRUD(t *testing.T) {
	name := acctest.ResourceName("crud-policy")
	t.Cleanup(func() { deletePolicyFixture(name) })

	// testAccResourceUpdate flips policies_config[0].skip from false to true.
	acctest.RunCRUD(t, acctest.CRUD{
		ResourceType: "stackguardian_policy",
		Address:      "stackguardian_policy.this",
		Create:       fmt.Sprintf(testAccResource, "this", name),
		Update:       fmt.Sprintf(testAccResourceUpdate, "this", name),
		UpdateChecks: []statecheck.StateCheck{
			statecheck.ExpectKnownValue("stackguardian_policy.this",
				tfjsonpath.New("policies_config").AtSliceIndex(0).AtMapKey("skip"), knownvalue.Bool(true)),
		},
		Read: func(ctx context.Context, c *sgclient.Client, org string, rs *terraform.ResourceState) error {
			_, err := c.Policies.ReadPolicy(ctx, org, rs.Primary.ID)
			return err
		},
		Delete: func(ctx context.Context, c *sgclient.Client, org string, rs *terraform.ResourceState) error {
			return c.Policies.DeletePolicy(ctx, org, rs.Primary.ID)
		},
	})
}
