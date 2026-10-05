package stackdatasource

import (
	"context"
	"fmt"

	"github.com/StackGuardian/terraform-provider-stackguardian/internal/constants"
	stackresource "github.com/StackGuardian/terraform-provider-stackguardian/internal/resource/stack"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// Schema is derived from the stackguardian_stack resource's schema rather than restated: the data
// source reuses the resource's StackResourceModel, whose AttributeTypes() must match this schema
// exactly, and deriving it keeps the two (and their descriptions) from drifting apart.
func (d *stackDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	var rresp resource.SchemaResponse
	stackresource.NewResource().Schema(ctx, resource.SchemaRequest{}, &rresp)
	resp.Diagnostics.Append(rresp.Diagnostics...)
	if resp.Diagnostics.HasError() {
		return
	}

	attrs, err := computedAttributes(rresp.Schema.Attributes)
	if err != nil {
		resp.Diagnostics.AddError("Unable to build stack data source schema", err.Error())
		return
	}

	attrs["id"] = dsschema.StringAttribute{
		MarkdownDescription: constants.DatasourceId,
		Required:            true,
	}
	attrs["workflow_group_id"] = dsschema.StringAttribute{
		MarkdownDescription: constants.WorkflowWorkflowGroupId,
		Required:            true,
	}

	resp.Schema = dsschema.Schema{
		MarkdownDescription: "Reads an existing stack, including its workflows and actions.",
		Attributes:          attrs,
	}
}

func computedAttributes(in map[string]rschema.Attribute) (map[string]dsschema.Attribute, error) {
	out := make(map[string]dsschema.Attribute, len(in))
	for name, a := range in {
		c, err := computedAttribute(a)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}
		out[name] = c
	}
	return out, nil
}

func computedAttribute(a rschema.Attribute) (dsschema.Attribute, error) {
	switch v := a.(type) {
	case rschema.StringAttribute:
		return dsschema.StringAttribute{MarkdownDescription: v.MarkdownDescription, Sensitive: v.Sensitive, Computed: true}, nil
	case rschema.BoolAttribute:
		return dsschema.BoolAttribute{MarkdownDescription: v.MarkdownDescription, Sensitive: v.Sensitive, Computed: true}, nil
	case rschema.Int64Attribute:
		return dsschema.Int64Attribute{MarkdownDescription: v.MarkdownDescription, Sensitive: v.Sensitive, Computed: true}, nil
	case rschema.ListAttribute:
		return dsschema.ListAttribute{MarkdownDescription: v.MarkdownDescription, Sensitive: v.Sensitive, ElementType: v.ElementType, Computed: true}, nil
	case rschema.MapAttribute:
		return dsschema.MapAttribute{MarkdownDescription: v.MarkdownDescription, Sensitive: v.Sensitive, ElementType: v.ElementType, Computed: true}, nil
	case rschema.SingleNestedAttribute:
		nested, err := computedAttributes(v.Attributes)
		if err != nil {
			return nil, err
		}
		return dsschema.SingleNestedAttribute{MarkdownDescription: v.MarkdownDescription, Sensitive: v.Sensitive, Attributes: nested, Computed: true}, nil
	case rschema.ListNestedAttribute:
		nested, err := computedAttributes(v.NestedObject.Attributes)
		if err != nil {
			return nil, err
		}
		return dsschema.ListNestedAttribute{MarkdownDescription: v.MarkdownDescription, Sensitive: v.Sensitive, NestedObject: dsschema.NestedAttributeObject{Attributes: nested}, Computed: true}, nil
	case rschema.MapNestedAttribute:
		nested, err := computedAttributes(v.NestedObject.Attributes)
		if err != nil {
			return nil, err
		}
		return dsschema.MapNestedAttribute{MarkdownDescription: v.MarkdownDescription, Sensitive: v.Sensitive, NestedObject: dsschema.NestedAttributeObject{Attributes: nested}, Computed: true}, nil
	default:
		return nil, fmt.Errorf("unsupported attribute type %T", a)
	}
}
