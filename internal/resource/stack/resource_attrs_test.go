package stack_test

import (
	"regexp"
	"testing"

	"github.com/StackGuardian/terraform-provider-stackguardian/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

// TestAccStack_EnvironmentVariables covers environment_variables (Optional,
// not Computed): round trip with a value set, a stable no-op re-apply, and
// then omitting it entirely (must clear cleanly, not error — Optional here,
// unlike stack_template_revision's copy of this shape which is Required).
func TestAccStack_EnvironmentVariables(t *testing.T) {
	wfGrpName := "tf-provider-stack-envvars-wfgrp"
	wfTemplateName := "tf-provider-stack-envvars-wftmpl"
	stackTemplateName := "tf-provider-stack-envvars-stmpl"
	id := "tf-provider-stack-envvars"

	revision := setupStackDependencyChain(t, wfGrpName, wfTemplateName, stackTemplateName, id)

	envConfig := `
  environment_variables = [
    {
      kind = "PLAIN_TEXT"
      config = {
        var_name   = "MY_VAR"
        text_value = "my-value"
      }
    }
  ]
`

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader()),
		Steps: []resource.TestStep{
			{
				Config: testAccStackConfig(wfGrpName, revision, id, envConfig),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_stack.test", "environment_variables.0.kind", "PLAIN_TEXT"),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "environment_variables.0.config.var_name", "MY_VAR"),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "environment_variables.0.config.text_value", "my-value"),
				),
			},
			{
				// Round trips with no diff.
				Config:   testAccStackConfig(wfGrpName, revision, id, envConfig),
				PlanOnly: true,
			},
			{
				// Omitted entirely — Optional (not Required/Computed), must clear cleanly.
				Config: testAccStackConfig(wfGrpName, revision, id, ""),
				Check:  resource.TestCheckResourceAttr("stackguardian_stack.test", "environment_variables.#", "0"),
			},
		},
	})
}

// TestAccStack_DeploymentPlatformConfigInvalidKind verifies that an invalid
// deployment_platform_config.kind produces the "Invalid deployment platform
// config kind" diagnostic from expandDeploymentPlatformConfig (client-side,
// before any API call) instead of silently passing through.
func TestAccStack_DeploymentPlatformConfigInvalidKind(t *testing.T) {
	wfGrpName := "tf-provider-stack-dpcinvalid-wfgrp"
	wfTemplateName := "tf-provider-stack-dpcinvalid-wftmpl"
	stackTemplateName := "tf-provider-stack-dpcinvalid-stmpl"
	id := "tf-provider-stack-dpcinvalid"

	revision := setupStackDependencyChain(t, wfGrpName, wfTemplateName, stackTemplateName, id)

	invalidConfig := `
  deployment_platform_config = [
    {
      kind = "NOT_A_REAL_KIND"
      config = {
        integration_id = "does-not-matter"
      }
    }
  ]
`

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader()),
		Steps: []resource.TestStep{
			{
				Config:      testAccStackConfig(wfGrpName, revision, id, invalidConfig),
				ExpectError: regexp.MustCompile("Invalid deployment platform config kind"),
			},
		},
	})
}

// TestAccStack_MiniStepsAndUserSchedules covers the stack-level mini_steps
// (notifications/webhooks/wf_chaining) and user_schedules[].inputs (the
// StackAction shape: action_type only) round trip, and that inputs is
// omittable entirely.
func TestAccStack_MiniStepsAndUserSchedules(t *testing.T) {
	wfGrpName := "tf-provider-stack-ministeps-wfgrp"
	wfTemplateName := "tf-provider-stack-ministeps-wftmpl"
	stackTemplateName := "tf-provider-stack-ministeps-stmpl"
	id := "tf-provider-stack-ministeps"

	revision := setupStackDependencyChain(t, wfGrpName, wfTemplateName, stackTemplateName, id)

	config := `
  mini_steps = {
    notifications = {
      email = {
        completed = [
          { recipients = ["a@example.com"] }
        ]
      }
    }
    webhooks = {
      completed = [
        {
          webhook_name = "on-completed"
          webhook_url  = "https://example.com/hook"
        }
      ]
    }
  }

  user_schedules = [
    {
      name  = "nightly"
      desc  = "Nightly run"
      cron  = "0 8 ? * MON *"
      state = "ENABLED"
      inputs = {
        action_type = "apply"
      }
    },
    {
      cron  = "0 9 ? * TUE *"
      state = "DISABLED"
    }
  ]
`

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
					resource.TestCheckResourceAttr("stackguardian_stack.test", "mini_steps.notifications.email.completed.0.recipients.0", "a@example.com"),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "mini_steps.webhooks.completed.0.webhook_name", "on-completed"),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "mini_steps.webhooks.completed.0.webhook_url", "https://example.com/hook"),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "user_schedules.0.name", "nightly"),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "user_schedules.0.cron", "0 8 ? * MON *"),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "user_schedules.0.inputs.action_type", "apply"),
					// Second entry omits inputs entirely — must not error.
					resource.TestCheckResourceAttr("stackguardian_stack.test", "user_schedules.1.cron", "0 9 ? * TUE *"),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "user_schedules.1.state", "DISABLED"),
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
