package workflowfromtemplate_test

import (
	"context"
	"testing"

	sgclient "github.com/StackGuardian/sg-sdk-go/client"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccWorkflowUsingTemplate_CRUD(t *testing.T) {
	acctest.SkipUnlessAcceptance(t)
	templateID := setupWorkflowTemplate(t, acctest.ResourceName("crud-wf-from-tmpl-tmpl")) + ":1"
	wfGrpName := acctest.ResourceName("crud-wf-from-tmpl-wfgrp")
	id := acctest.ResourceName("crud-wf-from-tmpl")

	t.Cleanup(func() {
		deleteWorkflowUsingTemplateFixture(wfGrpName, id)
		deleteWorkflowGroupFixture(wfGrpName)
	})
	if err := createWorkflowGroupFixture(wfGrpName); err != nil {
		t.Fatalf("creating workflow group fixture: %s", err)
	}

	body := func(description, tfVersion, tag string) string {
		return `
		  description = "` + description + `"

		  terraform_config = {
		    terraform_version = "` + tfVersion + `"
		  }

		  tags = ["` + tag + `"]
		`
	}
	const address = "stackguardian_workflow_from_template.test"

	acctest.RunCRUD(t, acctest.CRUD{
		ResourceType: "stackguardian_workflow_from_template",
		Address:      address,
		Create:       testAccWorkflowUsingTemplate(wfGrpName, id, "TERRAFORM", templateID, body("created", "1.5.0", "crud")),
		Update:       testAccWorkflowUsingTemplate(wfGrpName, id, "TERRAFORM", templateID, body("updated", "1.5.7", "updated")),
		UpdateChecks: []statecheck.StateCheck{
			statecheck.ExpectKnownValue(address, tfjsonpath.New("description"), knownvalue.StringExact("updated")),
			statecheck.ExpectKnownValue(address,
				tfjsonpath.New("terraform_config").AtMapKey("terraform_version"), knownvalue.StringExact("1.5.7")),
		},
		ImportID: acctest.ImportIDFromAttrs(address, "workflow_group_id", "id"),
		Header:   acctest.TemplateHeader(),
		Read: func(ctx context.Context, c *sgclient.Client, org string, rs *terraform.ResourceState) error {
			_, err := c.Workflows.ReadWorkflow(ctx, org, rs.Primary.ID, rs.Primary.Attributes["workflow_group_id"])
			return err
		},
		Delete: func(ctx context.Context, c *sgclient.Client, org string, rs *terraform.ResourceState) error {
			_, err := c.Workflows.DeleteWorkflow(ctx, org, rs.Primary.ID, rs.Primary.Attributes["workflow_group_id"])
			return err
		},
	})
}
