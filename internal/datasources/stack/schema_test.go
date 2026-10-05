package stackdatasource

import (
	"context"
	"testing"

	stackresource "github.com/StackGuardian/terraform-provider-stackguardian/internal/resource/stack"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// The data source reuses stackresource.StackResourceModel, so its schema must have exactly the
// resource schema's type — a mismatch only surfaces as a runtime error on Read otherwise.
func TestStackDataSourceSchemaMatchesResource(t *testing.T) {
	ctx := context.Background()

	var dsResp datasource.SchemaResponse
	(&stackDataSource{}).Schema(ctx, datasource.SchemaRequest{}, &dsResp)
	if dsResp.Diagnostics.HasError() {
		t.Fatalf("data source schema: %v", dsResp.Diagnostics)
	}
	if diags := dsResp.Schema.ValidateImplementation(ctx); diags.HasError() {
		t.Fatalf("data source schema is invalid: %v", diags)
	}

	var rResp resource.SchemaResponse
	stackresource.NewResource().Schema(ctx, resource.SchemaRequest{}, &rResp)
	if rResp.Diagnostics.HasError() {
		t.Fatalf("resource schema: %v", rResp.Diagnostics)
	}

	if got, want := dsResp.Schema.Type(), rResp.Schema.Type(); !got.Equal(want) {
		t.Fatalf("data source schema type does not match the resource schema type:\n got: %s\nwant: %s", got, want)
	}

	for _, name := range []string{"id", "workflow_group_id"} {
		a, ok := dsResp.Schema.Attributes[name].(dsschema.StringAttribute)
		if !ok || !a.Required {
			t.Errorf("%s must be a Required string attribute", name)
		}
	}
	for name, a := range dsResp.Schema.Attributes {
		if name == "id" || name == "workflow_group_id" {
			continue
		}
		if !a.IsComputed() || a.IsOptional() || a.IsRequired() {
			t.Errorf("%s must be Computed only", name)
		}
	}
}
