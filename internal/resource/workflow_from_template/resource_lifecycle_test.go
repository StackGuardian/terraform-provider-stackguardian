package workflowfromtemplate_test

import (
	"testing"

	"github.com/StackGuardian/terraform-provider-stackguardian/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

// TestAccWorkflowUsingTemplate_WorkflowGroupRequiresReplace verifies that changing
// workflow_group_id forces a destroy+create (RequiresReplace), since the platform has no
// operation to move a workflow between groups.
func TestAccWorkflowUsingTemplate_WorkflowGroupRequiresReplace(t *testing.T) {
	templateID := setupWorkflowTemplate(t, "tf-wfgrp-replace-tpl") + ":1"
	wfGrpA := acctest.ResourceName("tf-wfgrp-replace-a")
	wfGrpB := acctest.ResourceName("tf-wfgrp-replace-b")
	id := acctest.ResourceName("tf-wfgrp-replace-wf")

	for _, g := range []string{wfGrpA, wfGrpB} {
		if err := createWorkflowGroupFixture(g); err != nil {
			t.Fatalf("wfgrp fixture %q: %s", g, err)
		}
		defer deleteWorkflowGroupFixture(g)
	}
	defer deleteWorkflowUsingTemplateFixture(wfGrpA, id)
	defer deleteWorkflowUsingTemplateFixture(wfGrpB, id)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks:   []tfversion.TerraformVersionCheck{tfversion.SkipBelow(tfversion.Version1_1_0)},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader()),
		Steps: []resource.TestStep{
			{
				// Create in group A.
				Config: testAccWorkflowUsingTemplate(wfGrpA, id, "TERRAFORM", templateID, ""),
				Check: resource.TestCheckResourceAttr(
					"stackguardian_workflow_from_template.test", "workflow_group_id", wfGrpA),
			},
			{
				// Move to group B -> must be a Replace (destroy + create), not an in-place update.
				Config: testAccWorkflowUsingTemplate(wfGrpB, id, "TERRAFORM", templateID, ""),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(
							"stackguardian_workflow_from_template.test", plancheck.ResourceActionReplace),
					},
				},
				Check: resource.TestCheckResourceAttr(
					"stackguardian_workflow_from_template.test", "workflow_group_id", wfGrpB),
			},
		},
	})
}

// TestAccWorkflowUsingTemplate_ManualDeleteRecreates verifies that when the workflow is
// deleted out-of-band (e.g. from the platform UI), the next refresh drops it from state and
// the following plan proposes to recreate it (rather than erroring). The
// deleteWorkflowUsingTemplateFixture call between steps simulates the out-of-band delete;
// ExpectNonEmptyPlan on the refresh step asserts the resource is planned for recreate.
func TestAccWorkflowUsingTemplate_ManualDeleteRecreates(t *testing.T) {
	templateID := setupWorkflowTemplate(t, "tf-manualdel-tpl") + ":1"
	wfGrp := acctest.ResourceName("tf-manualdel-wfgrp")
	id := acctest.ResourceName("tf-manualdel-wf")

	if err := createWorkflowGroupFixture(wfGrp); err != nil {
		t.Fatalf("wfgrp fixture: %s", err)
	}
	defer deleteWorkflowGroupFixture(wfGrp)
	defer deleteWorkflowUsingTemplateFixture(wfGrp, id)

	cfg := testAccWorkflowUsingTemplate(wfGrp, id, "TERRAFORM", templateID, "")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks:   []tfversion.TerraformVersionCheck{tfversion.SkipBelow(tfversion.Version1_1_0)},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader()),
		Steps: []resource.TestStep{
			{
				// Create the workflow.
				Config: cfg,
				Check: resource.TestCheckResourceAttr(
					"stackguardian_workflow_from_template.test", "id", id),
			},
			{
				// Delete it out-of-band, then run a refresh-only step (no Config — the
				// framework forbids Config + RefreshState): Read gets a 404 and removes it
				// from state, so the follow-up plan is non-empty (recreate proposed).
				PreConfig:          func() { deleteWorkflowUsingTemplateFixture(wfGrp, id) },
				RefreshState:       true,
				ExpectNonEmptyPlan: true,
			},
			{
				// Applying the same config recreates it and settles to a clean plan.
				Config: cfg,
				Check: resource.TestCheckResourceAttr(
					"stackguardian_workflow_from_template.test", "id", id),
			},
			{
				Config:   cfg,
				PlanOnly: true,
			},
		},
	})
}
