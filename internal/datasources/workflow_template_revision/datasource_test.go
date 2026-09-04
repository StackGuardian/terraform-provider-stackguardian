package workflowtemplaterevision_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	sgsdkgo "github.com/StackGuardian/sg-sdk-go"
	"github.com/StackGuardian/sg-sdk-go/core"
	"github.com/StackGuardian/sg-sdk-go/workflowsteptemplate"
	"github.com/StackGuardian/sg-sdk-go/workflowtemplaterevisions"
	"github.com/StackGuardian/sg-sdk-go/workflowtemplates"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/acctest"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

var org = config.Get().OrgName

// setupPopulatedRevision creates and publishes a CUSTOM template revision populated with the
// fields a workflow_from_template user would want to read back for override/merge logic:
// description, tags, env_vars, runner_constraints, a POPULATED deployment_platform_config
// (AZURE_OIDC list — the field whose data source schema was previously a single object and
// errored at runtime), terraform_config, mini_steps, and top-level wf_steps_config. Returns
// the revision id in the BARE "<name>:1" form the read endpoint expects (the full
// "/<org>/<name>:1" path returns a misleading 401). Registers cleanup.
func setupPopulatedRevision(t *testing.T, name, stepTemplateID string) string {
	t.Helper()
	client := acctest.SGClient()
	sck := workflowtemplates.WorkflowTemplateSourceConfigKindCustom
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

	desc := "datasource-desc"
	envText := "ds-var-value"
	dpcIntegration := "akash-azure-oidc" // pre-existing QA integration
	napprovals := 1
	cpu := 512
	mem := 1024
	approval := false
	to := 2100

	_, err = client.WorkflowTemplatesRevisions.CreateWorkflowTemplateRevision(context.TODO(), org, name,
		&workflowtemplaterevisions.CreateWorkflowTemplateRevisionsRequest{
			Alias: "v1", SourceConfigKind: &sck, IsPublic: sgsdkgo.IsPublicEnumZero.Ptr(),
			OwnerOrg:                  fmt.Sprintf("/orgs/%s", org),
			LongDescription:           &desc,
			Tags:                      []string{"alpha", "beta"},
			Approvers:                 []string{"akashsuresh0510@gmail.com"},
			ContextTags:               map[string]string{"env": "test"},
			NumberOfApprovalsRequired: &napprovals,
			UserJobCPU:                &cpu,
			UserJobMemory:             &mem,
			RunnerConstraints: &sgsdkgo.RunnerConstraints{
				Type: sgsdkgo.RunnerConstraintsTypeEnumShared.Ptr(),
			},
			DeploymentPlatformConfig: []*workflowtemplaterevisions.DeploymentPlatformConfig{
				{
					Kind: workflowtemplaterevisions.DeploymentPlatformConfigKindEnumAzureOidc,
					Config: workflowtemplaterevisions.DeploymentPlatformConfigConfig{
						IntegrationId: fmt.Sprintf("/integrations/%s", dpcIntegration),
						ProfileName:   &dpcIntegration,
					},
				},
			},
			EnvironmentVariables: []sgsdkgo.EnvVars{
				{
					Kind:   sgsdkgo.EnvVarsKindEnumPlainText,
					Config: &sgsdkgo.EnvVarConfig{VarName: "DS_VAR", TextValue: &envText},
				},
			},
			WfStepsConfig: []sgsdkgo.WfStepsConfig{
				{
					Name:             sgsdkgo.String("step-one"),
					WfStepTemplateId: &stepTemplateID,
					Approval:         &approval,
					Timeout:          &to,
					MountPoints:      []sgsdkgo.MountPoint{},
				},
			},
		})
	if err != nil && !is409(err) {
		t.Fatalf("create rev: %s", err)
	}

	revisionID := name + ":1"
	if _, err := client.WorkflowTemplatesRevisions.UpdateWorkflowTemplateRevision(context.TODO(), org, revisionID,
		&workflowtemplaterevisions.UpdateWorkflowTemplateRevisionRequest{IsPublic: sgsdkgo.Optional(sgsdkgo.IsPublicEnumOne)}); err != nil {
		t.Fatalf("publish rev: %s", err)
	}
	if _, err := client.WorkflowTemplates.UpdateWorkflowTemplate(context.TODO(), org, name,
		&workflowtemplates.UpdateWorkflowTemplateRequest{IsPublic: sgsdkgo.Optional(sgsdkgo.IsPublicEnumOne)}); err != nil {
		t.Fatalf("publish tpl: %s", err)
	}

	t.Cleanup(func() {
		eff := fmt.Sprintf("%d", time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC).Unix())
		msg := "cleanup"
		_, _ = client.WorkflowTemplatesRevisions.UpdateWorkflowTemplateRevision(context.TODO(), org, revisionID,
			&workflowtemplaterevisions.UpdateWorkflowTemplateRevisionRequest{
				Deprecation: sgsdkgo.Optional(workflowtemplaterevisions.Deprecation{EffectiveDate: &eff, Message: &msg})})
		_ = client.WorkflowTemplatesRevisions.DeleteWorkflowTemplateRevision(context.TODO(), org, revisionID, true)
		_ = client.WorkflowTemplates.DeleteWorkflowTemplate(context.TODO(), org, name)
	})

	return revisionID // bare "<name>:1"
}

// setupStepTemplate creates a referenceable wf step template ":1" for wf_steps_config.
func setupStepTemplate(t *testing.T, name string) string {
	t.Helper()
	client := acctest.SGClient()
	is409 := func(err error) bool {
		var apiErr *core.APIError
		return errors.As(err, &apiErr) && apiErr.StatusCode == 409
	}
	sourceKind := workflowsteptemplate.WorkflowStepTemplateSourceConfigKindDockerImageEnum
	isPublic := workflowsteptemplate.IsPublicEnumOne
	isPrivate := false
	_, err := client.WorkflowStepTemplate.CreateWorkflowStepTemplate(context.TODO(), org, true,
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
		})
	if err != nil && !is409(err) {
		t.Fatalf("create step tpl: %s", err)
	}
	t.Cleanup(func() {
		_ = client.WorkflowStepTemplate.DeleteWorkflowStepTemplate(context.TODO(), org, name)
	})
	return fmt.Sprintf("/%s/%s:1", org, name)
}

// TestAccWorkflowTemplateRevisionDataSource_Custom verifies the data source Matthias asked
// for — a data element to fetch a template revision's default values (for override/merge logic
// in Terraform on upgrade). It reads a CUSTOM-source revision populated with the tricky fields
// (incl. top-level wf_steps_config, which is CUSTOM-only) and asserts they flatten correctly.
// The deployment_platform_config assertions are the regression guard: that attribute was
// declared as a single nested object while the model is a list, which errored at runtime for
// any revision carrying a populated deployment_platform_config. The TERRAFORM-source path
// (populated terraform_config) is covered by _Terraform below.
func TestAccWorkflowTemplateRevisionDataSource_Custom(t *testing.T) {
	stepTemplateID := setupStepTemplate(t, "tf-ds-wtr-step")
	revisionID := setupPopulatedRevision(t, "tf-ds-wtr-tpl", stepTemplateID)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(http.Header{}),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
data "stackguardian_workflow_template_revision" "test" {
  id = %q
}
`, revisionID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.stackguardian_workflow_template_revision.test", "description", "datasource-desc"),
					resource.TestCheckResourceAttr("data.stackguardian_workflow_template_revision.test", "tags.#", "2"),
					resource.TestCheckResourceAttr("data.stackguardian_workflow_template_revision.test", "user_job_cpu", "512"),
					resource.TestCheckResourceAttr("data.stackguardian_workflow_template_revision.test", "number_of_approvals_required", "1"),
					resource.TestCheckResourceAttr("data.stackguardian_workflow_template_revision.test", "environment_variables.0.config.var_name", "DS_VAR"),
					resource.TestCheckResourceAttr("data.stackguardian_workflow_template_revision.test", "runner_constraints.type", "shared"),
					// Regression guard: populated deployment_platform_config as a LIST.
					resource.TestCheckResourceAttr("data.stackguardian_workflow_template_revision.test", "deployment_platform_config.#", "1"),
					resource.TestCheckResourceAttr("data.stackguardian_workflow_template_revision.test", "deployment_platform_config.0.kind", "AZURE_OIDC"),
					resource.TestCheckResourceAttr("data.stackguardian_workflow_template_revision.test", "deployment_platform_config.0.config.integration_id", "/integrations/akash-azure-oidc"),
					resource.TestCheckResourceAttr("data.stackguardian_workflow_template_revision.test", "wf_steps_config.0.name", "step-one"),
					resource.TestCheckResourceAttr("data.stackguardian_workflow_template_revision.test", "wf_steps_config.0.wf_step_template_id", stepTemplateID),
				),
			},
		},
	})
}

// setupTerraformRevision creates and publishes a TERRAFORM-source template revision with a
// populated terraform_config (version + drift + managed state) and a git RuntimeSource — the
// fields a TERRAFORM template carries but a CUSTOM one does not. Exercises the terraform_config
// + runtime_source flatten path (the largest nested object) that the CUSTOM test never hits.
// Returns the bare "<name>:1" revision id. Registers cleanup.
func setupTerraformRevision(t *testing.T, name string) string {
	t.Helper()
	client := acctest.SGClient()
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

	tfVer := "TERRAFORM-1.5.7"
	driftCron := "0 */6 * * ? *"
	driftCheck := true
	managed := true
	envText := "var1"
	_, err = client.WorkflowTemplatesRevisions.CreateWorkflowTemplateRevision(context.TODO(), org, name,
		&workflowtemplaterevisions.CreateWorkflowTemplateRevisionsRequest{
			Alias: "v1", SourceConfigKind: &sck, IsPublic: sgsdkgo.IsPublicEnumZero.Ptr(),
			OwnerOrg: fmt.Sprintf("/orgs/%s", org),
			RunnerConstraints: &sgsdkgo.RunnerConstraints{
				Type: sgsdkgo.RunnerConstraintsTypeEnumShared.Ptr(),
			},
			RuntimeSource: &workflowtemplates.RuntimeSource{
				SourceConfigDestKind: workflowtemplates.SourceConfigDestKindEnumGitOther.Ptr(),
				Config: &workflowtemplates.RuntimeSourceConfig{
					Repo: "https://github.com/AkashS0510/terraform-null-cat",
				},
			},
			TerraformConfig: &sgsdkgo.TerraformConfig{
				TerraformVersion:      &tfVer,
				DriftCheck:            &driftCheck,
				DriftCron:             &driftCron,
				ManagedTerraformState: &managed,
			},
			EnvironmentVariables: []sgsdkgo.EnvVars{
				{
					Kind:   sgsdkgo.EnvVarsKindEnumPlainText,
					Config: &sgsdkgo.EnvVarConfig{VarName: "test1", TextValue: &envText},
				},
			},
		})
	if err != nil && !is409(err) {
		t.Fatalf("create rev: %s", err)
	}

	revisionID := name + ":1"
	if _, err := client.WorkflowTemplatesRevisions.UpdateWorkflowTemplateRevision(context.TODO(), org, revisionID,
		&workflowtemplaterevisions.UpdateWorkflowTemplateRevisionRequest{IsPublic: sgsdkgo.Optional(sgsdkgo.IsPublicEnumOne)}); err != nil {
		t.Fatalf("publish rev: %s", err)
	}
	if _, err := client.WorkflowTemplates.UpdateWorkflowTemplate(context.TODO(), org, name,
		&workflowtemplates.UpdateWorkflowTemplateRequest{IsPublic: sgsdkgo.Optional(sgsdkgo.IsPublicEnumOne)}); err != nil {
		t.Fatalf("publish tpl: %s", err)
	}

	t.Cleanup(func() {
		eff := fmt.Sprintf("%d", time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC).Unix())
		msg := "cleanup"
		_, _ = client.WorkflowTemplatesRevisions.UpdateWorkflowTemplateRevision(context.TODO(), org, revisionID,
			&workflowtemplaterevisions.UpdateWorkflowTemplateRevisionRequest{
				Deprecation: sgsdkgo.Optional(workflowtemplaterevisions.Deprecation{EffectiveDate: &eff, Message: &msg})})
		_ = client.WorkflowTemplatesRevisions.DeleteWorkflowTemplateRevision(context.TODO(), org, revisionID, true)
		_ = client.WorkflowTemplates.DeleteWorkflowTemplate(context.TODO(), org, name)
	})

	return revisionID
}

// TestAccWorkflowTemplateRevisionDataSource_Terraform reads a TERRAFORM-source revision through
// the data source and asserts the terraform_config + runtime_source fields flatten with no type
// mismatch. This is the other half of the two-kind coverage: CUSTOM templates carry top-level
// wf_steps_config, TERRAFORM templates carry a populated terraform_config (the largest nested
// object). terraform_version is normalized to the BARE form ("1.5.7", not "TERRAFORM-1.5.7")
// by the data source Read so it feeds workflow_from_template without a perpetual diff — this
// assertion is the guard for that normalization.
func TestAccWorkflowTemplateRevisionDataSource_Terraform(t *testing.T) {
	revisionID := setupTerraformRevision(t, "tf-ds-wtr-tftpl")

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(http.Header{}),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
data "stackguardian_workflow_template_revision" "tf" {
  id = %q
}
`, revisionID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.stackguardian_workflow_template_revision.tf", "source_config_kind", "TERRAFORM"),
					// terraform_config (the largest nested object) flattens fully. terraform_version
					// is normalized to the bare form (prefix stripped) by the data source Read.
					resource.TestCheckResourceAttr("data.stackguardian_workflow_template_revision.tf", "terraform_config.terraform_version", "1.5.7"),
					resource.TestCheckResourceAttr("data.stackguardian_workflow_template_revision.tf", "terraform_config.drift_check", "true"),
					resource.TestCheckResourceAttr("data.stackguardian_workflow_template_revision.tf", "terraform_config.drift_cron", "0 */6 * * ? *"),
					resource.TestCheckResourceAttr("data.stackguardian_workflow_template_revision.tf", "terraform_config.managed_terraform_state", "true"),
					resource.TestCheckResourceAttr("data.stackguardian_workflow_template_revision.tf", "environment_variables.0.config.var_name", "test1"),
					resource.TestCheckResourceAttr("data.stackguardian_workflow_template_revision.tf", "runner_constraints.type", "shared"),
					resource.TestCheckResourceAttr("data.stackguardian_workflow_template_revision.tf", "runtime_source.source_config_dest_kind", "GIT_OTHER"),
				),
			},
		},
	})
}
