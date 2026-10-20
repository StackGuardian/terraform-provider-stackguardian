package stack_test

import (
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

func actionsConfig(apiURI, revision, actions, workflowDescription string, workflowIds ...string) string {
	entries := ""
	for i, id := range workflowIds {
		if i == 0 {
			entries += fmt.Sprintf("{ id = %q, description = %q },\n", id, workflowDescription)
			continue
		}
		entries += fmt.Sprintf("{ id = %q },\n", id)
	}
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
  %s
  workflows_config = {
    workflows = [
      %s
    ]
  }
}
`, overrideRemovalOrg, apiURI, revision, actions, entries)
}

// Actions the configuration does not declare come from the template with their default flag,
// and an unrelated update leaves them to the platform instead of re-sending them as custom.
func TestStack_UndeclaredActionsKeepDefaultAndAreNotResent(t *testing.T) {
	srv, api := newFakeStackAPI(t, `{
	  "Actions": {"apply": {"name": "apply", "default": true, "order": {"wf-a": {}}}},
	  "WorkflowsConfig": {"workflows": [{"id": "wf-a", "ResourceName": "wf-a-0"}]}
	}`)

	noActionsInPatch := func(*terraform.State) error {
		patch := api.lastPatch()
		if patch == nil {
			return fmt.Errorf("no update was sent")
		}
		if _, ok := patch["Actions"]; ok {
			return fmt.Errorf("the update re-sent Actions: %v", patch["Actions"])
		}
		return nil
	}

	runFakeAPITest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProviderFactories(http.Header{}),
		Steps: []resource.TestStep{
			{
				Config: actionsConfig(srv.URL, "tpl:1", "", "first", "wf-a"),
				Check:  resource.TestCheckResourceAttr("stackguardian_stack.test", "actions.apply.default", "true"),
			},
			{
				Config: actionsConfig(srv.URL, "tpl:1", "", "second", "wf-a"),
				Check: resource.ComposeAggregateTestCheckFunc(
					noActionsInPatch,
					resource.TestCheckResourceAttr("stackguardian_stack.test", "actions.apply.default", "true"),
				),
			},
		},
	})
}

// Declared actions are sent on every update, with the default flag the platform holds when the
// configuration leaves it unset.
func TestStack_DeclaredActionsSentWithKnownDefault(t *testing.T) {
	srv, api := newFakeStackAPI(t, `{"WorkflowsConfig": {"workflows": [{"id": "wf-a", "ResourceName": "wf-a-0"}]}}`)
	declared := `actions = { deploy = { name = "deploy", order = { "wf-a" = {} } } }`

	sentDefault := func(*terraform.State) error {
		patch := api.lastPatch()
		actions, _ := patch["Actions"].(map[string]any)
		deploy, _ := actions["deploy"].(map[string]any)
		if deploy == nil {
			return fmt.Errorf("the update did not send the declared actions: %v", patch)
		}
		if deploy["default"] != false {
			return fmt.Errorf("sent default = %v, want false (the value state holds)", deploy["default"])
		}
		return nil
	}

	runFakeAPITest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProviderFactories(http.Header{}),
		Steps: []resource.TestStep{
			{
				// Sent without default: the platform reads an absent flag as custom.
				Config: actionsConfig(srv.URL, "tpl:1", declared, "first", "wf-a"),
				Check:  resource.TestCheckResourceAttr("stackguardian_stack.test", "actions.deploy.default", "false"),
			},
			{
				Config: actionsConfig(srv.URL, "tpl:1", declared, "second", "wf-a"),
				Check:  sentDefault,
			},
			{
				// A declared flag is sent as written, not overwritten.
				Config: actionsConfig(srv.URL, "tpl:1",
					`actions = { deploy = { name = "deploy", default = true, order = { "wf-a" = {} } } }`, "second", "wf-a"),
				Check: resource.TestCheckResourceAttr("stackguardian_stack.test", "actions.deploy.default", "true"),
			},
		},
	})
}

// Undeclared actions that reference a workflow the new revision drops, with no actions on the
// new revision to replace them, fail the plan instead of the apply.
func TestStack_InheritedActionsReferencingRemovedWorkflowRejectedAtPlan(t *testing.T) {
	srv, _ := newFakeStackAPIWithRevisions(t, map[string]string{
		"tpl:1": `{
		  "Actions": {"apply": {"name": "apply", "default": true, "order": {"wf-a": {}, "wf-b": {}}}},
		  "WorkflowsConfig": {"workflows": [{"id": "wf-a", "ResourceName": "a"}, {"id": "wf-b", "ResourceName": "b"}]}
		}`,
		"tpl:2": `{"WorkflowsConfig": {"workflows": [{"id": "wf-a", "ResourceName": "a"}]}}`,
	})

	runFakeAPITest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProviderFactories(http.Header{}),
		Steps: []resource.TestStep{
			{
				Config: actionsConfig(srv.URL, "tpl:1", "", "first", "wf-a", "wf-b"),
			},
			{
				Config:      actionsConfig(srv.URL, "tpl:2", "", "first", "wf-a"),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`(?s)references workflow "wf-b".*not in your configuration`),
			},
		},
	})
}
