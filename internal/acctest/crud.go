package acctest

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	sgclient "github.com/StackGuardian/sg-sdk-go/client"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/apierrors"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// CRUD step layout. Every TestAcc*_CRUD test runs these steps in this order, and
// scripts/crud-matrix maps a failing "Step N/M" back to the operation through
// it. Keep the two in sync.
//
//	Step 1: create  -- apply Config, then the framework's refresh + empty-plan check
//	Step 2: update  -- plan must be an in-place update of Address, then apply
//	Step 3: import  -- import by ID, then ImportStateVerify against the state
//	Step 4: gone    -- (only with Delete) delete the object behind the API's back;
//	                   refresh must drop it from state and plan a create
//	destroy         -- framework destroy, then CheckDestroy reads the API for a 404
const (
	CRUDStepCreate = 1
	CRUDStepUpdate = 2
	CRUDStepImport = 3
	CRUDStepGone   = 4
)

// ReadFunc fetches the API object behind one resource left in Terraform state.
// It returns the SDK error unchanged so CheckDestroy can tell a 404 apart from a
// real failure.
type ReadFunc func(ctx context.Context, client *sgclient.Client, org string, rs *terraform.ResourceState) error

// DeleteFunc deletes the API object behind one resource in Terraform state.
type DeleteFunc func(ctx context.Context, client *sgclient.Client, org string, rs *terraform.ResourceState) error

// CRUD describes one resource's create -> update -> import -> delete cycle.
type CRUD struct {
	// ResourceType is the Terraform type, e.g. "stackguardian_connector".
	ResourceType string
	// Address is the resource under test, e.g. "stackguardian_connector.this".
	Address string

	// Create and Update are complete configurations. Update must change at least
	// one attribute of Address in place.
	Create string
	Update string

	CreateChecks []statecheck.StateCheck
	UpdateChecks []statecheck.StateCheck

	// ImportID builds the import ID from state. Nil imports by the "id" attribute.
	ImportID resource.ImportStateIdFunc
	// ImportVerifyIdentifier is the attribute ImportStateVerify matches the
	// imported resource on. Empty means "id"; set it for resources without one.
	ImportVerifyIdentifier string
	// ImportIgnore lists attributes the API cannot return (write-only, or
	// config-only) and so are excluded from ImportStateVerify.
	ImportIgnore []string

	// Header is sent by the provider and by CheckDestroy's client. Template
	// resources need TemplateHeader().
	Header http.Header

	// Read is used by CheckDestroy for every ResourceType left in state.
	Read ReadFunc
	// Delete, when set, adds the "gone" step: the object is deleted outside
	// Terraform, and the provider must notice on refresh and plan to recreate it.
	Delete DeleteFunc
}

// RunCRUD runs the standard CRUD cycle for c.
func RunCRUD(t *testing.T, c CRUD) {
	t.Helper()
	SkipUnlessAcceptance(t)

	header := c.Header
	if header == nil {
		header = http.Header{}
	}

	// The update step's state, for the gone step to delete from. Import steps
	// do not persist their state, so this is still the live object.
	var updated *terraform.ResourceState

	steps := []resource.TestStep{
		{
			Config:            c.Create,
			ConfigStateChecks: c.CreateChecks,
		},
		{
			Config: c.Update,
			ConfigPlanChecks: resource.ConfigPlanChecks{
				PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction(c.Address, plancheck.ResourceActionUpdate),
				},
			},
			ConfigStateChecks: c.UpdateChecks,
			Check: func(s *terraform.State) error {
				updated = s.RootModule().Resources[c.Address]
				if updated == nil {
					return fmt.Errorf("%s not found in state after update", c.Address)
				}
				return nil
			},
		},
		{
			Config:                  c.Update,
			ResourceName:            c.Address,
			ImportState:             true,
			ImportStateIdFunc:       c.ImportID,
			ImportStateVerify:       true,
			ImportStateVerifyIgnore: c.ImportIgnore,

			ImportStateVerifyIdentifierAttribute: c.ImportVerifyIdentifier,
		},
	}

	if c.Delete != nil {
		steps = append(steps, resource.TestStep{
			PreConfig: func() {
				if err := c.Delete(context.Background(), SGClientWithHeader(header), config.Get().OrgName, updated); err != nil {
					t.Fatalf("deleting %s outside Terraform: %s", c.Address, err)
				}
			},
			Config: c.Update,
			ConfigPlanChecks: resource.ConfigPlanChecks{
				PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction(c.Address, plancheck.ResourceActionCreate),
				},
			},
		})
	}

	// scripts/crud-matrix reads this line to know whether the gone step ran.
	t.Logf("crud: resource=%s steps=%d", c.ResourceType, len(steps))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { TestAccPreCheck(t) },
		TerraformVersionChecks:   VersionChecks(),
		ProtoV6ProviderFactories: ProviderFactories(header),
		CheckDestroy:             CheckDestroy(c.ResourceType, header, c.Read),
		Steps:                    steps,
	})
}

// CheckDestroy asserts that no resource of resourceType left in state still
// exists in the API. Terraform's destroy only proves the provider returned no
// error; this proves the object is actually gone.
func CheckDestroy(resourceType string, header http.Header, read ReadFunc) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ctx := context.Background()
		client := SGClientWithHeader(header)
		org := config.Get().OrgName

		for name, rs := range s.RootModule().Resources {
			if rs.Type != resourceType {
				continue
			}
			err := read(ctx, client, org, rs)
			if err == nil {
				return fmt.Errorf("%s (id %q) still exists in the API after destroy", name, rs.Primary.ID)
			}
			if !IsNotFound(err) {
				return fmt.Errorf("checking %s (id %q) after destroy: %w", name, rs.Primary.ID, err)
			}
		}
		return nil
	}
}

// TemplateHeader returns the header the template APIs (stack, workflow and
// workflow step templates and their revisions) need to resolve the org.
func TemplateHeader() http.Header {
	h := http.Header{}
	h.Set("x-sg-internal-auth-orgid", config.Get().OrgName)
	return h
}

// IsNotFound reports whether err says the object does not exist. It is the
// provider's own definition, so tests and Read/Delete agree on what "gone" is.
func IsNotFound(err error) bool {
	return apierrors.IsNotFound(err)
}

// ImportIDFromAttrs returns an import ID func joining the given state
// attributes of address with "/" -- e.g. "workflow_group_id", "id" for
// workflows.
func ImportIDFromAttrs(address string, attrs ...string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[address]
		if !ok {
			return "", fmt.Errorf("%s not found in state", address)
		}
		id := ""
		for i, a := range attrs {
			v, ok := rs.Primary.Attributes[a]
			if !ok || v == "" {
				return "", fmt.Errorf("%s has no %q attribute", address, a)
			}
			if i > 0 {
				id += "/"
			}
			id += v
		}
		return id, nil
	}
}
