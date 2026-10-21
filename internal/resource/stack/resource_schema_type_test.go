package stack_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"testing"

	"github.com/StackGuardian/terraform-provider-stackguardian/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// These tests run real `terraform` steps against the in-memory stack API fake in
// resource_override_removal_test.go (no credentials, no TF_ACC).

func noCodeJSONConfig(apiURI, description string) string {
	return iacInputDataConfig(apiURI, description, `{
          schema_type = "NO_CODE_JSON"
          data        = jsonencode({ region = "eu-central-1" })
        }`)
}

func iacInputDataConfig(apiURI, description, iacInputData string) string {
	return fmt.Sprintf(`
provider "stackguardian" {
  api_key  = "test"
  org_name = %q
  api_uri  = %q
}

resource "stackguardian_stack" "test" {
  id                = "override-removal"
  workflow_group_id = "override-removal-wfgrp"
  template_group_id = "tpl:1"
  workflows_config = {
    workflows = [{
      id          = "wf-a"
      description = %q
      vcs_config = {
        iac_input_data = %s
      }
    }]
  }
}
`, overrideRemovalOrg, apiURI, description, iacInputData)
}

// sentIacInputData returns vcs_config.iac_input_data of the first workflow in body, as sent.
func sentIacInputData(body map[string]any) (map[string]any, bool) {
	wfc, _ := body["WorkflowsConfig"].(map[string]any)
	workflows, _ := wfc["workflows"].([]any)
	if len(workflows) == 0 {
		return nil, false
	}
	wf, _ := workflows[0].(map[string]any)
	vcs, _ := wf["VCSConfig"].(map[string]any)
	input, ok := vcs["iacInputData"].(map[string]any)
	return input, ok
}

// sentSchemaType returns vcs_config.iac_input_data.schemaType of the first workflow in body.
func sentSchemaType(body map[string]any) any {
	input, _ := sentIacInputData(body)
	return input["schemaType"]
}

// sentEmptyIacInputData checks that body carries iacInputData: {} for the first workflow.
func sentEmptyIacInputData(what string, body map[string]any) error {
	input, ok := sentIacInputData(body)
	if !ok || len(input) != 0 {
		sent, _ := json.Marshal(input)
		return fmt.Errorf("%s with iacInputData %s, want {}", what, sent)
	}
	return nil
}

// NO_CODE_JSON is a schema type the platform accepts and returns, though the SDK enum lacks it:
// a stack carrying it is created and updated with the value sent unchanged.
func TestStack_NoCodeJSONSchemaTypeIsSent(t *testing.T) {
	srv, api := newFakeStackAPI(t, createFailureRevision)

	runFakeAPITest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProviderFactories(http.Header{}),
		Steps: []resource.TestStep{
			{
				Config: noCodeJSONConfig(srv.URL, "first"),
				Check: func(*terraform.State) error {
					if got := sentSchemaType(api.storedStack("override-removal")); got != "NO_CODE_JSON" {
						return fmt.Errorf("created with schemaType %v, want NO_CODE_JSON", got)
					}
					return nil
				},
			},
			{
				Config: noCodeJSONConfig(srv.URL, "second"),
				Check: func(*terraform.State) error {
					if got := sentSchemaType(api.lastPatch()); got != "NO_CODE_JSON" {
						return fmt.Errorf("updated with schemaType %v, want NO_CODE_JSON", got)
					}
					return nil
				},
			},
		},
	})
}

// iac_input_data = {} is StackGuardian's "no inputs": it is sent as {}, read back as an object
// with nothing set, and plans clean.
func TestStack_EmptyIacInputDataSentAsNoInputs(t *testing.T) {
	srv, api := newFakeStackAPI(t, createFailureRevision)

	runFakeAPITest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProviderFactories(http.Header{}),
		Steps: []resource.TestStep{
			{
				Config: iacInputDataConfig(srv.URL, "first", "{}"),
				Check: resource.ComposeAggregateTestCheckFunc(
					func(*terraform.State) error {
						return sentEmptyIacInputData("created", api.storedStack("override-removal"))
					},
					resource.TestCheckNoResourceAttr("stackguardian_stack.test", "workflows_config.workflows.0.vcs_config.iac_input_data.schema_type"),
				),
			},
			{
				Config: iacInputDataConfig(srv.URL, "second", "{}"),
				Check: func(*terraform.State) error {
					return sentEmptyIacInputData("updated", api.lastPatch())
				},
			},
		},
	})
}

// Input data with only one of schema_type and data fails the plan, as StackGuardian would
// reject it at apply.
func TestStack_IncompleteIacInputDataRejectedAtPlan(t *testing.T) {
	srv, _ := newFakeStackAPI(t, createFailureRevision)

	for name, input := range map[string]string{
		"schema_type without data": `{ schema_type = "RAW_JSON" }`,
		"data without schema_type": `{ data = jsonencode({ a = 1 }) }`,
	} {
		t.Run(name, func(t *testing.T) {
			runFakeAPITest(t, resource.TestCase{
				ProtoV6ProviderFactories: acctest.ProviderFactories(http.Header{}),
				Steps: []resource.TestStep{{
					Config:      iacInputDataConfig(srv.URL, "first", input),
					PlanOnly:    true,
					ExpectError: regexp.MustCompile(`Incomplete iac_input_data`),
				}},
			})
		})
	}
}
