package workflowgroup_test

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

const testAccCRUDConfig = `resource "stackguardian_workflow_group" "this" {
  resource_name = %q
  description   = %q
  tags          = %s
}`

func TestAccWorkflowGroup_CRUD(t *testing.T) {
	name := acctest.ResourceName("crud-wfgrp")
	t.Cleanup(func() { deleteWorkflowGroupFixture(name) })

	acctest.RunCRUD(t, acctest.CRUD{
		ResourceType: "stackguardian_workflow_group",
		Address:      "stackguardian_workflow_group.this",
		Create:       fmt.Sprintf(testAccCRUDConfig, name, "created", `["crud"]`),
		Update:       fmt.Sprintf(testAccCRUDConfig, name, "updated", `["crud", "updated"]`),
		UpdateChecks: []statecheck.StateCheck{
			statecheck.ExpectKnownValue("stackguardian_workflow_group.this",
				tfjsonpath.New("description"), knownvalue.StringExact("updated")),
			statecheck.ExpectKnownValue("stackguardian_workflow_group.this",
				tfjsonpath.New("tags"), knownvalue.ListSizeExact(2)),
		},
		Read: func(ctx context.Context, c *sgclient.Client, org string, rs *terraform.ResourceState) error {
			_, err := c.WorkflowGroups.ReadWorkflowGroup(ctx, org, rs.Primary.ID)
			return err
		},
		Delete: func(ctx context.Context, c *sgclient.Client, org string, rs *terraform.ResourceState) error {
			_, err := c.WorkflowGroups.DeleteWorkflowGroup(ctx, org, rs.Primary.ID)
			return err
		},
	})
}
