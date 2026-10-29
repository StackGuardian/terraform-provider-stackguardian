package stacktemplaterevision_test

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"testing"
	"time"

	sgsdkgo "github.com/StackGuardian/sg-sdk-go"
	sgclient "github.com/StackGuardian/sg-sdk-go/client"
	sgoption "github.com/StackGuardian/sg-sdk-go/option"
	"github.com/StackGuardian/sg-sdk-go/stacktemplaterevisions"
	"github.com/StackGuardian/sg-sdk-go/stacktemplates"
	"github.com/StackGuardian/sg-sdk-go/workflowtemplates"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/acctest"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

func getClient() *sgclient.Client {
	customHeader := http.Header{}
	customHeader.Set("x-sg-internal-auth-orgid", "sg-provider-test")

	return sgclient.NewClient(
		sgoption.WithApiKey(config.Get().FormatApiKey()),
		sgoption.WithBaseURL(config.Get().ApiUri),
		sgoption.WithHTTPHeader(customHeader),
	)
}

// --- Fixture helpers: create resources via SDK ---

func createWorkflowTemplateFixture(templateName, sourceConfigKind string) error {
	client := getClient()
	kind := workflowtemplates.WorkflowTemplateSourceConfigKindEnum(sourceConfigKind)
	// createFirstRevision=true: iac_template_id must name an existing revision
	// (see qualifiedWorkflowTemplateRevisionId), so the template needs ":1".
	_, err := client.WorkflowTemplates.CreateWorkflowTemplate(context.TODO(), config.Get().OrgName, true, &workflowtemplates.CreateWorkflowTemplateRequest{
		Id:               &templateName,
		TemplateName:     templateName,
		SourceConfigKind: &kind,
		TemplateType:     sgsdkgo.TemplateTypeEnumIac,
		IsPublic:         sgsdkgo.IsPublicEnumZero.Ptr(),
		OwnerOrg:         fmt.Sprintf("/orgs/%s", config.Get().OrgName),
	})
	return err
}

func createStackTemplateFixture(templateName, sourceConfigKind string) error {
	client := getClient()
	kind := stacktemplates.StackTemplateSourceConfigKindEnum(sourceConfigKind)
	_, err := client.StackTemplates.CreateStackTemplate(context.TODO(), config.Get().OrgName, false, &stacktemplates.CreateStackTemplateRequest{
		Id:               &templateName,
		TemplateName:     templateName,
		SourceConfigKind: &kind,
		TemplateType:     sgsdkgo.TemplateTypeEnumIacGroup,
		IsPublic:         sgsdkgo.IsPublicEnumZero.Ptr(),
		OwnerOrg:         fmt.Sprintf("/orgs/%s", config.Get().OrgName),
	})
	return err
}

// --- Fixture helpers: delete resources via SDK (best-effort, for cleanup) ---

func deleteWorkflowTemplateFixture(templateId string) {
	client := getClient()
	// The template can't be deleted while its revision :1 (created by
	// createWorkflowTemplateFixture) still exists, so that goes first.
	_ = client.WorkflowTemplatesRevisions.DeleteWorkflowTemplateRevision(context.TODO(), config.Get().OrgName, templateId+":1", true)
	_ = client.WorkflowTemplates.DeleteWorkflowTemplate(context.TODO(), config.Get().OrgName, templateId)
}

// qualifiedWorkflowTemplateRevisionId returns the full "/<org>/<template-id>:1" id of
// the revision createWorkflowTemplateFixture creates — the form iac_template_id
// requires (the org even for your own organization, and an existing revision).
func qualifiedWorkflowTemplateRevisionId(templateName string) string {
	return fmt.Sprintf("/%s/%s:1", config.Get().OrgName, templateName)
}

func deleteStackTemplateFixture(templateId string) {
	client := getClient()
	_ = client.StackTemplates.DeleteStackTemplate(context.TODO(), config.Get().OrgName, templateId)
}

func deleteStackTemplateRevisionFixture(revisionId string) {
	client := getClient()
	_ = client.StackTemplateRevisions.DeleteStackTemplateRevision(context.TODO(), config.Get().OrgName, revisionId, true)
}

// deprecateStackTemplateRevisionFixture deprecates a revision so it can be deleted
// (required when its is_public is "1"). The API call's result is ignored on purpose: if
// the revision was never published, deprecation isn't applicable and the call may fail —
// that's fine, since deletion doesn't need it in that case either.
func deprecateStackTemplateRevisionFixture(revisionId string) {
	client := getClient()
	// effectiveDate must already be in the past for the deprecation to take
	// effect immediately — a future date only schedules it, and the revision
	// still counts as actively published (blocking delete) until then. This
	// fixture always runs well after the revision was created, so a second
	// back is enough margin without risking landing before creation.
	effectiveDate := fmt.Sprintf("%d", time.Now().Add(-1*time.Second).Unix())
	message := "This revision is deprecated"
	_, _ = client.StackTemplateRevisions.UpdateStackTemplateRevision(context.TODO(), config.Get().OrgName, revisionId, &stacktemplaterevisions.UpdateStackTemplateRevisionRequest{
		Deprecation: sgsdkgo.Optional(stacktemplaterevisions.Deprecation{
			EffectiveDate: &effectiveDate,
			Message:       &message,
		}),
	})
}

// --- Terraform config generators ---

// testAccStackTemplateRevisionConfig generates a basic stack template revision config.
// notes and description are the displayed strings for those fields.
func testAccStackTemplateRevisionConfig(stackTemplateID, wfTemplateID, alias, notes, description string) string {
	return fmt.Sprintf(`
resource "stackguardian_stack_template_revision" "test" {
  parent_template_id = "%s"
  alias              = "%s"
  notes              = "%s"
  description        = "%s"
  source_config_kind = "TERRAFORM"

  workflows_config = {
    workflows = [
      {
        id            = "d8dfaf15-2ad9-da29-8af0-c6b288b12089"
        resource_name = "wf-1"

        vcs_config = {
          iac_vcs_config = {
            iac_template_id = "%s"
          }
        }

        terraform_config = {
          managed_terraform_state = true
          terraform_version       = "1.5.7"
        }
      }
    ]
  }
}
`, stackTemplateID, alias, notes, description, qualifiedWorkflowTemplateRevisionId(wfTemplateID))
}

func testAccStackTemplateRevisionWithWorkflowsConfig(stackTemplateID, wfTemplateID, alias string) string {
	return fmt.Sprintf(`
resource "stackguardian_stack_template_revision" "test" {
  parent_template_id = "%s"
  alias              = "%s"
  notes              = "Revision with workflows config"
  description        = "Test revision with workflows configuration"
  source_config_kind = "TERRAFORM"

  workflows_config = {
    workflows = [
      {
        id            = "d8dfaf15-2ad9-da29-8af0-c6b288b12089"
        resource_name = "wf-1"

        vcs_config = {
          iac_vcs_config = {
            iac_template_id = "%s"
          }
          iac_input_data = {
            schema_type = "RAW_JSON"
            data        = jsonencode({
              bucket_region = "eu-central-1"
            })
          }
        }

        terraform_config = {
          managed_terraform_state = true
          terraform_version       = "1.5.7"
        }
      }
    ]
  }
}
`, stackTemplateID, alias, qualifiedWorkflowTemplateRevisionId(wfTemplateID))
}

// --- Tests ---

func TestAccStackTemplateRevision_Basic(t *testing.T) {
	acctest.SkipUnlessAcceptance(t)
	stackTemplateID := acctest.ResourceName("provider-test-stack-template-rev1")
	wfTemplateID := acctest.ResourceName("provider-test-wft-for-stack-rev1")
	revisionAlias := "v1"

	// Register cleanup before creation so defers run even if a later create fails. The
	// revision must be deprecated before it can be deleted (required when its is_public is
	// "1", harmless otherwise), and deleted before the parent stack template.
	defer deleteWorkflowTemplateFixture(wfTemplateID)
	defer deleteStackTemplateFixture(stackTemplateID)
	defer deleteStackTemplateRevisionFixture(fmt.Sprintf("%s:1", stackTemplateID))
	defer deprecateStackTemplateRevisionFixture(fmt.Sprintf("%s:1", stackTemplateID))

	// Create prerequisite resources via SDK
	if err := createWorkflowTemplateFixture(wfTemplateID, "TERRAFORM"); err != nil {
		t.Fatalf("failed to create workflow template fixture: %s", err)
	}
	if err := createStackTemplateFixture(stackTemplateID, "TERRAFORM"); err != nil {
		t.Fatalf("failed to create stack template fixture: %s", err)
	}

	customHeader := http.Header{}
	customHeader.Set("x-sg-internal-auth-orgid", "sg-provider-test")

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader),
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccStackTemplateRevisionConfig(stackTemplateID, wfTemplateID, revisionAlias, "Test revision notes", "Test revision description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_stack_template_revision.test", "alias", revisionAlias),
					resource.TestCheckResourceAttr("stackguardian_stack_template_revision.test", "notes", "Test revision notes"),
					resource.TestCheckResourceAttr("stackguardian_stack_template_revision.test", "description", "Test revision description"),
					resource.TestCheckResourceAttrSet("stackguardian_stack_template_revision.test", "id"),
					resource.TestCheckResourceAttr("stackguardian_stack_template_revision.test", "parent_template_id", stackTemplateID),
					resource.TestCheckResourceAttrSet("stackguardian_stack_template_revision.test", "template_id"),
					resource.TestCheckResourceAttrSet("stackguardian_stack_template_revision.test", "workflows_config.%"),
				),
			},
			// Update and Read testing
			{
				Config: testAccStackTemplateRevisionConfig(stackTemplateID, wfTemplateID, revisionAlias, "Updated revision notes", "Updated revision description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_stack_template_revision.test", "alias", revisionAlias),
					resource.TestCheckResourceAttr("stackguardian_stack_template_revision.test", "notes", "Updated revision notes"),
					resource.TestCheckResourceAttr("stackguardian_stack_template_revision.test", "description", "Updated revision description"),
					resource.TestCheckResourceAttrSet("stackguardian_stack_template_revision.test", "id"),
					resource.TestCheckResourceAttr("stackguardian_stack_template_revision.test", "parent_template_id", stackTemplateID),
				),
			},
			// Delete testing automatically occurs
		},
	})
}

// TestAccStackTemplateRevision_Import checks an imported revision gets parent_template_id
// from its id, so the next plan is empty instead of replacing the revision
// (parent_template_id is RequiresReplace).
func TestAccStackTemplateRevision_Import(t *testing.T) {
	acctest.SkipUnlessAcceptance(t)
	stackTemplateID := acctest.ResourceName("provider-test-stack-template-revimp")
	wfTemplateID := acctest.ResourceName("provider-test-wft-for-stack-revimp")

	defer deleteWorkflowTemplateFixture(wfTemplateID)
	defer deleteStackTemplateFixture(stackTemplateID)
	defer deleteStackTemplateRevisionFixture(fmt.Sprintf("%s:1", stackTemplateID))
	defer deprecateStackTemplateRevisionFixture(fmt.Sprintf("%s:1", stackTemplateID))

	if err := createWorkflowTemplateFixture(wfTemplateID, "TERRAFORM"); err != nil {
		t.Fatalf("failed to create workflow template fixture: %s", err)
	}
	if err := createStackTemplateFixture(stackTemplateID, "TERRAFORM"); err != nil {
		t.Fatalf("failed to create stack template fixture: %s", err)
	}

	customHeader := http.Header{}
	customHeader.Set("x-sg-internal-auth-orgid", "sg-provider-test")

	hcl := testAccStackTemplateRevisionConfig(stackTemplateID, wfTemplateID, "v1", "notes", "description")

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader),
		Steps: []resource.TestStep{
			{
				Config: hcl,
			},
			// Import into a fresh state and compare with the state the create produced.
			{
				Config:            hcl,
				ResourceName:      "stackguardian_stack_template_revision.test",
				ImportState:       true,
				ImportStateId:     stackTemplateID + ":1",
				ImportStateVerify: true,
				ImportStateCheck: func(states []*terraform.InstanceState) error {
					if got := states[0].Attributes["parent_template_id"]; got != stackTemplateID {
						return fmt.Errorf("imported parent_template_id = %q, want %q", got, stackTemplateID)
					}
					return nil
				},
			},
		},
	})
}

func TestAccStackTemplateRevision_WithWorkflowsConfig(t *testing.T) {
	acctest.SkipUnlessAcceptance(t)
	stackTemplateID := acctest.ResourceName("provider-test-stack-template-rev2")
	wfTemplateID := acctest.ResourceName("provider-test-wft-for-stack-rev2")
	revisionAlias := "v1"

	// Register cleanup before creation so defers run even if a later create fails. The
	// revision must be deprecated before it can be deleted (required when its is_public is
	// "1", harmless otherwise), and deleted before the parent stack template.
	defer deleteWorkflowTemplateFixture(wfTemplateID)
	defer deleteStackTemplateFixture(stackTemplateID)
	defer deleteStackTemplateRevisionFixture(fmt.Sprintf("%s:1", stackTemplateID))
	defer deprecateStackTemplateRevisionFixture(fmt.Sprintf("%s:1", stackTemplateID))

	// Create prerequisite resources via SDK
	if err := createWorkflowTemplateFixture(wfTemplateID, "TERRAFORM"); err != nil {
		t.Fatalf("failed to create workflow template fixture: %s", err)
	}
	if err := createStackTemplateFixture(stackTemplateID, "TERRAFORM"); err != nil {
		t.Fatalf("failed to create stack template fixture: %s", err)
	}

	customHeader := http.Header{}
	customHeader.Set("x-sg-internal-auth-orgid", "sg-provider-test")

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader),
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccStackTemplateRevisionWithWorkflowsConfig(stackTemplateID, wfTemplateID, revisionAlias),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_stack_template_revision.test", "alias", revisionAlias),
					resource.TestCheckResourceAttr("stackguardian_stack_template_revision.test", "notes", "Revision with workflows config"),
					resource.TestCheckResourceAttr("stackguardian_stack_template_revision.test", "description", "Test revision with workflows configuration"),
					resource.TestCheckResourceAttrSet("stackguardian_stack_template_revision.test", "id"),
					resource.TestCheckResourceAttr("stackguardian_stack_template_revision.test", "parent_template_id", stackTemplateID),
					resource.TestCheckResourceAttrSet("stackguardian_stack_template_revision.test", "template_id"),
					resource.TestCheckResourceAttrSet("stackguardian_stack_template_revision.test", "workflows_config.%"),
				),
			},
		},
	})
}

// TestAccStackTemplateRevision_EmptyIacInputData verifies a workflow can set
// vcs_config.iac_input_data = {} with no fields. The API accepts an explicitly empty
// iacInputData and treats it as "No Inputs", which is
// different from leaving iac_input_data out. Step 1 checks the apply succeeds, the
// platform stores an explicit empty iacInputData (read through the SDK) and state holds
// iac_input_data with no fields set; step 2 checks a re-plan shows no diff.
func TestAccStackTemplateRevision_EmptyIacInputData(t *testing.T) {
	acctest.SkipUnlessAcceptance(t)
	stackTemplateID := acctest.ResourceName("provider-test-stack-template-emptyinput")
	wfTemplateID := acctest.ResourceName("provider-test-wft-for-stack-emptyinput")
	revisionID := fmt.Sprintf("%s:1", stackTemplateID)

	// Cleanup runs in reverse: deprecate and delete the revision, then the stack template,
	// then the workflow template.
	defer deleteWorkflowTemplateFixture(wfTemplateID)
	defer deleteStackTemplateFixture(stackTemplateID)
	defer deleteStackTemplateRevisionFixture(revisionID)
	defer deprecateStackTemplateRevisionFixture(revisionID)

	if err := createWorkflowTemplateFixture(wfTemplateID, "TERRAFORM"); err != nil {
		t.Fatalf("failed to create workflow template fixture: %s", err)
	}
	if err := createStackTemplateFixture(stackTemplateID, "TERRAFORM"); err != nil {
		t.Fatalf("failed to create stack template fixture: %s", err)
	}

	hcl := fmt.Sprintf(`
resource "stackguardian_stack_template_revision" "test" {
  parent_template_id = %q
  alias              = "v1"
  notes              = "Revision with empty iac_input_data"
  source_config_kind = "TERRAFORM"

  workflows_config = {
    workflows = [
      {
        id            = "d8dfaf15-2ad9-da29-8af0-c6b288b12089"
        resource_name = "wf-1"

        vcs_config = {
          iac_vcs_config = {
            iac_template_id = %q
          }
          iac_input_data = {}
        }

        terraform_config = {
          managed_terraform_state = true
          terraform_version       = "1.5.7"
        }
      }
    ]
  }
}
`, stackTemplateID, qualifiedWorkflowTemplateRevisionId(wfTemplateID))

	const res = "stackguardian_stack_template_revision.test"
	iacInput := "workflows_config.workflows.0.vcs_config.iac_input_data."

	// checkPlatformStoredEmptyInput reads the revision through the SDK and checks the
	// workflow's iacInputData is present but empty, i.e. "No Inputs" rather than absent.
	checkPlatformStoredEmptyInput := func(*terraform.State) error {
		readResp, err := getClient().StackTemplateRevisions.ReadStackTemplateRevision(context.TODO(), config.Get().OrgName, revisionID)
		if err != nil {
			return fmt.Errorf("read stack template revision %q: %w", revisionID, err)
		}
		wfc := readResp.Msg.WorkflowsConfig
		if wfc == nil || len(wfc.Workflows) == 0 || wfc.Workflows[0] == nil {
			return fmt.Errorf("revision %q: no workflows in the response", revisionID)
		}
		vcs := wfc.Workflows[0].VcsConfig
		if vcs == nil || vcs.IacInputData == nil {
			return fmt.Errorf("revision %q: iacInputData was not stored; want an explicit empty object", revisionID)
		}
		in := vcs.IacInputData
		if in.SchemaType != nil || in.SchemaId != nil || (in.Data != nil && len(*in.Data) > 0) {
			return fmt.Errorf("revision %q: iacInputData = %+v, want an empty object", revisionID, in)
		}
		return nil
	}

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
				Config: hcl,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr(res, iacInput+"schema_type"),
					resource.TestCheckNoResourceAttr(res, iacInput+"schema_id"),
					resource.TestCheckNoResourceAttr(res, iacInput+"data"),
					checkPlatformStoredEmptyInput,
				),
			},
			{
				// Round trips with no diff.
				Config:   hcl,
				PlanOnly: true,
			},
		},
	})
}

// TestAccStackTemplateRevision_IacInputDataPartialRejected verifies that once any
// iac_input_data field is set, schema_type and data are both required (AlsoRequires
// validators), even though schema_type is Optional to allow iac_input_data = {}. The
// validators fail before any API call, so no fixtures are needed.
func TestAccStackTemplateRevision_IacInputDataPartialRejected(t *testing.T) {
	acctest.SkipUnlessAcceptance(t)

	withIacInputData := func(body string) string {
		return fmt.Sprintf(`
resource "stackguardian_stack_template_revision" "test" {
  parent_template_id = "my-stack-template"
  alias              = "v1"
  source_config_kind = "TERRAFORM"

  workflows_config = {
    workflows = [
      {
        id = "d8dfaf15-2ad9-da29-8af0-c6b288b12089"
        vcs_config = {
          iac_vcs_config = {
            iac_template_id = "/%s/my-workflow-template:1"
          }
          iac_input_data = %s
        }
      }
    ]
  }
}
`, config.Get().OrgName, body)
	}

	// AlsoRequires reports one "Invalid Attribute Combination" error per missing
	// attribute, naming it by its full path.
	missing := func(attr string) *regexp.Regexp {
		return regexp.MustCompile(`(?s)Invalid Attribute Combination.*iac_input_data\.` + attr)
	}
	// Both schema_type and data missing: two errors, in either order.
	missingBoth := regexp.MustCompile(`(?s)Invalid Attribute Combination.*(iac_input_data\.schema_type.*iac_input_data\.data|iac_input_data\.data.*iac_input_data\.schema_type)`)

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
				// schema_type without data.
				Config:      withIacInputData(`{ schema_type = "RAW_JSON" }`),
				PlanOnly:    true,
				ExpectError: missing("data"),
			},
			{
				// data without schema_type.
				Config:      withIacInputData(`{ data = jsonencode({ key = "value" }) }`),
				PlanOnly:    true,
				ExpectError: missing("schema_type"),
			},
			{
				// schema_id alone: both schema_type and data are missing.
				Config:      withIacInputData(`{ schema_id = "my-schema" }`),
				PlanOnly:    true,
				ExpectError: missingBoth,
			},
		},
	})
}

// TestAccStackTemplateRevision_ActionsRejectsEmpty verifies actions = {} (a known,
// explicitly empty map — not the same as leaving the attribute unset) is rejected at
// plan time by mapvalidator.SizeAtLeast(1) on "actions" (schema.go). The API generates
// default actions on create when given none, and rejects an empty map on update, so {}
// could never round-trip. The validator fails before any API call, so no fixtures are
// needed.
func TestAccStackTemplateRevision_ActionsRejectsEmpty(t *testing.T) {
	acctest.SkipUnlessAcceptance(t)
	stackTemplateID := acctest.ResourceName("provider-test-stack-template-actempty")
	wfTemplateID := acctest.ResourceName("provider-test-wft-for-stack-actempty")

	config := strings.TrimSuffix(strings.TrimSpace(testAccStackTemplateRevisionConfig(stackTemplateID, wfTemplateID, "v1", "notes", "description")), "}") + `
  actions = {}
}
`

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
				Config:      config,
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`(?s)actions.*at least 1`),
			},
		},
	})
}

// checkActionsDefaultTrue checks that every action stored on the revision has
// default=true on the platform, not just in state.
func checkActionsDefaultTrue(revisionID string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		readResp, err := getClient().StackTemplateRevisions.ReadStackTemplateRevision(context.TODO(), config.Get().OrgName, revisionID)
		if err != nil {
			return fmt.Errorf("reading revision %s: %w", revisionID, err)
		}
		if len(readResp.Msg.Actions) == 0 {
			return fmt.Errorf("revision %s has no actions", revisionID)
		}
		for name, action := range readResp.Msg.Actions {
			if action == nil || action.Default == nil || !*action.Default {
				return fmt.Errorf("action %q: default is not true on the platform", name)
			}
		}
		return nil
	}
}

// TestAccStackTemplateRevision_ActionsDefaultAlwaysTrue checks actions[*].default is
// always true: for the actions the API generates when none are declared, and for
// declared actions, where the provider sends true since default isn't user-settable.
func TestAccStackTemplateRevision_ActionsDefaultAlwaysTrue(t *testing.T) {
	acctest.SkipUnlessAcceptance(t)
	stackTemplateID := acctest.ResourceName("provider-test-stack-template-actdef")
	wfTemplateID := acctest.ResourceName("provider-test-wft-for-stack-actdef")
	revisionID := stackTemplateID + ":1"

	defer deleteWorkflowTemplateFixture(wfTemplateID)
	defer deleteStackTemplateFixture(stackTemplateID)
	defer deleteStackTemplateRevisionFixture(revisionID)
	defer deprecateStackTemplateRevisionFixture(revisionID)

	if err := createWorkflowTemplateFixture(wfTemplateID, "TERRAFORM"); err != nil {
		t.Fatalf("failed to create workflow template fixture: %s", err)
	}
	if err := createStackTemplateFixture(stackTemplateID, "TERRAFORM"); err != nil {
		t.Fatalf("failed to create stack template fixture: %s", err)
	}

	withoutActions := testAccStackTemplateRevisionConfig(stackTemplateID, wfTemplateID, "v1", "notes", "description")
	withActions := strings.TrimSuffix(strings.TrimSpace(withoutActions), "}") + `
  actions = {
    custom = {
      name        = "Custom"
      description = "a declared action"
      order = {
        "d8dfaf15-2ad9-da29-8af0-c6b288b12089" = {}
      }
    }
  }
}
`

	customHeader := http.Header{}
	customHeader.Set("x-sg-internal-auth-orgid", "sg-provider-test")

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader),
		Steps: []resource.TestStep{
			// No actions declared: the API generates apply/plan/destroy, all default=true.
			{
				Config: withoutActions,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_stack_template_revision.test", "actions.apply.default", "true"),
					resource.TestCheckResourceAttr("stackguardian_stack_template_revision.test", "actions.plan.default", "true"),
					resource.TestCheckResourceAttr("stackguardian_stack_template_revision.test", "actions.destroy.default", "true"),
					checkActionsDefaultTrue(revisionID),
				),
			},
			// Declared actions, without default: the provider sends default=true.
			{
				Config: withActions,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_stack_template_revision.test", "actions.%", "1"),
					resource.TestCheckResourceAttr("stackguardian_stack_template_revision.test", "actions.custom.default", "true"),
					checkActionsDefaultTrue(revisionID),
				),
			},
			{
				Config:   withActions,
				PlanOnly: true,
			},
		},
	})
}

// TestAccStackTemplateRevision_IacTemplateIdRequiresOrg verifies iac_template_id is
// rejected at plan time unless it is the full "/<org>/<template-id>:<revision>" id of a
// revision, the org given even for your own organization
// (validators.QualifiedTemplateRevisionId). The validator fails before any API call,
// so no fixtures are needed.
func TestAccStackTemplateRevision_IacTemplateIdRequiresOrg(t *testing.T) {
	acctest.SkipUnlessAcceptance(t)

	bare := testAccStackTemplateRevisionConfig("my-stack-template", "unused", "v1", "notes", "description")
	bare = strings.Replace(bare, qualifiedWorkflowTemplateRevisionId("unused"), "my-workflow-template:1", 1)
	noRevision := strings.Replace(bare, "my-workflow-template:1", "/"+config.Get().OrgName+"/my-workflow-template", 1)

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
				Config:      bare,
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`(?s)iac_template_id.*full id`),
			},
			{
				Config:      noRevision,
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`(?s)iac_template_id.*full id`),
			},
		},
	})
}

func TestAccStackTemplateRevision_Lifecycle(t *testing.T) {
	acctest.SkipUnlessAcceptance(t)
	stackTemplateName := acctest.ResourceName("tf-provider-stack-template-lifecycle")
	wfTemplateName := acctest.ResourceName("tf-provider-wft-lifecycle-for-stack")
	alias := "v1"
	// Must land after the revision is created (Step 1) but before Step 3
	// actually applies the deprecation — see Step 3 below. This is computed
	// now, before Step 1 has even run, so a negative offset would put it
	// before creation; a few seconds into the future instead reliably lands
	// after creation and has already elapsed into the past by the time Step 3
	// runs, since Steps 1 and 2 each involve a real API round trip.
	deprecationEffectiveDate := fmt.Sprintf("%d", time.Now().Add(3*time.Second).Unix())

	// Safety-net cleanup. Defers run LIFO, so registration order here is the
	// reverse of execution order:
	//   execution: deprecate revision → delete revision → delete stack template → delete workflow template
	defer deleteWorkflowTemplateFixture(wfTemplateName)
	defer deleteStackTemplateFixture(stackTemplateName)
	defer deleteStackTemplateRevisionFixture(fmt.Sprintf("%s:1", stackTemplateName))
	defer deprecateStackTemplateRevisionFixture(fmt.Sprintf("%s:1", stackTemplateName))

	customHeader := http.Header{}
	customHeader.Set("x-sg-internal-auth-orgid", "sg-provider-test")

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader),
		Steps: []resource.TestStep{
			// Step 1: Create parent stack template, parent workflow template, and revision
			{
				Config: testAccStackTemplateRevisionLifecycleConfig(stackTemplateName, wfTemplateName, alias, "", false, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_stack_template.parent", "template_name", stackTemplateName),
					resource.TestCheckResourceAttrSet("stackguardian_stack_template.parent", "id"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template.wf_parent", "template_name", wfTemplateName),
					resource.TestCheckResourceAttrSet("stackguardian_workflow_template.wf_parent", "id"),
					resource.TestCheckResourceAttr("stackguardian_stack_template_revision.test", "alias", alias),
					resource.TestCheckResourceAttr("stackguardian_stack_template_revision.test", "notes", "Initial revision"),
					resource.TestCheckResourceAttrSet("stackguardian_stack_template_revision.test", "id"),
					resource.TestCheckResourceAttr("stackguardian_stack_template_revision.test", "parent_template_id", stackTemplateName),
				),
			},
			// Step 2: Publish the revision
			{
				Config: testAccStackTemplateRevisionLifecycleConfig(stackTemplateName, wfTemplateName, alias, "1", false, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_stack_template_revision.test", "is_public", "1"),
				),
			},
			// Step 3: Deprecate the revision. effectiveDate must already be in
			// the past, or the revision still counts as actively published
			// and the implicit destroy below 400s with "Published template
			// cannot be deleted. Please deprecate the template first."
			{
				Config: testAccStackTemplateRevisionLifecycleConfig(stackTemplateName, wfTemplateName, alias, "1", true, deprecationEffectiveDate),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_stack_template_revision.test", "deprecation.message", "This revision is deprecated"),
					resource.TestCheckResourceAttr("stackguardian_stack_template_revision.test", "deprecation.effective_date", deprecationEffectiveDate),
				),
			},
			// Terraform destroy automatically deletes the revision, stack template, and workflow template
		},
	})
}

// testAccStackTemplateRevisionLifecycleConfig generates the lifecycle Terraform config.
// isPublic controls the revision's is_public value (pass "" to omit).
// deprecated=true adds a deprecation block to the revision using effectiveDate.
func testAccStackTemplateRevisionLifecycleConfig(stackTemplateName, wfTemplateName, alias, isPublic string, deprecated bool, effectiveDate string) string {
	isPublicLine := ""
	if isPublic != "" {
		isPublicLine = fmt.Sprintf("  is_public          = %q\n", isPublic)
	}
	deprecationBlock := ""
	if deprecated {
		deprecationBlock = fmt.Sprintf(`
  deprecation = {
    effective_date = %q
    message        = "This revision is deprecated"
  }`, effectiveDate)
	}
	return fmt.Sprintf(`
resource "stackguardian_workflow_template" "wf_parent" {
  template_name      = "%s"
  source_config_kind = "TERRAFORM"
  is_public          = "0"
  tags               = ["test", "lifecycle"]
}

# iac_template_id must name an existing revision, so the workflow template
# needs one of its own.
resource "stackguardian_workflow_template_revision" "wf_rev" {
  template_id        = stackguardian_workflow_template.wf_parent.id
  alias              = "v1"
  source_config_kind = "TERRAFORM"
  user_job_cpu       = 500
  user_job_memory    = 1024
}

resource "stackguardian_stack_template" "parent" {
  template_name      = "%s"
  source_config_kind = "TERRAFORM"
  is_public          = "0"
  tags               = ["test", "lifecycle"]
}

resource "stackguardian_stack_template_revision" "test" {
  parent_template_id = stackguardian_stack_template.parent.id
  alias              = "%s"
  notes              = "Initial revision"
  description        = "Lifecycle test revision"
  source_config_kind = "TERRAFORM"
%s
  workflows_config = {
    workflows = [
      {
        id            = "d8dfaf15-2ad9-da29-8af0-c6b288b12089"
        resource_name = "wf-1"

        vcs_config = {
          iac_vcs_config = {
            iac_template_id = "/%s/${stackguardian_workflow_template_revision.wf_rev.id}"
          }
        }

        terraform_config = {
          managed_terraform_state = true
          terraform_version       = "1.5.7"
        }
      }
    ]
  }
%s
}
`, wfTemplateName, stackTemplateName, alias, isPublicLine, config.Get().OrgName, deprecationBlock)
}
