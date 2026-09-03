package stack_test

import (
	"fmt"
	"regexp"
	"testing"

	sgsdkgo "github.com/StackGuardian/sg-sdk-go"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

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

// TestAccStack_WorkflowsConfigRevisionReResolution — REVISION SWITCH test.
// Purpose: a per-workflow field the stack leaves unset (number_of_approvals_required) must
// pick up a value the new revision provides that the old one didn't (revision1 -> revision2,
// step 2), and must go back to empty when a later revision drops it again (revision2 ->
// revision1, step 3) — rather than the value getting stuck at whatever UseStateForUnknown last
// carried forward (see reResolveWorkflowsConfigOnRevisionChange).
func TestAccStack_WorkflowsConfigRevisionReResolution(t *testing.T) {
	wfGrpName := "tf-provider-stack-wfrevre-wfgrp"
	wfTemplateName := "tf-provider-stack-wfrevre-wftmpl"
	stackTemplateName := "tf-provider-stack-wfrevre-stmpl"
	id := "tf-provider-stack-wfrevre"

	revision1 := setupStackDependencyChain(t, wfGrpName, wfTemplateName, stackTemplateName, id)
	revision2 := setupSecondStackTemplateRevision(t, stackTemplateName, wfTemplateName, "revision two", sgsdkgo.Int(2))

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
				// revision1 never sets number_of_approvals_required.
				Config: testAccStackConfig(wfGrpName, revision1, id, config),
				Check:  resource.TestCheckResourceAttr("stackguardian_stack.test", "workflows_config.workflows.0.number_of_approvals_required", "0"),
			},
			{
				// revision2 sets it to 2 — must now appear, though nothing in
				// this stack's own config changed.
				Config: testAccStackConfig(wfGrpName, revision2, id, config),
				Check:  resource.TestCheckResourceAttr("stackguardian_stack.test", "workflows_config.workflows.0.number_of_approvals_required", "2"),
			},
			{
				// Back to revision1 — must clear again, not stay stuck at 2.
				Config: testAccStackConfig(wfGrpName, revision1, id, config),
				Check:  resource.TestCheckResourceAttr("stackguardian_stack.test", "workflows_config.workflows.0.number_of_approvals_required", "0"),
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

// TestAccStack_WorkflowsConfigAddWorkflowSlot verifies that adding a second
// entry to workflows_config.workflows[] on an already-existing stack applies
// correctly — the existing slot's data must be undisturbed, and the new slot
// must appear with its own values — then that removing it again shrinks the
// list back down cleanly. setupStackTemplateChainNoActions wires two workflow
// slots (testWfSlotId, secondWfSlotId) on the template so both are available
// to declare; this is the one workflows_config scenario no other test covers:
// the workflows list itself changing shape via update, not just values within
// an already-declared slot.
func TestAccStack_WorkflowsConfigAddWorkflowSlot(t *testing.T) {
	wfGrpName := "tf-provider-stack-wfadd-wfgrp"
	wfTemplateName := "tf-provider-stack-wfadd-wftmpl"
	stackTemplateName := "tf-provider-stack-wfadd-stmpl"
	id := "tf-provider-stack-wfadd"

	t.Cleanup(func() {
		logCleanupErr(t, fmt.Sprintf("delete workflow group %q", wfGrpName), deleteWorkflowGroupFixture(wfGrpName))
	})
	if err := createWorkflowGroupFixture(wfGrpName); err != nil && !is409(err) {
		t.Fatalf("TestAccStack_WorkflowsConfigAddWorkflowSlot: create workflow group %q: %s", wfGrpName, err)
	}
	workflowTemplateID := setupStackWorkflowTemplate(t, wfTemplateName)
	revision := setupStackTemplateChainNoActions(t, stackTemplateName, workflowTemplateID)
	t.Cleanup(func() { deleteStackFixture(wfGrpName, id) })

	oneSlot := fmt.Sprintf(`
  workflows_config = {
    workflows = [
      { id = %q, tags = ["first"] }
    ]
  }
`, testWfSlotId)

	twoSlots := fmt.Sprintf(`
  workflows_config = {
    workflows = [
      { id = %q, tags = ["first"] },
      { id = %q, tags = ["second"] }
    ]
  }
`, testWfSlotId, secondWfSlotId)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader()),
		Steps: []resource.TestStep{
			{
				Config: testAccStackConfig(wfGrpName, revision, id, oneSlot),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_stack.test", "workflows_config.workflows.#", "1"),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "workflows_config.workflows.0.id", testWfSlotId),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "workflows_config.workflows.0.tags.0", "first"),
				),
			},
			{
				// Add a second slot — the first slot's data must be undisturbed,
				// and the new slot must appear with its own values.
				Config: testAccStackConfig(wfGrpName, revision, id, twoSlots),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_stack.test", "workflows_config.workflows.#", "2"),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "workflows_config.workflows.0.tags.0", "first"),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "workflows_config.workflows.1.id", secondWfSlotId),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "workflows_config.workflows.1.tags.0", "second"),
				),
			},
			{
				// Remove the second slot again — must shrink back to one cleanly.
				Config: testAccStackConfig(wfGrpName, revision, id, oneSlot),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_stack.test", "workflows_config.workflows.#", "1"),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "workflows_config.workflows.0.id", testWfSlotId),
				),
			},
		},
	})
}
