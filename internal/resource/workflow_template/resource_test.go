package workflowtemplate_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/StackGuardian/terraform-provider-stackguardian/internal/acctest"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/config"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/constants"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

var sourceConfigKind = "TERRAFORM"

// deleteWorkflowTemplateFixture is a safety-net cleanup: Terraform's own destroy step tears
// the template down in the normal case. This exists so a test that fails before reaching
// destroy (e.g. a failed assertion) doesn't leave it behind. Errors are ignored — the
// resource may already be gone. id is the template's id, which defaults to its
// template_name when the config doesn't set an explicit id.
func deleteWorkflowTemplateFixture(id string) {
	client := acctest.SGClient()
	client.WorkflowTemplates.DeleteWorkflowTemplate(context.TODO(), config.Get().OrgName, id)
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

// TestAccWorkflowTemplate_UpdateWithIsPublicUnchanged confirms updating an unrelated field
// (description) leaves is_public alone when the config keeps it at "0" across both steps.
// is_public is Optional+Computed with UseStateForUnknown(), so this guards against a
// nested-attribute-style regression where a field like this drifts or resets on an update
// that doesn't touch it.
func TestAccWorkflowTemplate_UpdateWithIsPublicUnchanged(t *testing.T) {
	templateName := acctest.ResourceName("tf-provider-workflow-template-5")

	t.Cleanup(func() { deleteWorkflowTemplateFixture(templateName) })

	customHeader := http.Header{}
	customHeader.Set("x-sg-internal-auth-orgid", "sg-provider-test")

	templateCallback := func(description string) string {
		return fmt.Sprintf(`
		  is_public   = "0"
		  description = %q
		`, description)
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
				Config: testAccWorkflowTemplate(templateName, sourceConfigKind, templateCallback("Initial description")),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_template.test", "template_name", templateName),
					resource.TestCheckResourceAttr("stackguardian_workflow_template.test", "is_public", "0"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template.test", "description", "Initial description"),
				),
			},
			// Update description only; is_public stays "0" in config on both steps.
			{
				Config: testAccWorkflowTemplate(templateName, sourceConfigKind, templateCallback("Updated description")),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_template.test", "template_name", templateName),
					resource.TestCheckResourceAttr("stackguardian_workflow_template.test", "is_public", "0"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template.test", "description", "Updated description"),
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
					resource.TestCheckResourceAttr("stackguardian_workflow_template.test", "runtime_source.source_config_dest_kind", constants.GitOther),
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

// TestAccWorkflowTemplate_IdRequiresReplace verifies that a
// practitioner-supplied id forces a destroy-and-recreate when changed — id
// has stringplanmodifier.RequiresReplace() (schema.go) alongside
// UseStateForUnknown(). template_name stays fixed across both steps so only
// id itself differs, isolating this from template_name's own behavior.
func TestAccWorkflowTemplate_IdRequiresReplace(t *testing.T) {
	name1 := acctest.ResourceName("tf-provider-workflow-template-id-replace-a")
	name2 := acctest.ResourceName("tf-provider-workflow-template-id-replace-b")

	t.Cleanup(func() { deleteWorkflowTemplateFixture(name1) })
	t.Cleanup(func() { deleteWorkflowTemplateFixture(name2) })

	customHeader := http.Header{}
	customHeader.Set("x-sg-internal-auth-orgid", "sg-provider-test")

	config := func(id string) string {
		return fmt.Sprintf(`id = %q`, id)
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowTemplate(name1, sourceConfigKind, config(name1)),
				Check:  resource.TestCheckResourceAttr("stackguardian_workflow_template.test", "id", name1),
			},
			{
				Config: testAccWorkflowTemplate(name1, sourceConfigKind, config(name2)),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("stackguardian_workflow_template.test", plancheck.ResourceActionReplace),
					},
				},
				Check: resource.TestCheckResourceAttr("stackguardian_workflow_template.test", "id", name2),
			},
		},
	})
}
