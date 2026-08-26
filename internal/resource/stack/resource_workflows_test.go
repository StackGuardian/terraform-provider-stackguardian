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
					resource.TestCheckResourceAttrSet("stackguardian_stack.test", "workflows_config.workflows.0.template_id"),
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

// TestAccStack_WorkflowsConfigRevisionReResolution verifies
// reResolveWorkflowsConfigOnRevisionChange: a per-workflow field the stack
// leaves unset (number_of_approvals_required) must pick up a value the new
// revision provides that the old one didn't (step 2), and must go back to
// empty when a later revision drops it again (step 3) — rather than the
// value getting stuck at whatever UseStateForUnknown last carried forward.
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
				Check:  resource.TestCheckNoResourceAttr("stackguardian_stack.test", "workflows_config.workflows.0.number_of_approvals_required"),
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
				Check:  resource.TestCheckNoResourceAttr("stackguardian_stack.test", "workflows_config.workflows.0.number_of_approvals_required"),
			},
		},
	})
}

// TestAccStack_WorkflowsConfigInvalidEnums verifies invalid wf_type,
// parallel_execution, and is_active values each produce their own
// provider-side diagnostic (expandWorkflowsConfig) instead of being passed
// through to the API. Each step's apply fails before anything is created, so
// they can safely share one resource address across steps.
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
			{
				Config:      testAccStackConfig(wfGrpName, revision, id, workflowConfig("is_active", "maybe")),
				ExpectError: regexp.MustCompile("Invalid is_active"),
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
// lower layers.
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
          terraform_version = "1.6.0"
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
					"stackguardian_stack.test", "workflows_config.workflows.0.terraform_config.terraform_version", "1.6.0"),
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
// per-workflow attributes together: input_schemas (Optional+Computed guard
// on id — must not be forced to an empty string), approvers, user_schedules
// (per-workflow shape — cron/state Required, no "inputs" field, unlike the
// stack-level copy), context_tags, runner_constraints, and mini_steps.
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

        input_schemas = [
          {
            type = "FORM_JSONSCHEMA"
          }
        ]

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
					resource.TestCheckResourceAttr("stackguardian_stack.test", "workflows_config.workflows.0.input_schemas.0.type", "FORM_JSONSCHEMA"),
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
