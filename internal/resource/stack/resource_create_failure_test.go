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
// resource_override_removal_test.go (no credentials, no TF_ACC). They pin down that a stack
// the platform stored is never left out of the Terraform state, and that a stack this
// configuration did not create is never taken over.

const createFailureRevision = `{"WorkflowsConfig": {"workflows": [{"id": "wf-a", "ResourceName": "wf-a-0"}]}}`

func expectCounts(api *fakeStackAPI, wantCreatesAtLeast, wantDeletes int) resource.TestCheckFunc {
	return func(*terraform.State) error {
		creates, deletes := api.counts()
		if creates < wantCreatesAtLeast || deletes != wantDeletes {
			return fmt.Errorf("creates = %d, deletes = %d; want creates >= %d, deletes = %d", creates, deletes, wantCreatesAtLeast, wantDeletes)
		}
		return nil
	}
}

// The create succeeds but reading the stack back fails: the stack is kept in state as tainted,
// and the next apply replaces it instead of failing with "already exists".
func TestStack_CreatedStackTrackedWhenReadBackFails(t *testing.T) {
	srv, api := newFakeStackAPI(t, createFailureRevision)
	api.set(func(f *fakeStackAPI) { f.failReadsAfterCreate = true })
	config := actionsConfig(srv.URL, "tpl:1", "", "first", "wf-a")

	runFakeAPITest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProviderFactories(http.Header{}),
		Steps: []resource.TestStep{
			{
				Config:      config,
				ExpectError: regexp.MustCompile(`Error reading created stack`),
			},
			{
				PreConfig: func() { api.set(func(f *fakeStackAPI) { f.failReadsAfterCreate = false }) },
				Config:    config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_stack.test", "id", "override-removal"),
					// Replaced: the tainted stack was deleted before being created again.
					expectCounts(api, 2, 1),
				),
			},
		},
	})
}

// The platform stores the stack, then a workflow fails and the create returns an error: the
// partial stack is tracked as tainted, so the next apply replaces it.
func TestStack_PartialStackTrackedWhenCreateFails(t *testing.T) {
	srv, api := newFakeStackAPI(t, createFailureRevision)
	api.set(func(f *fakeStackAPI) { f.failCreate = true })
	config := actionsConfig(srv.URL, "tpl:1", "", "first", "wf-a")

	runFakeAPITest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProviderFactories(http.Header{}),
		Steps: []resource.TestStep{
			{
				Config:      config,
				ExpectError: regexp.MustCompile(`(?s)Failed to create workflow.*already stored stack`),
			},
			{
				PreConfig: func() { api.set(func(f *fakeStackAPI) { f.failCreate = false }) },
				Config:    config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_stack.test", "id", "override-removal"),
					expectCounts(api, 2, 1),
				),
			},
		},
	})
}

// A stack that already exists under the configured id is reported, never created over, and —
// even when the create would fail for another reason — never put in state, where a later
// destroy would delete it.
func TestStack_ExistingStackNotTakenOver(t *testing.T) {
	for _, tc := range []struct {
		name       string
		failCreate bool
	}{
		{"create would succeed", false},
		{"create would fail", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv, api := newFakeStackAPI(t, createFailureRevision)
			api.preload("override-removal", map[string]any{"Id": "override-removal", "ResourceName": "someone else's"})
			api.set(func(f *fakeStackAPI) { f.failCreate = tc.failCreate })

			runFakeAPITest(t, resource.TestCase{
				ProtoV6ProviderFactories: acctest.ProviderFactories(http.Header{}),
				Steps: []resource.TestStep{{
					Config:      actionsConfig(srv.URL, "tpl:1", "", "first", "wf-a"),
					ExpectError: regexp.MustCompile(`(?s)Stack already exists.*terraform import`),
				}},
				CheckDestroy: func(*terraform.State) error {
					creates, deletes := api.counts()
					if creates != 0 || deletes != 0 {
						return fmt.Errorf("creates = %d, deletes = %d; want neither", creates, deletes)
					}
					if api.storedStack("override-removal") == nil {
						return fmt.Errorf("the existing stack is gone")
					}
					return nil
				},
			})
		})
	}
}

// A delete the platform refuses because a run is active explains why, and leaves the stack in
// state so a later destroy can finish the job.
func TestStack_DeleteRefusedWhileRunActiveExplainsWhy(t *testing.T) {
	srv, api := newFakeStackAPI(t, createFailureRevision)
	config := actionsConfig(srv.URL, "tpl:1", "", "first", "wf-a")

	runFakeAPITest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProviderFactories(http.Header{}),
		Steps: []resource.TestStep{
			{Config: config},
			{
				PreConfig:   func() { api.set(func(f *fakeStackAPI) { f.failDelete = true }) },
				Config:      config,
				Destroy:     true,
				ExpectError: regexp.MustCompile(`run\s+in\s+progress\s+or\s+queued`),
			},
			{
				// Still tracked: once the run is over, the closing destroy succeeds.
				PreConfig: func() { api.set(func(f *fakeStackAPI) { f.failDelete = false }) },
				Config:    config,
				Check:     resource.TestCheckResourceAttr("stackguardian_stack.test", "id", "override-removal"),
			},
		},
	})
}
