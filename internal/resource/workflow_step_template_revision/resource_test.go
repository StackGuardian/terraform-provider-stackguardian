package workflowsteptemplaterevision_test

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
	"github.com/StackGuardian/sg-sdk-go/workflowsteptemplaterevision"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

var stepTemplateOrg = os.Getenv("STACKGUARDIAN_ORG_NAME")

func getStepTemplateTestClient() *sgclient.Client {
	customHeader := http.Header{}
	customHeader.Set("x-sg-internal-auth-orgid", "sg-provider-test")
	return sgclient.NewClient(
		sgoption.WithApiKey(fmt.Sprintf("apikey %s", os.Getenv("STACKGUARDIAN_API_KEY"))),
		sgoption.WithBaseURL(os.Getenv("STACKGUARDIAN_API_URI")),
		sgoption.WithHTTPHeader(customHeader),
	)
}

func deleteStepTemplateFixture(templateId string) {
	client := getStepTemplateTestClient()
	client.WorkflowStepTemplate.DeleteWorkflowStepTemplate(context.TODO(), stepTemplateOrg, templateId)
}

func deleteStepTemplateRevisionFixture(revisionId string) {
	client := getStepTemplateTestClient()
	client.WorkflowStepTemplateRevision.DeleteWorkflowStepTemplateRevision(context.TODO(), stepTemplateOrg, revisionId, true)
}

func deprecateStepTemplateRevisionFixture(revisionId string) {
	client := getStepTemplateTestClient()
	effectiveDate := fmt.Sprintf("%d", time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC).Unix())
	message := "This revision is deprecated"
	client.WorkflowStepTemplateRevision.UpdateWorkflowStepTemplateRevision(context.TODO(), stepTemplateOrg, revisionId, &workflowsteptemplaterevision.UpdateWorkflowStepTemplateRevisionModel{
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
	templateName := "provider-test-workflow-step-template1"
	revisionAlias := "v1"

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
	templateName := "tf-provider-step-template-lifecycle"
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
				Config: testAccStepTemplateRevisionLifecycleCreate(templateName, alias),
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
				Config: testAccStepTemplateRevisionLifecyclePublish(templateName, alias),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_step_template_revision.test", "is_public", "1"),
				),
			},
			// Step 3: Deprecate the revision
			{
				Config: testAccStepTemplateRevisionLifecycleDeprecate(templateName, alias),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_step_template_revision.test", "deprecation.message", "This revision is deprecated"),
					resource.TestCheckResourceAttr("stackguardian_workflow_step_template_revision.test", "deprecation.effective_date", fmt.Sprintf("%d", time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC).Unix())),
				),
			},
			// Terraform destroy automatically deletes both the revision and the parent step template
		},
	})
}

func testAccStepTemplateRevisionLifecycleCreate(templateName, alias string) string {
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

  runtime_source = {
    source_config_dest_kind = "CONTAINER_REGISTRY"
    config = {
      docker_image = "ubuntu:20.04"
      is_private   = false
    }
  }
}
`, templateName, alias)
}

func testAccStepTemplateRevisionLifecyclePublish(templateName, alias string) string {
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

  is_public          = "1"

  runtime_source = {
    source_config_dest_kind = "CONTAINER_REGISTRY"
    config = {
      docker_image = "ubuntu:20.04"
      is_private   = false
    }
  }
}
`, templateName, alias)
}

func testAccStepTemplateRevisionLifecycleDeprecate(templateName, alias string) string {
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


  runtime_source = {
    source_config_dest_kind = "CONTAINER_REGISTRY"
    config = {
      docker_image = "ubuntu:20.04"
      is_private   = false
    }
  }

  deprecation = {
    effective_date = "%d"
    message        = "This revision is deprecated"
  }
}
`, templateName, alias, time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC).Unix())
}
