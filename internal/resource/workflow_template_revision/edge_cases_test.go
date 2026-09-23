package workflowtemplaterevision_test

import (
	"fmt"
	"net/http"
	"regexp"
	"testing"

	"github.com/StackGuardian/terraform-provider-stackguardian/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

// TestAccWorkflowTemplateRevision_GitCoreAutoCrlfPersistsAfterRemoval mirrors
// workflow_template's TestAccWorkflowTemplate_GitCoreAutoCrlfPersistsAfterRemoval:
// git_core_auto_crlf, once set explicitly at create, keeps its value after
// the attribute is removed from config on a later update — it is
// Optional+Computed with UseStateForUnknown() (workflowtemplate.WorkflowTemplateRuntimeSourceConfig(),
// reused by this resource's schema.go), so an omitted value on Update
// carries forward whatever is already in state rather than resetting to
// false or erroring. ref also changes between the two steps, so this
// exercises a genuine Update, not a no-op plan.
func TestAccWorkflowTemplateRevision_GitCoreAutoCrlfPersistsAfterRemoval(t *testing.T) {
	templateID := acctest.ResourceName("tf-provider-wftr-crlf-persist")
	alias := "revision-crlf-persist"

	registerWorkflowTemplateCleanup(t, templateID, 1)

	err := createWorkflowTemplateFixture(templateID, "TERRAFORM")
	if err != nil {
		t.Fatal(err)
	}

	customHeader := http.Header{}
	customHeader.Set("x-sg-internal-auth-orgid", "sg-provider-test")

	withCrlf := fmt.Sprintf(`
	  alias = %q

	  runtime_source = {
		source_config_dest_kind = "GIT_OTHER"
		config = {
		  is_private         = false
		  repo               = "https://github.com/StackGuardian/tf-null-resource.git"
		  ref                = "main"
		  git_core_auto_crlf = true
		}
	  }
	`, alias)

	withoutCrlf := fmt.Sprintf(`
	  alias = %q

	  runtime_source = {
		source_config_dest_kind = "GIT_OTHER"
		config = {
		  is_private = false
		  repo       = "https://github.com/StackGuardian/tf-null-resource.git"
		  ref        = "develop"
		}
	  }
	`, alias)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowTemplateRevision(templateID, "TERRAFORM", 500, 1024, withCrlf),
				Check:  resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "runtime_source.config.git_core_auto_crlf", "true"),
			},
			{
				// git_core_auto_crlf removed from config; ref changes too, so this is a
				// genuine update, not just a no-op plan.
				Config: testAccWorkflowTemplateRevision(templateID, "TERRAFORM", 500, 1024, withoutCrlf),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "runtime_source.config.ref", "develop"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "runtime_source.config.git_core_auto_crlf", "true"),
				),
			},
		},
	})
}

// TestAccWorkflowTemplateRevision_GitCoreAutoCrlfAddedThenRemoved mirrors
// workflow_template's TestAccWorkflowTemplate_GitCoreAutoCrlfAddedThenRemoved:
// starts UNSET at create (so it resolves to whatever the server assigns), is
// then explicitly set on an update, and finally removed from config again —
// it must keep the explicitly-set value rather than reverting to the
// original server default or becoming unknown/erroring. ref also changes on
// the final step so a real diff forces Update() to actually run.
func TestAccWorkflowTemplateRevision_GitCoreAutoCrlfAddedThenRemoved(t *testing.T) {
	templateID := acctest.ResourceName("tf-provider-wftr-crlf-addrm")
	alias := "revision-crlf-addrm"

	registerWorkflowTemplateCleanup(t, templateID, 1)

	err := createWorkflowTemplateFixture(templateID, "TERRAFORM")
	if err != nil {
		t.Fatal(err)
	}

	customHeader := http.Header{}
	customHeader.Set("x-sg-internal-auth-orgid", "sg-provider-test")

	unset := func(ref string) string {
		return fmt.Sprintf(`
		  alias = %q

		  runtime_source = {
			source_config_dest_kind = "GIT_OTHER"
			config = {
			  is_private = false
			  repo       = "https://github.com/StackGuardian/tf-null-resource.git"
			  ref        = %q
			}
		  }
		`, alias, ref)
	}

	withCrlfTrue := fmt.Sprintf(`
	  alias = %q

	  runtime_source = {
		source_config_dest_kind = "GIT_OTHER"
		config = {
		  is_private         = false
		  repo               = "https://github.com/StackGuardian/tf-null-resource.git"
		  ref                = "main"
		  git_core_auto_crlf = true
		}
	  }
	`, alias)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader),
		Steps: []resource.TestStep{
			{
				// Left unset at create — must resolve to a real server-assigned
				// value, not error.
				Config: testAccWorkflowTemplateRevision(templateID, "TERRAFORM", 500, 1024, unset("main")),
				Check:  resource.TestCheckResourceAttrSet("stackguardian_workflow_template_revision.test", "runtime_source.config.git_core_auto_crlf"),
			},
			{
				// Explicitly added on update — git_core_auto_crlf itself changes
				// here, so this step is already a genuine update.
				Config: testAccWorkflowTemplateRevision(templateID, "TERRAFORM", 500, 1024, withCrlfTrue),
				Check:  resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "runtime_source.config.git_core_auto_crlf", "true"),
			},
			{
				// Removed again, with ref also changed so this is a genuine
				// update — must keep the explicitly-set value, not revert to the
				// original server default or error.
				Config: testAccWorkflowTemplateRevision(templateID, "TERRAFORM", 500, 1024, unset("develop")),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "runtime_source.config.ref", "develop"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "runtime_source.config.git_core_auto_crlf", "true"),
				),
			},
		},
	})
}

// TestAccWorkflowTemplateRevision_RuntimeSourceAuthUpdates verifies that
// changing runtime_source.config.auth on an update actually applies. Both
// this resource and workflow_template route their ToUpdateAPIModel through
// the shared workflowtemplate.ConvertRuntimeSourceToUpdateAPI (model.go),
// which sets Auth on the update path.
func TestAccWorkflowTemplateRevision_RuntimeSourceAuthUpdates(t *testing.T) {
	templateID := acctest.ResourceName("tf-provider-wftr-auth-upd")
	alias := "revision-auth-upd"

	registerWorkflowTemplateCleanup(t, templateID, 1)

	err := createWorkflowTemplateFixture(templateID, "TERRAFORM")
	if err != nil {
		t.Fatal(err)
	}

	customHeader := http.Header{}
	customHeader.Set("x-sg-internal-auth-orgid", "sg-provider-test")

	config := func(auth string) string {
		return fmt.Sprintf(`
		  alias = %q

		  runtime_source = {
			source_config_dest_kind = "GITHUB_COM"
			config = {
			  is_private = true
			  auth       = %q
			  repo       = "https://github.com/StackGuardian/tf-null-resource.git"
			}
		  }
		`, alias, auth)
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowTemplateRevision(templateID, "TERRAFORM", 500, 1024, config("/integrations/tf-provider-test-connector")),
				Check:  resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "runtime_source.config.auth", "/integrations/tf-provider-test-connector"),
			},
			{
				// A different (fabricated, prefix-only validated) auth reference,
				// to prove the change reaches the API.
				Config: testAccWorkflowTemplateRevision(templateID, "TERRAFORM", 500, 1024, config("/integrations/tf-provider-test-connector-2")),
				Check:  resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "runtime_source.config.auth", "/integrations/tf-provider-test-connector-2"),
			},
		},
	})
}

// TestAccWorkflowTemplateRevision_IsPrivatePersistsAfterRemoval mirrors
// TestAccWorkflowTemplateRevision_GitCoreAutoCrlfPersistsAfterRemoval for
// is_private. Both this resource's and workflow_template's Update paths
// route through the shared, guarded workflowtemplate.RuntimeSourceModel.
// ToAPIModel (via ConvertRuntimeSourceToUpdateAPI, model.go), so
// IsPrivate/GitCoreAutoCRLF/Ref are all guarded against null/unknown.
func TestAccWorkflowTemplateRevision_IsPrivatePersistsAfterRemoval(t *testing.T) {
	templateID := acctest.ResourceName("tf-provider-wftr-ispriv-persist")
	alias := "revision-ispriv-persist"

	registerWorkflowTemplateCleanup(t, templateID, 1)

	err := createWorkflowTemplateFixture(templateID, "TERRAFORM")
	if err != nil {
		t.Fatal(err)
	}

	customHeader := http.Header{}
	customHeader.Set("x-sg-internal-auth-orgid", "sg-provider-test")

	withIsPrivate := fmt.Sprintf(`
	  alias = %q

	  runtime_source = {
		source_config_dest_kind = "GIT_OTHER"
		config = {
		  is_private = true
		  auth       = "/secrets/tf-provider-test-secret"
		  repo       = "https://github.com/StackGuardian/tf-null-resource.git"
		  ref        = "main"
		}
	  }
	`, alias)

	withoutIsPrivate := fmt.Sprintf(`
	  alias = %q

	  runtime_source = {
		source_config_dest_kind = "GIT_OTHER"
		config = {
		  auth = "/secrets/tf-provider-test-secret"
		  repo = "https://github.com/StackGuardian/tf-null-resource.git"
		  ref  = "develop"
		}
	  }
	`, alias)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowTemplateRevision(templateID, "TERRAFORM", 500, 1024, withIsPrivate),
				Check:  resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "runtime_source.config.is_private", "true"),
			},
			{
				// is_private removed from config; ref changes too, so this is a
				// genuine update, not just a no-op plan.
				Config: testAccWorkflowTemplateRevision(templateID, "TERRAFORM", 500, 1024, withoutIsPrivate),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "runtime_source.config.ref", "develop"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "runtime_source.config.is_private", "true"),
				),
			},
		},
	})
}

// TestAccWorkflowTemplateRevision_TerraformConfigFieldsUpdate verifies that
// terraform_config's Optional+Computed fields other than terraform_version
// (the only one TestAccWorkflowTemplateRevision_WithTerraformConfig ever
// changes) actually apply on a genuine update, not just at create.
func TestAccWorkflowTemplateRevision_TerraformConfigFieldsUpdate(t *testing.T) {
	templateID := acctest.ResourceName("tf-provider-wftr-tfconfig-upd")
	alias := "revision-tfconfig-upd"

	registerWorkflowTemplateCleanup(t, templateID, 1)

	err := createWorkflowTemplateFixture(templateID, "TERRAFORM")
	if err != nil {
		t.Fatal(err)
	}

	customHeader := http.Header{}
	customHeader.Set("x-sg-internal-auth-orgid", "sg-provider-test")

	config := func(driftCron, planOptions string, timeout int) string {
		return fmt.Sprintf(`
		  alias = %q

		  terraform_config = {
		    terraform_version      = "1.5.7"
		    drift_check            = true
		    drift_cron             = %q
		    managed_terraform_state = true
		    terraform_plan_options = %q
		    timeout                = %d
		  }
		`, alias, driftCron, planOptions, timeout)
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowTemplateRevision(templateID, "TERRAFORM", 500, 1024, config("0 0 * * *", "-lock=false", 3600)),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "terraform_config.drift_cron", "0 0 * * *"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "terraform_config.terraform_plan_options", "-lock=false"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "terraform_config.timeout", "3600"),
				),
			},
			{
				// Every value changes on this update — must actually apply, not
				// just round-trip the create-time values.
				Config: testAccWorkflowTemplateRevision(templateID, "TERRAFORM", 500, 1024, config("0 6 * * *", "-lock=true", 7200)),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "terraform_config.drift_cron", "0 6 * * *"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "terraform_config.terraform_plan_options", "-lock=true"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "terraform_config.timeout", "7200"),
				),
			},
		},
	})
}

// TestAccWorkflowTemplateRevision_DeploymentPlatformConfigUpdate adds an
// update step to TestAccWorkflowTemplateRevision_WithDeploymentPlatformConfig's
// scenario, which is create-only today — deployment_platform_config is
// top-level Optional+Computed with UseStateForUnknown(), but no test ever
// changes its value after create.
func TestAccWorkflowTemplateRevision_DeploymentPlatformConfigUpdate(t *testing.T) {
	templateID := acctest.ResourceName("tf-provider-wftr-dpc-upd")
	alias := "revision-dpc-upd"

	registerWorkflowTemplateCleanup(t, templateID, 1)

	err := createWorkflowTemplateFixture(templateID, "TERRAFORM")
	if err != nil {
		t.Fatal(err)
	}

	customHeader := http.Header{}
	customHeader.Set("x-sg-internal-auth-orgid", "sg-provider-test")

	config := func(profileName string) string {
		return fmt.Sprintf(`
		  alias     = %q
		  is_public = "0"

		  deployment_platform_config = [{
		    kind = "AWS_RBAC"
		    config = {
		      integration_id = "/integrations/test-integration"
		      profile_name   = %q
		    }
		  }]
		`, alias, profileName)
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowTemplateRevision(templateID, "TERRAFORM", 500, 1024, config("test-profile")),
				Check:  resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "deployment_platform_config.0.config.profile_name", "test-profile"),
			},
			{
				Config: testAccWorkflowTemplateRevision(templateID, "TERRAFORM", 500, 1024, config("updated-profile")),
				Check:  resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "deployment_platform_config.0.config.profile_name", "updated-profile"),
			},
		},
	})
}

// TestAccWorkflowTemplateRevision_UserJobCpuMemoryUpdate verifies user_job_cpu
// and user_job_memory (both Required, no plan modifier) actually apply on a
// genuine in-place update. The only existing coverage that changes these
// values is published_test.go, where the change is expected to be
// REJECTED (published-revision restriction) — this is the first test to
// confirm they update successfully on a non-published revision.
func TestAccWorkflowTemplateRevision_UserJobCpuMemoryUpdate(t *testing.T) {
	templateID := acctest.ResourceName("tf-provider-wftr-cpumem-upd")
	alias := "revision-cpumem-upd"

	registerWorkflowTemplateCleanup(t, templateID, 1)

	err := createWorkflowTemplateFixture(templateID, "TERRAFORM")
	if err != nil {
		t.Fatal(err)
	}

	customHeader := http.Header{}
	customHeader.Set("x-sg-internal-auth-orgid", "sg-provider-test")

	config := fmt.Sprintf(`alias = %q`, alias)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowTemplateRevision(templateID, "TERRAFORM", 500, 1024, config),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "user_job_cpu", "500"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "user_job_memory", "1024"),
				),
			},
			{
				Config: testAccWorkflowTemplateRevision(templateID, "TERRAFORM", 1000, 2048, config),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "user_job_cpu", "1000"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "user_job_memory", "2048"),
				),
			},
		},
	})
}

// TestAccWorkflowTemplateRevision_SourceConfigKindRejectedOnChange verifies
// that changing source_config_kind on an existing revision is rejected at
// plan time (ModifyPlan, resource.go) rather than silently updating in place
// or destroying and recreating the revision — neither is appropriate here:
// the API has no endpoint to change it in place, and replacing the revision
// would change its identity (a new revision number), breaking anything
// pinned to the old one. Uses OPENTOFU as the target since, like TERRAFORM,
// it disallows wf_steps_config — this config doesn't set any, so the two
// kinds are otherwise unconstrained.
func TestAccWorkflowTemplateRevision_SourceConfigKindRejectedOnChange(t *testing.T) {
	templateID := acctest.ResourceName("tf-provider-wftr-kind-rejected")
	alias := "revision-kind-rejected"

	registerWorkflowTemplateCleanup(t, templateID, 1)

	err := createWorkflowTemplateFixture(templateID, "TERRAFORM")
	if err != nil {
		t.Fatal(err)
	}

	customHeader := http.Header{}
	customHeader.Set("x-sg-internal-auth-orgid", "sg-provider-test")

	config := fmt.Sprintf(`alias = %q`, alias)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowTemplateRevision(templateID, "TERRAFORM", 500, 1024, config),
				Check:  resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "source_config_kind", "TERRAFORM"),
			},
			{
				Config:      testAccWorkflowTemplateRevision(templateID, "OPENTOFU", 500, 1024, config),
				ExpectError: regexp.MustCompile("source_config_kind cannot be changed"),
			},
		},
	})
}

// TestAccWorkflowTemplateRevision_NumberOfApprovalsRequiredUpdate verifies
// number_of_approvals_required (Optional+Computed, UseStateForUnknown())
// applies on a genuine update — TestAccWorkflowTemplateRevision_WithConfig
// only ever sets it once at create.
func TestAccWorkflowTemplateRevision_NumberOfApprovalsRequiredUpdate(t *testing.T) {
	templateID := acctest.ResourceName("tf-provider-wftr-approvals-upd")
	alias := "revision-approvals-upd"

	registerWorkflowTemplateCleanup(t, templateID, 1)

	err := createWorkflowTemplateFixture(templateID, "TERRAFORM")
	if err != nil {
		t.Fatal(err)
	}

	customHeader := http.Header{}
	customHeader.Set("x-sg-internal-auth-orgid", "sg-provider-test")

	config := func(n int) string {
		return fmt.Sprintf(`
		  alias                        = %q
		  number_of_approvals_required = %d
		`, alias, n)
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowTemplateRevision(templateID, "TERRAFORM", 500, 1024, config(1)),
				Check:  resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "number_of_approvals_required", "1"),
			},
			{
				Config: testAccWorkflowTemplateRevision(templateID, "TERRAFORM", 500, 1024, config(3)),
				Check:  resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "number_of_approvals_required", "3"),
			},
		},
	})
}

// TestAccWorkflowTemplateRevision_RepoRejectedOnChange verifies that
// changing runtime_source.config.repo on an existing revision is rejected
// at plan time (ModifyPlan, resource.go) rather than silently updating in
// place or destroying and recreating the revision — this holds regardless
// of the revision's publish status, unlike the published-revision
// restriction covered in published_test.go.
func TestAccWorkflowTemplateRevision_RepoRejectedOnChange(t *testing.T) {
	templateID := acctest.ResourceName("tf-provider-wftr-repo-rejected")
	alias := "revision-repo-rejected"

	registerWorkflowTemplateCleanup(t, templateID, 1)

	err := createWorkflowTemplateFixture(templateID, "TERRAFORM")
	if err != nil {
		t.Fatal(err)
	}

	customHeader := http.Header{}
	customHeader.Set("x-sg-internal-auth-orgid", "sg-provider-test")

	config := func(repo string) string {
		return fmt.Sprintf(`
		  alias = %q

		  runtime_source = {
			source_config_dest_kind = "GIT_OTHER"
			config = {
			  is_private = false
			  repo       = %q
			}
		  }
		`, alias, repo)
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowTemplateRevision(templateID, "TERRAFORM", 500, 1024, config("https://github.com/StackGuardian/tf-null-resource.git")),
				Check:  resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "runtime_source.config.repo", "https://github.com/StackGuardian/tf-null-resource.git"),
			},
			{
				Config:      testAccWorkflowTemplateRevision(templateID, "TERRAFORM", 500, 1024, config("https://github.com/StackGuardian/terraform-provider-stackguardian.git")),
				ExpectError: regexp.MustCompile("runtime_source.config.repo cannot be changed"),
			},
		},
	})
}
