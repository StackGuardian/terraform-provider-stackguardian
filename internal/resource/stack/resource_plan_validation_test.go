package stack_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/StackGuardian/terraform-provider-stackguardian/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// These tests run a real `terraform plan` against a fake API (no credentials, no TF_ACC)
// to pin down that workflows_config is checked against the stack template revision at
// plan time — ModifyPlan — not only when Create runs at apply.

const planValidationOrg = "plan-validation-org"

// fakeStackTemplateRevisionAPI serves GET .../templatetypes/IAC_GROUP/<org>/<revision>/ for the
// revisions given (revision id -> workflow ids, in order) and 404s everything else. It counts
// the revision reads, so a test can tell the check ran during plan.
func fakeStackTemplateRevisionAPI(t *testing.T, revisions map[string][]string) (*httptest.Server, *atomic.Int32) {
	t.Helper()
	var reads atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		prefix := fmt.Sprintf("/api/v1/templatetypes/IAC_GROUP/%s/", planValidationOrg)
		revision := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, prefix), "/")
		ids, ok := revisions[revision]
		if r.Method != http.MethodGet || !strings.HasPrefix(r.URL.Path, prefix) || !ok {
			http.Error(w, `{"msg":"not found"}`, http.StatusNotFound)
			return
		}
		reads.Add(1)
		workflows := make([]string, 0, len(ids))
		for _, id := range ids {
			workflows = append(workflows, fmt.Sprintf(`{"id":%q}`, id))
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"msg":{"WorkflowsConfig":{"workflows":[%s]}}}`, strings.Join(workflows, ","))
	}))
	t.Cleanup(srv.Close)
	return srv, &reads
}

func planValidationConfig(apiURI, revision string, workflowIds ...string) string {
	entries := make([]string, 0, len(workflowIds))
	for _, id := range workflowIds {
		entries = append(entries, fmt.Sprintf("{ id = %q }", id))
	}
	return fmt.Sprintf(`
provider "stackguardian" {
  api_key  = "test"
  org_name = %q
  api_uri  = %q
}

resource "stackguardian_stack" "test" {
  id                = "plan-validation"
  workflow_group_id = "plan-validation-wfgrp"
  template_group_id = %q
  workflows_config = {
    workflows = [%s]
  }
}
`, planValidationOrg, apiURI, revision, strings.Join(entries, ", "))
}

func TestStack_PlanRejectsWorkflowsConfigMismatch(t *testing.T) {
	srv, _ := fakeStackTemplateRevisionAPI(t, map[string][]string{"tpl:1": {"wf-a", "wf-b"}})

	for _, tc := range []struct {
		name string
		ids  []string
	}{
		{"wrong order", []string{"wf-b", "wf-a"}},
		{"missing workflow", []string{"wf-a"}},
		{"unknown workflow", []string{"wf-a", "wf-b", "wf-c"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resource.UnitTest(t, resource.TestCase{
				ProtoV6ProviderFactories: acctest.ProviderFactories(http.Header{}),
				Steps: []resource.TestStep{{
					Config:      planValidationConfig(srv.URL, "tpl:1", tc.ids...),
					PlanOnly:    true,
					ExpectError: regexp.MustCompile(`workflows_config.workflows does not match the stack\s+template revision`),
				}},
			})
		})
	}
}

func TestStack_PlanAcceptsMatchingWorkflowsConfig(t *testing.T) {
	srv, reads := fakeStackTemplateRevisionAPI(t, map[string][]string{"tpl:1": {"wf-a", "wf-b"}})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProviderFactories(http.Header{}),
		Steps: []resource.TestStep{{
			Config:             planValidationConfig(srv.URL, "tpl:1", "wf-a", "wf-b"),
			PlanOnly:           true,
			ExpectNonEmptyPlan: true,
		}},
	})
	if reads.Load() == 0 {
		t.Fatal("the stack template revision was never read during plan")
	}
}

// A revision that does not exist yet at plan time (created by the same apply) must not fail
// the plan; Create checks it at apply.
func TestStack_PlanSkipsCheckWhenRevisionNotFoundOnCreate(t *testing.T) {
	srv, _ := fakeStackTemplateRevisionAPI(t, map[string][]string{})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProviderFactories(http.Header{}),
		Steps: []resource.TestStep{{
			Config:             planValidationConfig(srv.URL, "tpl:1", "wf-a"),
			PlanOnly:           true,
			ExpectNonEmptyPlan: true,
		}},
	})
}
