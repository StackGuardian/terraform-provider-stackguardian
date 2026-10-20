package stack_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"testing"

	"github.com/StackGuardian/terraform-provider-stackguardian/internal/acctest"
	stackresource "github.com/StackGuardian/terraform-provider-stackguardian/internal/resource/stack"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// These tests run real `terraform` steps against the in-memory stack API fake in
// resource_override_removal_test.go (no credentials, no TF_ACC).

const (
	networkSlot = "aaaa1111-0000-4000-8000-000000000001"
	appSlot     = "bbbb2222-0000-4000-8000-000000000002"
)

// twoWorkflowRevision is a stack template revision whose two workflows run the
// same workflow template, so the provider computes their workflow ids.
const twoWorkflowRevision = `{
  "WorkflowsConfig": {"workflows": [
    {"id": "` + networkSlot + `", "ResourceName": "network",
     "VCSConfig": {"iacVCSConfig": {"useMarketplaceTemplate": true, "iacTemplateId": "/override-removal-org/net-tpl:1"}}},
    {"id": "` + appSlot + `", "ResourceName": "app",
     "VCSConfig": {"iacVCSConfig": {"useMarketplaceTemplate": true, "iacTemplateId": "/override-removal-org/net-tpl:1"}}}
  ]}
}`

func inputDataConfig(apiURI, appData, appEnvValue, appDescription string) string {
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
    workflows = [
      { id = %q },
      {
        id          = %q
        description = %q
        vcs_config = {
          iac_input_data = {
            schema_type = "RAW_JSON"
            data        = %s
          }
        }
        environment_variables = [
          { kind = "PLAIN_TEXT", config = { var_name = "VPC_ID", text_value = %q } }
        ]
      }
    ]
  }
}
`, overrideRemovalOrg, apiURI, networkSlot, appSlot, appDescription, appData, appEnvValue)
}

// storedAppWorkflow returns the second workflow of the stack the fake stored.
func storedAppWorkflow(t *testing.T, api *fakeStackAPI) map[string]any {
	t.Helper()
	stack := api.storedStack("override-removal")
	wfc, _ := stack["WorkflowsConfig"].(map[string]any)
	workflows, _ := wfc["workflows"].([]any)
	if len(workflows) != 2 {
		t.Fatalf("stored workflows = %v, want 2", workflows)
	}
	wf, _ := workflows[1].(map[string]any)
	return wf
}

func TestStack_InputDataReferencesResolvedAndKept(t *testing.T) {
	srv, api := newFakeStackAPI(t, twoWorkflowRevision)

	// A reference to the first workflow's output, written the way stack templates write it.
	// In HCL the leading $ is doubled, or Terraform would read ${...} as its own interpolation.
	reference := "${reference::template-group.0.outputs.vpc_id}"
	hclReference := "$" + reference
	data := fmt.Sprintf(`jsonencode({ vpc_id = %q, region = "eu-central-1" })`, hclReference)
	resolved := fmt.Sprintf("${workflow::override-removal-wfgrp.override-removal.%s.outputs.vpc_id}",
		stackresource.ComputeWorkflowId("/override-removal-org/net-tpl:1", networkSlot))

	// The platform must receive the resolved reference, never the template form.
	checkStoredResolved := func(s *terraform.State) error {
		wf := storedAppWorkflow(t, api)
		vcs, _ := wf["VCSConfig"].(map[string]any)
		input, _ := vcs["iacInputData"].(map[string]any)
		inputs, _ := input["data"].(map[string]any)
		if inputs["vpc_id"] != resolved {
			return fmt.Errorf("stored iacInputData.data.vpc_id = %v, want %q", inputs["vpc_id"], resolved)
		}
		envs, _ := wf["EnvironmentVariables"].([]any)
		env, _ := envs[0].(map[string]any)
		cfg, _ := env["config"].(map[string]any)
		if cfg["textValue"] != resolved {
			return fmt.Errorf("stored environment variable = %v, want %q", cfg["textValue"], resolved)
		}
		return nil
	}

	// State keeps the configured form, so plan, apply and refresh agree.
	wantData, _ := json.Marshal(map[string]string{"region": "eu-central-1", "vpc_id": reference})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProviderFactories(http.Header{}),
		Steps: []resource.TestStep{
			{
				Config: inputDataConfig(srv.URL, data, hclReference, "first"),
				Check: resource.ComposeAggregateTestCheckFunc(
					checkStoredResolved,
					resource.TestCheckResourceAttr("stackguardian_stack.test", "workflows_config.workflows.1.vcs_config.iac_input_data.data", string(wantData)),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "workflows_config.workflows.1.environment_variables.0.config.text_value", reference),
				),
			},
			{
				// An update — where the platform does not resolve references itself.
				Config: inputDataConfig(srv.URL, data, hclReference, "second"),
				Check: resource.ComposeAggregateTestCheckFunc(
					checkStoredResolved,
					resource.TestCheckResourceAttr("stackguardian_stack.test", "workflows_config.workflows.1.description", "second"),
				),
			},
			{
				Config:   inputDataConfig(srv.URL, data, hclReference, "second"),
				PlanOnly: true,
			},
		},
	})
}

// JSON written by hand rather than with jsonencode keeps its own formatting in state.
func TestStack_InputDataFormattingKept(t *testing.T) {
	srv, _ := newFakeStackAPI(t, twoWorkflowRevision)
	data := `"{ \"region\": \"eu-central-1\",  \"count\": 2 }"`

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProviderFactories(http.Header{}),
		Steps: []resource.TestStep{
			{
				Config: inputDataConfig(srv.URL, data, "plain", "first"),
				Check: resource.TestCheckResourceAttr("stackguardian_stack.test",
					"workflows_config.workflows.1.vcs_config.iac_input_data.data", `{ "region": "eu-central-1",  "count": 2 }`),
			},
		},
	})
}

func TestStack_InputDataInvalidJSONRejectedAtPlan(t *testing.T) {
	srv, _ := newFakeStackAPI(t, twoWorkflowRevision)

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProviderFactories(http.Header{}),
		Steps: []resource.TestStep{
			{
				Config:      inputDataConfig(srv.URL, `"{ region = \"eu\" }"`, "plain", "first"),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`Invalid JSON object`),
			},
		},
	})
}
