package connector_test

import (
	"testing"

	"github.com/StackGuardian/terraform-provider-stackguardian/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
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
      external_id      = "sg-provider-test:ElfygiFglfldTwnDFpAScQkvgvHTGV"
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
      external_id      = "sg-provider-test:ElfygiFglfldTwnDFpAScQkvgvHTGV"
      duration_seconds = "3600"
    }]
  }
}`
)

func TestAccConnector(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccResource,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"stackguardian_connector.aws-cloud-connector-example",
						tfjsonpath.New("settings").AtMapKey("config").AtSliceIndex(0).AtMapKey("external_id"),
						knownvalue.StringExact("sg-provider-test:ElfygiFglfldTwnDFpAScQkvgvHTGV"),
					),
				},
			},
			{
				Config: testAccResourceUpdate,
			},
		},
	})
}

func TestAccConnectorIncompatibleResourceName(t *testing.T) {
	// Test if the resource has name that is not compatible with the
	testResource := `resource "stackguardian_connector" "aws-cloud-connector-example" {
  resource_name = "aws rbac connector"
  description   = "AWS Cloud Connector"

  settings = {
    kind = "AWS_RBAC"

    config = [{
      role_arn         = "arn:aws:iam::209502960327:role/StackGuardian"
      external_id      = "sg-provider-test:ElfygiFglfldTwnDFpAScQkvgvHTGV"
      duration_seconds = "3600"
    }]
  }
}`
	testUpdateResource := `resource "stackguardian_connector" "aws-cloud-connector-example" {
  resource_name = "aws rbac connector"
  description   = "AWS Cloud Connector"

  settings = {
    kind = "AWS_RBAC"

    config = [{
      role_arn         = "arn:aws:iam::209502960327:role/StackGuardian"
      external_id      = "sg-provider-test:ElfygiFglfldTwnDFpAScQkvgvHTGV"
      duration_seconds = "360"
    }]
  }
}`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testResource,
			},
			{
				Config: testUpdateResource,
			},
		},
	})
}
