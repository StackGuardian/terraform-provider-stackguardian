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
// used only by TestAccStack_GeneratedDefaultActions to exercise the
// dependency chaining across more than one workflow.
const secondWfSlotId = "3f7c9e2a-5b1d-4e6f-8a2c-9d4b6e1f0a3c"

// TestAccStack_DefaultActionsComputedOnly verifies default_actions cannot be
// set directly in config (it's Computed-only, reflecting the server's merged
// actions map — custom_actions is the user-authored counterpart). Terraform
// itself rejects this at plan time before any provider code runs.
func TestAccStack_DefaultActionsComputedOnly(t *testing.T) {
	wfGrpName := "tf-provider-stack-defactcomp-wfgrp"
	wfTemplateName := "tf-provider-stack-defactcomp-wftmpl"
	stackTemplateName := "tf-provider-stack-defactcomp-stmpl"
	id := "tf-provider-stack-defactcomp"

	revision := setupStackDependencyChain(t, wfGrpName, wfTemplateName, stackTemplateName, id)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader()),
		Steps: []resource.TestStep{
			{
				Config: testAccStackConfig(wfGrpName, revision, id, `
  default_actions = {
    apply = {
      name = "apply"
    }
  }
`),
				ExpectError: regexp.MustCompile("Invalid Configuration for Read-Only Attribute"),
			},
		},
	})
}

// TestAccStack_ActionsDuplicateKey verifies that an action key present in
// both default_actions and custom_actions produces the "Duplicate action
// key" diagnostic instead of one silently overwriting the other.
//
// default_actions can't be set directly in config (see
// TestAccStack_DefaultActionsComputedOnly), but it becomes a real, known
// value on the plan when template_group_id changes — reResolveOnRevisionChange
// (ModifyPlan) resolves it from the NEW revision's own Actions unconditionally,
// since default_actions has no config counterpart to check. Switching
// template_group_id to a revision whose template defines "apply", while
// custom_actions ALSO declares "apply", collides.
func TestAccStack_ActionsDuplicateKey(t *testing.T) {
	wfGrpName := "tf-provider-stack-actdupe-wfgrp"
	wfTemplateName := "tf-provider-stack-actdupe-wftmpl"
	stackTemplateName := "tf-provider-stack-actdupe-stmpl"
	id := "tf-provider-stack-actdupe"

	revision1 := setupStackDependencyChain(t, wfGrpName, wfTemplateName, stackTemplateName, id)
	revision2 := setupSecondStackTemplateRevision(t, stackTemplateName, wfTemplateName, "revision two", nil)

	collidingCustomActions := fmt.Sprintf(`
  custom_actions = {
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
				Config: testAccStackConfig(wfGrpName, revision1, id, ""),
			},
			{
				Config:      testAccStackConfig(wfGrpName, revision2, id, collidingCustomActions),
				ExpectError: regexp.MustCompile("Duplicate action key"),
			},
		},
	})
}

// TestAccStack_CustomActionsRevisionRemovedWorkflow verifies that switching
// template_group_id to a revision that dropped a workflow slot custom_actions
// still references is rejected at plan time (validateCustomActionsAgainstRevision),
// rather than sending a dangling reference the API would reject with a less
// actionable error. revision1 (setupStackTemplateChainNoActions) wires
// testWfSlotId and secondWfSlotId; revision2 (setupSecondStackTemplateRevision)
// only re-declares testWfSlotId, so secondWfSlotId is the removed workflow.
func TestAccStack_CustomActionsRevisionRemovedWorkflow(t *testing.T) {
	wfGrpName := "tf-provider-stack-actrmwf-wfgrp"
	wfTemplateName := "tf-provider-stack-actrmwf-wftmpl"
	stackTemplateName := "tf-provider-stack-actrmwf-stmpl"
	id := "tf-provider-stack-actrmwf"

	t.Cleanup(func() {
		logCleanupErr(t, fmt.Sprintf("delete workflow group %q", wfGrpName), deleteWorkflowGroupFixture(wfGrpName))
	})
	if err := createWorkflowGroupFixture(wfGrpName); err != nil && !is409(err) {
		t.Fatalf("TestAccStack_CustomActionsRevisionRemovedWorkflow: create workflow group %q: %s", wfGrpName, err)
	}
	workflowTemplateID := setupStackWorkflowTemplate(t, wfTemplateName)
	revision1 := setupStackTemplateChainNoActions(t, stackTemplateName, workflowTemplateID)
	revision2 := setupSecondStackTemplateRevision(t, stackTemplateName, workflowTemplateID, "revision two", nil)
	t.Cleanup(func() { deleteStackFixture(wfGrpName, id) })

	referencesSecondSlot := fmt.Sprintf(`
  custom_actions = {
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
				Check:  resource.TestCheckResourceAttr("stackguardian_stack.test", "custom_actions.notify.name", "notify"),
			},
			{
				// revision2 dropped secondWfSlotId — custom_actions still
				// references it, so the switch must be rejected.
				Config:      testAccStackConfig(wfGrpName, revision2, id, referencesSecondSlot),
				ExpectError: regexp.MustCompile("custom_actions references a removed workflow"),
			},
		},
	})
}

// TestAccStack_CustomActionsRoundTrip covers custom_actions' nested shape —
// order[].parameters.terraform_action.action, order[].parameters.
// environment_variables, and order[].dependencies (id, condition.
// latest_status) — round tripping stably, and that removing an action from
// custom_actions on update actually drops it from the merged payload rather
// than leaving it orphaned. Also confirms an action the template defines but
// custom_actions doesn't override ("plan") lands in default_actions with
// Default=true, per expandActionsMap's template-fallback behavior.
//
// deployment_platform_config and wf_steps_config inside order[].parameters
// aren't covered here — both need a real integration_id / workflow step
// template fixture this test doesn't set up.
func TestAccStack_CustomActionsRoundTrip(t *testing.T) {
	wfGrpName := "tf-provider-stack-actrt-wfgrp"
	wfTemplateName := "tf-provider-stack-actrt-wftmpl"
	stackTemplateName := "tf-provider-stack-actrt-stmpl"
	id := "tf-provider-stack-actrt"

	revision := setupStackDependencyChain(t, wfGrpName, wfTemplateName, stackTemplateName, id)

	withDestroy := fmt.Sprintf(`
  custom_actions = {
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
  custom_actions = {
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
					resource.TestCheckResourceAttr("stackguardian_stack.test", "custom_actions.apply.name", "apply"),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "custom_actions.apply.description", "Custom apply action"),
					resource.TestCheckResourceAttr("stackguardian_stack.test",
						fmt.Sprintf("custom_actions.apply.order.%s.parameters.terraform_action.action", testWfSlotId), "apply"),
					resource.TestCheckResourceAttr("stackguardian_stack.test",
						fmt.Sprintf("custom_actions.apply.order.%s.parameters.environment_variables.0.config.var_name", testWfSlotId), "ACTION_VAR"),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "custom_actions.destroy.name", "destroy"),
					resource.TestCheckResourceAttr("stackguardian_stack.test",
						fmt.Sprintf("custom_actions.destroy.order.%s.dependencies.0.id", testWfSlotId), testWfSlotId),
					resource.TestCheckResourceAttr("stackguardian_stack.test",
						fmt.Sprintf("custom_actions.destroy.order.%s.dependencies.0.condition.latest_status", testWfSlotId), "COMPLETED"),
					// "plan" isn't in custom_actions — inherited from the template as a
					// default action instead.
					resource.TestCheckResourceAttr("stackguardian_stack.test", "default_actions.plan.name", "plan"),
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
				Check:  resource.TestCheckNoResourceAttr("stackguardian_stack.test", "custom_actions.destroy"),
			},
		},
	})
}

// setupStackTemplateChainNoActions creates and publishes a stack template +
// revision :1 via the SDK, like setupStackTemplateChain, but wires TWO
// workflow slots (testWfSlotId, secondWfSlotId — both pointing at the same
// workflow template) instead of one, and defines no Actions of its own at
// all. Used only by TestAccStack_GeneratedDefaultActions, which needs a
// template that supplies neither apply/plan/destroy NOR a dependency chain to
// inherit, so the only source for them is the provider's own generation
// (generateStackActions), and needs a second workflow to prove that
// generation actually chains multiple workflows rather than only handling
// the trivial single-workflow case. Registers cleanup. Returns the bare
// revision id ("<name>:1").
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
			// apply/plan/destroy can only come from the provider's own generation.
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

// TestAccStack_GeneratedDefaultActions verifies generateStackActions against
// default_actions_generation_doc.txt's exact algorithm: when the stack
// template revision defines no Actions of its own, apply/plan/destroy are
// synthesized from the revision's OWN WorkflowsConfig.workflows list — never
// the stack's own workflows_config, which this algorithm doesn't consult at
// all (setupStackTemplateChainNoActions wires two slots on the template,
// testWfSlotId then secondWfSlotId, and the stack in this test declares no
// workflows_config of its own). Order map keys are the bare template slot ids
// (StackTemplateRevisionWorkflow.Id) — not the workflow's own post-creation
// resource id, since at create time that doesn't exist yet.
func TestAccStack_GeneratedDefaultActions(t *testing.T) {
	wfGrpName := "tf-provider-stack-gendef-wfgrp"
	wfTemplateName := "tf-provider-stack-gendef-wftmpl"
	stackTemplateName := "tf-provider-stack-gendef-stmpl"
	id := "tf-provider-stack-gendef"

	t.Cleanup(func() {
		logCleanupErr(t, fmt.Sprintf("delete workflow group %q", wfGrpName), deleteWorkflowGroupFixture(wfGrpName))
	})
	if err := createWorkflowGroupFixture(wfGrpName); err != nil && !is409(err) {
		t.Fatalf("TestAccStack_GeneratedDefaultActions: create workflow group %q: %s", wfGrpName, err)
	}
	workflowTemplateID := setupStackWorkflowTemplate(t, wfTemplateName)
	revision := setupStackTemplateChainNoActions(t, stackTemplateName, workflowTemplateID)
	t.Cleanup(func() { deleteStackFixture(wfGrpName, id) })

	// custom_actions only declares an unrelated "notify" action — apply/plan/
	// destroy come from neither custom_actions nor the template (which has
	// none), so they must be generated. Its own order key is unrelated to
	// generation and keys by the slot id like everything else here.
	withNotifyOnly := fmt.Sprintf(`
  custom_actions = {
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
`, testWfSlotId)

	// Step 2: custom_actions now ALSO declares "apply" itself — proving
	// generation is RE-EVALUATED on Update (ToUpdateAPIModel), not left stuck
	// with the create-time result: "apply" must move out of default_actions
	// (generated) into custom_actions (user-authored), while plan/destroy
	// remain generated. Since generateStackActions's source (the template's
	// WorkflowsConfig.workflows) never changes, this is the only thing that
	// CAN change generation's outcome between create and update here.
	withNotifyAndApply := fmt.Sprintf(`
  custom_actions = {
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
				Config: testAccStackConfig(wfGrpName, revision, id, withNotifyOnly),
				Check: resource.ComposeAggregateTestCheckFunc(
					// custom_actions survives untouched.
					resource.TestCheckResourceAttr("stackguardian_stack.test", "custom_actions.notify.name", "notify"),

					// Generated apply/plan/destroy, matching default_actions.json's shape.
					resource.TestCheckResourceAttr("stackguardian_stack.test", "default_actions.apply.name", "Create"),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "default_actions.apply.description", "use this action to create resources in the stack"),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "default_actions.plan.name", "Plan"),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "default_actions.plan.description", "use this action to plan resources in the stack"),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "default_actions.destroy.name", "Destroy"),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "default_actions.destroy.description", "use this action to destroy resources in the stack"),

					// apply/plan chain in the template's own declaration order:
					// testWfSlotId first (no dependencies), secondWfSlotId depends on it.
					resource.TestCheckResourceAttr("stackguardian_stack.test",
						fmt.Sprintf("default_actions.apply.order.%s.dependencies.#", testWfSlotId), "0"),
					resource.TestCheckResourceAttr("stackguardian_stack.test",
						fmt.Sprintf("default_actions.apply.order.%s.dependencies.0.id", secondWfSlotId), testWfSlotId),
					resource.TestCheckResourceAttr("stackguardian_stack.test",
						fmt.Sprintf("default_actions.apply.order.%s.dependencies.0.condition.latest_status", secondWfSlotId), "COMPLETED"),
					resource.TestCheckResourceAttr("stackguardian_stack.test",
						fmt.Sprintf("default_actions.apply.order.%s.parameters.terraform_action.action", testWfSlotId), "apply"),
					resource.TestCheckResourceAttr("stackguardian_stack.test",
						fmt.Sprintf("default_actions.plan.order.%s.parameters.terraform_action.action", testWfSlotId), "plan"),

					// destroy chains in REVERSE: secondWfSlotId first (no dependencies),
					// testWfSlotId depends on it.
					resource.TestCheckResourceAttr("stackguardian_stack.test",
						fmt.Sprintf("default_actions.destroy.order.%s.dependencies.#", secondWfSlotId), "0"),
					resource.TestCheckResourceAttr("stackguardian_stack.test",
						fmt.Sprintf("default_actions.destroy.order.%s.dependencies.0.id", testWfSlotId), secondWfSlotId),
					resource.TestCheckResourceAttr("stackguardian_stack.test",
						fmt.Sprintf("default_actions.destroy.order.%s.parameters.terraform_action.action", secondWfSlotId), "destroy"),
				),
			},
			{
				// Round trips with no diff.
				Config:   testAccStackConfig(wfGrpName, revision, id, withNotifyOnly),
				PlanOnly: true,
			},
			{
				// Update: custom_actions now covers "apply" itself.
				Config: testAccStackConfig(wfGrpName, revision, id, withNotifyAndApply),
				Check: resource.ComposeAggregateTestCheckFunc(
					// "apply" moved to custom_actions (user-authored, Default=false)...
					resource.TestCheckResourceAttr("stackguardian_stack.test", "custom_actions.apply.name", "apply"),
					// ...and is no longer in default_actions.
					resource.TestCheckNoResourceAttr("stackguardian_stack.test", "default_actions.apply"),
					// plan/destroy are still generated (the template still defines
					// neither, and custom_actions still doesn't cover them).
					resource.TestCheckResourceAttr("stackguardian_stack.test", "default_actions.plan.name", "Plan"),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "default_actions.destroy.name", "Destroy"),
				),
			},
			{
				// And the post-update state round trips with no diff.
				Config:   testAccStackConfig(wfGrpName, revision, id, withNotifyAndApply),
				PlanOnly: true,
			},
		},
	})
}
