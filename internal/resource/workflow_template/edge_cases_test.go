package workflowtemplate_test

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"testing"

	sgsdkgo "github.com/StackGuardian/sg-sdk-go"
	"github.com/StackGuardian/sg-sdk-go/workflowtemplates"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/acctest"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/config"
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

// TestAccWorkflowTemplate_EmptyStringListsRoundTrip verifies that an explicitly empty
// tags or shared_orgs_list is stored and read back as an empty list, not null, on both
// create and update. `field = []` plans as a known empty list, so reading it back as
// null fails with "Provider produced inconsistent result after apply".
func TestAccWorkflowTemplate_EmptyStringListsRoundTrip(t *testing.T) {
	cases := []struct {
		name      string
		field     string
		populated string // non-empty value set before the update to []
		onUpdate  bool   // false: [] at create; true: populated at create, [] on update
	}{
		{name: "tags_create", field: "tags"},
		{name: "tags_update", field: "tags", populated: `["tf-provider-test"]`, onUpdate: true},
		{name: "shared_orgs_list_create", field: "shared_orgs_list"},
		{name: "shared_orgs_list_update", field: "shared_orgs_list", populated: `["sg-provider-test-shared-org"]`, onUpdate: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			templateName := acctest.ResourceName("tf-provider-workflow-template-empty-" + tc.name)

			t.Cleanup(func() { deleteWorkflowTemplateFixture(templateName) })

			customHeader := http.Header{}
			customHeader.Set("x-sg-internal-auth-orgid", "sg-provider-test")

			emptyStep := resource.TestStep{
				Config: testAccWorkflowTemplate(templateName, sourceConfigKind, fmt.Sprintf("%s = []", tc.field)),
				Check:  resource.TestCheckResourceAttr("stackguardian_workflow_template.test", tc.field+".#", "0"),
			}

			steps := []resource.TestStep{emptyStep}
			if tc.onUpdate {
				steps = []resource.TestStep{
					{
						Config: testAccWorkflowTemplate(templateName, sourceConfigKind, fmt.Sprintf("%s = %s", tc.field, tc.populated)),
						Check:  resource.TestCheckResourceAttr("stackguardian_workflow_template.test", tc.field+".#", "1"),
					},
					emptyStep,
				}
			}

			resource.Test(t, resource.TestCase{
				PreCheck: func() { acctest.TestAccPreCheck(t) },
				TerraformVersionChecks: []tfversion.TerraformVersionCheck{
					tfversion.SkipBelow(tfversion.Version1_1_0),
				},
				ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader),
				Steps:                    steps,
			})
		})
	}
}

// TestAccWorkflowTemplate_CoreEmptyStringListBehavior calls the API through the SDK,
// without Terraform, to pin down the core behavior the provider relies on for empty lists:
//   - create without Tags: core stores its default, so GET returns Tags = [] (why tags is
//     Optional+Computed);
//   - create without SharedOrgsList: core stores nothing, so GET omits it;
//   - create with SharedOrgsList = []: the SDK sends [] and GET returns it;
//   - update with Tags = [] and SharedOrgsList = []: core stores and returns [] for both.
func TestAccWorkflowTemplate_CoreEmptyStringListBehavior(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("acceptance test: set TF_ACC=1 to run")
	}
	acctest.TestAccPreCheck(t)

	ctx := context.Background()
	client := acctest.SGClient()
	org := config.Get().OrgName
	templateName := acctest.ResourceName("tf-provider-workflow-template-core-empty")

	t.Cleanup(func() { deleteWorkflowTemplateFixture(templateName) })

	kind := workflowtemplates.WorkflowTemplateSourceConfigKindEnum(sourceConfigKind)
	_, err := client.WorkflowTemplates.CreateWorkflowTemplate(ctx, org, false, &workflowtemplates.CreateWorkflowTemplateRequest{
		TemplateName:     templateName,
		OwnerOrg:         fmt.Sprintf("/orgs/%s", org),
		SourceConfigKind: &kind,
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	read, err := client.WorkflowTemplates.ReadWorkflowTemplate(ctx, org, templateName)
	if err != nil {
		t.Fatalf("read after create: %v", err)
	}
	if read.Msg.Tags == nil {
		t.Errorf("after create without Tags: GET omitted Tags, expected core to default it to []")
	} else if len(read.Msg.Tags) != 0 {
		t.Errorf("after create without Tags: got Tags = %v, want []", read.Msg.Tags)
	}
	if read.Msg.SharedOrgsList != nil {
		t.Errorf("after create without SharedOrgsList: got SharedOrgsList = %#v, expected GET to omit it", read.Msg.SharedOrgsList)
	}

	emptyTemplateName := acctest.ResourceName("tf-provider-workflow-template-core-empty-explicit")
	t.Cleanup(func() { deleteWorkflowTemplateFixture(emptyTemplateName) })

	_, err = client.WorkflowTemplates.CreateWorkflowTemplate(ctx, org, false, &workflowtemplates.CreateWorkflowTemplateRequest{
		TemplateName:     emptyTemplateName,
		OwnerOrg:         fmt.Sprintf("/orgs/%s", org),
		SourceConfigKind: &kind,
		SharedOrgsList:   &[]string{},
	})
	if err != nil {
		t.Fatalf("create with SharedOrgsList = []: %v", err)
	}

	emptyRead, err := client.WorkflowTemplates.ReadWorkflowTemplate(ctx, org, emptyTemplateName)
	if err != nil {
		t.Fatalf("read after create with SharedOrgsList = []: %v", err)
	}
	if emptyRead.Msg.SharedOrgsList == nil || len(emptyRead.Msg.SharedOrgsList) != 0 {
		t.Errorf("after create with SharedOrgsList = []: got %#v, want []", emptyRead.Msg.SharedOrgsList)
	}

	_, err = client.WorkflowTemplates.UpdateWorkflowTemplate(ctx, org, templateName, &workflowtemplates.UpdateWorkflowTemplateRequest{
		Tags:           sgsdkgo.Optional([]string{}),
		SharedOrgsList: sgsdkgo.Optional([]string{}),
	})
	if err != nil {
		t.Fatalf("update: %v", err)
	}

	read, err = client.WorkflowTemplates.ReadWorkflowTemplate(ctx, org, templateName)
	if err != nil {
		t.Fatalf("read after update: %v", err)
	}
	if read.Msg.Tags == nil || len(read.Msg.Tags) != 0 {
		t.Errorf("after update with Tags = []: got %#v, want []", read.Msg.Tags)
	}
	if read.Msg.SharedOrgsList == nil || len(read.Msg.SharedOrgsList) != 0 {
		t.Errorf("after update with SharedOrgsList = []: got %#v, want []", read.Msg.SharedOrgsList)
	}
}

// TestAccWorkflowTemplate_OmittedTags verifies that leaving tags out works now that an
// API [] reads back as []. Core stores a default of [] when the create request has no
// Tags; tags is Optional+Computed, so the plan leaves it unknown and accepts that [].
// A second step removes tags after they were set: UseStateForUnknown carries the state
// value forward, so the old tags stay (set `tags = []` to clear them).
func TestAccWorkflowTemplate_OmittedTags(t *testing.T) {
	templateName := acctest.ResourceName("tf-provider-workflow-template-omitted-tags")

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
				Config: testAccWorkflowTemplate(templateName, sourceConfigKind, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_template.test", "template_name", templateName),
					resource.TestCheckResourceAttr("stackguardian_workflow_template.test", "tags.#", "0"),
				),
			},
			{
				Config: testAccWorkflowTemplate(templateName, sourceConfigKind, `tags = ["tf-provider-test"]`),
				Check:  resource.TestCheckResourceAttr("stackguardian_workflow_template.test", "tags.0", "tf-provider-test"),
			},
			{
				Config: testAccWorkflowTemplate(templateName, sourceConfigKind, `description = "tags omitted"`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_template.test", "description", "tags omitted"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template.test", "tags.0", "tf-provider-test"),
				),
			},
		},
	})
}
