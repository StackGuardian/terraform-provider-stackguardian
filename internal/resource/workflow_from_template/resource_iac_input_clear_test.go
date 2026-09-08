package workflowfromtemplate_test

import (
	"context"
	"encoding/base64"
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
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

// setupTemplateWithInputDefaults creates and publishes a TERRAFORM template revision whose
// FORM_JSONSCHEMA carries property defaults (env=staging, region=eu) — so the provider's
// templateDefaultInputData resolves a non-empty default iac_input_data for workflows that
// omit it. Backs the explicit-empty (WFT-CLEAR-002) test. Registers cleanup.
func setupTemplateWithInputDefaults(t *testing.T, name string) string {
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

	formSchema := base64.StdEncoding.EncodeToString([]byte(`{
		"$schema": "http://json-schema.org/draft-07/schema#",
		"type": "object",
		"properties": {
			"env":    {"type": "string", "default": "staging"},
			"region": {"type": "string", "default": "eu"}
		}
	}`))
	isCommitted := true
	_, err = client.WorkflowTemplatesRevisions.CreateWorkflowTemplateRevision(context.TODO(), org, name,
		&workflowtemplaterevisions.CreateWorkflowTemplateRevisionsRequest{
			Alias: "v1", SourceConfigKind: &sck, IsPublic: sgsdkgo.IsPublicEnumZero.Ptr(),
			OwnerOrg: fmt.Sprintf("/orgs/%s", org),
			InputSchemas: []sgsdkgo.InputSchemas{
				{
					Type:        sgsdkgo.InputSchemasTypeEnumFormJsonschema,
					EncodedData: &formSchema,
					IsCommitted: &isCommitted,
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
	return fmt.Sprintf("/%s/%s:1", org, name)
}

// TestAccWorkflowUsingTemplate_ExplicitEmptyIacInputData verifies WFT-CLEAR-002: a template
// carries default input data (FORM_JSONSCHEMA defaults), and the user declares an EXPLICIT
// empty data ("{}") to CLEAR it instead of inheriting. Previously this failed with
// 'VCSConfig.iacInputData.data: This field is required' — the SDK's json omitempty on the
// plain-map Data field silently dropped the empty map from the payload. With Data as a
// pointer, "data": {} reaches the wire, the API accepts it, and inheritance is suppressed.
// Also asserts the omit case still inherits (control), and the clear round-trips (clean plan).
func TestAccWorkflowUsingTemplate_ExplicitEmptyIacInputData(t *testing.T) {
	templateID := setupTemplateWithInputDefaults(t, "tf-iacclear-tpl")
	wfGrp := "tf-iacclear-wfgrp"
	if err := createWorkflowGroupFixture(wfGrp); err != nil {
		t.Fatalf("wfgrp fixture: %s", err)
	}
	defer deleteWorkflowGroupFixture(wfGrp)
	defer deleteWorkflowUsingTemplateFixture(wfGrp, "tf-iacclear-inherit")
	defer deleteWorkflowUsingTemplateFixture(wfGrp, "tf-iacclear-empty")

	// Control: omits iac_input_data -> inherits the template defaults.
	// Subject: declares data = "{}" -> clears (suppresses inheritance).
	cfg := fmt.Sprintf(`
resource "stackguardian_workflow_from_template" "inherit" {
  workflow_group_id = %q
  id                = "tf-iacclear-inherit"
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

resource "stackguardian_workflow_from_template" "empty" {
  workflow_group_id = %q
  id                = "tf-iacclear-empty"
  wf_type           = "TERRAFORM"
  vcs_config = {
    iac_vcs_config = {
      iac_template_id = %q
    }
    iac_input_data = {
      schema_type = "RAW_JSON"
      data        = "{}"
    }
  }
  terraform_config = {
    terraform_version = "1.5.0"
  }
}
`, wfGrp, templateID, wfGrp, templateID)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks:   []tfversion.TerraformVersionCheck{tfversion.SkipBelow(tfversion.Version1_1_0)},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader()),
		Steps: []resource.TestStep{
			{
				Config: cfg,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Control inherits the FORM_JSONSCHEMA defaults.
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.inherit", "vcs_config.iac_input_data.data", `{"env":"staging","region":"eu"}`),
					// Subject's explicit empty is honored — template defaults NOT inherited.
					resource.TestCheckResourceAttr("stackguardian_workflow_from_template.empty", "vcs_config.iac_input_data.data", "{}"),
				),
			},
			{
				// Round-trip stability: no perpetual diff on the cleared value.
				Config:   cfg,
				PlanOnly: true,
			},
		},
	})
}
