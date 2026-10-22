package stack_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/StackGuardian/terraform-provider-stackguardian/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// These tests run real `terraform` steps against the in-memory stack API fake in
// resource_override_removal_test.go (no credentials, no TF_ACC).

var unknownOnUpgradeRevisions = map[string]string{
	"tpl:1": `{"WorkflowsConfig": {"workflows": [{"id": "wf-a", "ResourceName": "wf-a-0"}]}}`,
	"tpl:2": `{"WorkflowsConfig": {"workflows": [{"id": "wf-a", "ResourceName": "wf-a-0", "Description": "from the template"}]}}`,
}

// unknownOnUpgradeConfig declares the workflow's attribute from terraform_data.<source>.id, which
// is unknown at plan time whenever that terraform_data is new.
func unknownOnUpgradeConfig(apiURI, revision, source, workflowAttr string) string {
	return fmt.Sprintf(`
provider "stackguardian" {
  api_key  = "test"
  org_name = %q
  api_uri  = %q
}

resource "terraform_data" %q {
  input = %q
}

resource "stackguardian_stack" "test" {
  id                = "override-removal"
  workflow_group_id = "override-removal-wfgrp"
  template_group_id = %q
  workflows_config = {
    workflows = [{
      id = "wf-a"
      %s
    }]
  }
}
`, overrideRemovalOrg, apiURI, source, source, revision, fmt.Sprintf(workflowAttr, "terraform_data."+source+".id"))
}

// Changing template_group_id in the same apply that creates the resource a workflow attribute
// comes from: the attribute is unknown at plan time and must stay unknown in the plan, and the
// update then sends the configured value.
func TestStack_UnknownWorkflowValueOnRevisionChange(t *testing.T) {
	for name, attr := range map[string]string{
		"description":           `description = "Deploys ${%s}"`,
		"iac_input_data.data":   `vcs_config = { iac_input_data = { schema_type = "RAW_JSON", data = jsonencode({ id = %s }) } }`,
		"environment_variables": `environment_variables = [{ kind = "PLAIN_TEXT", config = { var_name = "ID", text_value = %s } }]`,
	} {
		t.Run(name, func(t *testing.T) {
			srv, api := newFakeStackAPIWithRevisions(t, unknownOnUpgradeRevisions)

			// The update must carry terraform_data.second's id, which only exists at apply.
			sentNewId := func(s *terraform.State) error {
				id := s.RootModule().Resources["terraform_data.second"].Primary.ID
				patch, _ := json.Marshal(api.lastPatch()["WorkflowsConfig"])
				if !strings.Contains(string(patch), id) {
					return fmt.Errorf("the update did not send terraform_data.second's id %q: %s", id, patch)
				}
				return nil
			}

			runFakeAPITest(t, resource.TestCase{
				ProtoV6ProviderFactories: acctest.ProviderFactories(http.Header{}),
				Steps: []resource.TestStep{
					{Config: unknownOnUpgradeConfig(srv.URL, "tpl:1", "first", attr)},
					{Config: unknownOnUpgradeConfig(srv.URL, "tpl:2", "second", attr), Check: sentNewId},
				},
			})
		})
	}
}
