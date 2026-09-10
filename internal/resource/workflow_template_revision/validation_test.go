package workflowtemplaterevision_test

import (
	"net/http"
	"testing"

	"github.com/StackGuardian/terraform-provider-stackguardian/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

// TestAccWorkflowTemplateRevision_ValidateWfStepsConfigNotAllowedForTerraform confirms
// wf_steps_config is rejected at plan time when source_config_kind is TERRAFORM or
// OPENTOFU — those kinds use fixed, built-in run steps instead. Since ValidateConfig runs
// before Create, no API call happens and no fixture/cleanup is needed.
func TestAccWorkflowTemplateRevision_ValidateWfStepsConfigNotAllowedForTerraform(t *testing.T) {
	customHeader := http.Header{}
	customHeader.Set("x-sg-internal-auth-orgid", "sg-provider-test")

	wfStepsConfig := `
  alias = "revision-validate-wfsteps"

  wf_steps_config = [
    {
      name                = "step-1"
      wf_step_template_id = "/tf-provider-test-org/dummy-step-template:1"
    }
  ]
`

	for _, sourceConfigKind := range []string{"TERRAFORM", "OPENTOFU"} {
		t.Run(sourceConfigKind, func(t *testing.T) {
			resource.Test(t, resource.TestCase{
				PreCheck: func() { acctest.TestAccPreCheck(t) },
				TerraformVersionChecks: []tfversion.TerraformVersionCheck{
					tfversion.SkipBelow(tfversion.Version1_1_0),
				},
				ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader),
				Steps: []resource.TestStep{
					{
						Config:      testAccWorkflowTemplateRevision("does-not-need-to-exist", sourceConfigKind, 500, 1024, wfStepsConfig),
						ExpectError: acctest.ErrorPattern("wf_steps_config is not allowed when source_config_kind is TERRAFORM or OPENTOFU; those workflow types use fixed, built-in run steps instead."),
					},
				},
			})
		})
	}
}

// TestAccWorkflowTemplateRevision_ValidateRuntimeSourceAuthRequired confirms the shared
// runtime_source is_private/auth validator (internal/resource/workflow_template's
// ValidateRuntimeSourceAuth) is wired up on the revision resource too. The full rule matrix
// is covered once, on workflow_template, since the validator is shared — this just proves
// the wiring here, not the whole rule set again.
func TestAccWorkflowTemplateRevision_ValidateRuntimeSourceAuthRequired(t *testing.T) {
	customHeader := http.Header{}
	customHeader.Set("x-sg-internal-auth-orgid", "sg-provider-test")

	config := testAccWorkflowTemplateRevision("does-not-need-to-exist", "CUSTOM", 500, 1024, `
  alias = "revision-validate-runtime-source"

  runtime_source = {
    source_config_dest_kind = "GITHUB_COM"
    config = {
      is_private = true
      repo       = "https://example.com/repo.git"
    }
  }
`)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader),
		Steps: []resource.TestStep{
			{
				Config:      config,
				ExpectError: acctest.ErrorPattern("auth is required for this runtime_source"),
			},
		},
	})
}
