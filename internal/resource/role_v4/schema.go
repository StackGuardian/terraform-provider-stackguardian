package rolev4

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
)

// Schema defines the schema for the resource.
func (r *RoleV4Resource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	r.RoleResource.Schema(ctx, req, resp)
	resp.Schema.Attributes["doc_version"] = schema.StringAttribute{
		Computed: true,
		Default:  stringdefault.StaticString("V4"),
	}
}
