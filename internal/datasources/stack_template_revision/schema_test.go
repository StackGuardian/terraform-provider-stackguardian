package stacktemplaterevisiondatasource

import (
	"context"
	"testing"

	stacktemplaterevision "github.com/StackGuardian/terraform-provider-stackguardian/internal/resource/stack_template_revision"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Read decodes config into, and sets state from, stacktemplaterevision.StackTemplateRevisionResourceModel,
// so this schema must have exactly the resource schema's type — a mismatch only surfaces as a runtime
// error otherwise.
func TestStackTemplateRevisionDataSourceSchemaMatchesResource(t *testing.T) {
	ctx := context.Background()

	var dsResp datasource.SchemaResponse
	NewDataSource().Schema(ctx, datasource.SchemaRequest{}, &dsResp)
	if dsResp.Diagnostics.HasError() {
		t.Fatalf("data source schema: %v", dsResp.Diagnostics)
	}
	if diags := dsResp.Schema.ValidateImplementation(ctx); diags.HasError() {
		t.Fatalf("data source schema is invalid: %v", diags)
	}

	var rResp resource.SchemaResponse
	stacktemplaterevision.NewResource().Schema(ctx, resource.SchemaRequest{}, &rResp)
	if rResp.Diagnostics.HasError() {
		t.Fatalf("resource schema: %v", rResp.Diagnostics)
	}

	if got, want := dsResp.Schema.Type(), rResp.Schema.Type(); !got.Equal(want) {
		t.Fatalf("data source schema type does not match the resource schema type:\n got: %s\nwant: %s", got, want)
	}

	for name, a := range dsResp.Schema.Attributes {
		if name == "id" {
			if !a.IsRequired() {
				t.Errorf("id must be Required")
			}
			continue
		}
		if !a.IsComputed() || a.IsOptional() || a.IsRequired() {
			t.Errorf("%s must be Computed only", name)
		}
	}
}
