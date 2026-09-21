package stack_test

import (
	"fmt"
	"testing"

	sgsdkgo "github.com/StackGuardian/sg-sdk-go"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
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
// Purpose: workflows_config.workflows[] growing to include a new slot, then shrinking back —
// the existing slot's data must be left undisturbed when a second slot is added, and the new
// slot must appear with its own values.
//
// Why this has to be a revision switch: workflows_config must declare exactly the workflow
// slots the ACTIVE stack template revision defines, in the same order, on every single apply
// (validateWorkflowsConfigMatchesRevision, resource.go/model.go) — there's no longer a way to
// leave a slot out and add it in a later update while template_group_id stays put; that update
// would be rejected on the very step that omits it. So the only way workflows_config's shape can
// change at all is a revision switch that itself adds or drops a slot:
//
//   - revision1 (setupStackDependencyChain) wires only testWfSlotId.
//   - revision2 (setupSecondStackTemplateRevisionTwoSlots) wires testWfSlotId AND secondWfSlotId.
//
// Step 1 creates against revision1 (one slot declared). Step 2 switches template_group_id to
// revision2, which now requires secondWfSlotId to be declared too — the existing slot's tags
// must be undisturbed, and the new slot must appear with its own values. Step 3 switches back to
// revision1, which now requires secondWfSlotId to be ABSENT — workflows_config must shrink back
// to one entry cleanly, not error or leave the removed slot lingering in state.
func TestAccStack_WorkflowsConfigAddSecondWorkflowOverride(t *testing.T) {
	wfGrpName := "tf-provider-stack-wfadd-wfgrp"
	wfTemplateName := "tf-provider-stack-wfadd-wftmpl"
	stackTemplateName := "tf-provider-stack-wfadd-stmpl"
	id := "tf-provider-stack-wfadd"

	revision1 := setupStackDependencyChain(t, wfGrpName, wfTemplateName, stackTemplateName, id)
	revision2 := setupSecondStackTemplateRevisionTwoSlots(t, stackTemplateName, wfTemplateName)

	oneSlot := fmt.Sprintf(`
  workflows_config = {
    workflows = [
      { id = %q, tags = ["first"] }
    ]
  }
`, testWfSlotId)

	twoSlots := fmt.Sprintf(`
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
				Config: testAccStackConfig(wfGrpName, revision1, id, oneSlot),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_stack.test", "workflows_config.workflows.#", "1"),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "workflows_config.workflows.0.id", testWfSlotId),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "workflows_config.workflows.0.tags.0", "first"),
				),
			},
			{
				// Switch to revision2 — its second slot must now be declared, the
				// first slot's data must be undisturbed, and the new slot must
				// appear with its own values.
				Config: testAccStackConfig(wfGrpName, revision2, id, twoSlots),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_stack.test", "workflows_config.workflows.#", "2"),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "workflows_config.workflows.0.tags.0", "first"),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "workflows_config.workflows.1.id", secondWfSlotId),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "workflows_config.workflows.1.tags.0", "second"),
				),
			},
			{
				// Switch back to revision1 — must shrink back to one cleanly.
				Config: testAccStackConfig(wfGrpName, revision1, id, oneSlot),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_stack.test", "workflows_config.workflows.#", "1"),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "workflows_config.workflows.0.id", testWfSlotId),
				),
			},
		},
	})
}
