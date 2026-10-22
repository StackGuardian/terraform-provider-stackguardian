package stack_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/StackGuardian/terraform-provider-stackguardian/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// These tests run real `terraform` steps against the in-memory stack API fake in
// resource_override_removal_test.go (no credentials, no TF_ACC).

const templateHooksTerraformConfig = `"TerraformConfig": {
  "preInitHooks": ["echo tpl"],
  "terraformBinPath": [{"source": "/opt/tf", "target": "/bin/terraform"}],
  "postApplyWfStepsConfig": [{"name": "notify", "wfStepTemplateId": "/demo/notify:1"}]
}`

var emptyListRevisions = map[string]string{
	"tpl:1": `{"WorkflowsConfig": {"workflows": [{"id": "wf-a", "ResourceName": "wf-a-0"}]}}`,
	"tpl:2": `{"WorkflowsConfig": {"workflows": [{"id": "wf-a", "ResourceName": "wf-a-0", ` + templateHooksTerraformConfig + `}]}}`,
}

func emptyListConfig(apiURI, revision, terraformConfig string) string {
	return fmt.Sprintf(`
provider "stackguardian" {
  api_key  = "test"
  org_name = %q
  api_uri  = %q
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
`, overrideRemovalOrg, apiURI, revision, terraformConfig)
}

// sentTerraformConfig returns TerraformConfig of the first workflow in body, as sent.
func sentTerraformConfig(body map[string]any) map[string]any {
	wfc, _ := body["WorkflowsConfig"].(map[string]any)
	workflows, _ := wfc["workflows"].([]any)
	if len(workflows) == 0 {
		return nil
	}
	wf, _ := workflows[0].(map[string]any)
	tc, _ := wf["TerraformConfig"].(map[string]any)
	return tc
}

// A declared [] clears hooks, steps and bin paths the template sets: it is sent as [] and reads
// back as [].
func TestStack_EmptyListOverridesTemplate(t *testing.T) {
	srv, api := newFakeStackAPIWithRevisions(t, emptyListRevisions)
	const cleared = `terraform_config = {
        pre_init_hooks              = []
        terraform_bin_path          = []
        post_apply_wf_steps_config  = []
      }`

	sentEmpty := func(*terraform.State) error {
		tc := sentTerraformConfig(api.storedStack("override-removal"))
		for _, key := range []string{"preInitHooks", "terraformBinPath", "postApplyWfStepsConfig"} {
			if list, ok := tc[key].([]any); !ok || len(list) != 0 {
				sent, _ := json.Marshal(tc[key])
				return fmt.Errorf("%s sent as %s, want []", key, sent)
			}
		}
		return nil
	}

	runFakeAPITest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProviderFactories(http.Header{}),
		Steps: []resource.TestStep{{
			Config: emptyListConfig(srv.URL, "tpl:2", cleared),
			Check: resource.ComposeAggregateTestCheckFunc(
				sentEmpty,
				resource.TestCheckResourceAttr("stackguardian_stack.test", "workflows_config.workflows.0.terraform_config.pre_init_hooks.#", "0"),
				resource.TestCheckResourceAttr("stackguardian_stack.test", "workflows_config.workflows.0.terraform_config.terraform_bin_path.#", "0"),
				resource.TestCheckResourceAttr("stackguardian_stack.test", "workflows_config.workflows.0.terraform_config.post_apply_wf_steps_config.#", "0"),
			),
		}},
	})
}

// A list left unset still adopts the new revision's value on upgrade, though state holds it as
// [] from the old revision.
func TestStack_UnsetListAdoptsTemplateOnUpgrade(t *testing.T) {
	for name, terraformConfig := range map[string]string{
		"no terraform_config":       ``,
		"terraform_config, no list": `terraform_config = { timeout = 30 }`,
	} {
		t.Run(name, func(t *testing.T) {
			srv, _ := newFakeStackAPIWithRevisions(t, emptyListRevisions)

			runFakeAPITest(t, resource.TestCase{
				ProtoV6ProviderFactories: acctest.ProviderFactories(http.Header{}),
				Steps: []resource.TestStep{
					{Config: emptyListConfig(srv.URL, "tpl:1", terraformConfig)},
					{
						Config: emptyListConfig(srv.URL, "tpl:2", terraformConfig),
						Check: resource.ComposeAggregateTestCheckFunc(
							resource.TestCheckResourceAttr("stackguardian_stack.test", "workflows_config.workflows.0.terraform_config.pre_init_hooks.0", "echo tpl"),
							resource.TestCheckResourceAttr("stackguardian_stack.test", "workflows_config.workflows.0.terraform_config.terraform_bin_path.0.source", "/opt/tf"),
							resource.TestCheckResourceAttr("stackguardian_stack.test", "workflows_config.workflows.0.terraform_config.post_apply_wf_steps_config.0.name", "notify"),
						),
					},
				},
			})
		})
	}
}
