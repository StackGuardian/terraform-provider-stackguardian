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

// registerWorkflowTemplateCleanup registers a single t.Cleanup that tears down
// templateID and its revisions 1..revisionCount, in the order the API
// requires: a revision must be deprecated before it can be deleted (required
// when its is_public is "1", harmless otherwise), and every revision must be
// deleted before the parent template can be deleted.
func registerWorkflowTemplateCleanup(t *testing.T, templateID string, revisionCount int) {
	t.Helper()
	t.Cleanup(func() {
		for i := 1; i <= revisionCount; i++ {
			revisionID := fmt.Sprintf("%s:%d", templateID, i)
			deprecateWorkflowTemplateRevisionFixture(revisionID)
			deleteWorkflowTemplateRevisionFixture(revisionID)
		}
		deleteWorkflowTemplateFixture(templateID)
	})
}

// testAccWorkflowTemplateRevision wraps the resource's required attributes
// (template_id, source_config_kind, user_job_cpu, user_job_memory) and injects
// additionalConfig for everything else under test.
func testAccWorkflowTemplateRevision(templateID, sourceConfigKind string, userJobCPU, userJobMemory int, additionalConfig string) string {
	return fmt.Sprintf(`
resource "stackguardian_workflow_template_revision" "test" {
  template_id         = %q
  source_config_kind  = %q
  user_job_cpu        = %d
  user_job_memory     = %d

  %s
}
`, templateID, sourceConfigKind, userJobCPU, userJobMemory, additionalConfig)
}

func TestAccWorkflowTemplateRevision_Basic(t *testing.T) {
	templateID := "tf-provider-workflow-template-revision-1"
	alias := "revision1"

	registerWorkflowTemplateCleanup(t, templateID, 1)

	err := createWorkflowTemplateFixture(templateID, "TERRAFORM")
	if err != nil {
		t.Error(err)
	}

	customHeader := http.Header{}
	customHeader.Set("x-sg-internal-auth-orgid", "sg-provider-test")

	config := func(description, notes, tags string) string {
		return fmt.Sprintf(`
  alias       = %q
  is_public   = "0"
  description = %q
  notes       = %q
  tags        = %s
`, alias, description, notes, tags)
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
				Config: testAccWorkflowTemplateRevision(templateID, "TERRAFORM", 500, 1024, config("Initial description", "Initial notes", `["test", "terraform"]`)),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "template_id", templateID),
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "alias", alias),
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "source_config_kind", "TERRAFORM"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "is_public", "0"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "description", "Initial description"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "notes", "Initial notes"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "tags.0", "test"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "tags.1", "terraform"),
					resource.TestCheckResourceAttrSet("stackguardian_workflow_template_revision.test", "id"),
				),
			},
			// Update and Read testing
			{
				Config: testAccWorkflowTemplateRevision(templateID, "TERRAFORM", 500, 1024, config("Updated description", "Updated revision notes", `["test", "terraform", "updated"]`)),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "alias", alias),
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "description", "Updated description"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "notes", "Updated revision notes"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "tags.2", "updated"),
				),
			},
			// Delete testing automatically occurs
		},
	})
}

func TestAccWorkflowTemplateRevision_WithConfig(t *testing.T) {
	templateID := "test-workflow-template-revision"
	alias := "revision2"

	registerWorkflowTemplateCleanup(t, templateID, 1)

	err := createWorkflowTemplateFixture(templateID, "TERRAFORM")
	if err != nil {
		t.Fatal(err)
	}

	customHeader := http.Header{}
	customHeader.Set("x-sg-internal-auth-orgid", "sg-provider-test")

	config := fmt.Sprintf(`
  alias                        = %q
  is_public                    = "0"
  notes                        = "Revision with detailed configuration"
  number_of_approvals_required = 1
  tags                         = ["test", "terraform", "detailed"]
  approvers                    = ["approver1", "approver2"]
`, alias)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader),
		Steps: []resource.TestStep{
			// Create and Read testing with full configuration
			{
				Config: testAccWorkflowTemplateRevision(templateID, "TERRAFORM", 2, 4096, config),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "template_id", templateID),
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "alias", alias),
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "user_job_cpu", "2"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "user_job_memory", "4096"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "number_of_approvals_required", "1"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "approvers.0", "approver1"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "approvers.1", "approver2"),
				),
			},
			// Delete testing automatically occurs
		},
	})
}

func TestAccWorkflowTemplateRevision_WithDeploymentPlatformConfig(t *testing.T) {
	templateID := "test-workflow-template-revision-dpc"
	alias := "revision-dpc"

	registerWorkflowTemplateCleanup(t, templateID, 1)

	err := createWorkflowTemplateFixture(templateID, "TERRAFORM")
	if err != nil {
		t.Fatal(err)
	}

	customHeader := http.Header{}
	customHeader.Set("x-sg-internal-auth-orgid", "sg-provider-test")

	config := fmt.Sprintf(`
  alias     = %q
  is_public = "0"

  deployment_platform_config = [{
    kind = "AWS_RBAC"
    config = {
      integration_id = "/integrations/test-integration"
      profile_name   = "test-profile"
    }
  }]
`, alias)

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
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "template_id", templateID),
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "alias", alias),
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "deployment_platform_config.0.kind", "AWS_RBAC"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "deployment_platform_config.0.config.integration_id", "/integrations/test-integration"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "deployment_platform_config.0.config.profile_name", "test-profile"),
				),
			},
		},
	})
}

func TestAccWorkflowTemplateRevision_WithEnvironmentVariablesAndContextTags(t *testing.T) {
	templateID := "tf-provider-wftr-envvars"
	alias := "revision-envvars"

	registerWorkflowTemplateCleanup(t, templateID, 1)

	err := createWorkflowTemplateFixture(templateID, "TERRAFORM")
	if err != nil {
		t.Fatal(err)
	}

	customHeader := http.Header{}
	customHeader.Set("x-sg-internal-auth-orgid", "sg-provider-test")

	config := func(textValue, ctxVal string) string {
		return fmt.Sprintf(`
  alias = %q

  environment_variables = [
    {
      kind = "PLAIN_TEXT"
      config = {
        var_name   = "MY_VAR"
        text_value = %q
      }
    }
  ]

  context_tags = {
    env = %q
  }
`, alias, textValue, ctxVal)
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowTemplateRevision(templateID, "TERRAFORM", 500, 1024, config("initial-value", "staging")),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "environment_variables.0.kind", "PLAIN_TEXT"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "environment_variables.0.config.var_name", "MY_VAR"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "environment_variables.0.config.text_value", "initial-value"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "context_tags.env", "staging"),
				),
			},
			{
				Config: testAccWorkflowTemplateRevision(templateID, "TERRAFORM", 500, 1024, config("updated-value", "production")),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "environment_variables.0.config.text_value", "updated-value"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "context_tags.env", "production"),
				),
			},
		},
	})
}

func TestAccWorkflowTemplateRevision_WithInputSchemas(t *testing.T) {
	templateID := "tf-provider-wftr-inputschemas"
	alias := "revision-inputschemas"

	registerWorkflowTemplateCleanup(t, templateID, 1)

	err := createWorkflowTemplateFixture(templateID, "TERRAFORM")
	if err != nil {
		t.Fatal(err)
	}

	customHeader := http.Header{}
	customHeader.Set("x-sg-internal-auth-orgid", "sg-provider-test")

	config := func(name string) string {
		return fmt.Sprintf(`
  alias = %q

  input_schemas = [
    {
      name           = %q
      type           = "RAW_JSON"
      encoded_data   = "eyJmb28iOiJiYXIifQ=="
      ui_schema_data = "eyJ1aSI6dHJ1ZX0="
    }
  ]
`, alias, name)
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowTemplateRevision(templateID, "TERRAFORM", 500, 1024, config("input-1")),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "input_schemas.0.name", "input-1"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "input_schemas.0.type", "RAW_JSON"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "input_schemas.0.encoded_data", "eyJmb28iOiJiYXIifQ=="),
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "input_schemas.0.ui_schema_data", "eyJ1aSI6dHJ1ZX0="),
				),
			},
			{
				Config: testAccWorkflowTemplateRevision(templateID, "TERRAFORM", 500, 1024, config("input-1-renamed")),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "input_schemas.0.name", "input-1-renamed"),
				),
			},
		},
	})
}

func TestAccWorkflowTemplateRevision_WithUserSchedules(t *testing.T) {
	templateID := "tf-provider-wftr-schedules"
	alias := "revision-schedules"

	registerWorkflowTemplateCleanup(t, templateID, 1)

	err := createWorkflowTemplateFixture(templateID, "TERRAFORM")
	if err != nil {
		t.Fatal(err)
	}

	customHeader := http.Header{}
	customHeader.Set("x-sg-internal-auth-orgid", "sg-provider-test")

	config := func(cron string) string {
		return fmt.Sprintf(`
  alias = %q

  user_schedules = [
    {
      cron  = %q
      state = "ENABLED"
      desc  = "Runs on schedule"
      name  = "weekly"
    }
  ]
`, alias, cron)
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowTemplateRevision(templateID, "TERRAFORM", 500, 1024, config("0 8 ? * MON *")),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "user_schedules.0.cron", "0 8 ? * MON *"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "user_schedules.0.state", "ENABLED"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "user_schedules.0.desc", "Runs on schedule"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "user_schedules.0.name", "weekly"),
				),
			},
			{
				Config: testAccWorkflowTemplateRevision(templateID, "TERRAFORM", 500, 1024, config("0 9 ? * MON *")),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "user_schedules.0.cron", "0 9 ? * MON *"),
				),
			},
		},
	})
}

func TestAccWorkflowTemplateRevision_WithRunnerConstraints(t *testing.T) {
	templateID := "tf-provider-wftr-runner"
	alias := "revision-runner"

	registerWorkflowTemplateCleanup(t, templateID, 1)

	err := createWorkflowTemplateFixture(templateID, "TERRAFORM")
	if err != nil {
		t.Fatal(err)
	}

	customHeader := http.Header{}
	customHeader.Set("x-sg-internal-auth-orgid", "sg-provider-test")

	configShared := fmt.Sprintf(`
  alias = %q

  runner_constraints = {
    type = "shared"
  }
`, alias)

	configWithNames := fmt.Sprintf(`
  alias = %q

  runner_constraints = {
    type  = "private"
    names = ["runner-1"]
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
				Config: testAccWorkflowTemplateRevision(templateID, "TERRAFORM", 500, 1024, configShared),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "runner_constraints.type", "shared"),
				),
			},
			{
				Config: testAccWorkflowTemplateRevision(templateID, "TERRAFORM", 500, 1024, configWithNames),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "runner_constraints.type", "private"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "runner_constraints.names.0", "runner-1"),
				),
			},
		},
	})
}

func TestAccWorkflowTemplateRevision_WithMiniSteps(t *testing.T) {
	templateID := "tf-provider-wftr-ministeps"
	alias := "revision-ministeps"

	registerWorkflowTemplateCleanup(t, templateID, 1)

	err := createWorkflowTemplateFixture(templateID, "TERRAFORM")
	if err != nil {
		t.Fatal(err)
	}

	customHeader := http.Header{}
	customHeader.Set("x-sg-internal-auth-orgid", "sg-provider-test")

	config := func(payloadExpr string) string {
		return fmt.Sprintf(`
  alias = %q

  mini_steps = {
    notifications = {
      email = {
        errored           = [{ recipients = ["oncall@example.com"] }]
        approval_required = [{ recipients = ["approver@example.com"] }]
      }
    }
    webhooks = {
      errored = [{
        webhook_name = "notify-errored"
        webhook_url  = "https://example.com/hook"
      }]
      completed = [{
        webhook_name   = "notify-completed"
        webhook_url    = "https://example.com/hook-completed"
        webhook_secret = "shh"
      }]
    }
    wf_chaining = {
      errored = [{
        workflow_group_id    = "kk"
        workflow_id          = "retest-of-bug-cewgh6dt-i7vp-0vo474rl"
        workflow_run_payload = %s
      }]
      completed = [{
        workflow_group_id = "kk"
        stack_id          = "stack-1"
        stack_run_payload = %s
      }]
    }
  }
`, alias, payloadExpr, payloadExpr)
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowTemplateRevision(templateID, "TERRAFORM", 500, 1024, config(`jsonencode({"test" = "value"})`)),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "mini_steps.notifications.email.errored.0.recipients.0", "oncall@example.com"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "mini_steps.notifications.email.approval_required.0.recipients.0", "approver@example.com"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "mini_steps.webhooks.errored.0.webhook_name", "notify-errored"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "mini_steps.webhooks.errored.0.webhook_url", "https://example.com/hook"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "mini_steps.webhooks.completed.0.webhook_name", "notify-completed"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "mini_steps.webhooks.completed.0.webhook_secret", "shh"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "mini_steps.wf_chaining.errored.0.workflow_group_id", "kk"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "mini_steps.wf_chaining.errored.0.workflow_id", "retest-of-bug-cewgh6dt-i7vp-0vo474rl"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "mini_steps.wf_chaining.errored.0.workflow_run_payload", `{"test":"value"}`),
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "mini_steps.wf_chaining.completed.0.stack_id", "stack-1"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "mini_steps.wf_chaining.completed.0.stack_run_payload", `{"test":"value"}`),
				),
			},
			{
				Config: testAccWorkflowTemplateRevision(templateID, "TERRAFORM", 500, 1024, config(`jsonencode({"test" = "updated"})`)),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "mini_steps.wf_chaining.errored.0.workflow_run_payload", `{"test":"updated"}`),
				),
			},
		},
	})
}

func TestAccWorkflowTemplateRevision_WithRuntimeSource(t *testing.T) {
	templateID := "tf-provider-wftr-runtime"
	alias := "revision-runtime"

	registerWorkflowTemplateCleanup(t, templateID, 1)

	err := createWorkflowTemplateFixture(templateID, "TERRAFORM")
	if err != nil {
		t.Fatal(err)
	}

	customHeader := http.Header{}
	customHeader.Set("x-sg-internal-auth-orgid", "sg-provider-test")

	config := func(ref string) string {
		return fmt.Sprintf(`
  alias = %q

  runtime_source = {
    source_config_dest_kind = "GITHUB_COM"
    config = {
      is_private                 = false
      auth                       = "/integrations/tf-provider-test-connector"
      repo                       = "https://github.com/taherkk/taher-null-resource.git"
      ref                        = %q
      working_dir                = "src"
      include_sub_module         = true
      git_core_auto_crlf         = true
      git_sparse_checkout_config = "--no-cone infra/**"
    }
  }
`, alias, ref)
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowTemplateRevision(templateID, "TERRAFORM", 500, 1024, config("main")),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "runtime_source.source_config_dest_kind", "GITHUB_COM"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "runtime_source.config.repo", "https://github.com/taherkk/taher-null-resource.git"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "runtime_source.config.is_private", "false"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "runtime_source.config.auth", "/integrations/tf-provider-test-connector"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "runtime_source.config.ref", "main"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "runtime_source.config.working_dir", "src"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "runtime_source.config.include_sub_module", "true"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "runtime_source.config.git_core_auto_crlf", "true"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "runtime_source.config.git_sparse_checkout_config", "--no-cone infra/**"),
				),
			},
			{
				Config: testAccWorkflowTemplateRevision(templateID, "TERRAFORM", 500, 1024, config("develop")),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "runtime_source.config.ref", "develop"),
				),
			},
		},
	})
}

func TestAccWorkflowTemplateRevision_WithTerraformConfig(t *testing.T) {
	templateID := "tf-provider-wftr-tfconfig"
	alias := "revision-tfconfig"

	registerWorkflowTemplateCleanup(t, templateID, 1)

	err := createWorkflowTemplateFixture(templateID, "TERRAFORM")
	if err != nil {
		t.Fatal(err)
	}

	customHeader := http.Header{}
	customHeader.Set("x-sg-internal-auth-orgid", "sg-provider-test")

	config := func(tfVersion string) string {
		return fmt.Sprintf(`
  alias = %q

  terraform_config = {
    terraform_version           = %q
    drift_check                 = true
    drift_cron                  = "0 0 * * *"
    managed_terraform_state     = true
    approval_pre_apply          = true
    terraform_plan_options      = "-lock=false"
    terraform_init_options      = "-upgrade"
    timeout                     = 3600
    run_pre_init_hooks_on_drift = true
    pre_init_hooks              = ["echo pre-init"]
    pre_plan_hooks              = ["echo pre-plan"]
    post_plan_hooks             = ["echo post-plan"]
    pre_apply_hooks             = ["echo pre-apply"]
    post_apply_hooks            = ["echo post-apply"]

    terraform_bin_path = [{
      source = "/usr/local/bin/terraform"
      target = "/usr/bin/terraform"
    }]

    pre_apply_wf_steps_config = [{
      name                = "pre-apply-step"
      wf_step_template_id = "/tf-provider-test-org/dummy-step-template:1"
    }]
  }
`, alias, tfVersion)
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowTemplateRevision(templateID, "TERRAFORM", 500, 1024, config("1.5.0")),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "terraform_config.terraform_version", "1.5.0"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "terraform_config.drift_check", "true"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "terraform_config.drift_cron", "0 0 * * *"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "terraform_config.managed_terraform_state", "true"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "terraform_config.approval_pre_apply", "true"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "terraform_config.terraform_plan_options", "-lock=false"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "terraform_config.terraform_init_options", "-upgrade"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "terraform_config.timeout", "3600"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "terraform_config.run_pre_init_hooks_on_drift", "true"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "terraform_config.pre_init_hooks.0", "echo pre-init"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "terraform_config.pre_plan_hooks.0", "echo pre-plan"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "terraform_config.post_plan_hooks.0", "echo post-plan"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "terraform_config.pre_apply_hooks.0", "echo pre-apply"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "terraform_config.post_apply_hooks.0", "echo post-apply"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "terraform_config.terraform_bin_path.0.source", "/usr/local/bin/terraform"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "terraform_config.terraform_bin_path.0.target", "/usr/bin/terraform"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "terraform_config.pre_apply_wf_steps_config.0.name", "pre-apply-step"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "terraform_config.pre_apply_wf_steps_config.0.wf_step_template_id", "/tf-provider-test-org/dummy-step-template:1"),
				),
			},
			{
				Config: testAccWorkflowTemplateRevision(templateID, "TERRAFORM", 500, 1024, config("1.5.7")),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "terraform_config.terraform_version", "1.5.7"),
				),
			},
		},
	})
}

func TestAccWorkflowTemplateRevision_WithWfStepsConfig(t *testing.T) {
	templateID := "tf-provider-wftr-wfsteps"
	alias := "revision-wfsteps"

	registerWorkflowTemplateCleanup(t, templateID, 1)

	// wf_steps_config is only registered by the API when the template's
	// source_config_kind is not TERRAFORM or OPENTOFU.
	err := createWorkflowTemplateFixture(templateID, "CUSTOM")
	if err != nil {
		t.Fatal(err)
	}

	customHeader := http.Header{}
	customHeader.Set("x-sg-internal-auth-orgid", "sg-provider-test")

	config := func(cmdOverride string) string {
		return fmt.Sprintf(`
  alias = %q

  wf_steps_config = [
    {
      name                = "step-1"
      wf_step_template_id = "/tf-provider-test-org/dummy-step-template:1"
      cmd_override        = %q
      timeout             = 600
      approval            = false

      environment_variables = [
        {
          kind = "PLAIN_TEXT"
          config = {
            var_name   = "STEP_VAR"
            text_value = "step-value"
          }
        }
      ]

      mount_points = [
        {
          source    = "/data"
          target    = "/mnt/data"
          read_only = true
        }
      ]

      wf_step_input_data = {
        schema_type = "FORM_JSONSCHEMA"
        data        = jsonencode({ "foo" = "bar" })
      }
    }
  ]
`, alias, cmdOverride)
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowTemplateRevision(templateID, "CUSTOM", 500, 1024, config("echo initial")),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "wf_steps_config.0.name", "step-1"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "wf_steps_config.0.wf_step_template_id", "/tf-provider-test-org/dummy-step-template:1"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "wf_steps_config.0.cmd_override", "echo initial"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "wf_steps_config.0.timeout", "600"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "wf_steps_config.0.approval", "false"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "wf_steps_config.0.environment_variables.0.config.var_name", "STEP_VAR"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "wf_steps_config.0.environment_variables.0.config.text_value", "step-value"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "wf_steps_config.0.mount_points.0.source", "/data"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "wf_steps_config.0.mount_points.0.target", "/mnt/data"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "wf_steps_config.0.mount_points.0.read_only", "true"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "wf_steps_config.0.wf_step_input_data.schema_type", "FORM_JSONSCHEMA"),
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "wf_steps_config.0.wf_step_input_data.data", `{"foo":"bar"}`),
				),
			},
			{
				Config: testAccWorkflowTemplateRevision(templateID, "CUSTOM", 500, 1024, config("echo updated")),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_template_revision.test", "wf_steps_config.0.cmd_override", "echo updated"),
				),
			},
		},
	})
}

func TestAccWorkflowTemplateRevision_Lifecycle(t *testing.T) {
	templateName := "tf-provider-wf-template-lifecycle"
	alias := "v1"

	// Safety-net cleanup: deprecates then deletes the revision, then deletes
	// the parent template.
	registerWorkflowTemplateCleanup(t, templateName, 1)

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
