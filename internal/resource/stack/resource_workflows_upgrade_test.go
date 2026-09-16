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
