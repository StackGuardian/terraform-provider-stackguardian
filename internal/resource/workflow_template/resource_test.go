package workflowtemplate_test

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/StackGuardian/terraform-provider-stackguardian/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

var sourceConfigKind = "TERRAFORM"
var org = os.Getenv("STACKGUARDIAN_ORG_NAME")

// deleteWorkflowTemplateFixture is a safety-net cleanup: Terraform's own destroy step tears
// the template down in the normal case. This exists so a test that fails before reaching
// destroy (e.g. a failed assertion) doesn't leave it behind. Errors are ignored — the
// resource may already be gone. id is the template's id, which defaults to its
// template_name when the config doesn't set an explicit id.
func deleteWorkflowTemplateFixture(id string) {
	client := acctest.SGClient()
	client.WorkflowTemplates.DeleteWorkflowTemplate(context.TODO(), org, id)
}

func testAccWorkflowTemplate(name, sourceConfigKind, additionalConfig string) string {
	return fmt.Sprintf(`
		resource "stackguardian_workflow_template" "test" {
		  template_name      = %q
		  source_config_kind = %q
		
		  %s
		}
		`, name, sourceConfigKind, additionalConfig)
}

func TestAccWorkflowTemplate_Basic(t *testing.T) {
	templateName := acctest.ResourceName("tf-provider-workflow-template-1")

	t.Cleanup(func() { deleteWorkflowTemplateFixture(templateName) })

	customHeader := http.Header{}
	customHeader.Set("x-sg-internal-auth-orgid", "sg-provider-test")

	templateCallback := func(isPublic, description string, extraTag string) string {
		tags := `["test", "terraform"]`
		if extraTag != "" {
			tags = fmt.Sprintf(`["test", "terraform", %q]`, extraTag)
		}

		return fmt.Sprintf(`
		  is_public   = %q
		  description = %q
		  tags        = %s
		`, isPublic, description, tags)
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader),
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccWorkflowTemplate(templateName, sourceConfigKind, templateCallback("0", "Initial description", "")),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_template.test", "template_name", templateName),
					resource.TestCheckResourceAttr("stackguardian_workflow_template.test", "source_config_kind", sourceConfigKind),
					resource.TestCheckResourceAttr("stackguardian_workflow_template.test", "is_public", "0"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template.test", "description", "Initial description"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template.test", "tags.0", "test"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template.test", "tags.1", "terraform"),
					resource.TestCheckResourceAttrSet("stackguardian_workflow_template.test", "id"),
					resource.TestCheckResourceAttrSet("stackguardian_workflow_template.test", "owner_org"),
				),
			},
			// Update and Read testing
			{
				Config: testAccWorkflowTemplate(templateName, sourceConfigKind, templateCallback("1", "Updated description", "updated")),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_template.test", "template_name", templateName),
					resource.TestCheckResourceAttr("stackguardian_workflow_template.test", "is_public", "1"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template.test", "description", "Updated description"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template.test", "tags.2", "updated"),
				),
			},
			// Delete testing automatically occurs
		},
	})
}

func TestAccWorkflowTemplate_WithRuntime(t *testing.T) {
	templateName := acctest.ResourceName("tf-provider-workflow-template-2")

	t.Cleanup(func() { deleteWorkflowTemplateFixture(templateName) })

	customHeader := http.Header{}
	customHeader.Set("x-sg-internal-auth-orgid", "sg-provider-test")

	// GIT_OTHER is the only source_config_dest_kind that can be public/authless per the
	// runtime_source is_private/auth validation — every other kind requires auth.
	templateCallback := func(ref, workingDir string) string {
		return fmt.Sprintf(`
		  runtime_source = {
			source_config_dest_kind = "GIT_OTHER"
			config = {
			  is_private                 = false
			  repo                       = "https://github.com/StackGuardian/tf-null-resource.git"
			  ref                        = %q
			  working_dir                = %q
			  include_sub_module         = true
			  git_core_auto_crlf         = true
			  git_sparse_checkout_config = "--no-cone infra/**"
			}
		  }
		`, ref, workingDir)
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader),
		Steps: []resource.TestStep{
			// Create and Read testing with runtime source
			{
				Config: testAccWorkflowTemplate(templateName, sourceConfigKind, templateCallback("main", "src")),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_template.test", "template_name", templateName),
					resource.TestCheckResourceAttr("stackguardian_workflow_template.test", "runtime_source.source_config_dest_kind", "GIT_OTHER"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template.test", "runtime_source.config.repo", "https://github.com/StackGuardian/tf-null-resource.git"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template.test", "runtime_source.config.is_private", "false"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template.test", "runtime_source.config.ref", "main"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template.test", "runtime_source.config.working_dir", "src"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template.test", "runtime_source.config.include_sub_module", "true"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template.test", "runtime_source.config.git_core_auto_crlf", "true"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template.test", "runtime_source.config.git_sparse_checkout_config", "--no-cone infra/**"),
				),
			},
			// Update the runtime source config (repo has RequiresReplace, so it is left unchanged)
			{
				Config: testAccWorkflowTemplate(templateName, sourceConfigKind, templateCallback("develop", "modules")),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_template.test", "runtime_source.config.ref", "develop"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template.test", "runtime_source.config.working_dir", "modules"),
				),
			},
			// Delete testing automatically occurs
		},
	})
}

func TestAccWorkflowTemplate_WithContextTagsAndSharedOrgs(t *testing.T) {
	templateName := "tf-provider-workflow-template-3"

	customHeader := http.Header{}
	customHeader.Set("x-sg-internal-auth-orgid", "sg-provider-test")

	templateCallback := func(ctxVal string) string {
		return fmt.Sprintf(`
		  context_tags = {
			env = %q
		  }
		
		  shared_orgs_list = ["sg-provider-test-shared-org"]
		`, ctxVal)
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowTemplate(templateName, sourceConfigKind, templateCallback("staging")),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_template.test", "template_name", templateName),
					resource.TestCheckResourceAttr("stackguardian_workflow_template.test", "context_tags.env", "staging"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template.test", "shared_orgs_list.0", "sg-provider-test-shared-org"),
				),
			},
			{
				Config: testAccWorkflowTemplate(templateName, sourceConfigKind, templateCallback("production")),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_template.test", "context_tags.env", "production"),
				),
			},
		},
	})
}

func TestAccWorkflowTemplate_WithVCSTriggers(t *testing.T) {
	templateName := "tf-provider-workflow-template-4"

	customHeader := http.Header{}
	customHeader.Set("x-sg-internal-auth-orgid", "sg-provider-test")

	templateCallback := func(enabled bool) string {
		return fmt.Sprintf(`
		  runtime_source = {
			source_config_dest_kind = "GITHUB_COM"
			config = {
			  is_private = true
			  auth       = "/integrations/tf-provider-test-connector"
			  repo       = "https://github.com/StackGuardian/tf-null-resource.git"
			}
		  }
		
		  vcs_triggers = {
			type = "GITHUB_COM"
		
			create_tag = {
			  create_revision = {
				enabled = %t
			  }
			}
		  }
		`, enabled)
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowTemplate(templateName, sourceConfigKind, templateCallback(true)),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_template.test", "template_name", templateName),
					resource.TestCheckResourceAttr("stackguardian_workflow_template.test", "vcs_triggers.type", "GITHUB_COM"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template.test", "vcs_triggers.create_tag.create_revision.enabled", "true"),
				),
			},
			{
				Config: testAccWorkflowTemplate(templateName, sourceConfigKind, templateCallback(false)),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_template.test", "vcs_triggers.create_tag.create_revision.enabled", "false"),
				),
			},
		},
	})
}
