package stack_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/StackGuardian/terraform-provider-stackguardian/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
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
// setupStackTemplateChainNoActions's two workflows) are therefore a
// live contract with the API, not something client-side code produces —
// see the project plan's note that this is the highest-risk test to re-run
// after that removal, since these assertions now exercise real API
// behavior for the first time rather than a deleted client-side replica.
// What IS still provider logic and exercised here: translateActionsOrderKeys
// (workflow UUID <-> real workflow id substitution) and flattenActionsMap's full
// nested mapping (dependencies, conditions, terraform_action) on the
// round-trip Read below. Order map keys are the bare workflow UUIDs
// (StackTemplateRevisionWorkflow.Id) — not the workflow's own post-creation
// resource id, since at create time that doesn't exist yet.
//
// It also verifies that once the user DOES declare "actions" in config, that
// value wholesale replaces the API's default set — expandActionsMap is an
// override, not a per-key merge, so plan/destroy (not redeclared) disappear
// rather than staying inherited alongside the user's own "apply".
func TestAccStack_ActionsGeneratedFromTemplate(t *testing.T) {
	acctest.SkipUnlessAcceptance(t)
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

	// setupStackTemplateChainNoActions declares two workflows — testWorkflowUUID then
	// secondWorkflowUUID — so workflows_config must declare both, in that order,
	// on every step (validateWorkflowsConfigMatchesRevision).
	twoWorkflowsConfig := fmt.Sprintf(`
  workflows_config = {
    workflows = [
      { id = %q },
      { id = %q }
    ]
  }
`, testWorkflowUUID, secondWorkflowUUID)

	// actions declared explicitly — replaces the generated set entirely, not
	// just its "apply" key.
	withActionsOverride := twoWorkflowsConfig + fmt.Sprintf(`
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
`, testWorkflowUUID)

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
				Config: testAccStackConfig(wfGrpName, revision, id, twoWorkflowsConfig),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_stack.test", "actions.apply.name", "Create"),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "actions.apply.description", "use this action to create resources in the stack"),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "actions.plan.name", "Plan"),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "actions.plan.description", "use this action to plan resources in the stack"),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "actions.destroy.name", "Destroy"),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "actions.destroy.description", "use this action to destroy resources in the stack"),

					// apply/plan chain in the template's own declaration order:
					// testWorkflowUUID first (no dependencies), secondWorkflowUUID depends on it.
					resource.TestCheckResourceAttr("stackguardian_stack.test",
						fmt.Sprintf("actions.apply.order.%s.dependencies.#", testWorkflowUUID), "0"),
					resource.TestCheckResourceAttr("stackguardian_stack.test",
						fmt.Sprintf("actions.apply.order.%s.dependencies.0.id", secondWorkflowUUID), testWorkflowUUID),
					resource.TestCheckResourceAttr("stackguardian_stack.test",
						fmt.Sprintf("actions.apply.order.%s.dependencies.0.condition.latest_status", secondWorkflowUUID), "COMPLETED"),
					resource.TestCheckResourceAttr("stackguardian_stack.test",
						fmt.Sprintf("actions.apply.order.%s.parameters.terraform_action.action", testWorkflowUUID), "apply"),
					resource.TestCheckResourceAttr("stackguardian_stack.test",
						fmt.Sprintf("actions.plan.order.%s.parameters.terraform_action.action", testWorkflowUUID), "plan"),

					// destroy chains in REVERSE: secondWorkflowUUID first (no dependencies),
					// testWorkflowUUID depends on it.
					resource.TestCheckResourceAttr("stackguardian_stack.test",
						fmt.Sprintf("actions.destroy.order.%s.dependencies.#", secondWorkflowUUID), "0"),
					resource.TestCheckResourceAttr("stackguardian_stack.test",
						fmt.Sprintf("actions.destroy.order.%s.dependencies.0.id", testWorkflowUUID), secondWorkflowUUID),
					resource.TestCheckResourceAttr("stackguardian_stack.test",
						fmt.Sprintf("actions.destroy.order.%s.parameters.terraform_action.action", secondWorkflowUUID), "destroy"),
				),
			},
			{
				// Round trips with no diff.
				Config:   testAccStackConfig(wfGrpName, revision, id, twoWorkflowsConfig),
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
				Config: testAccStackConfig(wfGrpName, revision, id, twoWorkflowsConfig+`description = "updated after create"`),
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

// checkStackActionDefaults checks each named action's "default" flag both in
// Terraform state (actions.<name>.default) and on the platform, read through the
// SDK, so a state value that merely agrees with the plan can't hide what was
// actually stored. want maps action name -> expected flag; a nil Default from the
// API counts as false.
func checkStackActionDefaults(wfGrpName, id string, want map[string]bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for name, wantDefault := range want {
			if err := resource.TestCheckResourceAttr("stackguardian_stack.test",
				fmt.Sprintf("actions.%s.default", name), fmt.Sprintf("%t", wantDefault))(s); err != nil {
				return err
			}
		}
		readResp, err := getClient().Stacks.ReadStack(context.TODO(), org, id, wfGrpName, nil)
		if err != nil {
			return fmt.Errorf("read stack %q: %w", id, err)
		}
		if readResp == nil || readResp.Msg == nil {
			return fmt.Errorf("read stack %q: empty response", id)
		}
		for name, wantDefault := range want {
			action, ok := readResp.Msg.Actions[name]
			if !ok || action == nil {
				return fmt.Errorf("action %q not found on stack %q", name, id)
			}
			got := action.Default != nil && *action.Default
			if got != wantDefault {
				return fmt.Errorf("action %q: default = %v, want %v", name, got, wantDefault)
			}
		}
		return nil
	}
}

// TestAccStack_ActionsKeepDefaultFlag verifies each action's "default" flag
// survives updates. In core, default=true marks an action that comes from the
// stack template (replaced or dropped when a stack is upgraded without sending
// Actions, e.g. from the UI), and default=false marks a custom action (kept
// across upgrades); see merge_actions_for_upgrade in core. The provider
// re-sends the stored actions on every update, so it must keep the platform's
// flag instead of stamping every action as custom.
//
// The flag is exposed as the Computed actions[*].default, planned by
// actionDefaultModifier; checkStackActionDefaults checks state and the API.
//
//  1. Create with actions unset: the revision has no actions of its own
//     (setupStackTemplateChainNoActions), so the platform generates
//     apply/plan/destroy with default=true. This step checks that assumption.
//  2. An unrelated update (description) must leave all three default=true.
//  3. Declaring actions makes them the user's own: default=false.
//  4. No longer declaring actions keeps them as the user's (default=false),
//     carried forward from state.
func TestAccStack_ActionsKeepDefaultFlag(t *testing.T) {
	acctest.SkipUnlessAcceptance(t)
	wfGrpName := acctest.ResourceName("stack-actdef-wfgrp")
	wfTemplateName := acctest.ResourceName("stack-actdef-wftmpl")
	stackTemplateName := acctest.ResourceName("stack-actdef-stmpl")
	id := acctest.ResourceName("stack-actdef")

	t.Cleanup(func() {
		logCleanupErr(t, fmt.Sprintf("delete workflow group %q", wfGrpName), deleteWorkflowGroupFixture(wfGrpName))
	})
	if err := createWorkflowGroupFixture(wfGrpName); err != nil && !is409(err) {
		t.Fatalf("TestAccStack_ActionsKeepDefaultFlag: create workflow group %q: %s", wfGrpName, err)
	}
	workflowTemplateID := setupStackWorkflowTemplate(t, wfTemplateName)
	revision := setupStackTemplateChainNoActions(t, stackTemplateName, workflowTemplateID)
	t.Cleanup(func() { deleteStackFixture(wfGrpName, id) })

	// setupStackTemplateChainNoActions declares two workflows, in this order.
	twoWorkflowsConfig := fmt.Sprintf(`
  workflows_config = {
    workflows = [
      { id = %q },
      { id = %q }
    ]
  }
`, testWorkflowUUID, secondWorkflowUUID)

	withDeclaredApply := twoWorkflowsConfig + fmt.Sprintf(`
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
`, testWorkflowUUID)

	allTemplateDefaults := map[string]bool{"apply": true, "plan": true, "destroy": true}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader()),
		Steps: []resource.TestStep{
			{
				Config: testAccStackConfig(wfGrpName, revision, id, twoWorkflowsConfig),
				Check:  checkStackActionDefaults(wfGrpName, id, allTemplateDefaults),
			},
			{
				// Unrelated update, actions still unset.
				Config: testAccStackConfig(wfGrpName, revision, id, twoWorkflowsConfig+`description = "updated after create"`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_stack.test", "description", "updated after create"),
					checkStackActionDefaults(wfGrpName, id, allTemplateDefaults),
				),
			},
			{
				// Round trips with no diff.
				Config:   testAccStackConfig(wfGrpName, revision, id, twoWorkflowsConfig+`description = "updated after create"`),
				PlanOnly: true,
			},
			{
				// Declared actions are the user's own.
				Config: testAccStackConfig(wfGrpName, revision, id, withDeclaredApply+`description = "updated after create"`),
				Check:  checkStackActionDefaults(wfGrpName, id, map[string]bool{"apply": false}),
			},
			{
				// No longer declared: still the user's own, carried from state.
				Config: testAccStackConfig(wfGrpName, revision, id, twoWorkflowsConfig+`description = "declared actions removed"`),
				Check:  checkStackActionDefaults(wfGrpName, id, map[string]bool{"apply": false}),
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
// deployment_platform_config and wf_steps_config inside order[].parameters are
// covered by TestAccStack_ActionsParametersStepsAndDeploymentPlatform.
func TestAccStack_ActionsRoundTrip(t *testing.T) {
	acctest.SkipUnlessAcceptance(t)
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
`, testWorkflowUUID)

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
`, testWorkflowUUID)

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
						fmt.Sprintf("actions.apply.order.%s.parameters.terraform_action.action", testWorkflowUUID), "apply"),
					resource.TestCheckResourceAttr("stackguardian_stack.test",
						fmt.Sprintf("actions.apply.order.%s.parameters.environment_variables.0.config.var_name", testWorkflowUUID), "ACTION_VAR"),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "actions.destroy.name", "destroy"),
					resource.TestCheckResourceAttr("stackguardian_stack.test",
						fmt.Sprintf("actions.destroy.order.%s.dependencies.0.id", testWorkflowUUID), testWorkflowUUID),
					resource.TestCheckResourceAttr("stackguardian_stack.test",
						fmt.Sprintf("actions.destroy.order.%s.dependencies.0.condition.latest_status", testWorkflowUUID), "COMPLETED"),
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

// TestAccStack_ActionsParametersStepsAndDeploymentPlatform covers the two
// order[].parameters fields TestAccStack_ActionsRoundTrip doesn't:
// wf_steps_config and deployment_platform_config. Each goes through its own
// expand/flatten path, separate from the workflow-level fields of the same
// name, and deployment_platform_config's shape changed from a JSON string to
// an object with integration_id/profile_name, so both are checked to round
// trip with no diff. wf_step_template_id needs a real workflow step template
// revision (setupStackWorkflowStepTemplate); integration_id is a placeholder,
// as in TestAccStack_WorkflowsConfigRoundTripRemainingFields, since nothing
// runs the action.
func TestAccStack_ActionsParametersStepsAndDeploymentPlatform(t *testing.T) {
	acctest.SkipUnlessAcceptance(t)
	wfGrpName := acctest.ResourceName("stack-actparams-wfgrp")
	wfTemplateName := acctest.ResourceName("stack-actparams-wftmpl")
	stackTemplateName := acctest.ResourceName("stack-actparams-stmpl")
	id := acctest.ResourceName("stack-actparams")

	revision := setupStackDependencyChain(t, wfGrpName, wfTemplateName, stackTemplateName, id)
	stepTemplateID := setupStackWorkflowStepTemplate(t, acctest.ResourceName("stack-actparams-step"))

	config := fmt.Sprintf(`
  actions = {
    apply = {
      name = "apply"
      order = {
        %[1]q = {
          parameters = {
            terraform_action = {
              action = "apply"
            }
            wf_steps_config = [
              {
                name                = "action-step"
                wf_step_template_id = %[2]q
                approval            = true
                timeout             = 300
                cmd_override        = "echo from-action"

                environment_variables = [
                  {
                    kind = "PLAIN_TEXT"
                    config = {
                      var_name   = "ACTION_STEP_VAR"
                      text_value = "action-step-value"
                    }
                  }
                ]
              }
            ]
            deployment_platform_config = [
              {
                kind = "AWS_RBAC"
                config = {
                  integration_id = "/integrations/test-integration"
                  profile_name   = "action-profile"
                }
              }
            ]
          }
        }
      }
    }
  }
`, testWorkflowUUID, stepTemplateID)

	const res = "stackguardian_stack.test"
	params := fmt.Sprintf("actions.apply.order.%s.parameters.", testWorkflowUUID)
	step := params + "wf_steps_config.0."
	dpc := params + "deployment_platform_config.0."

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader()),
		Steps: []resource.TestStep{
			{
				Config: testAccStackConfig(wfGrpName, revision, id, config),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(res, params+"terraform_action.action", "apply"),
					resource.TestCheckResourceAttr(res, params+"wf_steps_config.#", "1"),
					resource.TestCheckResourceAttr(res, step+"name", "action-step"),
					resource.TestCheckResourceAttr(res, step+"wf_step_template_id", stepTemplateID),
					resource.TestCheckResourceAttr(res, step+"approval", "true"),
					resource.TestCheckResourceAttr(res, step+"timeout", "300"),
					resource.TestCheckResourceAttr(res, step+"cmd_override", "echo from-action"),
					resource.TestCheckResourceAttr(res, step+"environment_variables.0.config.var_name", "ACTION_STEP_VAR"),
					resource.TestCheckResourceAttr(res, step+"environment_variables.0.config.text_value", "action-step-value"),
					resource.TestCheckResourceAttr(res, params+"deployment_platform_config.#", "1"),
					resource.TestCheckResourceAttr(res, dpc+"kind", "AWS_RBAC"),
					resource.TestCheckResourceAttr(res, dpc+"config.integration_id", "/integrations/test-integration"),
					resource.TestCheckResourceAttr(res, dpc+"config.profile_name", "action-profile"),
				),
			},
			{
				// Round trips with no diff.
				Config:   testAccStackConfig(wfGrpName, revision, id, config),
				PlanOnly: true,
			},
		},
	})
}

// TestAccStack_ActionsRejectsEmpty verifies actions = {} (a known, explicitly empty map — not
// the same as leaving the attribute unset) is rejected at plan time by the schema's
// mapvalidator.SizeAtLeast(1) on "actions" (schema.go), rather than reaching the API and
// surfacing its own less actionable "Stack actions are empty" error.
func TestAccStack_ActionsRejectsEmpty(t *testing.T) {
	acctest.SkipUnlessAcceptance(t)
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
	acctest.SkipUnlessAcceptance(t)
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
	acctest.SkipUnlessAcceptance(t)
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
	acctest.SkipUnlessAcceptance(t)
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
