package stack_test

import (
	"fmt"
	"testing"

	"github.com/StackGuardian/terraform-provider-stackguardian/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

// tags resolution rules (source material for docs-templates/resources/stack.md.tmpl):
//
//  1. Declaring tags in the resource config always wins — it overrides
//     whatever the stack template revision itself defines.
//  2. Leaving tags unset resolves it from the stack template revision's own
//     tags, if the revision has any.
//  3. Setting tags to an explicit [] is a deliberate clear, distinct from
//     leaving it unset — it always results in an empty tag list, even when
//     the template has tags of its own. tags = [] is a known, non-null,
//     zero-length value, not the same thing as omitting the attribute — see
//     TestAccStack_WorkflowsConfigUpdate's comment (resource_workflows_test.go)
//     for the established rationale behind that distinction.
//  4. Leaving tags unset, then changing template_group_id to a revision with
//     no tags of its own, re-resolves them fresh against the NEW revision —
//     it does NOT stay stuck on the old revision's tags. Since the new
//     revision has none, tags resolves to [] (see
//     TestAccStack_Tags_ClearedWhenTemplateHasNone below).
func TestAccStack_Tags_Resolution(t *testing.T) {
	wfGrpName := "tf-provider-stack-tagsres-wfgrp"
	wfTemplateName := "tf-provider-stack-tagsres-wftmpl"
	stackTemplateName := "tf-provider-stack-tagsres-stmpl"
	id := "tf-provider-stack-tagsres"

	t.Cleanup(func() {
		logCleanupErr(t, fmt.Sprintf("delete workflow group %q", wfGrpName), deleteWorkflowGroupFixture(wfGrpName))
	})
	if err := createWorkflowGroupFixture(wfGrpName); err != nil && !is409(err) {
		t.Fatalf("TestAccStack_Tags_Resolution: create workflow group %q: %s", wfGrpName, err)
	}
	workflowTemplateID := setupStackWorkflowTemplate(t, wfTemplateName)
	revision := setupStackTemplateChainWithFields(t, stackTemplateName, workflowTemplateID, nil, []string{"tmpl-tag"}, nil)
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
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_stack.test", "tags.#", "1"),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "tags.0", "tmpl-tag"),
				),
			},
			{
				// Scenario 1: declared explicitly — overrides the template's.
				Config: testAccStackConfig(wfGrpName, revision, id, `tags = ["user-tag-a", "user-tag-b"]`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_stack.test", "tags.#", "2"),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "tags.0", "user-tag-a"),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "tags.1", "user-tag-b"),
				),
			},
			{
				// Scenario 3: an explicit [] is a real clear, not the same as omitting.
				Config: testAccStackConfig(wfGrpName, revision, id, `tags = []`),
				Check:  resource.TestCheckResourceAttr("stackguardian_stack.test", "tags.#", "0"),
			},
			{
				// Removing the attribute again does NOT re-inherit the template's
				// tags — it carries forward whatever's already in state (the
				// explicit [] from the previous step).
				Config: testAccStackConfig(wfGrpName, revision, id, ""),
				Check:  resource.TestCheckResourceAttr("stackguardian_stack.test", "tags.#", "0"),
			},
		},
	})
}

// TestAccStack_Tags_ClearedWhenTemplateHasNone covers scenario 4 above: start on a revision
// WITH tags (tags left unset in config, so they resolve from the template), then switch
// template_group_id to a revision with NO tags of its own — tags must re-resolve to [], not
// stay stuck on the old revision's value.
//
// Regression test for the ToUpdateAPIModel / reResolveOnRevisionChange coupling — same failure
// mode as TestAccStack_Description_ClearedWhenTemplateHasNone, but for tags:
// reResolveOnRevisionChange must ALWAYS re-derive tags fresh from the new revision when the user
// left them unset, landing on a known-empty [] when the new revision has none, so
// ToUpdateAPIModel's known-non-null branch sends that [] as a real, explicit clear rather than
// leaving the old tags in place.
func TestAccStack_Tags_ClearedWhenTemplateHasNone(t *testing.T) {
	wfGrpName := "tf-provider-stack-tagsclr-wfgrp"
	wfTemplateName := "tf-provider-stack-tagsclr-wftmpl"
	stackTemplateName := "tf-provider-stack-tagsclr-stmpl"
	id := "tf-provider-stack-tagsclr"

	revisionNoTags := setupStackDependencyChain(t, wfGrpName, wfTemplateName, stackTemplateName, id)
	revisionWithTags := setupSecondStackTemplateRevisionWithFields(t, stackTemplateName, wfTemplateName, "", nil, []string{"original-tag"}, nil, defaultSecondRevisionActions())

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader()),
		Steps: []resource.TestStep{
			{
				Config: testAccStackConfig(wfGrpName, revisionWithTags, id, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_stack.test", "template_group_id", revisionWithTags),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "tags.#", "1"),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "tags.0", "original-tag"),
				),
			},
			{
				Config: testAccStackConfig(wfGrpName, revisionNoTags, id, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_stack.test", "template_group_id", revisionNoTags),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "tags.#", "0"),
				),
			},
		},
	})
}
