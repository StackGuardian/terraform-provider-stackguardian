package stacktemplaterevision_test

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	sgsdkgo "github.com/StackGuardian/sg-sdk-go"
	sgclient "github.com/StackGuardian/sg-sdk-go/client"
	sgoption "github.com/StackGuardian/sg-sdk-go/option"
	"github.com/StackGuardian/sg-sdk-go/stacktemplaterevisions"
	"github.com/StackGuardian/sg-sdk-go/stacktemplates"
	"github.com/StackGuardian/sg-sdk-go/workflowtemplates"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

var org = os.Getenv("STACKGUARDIAN_ORG_NAME")

func getClient() *sgclient.Client {
	customHeader := http.Header{}
	customHeader.Set("x-sg-internal-auth-orgid", "sg-provider-test")

	return sgclient.NewClient(
		sgoption.WithApiKey(fmt.Sprintf("apikey %s", os.Getenv("STACKGUARDIAN_API_KEY"))),
		sgoption.WithBaseURL(os.Getenv("STACKGUARDIAN_API_URI")),
		sgoption.WithHTTPHeader(customHeader),
	)
}

// --- Fixture helpers: create resources via SDK ---

func createWorkflowTemplateFixture(templateName, sourceConfigKind string) error {
	client := getClient()
	kind := workflowtemplates.WorkflowTemplateSourceConfigKindEnum(sourceConfigKind)
	_, err := client.WorkflowTemplates.CreateWorkflowTemplate(context.TODO(), org, false, &workflowtemplates.CreateWorkflowTemplateRequest{
		Id:               &templateName,
		TemplateName:     templateName,
		SourceConfigKind: &kind,
		TemplateType:     sgsdkgo.TemplateTypeEnumIac,
		IsPublic:         sgsdkgo.IsPublicEnumZero.Ptr(),
		OwnerOrg:         fmt.Sprintf("/orgs/%s", org),
	})
	return err
}

func createStackTemplateFixture(templateName, sourceConfigKind string) error {
	client := getClient()
	kind := stacktemplates.StackTemplateSourceConfigKindEnum(sourceConfigKind)
	_, err := client.StackTemplates.CreateStackTemplate(context.TODO(), org, false, &stacktemplates.CreateStackTemplateRequest{
		Id:               &templateName,
		TemplateName:     templateName,
		SourceConfigKind: &kind,
		TemplateType:     sgsdkgo.TemplateTypeEnumIacGroup,
		IsPublic:         sgsdkgo.IsPublicEnumZero.Ptr(),
		OwnerOrg:         fmt.Sprintf("/orgs/%s", org),
	})
	return err
}

// --- Fixture helpers: delete resources via SDK (best-effort, for cleanup) ---

func deleteWorkflowTemplateFixture(templateId string) {
	client := getClient()
	client.WorkflowTemplates.DeleteWorkflowTemplate(context.TODO(), org, templateId)
}

func deleteStackTemplateFixture(templateId string) {
	client := getClient()
	client.StackTemplates.DeleteStackTemplate(context.TODO(), org, templateId)
}

func deleteStackTemplateRevisionFixture(revisionId string) {
	client := getClient()
	client.StackTemplateRevisions.DeleteStackTemplateRevision(context.TODO(), org, revisionId, true)
}

func deprecateStackTemplateRevisionFixture(revisionId string) {
	client := getClient()
	effectiveDate := fmt.Sprintf("%d", time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC).Unix())
	message := "This revision is deprecated"
	client.StackTemplateRevisions.UpdateStackTemplateRevision(context.TODO(), org, revisionId, &stacktemplaterevisions.UpdateStackTemplateRevisionRequest{
		Deprecation: sgsdkgo.Optional(stacktemplaterevisions.Deprecation{
			EffectiveDate: &effectiveDate,
			Message:       &message,
		}),
	})
}

// --- Terraform config generators ---

func testAccStackTemplateRevisionConfig(stackTemplateID, wfTemplateID, alias string) string {
	return fmt.Sprintf(`
resource "stackguardian_stack_template_revision" "test" {
  parent_template_id = "%s"
  alias              = "%s"
  notes              = "Test revision notes"
  description        = "Test revision description"
  source_config_kind = "TERRAFORM"

  workflows_config = {
    workflows = [
      {
        id            = "d8dfaf15-2ad9-da29-8af0-c6b288b12089"
        template_id   = "%s"
        resource_name = "wf-1"

        terraform_config = {
          managed_terraform_state = true
          terraform_version       = "1.5.7"
        }
      }
    ]
  }
}
`, stackTemplateID, alias, wfTemplateID)
}

func testAccStackTemplateRevisionConfigUpdated(stackTemplateID, wfTemplateID, alias string) string {
	return fmt.Sprintf(`
resource "stackguardian_stack_template_revision" "test" {
  parent_template_id = "%s"
  alias              = "%s"
  notes              = "Updated revision notes"
  description        = "Updated revision description"
  source_config_kind = "TERRAFORM"

  workflows_config = {
    workflows = [
      {
        id            = "d8dfaf15-2ad9-da29-8af0-c6b288b12089"
        template_id   = "%s"
        resource_name = "wf-1"

        terraform_config = {
          managed_terraform_state = true
          terraform_version       = "1.5.7"
        }
      }
    ]
  }
}
`, stackTemplateID, alias, wfTemplateID)
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
        template_id   = "%s"
        resource_name = "wf-1"

        vcs_config = {
          iac_vcs_config = {
            use_marketplace_template = true
            iac_template_id          = "%s"
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
`, stackTemplateID, alias, wfTemplateID, wfTemplateID)
}

// --- Tests ---

func TestAccStackTemplateRevision_Basic(t *testing.T) {
	stackTemplateID := "provider-test-stack-template-rev1"
	wfTemplateID := "provider-test-wft-for-stack-rev1"
	revisionAlias := "v1"

	// Register cleanup before creation so defers run even if a later create fails
	defer deleteStackTemplateRevisionFixture(fmt.Sprintf("%s:1", stackTemplateID))
	defer deleteStackTemplateFixture(stackTemplateID)
	defer deleteWorkflowTemplateFixture(wfTemplateID)

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
				Config: testAccStackTemplateRevisionConfig(stackTemplateID, wfTemplateID, revisionAlias),
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
				Config: testAccStackTemplateRevisionConfigUpdated(stackTemplateID, wfTemplateID, revisionAlias),
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

func TestAccStackTemplateRevision_WithWorkflowsConfig(t *testing.T) {
	stackTemplateID := "provider-test-stack-template-rev2"
	wfTemplateID := "provider-test-wft-for-stack-rev2"
	revisionAlias := "v1"

	// Register cleanup before creation so defers run even if a later create fails
	defer deleteStackTemplateRevisionFixture(fmt.Sprintf("%s:1", stackTemplateID))
	defer deleteStackTemplateFixture(stackTemplateID)
	defer deleteWorkflowTemplateFixture(wfTemplateID)

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

func TestAccStackTemplateRevision_Lifecycle(t *testing.T) {
	stackTemplateName := "tf-provider-stack-template-lifecycle"
	wfTemplateName := "tf-provider-wft-lifecycle-for-stack"
	alias := "v1"

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
				Config: testAccStackTemplateRevisionLifecycleCreate(stackTemplateName, wfTemplateName, alias),
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
				Config: testAccStackTemplateRevisionLifecyclePublish(stackTemplateName, wfTemplateName, alias),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_stack_template_revision.test", "is_active", "1"),
					resource.TestCheckResourceAttr("stackguardian_stack_template_revision.test", "is_public", "1"),
				),
			},
			// Step 3: Deprecate the revision
			{
				Config: testAccStackTemplateRevisionLifecycleDeprecate(stackTemplateName, wfTemplateName, alias),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_stack_template_revision.test", "is_active", "1"),
					resource.TestCheckResourceAttr("stackguardian_stack_template_revision.test", "deprecation.message", "This revision is deprecated"),
					resource.TestCheckResourceAttr("stackguardian_stack_template_revision.test", "deprecation.effective_date", "1776011863"),
				),
			},
			// Terraform destroy automatically deletes the revision, stack template, and workflow template
		},
	})
}

func testAccStackTemplateRevisionLifecycleCreate(stackTemplateName, wfTemplateName, alias string) string {
	return fmt.Sprintf(`
resource "stackguardian_workflow_template" "wf_parent" {
  template_name      = "%s"
  source_config_kind = "TERRAFORM"
  is_public          = "0"
  tags               = ["test", "lifecycle"]
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

  workflows_config = {
    workflows = [
      {
        id            = "d8dfaf15-2ad9-da29-8af0-c6b288b12089"
        template_id   = stackguardian_workflow_template.wf_parent.id
        resource_name = "wf-1"

        terraform_config = {
          managed_terraform_state = true
          terraform_version       = "1.5.7"
        }
      }
    ]
  }
}
`, wfTemplateName, stackTemplateName, alias)
}

func testAccStackTemplateRevisionLifecyclePublish(stackTemplateName, wfTemplateName, alias string) string {
	return fmt.Sprintf(`
resource "stackguardian_workflow_template" "wf_parent" {
  template_name      = "%s"
  source_config_kind = "TERRAFORM"
  is_public          = "0"
  tags               = ["test", "lifecycle"]
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
  is_active          = "1"
  is_public          = "1"

  workflows_config = {
    workflows = [
      {
        id            = "d8dfaf15-2ad9-da29-8af0-c6b288b12089"
        template_id   = stackguardian_workflow_template.wf_parent.id
        resource_name = "wf-1"

        terraform_config = {
          managed_terraform_state = true
          terraform_version       = "1.5.7"
        }
      }
    ]
  }
}
`, wfTemplateName, stackTemplateName, alias)
}

func testAccStackTemplateRevisionLifecycleDeprecate(stackTemplateName, wfTemplateName, alias string) string {
	return fmt.Sprintf(`
resource "stackguardian_workflow_template" "wf_parent" {
  template_name      = "%s"
  source_config_kind = "TERRAFORM"
  is_public          = "0"
  tags               = ["test", "lifecycle"]
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
  is_active          = "1"

  workflows_config = {
    workflows = [
      {
        id            = "d8dfaf15-2ad9-da29-8af0-c6b288b12089"
        template_id   = stackguardian_workflow_template.wf_parent.id
        resource_name = "wf-1"

        terraform_config = {
          managed_terraform_state = true
          terraform_version       = "1.5.7"
        }
      }
    ]
  }

  deprecation = {
    effective_date = "1776011863"
    message        = "This revision is deprecated"
  }
}
`, wfTemplateName, stackTemplateName, alias)
}
