package stacktemplate_test

import (
	"context"
	"testing"

	sgclient "github.com/StackGuardian/sg-sdk-go/client"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccStackTemplate_CRUD(t *testing.T) {
	name := acctest.ResourceName("crud-stack-template")
	t.Cleanup(func() { deleteStackTemplateFixture(name) })

	// The updated config flips is_public, sets a description and adds a tag.
	acctest.RunCRUD(t, acctest.CRUD{
		ResourceType: "stackguardian_stack_template",
		Address:      "stackguardian_stack_template.test",
		Create:       testAccStackTemplateConfig(name, sourceConfigKind),
		Update:       testAccStackTemplateConfigUpdated(name, sourceConfigKind),
		UpdateChecks: []statecheck.StateCheck{
			statecheck.ExpectKnownValue("stackguardian_stack_template.test",
				tfjsonpath.New("description"), knownvalue.StringExact("Updated description")),
			statecheck.ExpectKnownValue("stackguardian_stack_template.test",
				tfjsonpath.New("is_public"), knownvalue.StringExact("1")),
		},
		Header: acctest.TemplateHeader(),
		Read: func(ctx context.Context, c *sgclient.Client, org string, rs *terraform.ResourceState) error {
			_, err := c.StackTemplates.ReadStackTemplate(ctx, org, rs.Primary.ID)
			return err
		},
		Delete: func(ctx context.Context, c *sgclient.Client, org string, rs *terraform.ResourceState) error {
			return c.StackTemplates.DeleteStackTemplate(ctx, org, rs.Primary.ID)
		},
	})
}
