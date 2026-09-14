package connector_test

import (
	"context"
	"net/http"
	"os"
	"testing"

	sgsdkgo "github.com/StackGuardian/sg-sdk-go"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

var org = os.Getenv("STACKGUARDIAN_ORG_NAME")

// deleteConnectorByResourceName is a safety-net cleanup for a test that fails before
// Terraform's own destroy step runs. A connector's id is server-generated when the config
// doesn't set one explicitly, so this looks it up by resource_name via ListAllConnectors
// before deleting. Best-effort: errors and no-matches are ignored.
func deleteConnectorByResourceName(resourceName string) {
	client := acctest.SGClient()
	resp, err := client.Connectors.ListAllConnectors(context.TODO(), org, &sgsdkgo.ListAllConnectorsRequest{
		ResourceNames: sgsdkgo.String(resourceName),
	})
	if err != nil {
		return
	}
	for _, c := range resp.Msg {
		if c != nil && c.Msg != nil {
			client.Connectors.DeleteConnector(context.TODO(), c.Msg.Id, org)
		}
	}
}

// deleteConnectorByID is a safety-net cleanup for a test with an explicit, known id.
func deleteConnectorByID(id string) {
	client := acctest.SGClient()
	client.Connectors.DeleteConnector(context.TODO(), id, org)
}

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
	acctest.SkipUnlessAcceptance(t)

	t.Cleanup(func() { deleteConnectorByResourceName("aws-rbac-connector") })

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks:   acctest.VersionChecks(),
		ProtoV6ProviderFactories: acctest.ProviderFactories(http.Header{}),
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
	acctest.SkipUnlessAcceptance(t)

	t.Cleanup(func() { deleteConnectorByResourceName("aws rbac connector") })

	// Test if the resource has name that is not compatible with the
	testResource := `resource "stackguardian_connector" "aws-cloud-connector-example1" {
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
	testUpdateResource := `resource "stackguardian_connector" "aws-cloud-connector-example1" {
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
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks:   acctest.VersionChecks(),
		ProtoV6ProviderFactories: acctest.ProviderFactories(http.Header{}),
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

func TestAccConnectorOptionalId(t *testing.T) {
	acctest.SkipUnlessAcceptance(t)

	t.Cleanup(func() { deleteConnectorByID("aws_rbac_connector2") })

	// Test if the resource has name that is not compatible with the
	testResource := `resource "stackguardian_connector" "aws-cloud-connector-example2" {
  id = "aws_rbac_connector2"
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
	testUpdateResource := `resource "stackguardian_connector" "aws-cloud-connector-example2" {
  id = "aws_rbac_connector2"
  resource_name = "aws rbac connector update"
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
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks:   acctest.VersionChecks(),
		ProtoV6ProviderFactories: acctest.ProviderFactories(http.Header{}),
		Steps: []resource.TestStep{
			{
				Config: testResource,
				//Check:  resource.TestCheckResourceAttr("aws-cloud-connector-example2"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"stackguardian_connector.aws-cloud-connector-example2",
						tfjsonpath.New("id"),
						knownvalue.StringExact("aws_rbac_connector2"),
					),
				},
			},
			{
				Config: testUpdateResource,
			},
		},
	})
}
