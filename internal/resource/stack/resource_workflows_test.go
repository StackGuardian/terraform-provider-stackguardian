package stack_test

import (
	"context"
	"fmt"
	"regexp"
	"strings"
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
`, testWfSlotId)

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
					resource.TestCheckResourceAttr("stackguardian_stack.test", "workflows_config.workflows.0.id", testWfSlotId),
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
`, testWfSlotId, field, value)
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
// referenced stack template revision's own slot list, in both directions of
// mismatch that check guards against: a missing slot, and the same slots
// declared out of order. See that function's doc comment for why both are
// enforced (completeness, so a template's slot never silently goes
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
	// setupStackTemplateChainNoActions wires two slots — testWfSlotId then
	// secondWfSlotId, in that order — so both directions of mismatch can be
	// exercised against a single fixture.
	revision := setupStackTemplateChainNoActions(t, stackTemplateName, workflowTemplateID)
	t.Cleanup(func() { deleteStackFixture(wfGrpName, id) })

	missingSlot := fmt.Sprintf(`
  workflows_config = {
    workflows = [
      { id = %q }
    ]
  }
`, testWfSlotId)

	wrongOrder := fmt.Sprintf(`
  workflows_config = {
    workflows = [
      { id = %q },
      { id = %q }
    ]
  }
`, secondWfSlotId, testWfSlotId)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader()),
		Steps: []resource.TestStep{
			{
				// The revision defines testWfSlotId AND secondWfSlotId — declaring
				// only one is a missing slot, not a valid partial subset.
				Config:      testAccStackConfig(wfGrpName, revision, id, missingSlot),
				ExpectError: regexp.MustCompile("does not match the stack template revision"),
			},
			{
				// The same two slots, declared in the opposite order from the
				// revision's own declaration order.
				Config:      testAccStackConfig(wfGrpName, revision, id, wrongOrder),
				ExpectError: regexp.MustCompile("does not match the stack template revision"),
			},
		},
	})
}

// TestAccStack_WorkflowsConfigPrecedenceMerge exercises the three-way
// precedence merge for a workflow slot's terraform_config: workflow template
// revision default (lowest, terraform_version 1.5.0 — see
// setupStackWorkflowTemplate) < stack template revision's override for the
// slot (middle, terraform_version 1.5.7 — see setupStackTemplateChain) < the
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
`, testWfSlotId)

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
`, testWfSlotId)

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
// workflow slot (see resolveWorkflowTemplates/mergeWorkflowWithStackTemplateOverride).
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
`, testWfSlotId)

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
`, testWfSlotId)

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
`, testWfSlotId)

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
`, testWfSlotId)

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
`, testWfSlotId)

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

// TestAccStack_WorkflowsConfigDriftRestoredAfterOutOfBandDelete verifies that deleting the
// live workflow behind a workflows_config.workflows[] entry directly (not the stack itself)
// is detected as drift on the next refresh, and that re-applying the SAME, unchanged config
// restores it — the platform recreates the missing workflow to match the declared
// workflows_config, rather than the provider erroring or silently leaving it gone.
//
// workflow_id (the newly-added computed field, model.go's computeWorkflowId) is what makes
// this test possible without any extra lookup: it's derived purely from the resolved
// iac_template_id and the workflow's own declared uuid (workflows_config.workflows[].id),
// so this test computes the exact same value independently and uses it to delete the live
// workflow directly via client.Workflows.DeleteWorkflow — the stack resource itself is
// untouched by that call.
func TestAccStack_WorkflowsConfigDriftRestoredAfterOutOfBandDelete(t *testing.T) {
	wfGrpName := "tf-provider-stack-wfdrift-wfgrp"
	wfTemplateName := "tf-provider-stack-wfdrift-wftmpl"
	stackTemplateName := "tf-provider-stack-wfdrift-stmpl"
	id := "tf-provider-stack-wfdrift"

	revision := setupStackDependencyChain(t, wfGrpName, wfTemplateName, stackTemplateName, id)

	// Mirrors computeWorkflowId's formula (model.go): "<template-name>-<first octet of the
	// declared workflow uuid>". wfTemplateName is already the bare template name
	// (setupStackWorkflowTemplate's own id), so no org/revision stripping is needed here.
	expectedWorkflowId := fmt.Sprintf("%s-%s", wfTemplateName, strings.SplitN(testWfSlotId, "-", 2)[0])

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
				Config: testAccStackConfig(wfGrpName, revision, id, config),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_stack.test", "workflows_config.workflows.0.workflow_id", expectedWorkflowId),
					resource.TestCheckResourceAttrSet("stackguardian_stack.test", "workflows_config.workflows.0.resource_name"),
				),
			},
			{
				// Delete the live workflow directly — the stack resource itself is never
				// touched by this call, so Terraform's own state has no idea this happened
				// until the refresh below re-reads the stack.
				PreConfig: func() {
					if _, err := getClient().Workflows.DeleteWorkflow(context.TODO(), org, expectedWorkflowId, wfGrpName); err != nil {
						t.Fatalf("failed to delete workflow %q out-of-band: %s", expectedWorkflowId, err)
					}
				},
				RefreshState:       true,
				ExpectNonEmptyPlan: true,
			},
			{
				// Same config, a real apply — the provider must restore the missing
				// workflow to match workflows_config, not error or leave it gone.
				Config: testAccStackConfig(wfGrpName, revision, id, config),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_stack.test", "workflows_config.workflows.#", "1"),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "workflows_config.workflows.0.id", testWfSlotId),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "workflows_config.workflows.0.workflow_id", expectedWorkflowId),
					resource.TestCheckResourceAttrSet("stackguardian_stack.test", "workflows_config.workflows.0.resource_name"),
				),
			},
			{
				// And the restored state round trips with no diff.
				Config:   testAccStackConfig(wfGrpName, revision, id, config),
				PlanOnly: true,
			},
		},
	})
}
