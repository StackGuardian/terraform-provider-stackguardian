package rolev4_test

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

func TestAccRoleV4_CRUD(t *testing.T) {
	wfGrpName := acctest.ResourceName("crud-role-wfgrp")
	roleName := acctest.ResourceName("crud-role")
	t.Cleanup(func() {
		deleteRoleFixture(roleName)
		deleteWorkflowGroupFixture(wfGrpName)
	})

	// testAccResourceUpdate changes the description and adds a tag.
	acctest.RunCRUD(t, acctest.CRUD{
		ResourceType: "stackguardian_rolev4",
		Address:      "stackguardian_rolev4.this",
		Create:       fmt.Sprintf(testAccResource, "dep", wfGrpName, "this", roleName, "dep"),
		Update:       fmt.Sprintf(testAccResourceUpdate, "dep", wfGrpName, "this", roleName, "dep"),
		UpdateChecks: []statecheck.StateCheck{
			statecheck.ExpectKnownValue("stackguardian_rolev4.this",
				tfjsonpath.New("tags"), knownvalue.ListSizeExact(2)),
		},
		Read: func(ctx context.Context, c *sgclient.Client, org string, rs *terraform.ResourceState) error {
			_, err := c.AccessManagement.ReadRole(ctx, org, rs.Primary.ID)
			return err
		},
		Delete: func(ctx context.Context, c *sgclient.Client, org string, rs *terraform.ResourceState) error {
			return c.AccessManagement.DeleteRole(ctx, org, rs.Primary.ID)
		},
	})
}
