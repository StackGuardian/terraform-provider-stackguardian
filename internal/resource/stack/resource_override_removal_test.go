package stack_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/StackGuardian/terraform-provider-stackguardian/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

// These tests run real `terraform apply` / `plan` steps against an in-memory fake of the stack
// API (no credentials, no TF_ACC), to pin down that removing an override from the config
// reverts the attribute to the value the stack template revision gives it, instead of
// silently keeping the old value.

const overrideRemovalOrg = "override-removal-org"

// fakeStackAPI stores each stack as the JSON the provider sent and returns it on read: the
// create request (Stack) and the read response (StackData) share their JSON keys, and an
// update merges the top-level keys it carries, like a PATCH.
type fakeStackAPI struct {
	mu            sync.Mutex
	revisions     map[string]string // stack template revision id -> its read "msg"
	stacks        map[string]map[string]any
	patches       []map[string]any // every update body, in order
	templateReads []string         // workflow template revision ids read, in order
}

// newFakeStackAPI serves one stack template revision, "tpl:1".
func newFakeStackAPI(t *testing.T, revision string) (*httptest.Server, *fakeStackAPI) {
	t.Helper()
	return newFakeStackAPIWithRevisions(t, map[string]string{"tpl:1": revision})
}

func newFakeStackAPIWithRevisions(t *testing.T, revisions map[string]string) (*httptest.Server, *fakeStackAPI) {
	t.Helper()
	api := &fakeStackAPI{revisions: revisions, stacks: map[string]map[string]any{}}
	srv := httptest.NewServer(api)
	t.Cleanup(srv.Close)
	return srv, api
}

// storedStack returns the stack as the fake last stored it, round-tripped through JSON.
func (f *fakeStackAPI) storedStack(id string) map[string]any {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.stacks[id]
}

func (f *fakeStackAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	w.Header().Set("Content-Type", "application/json")

	revisionPrefix := fmt.Sprintf("/api/v1/templatetypes/IAC_GROUP/%s/", overrideRemovalOrg)
	stacksPrefix := fmt.Sprintf("/api/v1/orgs/%s/wfgrps/override-removal-wfgrp/stacks/", overrideRemovalOrg)

	switch {
	case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, revisionPrefix):
		revision, ok := f.revisions[strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, revisionPrefix), "/")]
		if !ok {
			http.Error(w, `{"msg":"not found"}`, http.StatusNotFound)
			return
		}
		fmt.Fprintf(w, `{"msg":%s}`, revision)
		return
	case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/api/v1/templatetypes/IAC/"):
		f.templateReads = append(f.templateReads, strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/api/v1/templatetypes/IAC/"), "/"))
		// A workflow template revision with no defaults of its own.
		fmt.Fprint(w, `{"msg":{}}`)
		return
	case !strings.HasPrefix(r.URL.Path, stacksPrefix):
		http.Error(w, `{"msg":"not found"}`, http.StatusNotFound)
		return
	}

	id := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, stacksPrefix), "/")
	switch r.Method {
	case http.MethodPost:
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, `{"msg":"bad request"}`, http.StatusBadRequest)
			return
		}
		stackId, _ := body["Id"].(string)
		f.stacks[stackId] = body
		stored, _ := json.Marshal(body)
		fmt.Fprintf(w, `{"msg":"created","data":{"stack":%s}}`, stored)
	case http.MethodGet:
		stack, ok := f.stacks[id]
		if !ok {
			http.Error(w, `{"msg":"not found"}`, http.StatusNotFound)
			return
		}
		stored, _ := json.Marshal(stack)
		fmt.Fprintf(w, `{"msg":%s}`, stored)
	case http.MethodPatch:
		stack, ok := f.stacks[id]
		if !ok {
			http.Error(w, `{"msg":"not found"}`, http.StatusNotFound)
			return
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, `{"msg":"bad request"}`, http.StatusBadRequest)
			return
		}
		f.patches = append(f.patches, body)
		for k, v := range body {
			stack[k] = v
		}
		fmt.Fprint(w, `{"msg":"updated"}`)
	case http.MethodDelete:
		delete(f.stacks, id)
		fmt.Fprint(w, `{"msg":"deleted"}`)
	default:
		http.Error(w, `{"msg":"method not allowed"}`, http.StatusMethodNotAllowed)
	}
}

func overrideRemovalConfig(apiURI, stackAttrs, workflowAttrs string) string {
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
  %s
  workflows_config = {
    workflows = [
      {
        id = "wf-a"
        %s
      }
    ]
  }
}
`, overrideRemovalOrg, apiURI, stackAttrs, workflowAttrs)
}

func TestStack_RemovedOverrideRevertsToTemplateValue(t *testing.T) {
	srv, _ := newFakeStackAPI(t, `{
	  "LongDescription": "from template",
	  "WorkflowsConfig": {"workflows": [
	    {"id": "wf-a", "ResourceName": "wf-a-0", "NumberOfApprovalsRequired": 1,
	     "TerraformConfig": {"driftCheck": true, "driftCron": "0 0 * * ? *"}}
	  ]}
	}`)

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProviderFactories(http.Header{}),
		Steps: []resource.TestStep{
			{
				// Override a stack-level field, a workflow field and one nested field.
				Config: overrideRemovalConfig(srv.URL,
					`description = "custom"`,
					`number_of_approvals_required = 2
        terraform_config = { drift_check = false, managed_terraform_state = true }`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_stack.test", "description", "custom"),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "workflows_config.workflows.0.number_of_approvals_required", "2"),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "workflows_config.workflows.0.terraform_config.drift_check", "false"),
				),
			},
			{
				// Drop all three overrides; terraform_config itself stays declared.
				Config: overrideRemovalConfig(srv.URL, ``,
					`terraform_config = { managed_terraform_state = true }`),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("stackguardian_stack.test", plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_stack.test", "description", "from template"),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "workflows_config.workflows.0.number_of_approvals_required", "1"),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "workflows_config.workflows.0.terraform_config.drift_check", "true"),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "workflows_config.workflows.0.terraform_config.managed_terraform_state", "true"),
				),
			},
			{
				// The reverted values are now in state: no further diff.
				Config:   overrideRemovalConfig(srv.URL, ``, `terraform_config = { managed_terraform_state = true }`),
				PlanOnly: true,
			},
		},
	})
}

// A field the template does not set goes back to its default when its override is removed.
func TestStack_RemovedOverrideWithoutTemplateValueResets(t *testing.T) {
	srv, _ := newFakeStackAPI(t, `{"WorkflowsConfig": {"workflows": [{"id": "wf-a", "ResourceName": "wf-a-0"}]}}`)

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProviderFactories(http.Header{}),
		Steps: []resource.TestStep{
			{
				Config: overrideRemovalConfig(srv.URL, ``, `approvers = ["alice@example.com"]`),
				Check:  resource.TestCheckResourceAttr("stackguardian_stack.test", "workflows_config.workflows.0.approvers.#", "1"),
			},
			{
				Config: overrideRemovalConfig(srv.URL, ``, ``),
				Check:  resource.TestCheckResourceAttr("stackguardian_stack.test", "workflows_config.workflows.0.approvers.#", "0"),
			},
		},
	})
}

// lastPatch returns the body of the most recent update.
func (f *fakeStackAPI) lastPatch() map[string]any {
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.patches) == 0 {
		return nil
	}
	return f.patches[len(f.patches)-1]
}

// workflowTemplateReads returns the workflow template revision ids read so far.
func (f *fakeStackAPI) workflowTemplateReads() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]string(nil), f.templateReads...)
}
