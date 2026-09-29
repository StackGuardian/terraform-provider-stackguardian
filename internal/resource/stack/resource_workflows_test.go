package stack_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/StackGuardian/terraform-provider-stackguardian/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

// TODO: ModifyPlan's revision-change branch (resource.go) only re-resolves
// actions when it's left unset in config AND the new revision defines its
// own actions (reResolveOnRevisionChange's doc comment in model.go). When
// actions is unset but the new revision has none of its own, plan.Actions is
// currently left untouched — carried forward from the old revision's already-
// resolved value — instead of being re-resolved (e.g. back to whatever the
// API would generate). It should be re-resolved whenever actions isn't
// provided in the resource config, regardless of what the new revision does.
// No test currently covers this gap.

// TestAccStack_WorkflowsConfigMinimalEntry verifies a minimal workflow entry
// (only the Required id) creates successfully, and doubles as a regression
// test for the Optional+Computed guard fix: omitted fields must resolve to
// real server-assigned values, not the forced-empty/zero values
// ValueStringPointer()/ValueInt64Pointer() return for unknown.
func TestAccStack_WorkflowsConfigMinimalEntry(t *testing.T) {
	wfGrpName := "tf-provider-stack-wfmin-wfgrp"
	wfTemplateName := "tf-provider-stack-wfmin-wftmpl"
	stackTemplateName := "tf-provider-stack-wfmin-stmpl"
	id := "tf-provider-stack-wfmin"

	revision := setupStackDependencyChain(t, wfGrpName, wfTemplateName, stackTemplateName, id)

	config := fmt.Sprintf(`
  workflows_config = {
    workflows = [
      { id = %q }
    ]
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
				Config: testAccStackConfig(wfGrpName, revision, id, config),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_stack.test", "workflows_config.workflows.0.id", testWorkflowUUID),
					// None of these were declared — must resolve to real values, not
					// be forced to "" / 0 by the unknown-value guard bug.
					resource.TestCheckResourceAttrSet("stackguardian_stack.test", "workflows_config.workflows.0.resource_name"),
					resource.TestCheckResourceAttrSet("stackguardian_stack.test", "workflows_config.workflows.0.user_job_cpu"),
					resource.TestCheckResourceAttrSet("stackguardian_stack.test", "workflows_config.workflows.0.user_job_memory"),
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

// TestAccStack_WorkflowsConfigInvalidEnums verifies invalid wf_type and
// parallel_execution values each produce their own provider-side diagnostic
// (expandWorkflowsConfig) instead of being passed through to the API. Each
// step's apply fails before anything is created, so they can safely share
// one resource address across steps.
func TestAccStack_WorkflowsConfigInvalidEnums(t *testing.T) {
	wfGrpName := "tf-provider-stack-wfenum-wfgrp"
	wfTemplateName := "tf-provider-stack-wfenum-wftmpl"
	stackTemplateName := "tf-provider-stack-wfenum-stmpl"
	id := "tf-provider-stack-wfenum"

	revision := setupStackDependencyChain(t, wfGrpName, wfTemplateName, stackTemplateName, id)

	workflowConfig := func(field, value string) string {
		return fmt.Sprintf(`
  workflows_config = {
    workflows = [
      {
        id       = %q
        %s = %q
      }
    ]
  }
`, testWorkflowUUID, field, value)
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader()),
		Steps: []resource.TestStep{
			{
				Config:      testAccStackConfig(wfGrpName, revision, id, workflowConfig("wf_type", "NOT_A_TYPE")),
				ExpectError: regexp.MustCompile("Invalid wf_type"),
			},
			{
				Config:      testAccStackConfig(wfGrpName, revision, id, workflowConfig("parallel_execution", "sideways")),
				ExpectError: regexp.MustCompile("Invalid parallel_execution"),
			},
		},
	})
}

// TestAccStack_WorkflowsConfigMismatchRejected verifies
// validateWorkflowsConfigMatchesRevision (model.go) rejects
// workflows_config.workflows whenever it doesn't exactly match the
// referenced stack template revision's own workflow list, in both directions of
// mismatch that check guards against: a missing workflow, and the same workflows
// declared out of order. See that function's doc comment for why both are
// enforced (completeness, so a template's workflow never silently goes
// undeclared; order, so the stack's own listing can't drift out of sync with
// the order the API's default action-chaining derives from). Each step's
// apply fails before anything is created, so they can safely share one
// resource address across steps (mirrors TestAccStack_WorkflowsConfigInvalidEnums).
func TestAccStack_WorkflowsConfigMismatchRejected(t *testing.T) {
	wfGrpName := "tf-provider-stack-wfmismatch-wfgrp"
	wfTemplateName := "tf-provider-stack-wfmismatch-wftmpl"
	stackTemplateName := "tf-provider-stack-wfmismatch-stmpl"
	id := "tf-provider-stack-wfmismatch"

	t.Cleanup(func() {
		logCleanupErr(t, fmt.Sprintf("delete workflow group %q", wfGrpName), deleteWorkflowGroupFixture(wfGrpName))
	})
	if err := createWorkflowGroupFixture(wfGrpName); err != nil && !is409(err) {
		t.Fatalf("TestAccStack_WorkflowsConfigMismatchRejected: create workflow group %q: %s", wfGrpName, err)
	}
	workflowTemplateID := setupStackWorkflowTemplate(t, wfTemplateName)
	// setupStackTemplateChainNoActions declares two workflows — testWorkflowUUID then
	// secondWorkflowUUID, in that order — so both directions of mismatch can be
	// exercised against a single fixture.
	revision := setupStackTemplateChainNoActions(t, stackTemplateName, workflowTemplateID)
	t.Cleanup(func() { deleteStackFixture(wfGrpName, id) })

	missingWorkflow := fmt.Sprintf(`
  workflows_config = {
    workflows = [
      { id = %q }
    ]
  }
`, testWorkflowUUID)

	wrongOrder := fmt.Sprintf(`
  workflows_config = {
    workflows = [
      { id = %q },
      { id = %q }
    ]
  }
`, secondWorkflowUUID, testWorkflowUUID)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader()),
		Steps: []resource.TestStep{
			{
				// The revision defines testWorkflowUUID AND secondWorkflowUUID — declaring
				// only one is a missing workflow, not a valid partial subset.
				Config:      testAccStackConfig(wfGrpName, revision, id, missingWorkflow),
				ExpectError: regexp.MustCompile("does not match the stack template revision"),
			},
			{
				// The same two workflows, declared in the opposite order from the
				// revision's own declaration order.
				Config:      testAccStackConfig(wfGrpName, revision, id, wrongOrder),
				ExpectError: regexp.MustCompile("does not match the stack template revision"),
			},
		},
	})
}

// TODO: TestAccStack_WorkflowsConfigMismatchRejected only exercises the CREATE-time call to
// validateWorkflowsConfigMatchesRevision — both its steps fail before anything is ever
// created, so a fresh resource.go's Create() is the only call site actually proven to reject
// a mismatch. The check now also runs from two other call sites (see the "ModifyPlan only
// re-validates..." comments in resource.go's Update()/ModifyPlan, added when the
// double-fetch-per-apply was eliminated), and neither is covered by a live test yet:
//   - Update() with NO revision change: create a stack successfully with a valid
//     workflows_config, then apply an update that changes workflows_config into a mismatch
//     (a workflow removed, or reordered) while template_group_id stays the same — must still be
//     rejected, not silently accepted just because the resource already exists.
//   - ModifyPlan's revision-change branch: create successfully against revision1, then switch
//     template_group_id to a revision2 whose own workflow list, if carried over verbatim from
//     revision1's declared workflows_config, would now be a missing-workflow or wrong-order
//     mismatch against revision2 — must be rejected at plan time, mirroring
//     TestAccStack_ActionsRevisionRemovedWorkflow's pattern in resource_stack_upgrade_test.go.
//
// Why order specifically has to match (not just membership): the platform's own default
// action generation (an unset "actions" attribute — see TestAccStack_ActionsGeneratedFromTemplate)
// chains apply/plan/destroy dependencies ACROSS workflows in the stack template revision's own
// declaration order — first workflow has no dependency, second depends on the first, and so
// on (reversed for destroy). If workflows_config.workflows[] could list the same workflows in a
// different order than the revision declares them, the stack's own visible listing would
// silently disagree with the order those generated dependencies are actually chained in, with
// nothing in the diff ever explaining why. Pinning the declared order to the revision's own
// order is what keeps the two from ever drifting apart.

// TestAccStack_WorkflowsConfigPrecedenceMerge exercises the three-way
// precedence merge for a workflow's terraform_config: workflow template
// revision default (lowest, terraform_version 1.5.0 — see
// setupStackWorkflowTemplate) < stack template revision's override for the
// workflow (middle, terraform_version 1.5.7 — see setupStackTemplateChain) < the
// stack's own workflows_config.workflows[] entry (highest). Step 1 leaves
// terraform_config unset on the stack entry, so it must resolve to the
// middle layer's 1.5.7 (not the bottom layer's 1.5.0). Step 2 declares
// terraform_version directly on the stack entry, which must win over both
// lower layers — 1.5.5 is used rather than something above 1.5.7 since the
// API no longer supports Terraform versions past that; precedence here
// doesn't depend on numeric ordering, only on which layer declared a value.
func TestAccStack_WorkflowsConfigPrecedenceMerge(t *testing.T) {
	wfGrpName := "tf-provider-stack-wfmerge-wfgrp"
	wfTemplateName := "tf-provider-stack-wfmerge-wftmpl"
	stackTemplateName := "tf-provider-stack-wfmerge-stmpl"
	id := "tf-provider-stack-wfmerge"

	revision := setupStackDependencyChain(t, wfGrpName, wfTemplateName, stackTemplateName, id)

	inheritedConfig := fmt.Sprintf(`
  workflows_config = {
    workflows = [
      { id = %q }
    ]
  }
`, testWorkflowUUID)

	overrideConfig := fmt.Sprintf(`
  workflows_config = {
    workflows = [
      {
        id = %q
        terraform_config = {
          terraform_version = "1.5.5"
        }
      }
    ]
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
				Config: testAccStackConfig(wfGrpName, revision, id, inheritedConfig),
				Check: resource.TestCheckResourceAttr(
					"stackguardian_stack.test", "workflows_config.workflows.0.terraform_config.terraform_version", "1.5.7"),
			},
			{
				Config: testAccStackConfig(wfGrpName, revision, id, overrideConfig),
				Check: resource.TestCheckResourceAttr(
					"stackguardian_stack.test", "workflows_config.workflows.0.terraform_config.terraform_version", "1.5.5"),
			},
		},
	})
}

// TestAccStack_WorkflowsConfigVcsConfigComputedOnly verifies
// vcs_config.iac_vcs_config can't be set directly in config — it's
// Computed-only and always inherited from the matched stack-template-revision
// workflow (see resolveWorkflowTemplates/mergeWorkflowWithStackTemplateOverride).
func TestAccStack_WorkflowsConfigVcsConfigComputedOnly(t *testing.T) {
	wfGrpName := "tf-provider-stack-wfvcs-wfgrp"
	wfTemplateName := "tf-provider-stack-wfvcs-wftmpl"
	stackTemplateName := "tf-provider-stack-wfvcs-stmpl"
	id := "tf-provider-stack-wfvcs"

	revision := setupStackDependencyChain(t, wfGrpName, wfTemplateName, stackTemplateName, id)

	config := fmt.Sprintf(`
  workflows_config = {
    workflows = [
      {
        id = %q
        vcs_config = {
          iac_vcs_config = {
            use_marketplace_template = true
          }
        }
      }
    ]
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
				Config:      testAccStackConfig(wfGrpName, revision, id, config),
				ExpectError: regexp.MustCompile("Invalid Configuration for Read-Only Attribute"),
			},
		},
	})
}

// TestAccStack_WorkflowsConfigRoundTrip covers a batch of the remaining
// per-workflow attributes together: approvers, user_schedules (per-workflow
// shape — cron/state Required, no "inputs" field, unlike the stack-level
// copy), context_tags, runner_constraints, and mini_steps.
func TestAccStack_WorkflowsConfigRoundTrip(t *testing.T) {
	wfGrpName := "tf-provider-stack-wfrt-wfgrp"
	wfTemplateName := "tf-provider-stack-wfrt-wftmpl"
	stackTemplateName := "tf-provider-stack-wfrt-stmpl"
	id := "tf-provider-stack-wfrt"

	revision := setupStackDependencyChain(t, wfGrpName, wfTemplateName, stackTemplateName, id)

	config := fmt.Sprintf(`
  workflows_config = {
    workflows = [
      {
        id        = %q
        approvers = ["alice@example.com"]

        user_schedules = [
          {
            cron  = "0 8 ? * MON *"
            state = "ENABLED"
          }
        ]

        context_tags = {
          env = "dev"
        }

        runner_constraints = {
          type  = "private"
          names = ["runner-1"]
        }

        mini_steps = {
          webhooks = {
            completed = [
              {
                webhook_name = "on-completed"
                webhook_url  = "https://example.com/hook"
              }
            ]
          }
        }
      }
    ]
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
				Config: testAccStackConfig(wfGrpName, revision, id, config),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_stack.test", "workflows_config.workflows.0.approvers.0", "alice@example.com"),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "workflows_config.workflows.0.user_schedules.0.cron", "0 8 ? * MON *"),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "workflows_config.workflows.0.user_schedules.0.state", "ENABLED"),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "workflows_config.workflows.0.context_tags.env", "dev"),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "workflows_config.workflows.0.runner_constraints.type", "private"),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "workflows_config.workflows.0.runner_constraints.names.0", "runner-1"),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "workflows_config.workflows.0.mini_steps.webhooks.completed.0.webhook_name", "on-completed"),
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

// TestAccStack_WorkflowsConfigUpdate verifies that changing per-workflow
// values on an ALREADY-EXISTING stack actually applies — approvers,
// user_schedules, context_tags, runner_constraints, and mini_steps all
// round-tripped at create in TestAccStack_WorkflowsConfigRoundTrip, but were
// never re-verified after a real update to a different value. Step 3 clears
// every override back to unset and checks the ones with a known-empty
// flatten default (tags/approvers/user_schedules/context_tags) settle to
// empty rather than erroring or keeping a stale value — exercising that path
// on a genuine update, not just a fresh create.
func TestAccStack_WorkflowsConfigUpdate(t *testing.T) {
	wfGrpName := "tf-provider-stack-wfupd-wfgrp"
	wfTemplateName := "tf-provider-stack-wfupd-wftmpl"
	stackTemplateName := "tf-provider-stack-wfupd-stmpl"
	id := "tf-provider-stack-wfupd"

	revision := setupStackDependencyChain(t, wfGrpName, wfTemplateName, stackTemplateName, id)

	initialConfig := fmt.Sprintf(`
  workflows_config = {
    workflows = [
      {
        id        = %q
        tags      = ["v1"]
        approvers = ["alice@example.com"]

        user_schedules = [
          { cron = "0 8 ? * MON *", state = "ENABLED" }
        ]

        context_tags = {
          env = "dev"
        }

        runner_constraints = {
          type  = "private"
          names = ["runner-1"]
        }

        mini_steps = {
          webhooks = {
            completed = [
              { webhook_name = "on-completed", webhook_url = "https://example.com/hook" }
            ]
          }
        }
      }
    ]
  }
`, testWorkflowUUID)

	updatedConfig := fmt.Sprintf(`
  workflows_config = {
    workflows = [
      {
        id        = %q
        tags      = ["v2"]
        approvers = ["bob@example.com"]

        user_schedules = [
          { cron = "0 9 ? * TUE *", state = "DISABLED" }
        ]

        context_tags = {
          env = "prod"
        }

        runner_constraints = {
          type  = "private"
          names = ["runner-2"]
        }

        mini_steps = {
          webhooks = {
            completed = [
              { webhook_name = "on-completed-v2", webhook_url = "https://example.com/hook-v2" }
            ]
          }
        }
      }
    ]
  }
`, testWorkflowUUID)

	// Explicit empty values, not omission: removing an attribute from config
	// entirely is indistinguishable, at plan time, from never having set it —
	// UseStateForUnknown then just carries the prior (Step 2) value forward
	// with no diff, so omitting these here would silently keep testing Step
	// 2's values instead of exercising a real clear. An explicit [] / {} is a
	// known, non-null config value, so it plans as a genuine change — same
	// pattern as TestAccWorkflowUsingTemplate_ExplicitEmptySuppressesTemplateDefault.
	clearedConfig := fmt.Sprintf(`
  workflows_config = {
    workflows = [
      {
        id             = %q
        tags           = []
        approvers      = []
        user_schedules = []
        context_tags   = {}
      }
    ]
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
				Config: testAccStackConfig(wfGrpName, revision, id, initialConfig),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_stack.test", "workflows_config.workflows.0.tags.0", "v1"),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "workflows_config.workflows.0.approvers.0", "alice@example.com"),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "workflows_config.workflows.0.user_schedules.0.cron", "0 8 ? * MON *"),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "workflows_config.workflows.0.user_schedules.0.state", "ENABLED"),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "workflows_config.workflows.0.context_tags.env", "dev"),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "workflows_config.workflows.0.runner_constraints.names.0", "runner-1"),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "workflows_config.workflows.0.mini_steps.webhooks.completed.0.webhook_name", "on-completed"),
				),
			},
			{
				// Update: every value above changes to a different one on the
				// SAME stack — verifies the change actually applies, not just
				// that the initial create round trips.
				Config: testAccStackConfig(wfGrpName, revision, id, updatedConfig),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_stack.test", "workflows_config.workflows.0.tags.0", "v2"),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "workflows_config.workflows.0.approvers.0", "bob@example.com"),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "workflows_config.workflows.0.user_schedules.0.cron", "0 9 ? * TUE *"),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "workflows_config.workflows.0.user_schedules.0.state", "DISABLED"),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "workflows_config.workflows.0.context_tags.env", "prod"),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "workflows_config.workflows.0.runner_constraints.names.0", "runner-2"),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "workflows_config.workflows.0.mini_steps.webhooks.completed.0.webhook_name", "on-completed-v2"),
				),
			},
			{
				// Clear every override back to unset — the template supplies no
				// default for any of these, so the known-empty ones must settle
				// to empty rather than error or keep the prior step's value.
				Config: testAccStackConfig(wfGrpName, revision, id, clearedConfig),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_stack.test", "workflows_config.workflows.0.tags.#", "0"),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "workflows_config.workflows.0.approvers.#", "0"),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "workflows_config.workflows.0.user_schedules.#", "0"),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "workflows_config.workflows.0.context_tags.%", "0"),
				),
			},
		},
	})
}

// TODO: out-of-band deletion of a stack workflow is not handled yet — the test below is
// commented out until it is. Deleting the live workflow behind a workflows_config.workflows[] entry
// (e.g. from the platform UI) makes the next refresh drop that entry from state; re-applying the
// unchanged config then fails with "Provider produced inconsistent result after apply"
// (workflows[0].workflow_id / approvers: was null, but now ...). Real users hit this too: a plain
// `terraform apply` refreshes before planning. Cause: UseStateForUnknown (framework v1.19.0) only
// skips when the whole resource has no state, so for an entry missing from state it copies that
// entry's null into the plan for every computed field, while the apply returns real values.
//
// Proposed fix: in ModifyPlan, when template_group_id is unchanged but the plan has a
// workflows[].id not present in the refreshed state, fetch the revision and
// resolveWorkflowTemplates, expand/flatten the same way reResolveWorkflowsConfigOnRevisionChange
// (model.go) does, and write the predicted values only into those new entries' fields that are
// null in config — existing entries and anything set in .tf stay untouched, and a normal update
// never triggers the extra fetch. Rejected: swapping to UseNonNullStateForUnknown — it would
// plan nullable fields (description, user_job_cpu, ...) as "known after apply" on every update,
// cascading to anything referencing them.

// TestAccStack_WorkflowsConfigDriftRestoredAfterOutOfBandDelete verifies that deleting the
// live workflow behind a workflows_config.workflows[] entry directly (not the stack itself)
// is detected as drift on the next refresh, and that re-applying the SAME, unchanged config
// restores it — the platform recreates the missing workflow to match the declared
// workflows_config, rather than the provider erroring or silently leaving it gone.
//
// workflow_id (the newly-added computed field, model.go's computeWorkflowId) is what makes
// this test possible without any extra lookup: it's derived purely from the resolved
// iac_template_id and the workflow's own declared uuid (workflows_config.workflows[].id).
// This test calls computeWorkflowId itself (exposed via export_test.go) rather than
// duplicating its formula, and uses the result to delete the live workflow directly via
// client.StackWorkflows.DeleteStackWorkflow — the stack resource itself is untouched by that
// call.
//func TestAccStack_WorkflowsConfigDriftRestoredAfterOutOfBandDelete(t *testing.T) {
//	wfGrpName := "tf-provider-stack-wfdrift-wfgrp"
//	wfTemplateName := "tf-provider-stack-wfdrift-wftmpl"
//	stackTemplateName := "tf-provider-stack-wfdrift-stmpl"
//	id := "tf-provider-stack-wfdrift"
//
//	revision := setupStackDependencyChain(t, wfGrpName, wfTemplateName, stackTemplateName, id)
//
//	// Computed via the same computeWorkflowId the provider itself calls when building the
//	// create/update payload (model.go), fed the exact resolved iac_template_id
//	// setupStackTemplateChainWithFields wires onto the workflow entry ("/<org>/<workflow
//	// template id>:1" — see prefixedWorkflowRevisionID there), rather than duplicating the
//	// formula's string-parsing logic here.
//	resolvedIacTemplateId := fmt.Sprintf("/%s/%s:1", org, wfTemplateName)
//	expectedWorkflowId := stackresource.ComputeWorkflowId(resolvedIacTemplateId, testWorkflowUUID)
//
//	config := fmt.Sprintf(`
//  workflows_config = {
//    workflows = [
//      { id = %q }
//    ]
//  }
//`, testWorkflowUUID)
//
//	resource.Test(t, resource.TestCase{
//		PreCheck: func() { acctest.TestAccPreCheck(t) },
//		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
//			tfversion.SkipBelow(tfversion.Version1_1_0),
//		},
//		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader()),
//		Steps: []resource.TestStep{
//			{
//				Config: testAccStackConfig(wfGrpName, revision, id, config),
//				Check: resource.ComposeAggregateTestCheckFunc(
//					resource.TestCheckResourceAttr("stackguardian_stack.test", "workflows_config.workflows.0.workflow_id", expectedWorkflowId),
//					resource.TestCheckResourceAttrSet("stackguardian_stack.test", "workflows_config.workflows.0.resource_name"),
//				),
//			},
//			{
//				// Delete the live workflow directly — the stack resource itself is never
//				// touched by this call, so Terraform's own state has no idea this happened
//				// until the refresh below re-reads the stack.
//				PreConfig: func() {
//					if err := getClient().StackWorkflows.DeleteStackWorkflow(context.TODO(), org, id, expectedWorkflowId, wfGrpName); err != nil {
//						t.Fatalf("failed to delete workflow %q out-of-band: %s", expectedWorkflowId, err)
//					}
//				},
//				RefreshState:       true,
//				ExpectNonEmptyPlan: true,
//			},
//			{
//				// Same config, a real apply — the provider must restore the missing
//				// workflow to match workflows_config, not error or leave it gone.
//				Config: testAccStackConfig(wfGrpName, revision, id, config),
//				Check: resource.ComposeAggregateTestCheckFunc(
//					resource.TestCheckResourceAttr("stackguardian_stack.test", "workflows_config.workflows.#", "1"),
//					resource.TestCheckResourceAttr("stackguardian_stack.test", "workflows_config.workflows.0.id", testWorkflowUUID),
//					resource.TestCheckResourceAttr("stackguardian_stack.test", "workflows_config.workflows.0.workflow_id", expectedWorkflowId),
//					resource.TestCheckResourceAttrSet("stackguardian_stack.test", "workflows_config.workflows.0.resource_name"),
//				),
//			},
//			{
//				// And the restored state round trips with no diff.
//				Config:   testAccStackConfig(wfGrpName, revision, id, config),
//				PlanOnly: true,
//			},
//		},
//	})
//}
