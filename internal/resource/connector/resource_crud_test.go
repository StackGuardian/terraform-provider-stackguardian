package connector_test

import (
	"context"
	"fmt"
	"testing"

	sgclient "github.com/StackGuardian/sg-sdk-go/client"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

const testAccCRUDConfig = `resource "stackguardian_connector" "this" {
  resource_name = %q
  description   = %q

  settings = {
    kind = "AWS_RBAC"

    config = [{
      role_arn         = "arn:aws:iam::209502960327:role/StackGuardian"
      external_id      = "sg-provider-test:ElfygiFglfldTwnDFpAScQkvgvHTGV"
      duration_seconds = %q
    }]
  }
}`

func TestAccConnector_CRUD(t *testing.T) {
	name := acctest.ResourceName("crud-connector")
	t.Cleanup(func() { deleteConnectorByResourceName(name) })

	acctest.RunCRUD(t, acctest.CRUD{
		ResourceType: "stackguardian_connector",
		Address:      "stackguardian_connector.this",
		Create:       fmt.Sprintf(testAccCRUDConfig, name, "created", "3600"),
		Update:       fmt.Sprintf(testAccCRUDConfig, name, "updated", "1800"),
		UpdateChecks: []statecheck.StateCheck{
			statecheck.ExpectKnownValue("stackguardian_connector.this",
				tfjsonpath.New("description"), knownvalue.StringExact("updated")),
			statecheck.ExpectKnownValue("stackguardian_connector.this",
				tfjsonpath.New("settings").AtMapKey("config").AtSliceIndex(0).AtMapKey("duration_seconds"),
				knownvalue.StringExact("1800")),
		},
		// The API returns external_id as "[REDACTED]". Read keeps the value from
		// prior state, but an import has none, so the first plan after an import
		// shows an update for it.
		ImportIgnore: []string{"settings.config.0.external_id"},
		Read: func(ctx context.Context, c *sgclient.Client, org string, rs *terraform.ResourceState) error {
			_, err := c.Connectors.ReadConnector(ctx, rs.Primary.ID, org)
			return err
		},
		Delete: func(ctx context.Context, c *sgclient.Client, org string, rs *terraform.ResourceState) error {
			_, err := c.Connectors.DeleteConnector(ctx, rs.Primary.ID, org)
			return err
		},
	})
}
