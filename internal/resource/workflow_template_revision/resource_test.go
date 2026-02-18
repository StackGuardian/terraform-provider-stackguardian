package workflowtemplaterevision_test

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"

	sgsdkgo "github.com/StackGuardian/sg-sdk-go"
	sgclient "github.com/StackGuardian/sg-sdk-go/client"
	sgoption "github.com/StackGuardian/sg-sdk-go/option"
	"github.com/StackGuardian/sg-sdk-go/workflowtemplates"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

var org = os.Getenv("STACKGUARDIAN_ORG_NAME")

func GetClient() *sgclient.Client {
	customHeader := http.Header{}
	customHeader.Set("x-sg-internal-auth-orgid", "sg-provider-test")

	client := sgclient.NewClient(sgoption.WithApiKey(fmt.Sprintf("apikey %s", os.Getenv("STACKGUARDIAN_API_KEY"))), sgoption.WithBaseURL(os.Getenv("STACKGUARDIAN_API_URI")), sgoption.WithHTTPHeader(customHeader))

	return client
}

func SampleCreateWorkflowPayload(templateName, sourceConfigKind string) *workflowtemplates.CreateWorkflowTemplateRequest {
	var sampleCreateWorkflowTemplatePayload = workflowtemplates.CreateWorkflowTemplateRequest{
		TemplateName:     templateName,
		SourceConfigKind: (*workflowtemplates.WorkflowTemplateSourceConfigKindEnum)(&sourceConfigKind),
		TemplateType:     sgsdkgo.TemplateTypeEnum("IAC"),
		IsPublic:         sgsdkgo.IsPublicEnumZero.Ptr(),
		OwnerOrg:         fmt.Sprintf("/orgs/%s", org),
	}
	return &sampleCreateWorkflowTemplatePayload
}

func createWorkflowTemplateFixture(templateName, sourceConfigKind string) error {
	client := GetClient()

	templatePayload := SampleCreateWorkflowPayload(templateName, sourceConfigKind)

	_, err := client.WorkflowTemplates.CreateWorkflowTemplate(context.TODO(), org, false, templatePayload)
	if err != nil {
		return err
	}
	return nil
}

func deleteWorkflowTemplateFixture(templateId string) {
	client := GetClient()

	client.WorkflowTemplates.DeleteWorkflowTemplate(context.TODO(), org, templateId)
}

func deleteWorkflowTemplateRevisionFixture(revisionId string) {
	client := GetClient()

	client.WorkflowTemplatesRevisions.DeleteWorkflowTemplateRevision(context.TODO(), org, revisionId, true)
}

func TestAccWorkflowTemplateRevision_Basic(t *testing.T) {
	templateID := "tf-provider-workflow-template-revision-1"
	alias := "revision1"

	err := createWorkflowTemplateFixture(templateID, "TERRAFORM")
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer deleteWorkflowTemplateFixture(templateID)
	defer deleteWorkflowTemplateRevisionFixture(fmt.Sprintf("%s:1", templateID))

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
				Config: testAccWorkflowTemplateRevisionConfig(templateID, alias),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "template_id", templateID),
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "alias", alias),
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "is_public", "0"),
					resource.TestCheckResourceAttrSet("stackguardian_workflow_template_revision.test", "id"),
				),
			},
			// Update and Read testing
			{
				Config: testAccWorkflowTemplateRevisionConfigUpdated(templateID, alias),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "alias", alias),
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "notes", "Updated revision notes"),
				),
			},
			// Delete testing automatically occurs
		},
	})
}

func TestAccWorkflowTemplateRevision_WithConfig(t *testing.T) {
	templateID := "test-workflow-template-revision"
	alias := "revision2"

	err := createWorkflowTemplateFixture(templateID, "TERRAFORM")
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer deleteWorkflowTemplateFixture(templateID)
	defer deleteWorkflowTemplateRevisionFixture(fmt.Sprintf("%s:1", templateID))

	customHeader := http.Header{}
	customHeader.Set("x-sg-internal-auth-orgid", "sg-provider-test")

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader),
		Steps: []resource.TestStep{
			// Create and Read testing with full configuration
			{
				Config: testAccWorkflowTemplateRevisionConfigWithDetails(templateID, alias),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "template_id", templateID),
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "alias", alias),
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "user_job_cpu", "2"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "user_job_memory", "4096"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "number_of_approvals_required", "1"),
				),
			},
			// Delete testing automatically occurs
		},
	})
}

func testAccWorkflowTemplateRevisionConfig(templateID, alias string) string {
	return fmt.Sprintf(`
resource "stackguardian_workflow_template_revision" "test" {
  template_id = "%s"
  alias       = "%s"
  source_config_kind = "TERRAFORM"
  is_public   = "0"
  tags        = ["test", "terraform"]
}
`, templateID, alias)
}

func testAccWorkflowTemplateRevisionConfigUpdated(templateID, alias string) string {
	return fmt.Sprintf(`
resource "stackguardian_workflow_template_revision" "test" {
  template_id = "%s"
  alias       = "%s"
  source_config_kind = "TERRAFORM"
  is_public   = "0"
  notes       = "Updated revision notes"
  tags        = ["test", "terraform", "updated"]
}
`, templateID, alias)
}

func testAccWorkflowTemplateRevisionConfigWithDetails(templateID, alias string) string {
	return fmt.Sprintf(`
resource "stackguardian_workflow_template_revision" "test" {
  template_id                  = "%s"
  alias                        = "%s"
  source_config_kind = "TERRAFORM"
  is_public                    = "0"
  notes                        = "Revision with detailed configuration"
  user_job_cpu                 = 2
  user_job_memory              = 4096
  number_of_approvals_required = 1
  tags                         = ["test", "terraform", "detailed"]
  approvers                    = ["approver1", "approver2"]
}
`, templateID, alias)
}
