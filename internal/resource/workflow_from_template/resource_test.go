package workflowfromtemplate_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	sgsdkgo "github.com/StackGuardian/sg-sdk-go"
	sgclient "github.com/StackGuardian/sg-sdk-go/client"
	"github.com/StackGuardian/sg-sdk-go/core"
	sgoption "github.com/StackGuardian/sg-sdk-go/option"
	sgworkflows "github.com/StackGuardian/sg-sdk-go/workflows"
	"github.com/StackGuardian/sg-sdk-go/workflowtemplaterevisions"
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

func createWorkflowGroupFixture(wfGrpName string) error {
	client := getClient()
	parts := strings.Split(wfGrpName, "/")
	leafName := parts[len(parts)-1]
	payload := &sgsdkgo.WorkflowGroup{
		ResourceName: &leafName,
	}
	if len(parts) > 1 {
		parent := strings.Join(parts[:len(parts)-1], "/")
		_, err := client.WorkflowGroups.CreateChildWorkflowGroup(context.TODO(), org, parent, payload)
		return err
	}
	_, err := client.WorkflowGroups.CreateWorkflowGroup(context.TODO(), org, payload)
	return err
}

func deleteWorkflowGroupFixture(wfGrpName string) {
	client := getClient()
	client.WorkflowGroups.DeleteWorkflowGroup(context.TODO(), org, wfGrpName)
}

func deleteWorkflowUsingTemplateFixture(wfGrpName, workflowName string) {
	client := getClient()
	client.Workflows.DeleteWorkflow(context.TODO(), org, workflowName, wfGrpName)
}

// setupWorkflowTemplate creates a workflow template + published revision and registers
// cleanup that deprecates, deletes the revision, then deletes the template.
// Returns the template ID to use as iac_template_id.
func setupWorkflowTemplate(t *testing.T, templateID string) string {
	t.Helper()

	client := getClient()
	revisionID := fmt.Sprintf("%s:1", templateID)
	sourceConfigKind := workflowtemplates.WorkflowTemplateSourceConfigKindTerraform

	is409 := func(err error) bool {
		var apiErr *core.APIError
		return errors.As(err, &apiErr) && apiErr.StatusCode == 409
	}

	// 1. Create template (ignore 409 — already exists)
	_, err := client.WorkflowTemplates.CreateWorkflowTemplate(
		context.TODO(), org, false,
		&workflowtemplates.CreateWorkflowTemplateRequest{
			Id:               &templateID,
			TemplateName:     templateID,
			SourceConfigKind: &sourceConfigKind,
			TemplateType:     sgsdkgo.TemplateTypeEnum("IAC"),
			IsPublic:         sgsdkgo.IsPublicEnumZero.Ptr(),
			OwnerOrg:         fmt.Sprintf("/orgs/%s", org),
		},
	)
	if err != nil && !is409(err) {
		t.Fatalf("setupWorkflowTemplate: create template %q: %s", templateID, err)
	}

	// 2. Create revision with a default env var and user schedule (ignore 409 — already exists)
	alias := "v1"
	scheduleName := "tmpl-schedule"
	scheduleDesc := "Template default schedule"
	envTextValue := "tmpl-value"
	tmplTfVersion := "1.5.0"
	_, err = client.WorkflowTemplatesRevisions.CreateWorkflowTemplateRevision(
		context.TODO(), org, templateID,
		&workflowtemplaterevisions.CreateWorkflowTemplateRevisionsRequest{
			Alias:            alias,
			SourceConfigKind: &sourceConfigKind,
			IsPublic:         sgsdkgo.IsPublicEnumZero.Ptr(),
			OwnerOrg:         fmt.Sprintf("/orgs/%s", org),
			TerraformConfig: &sgsdkgo.TerraformConfig{
				TerraformVersion: &tmplTfVersion,
			},
			EnvironmentVariables: []sgsdkgo.EnvVars{
				{
					Kind: sgsdkgo.EnvVarsKindEnumPlainText,
					Config: &sgsdkgo.EnvVarConfig{
						VarName:   "TMPL_VAR",
						TextValue: &envTextValue,
					},
				},
			},
			UserSchedules: []workflowtemplaterevisions.UserSchedules{
				{
					Cron:  "0 8 ? * MON *",
					State: workflowtemplaterevisions.UserSchedulesStateEnumEnabled,
					Name:  &scheduleName,
					Desc:  &scheduleDesc,
				},
			},
		},
	)
	if err != nil && !is409(err) {
		client.WorkflowTemplates.DeleteWorkflowTemplate(context.TODO(), org, templateID)
		t.Fatalf("setupWorkflowTemplate: create revision for %q: %s", templateID, err)
	}

	// 3. Publish revision
	_, err = client.WorkflowTemplatesRevisions.UpdateWorkflowTemplateRevision(
		context.TODO(), org, revisionID,
		&workflowtemplaterevisions.UpdateWorkflowTemplateRevisionRequest{
			IsPublic: sgsdkgo.Optional(sgsdkgo.IsPublicEnumOne),
		},
	)
	if err != nil {
		client.WorkflowTemplatesRevisions.DeleteWorkflowTemplateRevision(context.TODO(), org, revisionID, true)
		client.WorkflowTemplates.DeleteWorkflowTemplate(context.TODO(), org, templateID)
		t.Fatalf("setupWorkflowTemplate: publish revision %q: %s", revisionID, err)
	}

	// 4. Publish template
	_, err = client.WorkflowTemplates.UpdateWorkflowTemplate(
		context.TODO(), org, templateID,
		&workflowtemplates.UpdateWorkflowTemplateRequest{
			IsPublic: sgsdkgo.Optional(sgsdkgo.IsPublicEnumOne),
		},
	)
	if err != nil {
		client.WorkflowTemplatesRevisions.DeleteWorkflowTemplateRevision(context.TODO(), org, revisionID, true)
		client.WorkflowTemplates.DeleteWorkflowTemplate(context.TODO(), org, templateID)
		t.Fatalf("setupWorkflowTemplate: publish template %q: %s", templateID, err)
	}

	// 5. Register cleanup: deprecate revision → delete revision → delete template
	t.Cleanup(func() {
		effectiveDate := fmt.Sprintf("%d", time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC).Unix())
		message := "Test cleanup"
		client.WorkflowTemplatesRevisions.UpdateWorkflowTemplateRevision(
			context.TODO(), org, revisionID,
			&workflowtemplaterevisions.UpdateWorkflowTemplateRevisionRequest{
				Deprecation: sgsdkgo.Optional(workflowtemplaterevisions.Deprecation{
					EffectiveDate: &effectiveDate,
					Message:       &message,
				}),
			},
		)
		client.WorkflowTemplatesRevisions.DeleteWorkflowTemplateRevision(context.TODO(), org, revisionID, true)
		client.WorkflowTemplates.DeleteWorkflowTemplate(context.TODO(), org, templateID)
	})

	return fmt.Sprintf("/%s/%s", org, templateID)
}

func customHeader() http.Header {
	h := http.Header{}
	h.Set("x-sg-internal-auth-orgid", "sg-provider-test")
	return h
}

// testAccWorkflowUsingTemplate builds a Terraform config for the resource.
func testAccWorkflowUsingTemplate(wfGrpName, id, wfType, templateID, additionalConfig string) string {
	return fmt.Sprintf(`
resource "stackguardian_workflow_from_template" "test" {
  workflow_group_id = %q
  id                = %q
  wf_type           = %q

  vcs_config = {
    iac_vcs_config = {
      iac_template_id = %q
    }
  }

  %s
}
`, wfGrpName, id, wfType, templateID, additionalConfig)
}

func TestAccWorkflowUsingTemplate_Basic(t *testing.T) {
	templateID := setupWorkflowTemplate(t, "tf-provider-wf-tmpl-basic") + ":1"
	wfGrpName := "tf-provider-wf-template-basic-wfgrp"
	id := "tf-provider-wf-template-basic"

	if err := createWorkflowGroupFixture(wfGrpName); err != nil {
		t.Errorf("failed to create workflow group fixture: %s", err.Error())
	}
	defer deleteWorkflowGroupFixture(wfGrpName)
	defer deleteWorkflowUsingTemplateFixture(wfGrpName, id)

	// No overrides — template defaults should be resolved onto the top-level attributes.
	configNoOverrides := fmt.Sprintf(`
  terraform_config = {
    terraform_version = %q
  }
`, "1.5.0")

	// Override the env var set on the revision.
	configWithOverrides := fmt.Sprintf(`
  terraform_config = {
    terraform_version = %q
  }

  environment_variables = [
    {
      kind = "PLAIN_TEXT"
      config = {
        var_name   = "OVERRIDE_VAR"
        text_value = "override-value"
      }
    }
  ]
`, "1.5.0")

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader()),
		Steps: []resource.TestStep{
			{
				// Step 1: no overrides — top-level attributes should reflect template revision defaults.
				Config: testAccWorkflowUsingTemplate(wfGrpName, id, "TERRAFORM", templateID, configNoOverrides),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test", "workflow_group_id", wfGrpName),
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test", "id", id),
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test", "wf_type", "TERRAFORM"),
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test", "vcs_config.iac_vcs_config.iac_template_id", templateID),
					resource.TestCheckResourceAttrSet("stackguardian_workflow_from_template.test", "resource_name"),
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test", "environment_variables.0.config.var_name", "TMPL_VAR"),
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test", "environment_variables.0.config.text_value", "tmpl-value"),
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test", "user_schedules.0.cron", "0 8 ? * MON *"),
				),
			},
			{
				// Step 2: override env var — top-level attribute should reflect the override.
				Config: testAccWorkflowUsingTemplate(wfGrpName, id, "TERRAFORM", templateID, configWithOverrides),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test", "id", id),
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test", "environment_variables.0.config.var_name", "OVERRIDE_VAR"),
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test", "environment_variables.0.config.text_value", "override-value"),
				),
			},
		},
	})
}

func TestAccWorkflowUsingTemplate_WithDescription(t *testing.T) {
	templateID := setupWorkflowTemplate(t, "tf-provider-wf-tmpl-desc") + ":1"
	wfGrpName := "tf-provider-wf-template-desc-wfgrp"
	id := "tf-provider-wf-template-desc"

	if err := createWorkflowGroupFixture(wfGrpName); err != nil {
		t.Errorf("failed to create workflow group fixture: %s", err.Error())
	}
	defer deleteWorkflowGroupFixture(wfGrpName)
	defer deleteWorkflowUsingTemplateFixture(wfGrpName, id)

	config := func(description string) string {
		return fmt.Sprintf(`
  description = %q

  terraform_config = {
    terraform_version = %q
  }
`, description, "1.5.0")
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader()),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowUsingTemplate(wfGrpName, id, "TERRAFORM", templateID, config("initial description")),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test", "id", id),
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test", "description", "initial description"),
				),
			},
			{
				Config: testAccWorkflowUsingTemplate(wfGrpName, id, "TERRAFORM", templateID, config("updated description")),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test", "id", id),
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test", "description", "updated description"),
				),
			},
		},
	})
}

func TestAccWorkflowUsingTemplate_WithTerraformConfig(t *testing.T) {
	templateID := setupWorkflowTemplate(t, "tf-provider-wf-tmpl-tfcfg") + ":1"
	wfGrpName := "tf-provider-wf-template-tfcfg-wfgrp"
	id := "tf-provider-wf-template-tfcfg"

	if err := createWorkflowGroupFixture(wfGrpName); err != nil {
		t.Errorf("failed to create workflow group fixture: %s", err.Error())
	}
	defer deleteWorkflowGroupFixture(wfGrpName)
	defer deleteWorkflowUsingTemplateFixture(wfGrpName, id)

	config := func(tfVersion string) string {
		return fmt.Sprintf(`
  terraform_config = {
    terraform_version = %q
  }
`, tfVersion)
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader()),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowUsingTemplate(wfGrpName, id, "TERRAFORM", templateID, config("1.5.0")),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test", "id", id),
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test", "terraform_config.terraform_version", "1.5.0"),
				),
			},
			{
				Config: testAccWorkflowUsingTemplate(wfGrpName, id, "TERRAFORM", templateID, config("1.6.0")),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test", "id", id),
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test", "terraform_config.terraform_version", "1.6.0"),
				),
			},
		},
	})
}

func TestAccWorkflowUsingTemplate_WithEnvironmentVariables(t *testing.T) {
	templateID := setupWorkflowTemplate(t, "tf-provider-wf-tmpl-envvars") + ":1"
	wfGrpName := "tf-provider-wf-template-envvars-wfgrp"
	id := "tf-provider-wf-template-envvars"

	if err := createWorkflowGroupFixture(wfGrpName); err != nil {
		t.Errorf("failed to create workflow group fixture: %s", err.Error())
	}
	defer deleteWorkflowGroupFixture(wfGrpName)
	defer deleteWorkflowUsingTemplateFixture(wfGrpName, id)

	config := func(textValue string) string {
		return fmt.Sprintf(`
  terraform_config = {
    terraform_version = %q
  }

  environment_variables = [
    {
      kind = "PLAIN_TEXT"
      config = {
        var_name   = "MY_VAR"
        text_value = %q
      }
    }
  ]
`, "1.5.0", textValue)
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader()),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowUsingTemplate(wfGrpName, id, "TERRAFORM", templateID, config("initial-value")),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test", "id", id),
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test", "environment_variables.0.kind", "PLAIN_TEXT"),
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test", "environment_variables.0.config.var_name", "MY_VAR"),
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test", "environment_variables.0.config.text_value", "initial-value"),
				),
			},
			{
				Config: testAccWorkflowUsingTemplate(wfGrpName, id, "TERRAFORM", templateID, config("updated-value")),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test", "id", id),
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test", "environment_variables.0.config.text_value", "updated-value"),
				),
			},
		},
	})
}

func TestAccWorkflowUsingTemplate_WithUserSchedules(t *testing.T) {
	templateID := setupWorkflowTemplate(t, "tf-provider-wf-tmpl-schedules") + ":1"
	wfGrpName := "tf-provider-wf-template-schedules-wfgrp"
	id := "tf-provider-wf-template-schedules"

	if err := createWorkflowGroupFixture(wfGrpName); err != nil {
		t.Errorf("failed to create workflow group fixture: %s", err.Error())
	}
	defer deleteWorkflowGroupFixture(wfGrpName)
	defer deleteWorkflowUsingTemplateFixture(wfGrpName, id)

	config := func(cron string) string {
		return fmt.Sprintf(`
  terraform_config = {
    terraform_version = %q
  }

  user_schedules = [
    {
      cron  = %q
      state = "ENABLED"
      desc  = "Runs on schedule"
    }
  ]
`, "1.5.0", cron)
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader()),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowUsingTemplate(wfGrpName, id, "TERRAFORM", templateID, config("0 8 ? * MON *")),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test", "id", id),
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test", "user_schedules.0.cron", "0 8 ? * MON *"),
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test", "user_schedules.0.state", "ENABLED"),
				),
			},
			{
				Config: testAccWorkflowUsingTemplate(wfGrpName, id, "TERRAFORM", templateID, config("0 9 ? * MON *")),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test", "id", id),
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test", "user_schedules.0.cron", "0 9 ? * MON *"),
				),
			},
		},
	})
}

func TestAccWorkflowUsingTemplate_WithTagsAndContextTags(t *testing.T) {
	templateID := setupWorkflowTemplate(t, "tf-provider-wf-tmpl-tags") + ":1"
	wfGrpName := "tf-provider-wf-template-tags-wfgrp"
	id := "tf-provider-wf-template-tags"

	if err := createWorkflowGroupFixture(wfGrpName); err != nil {
		t.Errorf("failed to create workflow group fixture: %s", err.Error())
	}
	defer deleteWorkflowGroupFixture(wfGrpName)
	defer deleteWorkflowUsingTemplateFixture(wfGrpName, id)

	config := func(tag, ctxVal string) string {
		return fmt.Sprintf(`
  terraform_config = {
    terraform_version = %q
  }

  tags = [%q]

  context_tags = {
    env = %q
  }
`, "1.5.0", tag, ctxVal)
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader()),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowUsingTemplate(wfGrpName, id, "TERRAFORM", templateID, config("v1", "staging")),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test", "id", id),
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test", "tags.0", "v1"),
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test", "context_tags.env", "staging"),
				),
			},
			{
				Config: testAccWorkflowUsingTemplate(wfGrpName, id, "TERRAFORM", templateID, config("v2", "production")),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test", "id", id),
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test", "tags.0", "v2"),
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test", "context_tags.env", "production"),
				),
			},
		},
	})
}

func TestAccWorkflowUsingTemplate_WithApprovers(t *testing.T) {
	templateID := setupWorkflowTemplate(t, "tf-provider-wf-tmpl-approvers") + ":1"
	wfGrpName := "tf-provider-wf-template-approvers-wfgrp"
	id := "tf-provider-wf-template-approvers"

	if err := createWorkflowGroupFixture(wfGrpName); err != nil {
		t.Errorf("failed to create workflow group fixture: %s", err.Error())
	}
	defer deleteWorkflowGroupFixture(wfGrpName)
	defer deleteWorkflowUsingTemplateFixture(wfGrpName, id)

	config := func(numApprovals int) string {
		return fmt.Sprintf(`
  terraform_config = {
    terraform_version = %q
  }

  approvers                    = ["approver@example.com"]
  number_of_approvals_required = %d
`, "1.5.0", numApprovals)
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader()),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowUsingTemplate(wfGrpName, id, "TERRAFORM", templateID, config(1)),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test", "id", id),
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test", "approvers.0", "approver@example.com"),
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test", "number_of_approvals_required", "1"),
				),
			},
			{
				Config: testAccWorkflowUsingTemplate(wfGrpName, id, "TERRAFORM", templateID, config(2)),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test", "id", id),
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test", "number_of_approvals_required", "2"),
				),
			},
		},
	})
}

func TestAccWorkflowUsingTemplate_WithRunnerConstraints(t *testing.T) {
	templateID := setupWorkflowTemplate(t, "tf-provider-wf-tmpl-runner") + ":1"
	wfGrpName := "tf-provider-wf-template-runner-wfgrp"
	id := "tf-provider-wf-template-runner"

	if err := createWorkflowGroupFixture(wfGrpName); err != nil {
		t.Errorf("failed to create workflow group fixture: %s", err.Error())
	}
	defer deleteWorkflowGroupFixture(wfGrpName)
	defer deleteWorkflowUsingTemplateFixture(wfGrpName, id)

	configShared := fmt.Sprintf(`
  terraform_config = {
    terraform_version = %q
  }

  runner_constraints = {
    type = "shared"
  }
`, "1.5.0")

	configPrivate := fmt.Sprintf(`
  terraform_config = {
    terraform_version = %q
  }

  runner_constraints = {
    type  = "private"
    names = ["runner-1"]
  }
`, "1.5.0")

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader()),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowUsingTemplate(wfGrpName, id, "TERRAFORM", templateID, configShared),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test", "id", id),
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test", "runner_constraints.type", "shared"),
				),
			},
			{
				Config: testAccWorkflowUsingTemplate(wfGrpName, id, "TERRAFORM", templateID, configPrivate),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test", "id", id),
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test", "runner_constraints.type", "private"),
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test", "runner_constraints.names.0", "runner-1"),
				),
			},
		},
	})
}

func TestAccWorkflowUsingTemplate_WithIacInputData(t *testing.T) {
	templateID := setupWorkflowTemplate(t, "tf-provider-wf-tmpl-iac-input") + ":1"
	wfGrpName := "tf-provider-wf-template-iac-input-wfgrp"
	id := "tf-provider-wf-template-iac-input"

	if err := createWorkflowGroupFixture(wfGrpName); err != nil {
		t.Errorf("failed to create workflow group fixture: %s", err.Error())
	}
	defer deleteWorkflowGroupFixture(wfGrpName)
	defer deleteWorkflowUsingTemplateFixture(wfGrpName, id)

	configWithIacInput := func(dataExpr string) string {
		return fmt.Sprintf(`
resource "stackguardian_workflow_from_template" "test" {
  workflow_group_id = %q
  id                = %q
  wf_type           = "TERRAFORM"

  vcs_config = {
    iac_vcs_config = {
      iac_template_id = %q
    }
    iac_input_data = {
      schema_type = "RAW_JSON"
      data        = %s
    }
  }

  terraform_config = {
    terraform_version = %q
  }
}
`, wfGrpName, id, templateID, dataExpr, "1.5.0")
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader()),
		Steps: []resource.TestStep{
			{
				Config: configWithIacInput(`jsonencode({"env" = "staging"})`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test", "id", id),
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test", "vcs_config.iac_input_data.schema_type", "RAW_JSON"),
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test", "vcs_config.iac_input_data.data", `{"env":"staging"}`),
				),
			},
			{
				Config: configWithIacInput(`jsonencode({"env" = "production"})`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test", "id", id),
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test", "vcs_config.iac_input_data.data", `{"env":"production"}`),
				),
			},
		},
	})
}

func TestAccWorkflowUsingTemplate_InNestedWorkflowGroup(t *testing.T) {
	templateID := setupWorkflowTemplate(t, "tf-provider-wf-tmpl-nested") + ":1"
	parentWfGrpName := "tf-provider-wf-template-nested-parent"
	childWfGrpName := parentWfGrpName + "/tf-provider-wf-template-nested-child"
	id := "tf-provider-wf-template-nested"

	if err := createWorkflowGroupFixture(parentWfGrpName); err != nil {
		t.Errorf("failed to create parent workflow group fixture: %s", err.Error())
	}
	defer deleteWorkflowGroupFixture(parentWfGrpName)

	if err := createWorkflowGroupFixture(childWfGrpName); err != nil {
		t.Errorf("failed to create child workflow group fixture: %s", err.Error())
	}
	defer deleteWorkflowGroupFixture(childWfGrpName)
	defer deleteWorkflowUsingTemplateFixture(childWfGrpName, id)

	config := fmt.Sprintf(`
  terraform_config = {
    terraform_version = %q
  }
`, "1.5.0")

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader()),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowUsingTemplate(childWfGrpName, id, "TERRAFORM", templateID, config),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test", "workflow_group_id", childWfGrpName),
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test", "id", id),
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test", "vcs_config.iac_vcs_config.iac_template_id", templateID),
				),
			},
		},
	})
}

// TestAccWorkflowUsingTemplate_FullResolution verifies that, on create, the user's
// declared fields and the template-derived defaults are both resolved onto the
// top-level attributes (state mirrors the fully-merged API record).
func TestAccWorkflowUsingTemplate_FullResolution(t *testing.T) {
	templateID := setupWorkflowTemplate(t, "tf-provider-wf-tmpl-resolved") + ":1"
	wfGrpName := "tf-provider-wf-template-resolved-wfgrp"
	id := "tf-provider-wf-template-resolved"

	if err := createWorkflowGroupFixture(wfGrpName); err != nil {
		t.Errorf("failed to create workflow group fixture: %s", err.Error())
	}
	defer deleteWorkflowGroupFixture(wfGrpName)
	defer deleteWorkflowUsingTemplateFixture(wfGrpName, id)

	config := `
  description = "test resolved schema"

  terraform_config = {
    terraform_version = "1.5.0"
  }

  environment_variables = [
    {
      kind = "PLAIN_TEXT"
      config = {
        var_name   = "RESOLVED_VAR"
        text_value = "resolved-value"
      }
    }
  ]
`

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader()),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowUsingTemplate(wfGrpName, id, "TERRAFORM", templateID, config),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test", "id", id),
					// User-declared fields resolved onto top-level attributes.
					resource.TestCheckResourceAttrSet("stackguardian_workflow_from_template.test", "resource_name"),
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test", "description", "test resolved schema"),
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test", "terraform_config.terraform_version", "1.5.0"),
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test", "environment_variables.0.kind", "PLAIN_TEXT"),
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test", "environment_variables.0.config.var_name", "RESOLVED_VAR"),
					// user_job_cpu/memory were never declared — they must be resolved from
					// the template/platform defaults (Computed) rather than left null.
					resource.TestCheckResourceAttrSet("stackguardian_workflow_from_template.test", "user_job_cpu"),
					resource.TestCheckResourceAttrSet("stackguardian_workflow_from_template.test", "user_job_memory"),
				),
			},
		},
	})
}

// TestAccWorkflowUsingTemplate_TemplateDefaultsResolved verifies that a field the
// user never declares is populated from the template revision (provider-side merge),
// and that a clean plan immediately after apply shows no diff.
func TestAccWorkflowUsingTemplate_TemplateDefaultsResolved(t *testing.T) {
	templateID := setupWorkflowTemplate(t, "tf-provider-wf-tmpl-defaults") + ":1"
	wfGrpName := "tf-provider-wf-template-defaults-wfgrp"
	id := "tf-provider-wf-template-defaults"

	if err := createWorkflowGroupFixture(wfGrpName); err != nil {
		t.Errorf("failed to create workflow group fixture: %s", err.Error())
	}
	defer deleteWorkflowGroupFixture(wfGrpName)
	defer deleteWorkflowUsingTemplateFixture(wfGrpName, id)

	// The user declares nothing beyond the template reference. Env vars and the user
	// schedule come from the template revision created in setupWorkflowTemplate.
	config := ``

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader()),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowUsingTemplate(wfGrpName, id, "TERRAFORM", templateID, config),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test", "id", id),
					// Template-derived env var resolved even though the user declared none.
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test", "environment_variables.0.config.var_name", "TMPL_VAR"),
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test", "environment_variables.0.config.text_value", "tmpl-value"),
					// Template-derived schedule resolved.
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test", "user_schedules.0.cron", "0 8 ? * MON *"),
				),
			},
			{
				// Re-applying the identical config must produce no plan (state == reality).
				Config:   testAccWorkflowUsingTemplate(wfGrpName, id, "TERRAFORM", templateID, config),
				PlanOnly: true,
			},
		},
	})
}

// TestAccWorkflowUsingTemplate_DriftDetection verifies the core capability this design
// enables: when the workflow is changed out-of-band (simulating a dev-portal edit), the
// next plan detects the drift instead of reporting "no changes". The out-of-band change
// is made directly via the API between steps; the following refresh-and-plan must be
// non-empty.
func TestAccWorkflowUsingTemplate_DriftDetection(t *testing.T) {
	templateID := setupWorkflowTemplate(t, "tf-provider-wf-tmpl-drift") + ":1"
	wfGrpName := "tf-provider-wf-template-drift-wfgrp"
	id := "tf-provider-wf-template-drift"

	if err := createWorkflowGroupFixture(wfGrpName); err != nil {
		t.Errorf("failed to create workflow group fixture: %s", err.Error())
	}
	defer deleteWorkflowGroupFixture(wfGrpName)
	defer deleteWorkflowUsingTemplateFixture(wfGrpName, id)

	config := `
  description = "original description"

  terraform_config = {
    terraform_version = "1.5.0"
  }
`

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader()),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowUsingTemplate(wfGrpName, id, "TERRAFORM", templateID, config),
				Check: resource.TestCheckResourceAttr(
					"stackguardian_workflow_from_template.test", "description", "original description"),
			},
			{
				// Mutate the workflow out-of-band, then run a refresh-only step. Because
				// state now mirrors reality field-for-field, the refresh detects the
				// drifted description and the resulting plan is non-empty. (RefreshState
				// cannot be combined with Config, so this step intentionally omits Config.)
				PreConfig: func() {
					client := getClient()
					newDesc := "changed-out-of-band"
					_, err := client.Workflows.UpdateWorkflow(
						context.TODO(), org, id, wfGrpName,
						sgworkflows.UpgradeModeEnumPreserveSettings.Ptr(),
						&sgworkflows.PatchedWorkflow{
							Description: sgsdkgo.Optional(newDesc),
						},
					)
					if err != nil {
						t.Fatalf("drift setup: out-of-band update failed: %s", err)
					}
				},
				RefreshState:       true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}
