package stack_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/StackGuardian/terraform-provider-stackguardian/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

// This file holds every stackguardian_stack test that switches
// template_group_id from one stack template revision to another — a REVISION
// SWITCH test, in each function's own doc comment — one section per
// attribute, each preceded by a comment documenting that attribute's
// resolution rules across a revision change. Tests against a single,
// unchanging revision live in resource_attributes_test.go instead.
// workflows_config has its own pair of files (resource_workflows_test.go /
// resource_stack_upgrade_test.go) and isn't covered here.

// --- root ---

// TestAccStack_TemplateGroupIdReResolution — REVISION SWITCH test.
// Purpose: description and actions are left unset in config, so both start resolved from
// revision1; template_group_id then moves to revision2, which defines its OWN description and
// actions. Both fields must re-resolve to revision2's values, not stay stuck on revision1's (see
// ModifyPlan/reResolveOnRevisionChange).
//
// Mechanism this guards, for actions specifically: revision1 and revision2 both define their
// own apply/plan Actions verbatim (see setupStackTemplateChain/setupSecondStackTemplateRevision),
// so reResolveOnRevisionChange's actions block (which only re-resolves when the NEW revision has
// its own Actions — see expandActionsMap for why the provider never synthesizes a fresh action
// set) picks up revision2's set directly. Step 2's actions assertions confirm that. There is no
// "switching to a revision with no actions of its own" case to test separately — the API rejects
// publishing a stack template revision with an empty Actions map, so every revision a stack can
// reference always has at least one action (see resource_attributes_test.go's "actions resolution
// rules" comment).
func TestAccStack_TemplateGroupIdReResolution(t *testing.T) {
	wfGrpName := "tf-provider-stack-tmplswitch-wfgrp"
	wfTemplateName := "tf-provider-stack-tmplswitch-wftmpl"
	stackTemplateName := "tf-provider-stack-tmplswitch-stmpl"
	id := "tf-provider-stack-tmplswitch"

	revision1 := setupStackDependencyChain(t, wfGrpName, wfTemplateName, stackTemplateName, id)
	revision2 := setupSecondStackTemplateRevision(t, stackTemplateName, wfTemplateName, "revision two description", nil)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader()),
		Steps: []resource.TestStep{
			{
				// description left unset — revision1 has none, so it stays empty.
				// actions left unset too — revision1's own apply/plan are inherited
				// verbatim at CREATE time (rule 2 in resource_attributes_test.go's
				// "actions resolution rules" comment; Step 2 below re-proves it
				// across a revision switch instead).
				Config: testAccStackConfig(wfGrpName, revision1, id, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_stack.test", "template_group_id", revision1),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "actions.apply.name", "apply"),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "actions.plan.name", "plan"),
				),
			},
			{
				// Switch to revision2 — description must re-resolve to revision2's
				// value, not stay stuck at revision1's (empty).
				Config: testAccStackConfig(wfGrpName, revision2, id, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_stack.test", "template_group_id", revision2),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "description", "revision two description"),
					// revision2's own Actions, copied verbatim (no generation).
					resource.TestCheckResourceAttr("stackguardian_stack.test", "actions.apply.name", "apply"),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "actions.plan.name", "plan"),
				),
			},
		},
	})
}

// --- actions ---

// TestAccStack_ActionsRevisionRemovedWorkflow — REVISION SWITCH test.
// Purpose: template_group_id moves to a revision that dropped a workflow slot the user's own
// actions still references. That switch must be rejected at plan time
// (validateActionsAgainstRevision), rather than sending a dangling reference the API would
// reject with a less actionable error.
//
// Setup: revision1 (setupStackTemplateChainNoActions) wires testWfSlotId and secondWfSlotId;
// revision2 (setupSecondStackTemplateRevision) only re-declares testWfSlotId, so secondWfSlotId
// is the removed workflow.
func TestAccStack_ActionsRevisionRemovedWorkflow(t *testing.T) {
	wfGrpName := "tf-provider-stack-actrmwf-wfgrp"
	wfTemplateName := "tf-provider-stack-actrmwf-wftmpl"
	stackTemplateName := "tf-provider-stack-actrmwf-stmpl"
	id := "tf-provider-stack-actrmwf"

	t.Cleanup(func() {
		logCleanupErr(t, fmt.Sprintf("delete workflow group %q", wfGrpName), deleteWorkflowGroupFixture(wfGrpName))
	})
	if err := createWorkflowGroupFixture(wfGrpName); err != nil && !is409(err) {
		t.Fatalf("TestAccStack_ActionsRevisionRemovedWorkflow: create workflow group %q: %s", wfGrpName, err)
	}
	workflowTemplateID := setupStackWorkflowTemplate(t, wfTemplateName)
	revision1 := setupStackTemplateChainNoActions(t, stackTemplateName, workflowTemplateID)
	revision2 := setupSecondStackTemplateRevision(t, stackTemplateName, workflowTemplateID, "revision two", nil)
	t.Cleanup(func() { deleteStackFixture(wfGrpName, id) })

	// revision1 wires two slots (testWfSlotId, secondWfSlotId); revision2 only
	// re-declares testWfSlotId — workflows_config must match whichever
	// revision is active in that step (validateWorkflowsConfigMatchesRevision).
	revision1WorkflowsConfig := fmt.Sprintf(`
  workflows_config = {
    workflows = [
      { id = %q },
      { id = %q }
    ]
  }
`, testWfSlotId, secondWfSlotId)
	revision2WorkflowsConfig := fmt.Sprintf(`
  workflows_config = {
    workflows = [
      { id = %q }
    ]
  }
`, testWfSlotId)

	referencesSecondSlot := fmt.Sprintf(`
  actions = {
    notify = {
      name = "notify"
      order = {
        %[1]q = {
          parameters = {
            terraform_action = {
              action = "apply"
            }
          }
        }
      }
    }
  }
`, secondWfSlotId)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader()),
		Steps: []resource.TestStep{
			{
				Config: testAccStackConfig(wfGrpName, revision1, id, revision1WorkflowsConfig+referencesSecondSlot),
				Check:  resource.TestCheckResourceAttr("stackguardian_stack.test", "actions.notify.name", "notify"),
			},
			{
				// revision2 dropped secondWfSlotId — actions still references it, so
				// the switch must be rejected.
				Config:      testAccStackConfig(wfGrpName, revision2, id, revision2WorkflowsConfig+referencesSecondSlot),
				ExpectError: regexp.MustCompile("actions references a removed workflow"),
			},
		},
	})
}

// --- context_tags ---

// TestAccStack_ContextTags_ClearedWhenTemplateHasNone — REVISION SWITCH test.
// Purpose: context_tags is left unset in config, so it starts resolved from a template
// revision that HAS context_tags; template_group_id then moves to a revision with NONE.
// context_tags must re-resolve to {} on the new revision, not stay stuck on the old one's
// value (scenario 4 in TestAccStack_ContextTags_Resolution's doc comment, resource_attributes_test.go).
//
// Mechanism this guards: reResolveOnRevisionChange must ALWAYS re-derive context_tags fresh
// from the new revision when the user left them unset, landing on a known-empty {} when the
// new revision has none, so ToUpdateAPIModel's known-non-null branch sends that {} as a real,
// explicit clear rather than leaving the old context_tags in place. Same failure mode as the
// description/tags versions of this test.
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

// --- description ---

// TestAccStack_Description_ClearedWhenTemplateHasNone — REVISION SWITCH test.
// Purpose: description is left unset in config, so it starts resolved from a template revision
// that HAS a description; template_group_id then moves to a revision with NONE. description
// must re-resolve to "" on the new revision, not stay stuck on the old one's value (scenario 4
// in TestAccStack_Description_Resolution's doc comment, resource_attributes_test.go).
//
// Mechanism this guards: reResolveOnRevisionChange must ALWAYS re-derive description fresh from
// the new revision when the user left it unset — landing on a known "" (via
// knownEmptyStringIfNull) when the new revision has none, never an actual null and never the
// stale old value. ToUpdateAPIModel's known-non-null branch then sends that "" as a real,
// explicit clear. If reResolveOnRevisionChange regressed to producing an actual null instead
// (e.g. by reverting to plain flatteners.StringPtr without the knownEmptyStringIfNull wrapper),
// this step would fail with "Provider produced inconsistent result after apply" instead of the
// assertion below ever running, since ToUpdateAPIModel would then omit the field (leaving the
// old value in place) while the plan had predicted null.
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

// --- tags ---

// TestAccStack_Tags_ClearedWhenTemplateHasNone — REVISION SWITCH test.
// Purpose: tags is left unset in config, so it starts resolved from a template revision that
// HAS tags; template_group_id then moves to a revision with NONE. tags must re-resolve to []
// on the new revision, not stay stuck on the old one's value (scenario 4 in
// TestAccStack_Tags_Resolution's doc comment, resource_attributes_test.go).
//
// Mechanism this guards: reResolveOnRevisionChange must ALWAYS re-derive tags fresh from the
// new revision when the user left them unset, landing on a known-empty [] when the new revision
// has none, so ToUpdateAPIModel's known-non-null branch sends that [] as a real, explicit clear
// rather than leaving the old tags in place. Same failure mode as
// TestAccStack_Description_ClearedWhenTemplateHasNone.
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
