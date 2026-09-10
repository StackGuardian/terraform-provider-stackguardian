package workflowtemplate_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/StackGuardian/terraform-provider-stackguardian/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

var sourceConfigKind = "TERRAFORM"

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
	templateName := "tf-provider-workflow-template-1"

	customHeader := http.Header{}
	customHeader.Set("x-sg-internal-auth-orgid", "sg-provider-test")

	config := func(isPublic, description string, extraTag string) string {
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
				Config: testAccWorkflowTemplate(templateName, sourceConfigKind, config("0", "Initial description", "")),
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
				Config: testAccWorkflowTemplate(templateName, sourceConfigKind, config("1", "Updated description", "updated")),
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
	templateName := "tf-provider-workflow-template-2"

	customHeader := http.Header{}
	customHeader.Set("x-sg-internal-auth-orgid", "sg-provider-test")

	config := func(ref, workingDir string) string {
		return fmt.Sprintf(`
  runtime_source = {
    source_config_dest_kind = "GITHUB_COM"
    config = {
      is_private                 = false
      repo                       = "https://github.com/taherkk/taher-null-resource.git"
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
				Config: testAccWorkflowTemplate(templateName, sourceConfigKind, config("main", "src")),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_template.test", "template_name", templateName),
					resource.TestCheckResourceAttr("stackguardian_workflow_template.test", "runtime_source.source_config_dest_kind", "GITHUB_COM"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template.test", "runtime_source.config.repo", "https://github.com/taherkk/taher-null-resource.git"),
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
				Config: testAccWorkflowTemplate(templateName, sourceConfigKind, config("develop", "modules")),
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

	config := func(ctxVal string) string {
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
				Config: testAccWorkflowTemplate(templateName, sourceConfigKind, config("staging")),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_template.test", "template_name", templateName),
					resource.TestCheckResourceAttr("stackguardian_workflow_template.test", "context_tags.env", "staging"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template.test", "shared_orgs_list.0", "sg-provider-test-shared-org"),
				),
			},
			{
				Config: testAccWorkflowTemplate(templateName, sourceConfigKind, config("production")),
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

	config := func(enabled bool) string {
		return fmt.Sprintf(`
  runtime_source = {
    source_config_dest_kind = "GITHUB_COM"
    config = {
      is_private = true
      auth       = "/integrations/tf-provider-test-connector"
      repo       = "https://github.com/taherkk/taher-null-resource.git"
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
				Config: testAccWorkflowTemplate(templateName, sourceConfigKind, config(true)),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_template.test", "template_name", templateName),
					resource.TestCheckResourceAttr("stackguardian_workflow_template.test", "vcs_triggers.type", "GITHUB_COM"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template.test", "vcs_triggers.create_tag.create_revision.enabled", "true"),
				),
			},
			{
				Config: testAccWorkflowTemplate(templateName, sourceConfigKind, config(false)),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_template.test", "vcs_triggers.create_tag.create_revision.enabled", "false"),
				),
			},
		},
	})
}
