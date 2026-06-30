package workflowfromtemplate_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	sgsdkgo "github.com/StackGuardian/sg-sdk-go"
	"github.com/StackGuardian/sg-sdk-go/core"
	"github.com/StackGuardian/sg-sdk-go/workflowtemplaterevisions"
	"github.com/StackGuardian/sg-sdk-go/workflowtemplates"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

// setupTwoRevIdenticalDerivedFields builds a template whose rev1 and rev2 are IDENTICAL in every
// template-derived scalar/simple field we now compute concretely at plan time (description,
// tags, approvers, context_tags, env vars, approvals, cpu/mem), differing ONLY in
// terraform_version (so the upgrade is a real revision change). Used to verify that a
// dependent referencing those unchanged fields is NOT spuriously updated on upgrade.
func setupTwoRevIdenticalDerivedFields(t *testing.T, name string) (rev1, rev2 string) {
	t.Helper()
	client := getClient()
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

	desc := "stable-desc"
	envText := "stable-var-value"
	napprovals := 1
	cpu := 512
	mem := 1024
	mk := func(alias, tfver string) {
		d := desc
		na, c, m := napprovals, cpu, mem
		_, err := client.WorkflowTemplatesRevisions.CreateWorkflowTemplateRevision(context.TODO(), org, name,
			&workflowtemplaterevisions.CreateWorkflowTemplateRevisionsRequest{
				Alias: alias, SourceConfigKind: &sck, IsPublic: sgsdkgo.IsPublicEnumZero.Ptr(),
				OwnerOrg:                  fmt.Sprintf("/orgs/%s", org),
				LongDescription:           &d,
				Tags:                      []string{"alpha", "beta"},
				Approvers:                 []string{"akashsuresh0510@gmail.com"},
				ContextTags:               map[string]string{"env": "test", "team": "core"},
				NumberOfApprovalsRequired: &na,
				UserJobCPU:                &c,
				UserJobMemory:             &m,
				TerraformConfig:           &sgsdkgo.TerraformConfig{TerraformVersion: &tfver},
				RunnerConstraints: &sgsdkgo.RunnerConstraints{
					Type: sgsdkgo.RunnerConstraintsTypeEnumShared.Ptr(),
				},
				Ministeps: &workflowtemplaterevisions.Ministeps{
					Notifications: &workflowtemplaterevisions.MinistepsNotifications{
						Email: &workflowtemplaterevisions.MinistepsNotificationsEmail{
							COMPLETED: []workflowtemplaterevisions.MinistepsNotificationRecepients{
								{Recipients: []string{"akashsuresh0510@gmail.com"}},
							},
						},
					},
				},
				EnvironmentVariables: []sgsdkgo.EnvVars{
					{
						Kind:   sgsdkgo.EnvVarsKindEnumPlainText,
						Config: &sgsdkgo.EnvVarConfig{VarName: "STABLE_VAR", TextValue: &envText},
					},
				},
			})
		if err != nil && !is409(err) {
			t.Fatalf("create rev %s: %s", alias, err)
		}
	}
	mk("v1", "1.5.0")
	mk("v2", "1.6.0") // only terraform_version differs

	for _, rev := range []string{name + ":1", name + ":2"} {
		if _, err := client.WorkflowTemplatesRevisions.UpdateWorkflowTemplateRevision(context.TODO(), org, rev,
			&workflowtemplaterevisions.UpdateWorkflowTemplateRevisionRequest{IsPublic: sgsdkgo.Optional(sgsdkgo.IsPublicEnumOne)}); err != nil {
			t.Fatalf("publish %s: %s", rev, err)
		}
	}
	if _, err := client.WorkflowTemplates.UpdateWorkflowTemplate(context.TODO(), org, name,
		&workflowtemplates.UpdateWorkflowTemplateRequest{IsPublic: sgsdkgo.Optional(sgsdkgo.IsPublicEnumOne)}); err != nil {
		t.Fatalf("publish tpl: %s", err)
	}
	t.Cleanup(func() {
		eff := fmt.Sprintf("%d", time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC).Unix())
		msg := "cleanup"
		for _, rev := range []string{name + ":1", name + ":2"} {
			client.WorkflowTemplatesRevisions.UpdateWorkflowTemplateRevision(context.TODO(), org, rev,
				&workflowtemplaterevisions.UpdateWorkflowTemplateRevisionRequest{
					Deprecation: sgsdkgo.Optional(workflowtemplaterevisions.Deprecation{EffectiveDate: &eff, Message: &msg})})
			client.WorkflowTemplatesRevisions.DeleteWorkflowTemplateRevision(context.TODO(), org, rev, true)
		}
		client.WorkflowTemplates.DeleteWorkflowTemplate(context.TODO(), org, name)
	})
	return fmt.Sprintf("/%s/%s:1", org, name), fmt.Sprintf("/%s/%s:2", org, name)
}

// TestAccWorkflowUsingTemplate_RevisionUpgradeNoSpuriousDependentUpdate verifies the fix: upgrading test1 :1 -> :2 where
// the template-derived scalar/simple fields are IDENTICAL across revisions must NOT spuriously
// update test2, which references those fields. test1 itself updates (terraform_version
// changes + iac_template_id), but test2 must be NoOp. Also proves apply succeeds (no
// "inconsistent result after apply") by running the upgrade as a real apply step.
func TestAccWorkflowUsingTemplate_RevisionUpgradeNoSpuriousDependentUpdate(t *testing.T) {
	rev1, rev2 := setupTwoRevIdenticalDerivedFields(t, "tf-fixverify-tpl")
	wfGrp := "tf-fixverify-wfgrp"
	if err := createWorkflowGroupFixture(wfGrp); err != nil {
		t.Fatalf("wfgrp fixture: %s", err)
	}
	defer deleteWorkflowGroupFixture(wfGrp)
	defer deleteWorkflowUsingTemplateFixture(wfGrp, "tf-fixverify-1")
	defer deleteWorkflowUsingTemplateFixture(wfGrp, "tf-fixverify-2")

	cfg := func(rev string) string {
		return fmt.Sprintf(`
resource "stackguardian_workflow_from_template" "test1" {
  workflow_group_id = %q
  id                = "tf-fixverify-1"
  wf_type           = "TERRAFORM"
  vcs_config = {
    iac_vcs_config = {
      iac_template_id = %q
    }
  }
  terraform_config = {
    terraform_version = "1.5.0"
  }
}

resource "stackguardian_workflow_from_template" "test2" {
  workflow_group_id = %q
  id                = "tf-fixverify-2"
  wf_type           = "TERRAFORM"
  vcs_config = {
    iac_vcs_config = {
      iac_template_id = %q
    }
  }
  # References test1's COMPUTED, template-derived env vars (unchanged across rev1->rev2).
  environment_variables = stackguardian_workflow_from_template.test1.environment_variables
  # Also reference a template-derived int field (user_job_cpu) and a nested object
  # (runner_constraints) the template carries, to prove they resolve concretely and keep
  # dependents NoOp when unchanged.
  user_job_cpu               = stackguardian_workflow_from_template.test1.user_job_cpu
  runner_constraints         = stackguardian_workflow_from_template.test1.runner_constraints
  deployment_platform_config = stackguardian_workflow_from_template.test1.deployment_platform_config
  mini_steps                 = stackguardian_workflow_from_template.test1.mini_steps
  terraform_config = {
    terraform_version = "1.5.0"
  }
}
`, wfGrp, rev, wfGrp, rev1) // test2 always stays on rev1
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks:   []tfversion.TerraformVersionCheck{tfversion.SkipBelow(tfversion.Version1_1_0)},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader()),
		Steps: []resource.TestStep{
			{
				// Create both on rev1.
				Config: cfg(rev1),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test1", "environment_variables.0.config.var_name", "STABLE_VAR"),
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test2", "environment_variables.0.config.var_name", "STABLE_VAR"),
				),
			},
			{
				// Upgrade test1 -> rev2. Scalars/env vars identical, so test2 (referencing
				// test1.environment_variables) must be a NoOp. test1 itself updates.
				Config: cfg(rev2),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("stackguardian_workflow_from_template.test2", plancheck.ResourceActionNoop),
						plancheck.ExpectResourceAction("stackguardian_workflow_from_template.test1", plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				// Re-apply rev2 config: fully stable, no diffs anywhere.
				Config:   cfg(rev2),
				PlanOnly: true,
			},
		},
	})
}
