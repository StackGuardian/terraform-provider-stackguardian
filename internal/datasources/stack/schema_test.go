package stackdatasource

import (
	"context"
	"strings"
	"testing"

	stackresource "github.com/StackGuardian/terraform-provider-stackguardian/internal/resource/stack"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
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

// secretAttributeNames are the stack attributes that hold secrets and must be
// Sensitive wherever they appear, so Terraform redacts them in plan output.
var secretAttributeNames = map[string]bool{"webhook_secret": true, "text_value": true}

// TestStackSecretAttributesAreSensitive walks the resource and data source schemas and
// checks every occurrence of a secret attribute is Sensitive.
func TestStackSecretAttributesAreSensitive(t *testing.T) {
	ctx := context.Background()

	var rResp resource.SchemaResponse
	stackresource.NewResource().Schema(ctx, resource.SchemaRequest{}, &rResp)
	found := 0
	walkResourceAttributes("", rResp.Schema.Attributes, func(p string, a rschema.Attribute) {
		if secretAttributeNames[p[strings.LastIndex(p, ".")+1:]] {
			found++
			if !a.IsSensitive() {
				t.Errorf("resource attribute %s must be Sensitive", p)
			}
		}
	})
	if found == 0 {
		t.Fatal("no secret attributes found in the resource schema")
	}

	var dsResp datasource.SchemaResponse
	(&stackDataSource{}).Schema(ctx, datasource.SchemaRequest{}, &dsResp)
	dsFound := 0
	walkDataSourceAttributes("", dsResp.Schema.Attributes, func(p string, a dsschema.Attribute) {
		if secretAttributeNames[p[strings.LastIndex(p, ".")+1:]] {
			dsFound++
			if !a.IsSensitive() {
				t.Errorf("data source attribute %s must be Sensitive", p)
			}
		}
	})
	if dsFound != found {
		t.Errorf("found %d secret attributes in the data source schema, want %d", dsFound, found)
	}
}

// walkResourceAttributes calls fn for every attribute in attrs, recursing into nested attributes.
func walkResourceAttributes(prefix string, attrs map[string]rschema.Attribute, fn func(string, rschema.Attribute)) {
	for name, a := range attrs {
		p := prefix + name
		fn(p, a)
		switch n := a.(type) {
		case rschema.SingleNestedAttribute:
			walkResourceAttributes(p+".", n.Attributes, fn)
		case rschema.ListNestedAttribute:
			walkResourceAttributes(p+".", n.NestedObject.Attributes, fn)
		case rschema.MapNestedAttribute:
			walkResourceAttributes(p+".", n.NestedObject.Attributes, fn)
		case rschema.SetNestedAttribute:
			walkResourceAttributes(p+".", n.NestedObject.Attributes, fn)
		}
	}
}

// walkDataSourceAttributes calls fn for every attribute in attrs, recursing into nested attributes.
func walkDataSourceAttributes(prefix string, attrs map[string]dsschema.Attribute, fn func(string, dsschema.Attribute)) {
	for name, a := range attrs {
		p := prefix + name
		fn(p, a)
		switch n := a.(type) {
		case dsschema.SingleNestedAttribute:
			walkDataSourceAttributes(p+".", n.Attributes, fn)
		case dsschema.ListNestedAttribute:
			walkDataSourceAttributes(p+".", n.NestedObject.Attributes, fn)
		case dsschema.MapNestedAttribute:
			walkDataSourceAttributes(p+".", n.NestedObject.Attributes, fn)
		case dsschema.SetNestedAttribute:
			walkDataSourceAttributes(p+".", n.NestedObject.Attributes, fn)
		}
	}
}
