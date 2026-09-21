package stack_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/StackGuardian/terraform-provider-stackguardian/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

// This file holds every stackguardian_stack test that runs against a SINGLE,
// unchanging stack template revision — no template_group_id switch — one
// section per attribute, each preceded by a comment documenting that
// attribute's resolution rules. Tests that switch template_group_id live in
// resource_attributes_stack_upgrade_test.go instead. workflows_config has its
// own pair of files (resource_workflows_test.go /
// resource_stack_upgrade_test.go) and isn't covered here.

// --- actions ---

// actions resolution rules (source material for docs-templates/resources/stack.md.tmpl):
//
//  1. Declaring actions in the resource config always wins — it wholesale
//     replaces whatever the stack template revision itself defines, not a
//     per-key merge (see TestAccStack_ActionsRoundTrip).
//  2. Leaving actions unset resolves it from the stack template revision's
//     own actions verbatim, if the revision has any — both at create time
//     (TestAccStack_TemplateGroupIdReResolution's Step 1, resource_attributes_stack_upgrade_test.go)
//     and across a revision change (that same test's Step 2).
//  3. An explicit empty value (actions = {}) is rejected at plan time by a
//     schema validator (mapvalidator.SizeAtLeast(1) on "actions", schema.go)
//     — a stack must always have at least one action, matching the API's own
//     requirement for stack template revisions ("Stack actions are empty").
//     This only fires for a known, explicitly-empty map — leaving the
//     attribute unset is unaffected (rule 2 above still applies) — see
//     TestAccStack_ActionsRejectsEmpty.
//  4. If neither the resource nor the template supplies actions, the API
//     applies its own default at create time (TestAccStack_ActionsGeneratedFromTemplate) —
//     the provider does not synthesize one itself — and that default persists
//     through later updates that don't touch actions, rather than being
//     dropped or regenerated (TestAccStack_ActionsGeneratedFromTemplate's
//     unrelated-update step).
//
// There is no "leaving actions unset, then switching to a revision with none
// of its own" scenario tested here: it's unreachable. The API rejects
// publishing a stack template revision with an empty Actions map ("Stack
// actions are empty"), so every revision a stack can actually reference
// always has at least one action.

// TestAccStack_ActionsGeneratedFromTemplate verifies that leaving "actions"
// unset in config, with a template that supplies no Actions of its own
// either (setupStackTemplateChainNoActions), resolves to the API's own
// create-time default apply/plan/destroy set — expandActionsMap omits
// "actions" from the request entirely in that case (see its doc comment:
// neither the resource nor the template has a value, so there's nothing to
// send, and the provider does not synthesize one itself), so this set is
// entirely the platform's own doing, not provider logic. The specific
// values asserted below (names, descriptions, dependency chaining across
// setupStackTemplateChainNoActions's two workflow slots) are therefore a
// live contract with the API, not something client-side code produces —
// see the project plan's note that this is the highest-risk test to re-run
// after that removal, since these assertions now exercise real API
// behavior for the first time rather than a deleted client-side replica.
// What IS still provider logic and exercised here: translateActionsOrderKeys
// (slot-uuid <-> real workflow-id substitution) and flattenActionsMap's full
// nested mapping (dependencies, conditions, terraform_action) on the
// round-trip Read below. Order map keys are the bare template slot ids
// (StackTemplateRevisionWorkflow.Id) — not the workflow's own post-creation
// resource id, since at create time that doesn't exist yet.
//
// It also verifies that once the user DOES declare "actions" in config, that
// value wholesale replaces the API's default set — expandActionsMap is an
// override, not a per-key merge, so plan/destroy (not redeclared) disappear
// rather than staying inherited alongside the user's own "apply".
func TestAccStack_ActionsGeneratedFromTemplate(t *testing.T) {
	wfGrpName := "tf-provider-stack-gendef-wfgrp"
	wfTemplateName := "tf-provider-stack-gendef-wftmpl"
	stackTemplateName := "tf-provider-stack-gendef-stmpl"
	id := "tf-provider-stack-gendef"

	t.Cleanup(func() {
		logCleanupErr(t, fmt.Sprintf("delete workflow group %q", wfGrpName), deleteWorkflowGroupFixture(wfGrpName))
	})
	if err := createWorkflowGroupFixture(wfGrpName); err != nil && !is409(err) {
		t.Fatalf("TestAccStack_ActionsGeneratedFromTemplate: create workflow group %q: %s", wfGrpName, err)
	}
	workflowTemplateID := setupStackWorkflowTemplate(t, wfTemplateName)
	revision := setupStackTemplateChainNoActions(t, stackTemplateName, workflowTemplateID)
	t.Cleanup(func() { deleteStackFixture(wfGrpName, id) })

	// setupStackTemplateChainNoActions wires two slots — testWfSlotId then
	// secondWfSlotId — so workflows_config must declare both, in that order,
	// on every step (validateWorkflowsConfigMatchesRevision).
	twoSlotWorkflowsConfig := fmt.Sprintf(`
  workflows_config = {
    workflows = [
      { id = %q },
      { id = %q }
    ]
  }
`, testWfSlotId, secondWfSlotId)

	// actions declared explicitly — replaces the generated set entirely, not
	// just its "apply" key.
	withActionsOverride := twoSlotWorkflowsConfig + fmt.Sprintf(`
  actions = {
    apply = {
      name = "apply"
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
`, testWfSlotId)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader()),
		Steps: []resource.TestStep{
			{
				// actions left unset — generated apply/plan/destroy, matching
				// default_actions.json's shape.
				Config: testAccStackConfig(wfGrpName, revision, id, twoSlotWorkflowsConfig),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_stack.test", "actions.apply.name", "Create"),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "actions.apply.description", "use this action to create resources in the stack"),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "actions.plan.name", "Plan"),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "actions.plan.description", "use this action to plan resources in the stack"),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "actions.destroy.name", "Destroy"),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "actions.destroy.description", "use this action to destroy resources in the stack"),

					// apply/plan chain in the template's own declaration order:
					// testWfSlotId first (no dependencies), secondWfSlotId depends on it.
					resource.TestCheckResourceAttr("stackguardian_stack.test",
						fmt.Sprintf("actions.apply.order.%s.dependencies.#", testWfSlotId), "0"),
					resource.TestCheckResourceAttr("stackguardian_stack.test",
						fmt.Sprintf("actions.apply.order.%s.dependencies.0.id", secondWfSlotId), testWfSlotId),
					resource.TestCheckResourceAttr("stackguardian_stack.test",
						fmt.Sprintf("actions.apply.order.%s.dependencies.0.condition.latest_status", secondWfSlotId), "COMPLETED"),
					resource.TestCheckResourceAttr("stackguardian_stack.test",
						fmt.Sprintf("actions.apply.order.%s.parameters.terraform_action.action", testWfSlotId), "apply"),
					resource.TestCheckResourceAttr("stackguardian_stack.test",
						fmt.Sprintf("actions.plan.order.%s.parameters.terraform_action.action", testWfSlotId), "plan"),

					// destroy chains in REVERSE: secondWfSlotId first (no dependencies),
					// testWfSlotId depends on it.
					resource.TestCheckResourceAttr("stackguardian_stack.test",
						fmt.Sprintf("actions.destroy.order.%s.dependencies.#", secondWfSlotId), "0"),
					resource.TestCheckResourceAttr("stackguardian_stack.test",
						fmt.Sprintf("actions.destroy.order.%s.dependencies.0.id", testWfSlotId), secondWfSlotId),
					resource.TestCheckResourceAttr("stackguardian_stack.test",
						fmt.Sprintf("actions.destroy.order.%s.parameters.terraform_action.action", secondWfSlotId), "destroy"),
				),
			},
			{
				// Round trips with no diff.
				Config:   testAccStackConfig(wfGrpName, revision, id, twoSlotWorkflowsConfig),
				PlanOnly: true,
			},
			{
				// An unrelated update (description) — actions still left unset. The
				// API-generated default from the first step must persist through
				// this update, not be dropped or regenerated: the carried-forward
				// plan value is known and non-null, so expandActionsMap treats it as
				// a real declaration and re-sends it verbatim (see reResolveOnRevisionChange's
				// doc comment in model.go for why "unchanged" is the correct
				// prediction here).
				Config: testAccStackConfig(wfGrpName, revision, id, twoSlotWorkflowsConfig+`description = "updated after create"`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_stack.test", "description", "updated after create"),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "actions.apply.name", "Create"),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "actions.plan.name", "Plan"),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "actions.destroy.name", "Destroy"),
				),
			},
			{
				// User now declares actions explicitly — the whole generated set is
				// replaced, not merged key-by-key: plan/destroy must vanish since
				// they weren't redeclared.
				Config: testAccStackConfig(wfGrpName, revision, id, withActionsOverride),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_stack.test", "actions.apply.name", "apply"),
					resource.TestCheckNoResourceAttr("stackguardian_stack.test", "actions.plan"),
					resource.TestCheckNoResourceAttr("stackguardian_stack.test", "actions.destroy"),
				),
			},
			{
				// And the post-update state round trips with no diff.
				Config:   testAccStackConfig(wfGrpName, revision, id, withActionsOverride),
				PlanOnly: true,
			},
		},
	})
}

// TestAccStack_ActionsRoundTrip covers actions' nested shape —
// order[].parameters.terraform_action.action, order[].parameters.
// environment_variables, and order[].dependencies (id, condition.
// latest_status) — round tripping stably, and that removing an action from
// actions on update actually drops it from the payload rather than leaving it
// orphaned. It also confirms the override is total: setupStackDependencyChain's
// template defines its own "apply"/"plan" Actions, but since this test
// declares "actions" explicitly, none of the template's own actions (e.g.
// "plan") are merged in — only what the resource itself declares exists.
// deployment_platform_config and wf_steps_config inside order[].parameters
// aren't covered here — both need a real integration_id / workflow step
// template fixture this test doesn't set up.
func TestAccStack_ActionsRoundTrip(t *testing.T) {
	wfGrpName := "tf-provider-stack-actrt-wfgrp"
	wfTemplateName := "tf-provider-stack-actrt-wftmpl"
	stackTemplateName := "tf-provider-stack-actrt-stmpl"
	id := "tf-provider-stack-actrt"

	revision := setupStackDependencyChain(t, wfGrpName, wfTemplateName, stackTemplateName, id)

	withDestroy := fmt.Sprintf(`
  actions = {
    apply = {
      name        = "apply"
      description = "Custom apply action"
      order = {
        %[1]q = {
          parameters = {
            terraform_action = {
              action = "apply"
            }
            environment_variables = [
              {
                kind = "PLAIN_TEXT"
                config = {
                  var_name   = "ACTION_VAR"
                  text_value = "action-value"
                }
              }
            ]
          }
        }
      }
    }
    destroy = {
      name = "destroy"
      order = {
        %[1]q = {
          parameters = {
            terraform_action = {
              action = "destroy"
            }
          }
          dependencies = [
            {
              id = %[1]q
              condition = {
                latest_status = "COMPLETED"
              }
            }
          ]
        }
      }
    }
  }
`, testWfSlotId)

	withoutDestroy := fmt.Sprintf(`
  actions = {
    apply = {
      name = "apply"
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
`, testWfSlotId)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader()),
		Steps: []resource.TestStep{
			{
				Config: testAccStackConfig(wfGrpName, revision, id, withDestroy),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_stack.test", "actions.apply.name", "apply"),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "actions.apply.description", "Custom apply action"),
					resource.TestCheckResourceAttr("stackguardian_stack.test",
						fmt.Sprintf("actions.apply.order.%s.parameters.terraform_action.action", testWfSlotId), "apply"),
					resource.TestCheckResourceAttr("stackguardian_stack.test",
						fmt.Sprintf("actions.apply.order.%s.parameters.environment_variables.0.config.var_name", testWfSlotId), "ACTION_VAR"),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "actions.destroy.name", "destroy"),
					resource.TestCheckResourceAttr("stackguardian_stack.test",
						fmt.Sprintf("actions.destroy.order.%s.dependencies.0.id", testWfSlotId), testWfSlotId),
					resource.TestCheckResourceAttr("stackguardian_stack.test",
						fmt.Sprintf("actions.destroy.order.%s.dependencies.0.condition.latest_status", testWfSlotId), "COMPLETED"),
					// The template's own "plan" action is NOT inherited — actions is
					// declared, so it wholesale replaces the template's value.
					resource.TestCheckNoResourceAttr("stackguardian_stack.test", "actions.plan"),
				),
			},
			{
				// Round trips with no diff.
				Config:   testAccStackConfig(wfGrpName, revision, id, withDestroy),
				PlanOnly: true,
			},
			{
				// Remove "destroy" — must actually disappear, not linger.
				Config: testAccStackConfig(wfGrpName, revision, id, withoutDestroy),
				Check:  resource.TestCheckNoResourceAttr("stackguardian_stack.test", "actions.destroy"),
			},
		},
	})
}

// TestAccStack_ActionsRejectsEmpty verifies actions = {} (a known, explicitly empty map — not
// the same as leaving the attribute unset) is rejected at plan time by the schema's
// mapvalidator.SizeAtLeast(1) on "actions" (schema.go), rather than reaching the API and
// surfacing its own less actionable "Stack actions are empty" error.
func TestAccStack_ActionsRejectsEmpty(t *testing.T) {
	wfGrpName := "tf-provider-stack-actempty-wfgrp"
	wfTemplateName := "tf-provider-stack-actempty-wftmpl"
	stackTemplateName := "tf-provider-stack-actempty-stmpl"
	id := "tf-provider-stack-actempty"

	revision := setupStackDependencyChain(t, wfGrpName, wfTemplateName, stackTemplateName, id)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader()),
		Steps: []resource.TestStep{
			{
				Config:      testAccStackConfig(wfGrpName, revision, id, `actions = {}`),
				ExpectError: regexp.MustCompile(`map must contain at least 1 elements`),
			},
		},
	})
}

// --- context_tags ---

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
//     resolves to {} (see TestAccStack_ContextTags_ClearedWhenTemplateHasNone,
//     resource_attributes_stack_upgrade_test.go).
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

// --- description ---

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
//     TestAccStack_Description_ClearedWhenTemplateHasNone,
//     resource_attributes_stack_upgrade_test.go).
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

// --- tags ---

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
//     TestAccStack_Tags_ClearedWhenTemplateHasNone,
//     resource_attributes_stack_upgrade_test.go).
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
