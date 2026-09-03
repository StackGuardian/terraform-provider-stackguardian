package stack_test

import (
	"fmt"
	"testing"

	"github.com/StackGuardian/terraform-provider-stackguardian/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

// context_tags resolution rules (source material for docs-templates/resources/stack.md.tmpl):
//
//  1. Declaring context_tags in the resource config always wins — it
//     overrides whatever the stack template revision itself defines.
//  2. Leaving context_tags unset resolves it from the stack template
//     revision's own context_tags, if the revision has any.
//  3. Setting context_tags to an explicit {} is a deliberate clear, distinct
//     from leaving it unset — it always results in an empty map, even when
//     the template has context_tags of its own.
//  4. Leaving context_tags unset, then changing template_group_id to a
//     revision with no context_tags of its own, re-resolves them fresh
//     against the NEW revision — it does NOT stay stuck on the old
//     revision's value. Since the new revision has none, context_tags
//     resolves to {} (see TestAccStack_ContextTags_ClearedWhenTemplateHasNone
//     below).
func TestAccStack_ContextTags_Resolution(t *testing.T) {
	wfGrpName := "tf-provider-stack-ctxtagsres-wfgrp"
	wfTemplateName := "tf-provider-stack-ctxtagsres-wftmpl"
	stackTemplateName := "tf-provider-stack-ctxtagsres-stmpl"
	id := "tf-provider-stack-ctxtagsres"

	t.Cleanup(func() {
		logCleanupErr(t, fmt.Sprintf("delete workflow group %q", wfGrpName), deleteWorkflowGroupFixture(wfGrpName))
	})
	if err := createWorkflowGroupFixture(wfGrpName); err != nil && !is409(err) {
		t.Fatalf("TestAccStack_ContextTags_Resolution: create workflow group %q: %s", wfGrpName, err)
	}
	workflowTemplateID := setupStackWorkflowTemplate(t, wfTemplateName)
	revision := setupStackTemplateChainWithFields(t, stackTemplateName, workflowTemplateID, nil, nil, map[string]string{"team": "platform"})
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
				Check:  resource.TestCheckResourceAttr("stackguardian_stack.test", "context_tags.team", "platform"),
			},
			{
				// Scenario 1: declared explicitly — overrides the template's.
				Config: testAccStackConfig(wfGrpName, revision, id, `context_tags = { env = "prod" }`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_stack.test", "context_tags.env", "prod"),
					resource.TestCheckNoResourceAttr("stackguardian_stack.test", "context_tags.team"),
				),
			},
			{
				// Scenario 3: an explicit {} is a real clear, not the same as omitting.
				Config: testAccStackConfig(wfGrpName, revision, id, `context_tags = {}`),
				Check:  resource.TestCheckResourceAttr("stackguardian_stack.test", "context_tags.%", "0"),
			},
			{
				// Removing the attribute again does NOT re-inherit the template's
				// context_tags — it carries forward whatever's already in state
				// (the explicit {} from the previous step).
				Config: testAccStackConfig(wfGrpName, revision, id, ""),
				Check:  resource.TestCheckResourceAttr("stackguardian_stack.test", "context_tags.%", "0"),
			},
		},
	})
}

// TestAccStack_ContextTags_ClearedWhenTemplateHasNone covers scenario 4 above: start on a
// revision WITH context_tags (left unset in config, so they resolve from the template), then
// switch template_group_id to a revision with NO context_tags of its own — context_tags must
// re-resolve to {}, not stay stuck on the old revision's value.
//
// Regression test for the ToUpdateAPIModel / reResolveOnRevisionChange coupling — same failure
// mode as the description/tags versions, for context_tags: reResolveOnRevisionChange must
// ALWAYS re-derive context_tags fresh from the new revision when the user left them unset,
// landing on a known-empty {} when the new revision has none, so ToUpdateAPIModel's
// known-non-null branch sends that {} as a real, explicit clear rather than leaving the old
// context_tags in place.
func TestAccStack_ContextTags_ClearedWhenTemplateHasNone(t *testing.T) {
	wfGrpName := "tf-provider-stack-ctxtagsclr-wfgrp"
	wfTemplateName := "tf-provider-stack-ctxtagsclr-wftmpl"
	stackTemplateName := "tf-provider-stack-ctxtagsclr-stmpl"
	id := "tf-provider-stack-ctxtagsclr"

	revisionNoContextTags := setupStackDependencyChain(t, wfGrpName, wfTemplateName, stackTemplateName, id)
	revisionWithContextTags := setupSecondStackTemplateRevisionWithFields(t, stackTemplateName, wfTemplateName, "", nil, nil, map[string]string{"team": "platform"}, defaultSecondRevisionActions())

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader()),
		Steps: []resource.TestStep{
			{
				Config: testAccStackConfig(wfGrpName, revisionWithContextTags, id, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_stack.test", "template_group_id", revisionWithContextTags),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "context_tags.team", "platform"),
				),
			},
			{
				Config: testAccStackConfig(wfGrpName, revisionNoContextTags, id, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_stack.test", "template_group_id", revisionNoContextTags),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "context_tags.%", "0"),
				),
			},
		},
	})
}
