package connector_test

import (
	"testing"

	"github.com/StackGuardian/terraform-provider-stackguardian/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

const (
	testAccResource = `resource "stackguardian_connector" "aws-cloud-connector-example" {
  resource_name = "aws-rbac-connector"
  description   = "AWS Cloud Connector"

  settings = {
    kind = "AWS_RBAC"

    config = [{
      role_arn         = "arn:aws:iam::209502960327:role/StackGuardian"
      external_id      = "sg-provider-test:ElfygiFglfldTwnDFpAScQkvgvHTGV "
      duration_seconds = "3600"
    }]
  }
}`

	testAccResourceUpdate = `resource "stackguardian_connector" "aws-cloud-connector-example" {
  resource_name = "aws-rbac-connector"
  description   = "AWS Cloud Connector Update"

  settings = {
    kind = "AWS_RBAC"

    config = [{
      role_arn         = "arn:aws:iam::209502960327:role/StackGuardian"
      external_id      = "sg-provider-test:ElfygiFglfldTwnDFpAScQkvgvHTGV "
      duration_seconds = "3600"
    }]
  }
}`
)

func TestAccWorkflowGroup(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccResource,
			},
			{
				Config: testAccResourceUpdate,
			},
		},
	})
}
