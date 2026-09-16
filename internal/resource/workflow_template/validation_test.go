package workflowtemplate_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/StackGuardian/terraform-provider-stackguardian/internal/acctest"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/constants"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

// TestAccWorkflowTemplate_ValidateRuntimeSourceAuth exercises the runtime_source
// is_private/auth validator (ValidateRuntimeSourceAuth in validation.go), shared with
// workflow_template_revision. Every case here fails at plan time (ValidateConfig runs
// before Create), so no API call happens and no fixture/cleanup is needed.
func TestAccWorkflowTemplate_ValidateRuntimeSourceAuth(t *testing.T) {
	customHeader := http.Header{}
	customHeader.Set("x-sg-internal-auth-orgid", "sg-provider-test")

	runtimeSourceConfig := func(destKind string, isPrivate bool, auth string) string {
		authLine := ""
		if auth != "" {
			authLine = fmt.Sprintf("auth = %q", auth)
		}
		return fmt.Sprintf(`
  runtime_source = {
    source_config_dest_kind = %q
    config = {
      is_private = %t
      repo       = "https://example.com/repo.git"
      %s
    }
  }
`, destKind, isPrivate, authLine)
	}

	testCases := []struct {
		name        string
		additional  string
		expectError string
	}{
		{
			name:        "is_private true requires auth",
			additional:  runtimeSourceConfig(constants.GitOther, true, ""),
			expectError: `auth is required for this runtime_source`,
		},
		{
			name:        "non-GIT_OTHER requires auth even when is_private is false",
			additional:  runtimeSourceConfig(constants.GithubCom, false, ""),
			expectError: `auth is required for this runtime_source`,
		},
		{
			name:        "GIT_OTHER auth must start with /secrets/",
			additional:  runtimeSourceConfig(constants.GitOther, true, "/integrations/oops"),
			expectError: `auth must start with /secrets/ for ` + constants.GitOther,
		},
		{
			name:        "non-GIT_OTHER auth must start with /integration",
			additional:  runtimeSourceConfig(constants.GithubCom, true, "/secrets/oops"),
			expectError: `auth must start with /integration for this source_config_dest_kind`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resource.Test(t, resource.TestCase{
				PreCheck: func() { acctest.TestAccPreCheck(t) },
				TerraformVersionChecks: []tfversion.TerraformVersionCheck{
					tfversion.SkipBelow(tfversion.Version1_1_0),
				},
				ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader),
				Steps: []resource.TestStep{
					{
						Config:      testAccWorkflowTemplate("does-not-need-to-exist", sourceConfigKind, tc.additional),
						ExpectError: acctest.TFStandardErrorPattern(tc.expectError),
					},
				},
			})
		})
	}
}

// TestAccWorkflowTemplate_RuntimeSourceGitOtherAuthWithPublicRepo confirms GIT_OTHER may
// set auth (e.g. an access token) even when is_private is false — a public repo can still
// use a token, so this is not the is_private/auth contradiction the validator rejects for
// GIT_OTHER. Unlike the ValidateConfig cases above, this config is valid and actually
// creates a resource.
func TestAccWorkflowTemplate_RuntimeSourceGitOtherAuthWithPublicRepo(t *testing.T) {
	templateName := acctest.ResourceName("tf-provider-wt-git-other-public-auth")

	t.Cleanup(func() { deleteWorkflowTemplateFixture(templateName) })

	customHeader := http.Header{}
	customHeader.Set("x-sg-internal-auth-orgid", "sg-provider-test")

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowTemplate(templateName, sourceConfigKind, `
  runtime_source = {
    source_config_dest_kind = "GIT_OTHER"
    config = {
      is_private = false
      auth       = "/secrets/public-repo-token"
      repo       = "https://example.com/repo.git"
    }
  }
`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_template.test", "runtime_source.config.is_private", "false"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template.test", "runtime_source.config.auth", "/secrets/public-repo-token"),
				),
			},
		},
	})
}
