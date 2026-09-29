package stack_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	sgsdkgo "github.com/StackGuardian/sg-sdk-go"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/acctest"
	stackresource "github.com/StackGuardian/terraform-provider-stackguardian/internal/resource/stack"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

// TestAccStack_WorkflowsConfigRevisionReResolution — REVISION SWITCH test.
// Purpose: a per-workflow field the stack leaves unset (number_of_approvals_required) must
// pick up a value the new revision provides that the old one didn't (revision1 -> revision2,
// step 2), and must go back to empty when a later revision drops it again (revision2 ->
// revision1, step 3) — rather than the value getting stuck at whatever UseStateForUnknown last
// carried forward (see reResolveWorkflowsConfigOnRevisionChange).
func TestAccStack_WorkflowsConfigRevisionReResolution(t *testing.T) {
	wfGrpName := "tf-provider-stack-wfrevre-wfgrp"
	wfTemplateName := "tf-provider-stack-wfrevre-wftmpl"
	stackTemplateName := "tf-provider-stack-wfrevre-stmpl"
	id := "tf-provider-stack-wfrevre"

	revision1 := setupStackDependencyChain(t, wfGrpName, wfTemplateName, stackTemplateName, id)
	revision2 := setupSecondStackTemplateRevision(t, stackTemplateName, wfTemplateName, "revision two", sgsdkgo.Int(2))

	config := fmt.Sprintf(`
  workflows_config = {
    workflows = [
      { id = %q }
    ]
  }
`, testWfSlotId)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader()),
		Steps: []resource.TestStep{
			{
				// revision1 never sets number_of_approvals_required.
				Config: testAccStackConfig(wfGrpName, revision1, id, config),
				Check:  resource.TestCheckResourceAttr("stackguardian_stack.test", "workflows_config.workflows.0.number_of_approvals_required", "0"),
			},
			{
				// revision2 sets it to 2 — must now appear, though nothing in
				// this stack's own config changed.
				Config: testAccStackConfig(wfGrpName, revision2, id, config),
				Check:  resource.TestCheckResourceAttr("stackguardian_stack.test", "workflows_config.workflows.0.number_of_approvals_required", "2"),
			},
			{
				// Back to revision1 — must clear again, not stay stuck at 2.
				Config: testAccStackConfig(wfGrpName, revision1, id, config),
				Check:  resource.TestCheckResourceAttr("stackguardian_stack.test", "workflows_config.workflows.0.number_of_approvals_required", "0"),
			},
		},
	})
}

// TestAccStack_WorkflowsConfigAddSecondWorkflowOverride — REVISION SWITCH test.
// Purpose: workflows_config.workflows[] growing to include a new workflow entry, then
// shrinking back — the existing entry's data must be left undisturbed when a second entry
// is added, and the new entry must appear with its own values.
//
// Why this has to be a revision switch: workflows_config must declare exactly the workflow
// entries the ACTIVE stack template revision defines, in the same order, on every single
// apply (validateWorkflowsConfigMatchesRevision, resource.go/model.go) — there's no longer a
// way to leave an entry out and add it in a later update while template_group_id stays put;
// that update would be rejected on the very step that omits it. So the only way
// workflows_config's shape can change at all is a revision switch that itself adds or drops
// a workflow entry:
//
//   - revision1 (setupStackDependencyChain) declares only the testWfSlotId workflow entry.
//   - revision2 (setupSecondStackTemplateRevisionTwoSlots) declares both testWfSlotId AND
//     secondWfSlotId.
//
// Step 1 creates against revision1 (one workflow entry declared). Step 2 switches
// template_group_id to revision2, which now requires secondWfSlotId to be declared too — the
// existing entry's tags must be undisturbed, and the new entry must appear with its own
// values. Step 3 switches back to revision1, which now requires secondWfSlotId to be ABSENT
// — workflows_config must shrink back to one entry cleanly, not error or leave the removed
// entry lingering in state.
func TestAccStack_WorkflowsConfigAddSecondWorkflowOverride(t *testing.T) {
	wfGrpName := "tf-provider-stack-wfadd-wfgrp"
	wfTemplateName := "tf-provider-stack-wfadd-wftmpl"
	stackTemplateName := "tf-provider-stack-wfadd-stmpl"
	id := "tf-provider-stack-wfadd"

	revision1 := setupStackDependencyChain(t, wfGrpName, wfTemplateName, stackTemplateName, id)
	revision2 := setupSecondStackTemplateRevisionTwoSlots(t, stackTemplateName, wfTemplateName)

	oneWorkflowConfig := fmt.Sprintf(`
  workflows_config = {
    workflows = [
      { id = %q, tags = ["first"] }
    ]
  }
`, testWfSlotId)

	twoWorkflowsConfig := fmt.Sprintf(`
  workflows_config = {
    workflows = [
      { id = %q, tags = ["first"] },
      { id = %q, tags = ["second"] }
    ]
  }
`, testWfSlotId, secondWfSlotId)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader()),
		Steps: []resource.TestStep{
			{
				Config: testAccStackConfig(wfGrpName, revision1, id, oneWorkflowConfig),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_stack.test", "workflows_config.workflows.#", "1"),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "workflows_config.workflows.0.id", testWfSlotId),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "workflows_config.workflows.0.tags.0", "first"),
				),
			},
			{
				// Switch to revision2 — its second workflow entry must now be declared,
				// the first entry's data must be undisturbed, and the new entry must
				// appear with its own values.
				Config: testAccStackConfig(wfGrpName, revision2, id, twoWorkflowsConfig),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_stack.test", "workflows_config.workflows.#", "2"),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "workflows_config.workflows.0.tags.0", "first"),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "workflows_config.workflows.1.id", secondWfSlotId),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "workflows_config.workflows.1.tags.0", "second"),
				),
			},
			{
				// Switch back to revision1 — must shrink back to one cleanly.
				Config: testAccStackConfig(wfGrpName, revision1, id, oneWorkflowConfig),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_stack.test", "workflows_config.workflows.#", "1"),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "workflows_config.workflows.0.id", testWfSlotId),
				),
			},
		},
	})
}

// TestAccStack_WorkflowsConfigRemoveAndAddWorkflow — REVISION SWITCH test.
// Purpose: TestAccStack_WorkflowsConfigAddSecondWorkflowOverride only covers a pure grow
// (revision1 -> revision2 adds a workflow entry, keeping revision1's) and a pure shrink
// (revision2 -> revision1 drops it again) — the list length always changes by one, and the
// entry(entries) already declared are never themselves removed. This test instead covers a
// revision change that REMOVES a previously-declared workflow entry AND ADDS a different,
// new one in the same step, so workflows_config.workflows[] keeps the same length (one) but
// its membership changes entirely: revision1 (setupStackDependencyChain) declares only the
// testWfSlotId workflow entry, setupSecondStackTemplateRevisionRemoveAndAddWorkflow's
// revision2 declares only the secondWfSlotId entry instead. Switching from revision1 to
// revision2 must confirm testWfSlotId's live workflow is actually removed/cleaned up (not
// orphaned) while secondWfSlotId's is correctly created and populated — exercising both the
// "remove" and "add" halves of reResolveWorkflowsConfigOnRevisionChange and
// validateWorkflowsConfigMatchesRevision's exact-match check together, rather than each in
// isolation the way the existing grow/shrink test does.
func TestAccStack_WorkflowsConfigRemoveAndAddWorkflow(t *testing.T) {
	wfGrpName := "tf-provider-stack-wfremoveadd-wfgrp"
	wfTemplateName := "tf-provider-stack-wfremoveadd-wftmpl"
	stackTemplateName := "tf-provider-stack-wfremoveadd-stmpl"
	id := "tf-provider-stack-wfremoveadd"

	revision1 := setupStackDependencyChain(t, wfGrpName, wfTemplateName, stackTemplateName, id)
	revision2 := setupSecondStackTemplateRevisionRemoveAndAddWorkflow(t, stackTemplateName, wfTemplateName)

	firstWorkflowConfig := fmt.Sprintf(`
  workflows_config = {
    workflows = [
      { id = %q }
    ]
  }
`, testWfSlotId)

	secondWorkflowConfig := fmt.Sprintf(`
  workflows_config = {
    workflows = [
      { id = %q }
    ]
  }
`, secondWfSlotId)

	// The id testWfSlotId's workflow resolves to under revision1 — computed the same way
	// the provider itself does (model.go's computeWorkflowId, exposed via export_test.go),
	// fed the same resolved iac_template_id setupStackDependencyChain's chain wires onto
	// that workflow entry. Used after the revision change to confirm the platform actually
	// removed it.
	resolvedIacTemplateId := fmt.Sprintf("/%s/%s:1", org, wfTemplateName)
	oldWorkflowId := stackresource.ComputeWorkflowId(resolvedIacTemplateId, testWfSlotId)

	checkOldWorkflowRemoved := func(s *terraform.State) error {
		_, err := getClient().StackWorkflows.ReadStackWorkflow(context.TODO(), org, id, oldWorkflowId, wfGrpName)
		if err == nil {
			return fmt.Errorf("expected workflow %q (entry %s) to be removed after the revision change, but it still exists", oldWorkflowId, testWfSlotId)
		}
		// A missing workflow can come back as a 400 with {"msg":"Workflow does not exist"}
		// rather than a 404 (the API's api_helper decorator defaults any response without an
		// explicit status to 400) — match on the message since the status code alone can't
		// distinguish "removed" from a real failure.
		if !strings.Contains(err.Error(), "does not exist") {
			return fmt.Errorf("unexpected error checking removal of workflow %q: %w", oldWorkflowId, err)
		}
		return nil
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader()),
		Steps: []resource.TestStep{
			{
				Config: testAccStackConfig(wfGrpName, revision1, id, firstWorkflowConfig),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_stack.test", "workflows_config.workflows.#", "1"),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "workflows_config.workflows.0.id", testWfSlotId),
					// actions is unset in config, so it comes from revision1: apply/plan
					// each ordering only the testWfSlotId workflow.
					resource.TestCheckResourceAttr("stackguardian_stack.test", "actions.apply.order.%", "1"),
					resource.TestCheckResourceAttr("stackguardian_stack.test",
						fmt.Sprintf("actions.apply.order.%s.parameters.terraform_action.action", testWfSlotId), "apply"),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "actions.plan.order.%", "1"),
					resource.TestCheckResourceAttr("stackguardian_stack.test",
						fmt.Sprintf("actions.plan.order.%s.parameters.terraform_action.action", testWfSlotId), "plan"),
				),
			},
			{
				// Switch to revision2 — testWfSlotId's workflow must be removed (not
				// orphaned) and secondWfSlotId's added in its place; the list length
				// stays at one throughout, only its membership changes. actions (still
				// unset in config) must be re-resolved from revision2: the removed
				// workflow must no longer be ordered, and the added one must be.
				Config: testAccStackConfig(wfGrpName, revision2, id, secondWorkflowConfig),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_stack.test", "workflows_config.workflows.#", "1"),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "workflows_config.workflows.0.id", secondWfSlotId),
					resource.TestCheckResourceAttrSet("stackguardian_stack.test", "workflows_config.workflows.0.workflow_id"),
					resource.TestCheckResourceAttrSet("stackguardian_stack.test", "workflows_config.workflows.0.resource_name"),
					checkOldWorkflowRemoved,
					resource.TestCheckResourceAttr("stackguardian_stack.test", "actions.apply.order.%", "1"),
					resource.TestCheckResourceAttr("stackguardian_stack.test",
						fmt.Sprintf("actions.apply.order.%s.parameters.terraform_action.action", secondWfSlotId), "apply"),
					resource.TestCheckNoResourceAttr("stackguardian_stack.test",
						fmt.Sprintf("actions.apply.order.%s.parameters.terraform_action.action", testWfSlotId)),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "actions.plan.order.%", "1"),
					resource.TestCheckResourceAttr("stackguardian_stack.test",
						fmt.Sprintf("actions.plan.order.%s.parameters.terraform_action.action", secondWfSlotId), "plan"),
					resource.TestCheckNoResourceAttr("stackguardian_stack.test",
						fmt.Sprintf("actions.plan.order.%s.parameters.terraform_action.action", testWfSlotId)),
				),
			},
		},
	})
}
