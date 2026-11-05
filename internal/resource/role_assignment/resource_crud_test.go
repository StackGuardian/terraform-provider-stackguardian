package roleassignment_test

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

// Two roles; the assignment points at whichever one the last %s names. It uses
// roles, not the deprecated role attribute.
const testAccCRUDConfig = `
resource "stackguardian_role" "a" {
  resource_name = %q
  description   = "CRUD role assignment fixture A"
}

resource "stackguardian_role" "b" {
  resource_name = %q
  description   = "CRUD role assignment fixture B"
}

resource "stackguardian_role_assignment" "this" {
  user_id     = %q
  entity_type = "EMAIL"
  roles       = [stackguardian_role.%s.resource_name]
  send_email  = false
}`

func TestAccRoleAssignment_CRUD(t *testing.T) {
	userId := "crud.user@domain.com"
	roleA := acctest.ResourceName("crud-assign-role-a")
	roleB := acctest.ResourceName("crud-assign-role-b")

	deleteRoleAssignmentFixture(userId)
	t.Cleanup(func() {
		deleteRoleAssignmentFixture(userId)
		deleteRoleFixture(roleA)
		deleteRoleFixture(roleB)
	})

	acctest.RunCRUD(t, acctest.CRUD{
		ResourceType: "stackguardian_role_assignment",
		Address:      "stackguardian_role_assignment.this",
		Create:       fmt.Sprintf(testAccCRUDConfig, roleA, roleB, userId, "a"),
		Update:       fmt.Sprintf(testAccCRUDConfig, roleA, roleB, userId, "b"),
		UpdateChecks: []statecheck.StateCheck{
			statecheck.ExpectKnownValue("stackguardian_role_assignment.this",
				tfjsonpath.New("roles"), knownvalue.ListExact([]knownvalue.Check{knownvalue.StringExact(roleB)})),
		},
		ImportID:               acctest.ImportIDFromAttrs("stackguardian_role_assignment.this", "user_id"),
		ImportVerifyIdentifier: "user_id",
		// send_email only drives the invitation on create; the API never returns it.
		ImportIgnore: []string{"send_email"},
		Read: func(ctx context.Context, c *sgclient.Client, org string, rs *terraform.ResourceState) error {
			userId := rs.Primary.Attributes["user_id"]
			_, err := c.AccessManagement.ReadUser(ctx, org, &sgsdkgo.GetorRemoveUserFromOrganization{UserId: &userId})
			return err
		},
		Delete: func(ctx context.Context, c *sgclient.Client, org string, rs *terraform.ResourceState) error {
			userId := rs.Primary.Attributes["user_id"]
			_, err := c.AccessManagement.DeleteUser(ctx, org, &sgsdkgo.GetorRemoveUserFromOrganization{UserId: &userId})
			return err
		},
	})
}
