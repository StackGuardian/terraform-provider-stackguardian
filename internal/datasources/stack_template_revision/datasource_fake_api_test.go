package stacktemplaterevisiondatasource_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/StackGuardian/terraform-provider-stackguardian/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// Reads a revision through a real `terraform` run against a fake API (no credentials, no
// TF_ACC): the API response does not name the parent template, so parent_template_id must
// come from the id rather than being left null.
func TestStackTemplateRevisionDataSource_SetsParentTemplateId(t *testing.T) {
	acctest.SkipUnlessTerraformCLI(t)
	const org = "ds-parent-org"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != fmt.Sprintf("/api/v1/templatetypes/IAC_GROUP/%s/my-stack-template:1/", org) {
			http.Error(w, `{"msg":"not found"}`, http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"msg":{"Id":"my-stack-template:1","TemplateId":"/%s/my-stack-template:1"}}`, org)
	}))
	t.Cleanup(srv.Close)

	resource.ParallelTest(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: acctest.ProviderFactories(http.Header{}),
		Steps: []resource.TestStep{{
			Config: fmt.Sprintf(`
provider "stackguardian" {
  api_key  = "test"
  org_name = %q
  api_uri  = %q
}

data "stackguardian_stack_template_revision" "test" {
  id = "my-stack-template:1"
}
`, org, srv.URL),
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("data.stackguardian_stack_template_revision.test", "parent_template_id", "my-stack-template"),
				resource.TestCheckResourceAttr("data.stackguardian_stack_template_revision.test", "template_id", "/"+org+"/my-stack-template:1"),
			),
		}},
	})
}
