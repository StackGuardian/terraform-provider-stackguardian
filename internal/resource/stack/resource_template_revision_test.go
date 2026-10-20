package stack_test

import (
	"fmt"
	"net/http"
	"slices"
	"testing"

	"github.com/StackGuardian/terraform-provider-stackguardian/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// These tests run real `terraform` steps against the in-memory stack API fake in
// resource_override_removal_test.go (no credentials, no TF_ACC).

func singleWorkflowRevision(iacTemplateId string) string {
	return fmt.Sprintf(`{"WorkflowsConfig": {"workflows": [{"id": "wf-a", "ResourceName": "wf-a-0",
	  "VCSConfig": {"iacVCSConfig": {"useMarketplaceTemplate": true, "iacTemplateId": %q}}}]}}`, iacTemplateId)
}

func templateReadsCheck(api *fakeStackAPI, want []string) resource.TestCheckFunc {
	return func(*terraform.State) error {
		got := api.workflowTemplateReads()
		for _, id := range want {
			if !slices.Contains(got, id) {
				return fmt.Errorf("workflow template reads = %v, want %q among them", got, id)
			}
		}
		if len(want) == 0 && len(got) > 0 {
			return fmt.Errorf("workflow template reads = %v, want none", got)
		}
		return nil
	}
}

// A workflow template without a revision is not read as if it were one: there is no revision
// content to inherit, and the apply still succeeds.
func TestStack_UnpinnedWorkflowTemplateNotReadAsRevision(t *testing.T) {
	srv, api := newFakeStackAPI(t, singleWorkflowRevision("/override-removal-org/net-tpl"))

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProviderFactories(http.Header{}),
		Steps: []resource.TestStep{{
			Config: actionsConfig(srv.URL, "tpl:1", "", "first", "wf-a"),
			Check:  templateReadsCheck(api, nil),
		}},
	})
}

func TestStack_PinnedWorkflowTemplateRead(t *testing.T) {
	srv, api := newFakeStackAPI(t, singleWorkflowRevision("/override-removal-org/net-tpl:3"))

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProviderFactories(http.Header{}),
		Steps: []resource.TestStep{{
			Config: actionsConfig(srv.URL, "tpl:1", "", "first", "wf-a"),
			Check:  templateReadsCheck(api, []string{"override-removal-org/net-tpl:3"}),
		}},
	})
}

// A create response without the stack in it must not crash the provider: the stack exists
// under the configured id, so it is read back by that id.
func TestStack_CreateSurvivesResponseWithoutStack(t *testing.T) {
	srv, api := newFakeStackAPI(t, `{"WorkflowsConfig": {"workflows": [{"id": "wf-a", "ResourceName": "wf-a-0"}]}}`)
	api.emptyCreateResponse = true

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProviderFactories(http.Header{}),
		Steps: []resource.TestStep{{
			Config: actionsConfig(srv.URL, "tpl:1", "", "first", "wf-a"),
			Check:  resource.TestCheckResourceAttr("stackguardian_stack.test", "id", "override-removal"),
		}},
	})
}
