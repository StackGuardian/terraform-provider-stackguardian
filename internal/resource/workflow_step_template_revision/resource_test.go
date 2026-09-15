package workflowsteptemplaterevision_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	sgsdkgo "github.com/StackGuardian/sg-sdk-go"
	sgclient "github.com/StackGuardian/sg-sdk-go/client"
	sgoption "github.com/StackGuardian/sg-sdk-go/option"
	"github.com/StackGuardian/sg-sdk-go/workflowsteptemplaterevision"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/acctest"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

var stepTemplateOrg = config.Get().OrgName

func getStepTemplateTestClient() *sgclient.Client {
	customHeader := http.Header{}
	customHeader.Set("x-sg-internal-auth-orgid", "sg-provider-test")
	cfg := config.Get()
	return sgclient.NewClient(
		sgoption.WithApiKey(cfg.FormatApiKey()),
		sgoption.WithBaseURL(cfg.ApiUri),
		sgoption.WithHTTPHeader(customHeader),
	)
}

func deleteStepTemplateFixture(templateId string) {
	client := getStepTemplateTestClient()
	_ = client.WorkflowStepTemplate.DeleteWorkflowStepTemplate(context.TODO(), stepTemplateOrg, templateId)
}

func deleteStepTemplateRevisionFixture(revisionId string) {
	client := getStepTemplateTestClient()
	_ = client.WorkflowStepTemplateRevision.DeleteWorkflowStepTemplateRevision(context.TODO(), stepTemplateOrg, revisionId, true)
}

// deprecateStepTemplateRevisionFixture deprecates a revision so it can be deleted (required
// when its is_public is "1"). The API call's result is ignored on purpose: if the revision
// was never published, deprecation isn't applicable and the call may fail — that's fine,
// since deletion doesn't need it in that case either.
func deprecateStepTemplateRevisionFixture(revisionId string) {
	client := getStepTemplateTestClient()
	effectiveDate := fmt.Sprintf("%d", time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC).Unix())
	message := "This revision is deprecated"
	_, _ = client.WorkflowStepTemplateRevision.UpdateWorkflowStepTemplateRevision(context.TODO(), stepTemplateOrg, revisionId, &workflowsteptemplaterevision.UpdateWorkflowStepTemplateRevisionModel{
		Deprecation: sgsdkgo.Optional(workflowsteptemplaterevision.Deprecation{
			EffectiveDate: &effectiveDate,
			Message:       &message,
		}),
	})
}

func testAccWorkflowStepTemplateRevisionConfig(templateName string, revisionAlias string) string {
	return fmt.Sprintf(`
resource "stackguardian_workflow_step_template" "test" {
  template_name = "%s"
  is_public     = "0"
  description   = "Test template for revision"

  source_config_kind = "DOCKER_IMAGE"

  runtime_source = {
    source_config_dest_kind = "CONTAINER_REGISTRY"
    config = {
      docker_image = "ubuntu:latest"
      is_private   = false
    }
  }
}

resource "stackguardian_workflow_step_template_revision" "test" {
  template_id = stackguardian_workflow_step_template.test.id
  alias       = "%s"
  notes       = "Test revision notes"

  source_config_kind = "DOCKER_IMAGE"

  runtime_source = {
    source_config_dest_kind = "CONTAINER_REGISTRY"
    config = {
      docker_image = "ubuntu:20.04"
      is_private   = false
    }
  }
}
`, templateName, revisionAlias)
}

func TestAccWorkflowStepTemplateRevision_Basic(t *testing.T) {
	templateName := acctest.ResourceName("provider-test-workflow-step-template1")
	revisionAlias := "v1"

	// Safety-net cleanup: Terraform's own destroy step tears these down in the normal case.
	// The revision must be deprecated before it can be deleted (required when its is_public
	// is "1", harmless otherwise), and it must be deleted before the parent template.
	t.Cleanup(func() {
		deprecateStepTemplateRevisionFixture(fmt.Sprintf("%s:1", templateName))
		deleteStepTemplateRevisionFixture(fmt.Sprintf("%s:1", templateName))
		deleteStepTemplateFixture(templateName)
	})

	testAccResource := testAccWorkflowStepTemplateRevisionConfig(templateName, revisionAlias)
	customHeader := http.Header{}
	customHeader.Set("x-sg-internal-auth-orgid", "sg-provider-test")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader),
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccResource,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_step_template_revision.test", "alias", revisionAlias),
					resource.TestCheckResourceAttr("stackguardian_workflow_step_template_revision.test", "notes", "Test revision notes"),
					resource.TestCheckResourceAttr("stackguardian_workflow_step_template_revision.test", "source_config_kind", "DOCKER_IMAGE"),
					resource.TestCheckResourceAttrSet("stackguardian_workflow_step_template_revision.test", "id"),
					resource.TestCheckResourceAttrSet("stackguardian_workflow_step_template_revision.test", "template_id"),
				),
			},
		},
	})
}

func TestAccWorkflowStepTemplateRevision_Lifecycle(t *testing.T) {
	templateName := acctest.ResourceName("tf-provider-step-template-lifecycle")
	alias := "v1"

	// Safety-net cleanup. Defers run LIFO, so registration order here is the
	// reverse of execution order:
	//   execution: deprecate revision → delete revision → delete parent template
	defer deleteStepTemplateFixture(templateName)
	defer deleteStepTemplateRevisionFixture(fmt.Sprintf("%s:1", templateName))
	defer deprecateStepTemplateRevisionFixture(fmt.Sprintf("%s:1", templateName))

	customHeader := http.Header{}
	customHeader.Set("x-sg-internal-auth-orgid", "sg-provider-test")

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader),
		Steps: []resource.TestStep{
			// Step 1: Create parent step template and revision
			{
				Config: testAccStepTemplateRevisionLifecycleConfig(templateName, alias, "", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_step_template.parent", "template_name", templateName),
					resource.TestCheckResourceAttrSet("stackguardian_workflow_step_template.parent", "id"),
					resource.TestCheckResourceAttr("stackguardian_workflow_step_template_revision.test", "alias", alias),
					resource.TestCheckResourceAttr("stackguardian_workflow_step_template_revision.test", "source_config_kind", "DOCKER_IMAGE"),
					resource.TestCheckResourceAttrSet("stackguardian_workflow_step_template_revision.test", "id"),
				),
			},
			// Step 2: Publish the revision
			{
				Config: testAccStepTemplateRevisionLifecycleConfig(templateName, alias, "1", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_step_template_revision.test", "is_public", "1"),
				),
			},
			// Step 3: Deprecate the revision
			{
				Config: testAccStepTemplateRevisionLifecycleConfig(templateName, alias, "1", true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_step_template_revision.test", "deprecation.message", "This revision is deprecated"),
					resource.TestCheckResourceAttr("stackguardian_workflow_step_template_revision.test", "deprecation.effective_date", fmt.Sprintf("%d", time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC).Unix())),
				),
			},
			// Terraform destroy automatically deletes both the revision and the parent step template
		},
	})
}

// testAccStepTemplateRevisionLifecycleConfig generates the lifecycle Terraform config.
// isPublic controls the revision's is_public value (pass "" to omit).
// deprecated=true adds a deprecation block to the revision.
func testAccStepTemplateRevisionLifecycleConfig(templateName, alias, isPublic string, deprecated bool) string {
	isPublicLine := ""
	if isPublic != "" {
		isPublicLine = fmt.Sprintf("  is_public          = %q\n", isPublic)
	}
	deprecationBlock := ""
	if deprecated {
		deprecationBlock = fmt.Sprintf(`
  deprecation = {
    effective_date = "%d"
    message        = "This revision is deprecated"
  }`, time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC).Unix())
	}
	return fmt.Sprintf(`
resource "stackguardian_workflow_step_template" "parent" {
  template_name      = "%s"
  is_public          = "0"
  source_config_kind = "DOCKER_IMAGE"

  runtime_source = {
    source_config_dest_kind = "CONTAINER_REGISTRY"
    config = {
      docker_image = "ubuntu:latest"
      is_private   = false
    }
  }
}

resource "stackguardian_workflow_step_template_revision" "test" {
  template_id        = stackguardian_workflow_step_template.parent.id
  alias              = "%s"
  notes              = "Initial revision"
  source_config_kind = "DOCKER_IMAGE"
%s
  runtime_source = {
    source_config_dest_kind = "CONTAINER_REGISTRY"
    config = {
      docker_image = "ubuntu:20.04"
      is_private   = false
    }
  }
%s
}
`, templateName, alias, isPublicLine, deprecationBlock)
}
