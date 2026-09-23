package workflowtemplate_test

import (
	"fmt"
	"net/http"
	"regexp"
	"testing"

	"github.com/StackGuardian/terraform-provider-stackguardian/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

// TestAccWorkflowTemplate_GitCoreAutoCrlfPersistsAfterRemoval verifies that
// git_core_auto_crlf, once set explicitly at create, keeps its value after
// the attribute is removed from config on a later update — it is
// Optional+Computed with UseStateForUnknown() (schema.go), so an omitted
// value on Update carries forward whatever is already in state rather than
// resetting to false or erroring. ref also changes between the two steps, so
// this exercises a genuine Update, not a no-op plan.
func TestAccWorkflowTemplate_GitCoreAutoCrlfPersistsAfterRemoval(t *testing.T) {
	templateName := acctest.ResourceName("tf-provider-workflow-template-crlf-persist")

	t.Cleanup(func() { deleteWorkflowTemplateFixture(templateName) })

	customHeader := http.Header{}
	customHeader.Set("x-sg-internal-auth-orgid", "sg-provider-test")

	withCrlf := `
	  runtime_source = {
		source_config_dest_kind = "GIT_OTHER"
		config = {
		  is_private         = false
		  repo               = "https://github.com/StackGuardian/tf-null-resource.git"
		  ref                = "main"
		  git_core_auto_crlf = true
		}
	  }
	`

	withoutCrlf := `
	  runtime_source = {
		source_config_dest_kind = "GIT_OTHER"
		config = {
		  is_private = false
		  repo       = "https://github.com/StackGuardian/tf-null-resource.git"
		  ref        = "develop"
		}
	  }
	`

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowTemplate(templateName, sourceConfigKind, withCrlf),
				Check:  resource.TestCheckResourceAttr("stackguardian_workflow_template.test", "runtime_source.config.git_core_auto_crlf", "true"),
			},
			{
				// git_core_auto_crlf removed from config; ref changes too, so this is a
				// genuine update, not just a no-op plan.
				Config: testAccWorkflowTemplate(templateName, sourceConfigKind, withoutCrlf),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_template.test", "runtime_source.config.ref", "develop"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template.test", "runtime_source.config.git_core_auto_crlf", "true"),
				),
			},
		},
	})
}

// TestAccWorkflowTemplate_GitCoreAutoCrlfAddedThenRemoved verifies the same
// persistence contract when git_core_auto_crlf starts UNSET at create (so it
// resolves to whatever the server assigns), is then explicitly set on an
// update, and finally removed from config again — it must keep the
// explicitly-set value rather than reverting to the original server default
// or becoming unknown/erroring, since UseStateForUnknown() carries forward
// whatever is already in state. ref also changes on the final step: without
// that, the config would be byte-identical to step 1's except for the
// omitted attribute, and since git_core_auto_crlf is expected to plan as
// unchanged (carried forward from state), that step could show zero diff and
// Terraform might skip calling Update() entirely — proving nothing about the
// removal path this test exists to check.
func TestAccWorkflowTemplate_GitCoreAutoCrlfAddedThenRemoved(t *testing.T) {
	templateName := acctest.ResourceName("tf-provider-workflow-template-crlf-addrm")

	t.Cleanup(func() { deleteWorkflowTemplateFixture(templateName) })

	customHeader := http.Header{}
	customHeader.Set("x-sg-internal-auth-orgid", "sg-provider-test")

	unset := func(ref string) string {
		return fmt.Sprintf(`
		  runtime_source = {
			source_config_dest_kind = "GIT_OTHER"
			config = {
			  is_private = false
			  repo       = "https://github.com/StackGuardian/tf-null-resource.git"
			  ref        = %q
			}
		  }
		`, ref)
	}

	withCrlfTrue := `
	  runtime_source = {
		source_config_dest_kind = "GIT_OTHER"
		config = {
		  is_private         = false
		  repo               = "https://github.com/StackGuardian/tf-null-resource.git"
		  ref                = "main"
		  git_core_auto_crlf = true
		}
	  }
	`

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader),
		Steps: []resource.TestStep{
			{
				// Left unset at create — must resolve to a real server-assigned
				// value, not error (regression guard for the Unknown-value bug
				// pattern documented in ToAPIModel).
				Config: testAccWorkflowTemplate(templateName, sourceConfigKind, unset("main")),
				Check:  resource.TestCheckResourceAttrSet("stackguardian_workflow_template.test", "runtime_source.config.git_core_auto_crlf"),
			},
			{
				// Explicitly added on update — git_core_auto_crlf itself changes
				// here, so this step is already a genuine update.
				Config: testAccWorkflowTemplate(templateName, sourceConfigKind, withCrlfTrue),
				Check:  resource.TestCheckResourceAttr("stackguardian_workflow_template.test", "runtime_source.config.git_core_auto_crlf", "true"),
			},
			{
				// Removed again, with ref also changed so this is a genuine
				// update — must keep the explicitly-set value, not revert to the
				// original server default or error.
				Config: testAccWorkflowTemplate(templateName, sourceConfigKind, unset("develop")),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_template.test", "runtime_source.config.ref", "develop"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template.test", "runtime_source.config.git_core_auto_crlf", "true"),
				),
			},
		},
	})
}

// TestAccWorkflowTemplate_RuntimeSourceAuthUpdates verifies that changing
// runtime_source.config.auth on an update actually applies. auth is plain
// Optional (no Computed/UseStateForUnknown), so a value change here should
// be a straightforward update — but ToUpdateAPIModel's
// RuntimeSourceConfigUpdate struct literal (model.go) never sets Auth at
// all, unlike ToAPIModel's Create path, which does. If that gap is real,
// step 2's check fails because the API never actually received the new
// value.
func TestAccWorkflowTemplate_RuntimeSourceAuthUpdates(t *testing.T) {
	templateName := acctest.ResourceName("tf-provider-workflow-template-auth-upd")

	t.Cleanup(func() { deleteWorkflowTemplateFixture(templateName) })

	customHeader := http.Header{}
	customHeader.Set("x-sg-internal-auth-orgid", "sg-provider-test")

	config := func(auth string) string {
		return fmt.Sprintf(`
		  runtime_source = {
			source_config_dest_kind = "GITHUB_COM"
			config = {
			  is_private = true
			  auth       = %q
			  repo       = "https://github.com/StackGuardian/tf-null-resource.git"
			}
		  }
		`, auth)
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowTemplate(templateName, sourceConfigKind, config("/integrations/tf-provider-test-connector")),
				Check:  resource.TestCheckResourceAttr("stackguardian_workflow_template.test", "runtime_source.config.auth", "/integrations/tf-provider-test-connector"),
			},
			{
				// A different (fabricated, not required to resolve to a real
				// integration — see ValidateRuntimeSourceAuth, which only checks the
				// "/integration" prefix) auth reference, to prove the change reaches
				// the API rather than silently being dropped.
				Config: testAccWorkflowTemplate(templateName, sourceConfigKind, config("/integrations/tf-provider-test-connector-2")),
				Check:  resource.TestCheckResourceAttr("stackguardian_workflow_template.test", "runtime_source.config.auth", "/integrations/tf-provider-test-connector-2"),
			},
		},
	})
}

// TestAccWorkflowTemplate_IsPrivatePersistsAfterRemoval mirrors
// TestAccWorkflowTemplate_GitCoreAutoCrlfPersistsAfterRemoval for
// is_private: set explicitly at create, then removed from config on a later
// update (which also changes ref, so it's a genuine update) — is_private is
// Optional+Computed with UseStateForUnknown() (schema.go), so it must keep
// its value rather than resetting.
func TestAccWorkflowTemplate_IsPrivatePersistsAfterRemoval(t *testing.T) {
	templateName := acctest.ResourceName("tf-provider-workflow-template-ispriv-persist")

	t.Cleanup(func() { deleteWorkflowTemplateFixture(templateName) })

	customHeader := http.Header{}
	customHeader.Set("x-sg-internal-auth-orgid", "sg-provider-test")

	withIsPrivate := `
	  runtime_source = {
		source_config_dest_kind = "GIT_OTHER"
		config = {
		  is_private = true
		  auth       = "/secrets/tf-provider-test-secret"
		  repo       = "https://github.com/StackGuardian/tf-null-resource.git"
		  ref        = "main"
		}
	  }
	`

	withoutIsPrivate := `
	  runtime_source = {
		source_config_dest_kind = "GIT_OTHER"
		config = {
		  auth = "/secrets/tf-provider-test-secret"
		  repo = "https://github.com/StackGuardian/tf-null-resource.git"
		  ref  = "develop"
		}
	  }
	`

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowTemplate(templateName, sourceConfigKind, withIsPrivate),
				Check:  resource.TestCheckResourceAttr("stackguardian_workflow_template.test", "runtime_source.config.is_private", "true"),
			},
			{
				// is_private removed from config; ref changes too, so this is a
				// genuine update, not just a no-op plan.
				Config: testAccWorkflowTemplate(templateName, sourceConfigKind, withoutIsPrivate),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_template.test", "runtime_source.config.ref", "develop"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template.test", "runtime_source.config.is_private", "true"),
				),
			},
		},
	})
}

// TestAccWorkflowTemplate_RepoRejectedOnChange verifies that changing
// runtime_source.config.repo on an existing template is rejected at plan
// time (ModifyPlan, resource.go) rather than silently updating in place or
// destroying and recreating the template.
func TestAccWorkflowTemplate_RepoRejectedOnChange(t *testing.T) {
	templateName := acctest.ResourceName("tf-provider-workflow-template-repo-rejected")

	t.Cleanup(func() { deleteWorkflowTemplateFixture(templateName) })

	customHeader := http.Header{}
	customHeader.Set("x-sg-internal-auth-orgid", "sg-provider-test")

	config := func(repo string) string {
		return fmt.Sprintf(`
		  runtime_source = {
			source_config_dest_kind = "GIT_OTHER"
			config = {
			  is_private = false
			  repo       = %q
			}
		  }
		`, repo)
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowTemplate(templateName, sourceConfigKind, config("https://github.com/StackGuardian/tf-null-resource.git")),
				Check:  resource.TestCheckResourceAttr("stackguardian_workflow_template.test", "runtime_source.config.repo", "https://github.com/StackGuardian/tf-null-resource.git"),
			},
			{
				Config:      testAccWorkflowTemplate(templateName, sourceConfigKind, config("https://github.com/StackGuardian/terraform-provider-stackguardian.git")),
				ExpectError: regexp.MustCompile("runtime_source.config.repo cannot be changed"),
			},
		},
	})
}

// TestAccWorkflowTemplate_VcsTriggersPersistsAfterRemoval checks whether
// vcs_triggers, once set, survives being removed from config on a later
// update (which also changes description, so it's a genuine update).
// vcs_triggers is Optional+Computed but — unlike runtime_source, which has
// objectplanmodifier.UseStateForUnknown() — has NO plan modifier at all
// (schema.go). Per the UseStateForUnknown() gotcha documented in this
// codebase, that asymmetry means this step may not behave like
// git_core_auto_crlf/is_private above: expect either a value that resets
// instead of persisting, or a "Provider produced inconsistent result after
// apply" error, rather than a clean pass.
func TestAccWorkflowTemplate_VcsTriggersPersistsAfterRemoval(t *testing.T) {
	templateName := acctest.ResourceName("tf-provider-workflow-template-vcst-persist")

	t.Cleanup(func() { deleteWorkflowTemplateFixture(templateName) })

	customHeader := http.Header{}
	customHeader.Set("x-sg-internal-auth-orgid", "sg-provider-test")

	withVcsTriggers := `
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
			enabled = true
		  }
		}
	  }

	  description = "with vcs_triggers"
	`

	withoutVcsTriggers := `
	  runtime_source = {
		source_config_dest_kind = "GITHUB_COM"
		config = {
		  is_private = true
		  auth       = "/integrations/tf-provider-test-connector"
		  repo       = "https://github.com/StackGuardian/tf-null-resource.git"
		}
	  }

	  description = "vcs_triggers omitted"
	`

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowTemplate(templateName, sourceConfigKind, withVcsTriggers),
				Check:  resource.TestCheckResourceAttr("stackguardian_workflow_template.test", "vcs_triggers.type", "GITHUB_COM"),
			},
			{
				Config: testAccWorkflowTemplate(templateName, sourceConfigKind, withoutVcsTriggers),
				Check:  resource.TestCheckResourceAttr("stackguardian_workflow_template.test", "vcs_triggers.type", "GITHUB_COM"),
			},
		},
	})
}
