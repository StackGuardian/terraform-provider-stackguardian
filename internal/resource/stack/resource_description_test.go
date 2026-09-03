package stack_test

import (
	"fmt"
	"testing"

	"github.com/StackGuardian/terraform-provider-stackguardian/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

// description resolution rules (source material for docs-templates/resources/stack.md.tmpl):
//
//  1. Declaring description in the resource config always wins — it overrides
//     whatever the stack template revision itself defines.
//  2. Leaving description unset resolves it from the stack template
//     revision's own description, if the revision has one.
//  3. Setting description to an explicit "" is a deliberate clear, distinct
//     from leaving it unset — it always results in an empty description,
//     even when the template has one.
//  4. Leaving description unset, then changing template_group_id to a
//     revision with no description of its own, re-resolves it fresh against
//     the NEW revision — it does NOT stay stuck on the old revision's value.
//     Since the new revision has nothing, it resolves to "" (see
//     TestAccStack_Description_ClearedWhenTemplateHasNone below).
func TestAccStack_Description_Resolution(t *testing.T) {
	wfGrpName := "tf-provider-stack-descres-wfgrp"
	wfTemplateName := "tf-provider-stack-descres-wftmpl"
	stackTemplateName := "tf-provider-stack-descres-stmpl"
	id := "tf-provider-stack-descres"

	t.Cleanup(func() {
		logCleanupErr(t, fmt.Sprintf("delete workflow group %q", wfGrpName), deleteWorkflowGroupFixture(wfGrpName))
	})
	if err := createWorkflowGroupFixture(wfGrpName); err != nil && !is409(err) {
		t.Fatalf("TestAccStack_Description_Resolution: create workflow group %q: %s", wfGrpName, err)
	}
	workflowTemplateID := setupStackWorkflowTemplate(t, wfTemplateName)
	templateDescription := "template's own description"
	revision := setupStackTemplateChainWithFields(t, stackTemplateName, workflowTemplateID, &templateDescription, nil, nil)
	t.Cleanup(func() { deleteStackFixture(wfGrpName, id) })

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader()),
		Steps: []resource.TestStep{
			{
				// Scenario 2: unset — resolves from the template.
				Config: testAccStackConfig(wfGrpName, revision, id, ""),
				Check:  resource.TestCheckResourceAttr("stackguardian_stack.test", "description", templateDescription),
			},
			{
				// Scenario 1: declared explicitly — overrides the template's.
				Config: testAccStackConfig(wfGrpName, revision, id, `description = "user's own description"`),
				Check:  resource.TestCheckResourceAttr("stackguardian_stack.test", "description", "user's own description"),
			},
			{
				// Scenario 3: an explicit "" is a real clear, not the same as omitting.
				Config: testAccStackConfig(wfGrpName, revision, id, `description = ""`),
				Check:  resource.TestCheckResourceAttr("stackguardian_stack.test", "description", ""),
			},
			{
				// Removing the attribute again does NOT re-inherit the template's
				// description — it carries forward whatever's already in state
				// (the explicit "" from the previous step). Optional+Computed
				// attributes can't distinguish "never set" from "explicitly
				// cleared, then the declaration removed".
				Config: testAccStackConfig(wfGrpName, revision, id, ""),
				Check:  resource.TestCheckResourceAttr("stackguardian_stack.test", "description", ""),
			},
		},
	})
}

// TestAccStack_Description_ClearedWhenTemplateHasNone covers scenario 4 above: start on a
// revision WITH a description (description left unset in config, so it resolves from the
// template), then switch template_group_id to a revision with NO description of its own —
// description must re-resolve to "", not stay stuck on the old revision's value.
//
// Regression test for the ToUpdateAPIModel / reResolveOnRevisionChange coupling:
// reResolveOnRevisionChange must ALWAYS re-derive description fresh from the new revision when
// the user left it unset — landing on a known "" (via knownEmptyStringIfNull) when the new
// revision has none, never an actual null and never the stale old value. ToUpdateAPIModel's
// known-non-null branch then sends that "" as a real, explicit clear. If reResolveOnRevisionChange
// regressed to producing an actual null instead (e.g. by reverting to plain flatteners.StringPtr
// without the knownEmptyStringIfNull wrapper), this step would fail with "Provider produced
// inconsistent result after apply" instead of the assertion below ever running, since
// ToUpdateAPIModel would then omit the field (leaving the old value in place) while the plan had
// predicted null.
func TestAccStack_Description_ClearedWhenTemplateHasNone(t *testing.T) {
	wfGrpName := "tf-provider-stack-descclr-wfgrp"
	wfTemplateName := "tf-provider-stack-descclr-wftmpl"
	stackTemplateName := "tf-provider-stack-descclr-stmpl"
	id := "tf-provider-stack-descclr"

	revisionNoDesc := setupStackDependencyChain(t, wfGrpName, wfTemplateName, stackTemplateName, id)
	revisionWithDesc := setupSecondStackTemplateRevision(t, stackTemplateName, wfTemplateName, "original description", nil)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader()),
		Steps: []resource.TestStep{
			{
				Config: testAccStackConfig(wfGrpName, revisionWithDesc, id, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_stack.test", "template_group_id", revisionWithDesc),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "description", "original description"),
				),
			},
			{
				Config: testAccStackConfig(wfGrpName, revisionNoDesc, id, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_stack.test", "template_group_id", revisionNoDesc),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "description", ""),
				),
			},
		},
	})
}
