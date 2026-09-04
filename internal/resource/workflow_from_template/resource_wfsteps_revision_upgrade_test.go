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

// setupTwoRevIdenticalWfStepsConfig builds a CUSTOM template whose rev1 and rev2 carry an
// IDENTICAL top-level WfStepsConfig (three steps, mirroring the real "reordered steps"
// custom template: one plain step, one with wfStepInputData, one with cmdOverride). The
// revisions differ ONLY in Notes (metadata, not template-derived to any workflow attribute),
// so the upgrade is a real revision change while every workflow-visible field — including
// wf_steps_config — is unchanged. Used to verify a dependent referencing the COMPUTED
// wf_steps_config is NOT spuriously updated on upgrade (the field now resolves concretely at
// plan time). stepTemplateID must be a real, referenceable wf step template ":1".
func setupTwoRevIdenticalWfStepsConfig(t *testing.T, name, stepTemplateID string) (rev1, rev2 string) {
	t.Helper()
	client := getClient()
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

	// Identical steps on both revisions. approval/timeout/cmdOverride/wfStepInputData exercise
	// the fields the flattener round-trips; the middle step carries wfStepInputData (whose
	// data map the API reorders — order-independent after unmarshal, which is what we prove).
	steps := func() []sgsdkgo.WfStepsConfig {
		approval := false
		to := 2100
		cmd := "printenv && echo done"
		schemaType := sgsdkgo.WfStepInputDataSchemaTypeEnumFormJSONSchema
		return []sgsdkgo.WfStepsConfig{
			{
				Name:             sgsdkgo.String("step-plain"),
				WfStepTemplateId: &stepTemplateID,
				Approval:         &approval,
				Timeout:          &to,
				MountPoints:      []sgsdkgo.MountPoint{},
			},
			{
				Name:             sgsdkgo.String("step-input"),
				WfStepTemplateId: &stepTemplateID,
				Approval:         &approval,
				Timeout:          &to,
				MountPoints:      []sgsdkgo.MountPoint{},
				WfStepInputData: &sgsdkgo.WfStepInputData{
					SchemaType: &schemaType,
					Data: map[string]interface{}{
						"name":  "workflow-default",
						"age":   float64(99),
						"agree": true,
					},
				},
			},
			{
				Name:             sgsdkgo.String("step-cmd"),
				WfStepTemplateId: &stepTemplateID,
				Approval:         &approval,
				Timeout:          &to,
				MountPoints:      []sgsdkgo.MountPoint{},
				CmdOverride:      &cmd,
			},
		}
	}

	mk := func(alias, notes string) {
		_, err := client.WorkflowTemplatesRevisions.CreateWorkflowTemplateRevision(context.TODO(), org, name,
			&workflowtemplaterevisions.CreateWorkflowTemplateRevisionsRequest{
				Alias: alias, SourceConfigKind: &sck, IsPublic: sgsdkgo.IsPublicEnumZero.Ptr(),
				OwnerOrg:      fmt.Sprintf("/orgs/%s", org),
				Notes:         notes, // only-differing field -> real revision change
				WfStepsConfig: steps(),
			})
		if err != nil && !is409(err) {
			t.Fatalf("create rev %s: %s", alias, err)
		}
	}
	mk("v1", "rev one")
	mk("v2", "rev two") // only Notes differs

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
			_, _ = client.WorkflowTemplatesRevisions.UpdateWorkflowTemplateRevision(context.TODO(), org, rev,
				&workflowtemplaterevisions.UpdateWorkflowTemplateRevisionRequest{
					Deprecation: sgsdkgo.Optional(workflowtemplaterevisions.Deprecation{EffectiveDate: &eff, Message: &msg})})
			_ = client.WorkflowTemplatesRevisions.DeleteWorkflowTemplateRevision(context.TODO(), org, rev, true)
		}
		_ = client.WorkflowTemplates.DeleteWorkflowTemplate(context.TODO(), org, name)
	})
	return fmt.Sprintf("/%s/%s:1", org, name), fmt.Sprintf("/%s/%s:2", org, name)
}

// TestAccWorkflowUsingTemplate_WfStepsConfigRevisionUpgradeNoSpuriousDependentUpdate verifies
// the wf_steps_config concrete-resolution fix: upgrading test1 :1 -> :2 where the
// template-derived top-level wf_steps_config is IDENTICAL across revisions must NOT spuriously
// update test2, which references test1.wf_steps_config. test1 itself updates (iac_template_id
// changes), but test2 must be NoOp. Also proves apply succeeds (no "inconsistent result after
// apply") — the round-trip that was previously unproven, which is why the field was left
// unknown. wf_steps_config is CUSTOM-only, so both workflows are wf_type = "CUSTOM".
func TestAccWorkflowUsingTemplate_WfStepsConfigRevisionUpgradeNoSpuriousDependentUpdate(t *testing.T) {
	stepTemplateID := setupWorkflowStepTemplate(t, "tf-wfsteps-fixverify-step")
	rev1, rev2 := setupTwoRevIdenticalWfStepsConfig(t, "tf-wfsteps-fixverify-tpl", stepTemplateID)
	wfGrp := "tf-wfsteps-fixverify-wfgrp"
	if err := createWorkflowGroupFixture(wfGrp); err != nil {
		t.Fatalf("wfgrp fixture: %s", err)
	}
	defer deleteWorkflowGroupFixture(wfGrp)
	defer deleteWorkflowUsingTemplateFixture(wfGrp, "tf-wfsteps-fixverify-1")
	defer deleteWorkflowUsingTemplateFixture(wfGrp, "tf-wfsteps-fixverify-2")

	cfg := func(rev string) string {
		return fmt.Sprintf(`
resource "stackguardian_workflow_from_template" "test1" {
  workflow_group_id = %q
  id                = "tf-wfsteps-fixverify-1"
  wf_type           = "CUSTOM"
  vcs_config = {
    iac_vcs_config = {
      iac_template_id = %q
    }
  }
}

resource "stackguardian_workflow_from_template" "test2" {
  workflow_group_id = %q
  id                = "tf-wfsteps-fixverify-2"
  wf_type           = "CUSTOM"
  vcs_config = {
    iac_vcs_config = {
      iac_template_id = %q
    }
  }
  # References test1's COMPUTED, template-derived wf_steps_config (unchanged across
  # rev1->rev2). Must resolve concretely so this stays NoOp on test1's upgrade.
  wf_steps_config = stackguardian_workflow_from_template.test1.wf_steps_config
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
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test1", "wf_steps_config.0.name", "step-plain"),
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test1", "wf_steps_config.1.name", "step-input"),
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test1", "wf_steps_config.2.name", "step-cmd"),
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test2", "wf_steps_config.0.name", "step-plain"),
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.test2", "wf_steps_config.2.cmd_override", "printenv && echo done"),
				),
			},
			{
				// Upgrade test1 -> rev2. wf_steps_config identical, so test2 (referencing
				// test1.wf_steps_config) must be a NoOp. test1 itself updates.
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
