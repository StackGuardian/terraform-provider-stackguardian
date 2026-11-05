package workflowgit_test

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

func TestAccWorkflowGit_CRUD(t *testing.T) {
	acctest.SkipUnlessAcceptance(t)
	wfGrpName := acctest.ResourceName("crud-wf-git-wfgrp")
	id := acctest.ResourceName("crud-wf-git")

	t.Cleanup(func() {
		deleteWorkflowGitFixture(wfGrpName, id)
		deleteWorkflowGroupFixture(wfGrpName)
	})
	if err := createWorkflowGroupFixture(wfGrpName); err != nil {
		t.Fatalf("creating workflow group fixture: %s", err)
	}

	body := func(description, repo string) string {
		return `
		  description = "` + description + `"

		  vcs_config = {
		    iac_vcs_config = {
		      custom_source = {
		        config = {
		          is_private = false
		          repo       = "` + repo + `"
		        }
		        source_config_dest_kind = "GIT_OTHER"
		      }
		    }
		  }
		`
	}
	const address = "stackguardian_workflow_git.test"

	acctest.RunCRUD(t, acctest.CRUD{
		ResourceType: "stackguardian_workflow_git",
		Address:      address,
		Create:       testAccWorkflowGit(wfGrpName, id, "CUSTOM", body("created", "https://github.com/dummy/test-repo.git")),
		Update:       testAccWorkflowGit(wfGrpName, id, "CUSTOM", body("updated", "https://github.com/dummy/test-repo-2.git")),
		UpdateChecks: []statecheck.StateCheck{
			statecheck.ExpectKnownValue(address, tfjsonpath.New("description"), knownvalue.StringExact("updated")),
			statecheck.ExpectKnownValue(address,
				tfjsonpath.New("vcs_config").AtMapKey("iac_vcs_config").AtMapKey("custom_source").AtMapKey("config").AtMapKey("repo"),
				knownvalue.StringExact("https://github.com/dummy/test-repo-2.git")),
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
