package workflowtemplaterevision

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// topLevelListPaths returns the path of every top-level list attribute in attrs.
// Lists nested inside objects (terraform_config, runner_constraints, mini_steps) are
// held by shared SDK structs whose []T fields still drop [] and are not covered yet.
func topLevelListPaths(attrs map[string]schema.Attribute) []path.Path {
	var paths []path.Path
	for name, a := range attrs {
		switch a.(type) {
		case schema.ListAttribute, schema.ListNestedAttribute:
			paths = append(paths, path.Root(name))
		}
	}
	return paths
}

// revisionPayloads builds the create and update request bodies for a plan where the
// list at listPath is set to value (and its parent objects exist).
func revisionPayloads(t *testing.T, ctx context.Context, sch schema.Schema, listPath path.Path, value attr.Value) (create, update []byte) {
	t.Helper()

	plan := tfsdk.Plan{Schema: sch, Raw: tftypes.NewValue(sch.Type().TerraformType(ctx), nil)}
	if diags := plan.SetAttribute(ctx, path.Root("template_id"), types.StringValue("tpl")); diags.HasError() {
		t.Fatalf("set template_id: %v", diags)
	}
	if diags := plan.SetAttribute(ctx, path.Root("source_config_kind"), types.StringValue("CUSTOM")); diags.HasError() {
		t.Fatalf("set source_config_kind: %v", diags)
	}
	if diags := plan.SetAttribute(ctx, listPath, value); diags.HasError() {
		t.Fatalf("set %s: %v", listPath, diags)
	}

	var m WorkflowTemplateRevisionResourceModel
	if diags := plan.Get(ctx, &m); diags.HasError() {
		t.Fatalf("plan.Get with %s: %v", listPath, diags)
	}

	createReq, diags := m.ToAPIModel(ctx)
	if diags.HasError() {
		t.Fatalf("ToAPIModel with %s: %v", listPath, diags)
	}
	updateReq, diags := m.ToUpdateAPIModel(ctx)
	if diags.HasError() {
		t.Fatalf("ToUpdateAPIModel with %s: %v", listPath, diags)
	}

	var err error
	if create, err = json.Marshal(createReq); err != nil {
		t.Fatal(err)
	}
	if update, err = json.Marshal(updateReq); err != nil {
		t.Fatal(err)
	}
	return create, update
}

// TestEmptyListAttributesReachThePayload checks, for every top-level list attribute in
// the workflow_template_revision schema, that `attr = []` produces a different create and
// update request body than leaving the attribute null. If the two bodies are identical,
// the explicit empty list was dropped (typically by `omitempty` on a []T SDK field), so
// core never stores it, GET omits it, and it reads back as null against the [] plan —
// "Provider produced inconsistent result after apply".
func TestEmptyListAttributesReachThePayload(t *testing.T) {
	ctx := context.Background()

	var schemaResp resource.SchemaResponse
	(&workflowTemplateRevisionResource{}).Schema(ctx, resource.SchemaRequest{}, &schemaResp)
	sch := schemaResp.Schema

	for _, p := range topLevelListPaths(sch.Attributes) {
		t.Run(p.String(), func(t *testing.T) {
			attrType, diags := sch.TypeAtPath(ctx, p)
			if diags.HasError() {
				t.Fatalf("type at %s: %v", p, diags)
			}
			elemType := attrType.(types.ListType).ElemType

			nullCreate, nullUpdate := revisionPayloads(t, ctx, sch, p, types.ListNull(elemType))
			emptyCreate, emptyUpdate := revisionPayloads(t, ctx, sch, p, types.ListValueMust(elemType, []attr.Value{}))

			if bytes.Equal(nullCreate, emptyCreate) {
				t.Errorf("create: %s = [] was dropped from the request body: %s", p, emptyCreate)
			}
			if bytes.Equal(nullUpdate, emptyUpdate) {
				t.Errorf("update: %s = [] was dropped from the request body: %s", p, emptyUpdate)
			}
		})
	}
}
