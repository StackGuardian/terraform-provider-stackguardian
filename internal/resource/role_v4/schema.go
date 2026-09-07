package rolev4

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
)

// Schema defines the schema for the resource.
//
// The attribute set is inherited wholesale from stackguardian_role, so the
// schema-level description and the deprecation notice set there must both be
// overridden here — otherwise stackguardian_rolev4 would document itself as the
// deprecated V3 resource.
func (r *RoleV4Resource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	r.RoleResource.Schema(ctx, req, resp)
	resp.Schema.MarkdownDescription = "Manages a role using the V4 permission document. Prefer this resource over `stackguardian_role` for new configurations: `stackguardian_role` expands permissions by combining every path with every other path, whereas `stackguardian_rolev4` maps paths one to one, which is almost always what you want. Roles are granted to users through a `stackguardian_role_assignment`."
	resp.Schema.DeprecationMessage = ""
	resp.Schema.Attributes["doc_version"] = schema.StringAttribute{
		MarkdownDescription: "Version of the permission document format. Always `V4` for this resource; it is what distinguishes it from `stackguardian_role`, which uses V3 semantics.",
		Computed:            true,
		Default:             stringdefault.StaticString("V4"),
	}
}
