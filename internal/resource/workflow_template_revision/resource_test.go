package workflowtemplaterevision_test

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
	"github.com/StackGuardian/sg-sdk-go/workflowtemplaterevisions"
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
		Id:               &templateName,
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

func deprecateWorkflowTemplateRevisionFixture(revisionId string) {
	client := GetClient()
	effectiveDate := fmt.Sprintf("%d", time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC).Unix())
	message := "This revision is deprecated"
	client.WorkflowTemplatesRevisions.UpdateWorkflowTemplateRevision(context.TODO(), org, revisionId, &workflowtemplaterevisions.UpdateWorkflowTemplateRevisionRequest{
		Deprecation: sgsdkgo.Optional(workflowtemplaterevisions.Deprecation{
			EffectiveDate: &effectiveDate,
			Message:       &message,
		}),
	})
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
				Config: testAccWorkflowTemplateRevisionConfig(templateID, alias, "", `["test", "terraform"]`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "template_id", templateID),
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "alias", alias),
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "is_public", "0"),
					resource.TestCheckResourceAttrSet("stackguardian_workflow_template_revision.test", "id"),
				),
			},
			// Update and Read testing
			{
				Config: testAccWorkflowTemplateRevisionConfig(templateID, alias, "Updated revision notes", `["test", "terraform", "updated"]`),
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

// testAccWorkflowTemplateRevisionConfig returns a config for a basic revision.
// notes is optional (pass "" to omit). tags is the HCL list literal, e.g. `["test", "terraform"]`.
func testAccWorkflowTemplateRevisionConfig(templateID, alias, notes, tags string) string {
	notesLine := ""
	if notes != "" {
		notesLine = fmt.Sprintf("  notes              = %q\n", notes)
	}
	return fmt.Sprintf(`
resource "stackguardian_workflow_template_revision" "test" {
  template_id        = "%s"
  alias              = "%s"
  source_config_kind = "TERRAFORM"
  is_public          = "0"
  user_job_cpu       = 500
  user_job_memory    = 1024
  tags               = %s
%s}
`, templateID, alias, tags, notesLine)
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

func TestAccWorkflowTemplateRevision_Lifecycle(t *testing.T) {
	templateName := "tf-provider-wf-template-lifecycle"
	alias := "v1"

	// Safety-net cleanup. Defers run LIFO, so registration order here is the
	// reverse of execution order:
	//   execution: deprecate revision → delete revision → delete parent template
	defer deleteWorkflowTemplateFixture(templateName)
	defer deleteWorkflowTemplateRevisionFixture(fmt.Sprintf("%s:1", templateName))
	defer deprecateWorkflowTemplateRevisionFixture(fmt.Sprintf("%s:1", templateName))

	customHeader := http.Header{}
	customHeader.Set("x-sg-internal-auth-orgid", "sg-provider-test")

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader),
		Steps: []resource.TestStep{
			// Step 1: Create parent template and revision
			{
				Config: testAccWfTemplateRevisionLifecycleConfig(templateName, alias, "0", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_template.parent", "template_name", templateName),
					resource.TestCheckResourceAttrSet("stackguardian_workflow_template.parent", "id"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "alias", alias),
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "is_public", "0"),
					resource.TestCheckResourceAttrSet("stackguardian_workflow_template_revision.test", "id"),
				),
			},
			// Step 2: Publish the revision
			{
				Config: testAccWfTemplateRevisionLifecycleConfig(templateName, alias, "1", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "is_public", "1"),
				),
			},
			// Step 3: Deprecate the revision
			{
				Config: testAccWfTemplateRevisionLifecycleConfig(templateName, alias, "1", true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "is_public", "1"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "deprecation.message", "This revision is deprecated"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "deprecation.effective_date", fmt.Sprintf("%d", time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC).Unix())),
				),
			},
			// Terraform destroy automatically deletes both the revision and the parent template
		},
	})
}

// testAccWfTemplateRevisionLifecycleConfig generates the lifecycle Terraform config.
// isPublic controls the revision's is_public value.
// deprecated=true adds a deprecation block to the revision.
func testAccWfTemplateRevisionLifecycleConfig(templateName, alias, isPublic string, deprecated bool) string {
	deprecationBlock := ""
	if deprecated {
		deprecationBlock = fmt.Sprintf(`
  deprecation = {
    effective_date = "%d"
    message        = "This revision is deprecated"
  }`, time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC).Unix())
	}
	return fmt.Sprintf(`
resource "stackguardian_workflow_template" "parent" {
  template_name      = "%s"
  source_config_kind = "TERRAFORM"
  is_public          = "0"
  tags               = ["test", "lifecycle"]
}

resource "stackguardian_workflow_template_revision" "test" {
  template_id        = stackguardian_workflow_template.parent.id
  alias              = "%s"
  source_config_kind = "TERRAFORM"
  is_public          = "%s"
  user_job_cpu       = 500
  user_job_memory    = 1024
  tags               = ["test", "lifecycle"]
%s
}
`, templateName, alias, isPublic, deprecationBlock)
}
