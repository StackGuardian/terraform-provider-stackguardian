// Once a stackguardian_workflow_template_revision is published (is_public = "1"), the API
// only allows updating a small set of fields: LongDescription (the `description` attribute),
// Deprecation, Alias, and Notes. Every other attribute is rejected — but only when its value
// actually changes. Terraform always resends the entire config on every apply, not a diff,
// so an update to an allowed field on a revision that has other attributes set will always
// re-include those attributes' current values in the same request; if the API rejected a
// request merely for *containing* a disallowed field, rather than for that field's value
// differing from what's stored, updating an allowed field would be impossible on any
// published revision with other attributes set. The two tests in this file are the negative
// and positive halves of that same contract: DisallowedFieldUpdatesWhilePublished sweeps
// every non-allowed attribute expecting the API to reject an actual change to it, while
// UpdateAllowedFieldsWithOtherAttributesUnchangedWhilePublished sweeps the same attributes
// expecting success when they're present but held constant across the update.
//
// The provider doesn't enforce any of this client-side — nothing in ValidateConfig or
// ToUpdateAPIModel treats a published revision differently — so both are contract tests
// against the live API's own behavior, not tests of provider logic: they exist to catch
// drift if the API's rule is ever loosened, tightened, or the provider starts/stops sending
// a field it shouldn't.
//
// A second, unrelated reason every test here ends with a deprecation step: Terraform's own
// post-test destroy calls Delete() directly, with no deprecation logic of its own, and the
// API requires a published revision to be deprecated before it can be deleted. Skipping that
// final step fails the test run with "Error running post-test destroy, there may be dangling
// resources" even when everything the test itself checked was correct.
package workflowtemplaterevision_test

import (
	"fmt"
	"net/http"
	"regexp"
	"testing"
	"time"

	"github.com/StackGuardian/terraform-provider-stackguardian/internal/acctest"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/constants"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

// deprecationConfigBlock returns the HCL deprecation block appended to every test's final
// step in this file, or "" when deprecated is false.
func deprecationConfigBlock(deprecated bool) string {
	if !deprecated {
		return ""
	}
	return fmt.Sprintf(`
  deprecation = {
    effective_date = "%d"
    message        = "This revision is deprecated"
  }`, time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC).Unix())
}

// TestAccWorkflowTemplateRevision_DisallowedFieldUpdatesWhilePublished is a table-driven
// sweep asserting the API rejects an actual *change* to every attribute other than
// description/alias/notes/deprecation on a published revision. See the file-level doc
// comment for why this exists.
//
// The exact error text the API returns isn't pinned down here — ExpectError only asserts
// that apply fails, not what it says. Tighten each case's pattern with acctest.ErrorPattern
// once the real message is known from a live run.
func TestAccWorkflowTemplateRevision_DisallowedFieldUpdatesWhilePublished(t *testing.T) {
	const (
		baseCPU    = 500
		baseMemory = 1024
	)

	type testCase struct {
		name string
		// sourceConfigKind applies to both steps unless changedSourceConfigKind is set.
		// Defaults to TERRAFORM.
		sourceConfigKind        string
		changedSourceConfigKind string // overrides sourceConfigKind on the rejected step only
		changedUserJobCPU       int    // overrides baseCPU on the rejected step only
		changedUserJobMemory    int    // overrides baseMemory on the rejected step only
		changedConfig           string // HCL merged into the baseline config on the rejected step
	}

	cases := []testCase{
		{name: "tags", changedConfig: `tags = ["should-not-apply"]`},
		{name: "context_tags", changedConfig: `
  context_tags = {
    env = "should-not-apply"
  }`},
		{name: "environment_variables", changedConfig: `
  environment_variables = [
    {
      kind = "PLAIN_TEXT"
      config = {
        var_name   = "SHOULD_NOT_APPLY"
        text_value = "should-not-apply"
      }
    }
  ]`},
		{name: "input_schemas", changedConfig: `
  input_schemas = [
    {
      name = "should-not-apply"
      type = "RAW_JSON"
    }
  ]`},
		{name: "mini_steps", changedConfig: `
  mini_steps = {
    wf_chaining = {
      errored = [{
        workflow_group_id = "kk"
      }]
    }
  }`},
		{name: "runner_constraints", changedConfig: `
  runner_constraints = {
    type = "shared"
  }`},
		{name: "user_schedules", changedConfig: `
  user_schedules = [
    {
      cron  = "0 8 ? * MON *"
      state = "ENABLED"
    }
  ]`},
		{name: "approvers", changedConfig: `approvers = ["approver@example.com"]`},
		{name: "number_of_approvals_required", changedConfig: `number_of_approvals_required = 1`},
		{name: "user_job_cpu_and_memory", changedUserJobCPU: 1000, changedUserJobMemory: 2048},
		{name: "runtime_source", changedConfig: fmt.Sprintf(`
  runtime_source = {
    source_config_dest_kind = %q
    config = {
      is_private = false
      repo       = "https://github.com/StackGuardian/tf-null-resource.git"
    }
  }`, constants.GitOther)},
		{name: "terraform_config", changedConfig: `
  terraform_config = {
    terraform_version = "1.5.0"
  }`},
		{name: "deployment_platform_config", changedConfig: `
  deployment_platform_config = [{
    kind = "AWS_RBAC"
    config = {
      integration_id = "/integrations/test-integration"
    }
  }]`},
		{
			// wf_steps_config needs a non-TERRAFORM/OPENTOFU source_config_kind throughout
			// (both steps), or the provider's own wfStepsConfigNotAllowedForTerraformDiagnostics
			// ValidateConfig check would reject it before the request ever reaches the API —
			// a different failure than the one this test means to exercise.
			name:             "wf_steps_config",
			sourceConfigKind: "CUSTOM",
			changedConfig: `
  wf_steps_config = [
    {
      name                = "step-1"
      wf_step_template_id = "/tf-provider-test-org/dummy-step-template:1"
    }
  ]`,
		},
		{name: "source_config_kind", changedSourceConfigKind: "OPENTOFU"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			sourceConfigKind := tc.sourceConfigKind
			if sourceConfigKind == "" {
				sourceConfigKind = "TERRAFORM"
			}
			changedSourceConfigKind := tc.changedSourceConfigKind
			if changedSourceConfigKind == "" {
				changedSourceConfigKind = sourceConfigKind
			}
			changedCPU := baseCPU
			if tc.changedUserJobCPU != 0 {
				changedCPU = tc.changedUserJobCPU
			}
			changedMemory := baseMemory
			if tc.changedUserJobMemory != 0 {
				changedMemory = tc.changedUserJobMemory
			}

			templateID := acctest.ResourceName("tf-provider-wftr-published-disallowed-" + tc.name)
			alias := "revision-published-disallowed-" + tc.name

			registerWorkflowTemplateCleanup(t, templateID, 1)

			err := createWorkflowTemplateFixture(templateID, sourceConfigKind)
			if err != nil {
				t.Fatal(err)
			}

			customHeader := http.Header{}
			customHeader.Set("x-sg-internal-auth-orgid", "sg-provider-test")

			baseline := func(deprecated bool) string {
				return fmt.Sprintf(`
			  alias       = %q
			  is_public   = "1"
			  description = "Published revision"
			  notes       = "Published revision notes"
			  %s
			`, alias, deprecationConfigBlock(deprecated))
			}
			changed := fmt.Sprintf(`
			  alias       = %q
			  is_public   = "1"
			  description = "Published revision"
			  notes       = "Published revision notes"
			  %s
			`, alias, tc.changedConfig)

			resource.Test(t, resource.TestCase{
				PreCheck: func() { acctest.TestAccPreCheck(t) },
				TerraformVersionChecks: []tfversion.TerraformVersionCheck{
					tfversion.SkipBelow(tfversion.Version1_1_0),
				},
				ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader),
				Steps: []resource.TestStep{
					// Publish the revision.
					{
						Config: testAccWorkflowTemplateRevision(templateID, sourceConfigKind, baseCPU, baseMemory, baseline(false)),
						Check: resource.ComposeAggregateTestCheckFunc(
							resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "is_public", "1"),
						),
					},
					// Attempt to change the disallowed field; the API is expected to reject it.
					{
						Config:      testAccWorkflowTemplateRevision(templateID, changedSourceConfigKind, changedCPU, changedMemory, changed),
						ExpectError: regexp.MustCompile(`.`),
					},
					// Deprecate the still-published revision so the framework's own
					// post-test destroy doesn't hit "dangling resources".
					{
						Config: testAccWorkflowTemplateRevision(templateID, sourceConfigKind, baseCPU, baseMemory, baseline(true)),
						Check: resource.ComposeAggregateTestCheckFunc(
							resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "deprecation.message", "This revision is deprecated"),
						),
					},
				},
			})
		})
	}
}

// TestAccWorkflowTemplateRevision_UpdateAllowedFieldsWithOtherAttributesUnchangedWhilePublished
// is the positive counterpart to TestAccWorkflowTemplateRevision_DisallowedFieldUpdatesWhilePublished:
// a table-driven sweep confirming that updating description/alias/notes on a published
// revision succeeds even when every other attribute is present in the same request, as long
// as its value is identical to what's already stored. See the file-level doc comment for
// why this matters. The "baseline" case (no other attribute set at all) and the
// "runtime_source" case (a private repo with auth, present but unchanged) are what two
// narrower, now-removed tests used to cover individually; they're just two rows of this
// table now.
func TestAccWorkflowTemplateRevision_UpdateAllowedFieldsWithOtherAttributesUnchangedWhilePublished(t *testing.T) {
	const (
		baseCPU    = 500
		baseMemory = 1024
	)

	type testCase struct {
		name string
		// sourceConfigKind applies to every step. Defaults to TERRAFORM.
		sourceConfigKind string
		// constantConfig is HCL held byte-for-byte identical across every step — only
		// alias/description/notes (and, on the final step, deprecation) ever change.
		constantConfig string
	}

	cases := []testCase{
		{name: "baseline"},
		{name: "tags", constantConfig: `tags = ["stable-tag"]`},
		{name: "context_tags", constantConfig: `
  context_tags = {
    env = "stable"
  }`},
		{name: "environment_variables", constantConfig: `
  environment_variables = [
    {
      kind = "PLAIN_TEXT"
      config = {
        var_name   = "STABLE_VAR"
        text_value = "stable-value"
      }
    }
  ]`},
		{name: "input_schemas", constantConfig: `
  input_schemas = [
    {
      name = "stable-schema"
      type = "RAW_JSON"
    }
  ]`},
		{name: "mini_steps", constantConfig: `
  mini_steps = {
    wf_chaining = {
      errored = [{
        workflow_group_id = "kk"
      }]
    }
  }`},
		{name: "runner_constraints", constantConfig: `
  runner_constraints = {
    type = "shared"
  }`},
		{name: "user_schedules", constantConfig: `
  user_schedules = [
    {
      cron  = "0 8 ? * MON *"
      state = "ENABLED"
    }
  ]`},
		{name: "approvers", constantConfig: `approvers = ["approver@example.com"]`},
		{name: "number_of_approvals_required", constantConfig: `number_of_approvals_required = 1`},
		{name: "runtime_source", constantConfig: fmt.Sprintf(`
  runtime_source = {
    source_config_dest_kind = %q
    config = {
      is_private = true
      auth       = "/secrets/tf-provider-test-secret"
      repo       = "https://github.com/StackGuardian/tf-null-resource.git"
    }
  }`, constants.GitOther)},
		{name: "terraform_config", constantConfig: `
  terraform_config = {
    terraform_version = "1.5.0"
  }`},
		{name: "deployment_platform_config", constantConfig: `
  deployment_platform_config = [{
    kind = "AWS_RBAC"
    config = {
      integration_id = "/integrations/test-integration"
    }
  }]`},
		{
			// See the equivalent case in the negative test for why CUSTOM is needed here.
			name:             "wf_steps_config",
			sourceConfigKind: "CUSTOM",
			constantConfig: `
  wf_steps_config = [
    {
      name                = "step-1"
      wf_step_template_id = "/tf-provider-test-org/dummy-step-template:1"
    }
  ]`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			sourceConfigKind := tc.sourceConfigKind
			if sourceConfigKind == "" {
				sourceConfigKind = "TERRAFORM"
			}

			templateID := acctest.ResourceName("tf-provider-wftr-published-unchanged-" + tc.name)

			registerWorkflowTemplateCleanup(t, templateID, 1)

			err := createWorkflowTemplateFixture(templateID, sourceConfigKind)
			if err != nil {
				t.Fatal(err)
			}

			customHeader := http.Header{}
			customHeader.Set("x-sg-internal-auth-orgid", "sg-provider-test")

			templateRevisionCallback := func(alias, description, notes string, deprecated bool) string {
				return fmt.Sprintf(`
			  alias       = %q
			  is_public   = "1"
			  description = %q
			  notes       = %q
			  %s
			  %s
			`, alias, description, notes, tc.constantConfig, deprecationConfigBlock(deprecated))
			}

			resource.Test(t, resource.TestCase{
				PreCheck: func() { acctest.TestAccPreCheck(t) },
				TerraformVersionChecks: []tfversion.TerraformVersionCheck{
					tfversion.SkipBelow(tfversion.Version1_1_0),
				},
				ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader),
				Steps: []resource.TestStep{
					{
						Config: testAccWorkflowTemplateRevision(templateID, sourceConfigKind, baseCPU, baseMemory, templateRevisionCallback("revision-published-unchanged-"+tc.name+"-v1", "Initial description", "Initial notes", false)),
						Check: resource.ComposeAggregateTestCheckFunc(
							resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "is_public", "1"),
							resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "description", "Initial description"),
							resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "notes", "Initial notes"),
						),
					},
					// Update alias/description/notes only; tc.constantConfig (if any) is
					// resent with the exact same value — this must succeed.
					{
						Config: testAccWorkflowTemplateRevision(templateID, sourceConfigKind, baseCPU, baseMemory, templateRevisionCallback("revision-published-unchanged-"+tc.name+"-v2", "Updated description", "Updated notes", false)),
						Check: resource.ComposeAggregateTestCheckFunc(
							resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "is_public", "1"),
							resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "alias", "revision-published-unchanged-"+tc.name+"-v2"),
							resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "description", "Updated description"),
							resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "notes", "Updated notes"),
						),
					},
					// Deprecate before the framework's own post-test destroy runs.
					{
						Config: testAccWorkflowTemplateRevision(templateID, sourceConfigKind, baseCPU, baseMemory, templateRevisionCallback("revision-published-unchanged-"+tc.name+"-v2", "Updated description", "Updated notes", true)),
						Check: resource.ComposeAggregateTestCheckFunc(
							resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "deprecation.message", "This revision is deprecated"),
						),
					},
				},
			})
		})
	}
}
