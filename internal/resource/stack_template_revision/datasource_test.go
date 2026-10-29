package stacktemplaterevision_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/StackGuardian/terraform-provider-stackguardian/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

// TestAccStackTemplateRevision_DataSourceParentTemplateId checks the
// stackguardian_stack_template_revision data source reads parent_template_id back:
// the revision's id with the ":<revision>" suffix removed, the same value the
// resource was created with. It lives in the resource package, not the data source
// one, to reuse this package's fixtures.
func TestAccStackTemplateRevision_DataSourceParentTemplateId(t *testing.T) {
	acctest.SkipUnlessAcceptance(t)
	stackTemplateID := acctest.ResourceName("provider-test-stack-template-revds")
	wfTemplateID := acctest.ResourceName("provider-test-wft-for-stack-revds")

	defer deleteWorkflowTemplateFixture(wfTemplateID)
	defer deleteStackTemplateFixture(stackTemplateID)
	defer deleteStackTemplateRevisionFixture(fmt.Sprintf("%s:1", stackTemplateID))
	defer deprecateStackTemplateRevisionFixture(fmt.Sprintf("%s:1", stackTemplateID))

	if err := createWorkflowTemplateFixture(wfTemplateID, "TERRAFORM"); err != nil {
		t.Fatalf("failed to create workflow template fixture: %s", err)
	}
	if err := createStackTemplateFixture(stackTemplateID, "TERRAFORM"); err != nil {
		t.Fatalf("failed to create stack template fixture: %s", err)
	}

	customHeader := http.Header{}
	customHeader.Set("x-sg-internal-auth-orgid", "sg-provider-test")

	hcl := testAccStackTemplateRevisionConfig(stackTemplateID, wfTemplateID, "v1", "notes", "description") + `
data "stackguardian_stack_template_revision" "test" {
  id = stackguardian_stack_template_revision.test.id
}
`

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader),
		Steps: []resource.TestStep{
			{
				Config: hcl,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.stackguardian_stack_template_revision.test", "id", stackTemplateID+":1"),
					resource.TestCheckResourceAttr("data.stackguardian_stack_template_revision.test", "parent_template_id", stackTemplateID),
					resource.TestCheckResourceAttrPair(
						"data.stackguardian_stack_template_revision.test", "parent_template_id",
						"stackguardian_stack_template_revision.test", "parent_template_id",
					),
				),
			},
		},
	})
}
