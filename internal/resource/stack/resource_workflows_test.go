package stack_test

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"testing"

	sgsdkgo "github.com/StackGuardian/sg-sdk-go"
	"github.com/StackGuardian/sg-sdk-go/stacktemplaterevisions"
	"github.com/StackGuardian/sg-sdk-go/workflowtemplaterevisions"
	"github.com/StackGuardian/sg-sdk-go/workflowtemplates"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
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
	acctest.SkipUnlessAcceptance(t)
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
	acctest.SkipUnlessAcceptance(t)
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
	acctest.SkipUnlessAcceptance(t)
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

// TestAccStack_WorkflowsConfigMismatchRejectedAtPlan verifies a workflows_config
// that doesn't match the stack template revision (a missing workflow, or the
// right workflows in the wrong order) is rejected by `terraform plan`, not only
// when applying, both when creating a stack and when updating one without
// changing template_group_id. Rejecting it at plan time matters for pipelines
// that review a plan before applying it: a plan that succeeds would promise a
// change the apply then refuses. Every rejected step is PlanOnly, so if the
// check only ran at apply these steps would fail with "the plan was not empty"
// instead of the expected error.
//
// The two-workflow revision (setupStackTemplateChainNoActions) declares
// testWorkflowUUID then secondWorkflowUUID, so both kinds of mismatch can be
// exercised against one fixture.
func TestAccStack_WorkflowsConfigMismatchRejectedAtPlan(t *testing.T) {
	acctest.SkipUnlessAcceptance(t)
	wfGrpName := acctest.ResourceName("stack-wfmismatchplan-wfgrp")
	wfTemplateName := acctest.ResourceName("stack-wfmismatchplan-wftmpl")
	stackTemplateName := acctest.ResourceName("stack-wfmismatchplan-stmpl")
	id := acctest.ResourceName("stack-wfmismatchplan")

	t.Cleanup(func() {
		logCleanupErr(t, fmt.Sprintf("delete workflow group %q", wfGrpName), deleteWorkflowGroupFixture(wfGrpName))
	})
	if err := createWorkflowGroupFixture(wfGrpName); err != nil && !is409(err) {
		t.Fatalf("TestAccStack_WorkflowsConfigMismatchRejectedAtPlan: create workflow group %q: %s", wfGrpName, err)
	}
	workflowTemplateID := setupStackWorkflowTemplate(t, wfTemplateName)
	revision := setupStackTemplateChainNoActions(t, stackTemplateName, workflowTemplateID)
	t.Cleanup(func() { deleteStackFixture(wfGrpName, id) })

	workflows := func(ids ...string) string {
		entries := ""
		for _, wfId := range ids {
			entries += fmt.Sprintf("      { id = %q },\n", wfId)
		}
		return fmt.Sprintf("\n  workflows_config = {\n    workflows = [\n%s    ]\n  }\n", entries)
	}
	valid := workflows(testWorkflowUUID, secondWorkflowUUID)
	missingWorkflow := workflows(testWorkflowUUID)
	wrongOrder := workflows(secondWorkflowUUID, testWorkflowUUID)
	mismatch := regexp.MustCompile("does not match the stack template revision")

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader()),
		Steps: []resource.TestStep{
			// Create.
			{
				Config:      testAccStackConfig(wfGrpName, revision, id, missingWorkflow),
				PlanOnly:    true,
				ExpectError: mismatch,
			},
			{
				Config:      testAccStackConfig(wfGrpName, revision, id, wrongOrder),
				PlanOnly:    true,
				ExpectError: mismatch,
			},
			// Update with the same revision.
			{
				Config: testAccStackConfig(wfGrpName, revision, id, valid),
			},
			{
				Config:      testAccStackConfig(wfGrpName, revision, id, missingWorkflow),
				PlanOnly:    true,
				ExpectError: mismatch,
			},
			{
				Config:      testAccStackConfig(wfGrpName, revision, id, wrongOrder),
				PlanOnly:    true,
				ExpectError: mismatch,
			},
		},
	})
}

// TestAccStack_WorkflowsConfigMismatchRejectedOnRevisionChange verifies the same
// check when template_group_id moves to another revision: revision :1 declares
// only testWorkflowUUID, revision :2 declares testWorkflowUUID then
// secondWorkflowUUID. Keeping :1's workflows_config, or listing :2's workflows in
// the wrong order, must be rejected by `terraform plan` (ModifyPlan) before the
// upgrade is attempted.
func TestAccStack_WorkflowsConfigMismatchRejectedOnRevisionChange(t *testing.T) {
	acctest.SkipUnlessAcceptance(t)
	wfGrpName := acctest.ResourceName("stack-wfmismatchrev-wfgrp")
	wfTemplateName := acctest.ResourceName("stack-wfmismatchrev-wftmpl")
	stackTemplateName := acctest.ResourceName("stack-wfmismatchrev-stmpl")
	id := acctest.ResourceName("stack-wfmismatchrev")

	revision1 := setupStackDependencyChain(t, wfGrpName, wfTemplateName, stackTemplateName, id)
	revision2 := setupSecondStackTemplateRevisionTwoWorkflows(t, stackTemplateName, wfTemplateName)

	oneWorkflow := fmt.Sprintf(`
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
	mismatch := regexp.MustCompile("does not match the stack template revision")

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader()),
		Steps: []resource.TestStep{
			{
				Config: testAccStackConfig(wfGrpName, revision1, id, oneWorkflow),
			},
			{
				// Upgrading to :2 without declaring its second workflow.
				Config:      testAccStackConfig(wfGrpName, revision2, id, oneWorkflow),
				PlanOnly:    true,
				ExpectError: mismatch,
			},
			{
				// Upgrading to :2 with its workflows out of order.
				Config:      testAccStackConfig(wfGrpName, revision2, id, wrongOrder),
				PlanOnly:    true,
				ExpectError: mismatch,
			},
		},
	})
}

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
	acctest.SkipUnlessAcceptance(t)
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
	acctest.SkipUnlessAcceptance(t)
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
	acctest.SkipUnlessAcceptance(t)
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
//
// This also covers the updateWorkflowsFromConfig query param ToUpdateAPIModel
// sets: it tells the platform to push the updated workflows_config down to the
// live workflows. Every read, including the one Update does after the PATCH,
// passes refresh_workflows_config=true, so workflows_config is rebuilt from
// those live workflows. Without the param they would keep their old values,
// and the update steps here would fail with "inconsistent result after apply".
func TestAccStack_WorkflowsConfigUpdate(t *testing.T) {
	acctest.SkipUnlessAcceptance(t)
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

// TestAccStack_WorkflowsConfigRoundTripRemainingFields covers the per-workflow attributes the
// other workflows_config tests don't set: description, wf_type, parallel_execution,
// environment_variables, number_of_approvals_required, deployment_platform_config,
// iac_input_data, the rest of terraform_config (drift, approval, plan/init options, timeout,
// hooks, and a pre-plan workflow step), and mini_steps email notifications and wf_chaining.
// Each value must come back from the API exactly as sent — a mismatch fails the apply itself
// ("Provider produced inconsistent result after apply") — and the second step checks it then
// re-plans with no diff.
func TestAccStack_WorkflowsConfigRoundTripRemainingFields(t *testing.T) {
	acctest.SkipUnlessAcceptance(t)
	wfGrpName := acctest.ResourceName("stack-wfrest-wfgrp")
	wfTemplateName := acctest.ResourceName("stack-wfrest-wftmpl")
	stackTemplateName := acctest.ResourceName("stack-wfrest-stmpl")
	id := acctest.ResourceName("stack-wfrest")

	revision := setupStackDependencyChain(t, wfGrpName, wfTemplateName, stackTemplateName, id)
	stepTemplateID := setupStackWorkflowStepTemplate(t, acctest.ResourceName("stack-wfrest-step"))

	config := fmt.Sprintf(`
  workflows_config = {
    workflows = [
      {
        id                 = %[1]q
        description        = "workflow description"
        wf_type            = "TERRAFORM"
        parallel_execution = "enabled"

        approvers                    = ["alice@example.com"]
        number_of_approvals_required = 1

        environment_variables = [
          {
            kind = "PLAIN_TEXT"
            config = {
              var_name   = "WF_VAR"
              text_value = "wf-value"
            }
          }
        ]

        deployment_platform_config = [
          {
            kind = "AWS_RBAC"
            config = {
              integration_id = "/integrations/test-integration"
              profile_name   = "test-profile"
            }
          }
        ]

        vcs_config = {
          iac_input_data = {
            schema_type = "RAW_JSON"
            data        = jsonencode({ bucket_region = "eu-central-1" })
          }
        }

        terraform_config = {
          terraform_version       = "1.5.5"
          managed_terraform_state = true
          drift_check             = true
          drift_cron              = "0 */6 * * ? *"
          approval_pre_apply      = true
          terraform_plan_options  = "-input=false"
          terraform_init_options  = "-input=false"
          timeout                 = 3600
          pre_plan_hooks          = ["echo pre-plan"]
          post_apply_hooks        = ["echo post-apply"]

          pre_plan_wf_steps_config = [
            {
              name                = "pre-plan-step"
              wf_step_template_id = %[2]q
            }
          ]
        }

        mini_steps = {
          notifications = {
            email = {
              completed = [
                {
                  recipients = ["alice@example.com"]
                }
              ]
            }
          }
          wf_chaining = {
            errored = [
              {
                workflow_group_id = "kk"
                workflow_id       = "retest-of-bug-cewgh6dt-i7vp-0vo474rl"
              }
            ]
          }
        }
      }
    ]
  }
`, testWorkflowUUID, stepTemplateID)

	const res = "stackguardian_stack.test"
	wf := "workflows_config.workflows.0."
	tc := wf + "terraform_config."

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
					resource.TestCheckResourceAttr(res, wf+"description", "workflow description"),
					resource.TestCheckResourceAttr(res, wf+"wf_type", "TERRAFORM"),
					resource.TestCheckResourceAttr(res, wf+"parallel_execution", "enabled"),
					resource.TestCheckResourceAttr(res, wf+"approvers.0", "alice@example.com"),
					resource.TestCheckResourceAttr(res, wf+"number_of_approvals_required", "1"),

					resource.TestCheckResourceAttr(res, wf+"environment_variables.#", "1"),
					resource.TestCheckResourceAttr(res, wf+"environment_variables.0.kind", "PLAIN_TEXT"),
					resource.TestCheckResourceAttr(res, wf+"environment_variables.0.config.var_name", "WF_VAR"),
					resource.TestCheckResourceAttr(res, wf+"environment_variables.0.config.text_value", "wf-value"),

					resource.TestCheckResourceAttr(res, wf+"deployment_platform_config.#", "1"),
					resource.TestCheckResourceAttr(res, wf+"deployment_platform_config.0.kind", "AWS_RBAC"),
					resource.TestCheckResourceAttr(res, wf+"deployment_platform_config.0.config.integration_id", "/integrations/test-integration"),
					resource.TestCheckResourceAttr(res, wf+"deployment_platform_config.0.config.profile_name", "test-profile"),

					resource.TestCheckResourceAttr(res, wf+"vcs_config.iac_input_data.schema_type", "RAW_JSON"),
					resource.TestCheckResourceAttr(res, wf+"vcs_config.iac_input_data.data", `{"bucket_region":"eu-central-1"}`),

					resource.TestCheckResourceAttr(res, tc+"terraform_version", "1.5.5"),
					resource.TestCheckResourceAttr(res, tc+"managed_terraform_state", "true"),
					resource.TestCheckResourceAttr(res, tc+"drift_check", "true"),
					resource.TestCheckResourceAttr(res, tc+"drift_cron", "0 */6 * * ? *"),
					resource.TestCheckResourceAttr(res, tc+"approval_pre_apply", "true"),
					resource.TestCheckResourceAttr(res, tc+"terraform_plan_options", "-input=false"),
					resource.TestCheckResourceAttr(res, tc+"terraform_init_options", "-input=false"),
					resource.TestCheckResourceAttr(res, tc+"timeout", "3600"),
					resource.TestCheckResourceAttr(res, tc+"pre_plan_hooks.0", "echo pre-plan"),
					resource.TestCheckResourceAttr(res, tc+"post_apply_hooks.0", "echo post-apply"),
					resource.TestCheckResourceAttr(res, tc+"pre_plan_wf_steps_config.#", "1"),
					resource.TestCheckResourceAttr(res, tc+"pre_plan_wf_steps_config.0.name", "pre-plan-step"),
					resource.TestCheckResourceAttr(res, tc+"pre_plan_wf_steps_config.0.wf_step_template_id", stepTemplateID),

					resource.TestCheckResourceAttr(res, wf+"mini_steps.notifications.email.completed.0.recipients.0", "alice@example.com"),
					resource.TestCheckResourceAttr(res, wf+"mini_steps.wf_chaining.errored.0.workflow_group_id", "kk"),
					resource.TestCheckResourceAttr(res, wf+"mini_steps.wf_chaining.errored.0.workflow_id", "retest-of-bug-cewgh6dt-i7vp-0vo474rl"),
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

// TestAccStack_WorkflowsConfigRoundTripHooksAndMiniSteps covers the per-workflow attributes
// TestAccStack_WorkflowsConfigRoundTripRemainingFields doesn't set: the remaining
// terraform_config hooks and workflow steps (pre_apply, post_plan, post_apply),
// run_pre_init_hooks_on_drift, terraform_bin_path (on a private runner), a step's mount_points, a mini_steps
// webhook with a webhook_secret, and a wf_chaining entry that chains a stack (stack_id and
// stack_run_payload). Each value must come back from the API exactly as sent, and the
// second step re-plans with no diff.
func TestAccStack_WorkflowsConfigRoundTripHooksAndMiniSteps(t *testing.T) {
	acctest.SkipUnlessAcceptance(t)
	wfGrpName := acctest.ResourceName("stack-wfhooks-wfgrp")
	wfTemplateName := acctest.ResourceName("stack-wfhooks-wftmpl")
	stackTemplateName := acctest.ResourceName("stack-wfhooks-stmpl")
	id := acctest.ResourceName("stack-wfhooks")

	revision := setupStackDependencyChain(t, wfGrpName, wfTemplateName, stackTemplateName, id)
	stepTemplateID := setupStackWorkflowStepTemplate(t, acctest.ResourceName("stack-wfhooks-step"))

	config := fmt.Sprintf(`
  workflows_config = {
    workflows = [
      {
        id = %[1]q

        # terraform_bin_path only applies on a private runner: on any other the
        # platform drops it.
        runner_constraints = {
          type  = "private"
          names = ["runner-1"]
        }

        terraform_config = {
          run_pre_init_hooks_on_drift = true
          pre_apply_hooks             = ["echo pre-apply"]
          post_plan_hooks             = ["echo post-plan"]

          terraform_bin_path = [
            {
              source    = "/opt/terraform/bin"
              target    = "/usr/local/bin"
              read_only = true
            }
          ]

          pre_apply_wf_steps_config = [
            {
              name                = "pre-apply-step"
              wf_step_template_id = %[2]q
              mount_points = [
                {
                  source    = "/data/cache"
                  target    = "/cache"
                  read_only = false
                }
              ]
            }
          ]
          post_plan_wf_steps_config = [
            {
              name                = "post-plan-step"
              wf_step_template_id = %[2]q
            }
          ]
          post_apply_wf_steps_config = [
            {
              name                = "post-apply-step"
              wf_step_template_id = %[2]q
            }
          ]
        }

        mini_steps = {
          webhooks = {
            errored = [
              {
                webhook_name   = "on-errored"
                webhook_url    = "https://example.com/errored"
                webhook_secret = "hook-secret"
              }
            ]
          }
          wf_chaining = {
            completed = [
              {
                workflow_group_id = "chained-group"
                stack_id          = "chained-stack"
                stack_run_payload = jsonencode({ ActionName = "apply" })
              }
            ]
          }
        }
      }
    ]
  }
`, testWorkflowUUID, stepTemplateID)

	const res = "stackguardian_stack.test"
	wf := "workflows_config.workflows.0."
	tc := wf + "terraform_config."

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
					resource.TestCheckResourceAttr(res, tc+"run_pre_init_hooks_on_drift", "true"),
					resource.TestCheckResourceAttr(res, tc+"pre_apply_hooks.#", "1"),
					resource.TestCheckResourceAttr(res, tc+"pre_apply_hooks.0", "echo pre-apply"),
					resource.TestCheckResourceAttr(res, tc+"post_plan_hooks.#", "1"),
					resource.TestCheckResourceAttr(res, tc+"post_plan_hooks.0", "echo post-plan"),

					resource.TestCheckResourceAttr(res, tc+"terraform_bin_path.#", "1"),
					resource.TestCheckResourceAttr(res, tc+"terraform_bin_path.0.source", "/opt/terraform/bin"),
					resource.TestCheckResourceAttr(res, tc+"terraform_bin_path.0.target", "/usr/local/bin"),
					resource.TestCheckResourceAttr(res, tc+"terraform_bin_path.0.read_only", "true"),

					resource.TestCheckResourceAttr(res, tc+"pre_apply_wf_steps_config.#", "1"),
					resource.TestCheckResourceAttr(res, tc+"pre_apply_wf_steps_config.0.name", "pre-apply-step"),
					resource.TestCheckResourceAttr(res, tc+"pre_apply_wf_steps_config.0.wf_step_template_id", stepTemplateID),
					resource.TestCheckResourceAttr(res, tc+"pre_apply_wf_steps_config.0.mount_points.#", "1"),
					resource.TestCheckResourceAttr(res, tc+"pre_apply_wf_steps_config.0.mount_points.0.source", "/data/cache"),
					resource.TestCheckResourceAttr(res, tc+"pre_apply_wf_steps_config.0.mount_points.0.target", "/cache"),
					resource.TestCheckResourceAttr(res, tc+"pre_apply_wf_steps_config.0.mount_points.0.read_only", "false"),
					resource.TestCheckResourceAttr(res, tc+"post_plan_wf_steps_config.#", "1"),
					resource.TestCheckResourceAttr(res, tc+"post_plan_wf_steps_config.0.name", "post-plan-step"),
					resource.TestCheckResourceAttr(res, tc+"post_apply_wf_steps_config.#", "1"),
					resource.TestCheckResourceAttr(res, tc+"post_apply_wf_steps_config.0.name", "post-apply-step"),

					resource.TestCheckResourceAttr(res, wf+"mini_steps.webhooks.errored.#", "1"),
					resource.TestCheckResourceAttr(res, wf+"mini_steps.webhooks.errored.0.webhook_name", "on-errored"),
					resource.TestCheckResourceAttr(res, wf+"mini_steps.webhooks.errored.0.webhook_url", "https://example.com/errored"),
					resource.TestCheckResourceAttr(res, wf+"mini_steps.webhooks.errored.0.webhook_secret", "hook-secret"),

					resource.TestCheckResourceAttr(res, wf+"mini_steps.wf_chaining.completed.#", "1"),
					resource.TestCheckResourceAttr(res, wf+"mini_steps.wf_chaining.completed.0.workflow_group_id", "chained-group"),
					resource.TestCheckResourceAttr(res, wf+"mini_steps.wf_chaining.completed.0.stack_id", "chained-stack"),
					resource.TestCheckResourceAttr(res, wf+"mini_steps.wf_chaining.completed.0.stack_run_payload", `{"ActionName":"apply"}`),
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

// TestAccStack_TerraformBinPathRequiresPrivateRunner verifies terraform_bin_path is
// rejected at plan time on a workflow that doesn't run on a private runner. The
// platform drops it from such a workflow, so the apply used to fail with an
// "inconsistent result after apply" that didn't name the attribute. Step 1 declares a
// shared runner; step 2 declares no runner at all, so the organization's default would
// apply, which the provider can't see.
func TestAccStack_TerraformBinPathRequiresPrivateRunner(t *testing.T) {
	acctest.SkipUnlessAcceptance(t)
	wfGrpName := acctest.ResourceName("stack-binpath-wfgrp")
	wfTemplateName := acctest.ResourceName("stack-binpath-wftmpl")
	stackTemplateName := acctest.ResourceName("stack-binpath-stmpl")
	id := acctest.ResourceName("stack-binpath")

	revision := setupStackDependencyChain(t, wfGrpName, wfTemplateName, stackTemplateName, id)

	config := func(runnerConstraints string) string {
		return testAccStackConfig(wfGrpName, revision, id, fmt.Sprintf(`
  workflows_config = {
    workflows = [
      {
        id = %q
%s
        terraform_config = {
          terraform_bin_path = [
            {
              source = "/opt/terraform/bin"
              target = "/usr/local/bin"
            }
          ]
        }
      }
    ]
  }
`, testWorkflowUUID, runnerConstraints))
	}

	requiresPrivate := regexp.MustCompile(`terraform_bin_path requires a private runner`)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader()),
		Steps: []resource.TestStep{
			{
				Config: config(`
        runner_constraints = {
          type = "shared"
        }`),
				PlanOnly:    true,
				ExpectError: requiresPrivate,
			},
			{
				Config:      config(""),
				PlanOnly:    true,
				ExpectError: requiresPrivate,
			},
		},
	})
}

// TestAccStack_WorkflowsConfigCustomWfStepsConfig covers a workflow's top-level
// wf_steps_config, which only applies to CUSTOM workflows — so this stack's template revision
// is built on a CUSTOM workflow template rather than the TERRAFORM one the other tests use. The
// declared steps must come back from the API exactly as sent, then re-plan with no diff.
func TestAccStack_WorkflowsConfigCustomWfStepsConfig(t *testing.T) {
	acctest.SkipUnlessAcceptance(t)
	wfGrpName := acctest.ResourceName("stack-wfcustom-wfgrp")
	wfTemplateName := acctest.ResourceName("stack-wfcustom-wftmpl")
	stackTemplateName := acctest.ResourceName("stack-wfcustom-stmpl")
	id := acctest.ResourceName("stack-wfcustom")

	t.Cleanup(func() {
		logCleanupErr(t, fmt.Sprintf("delete workflow group %q", wfGrpName), deleteWorkflowGroupFixture(wfGrpName))
	})
	if err := createWorkflowGroupFixture(wfGrpName); err != nil && !is409(err) {
		t.Fatalf("TestAccStack_WorkflowsConfigCustomWfStepsConfig: create workflow group %q: %s", wfGrpName, err)
	}
	workflowTemplateID := setupStackWorkflowTemplateOfKind(t, wfTemplateName, workflowtemplates.WorkflowTemplateSourceConfigKindCustom)
	revision := setupStackTemplateChain(t, stackTemplateName, workflowTemplateID)
	stepTemplateID := setupStackWorkflowStepTemplate(t, acctest.ResourceName("stack-wfcustom-step"))
	t.Cleanup(func() { deleteStackFixture(wfGrpName, id) })

	config := fmt.Sprintf(`
  workflows_config = {
    workflows = [
      {
        id      = %[1]q
        wf_type = "CUSTOM"

        wf_steps_config = [
          {
            name                = "step-one"
            wf_step_template_id = %[2]q
            approval            = true
            timeout             = 600
            cmd_override        = "echo hello"

            environment_variables = [
              {
                kind = "PLAIN_TEXT"
                config = {
                  var_name   = "STEP_VAR"
                  text_value = "step-value"
                }
              }
            ]
          }
        ]
      }
    ]
  }
`, testWorkflowUUID, stepTemplateID)

	const res = "stackguardian_stack.test"
	step := "workflows_config.workflows.0.wf_steps_config.0."

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
					resource.TestCheckResourceAttr(res, "workflows_config.workflows.0.wf_type", "CUSTOM"),
					resource.TestCheckResourceAttr(res, "workflows_config.workflows.0.wf_steps_config.#", "1"),
					resource.TestCheckResourceAttr(res, step+"name", "step-one"),
					resource.TestCheckResourceAttr(res, step+"wf_step_template_id", stepTemplateID),
					resource.TestCheckResourceAttr(res, step+"approval", "true"),
					resource.TestCheckResourceAttr(res, step+"timeout", "600"),
					resource.TestCheckResourceAttr(res, step+"cmd_override", "echo hello"),
					resource.TestCheckResourceAttr(res, step+"environment_variables.0.config.var_name", "STEP_VAR"),
					resource.TestCheckResourceAttr(res, step+"environment_variables.0.config.text_value", "step-value"),
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

// templateEnvVars is a single PLAIN_TEXT environment variable for template
// fixtures that need a non-empty environment_variables value to inherit.
func templateEnvVars(name, value string) []sgsdkgo.EnvVars {
	return []sgsdkgo.EnvVars{
		{
			Kind:   sgsdkgo.EnvVarsKindEnumPlainText,
			Config: &sgsdkgo.EnvVarConfig{VarName: name, TextValue: sgsdkgo.String(value)},
		},
	}
}

// TestAccStack_WorkflowsConfigExplicitEmptyListOverridesTemplate verifies an
// explicit [] in the stack's own workflow entry is kept as a real value rather
// than treated as unset: the stack template revision's workflow supplies
// non-empty approvers, environment_variables and terraform_config
// pre_init_hooks/post_apply_hooks, and the stack config sets each to [].
// The resolved values must stay empty, while terraform_config fields the
// stack left unset (terraform_version) are still inherited from the template.
func TestAccStack_WorkflowsConfigExplicitEmptyListOverridesTemplate(t *testing.T) {
	acctest.SkipUnlessAcceptance(t)
	wfGrpName := acctest.ResourceName("stack-wfempty-wfgrp")
	wfTemplateName := acctest.ResourceName("stack-wfempty-wftmpl")
	stackTemplateName := acctest.ResourceName("stack-wfempty-stmpl")
	id := acctest.ResourceName("stack-wfempty")

	revision := setupStackDependencyChainWith(t, wfGrpName, wfTemplateName, stackTemplateName, id, nil,
		func(wf *stacktemplaterevisions.StackTemplateRevisionWorkflow) {
			wf.Approvers = []string{"alice@example.com"}
			envVars := templateEnvVars("STACK_TPL_VAR", "stack-tpl-value")
			wf.EnvironmentVariables = []*sgsdkgo.EnvVars{&envVars[0]}
			wf.TerraformConfig.PreInitHooks = []string{"echo stack-tpl-pre-init"}
			wf.TerraformConfig.PostApplyHooks = []string{"echo stack-tpl-post-apply"}
		})

	config := fmt.Sprintf(`
  workflows_config = {
    workflows = [
      {
        id                    = %q
        approvers             = []
        environment_variables = []
        terraform_config = {
          pre_init_hooks   = []
          post_apply_hooks = []
        }
      }
    ]
  }
`, testWorkflowUUID)

	res := "stackguardian_stack.test"
	wf := "workflows_config.workflows.0."

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
					resource.TestCheckResourceAttr(res, wf+"approvers.#", "0"),
					resource.TestCheckResourceAttr(res, wf+"environment_variables.#", "0"),
					resource.TestCheckResourceAttr(res, wf+"terraform_config.pre_init_hooks.#", "0"),
					resource.TestCheckResourceAttr(res, wf+"terraform_config.post_apply_hooks.#", "0"),
					resource.TestCheckResourceAttr(res, wf+"terraform_config.terraform_version", "1.5.7"),
				),
			},
			{
				// Re-planning the same config must be a no-op: [] in config
				// matches [] in state, with nothing pulled in from the template.
				Config:   testAccStackConfig(wfGrpName, revision, id, config),
				PlanOnly: true,
			},
		},
	})
}

// TestAccStack_WorkflowsConfigInheritsExplicitEmptyFromStackTemplate verifies
// an explicit [] on the stack template revision's workflow is inherited as a
// real value rather than treated as unset: the workflow template revision
// (lowest layer) supplies non-empty environment_variables and
// terraform_config pre_init_hooks, the stack template revision (middle layer)
// sets both to [], and the stack leaves them unset. The middle layer's empty
// value must win over the lowest layer's non-empty one.
func TestAccStack_WorkflowsConfigInheritsExplicitEmptyFromStackTemplate(t *testing.T) {
	acctest.SkipUnlessAcceptance(t)
	wfGrpName := acctest.ResourceName("stack-wftplempty-wfgrp")
	wfTemplateName := acctest.ResourceName("stack-wftplempty-wftmpl")
	stackTemplateName := acctest.ResourceName("stack-wftplempty-stmpl")
	id := acctest.ResourceName("stack-wftplempty")

	revision := setupStackDependencyChainWith(t, wfGrpName, wfTemplateName, stackTemplateName, id,
		func(req *workflowtemplaterevisions.CreateWorkflowTemplateRevisionsRequest) {
			req.EnvironmentVariables = templateEnvVars("WF_TPL_VAR", "wf-tpl-value")
			req.TerraformConfig.PreInitHooks = []string{"echo wf-tpl-pre-init"}
		},
		func(wf *stacktemplaterevisions.StackTemplateRevisionWorkflow) {
			wf.EnvironmentVariables = []*sgsdkgo.EnvVars{}
			wf.TerraformConfig.PreInitHooks = []string{}
		})

	res := "stackguardian_stack.test"
	wf := "workflows_config.workflows.0."

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader()),
		Steps: []resource.TestStep{
			{
				Config: testAccStackConfig(wfGrpName, revision, id, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(res, wf+"environment_variables.#", "0"),
					resource.TestCheckResourceAttr(res, wf+"terraform_config.pre_init_hooks.#", "0"),
					resource.TestCheckResourceAttr(res, wf+"terraform_config.terraform_version", "1.5.7"),
				),
			},
			{
				Config:   testAccStackConfig(wfGrpName, revision, id, ""),
				PlanOnly: true,
			},
		},
	})
}

// TestAccStack_InvalidJSONRejected verifies the JSON-object attributes inside a
// workflow (vcs_config.iac_input_data.data, wf_steps_config[*].wf_step_input_data.data
// and mini_steps.wf_chaining.*[*].workflow_run_payload) are checked at plan time
// (validators.JSONObject). Invalid JSON used to be replaced by {} when building the
// request, so the inputs were silently dropped and the apply failed with "inconsistent
// result after apply". The validator fails before any API call, so no fixtures are
// needed.
func TestAccStack_InvalidJSONRejected(t *testing.T) {
	acctest.SkipUnlessAcceptance(t)

	withWorkflow := func(body string) string {
		return testAccStackConfig("my-workflow-group", fmt.Sprintf("/%s/my-stack-template:1", org), "my-stack", fmt.Sprintf(`
  workflows_config = {
    workflows = [
      {
        id = %q
%s
      }
    ]
  }
`, testWorkflowUUID, body))
	}
	iacInputData := func(data string) string {
		return withWorkflow(fmt.Sprintf(`
        vcs_config = {
          iac_input_data = {
            schema_type = "RAW_JSON"
            data        = %q
          }
        }`, data))
	}
	stepInputData := func(data string) string {
		return withWorkflow(fmt.Sprintf(`
        wf_steps_config = [
          {
            name                = "step"
            wf_step_template_id = "/%s/my-step-template:1"
            wf_step_input_data = {
              schema_type = "RAW_JSON"
              data        = %q
            }
          }
        ]`, org, data))
	}
	runPayload := func(payload string) string {
		return withWorkflow(fmt.Sprintf(`
        mini_steps = {
          wf_chaining = {
            completed = [
              {
                workflow_group_id    = "my-workflow-group"
                workflow_id          = "my-workflow"
                workflow_run_payload = %q
              }
            ]
          }
        }`, payload))
	}

	invalid := regexp.MustCompile(`(?s)Invalid JSON object.*not valid JSON`)
	notObject := regexp.MustCompile(`(?s)Invalid JSON object.*must be a JSON object`)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader()),
		Steps: []resource.TestStep{
			{Config: iacInputData(`{"bucket_region":`), PlanOnly: true, ExpectError: invalid},
			{Config: iacInputData(`["eu-central-1"]`), PlanOnly: true, ExpectError: notObject},
			{Config: stepInputData(`{"key":`), PlanOnly: true, ExpectError: invalid},
			{Config: runPayload(`["apply"]`), PlanOnly: true, ExpectError: notObject},
		},
	})
}

// TestAccStack_ReferenceValuesRoundTrip verifies ${reference::template-group.<index>...}
// values, which point a stack workflow at a sibling workflow's outputs, round trip.
// When core creates a stack's workflows it rewrites them to
// ${workflow::<group>.<stack>.<workflow id>...} in vcs_config.iac_input_data.data and
// PLAIN_TEXT environment variables, so the read after create returned a different value
// from the one planned, and the apply failed with "Provider produced inconsistent result
// after apply"; a refresh would also show it as a diff.
//
// The second workflow of the two-workflow revision (setupStackTemplateChainNoActions)
// references the first one's output (index 0). Step 1 checks state keeps the
// ${reference::} form while the platform holds the rewritten ${workflow::} form (which
// shows the rewrite really happened); step 2 checks a refresh and re-plan show no diff.
func TestAccStack_ReferenceValuesRoundTrip(t *testing.T) {
	acctest.SkipUnlessAcceptance(t)
	wfGrpName := acctest.ResourceName("stack-wfref-wfgrp")
	wfTemplateName := acctest.ResourceName("stack-wfref-wftmpl")
	stackTemplateName := acctest.ResourceName("stack-wfref-stmpl")
	id := acctest.ResourceName("stack-wfref")

	t.Cleanup(func() {
		logCleanupErr(t, fmt.Sprintf("delete workflow group %q", wfGrpName), deleteWorkflowGroupFixture(wfGrpName))
	})
	if err := createWorkflowGroupFixture(wfGrpName); err != nil && !is409(err) {
		t.Fatalf("TestAccStack_ReferenceValuesRoundTrip: create workflow group %q: %s", wfGrpName, err)
	}
	workflowTemplateID := setupStackWorkflowTemplate(t, wfTemplateName)
	revision := setupStackTemplateChainNoActions(t, stackTemplateName, workflowTemplateID)
	t.Cleanup(func() { deleteStackFixture(wfGrpName, id) })

	const reference = "${reference::template-group.0.outputs.bucket_region}"

	// $${ escapes HCL interpolation, so Terraform sees a literal ${reference::...}.
	config := fmt.Sprintf(`
  workflows_config = {
    workflows = [
      { id = %[1]q },
      {
        id = %[2]q
        vcs_config = {
          iac_input_data = {
            schema_type = "RAW_JSON"
            data        = jsonencode({ bucket_region = "$${reference::template-group.0.outputs.bucket_region}" })
          }
        }
        environment_variables = [
          {
            kind = "PLAIN_TEXT"
            config = {
              var_name   = "BUCKET_REGION"
              text_value = "$${reference::template-group.0.outputs.bucket_region}"
            }
          }
        ]
      }
    ]
  }
`, testWorkflowUUID, secondWorkflowUUID)

	const res = "stackguardian_stack.test"
	wf := "workflows_config.workflows.1."

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
					resource.TestCheckResourceAttr(res, wf+"vcs_config.iac_input_data.data",
						fmt.Sprintf(`{"bucket_region":%q}`, reference)),
					resource.TestCheckResourceAttr(res, wf+"environment_variables.0.config.text_value", reference),
					checkReferencesResolved(wfGrpName, id),
				),
			},
			{
				// A refresh reads the rewritten form back; re-planning must show no diff.
				Config:   testAccStackConfig(wfGrpName, revision, id, config),
				PlanOnly: true,
			},
		},
	})
}

// TestAccStack_InheritsEmptyIacInputData verifies a stack can inherit an explicitly empty
// vcs_config.iac_input_data ("No Inputs") from its stack template revision. The API
// stores and returns it as an iacInputData with no schemaType, which the stack's reader
// (flattenVcsConfig) dereferenced, so reading the stack panicked. The same reader also
// builds the plan for an upgrade to such a revision (reResolveWorkflowsConfigOnRevisionChange).
//
// Step 1 creates a stack that declares only the workflow id, so iac_input_data is
// inherited, and checks state has no iac_input_data fields set while the platform holds
// the explicit empty iacInputData; step 2 checks a re-plan shows no diff.
func TestAccStack_InheritsEmptyIacInputData(t *testing.T) {
	acctest.SkipUnlessAcceptance(t)
	wfGrpName := acctest.ResourceName("stack-emptyinput-wfgrp")
	wfTemplateName := acctest.ResourceName("stack-emptyinput-wftmpl")
	stackTemplateName := acctest.ResourceName("stack-emptyinput-stmpl")
	id := acctest.ResourceName("stack-emptyinput")

	revision := setupStackDependencyChainWith(t, wfGrpName, wfTemplateName, stackTemplateName, id, nil,
		func(wf *stacktemplaterevisions.StackTemplateRevisionWorkflow) {
			// The fixture's VcsConfig is non-nil (it carries IacVcsConfig).
			wf.VcsConfig.IacInputData = &sgsdkgo.IacInputData{}
		})

	const res = "stackguardian_stack.test"
	iacInput := "workflows_config.workflows.0.vcs_config.iac_input_data."

	// checkPlatformStoredEmptyInput reads the stack's live workflow through the SDK and
	// checks its iacInputData is present but empty: "No Inputs" was inherited and stored.
	checkPlatformStoredEmptyInput := func(*terraform.State) error {
		readResp, err := getClient().Stacks.ReadStack(context.TODO(), org, id, wfGrpName,
			&sgsdkgo.ReadStackQueryParams{RefreshWorkflowsConfig: sgsdkgo.Bool(true)})
		if err != nil {
			return fmt.Errorf("read stack %q: %w", id, err)
		}
		if readResp == nil || readResp.Msg == nil || readResp.Msg.WorkflowsConfig == nil ||
			len(readResp.Msg.WorkflowsConfig.Workflows) == 0 || readResp.Msg.WorkflowsConfig.Workflows[0] == nil {
			return fmt.Errorf("read stack %q: no workflows in the response", id)
		}
		vcs := readResp.Msg.WorkflowsConfig.Workflows[0].VcsConfig
		if vcs == nil || vcs.IacInputData == nil {
			return fmt.Errorf("stack %q: iacInputData was not stored; want the inherited empty object", id)
		}
		in := vcs.IacInputData
		if in.SchemaType != nil || in.SchemaId != nil || (in.Data != nil && len(*in.Data) > 0) {
			return fmt.Errorf("stack %q: iacInputData = %+v, want an empty object", id, in)
		}
		return nil
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader()),
		Steps: []resource.TestStep{
			{
				// workflows_config declares only the workflow id (testAccStackConfig's
				// default), so vcs_config.iac_input_data is inherited from the revision.
				Config: testAccStackConfig(wfGrpName, revision, id, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr(res, iacInput+"schema_type"),
					resource.TestCheckNoResourceAttr(res, iacInput+"schema_id"),
					resource.TestCheckNoResourceAttr(res, iacInput+"data"),
					checkPlatformStoredEmptyInput,
				),
			},
			{
				// Round trips with no diff.
				Config:   testAccStackConfig(wfGrpName, revision, id, ""),
				PlanOnly: true,
			},
		},
	})
}

// TestAccStack_OverridesIacInputDataWithEmpty verifies a stack can set
// vcs_config.iac_input_data = {} to override inputs it would otherwise inherit from its
// stack template revision with "No Inputs". schema_type is Optional for this (once any
// field is set, schema_type and data are both required), the provider sends an
// explicitly empty iacInputData, and the API keeps it rather than filling in the
// template's (it only fills iacInputData when the workflow has none).
//
// The revision's workflow sets RAW_JSON inputs; the stack declares iac_input_data = {}.
// Step 1 checks state has no iac_input_data fields and the platform stored an empty
// iacInputData (not the template's); step 2 checks a re-plan shows no diff.
func TestAccStack_OverridesIacInputDataWithEmpty(t *testing.T) {
	acctest.SkipUnlessAcceptance(t)
	wfGrpName := acctest.ResourceName("stack-overrideinput-wfgrp")
	wfTemplateName := acctest.ResourceName("stack-overrideinput-wftmpl")
	stackTemplateName := acctest.ResourceName("stack-overrideinput-stmpl")
	id := acctest.ResourceName("stack-overrideinput")

	revision := setupStackDependencyChainWith(t, wfGrpName, wfTemplateName, stackTemplateName, id, nil,
		func(wf *stacktemplaterevisions.StackTemplateRevisionWorkflow) {
			data := map[string]interface{}{"bucket_region": "eu-central-1"}
			wf.VcsConfig.IacInputData = &sgsdkgo.IacInputData{
				SchemaType: sgsdkgo.IacInputDataSchemaTypeEnumRawJson.Ptr(),
				Data:       &data,
			}
		})

	config := fmt.Sprintf(`
  workflows_config = {
    workflows = [
      {
        id = %q
        vcs_config = {
          iac_input_data = {}
        }
      }
    ]
  }
`, testWorkflowUUID)

	const res = "stackguardian_stack.test"
	iacInput := "workflows_config.workflows.0.vcs_config.iac_input_data."

	// checkPlatformStoredEmptyInput reads the stack's live workflow through the SDK and
	// checks its iacInputData is the empty override, not the template's RAW_JSON inputs.
	checkPlatformStoredEmptyInput := func(*terraform.State) error {
		readResp, err := getClient().Stacks.ReadStack(context.TODO(), org, id, wfGrpName,
			&sgsdkgo.ReadStackQueryParams{RefreshWorkflowsConfig: sgsdkgo.Bool(true)})
		if err != nil {
			return fmt.Errorf("read stack %q: %w", id, err)
		}
		if readResp == nil || readResp.Msg == nil || readResp.Msg.WorkflowsConfig == nil ||
			len(readResp.Msg.WorkflowsConfig.Workflows) == 0 || readResp.Msg.WorkflowsConfig.Workflows[0] == nil {
			return fmt.Errorf("read stack %q: no workflows in the response", id)
		}
		vcs := readResp.Msg.WorkflowsConfig.Workflows[0].VcsConfig
		if vcs == nil || vcs.IacInputData == nil {
			return fmt.Errorf("stack %q: iacInputData was not stored; want the empty override", id)
		}
		in := vcs.IacInputData
		if in.SchemaType != nil || in.SchemaId != nil || (in.Data != nil && len(*in.Data) > 0) {
			return fmt.Errorf("stack %q: iacInputData = %+v, want the empty override, not the template's inputs", id, in)
		}
		return nil
	}

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
					resource.TestCheckNoResourceAttr(res, iacInput+"schema_type"),
					resource.TestCheckNoResourceAttr(res, iacInput+"data"),
					checkPlatformStoredEmptyInput,
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

// testReference is the ${reference::} value the reference tests use: the second
// workflow pointing at an output of the first (index 0).
const testReference = "${reference::template-group.0.outputs.bucket_region}"

// checkReferencesResolved reads the stack's live workflows through the SDK and checks
// the second workflow (secondWorkflowUUID) holds references in their resolved
// ${workflow::<group>.<stack>.<workflow id>...} form, in iac_input_data.data
// "bucket_region" and the PLAIN_TEXT BUCKET_REGION environment variable. The platform
// is expected to resolve ${reference::} values whenever it writes a workflow, on create
// and on update: storing the raw reference on an update would change a value that was
// previously resolved.
func checkReferencesResolved(wfGrpName, id string) resource.TestCheckFunc {
	return func(*terraform.State) error {
		readResp, err := getClient().Stacks.ReadStack(context.TODO(), org, id, wfGrpName,
			&sgsdkgo.ReadStackQueryParams{RefreshWorkflowsConfig: sgsdkgo.Bool(true)})
		if err != nil {
			return fmt.Errorf("read stack %q: %w", id, err)
		}
		if readResp == nil || readResp.Msg == nil || readResp.Msg.WorkflowsConfig == nil {
			return fmt.Errorf("read stack %q: no workflows_config in response", id)
		}
		for _, w := range readResp.Msg.WorkflowsConfig.Workflows {
			if w == nil || w.Id == nil || *w.Id != secondWorkflowUUID {
				continue
			}
			if w.VcsConfig == nil || w.VcsConfig.IacInputData == nil || w.VcsConfig.IacInputData.Data == nil {
				return fmt.Errorf("workflow %s: no iac_input_data.data on the platform", secondWorkflowUUID)
			}
			data, _ := (*w.VcsConfig.IacInputData.Data)["bucket_region"].(string)
			if !strings.HasPrefix(data, "${workflow::") {
				return fmt.Errorf("workflow %s: iac_input_data.data bucket_region = %q, want the resolved ${workflow::} form", secondWorkflowUUID, data)
			}
			for _, env := range w.EnvironmentVariables {
				if env != nil && env.Config != nil && env.Config.VarName == "BUCKET_REGION" {
					if env.Config.TextValue == nil || !strings.HasPrefix(*env.Config.TextValue, "${workflow::") {
						return fmt.Errorf("workflow %s: BUCKET_REGION text_value = %v, want the resolved ${workflow::} form", secondWorkflowUUID, env.Config.TextValue)
					}
					return nil
				}
			}
			return fmt.Errorf("workflow %s: BUCKET_REGION environment variable not found on the platform", secondWorkflowUUID)
		}
		return fmt.Errorf("workflow %s not found on stack %q", secondWorkflowUUID, id)
	}
}

// referenceWorkflowsConfig returns a workflows_config whose second workflow declares
// iac_input_data.data and a PLAIN_TEXT BUCKET_REGION environment variable both set to
// value (an HCL expression or literal, written into the config as is).
func referenceWorkflowsConfig(value string) string {
	return fmt.Sprintf(`
  workflows_config = {
    workflows = [
      { id = %[1]q },
      {
        id = %[2]q
        vcs_config = {
          iac_input_data = {
            schema_type = "RAW_JSON"
            data        = jsonencode({ bucket_region = %[3]s })
          }
        }
        environment_variables = [
          {
            kind = "PLAIN_TEXT"
            config = {
              var_name   = "BUCKET_REGION"
              text_value = %[3]s
            }
          }
        ]
      }
    ]
  }
`, testWorkflowUUID, secondWorkflowUUID, value)
}

// TestAccStack_ReferenceValuesResolvedOnUpdate verifies ${reference::} values declared
// on a stack workflow are resolved by the platform on update as well as on create, and
// still round trip in state. The provider plans and sends the ${reference::} form (it
// keeps that form in state, see preserveReferenceForms), so every update re-sends it; if
// the platform stored it unresolved on update, a value that was resolved at create would
// change in the platform's records.
//
// Step 1 creates with references (resolved on create). Step 2 is an unrelated update
// (description), which re-sends them. Step 3 switches them to a literal, step 4 back to
// references, which must be resolved by the update. Each apply also checks state keeps
// the declared form, and the last step checks a re-plan shows no diff.
func TestAccStack_ReferenceValuesResolvedOnUpdate(t *testing.T) {
	acctest.SkipUnlessAcceptance(t)
	wfGrpName := acctest.ResourceName("stack-refupd-wfgrp")
	wfTemplateName := acctest.ResourceName("stack-refupd-wftmpl")
	stackTemplateName := acctest.ResourceName("stack-refupd-stmpl")
	id := acctest.ResourceName("stack-refupd")

	t.Cleanup(func() {
		logCleanupErr(t, fmt.Sprintf("delete workflow group %q", wfGrpName), deleteWorkflowGroupFixture(wfGrpName))
	})
	if err := createWorkflowGroupFixture(wfGrpName); err != nil && !is409(err) {
		t.Fatalf("TestAccStack_ReferenceValuesResolvedOnUpdate: create workflow group %q: %s", wfGrpName, err)
	}
	workflowTemplateID := setupStackWorkflowTemplate(t, wfTemplateName)
	revision := setupStackTemplateChainNoActions(t, stackTemplateName, workflowTemplateID)
	t.Cleanup(func() { deleteStackFixture(wfGrpName, id) })

	// $${ escapes HCL interpolation, so Terraform sees a literal ${reference::...}.
	withReferences := referenceWorkflowsConfig(`"$${reference::template-group.0.outputs.bucket_region}"`)
	withLiteral := referenceWorkflowsConfig(`"eu-central-1"`)

	const res = "stackguardian_stack.test"
	wf := "workflows_config.workflows.1."
	stateKeepsReferences := resource.ComposeAggregateTestCheckFunc(
		resource.TestCheckResourceAttr(res, wf+"vcs_config.iac_input_data.data", fmt.Sprintf(`{"bucket_region":%q}`, testReference)),
		resource.TestCheckResourceAttr(res, wf+"environment_variables.0.config.text_value", testReference),
	)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader()),
		Steps: []resource.TestStep{
			{
				// Create: resolved on create.
				Config: testAccStackConfig(wfGrpName, revision, id, withReferences),
				Check:  resource.ComposeAggregateTestCheckFunc(stateKeepsReferences, checkReferencesResolved(wfGrpName, id)),
			},
			{
				// Unrelated update: the references are re-sent and must stay resolved.
				Config: testAccStackConfig(wfGrpName, revision, id, withReferences+`description = "unrelated update"`),
				Check:  resource.ComposeAggregateTestCheckFunc(stateKeepsReferences, checkReferencesResolved(wfGrpName, id)),
			},
			{
				// Switch to a literal value.
				Config: testAccStackConfig(wfGrpName, revision, id, withLiteral+`description = "unrelated update"`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(res, wf+"vcs_config.iac_input_data.data", `{"bucket_region":"eu-central-1"}`),
					resource.TestCheckResourceAttr(res, wf+"environment_variables.0.config.text_value", "eu-central-1"),
				),
			},
			{
				// Back to references, set by an update this time: must be resolved.
				Config: testAccStackConfig(wfGrpName, revision, id, withReferences+`description = "unrelated update"`),
				Check:  resource.ComposeAggregateTestCheckFunc(stateKeepsReferences, checkReferencesResolved(wfGrpName, id)),
			},
			{
				// Round trips with no diff.
				Config:   testAccStackConfig(wfGrpName, revision, id, withReferences+`description = "unrelated update"`),
				PlanOnly: true,
			},
		},
	})
}

// TestAccStack_InheritedReferenceValues verifies ${reference::} values a stack workflow
// inherits from its stack template revision (not declared on the stack) are resolved by
// the platform, on create and when an unrelated update re-sends them, and round trip
// with no diff. The revision's second workflow sets iac_input_data.data and a PLAIN_TEXT
// BUCKET_REGION environment variable to a reference to the first workflow's output; the
// stack declares only the workflow ids. Inherited values are unknown at plan time on
// create, so state takes whatever the platform returns.
func TestAccStack_InheritedReferenceValues(t *testing.T) {
	acctest.SkipUnlessAcceptance(t)
	wfGrpName := acctest.ResourceName("stack-refinh-wfgrp")
	wfTemplateName := acctest.ResourceName("stack-refinh-wftmpl")
	stackTemplateName := acctest.ResourceName("stack-refinh-stmpl")
	id := acctest.ResourceName("stack-refinh")

	t.Cleanup(func() {
		logCleanupErr(t, fmt.Sprintf("delete workflow group %q", wfGrpName), deleteWorkflowGroupFixture(wfGrpName))
	})
	if err := createWorkflowGroupFixture(wfGrpName); err != nil && !is409(err) {
		t.Fatalf("TestAccStack_InheritedReferenceValues: create workflow group %q: %s", wfGrpName, err)
	}
	workflowTemplateID := setupStackWorkflowTemplate(t, wfTemplateName)
	revision := setupStackTemplateChainNoActionsWith(t, stackTemplateName, workflowTemplateID,
		func(wf *stacktemplaterevisions.StackTemplateRevisionWorkflow) {
			data := map[string]interface{}{"bucket_region": testReference}
			wf.VcsConfig.IacInputData = &sgsdkgo.IacInputData{
				SchemaType: sgsdkgo.IacInputDataSchemaTypeEnumRawJson.Ptr(),
				Data:       &data,
			}
			wf.EnvironmentVariables = []*sgsdkgo.EnvVars{{
				Kind:   sgsdkgo.EnvVarsKindEnumPlainText,
				Config: &sgsdkgo.EnvVarConfig{VarName: "BUCKET_REGION", TextValue: sgsdkgo.String(testReference)},
			}}
		})
	t.Cleanup(func() { deleteStackFixture(wfGrpName, id) })

	idsOnly := fmt.Sprintf(`
  workflows_config = {
    workflows = [
      { id = %q },
      { id = %q }
    ]
  }
`, testWorkflowUUID, secondWorkflowUUID)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader()),
		Steps: []resource.TestStep{
			{
				// Create: the inherited references are resolved.
				Config: testAccStackConfig(wfGrpName, revision, id, idsOnly),
				Check:  checkReferencesResolved(wfGrpName, id),
			},
			{
				// Round trips with no diff.
				Config:   testAccStackConfig(wfGrpName, revision, id, idsOnly),
				PlanOnly: true,
			},
			{
				// Unrelated update: the inherited values are re-sent and must stay resolved.
				Config: testAccStackConfig(wfGrpName, revision, id, idsOnly+`description = "unrelated update"`),
				Check:  checkReferencesResolved(wfGrpName, id),
			},
			{
				// Round trips with no diff.
				Config:   testAccStackConfig(wfGrpName, revision, id, idsOnly+`description = "unrelated update"`),
				PlanOnly: true,
			},
		},
	})
}
