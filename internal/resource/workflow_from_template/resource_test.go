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
	"github.com/StackGuardian/sg-sdk-go/workflowsteptemplate"
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

// addSecondRevision creates and publishes revision :2 of an existing template, with a
// different env var (REV2_VAR) and terraform version so an upgrade is observable. It
// registers cleanup for the revision. templateID is the bare template name (no org).
func addSecondRevision(t *testing.T, templateID string) string {
	t.Helper()
	client := getClient()
	revisionID := fmt.Sprintf("%s:2", templateID)
	sourceConfigKind := workflowtemplates.WorkflowTemplateSourceConfigKindTerraform

	is409 := func(err error) bool {
		var apiErr *core.APIError
		return errors.As(err, &apiErr) && apiErr.StatusCode == 409
	}

	alias := "v2"
	envTextValue := "rev2-value"
	tmplTfVersion := "1.5.7"
	_, err := client.WorkflowTemplatesRevisions.CreateWorkflowTemplateRevision(
		context.TODO(), org, templateID,
		&workflowtemplaterevisions.CreateWorkflowTemplateRevisionsRequest{
			Alias:            alias,
			SourceConfigKind: &sourceConfigKind,
			IsPublic:         sgsdkgo.IsPublicEnumZero.Ptr(),
			OwnerOrg:         fmt.Sprintf("/orgs/%s", org),
			// Drift ENABLED with a valid 6-field cron (the API validates the cron when
			// drift_check is true), so it is meaningful and kept by the coupling. rev1 has
			// no terraform_config drift fields; the upgrade test asserts these re-resolve
			// onto the workflow when moving rev1 -> rev2.
			TerraformConfig: &sgsdkgo.TerraformConfig{
				TerraformVersion: &tmplTfVersion,
				DriftCheck:       sgsdkgo.Bool(true),
				DriftCron:        sgsdkgo.String("0 */6 * * ? *"),
			},
			EnvironmentVariables: []sgsdkgo.EnvVars{
				{
					Kind: sgsdkgo.EnvVarsKindEnumPlainText,
					Config: &sgsdkgo.EnvVarConfig{
						VarName:   "REV2_VAR",
						TextValue: &envTextValue,
					},
				},
			},
		},
	)
	if err != nil && !is409(err) {
		t.Fatalf("addSecondRevision: create revision for %q: %s", templateID, err)
	}

	_, err = client.WorkflowTemplatesRevisions.UpdateWorkflowTemplateRevision(
		context.TODO(), org, revisionID,
		&workflowtemplaterevisions.UpdateWorkflowTemplateRevisionRequest{
			IsPublic: sgsdkgo.Optional(sgsdkgo.IsPublicEnumOne),
		},
	)
	if err != nil {
		client.WorkflowTemplatesRevisions.DeleteWorkflowTemplateRevision(context.TODO(), org, revisionID, true)
		t.Fatalf("addSecondRevision: publish revision %q: %s", revisionID, err)
	}

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
	})

	return fmt.Sprintf("/%s/%s:2", org, templateID)
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

// setupWorkflowStepTemplate creates and publishes a WORKFLOW_STEP template (with a
// DOCKER_IMAGE runtime source) and returns its fully-qualified revision id
// ("/<org>/<name>:1") for use as wf_step_template_id. Registers cleanup.
func setupWorkflowStepTemplate(t *testing.T, name string) string {
	t.Helper()
	client := getClient()

	is409 := func(err error) bool {
		var apiErr *core.APIError
		return errors.As(err, &apiErr) && apiErr.StatusCode == 409
	}

	sourceKind := workflowsteptemplate.WorkflowStepTemplateSourceConfigKindDockerImageEnum
	isPublic := workflowsteptemplate.IsPublicEnumOne
	isPrivate := false
	// createFirstRevision=true so :1 exists and is referenceable immediately.
	_, err := client.WorkflowStepTemplate.CreateWorkflowStepTemplate(
		context.TODO(), org, true,
		&workflowsteptemplate.CreateWorkflowStepTemplate{
			TemplateName:     name,
			TemplateType:     workflowsteptemplate.TemplateTypeWorkflowStepEnum,
			SourceConfigKind: sourceKind,
			IsPublic:         &isPublic,
			OwnerOrg:         fmt.Sprintf("/orgs/%s", org),
			RuntimeSource: &workflowsteptemplate.WorkflowStepRuntimeSource{
				SourceConfigDestKind: workflowsteptemplate.SourceConfigDestKindContainerRegistryEnum,
				Config: &workflowsteptemplate.WorkflowStepRuntimeSourceConfig{
					DockerImage: "ubuntu:latest",
					IsPrivate:   &isPrivate,
				},
			},
		},
	)
	if err != nil && !is409(err) {
		t.Fatalf("setupWorkflowStepTemplate: create %q: %s", name, err)
	}

	t.Cleanup(func() {
		client.WorkflowStepTemplate.DeleteWorkflowStepTemplate(context.TODO(), org, name)
	})

	return fmt.Sprintf("/%s/%s:1", org, name)
}

// setupCustomWorkflowTemplate creates and publishes a CUSTOM-source workflow template
// (revision :1). Top-level wf_steps_config is only allowed for CUSTOM workflow types
// (the API rejects it for TERRAFORM/OPENTOFU — see constants.WfStepsConfig), so this
// fixture backs the top-level wf_steps_config test. Registers cleanup.
func setupCustomWorkflowTemplate(t *testing.T, templateID string) string {
	t.Helper()
	client := getClient()
	revisionID := fmt.Sprintf("%s:1", templateID)
	sourceConfigKind := workflowtemplates.WorkflowTemplateSourceConfigKindCustom

	is409 := func(err error) bool {
		var apiErr *core.APIError
		return errors.As(err, &apiErr) && apiErr.StatusCode == 409
	}

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
		t.Fatalf("setupCustomWorkflowTemplate: create template %q: %s", templateID, err)
	}

	alias := "v1"
	_, err = client.WorkflowTemplatesRevisions.CreateWorkflowTemplateRevision(
		context.TODO(), org, templateID,
		&workflowtemplaterevisions.CreateWorkflowTemplateRevisionsRequest{
			Alias:            alias,
			SourceConfigKind: &sourceConfigKind,
			IsPublic:         sgsdkgo.IsPublicEnumZero.Ptr(),
			OwnerOrg:         fmt.Sprintf("/orgs/%s", org),
		},
	)
	if err != nil && !is409(err) {
		client.WorkflowTemplates.DeleteWorkflowTemplate(context.TODO(), org, templateID)
		t.Fatalf("setupCustomWorkflowTemplate: create revision for %q: %s", templateID, err)
	}

	_, err = client.WorkflowTemplatesRevisions.UpdateWorkflowTemplateRevision(
		context.TODO(), org, revisionID,
		&workflowtemplaterevisions.UpdateWorkflowTemplateRevisionRequest{
			IsPublic: sgsdkgo.Optional(sgsdkgo.IsPublicEnumOne),
		},
	)
	if err != nil {
		client.WorkflowTemplatesRevisions.DeleteWorkflowTemplateRevision(context.TODO(), org, revisionID, true)
		client.WorkflowTemplates.DeleteWorkflowTemplate(context.TODO(), org, templateID)
		t.Fatalf("setupCustomWorkflowTemplate: publish revision %q: %s", revisionID, err)
	}

	_, err = client.WorkflowTemplates.UpdateWorkflowTemplate(
		context.TODO(), org, templateID,
		&workflowtemplates.UpdateWorkflowTemplateRequest{
			IsPublic: sgsdkgo.Optional(sgsdkgo.IsPublicEnumOne),
		},
	)
	if err != nil {
		client.WorkflowTemplatesRevisions.DeleteWorkflowTemplateRevision(context.TODO(), org, revisionID, true)
		client.WorkflowTemplates.DeleteWorkflowTemplate(context.TODO(), org, templateID)
		t.Fatalf("setupCustomWorkflowTemplate: publish template %q: %s", templateID, err)
	}

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

// TestAccWorkflowUsingTemplate_WithWfStepsConfig verifies a user-declared top-level
// wf_steps_config (with the Required wf_step_template_id) on a CUSTOM workflow is created
// and round-trips stably — the second PlanOnly step asserts no perpetual diff (gotcha #7)
// and that wf_step_template_id is preserved. Top-level wf_steps_config is CUSTOM-only.
func TestAccWorkflowUsingTemplate_WithWfStepsConfig(t *testing.T) {
	templateID := setupCustomWorkflowTemplate(t, "tf-provider-wf-tmpl-wfsteps") + ":1"
	stepTemplateID := setupWorkflowStepTemplate(t, "tf-provider-wf-step-tmpl")
	wfGrpName := acctest.ResourceName("tf-provider-wf-template-wfsteps-wfgrp")
	id := acctest.ResourceName("tf-provider-wf-template-wfsteps")

	if err := createWorkflowGroupFixture(wfGrpName); err != nil {
		t.Errorf("failed to create workflow group fixture: %s", err.Error())
	}
	defer deleteWorkflowGroupFixture(wfGrpName)
	defer deleteWorkflowUsingTemplateFixture(wfGrpName, id)

	config := fmt.Sprintf(`
  wf_steps_config = [
    {
      name                = "step-one"
      wf_step_template_id = %q
      approval            = true
    }
  ]
`, stepTemplateID)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader()),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowUsingTemplate(wfGrpName, id, "CUSTOM", templateID, config),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test", "wf_steps_config.0.name", "step-one"),
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test", "wf_steps_config.0.wf_step_template_id", stepTemplateID),
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test", "wf_steps_config.0.approval", "true"),
				),
			},
			{
				// Re-plan the same config: must be a no-op (no perpetual diff on the
				// nested Optional fields / wf_step_template_id).
				Config:   testAccWorkflowUsingTemplate(wfGrpName, id, "CUSTOM", templateID, config),
				PlanOnly: true,
			},
		},
	})
}

// TestAccWorkflowUsingTemplate_LifecycleWfStepsConfig verifies a user-declared lifecycle
// hook list inside terraform_config (post_apply_wf_steps_config), which reuses the same
// wfStepsConfig() nested object in a Computed context. Asserts create + round-trip
// stability (PlanOnly no-op) for the Required wf_step_template_id in that context.
func TestAccWorkflowUsingTemplate_LifecycleWfStepsConfig(t *testing.T) {
	templateID := setupWorkflowTemplate(t, "tf-provider-wf-tmpl-lifecyclesteps") + ":1"
	stepTemplateID := setupWorkflowStepTemplate(t, "tf-provider-wf-lifecycle-step-tmpl")
	wfGrpName := acctest.ResourceName("tf-provider-wf-template-lifecyclesteps-wfgrp")
	id := acctest.ResourceName("tf-provider-wf-template-lifecyclesteps")

	if err := createWorkflowGroupFixture(wfGrpName); err != nil {
		t.Errorf("failed to create workflow group fixture: %s", err.Error())
	}
	defer deleteWorkflowGroupFixture(wfGrpName)
	defer deleteWorkflowUsingTemplateFixture(wfGrpName, id)

	config := fmt.Sprintf(`
  terraform_config = {
    terraform_version = "1.5.0"
    post_apply_wf_steps_config = [
      {
        name                = "post-apply-step"
        wf_step_template_id = %q
      }
    ]
  }
`, stepTemplateID)

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
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test", "terraform_config.post_apply_wf_steps_config.0.name", "post-apply-step"),
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test", "terraform_config.post_apply_wf_steps_config.0.wf_step_template_id", stepTemplateID),
				),
			},
			{
				Config:   testAccWorkflowUsingTemplate(wfGrpName, id, "TERRAFORM", templateID, config),
				PlanOnly: true,
			},
		},
	})
}

func TestAccWorkflowUsingTemplate_Basic(t *testing.T) {
	templateID := setupWorkflowTemplate(t, "tf-provider-wf-tmpl-basic") + ":1"
	wfGrpName := acctest.ResourceName("tf-provider-wf-template-basic-wfgrp")
	id := acctest.ResourceName("tf-provider-wf-template-basic")

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
	wfGrpName := acctest.ResourceName("tf-provider-wf-template-desc-wfgrp")
	id := acctest.ResourceName("tf-provider-wf-template-desc")

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

// setupTemplateWithWfStepRevision creates a template whose revision :1 carries
// TerraformConfig.wfStepTemplateRevisionId (pointing at stepRev), so a workflow built from it
// inherits the value when the user does not declare it. Registers cleanup.
func setupTemplateWithWfStepRevision(t *testing.T, name, stepRev string) string {
	t.Helper()
	client := getClient()
	revisionID := fmt.Sprintf("%s:1", name)
	sck := workflowtemplates.WorkflowTemplateSourceConfigKindTerraform
	is409 := func(err error) bool {
		var apiErr *core.APIError
		return errors.As(err, &apiErr) && apiErr.StatusCode == 409
	}
	_, err := client.WorkflowTemplates.CreateWorkflowTemplate(context.TODO(), org, false,
		&workflowtemplates.CreateWorkflowTemplateRequest{
			Id: &name, TemplateName: name, SourceConfigKind: &sck,
			TemplateType: sgsdkgo.TemplateTypeEnum("IAC"), IsPublic: sgsdkgo.IsPublicEnumZero.Ptr(),
			OwnerOrg: fmt.Sprintf("/orgs/%s", org),
		})
	if err != nil && !is409(err) {
		t.Fatalf("create tpl: %s", err)
	}
	tfVer := "1.5.0"
	_, err = client.WorkflowTemplatesRevisions.CreateWorkflowTemplateRevision(context.TODO(), org, name,
		&workflowtemplaterevisions.CreateWorkflowTemplateRevisionsRequest{
			Alias: "v1", SourceConfigKind: &sck, IsPublic: sgsdkgo.IsPublicEnumZero.Ptr(),
			OwnerOrg: fmt.Sprintf("/orgs/%s", org),
			TerraformConfig: &sgsdkgo.TerraformConfig{
				TerraformVersion:         &tfVer,
				WfStepTemplateRevisionId: &stepRev,
			},
		})
	if err != nil && !is409(err) {
		client.WorkflowTemplates.DeleteWorkflowTemplate(context.TODO(), org, name)
		t.Fatalf("create rev: %s", err)
	}
	if _, err := client.WorkflowTemplatesRevisions.UpdateWorkflowTemplateRevision(context.TODO(), org, revisionID,
		&workflowtemplaterevisions.UpdateWorkflowTemplateRevisionRequest{IsPublic: sgsdkgo.Optional(sgsdkgo.IsPublicEnumOne)}); err != nil {
		client.WorkflowTemplatesRevisions.DeleteWorkflowTemplateRevision(context.TODO(), org, revisionID, true)
		client.WorkflowTemplates.DeleteWorkflowTemplate(context.TODO(), org, name)
		t.Fatalf("publish rev: %s", err)
	}
	if _, err := client.WorkflowTemplates.UpdateWorkflowTemplate(context.TODO(), org, name,
		&workflowtemplates.UpdateWorkflowTemplateRequest{IsPublic: sgsdkgo.Optional(sgsdkgo.IsPublicEnumOne)}); err != nil {
		client.WorkflowTemplatesRevisions.DeleteWorkflowTemplateRevision(context.TODO(), org, revisionID, true)
		client.WorkflowTemplates.DeleteWorkflowTemplate(context.TODO(), org, name)
		t.Fatalf("publish tpl: %s", err)
	}
	t.Cleanup(func() {
		eff := fmt.Sprintf("%d", time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC).Unix())
		msg := "cleanup"
		client.WorkflowTemplatesRevisions.UpdateWorkflowTemplateRevision(context.TODO(), org, revisionID,
			&workflowtemplaterevisions.UpdateWorkflowTemplateRevisionRequest{
				Deprecation: sgsdkgo.Optional(workflowtemplaterevisions.Deprecation{EffectiveDate: &eff, Message: &msg})})
		client.WorkflowTemplatesRevisions.DeleteWorkflowTemplateRevision(context.TODO(), org, revisionID, true)
		client.WorkflowTemplates.DeleteWorkflowTemplate(context.TODO(), org, name)
	})
	return fmt.Sprintf("/%s/%s:1", org, name)
}

// TestAccWorkflowUsingTemplate_WfStepRevisionInheritedFromTemplate verifies the
// Optional+Computed behavior: when the TEMPLATE carries wfStepTemplateRevisionId and the user
// does NOT declare it, the workflow INHERITS it (proving the Computed + merge path), and the
// value round-trips with no diff.
func TestAccWorkflowUsingTemplate_WfStepRevisionInheritedFromTemplate(t *testing.T) {
	stepRev := setupWorkflowStepTemplate(t, "tf-provider-wf-steprev-inh") // "/<org>/<name>:1"
	templateID := setupTemplateWithWfStepRevision(t, "tf-provider-wf-tmpl-steprev-inh", stepRev)
	wfGrpName := acctest.ResourceName("tf-provider-wf-template-steprev-inh-wfgrp")
	id := acctest.ResourceName("tf-provider-wf-template-steprev-inh")

	if err := createWorkflowGroupFixture(wfGrpName); err != nil {
		t.Errorf("failed to create workflow group fixture: %s", err.Error())
	}
	defer deleteWorkflowGroupFixture(wfGrpName)
	defer deleteWorkflowUsingTemplateFixture(wfGrpName, id)

	// User declares only terraform_version; wf_step_template_revision_id is inherited.
	config := `
  terraform_config = {
    terraform_version = "1.5.0"
  }
`
	// Suppress the inherited value with an explicit "" -> API maps "" to None -> default.
	suppressConfig := `
  terraform_config = {
    terraform_version            = "1.5.0"
    wf_step_template_revision_id = ""
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
				// Case 2: present in template, user omits -> inherited.
				Config: testAccWorkflowUsingTemplate(wfGrpName, id, "TERRAFORM", templateID, config),
				Check: resource.TestCheckResourceAttr(
					"stackguardian_workflow_from_template.test", "terraform_config.wf_step_template_revision_id", stepRev),
			},
			{
				// Inherited value round-trips with no diff.
				Config:   testAccWorkflowUsingTemplate(wfGrpName, id, "TERRAFORM", templateID, config),
				PlanOnly: true,
			},
			{
				// Case 4: present in template but user sets "" to suppress -> API maps to
				// None -> default; state holds "" (not the inherited template value).
				Config: testAccWorkflowUsingTemplate(wfGrpName, id, "TERRAFORM", templateID, suppressConfig),
				Check: resource.TestCheckResourceAttr(
					"stackguardian_workflow_from_template.test", "terraform_config.wf_step_template_revision_id", ""),
			},
			{
				// Suppressed ("") round-trips with no diff.
				Config:   testAccWorkflowUsingTemplate(wfGrpName, id, "TERRAFORM", templateID, suppressConfig),
				PlanOnly: true,
			},
		},
	})
}

// TestAccWorkflowUsingTemplate_WithWfStepTemplateRevisionId verifies the user-settable
// terraform_config.wf_step_template_revision_id round-trips: set it, confirm it's stored and
// read back unchanged (PlanOnly no-op), then update it to a different revision.
func TestAccWorkflowUsingTemplate_WithWfStepTemplateRevisionId(t *testing.T) {
	templateID := setupWorkflowTemplate(t, "tf-provider-wf-tmpl-wfsteprev") + ":1"
	step1 := setupWorkflowStepTemplate(t, "tf-provider-wf-steprev-1") // "/<org>/<name>:1"
	step2 := setupWorkflowStepTemplate(t, "tf-provider-wf-steprev-2")
	wfGrpName := acctest.ResourceName("tf-provider-wf-template-wfsteprev-wfgrp")
	id := acctest.ResourceName("tf-provider-wf-template-wfsteprev")

	if err := createWorkflowGroupFixture(wfGrpName); err != nil {
		t.Errorf("failed to create workflow group fixture: %s", err.Error())
	}
	defer deleteWorkflowGroupFixture(wfGrpName)
	defer deleteWorkflowUsingTemplateFixture(wfGrpName, id)

	config := func(stepRev string) string {
		return fmt.Sprintf(`
  terraform_config = {
    terraform_version            = "1.5.0"
    wf_step_template_revision_id = %q
  }
`, stepRev)
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader()),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowUsingTemplate(wfGrpName, id, "TERRAFORM", templateID, config(step1)),
				Check: resource.TestCheckResourceAttr(
					"stackguardian_workflow_from_template.test", "terraform_config.wf_step_template_revision_id", step1),
			},
			{
				// Round-trips with no diff.
				Config:   testAccWorkflowUsingTemplate(wfGrpName, id, "TERRAFORM", templateID, config(step1)),
				PlanOnly: true,
			},
			{
				// Update to a different step template revision.
				Config: testAccWorkflowUsingTemplate(wfGrpName, id, "TERRAFORM", templateID, config(step2)),
				Check: resource.TestCheckResourceAttr(
					"stackguardian_workflow_from_template.test", "terraform_config.wf_step_template_revision_id", step2),
			},
		},
	})
}

func TestAccWorkflowUsingTemplate_WithTerraformConfig(t *testing.T) {
	templateID := setupWorkflowTemplate(t, "tf-provider-wf-tmpl-tfcfg") + ":1"
	wfGrpName := acctest.ResourceName("tf-provider-wf-template-tfcfg-wfgrp")
	id := acctest.ResourceName("tf-provider-wf-template-tfcfg")

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
				Config: testAccWorkflowUsingTemplate(wfGrpName, id, "TERRAFORM", templateID, config("1.5.7")),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test", "id", id),
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test", "terraform_config.terraform_version", "1.5.7"),
				),
			},
		},
	})
}

// TestAccWorkflowUsingTemplate_NormalUpdate verifies ordinary in-place updates across a
// mix of attribute kinds in one go: a top-level scalar (description), a nested
// terraform_config scalar (terraform_version), a list (tags), and a map (context_tags).
// Step 1 creates; step 2 mutates every one of them; step 3 re-applies step 2's config to
// prove the update settled to a stable, no-diff state.
func TestAccWorkflowUsingTemplate_NormalUpdate(t *testing.T) {
	templateID := setupWorkflowTemplate(t, "tf-provider-wf-tmpl-normalupd") + ":1"
	wfGrpName := acctest.ResourceName("tf-provider-wf-template-normalupd-wfgrp")
	id := acctest.ResourceName("tf-provider-wf-template-normalupd")

	if err := createWorkflowGroupFixture(wfGrpName); err != nil {
		t.Errorf("failed to create workflow group fixture: %s", err.Error())
	}
	defer deleteWorkflowGroupFixture(wfGrpName)
	defer deleteWorkflowUsingTemplateFixture(wfGrpName, id)

	config := func(desc, tfVersion, tagVal, ctxVal string) string {
		return fmt.Sprintf(`
  description = %q

  terraform_config = {
    terraform_version = %q
  }

  tags = [%q]

  context_tags = {
    env = %q
  }
`, desc, tfVersion, tagVal, ctxVal)
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader()),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowUsingTemplate(wfGrpName, id, "TERRAFORM", templateID, config("first", "1.5.0", "tag-a", "dev")),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test", "description", "first"),
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test", "terraform_config.terraform_version", "1.5.0"),
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test", "tags.0", "tag-a"),
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test", "context_tags.env", "dev"),
				),
			},
			{
				// Update every attribute at once.
				Config: testAccWorkflowUsingTemplate(wfGrpName, id, "TERRAFORM", templateID, config("second", "1.5.7", "tag-b", "prod")),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test", "description", "second"),
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test", "terraform_config.terraform_version", "1.5.7"),
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test", "tags.0", "tag-b"),
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test", "context_tags.env", "prod"),
				),
			},
			{
				// Re-apply identical config: must be a stable no-op.
				Config:   testAccWorkflowUsingTemplate(wfGrpName, id, "TERRAFORM", templateID, config("second", "1.5.7", "tag-b", "prod")),
				PlanOnly: true,
			},
		},
	})
}

// TestAccWorkflowUsingTemplate_EmptyAllowBlankFalse verifies that supplying an empty
// string ("") for attributes the API treats as allow_blank=False does NOT produce a
// 400 from a blank payload: the provider omits empty allow_blank=False strings
// (isNonEmpty guard in ToAPIModel). Covers terraform_config.{terraform_plan_options,
// terraform_init_options} and a transition where drift_cron is "" while drift_check is
// false. Also asserts the plan is stable afterward (empty round-trips consistently).
func TestAccWorkflowUsingTemplate_EmptyAllowBlankFalse(t *testing.T) {
	templateID := setupWorkflowTemplate(t, "tf-provider-wf-tmpl-emptyblank") + ":1"
	wfGrpName := acctest.ResourceName("tf-provider-wf-template-emptyblank-wfgrp")
	id := acctest.ResourceName("tf-provider-wf-template-emptyblank")

	if err := createWorkflowGroupFixture(wfGrpName); err != nil {
		t.Errorf("failed to create workflow group fixture: %s", err.Error())
	}
	defer deleteWorkflowGroupFixture(wfGrpName)
	defer deleteWorkflowUsingTemplateFixture(wfGrpName, id)

	// All allow_blank=False strings set to "". Without the omit-empty guard the API would
	// reject the blank values; with it, they are dropped and creation succeeds.
	emptyConfig := `
  description = ""

  terraform_config = {
    terraform_version      = "1.5.0"
    terraform_plan_options = ""
    terraform_init_options = ""
    drift_check            = false
    drift_cron             = ""
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
				Config: testAccWorkflowUsingTemplate(wfGrpName, id, "TERRAFORM", templateID, emptyConfig),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test", "terraform_config.terraform_version", "1.5.0"),
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test", "terraform_config.terraform_plan_options", ""),
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test", "terraform_config.terraform_init_options", ""),
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test", "terraform_config.drift_check", "false"),
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test", "terraform_config.drift_cron", ""),
				),
			},
			{
				// Empty values must round-trip with no diff (no "inconsistent result",
				// no perpetual plan from "" vs omitted).
				Config:   testAccWorkflowUsingTemplate(wfGrpName, id, "TERRAFORM", templateID, emptyConfig),
				PlanOnly: true,
			},
			{
				// Drift-check (refresh from the API) on empty allow_blank=False values must
				// produce a CLEAN plan: the API stores nothing for the omitted blanks, the
				// read-back coerces null -> "" deterministically, so refreshed state matches
				// config and there is no spurious drift. RefreshState refreshes the previous
				// step's config; it cannot carry its own Config. ExpectNonEmptyPlan defaults
				// to false, so a non-empty plan after refresh fails the step.
				RefreshState: true,
			},
			{
				// Transition empty -> real value, proving an allow_blank=False field can be
				// set after being empty (a normal update on a previously-omitted field).
				Config: testAccWorkflowUsingTemplate(wfGrpName, id, "TERRAFORM", templateID, `
  terraform_config = {
    terraform_version      = "1.5.0"
    terraform_plan_options = "-input=false"
  }
`),
				Check: resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test", "terraform_config.terraform_plan_options", "-input=false"),
			},
			{
				// And the transitioned real value must itself round-trip cleanly.
				Config: testAccWorkflowUsingTemplate(wfGrpName, id, "TERRAFORM", templateID, `
  terraform_config = {
    terraform_version      = "1.5.0"
    terraform_plan_options = "-input=false"
  }
`),
				PlanOnly: true,
			},
		},
	})
}

// TestAccWorkflowUsingTemplate_ExplicitEmptySuppressesTemplateDefault verifies the other
// side of empty handling: an explicit empty list ([]) is NOT omitted — it is sent to
// suppress a template default (optionalIfPresent). The standard template supplies a
// TMPL_VAR env var; declaring environment_variables = [] must override it to empty rather
// than inherit it. Omitting the attribute entirely (control, step 2 of a separate run is
// covered elsewhere) would instead inherit TMPL_VAR.
func TestAccWorkflowUsingTemplate_ExplicitEmptySuppressesTemplateDefault(t *testing.T) {
	base := acctest.ResourceName("tf-provider-wf-tmpl-emptysuppress")
	rev1 := setupWorkflowTemplate(t, base) + ":1" // supplies TMPL_VAR
	rev2 := addSecondRevision(t, base)            // supplies REV2_VAR
	wfGrpName := acctest.ResourceName("tf-provider-wf-template-emptysuppress-wfgrp")
	id := acctest.ResourceName("tf-provider-wf-template-emptysuppress")

	if err := createWorkflowGroupFixture(wfGrpName); err != nil {
		t.Errorf("failed to create workflow group fixture: %s", err.Error())
	}
	defer deleteWorkflowGroupFixture(wfGrpName)
	defer deleteWorkflowUsingTemplateFixture(wfGrpName, id)

	// environment_variables = [] explicitly suppresses the template's env var default.
	suppressConfig := `
  terraform_config = {
    terraform_version = "1.5.0"
  }

  environment_variables = []
`

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader()),
		Steps: []resource.TestStep{
			{
				// Create on rev1: explicit [] wins over rev1's TMPL_VAR default → zero env vars.
				Config: testAccWorkflowUsingTemplate(wfGrpName, id, "TERRAFORM", rev1, suppressConfig),
				Check:  resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test", "environment_variables.#", "0"),
			},
			{
				// Explicit empty must round-trip stably (not re-inherit the template default).
				Config:   testAccWorkflowUsingTemplate(wfGrpName, id, "TERRAFORM", rev1, suppressConfig),
				PlanOnly: true,
			},
			{
				// Drift-check (refresh from the API) must keep the suppression: refreshed
				// state stays at zero env vars and the plan is clean — the template's
				// TMPL_VAR is NOT re-inherited on refresh. RefreshState refreshes the
				// previous step's config and cannot carry its own Config.
				RefreshState: true,
			},
			{
				// Upgrade rev1 -> rev2: the explicit [] suppression must STILL win — the
				// upgrade must not re-inherit rev2's REV2_VAR. Env vars stay at zero.
				Config: testAccWorkflowUsingTemplate(wfGrpName, id, "TERRAFORM", rev2, suppressConfig),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test", "vcs_config.iac_vcs_config.iac_template_id", rev2),
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test", "environment_variables.#", "0"),
				),
			},
			{
				// And the post-upgrade suppressed state must round-trip cleanly.
				Config:   testAccWorkflowUsingTemplate(wfGrpName, id, "TERRAFORM", rev2, suppressConfig),
				PlanOnly: true,
			},
		},
	})
}

func TestAccWorkflowUsingTemplate_WithEnvironmentVariables(t *testing.T) {
	templateID := setupWorkflowTemplate(t, "tf-provider-wf-tmpl-envvars") + ":1"
	wfGrpName := acctest.ResourceName("tf-provider-wf-template-envvars-wfgrp")
	id := acctest.ResourceName("tf-provider-wf-template-envvars")

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
	wfGrpName := acctest.ResourceName("tf-provider-wf-template-schedules-wfgrp")
	id := acctest.ResourceName("tf-provider-wf-template-schedules")

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
	wfGrpName := acctest.ResourceName("tf-provider-wf-template-tags-wfgrp")
	id := acctest.ResourceName("tf-provider-wf-template-tags")

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
	wfGrpName := acctest.ResourceName("tf-provider-wf-template-approvers-wfgrp")
	id := acctest.ResourceName("tf-provider-wf-template-approvers")

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
	wfGrpName := acctest.ResourceName("tf-provider-wf-template-runner-wfgrp")
	id := acctest.ResourceName("tf-provider-wf-template-runner")

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
	wfGrpName := acctest.ResourceName("tf-provider-wf-template-iac-input-wfgrp")
	id := acctest.ResourceName("tf-provider-wf-template-iac-input")

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
	parentWfGrpName := acctest.ResourceName("tf-provider-wf-template-nested-parent")
	childWfGrpName := parentWfGrpName + "/tf-provider-wf-template-nested-child"
	id := acctest.ResourceName("tf-provider-wf-template-nested")

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
	wfGrpName := acctest.ResourceName("tf-provider-wf-template-resolved-wfgrp")
	id := acctest.ResourceName("tf-provider-wf-template-resolved")

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
	wfGrpName := acctest.ResourceName("tf-provider-wf-template-defaults-wfgrp")
	id := acctest.ResourceName("tf-provider-wf-template-defaults")

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

// setupDriftEnabledTemplate creates and publishes a template whose revision :1 has drift
// checking ENABLED (driftCheck=true) with a driftCron. Used to verify the
// drift_check/drift_cron coupling: a workflow that sets drift_check=false must drop the
// inherited cron. Mirrors setupWorkflowTemplate's create/publish/cleanup flow.
func setupDriftEnabledTemplate(t *testing.T, templateID string) string {
	t.Helper()
	client := getClient()
	revisionID := fmt.Sprintf("%s:1", templateID)
	sourceConfigKind := workflowtemplates.WorkflowTemplateSourceConfigKindTerraform

	is409 := func(err error) bool {
		var apiErr *core.APIError
		return errors.As(err, &apiErr) && apiErr.StatusCode == 409
	}

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
		t.Fatalf("setupDriftEnabledTemplate: create template %q: %s", templateID, err)
	}

	alias := "v1"
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
				DriftCheck:       sgsdkgo.Bool(true),
				DriftCron:        sgsdkgo.String("0 */6 * * ? *"),
			},
		},
	)
	if err != nil && !is409(err) {
		client.WorkflowTemplates.DeleteWorkflowTemplate(context.TODO(), org, templateID)
		t.Fatalf("setupDriftEnabledTemplate: create revision for %q: %s", templateID, err)
	}

	_, err = client.WorkflowTemplatesRevisions.UpdateWorkflowTemplateRevision(
		context.TODO(), org, revisionID,
		&workflowtemplaterevisions.UpdateWorkflowTemplateRevisionRequest{
			IsPublic: sgsdkgo.Optional(sgsdkgo.IsPublicEnumOne),
		},
	)
	if err != nil {
		client.WorkflowTemplatesRevisions.DeleteWorkflowTemplateRevision(context.TODO(), org, revisionID, true)
		client.WorkflowTemplates.DeleteWorkflowTemplate(context.TODO(), org, templateID)
		t.Fatalf("setupDriftEnabledTemplate: publish revision %q: %s", revisionID, err)
	}

	_, err = client.WorkflowTemplates.UpdateWorkflowTemplate(
		context.TODO(), org, templateID,
		&workflowtemplates.UpdateWorkflowTemplateRequest{
			IsPublic: sgsdkgo.Optional(sgsdkgo.IsPublicEnumOne),
		},
	)
	if err != nil {
		client.WorkflowTemplatesRevisions.DeleteWorkflowTemplateRevision(context.TODO(), org, revisionID, true)
		client.WorkflowTemplates.DeleteWorkflowTemplate(context.TODO(), org, templateID)
		t.Fatalf("setupDriftEnabledTemplate: publish template %q: %s", templateID, err)
	}

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

// TestAccWorkflowUsingTemplate_DriftCronDroppedWhenCheckFalse verifies the
// drift_check/drift_cron coupling. The template revision has drift ENABLED with a cron;
// the user sets drift_check=false and does not declare a cron. The resolved drift_cron
// must be "" (the cron is meaningless when checking is off), and the plan must be stable
// afterward (no "inconsistent result after apply", no perpetual diff).
func TestAccWorkflowUsingTemplate_DriftCronDroppedWhenCheckFalse(t *testing.T) {
	templateID := setupDriftEnabledTemplate(t, "tf-provider-wf-tmpl-driftcron") + ":1"
	wfGrpName := acctest.ResourceName("tf-provider-wf-template-driftcron-wfgrp")
	id := acctest.ResourceName("tf-provider-wf-template-driftcron")

	if err := createWorkflowGroupFixture(wfGrpName); err != nil {
		t.Errorf("failed to create workflow group fixture: %s", err.Error())
	}
	defer deleteWorkflowGroupFixture(wfGrpName)
	defer deleteWorkflowUsingTemplateFixture(wfGrpName, id)

	// User disables drift_check and leaves drift_cron unset. The template would otherwise
	// supply drift_cron="0 */6 * * ? *"; the coupling must drop it.
	configCheckOff := `
  terraform_config = {
    terraform_version = "1.5.0"
    drift_check       = false
  }
`
	// Flip drift_check on with an explicit cron — the cron must now be kept.
	configCheckOn := `
  terraform_config = {
    terraform_version = "1.5.0"
    drift_check       = true
    drift_cron        = "0 */6 * * ? *"
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
				Config: testAccWorkflowUsingTemplate(wfGrpName, id, "TERRAFORM", templateID, configCheckOff),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test", "terraform_config.drift_check", "false"),
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test", "terraform_config.drift_cron", ""),
				),
			},
			{
				// Re-apply the same config: must be a no-op (coupling is idempotent/stable).
				Config:   testAccWorkflowUsingTemplate(wfGrpName, id, "TERRAFORM", templateID, configCheckOff),
				PlanOnly: true,
			},
			{
				// Enable drift with a cron — the cron is now meaningful and must be kept.
				Config: testAccWorkflowUsingTemplate(wfGrpName, id, "TERRAFORM", templateID, configCheckOn),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test", "terraform_config.drift_check", "true"),
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test", "terraform_config.drift_cron", "0 */6 * * ? *"),
				),
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
	wfGrpName := acctest.ResourceName("tf-provider-wf-template-drift-wfgrp")
	id := acctest.ResourceName("tf-provider-wf-template-drift")

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

// TestAccWorkflowUsingTemplate_RevisionUpgrade verifies that changing iac_template_id to a
// new revision re-resolves the unset (Computed) fields against the NEW revision, while
// fields the user declared in config are preserved. rev1 has env var TMPL_VAR; rev2 has
// env var REV2_VAR. The user never declares environment_variables, so after the upgrade it
// must reflect rev2's REV2_VAR (not rev1's TMPL_VAR). The user-declared description is
// preserved across the upgrade.
func TestAccWorkflowUsingTemplate_RevisionUpgrade(t *testing.T) {
	base := acctest.ResourceName("tf-provider-wf-tmpl-upgrade")
	rev1 := setupWorkflowTemplate(t, base) + ":1"
	rev2 := addSecondRevision(t, base)
	wfGrpName := acctest.ResourceName("tf-provider-wf-template-upgrade-wfgrp")
	id := acctest.ResourceName("tf-provider-wf-template-upgrade")

	if err := createWorkflowGroupFixture(wfGrpName); err != nil {
		t.Errorf("failed to create workflow group fixture: %s", err.Error())
	}
	defer deleteWorkflowGroupFixture(wfGrpName)
	defer deleteWorkflowUsingTemplateFixture(wfGrpName, id)

	// User declares only description + terraform_version; env vars come from the template.
	config := `
  description = "user-owned description"

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
				// Create on rev1 → env var resolves to TMPL_VAR.
				Config: testAccWorkflowUsingTemplate(wfGrpName, id, "TERRAFORM", rev1, config),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test", "vcs_config.iac_vcs_config.iac_template_id", rev1),
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test", "environment_variables.0.config.var_name", "TMPL_VAR"),
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test", "description", "user-owned description"),
				),
			},
			{
				// Upgrade to rev2 → unset env var re-resolves to REV2_VAR; user description preserved.
				// rev2 ADDS terraform_config drift fields (rev1 had none); they must re-resolve
				// without an inconsistent-result error while the user's terraform_version is kept.
				// rev2 has drift_check=true with a cron, so the cron is meaningful and kept.
				Config: testAccWorkflowUsingTemplate(wfGrpName, id, "TERRAFORM", rev2, config),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test", "vcs_config.iac_vcs_config.iac_template_id", rev2),
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test", "environment_variables.0.config.var_name", "REV2_VAR"),
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test", "description", "user-owned description"),
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test", "terraform_config.drift_check", "true"),
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test", "terraform_config.drift_cron", "0 */6 * * ? *"),
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test", "terraform_config.terraform_version", "1.5.0"),
				),
			},
		},
	})
}
