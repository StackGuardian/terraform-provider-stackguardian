package stackdatasource

import (
	"context"
	"testing"

	stackresource "github.com/StackGuardian/terraform-provider-stackguardian/internal/resource/stack"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
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

// webhook_secret must stay out of plan output on both the resource and the data source, which
// inherits Sensitive from the resource schema.
func TestStackWebhookSecretIsSensitive(t *testing.T) {
	ctx := context.Background()

	var rResp resource.SchemaResponse
	stackresource.NewResource().Schema(ctx, resource.SchemaRequest{}, &rResp)
	var dsResp datasource.SchemaResponse
	(&stackDataSource{}).Schema(ctx, datasource.SchemaRequest{}, &dsResp)

	for _, event := range []string{"approval_required", "cancelled", "completed", "drift_detected", "errored"} {
		p := path.Root("workflows_config").AtName("workflows").AtListIndex(0).
			AtName("mini_steps").AtName("webhooks").AtName(event).AtListIndex(0).AtName("webhook_secret")

		ra, diags := rResp.Schema.AttributeAtPath(ctx, p)
		if diags.HasError() {
			t.Fatalf("resource %s: %v", p, diags)
		}
		if !ra.IsSensitive() {
			t.Errorf("resource %s must be Sensitive", p)
		}

		da, diags := dsResp.Schema.AttributeAtPath(ctx, p)
		if diags.HasError() {
			t.Fatalf("data source %s: %v", p, diags)
		}
		if !da.IsSensitive() {
			t.Errorf("data source %s must be Sensitive", p)
		}
	}
}
