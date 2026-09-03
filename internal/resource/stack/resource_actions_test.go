package stack_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	sgsdkgo "github.com/StackGuardian/sg-sdk-go"
	"github.com/StackGuardian/sg-sdk-go/stacktemplaterevisions"
	"github.com/StackGuardian/sg-sdk-go/stacktemplates"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

// secondWfSlotId is a second workflow slot UUID, distinct from testWfSlotId,
// used only by TestAccStack_ActionsGeneratedFromTemplate to exercise the
// dependency chaining across more than one workflow.
const secondWfSlotId = "3f7c9e2a-5b1d-4e6f-8a2c-9d4b6e1f0a3c"

// setupStackTemplateChainNoActions creates and publishes a stack template +
// revision :1 via the SDK, like setupStackTemplateChain, but wires TWO
// workflow slots (testWfSlotId, secondWfSlotId — both pointing at the same
// workflow template) instead of one, and defines no Actions of its own at
// all. Used only by TestAccStack_ActionsGeneratedFromTemplate, which needs a
// template that supplies neither apply/plan/destroy NOR a dependency chain to
// inherit, so the only source for them is the API's own create-time default
// (the provider no longer synthesizes one itself — see expandActionsMap),
// and needs a second workflow to prove that default actually chains multiple
// workflows rather than only handling the trivial single-workflow case.
// Registers cleanup. Returns the bare revision id ("<name>:1").
func setupStackTemplateChainNoActions(t *testing.T, stackTemplateID, workflowTemplateID string) string {
	t.Helper()
	client := getClient()
	revisionID := fmt.Sprintf("%s:1", stackTemplateID)
	sourceConfigKind := stacktemplates.StackTemplateSourceConfigKindTerraform

	t.Cleanup(func() {
		logCleanupErr(t, fmt.Sprintf("deprecate stack template revision %q", revisionID), deprecateStackTemplateRevisionFixture(revisionID))
		logCleanupErr(t, fmt.Sprintf("delete stack template revision %q", revisionID), deleteStackTemplateRevisionFixture(revisionID))
		logCleanupErr(t, fmt.Sprintf("delete stack template %q", stackTemplateID), deleteStackTemplateFixture(stackTemplateID))
	})

	_, err := client.StackTemplates.CreateStackTemplate(
		context.TODO(), org, false,
		&stacktemplates.CreateStackTemplateRequest{
			Id:               &stackTemplateID,
			TemplateName:     stackTemplateID,
			SourceConfigKind: &sourceConfigKind,
			OwnerOrg:         fmt.Sprintf("/orgs/%s", org),
		},
	)
	if err != nil && !is409(err) {
		t.Fatalf("setupStackTemplateChainNoActions: create stack template %q: %s", stackTemplateID, err)
	}

	prefixedWorkflowTemplateID := fmt.Sprintf("/%s/%s", org, workflowTemplateID)
	prefixedWorkflowRevisionID := fmt.Sprintf("/%s/%s:1", org, workflowTemplateID)
	useMarketplace := true
	managedState := true
	tfVersion := "1.5.7"

	makeSlot := func(slotId, resourceName string) *stacktemplaterevisions.StackTemplateRevisionWorkflow {
		return &stacktemplaterevisions.StackTemplateRevisionWorkflow{
			Id:           sgsdkgo.String(slotId),
			TemplateId:   &prefixedWorkflowTemplateID,
			ResourceName: sgsdkgo.String(resourceName),
			VcsConfig: &sgsdkgo.VcsConfig{
				IacVcsConfig: &sgsdkgo.IacvcsConfig{
					UseMarketplaceTemplate: &useMarketplace,
					IacTemplateId:          &prefixedWorkflowRevisionID,
				},
			},
			TerraformConfig: &sgsdkgo.TerraformConfig{
				ManagedTerraformState: &managedState,
				TerraformVersion:      &tfVersion,
			},
		}
	}

	_, err = client.StackTemplateRevisions.CreateStackTemplateRevision(
		context.TODO(), org, stackTemplateID,
		&stacktemplaterevisions.CreateStackTemplateRevisionRequest{
			Alias:            "v1",
			SourceConfigKind: &sourceConfigKind,
			IsPublic:         sgsdkgo.IsPublicEnumZero.Ptr(),
			OwnerOrg:         fmt.Sprintf("/orgs/%s", org),
			WorkflowsConfig: &stacktemplaterevisions.StackTemplateRevisionWorkflowsConfig{
				Workflows: []*stacktemplaterevisions.StackTemplateRevisionWorkflow{
					makeSlot(testWfSlotId, "wf-1"),
					makeSlot(secondWfSlotId, "wf-2"),
				},
			},
			// Deliberately no Actions — the template supplies none of its own, so
			// apply/plan/destroy can only come from the API's own create-time default.
		},
	)
	if err != nil && !is409(err) {
		t.Fatalf("setupStackTemplateChainNoActions: create revision for %q: %s", stackTemplateID, err)
	}

	_, err = client.StackTemplateRevisions.UpdateStackTemplateRevision(
		context.TODO(), org, revisionID,
		&stacktemplaterevisions.UpdateStackTemplateRevisionRequest{
			IsPublic: sgsdkgo.Optional(sgsdkgo.IsPublicEnumOne),
		},
	)
	if err != nil {
		t.Fatalf("setupStackTemplateChainNoActions: publish revision %q: %s", revisionID, err)
	}

	_, err = client.StackTemplates.UpdateStackTemplate(
		context.TODO(), org, stackTemplateID,
		&stacktemplates.UpdateStackTemplateRequest{
			IsPublic: sgsdkgo.Optional(sgsdkgo.IsPublicEnumOne),
		},
	)
	if err != nil {
		t.Fatalf("setupStackTemplateChainNoActions: publish template %q: %s", stackTemplateID, err)
	}

	return revisionID
}

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

	// actions declared explicitly — replaces the generated set entirely, not
	// just its "apply" key.
	withActionsOverride := fmt.Sprintf(`
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
				Config: testAccStackConfig(wfGrpName, revision, id, ""),
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
				Config:   testAccStackConfig(wfGrpName, revision, id, ""),
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
				Config: testAccStackConfig(wfGrpName, revision, id, `description = "updated after create"`),
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

// TestAccStack_ActionsRevisionRemovedWorkflow verifies that switching
// template_group_id to a revision that dropped a workflow slot the user's own
// actions still references is rejected at plan time
// (validateActionsAgainstRevision), rather than sending a dangling reference
// the API would reject with a less actionable error. revision1
// (setupStackTemplateChainNoActions) wires testWfSlotId and secondWfSlotId;
// revision2 (setupSecondStackTemplateRevision) only re-declares testWfSlotId,
// so secondWfSlotId is the removed workflow.
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
				Config: testAccStackConfig(wfGrpName, revision1, id, referencesSecondSlot),
				Check:  resource.TestCheckResourceAttr("stackguardian_stack.test", "actions.notify.name", "notify"),
			},
			{
				// revision2 dropped secondWfSlotId — actions still references it, so
				// the switch must be rejected.
				Config:      testAccStackConfig(wfGrpName, revision2, id, referencesSecondSlot),
				ExpectError: regexp.MustCompile("actions references a removed workflow"),
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

// actions resolution rules (source material for docs-templates/resources/stack.md.tmpl):
//
//  1. Declaring actions in the resource config always wins — it wholesale
//     replaces whatever the stack template revision itself defines, not a
//     per-key merge (see TestAccStack_ActionsRoundTrip).
//  2. Leaving actions unset resolves it from the stack template revision's
//     own actions verbatim, if the revision has any — both at create time
//     (TestAccStack_TemplateGroupIdReResolution's Step 1, resource_root_test.go)
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
