package workflowtemplaterevision

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/StackGuardian/sg-sdk-go/workflowtemplaterevisions"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// listAttributePaths returns the path of every list attribute in attrs that can be set
// without creating list elements: top-level lists and lists inside single nested objects
// (terraform_config, runner_constraints, mini_steps). Lists nested inside list elements
// (e.g. wf_steps_config[*].mount_points) are skipped because reaching them needs an element
// to exist.
func listAttributePaths(attrs map[string]schema.Attribute, parent path.Path) []path.Path {
	var paths []path.Path
	for name, a := range attrs {
		p := parent.AtName(name)
		switch a := a.(type) {
		case schema.ListAttribute, schema.ListNestedAttribute:
			paths = append(paths, p)
		case schema.SingleNestedAttribute:
			paths = append(paths, listAttributePaths(a.Attributes, p)...)
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

// TestEmptyListAttributesReachThePayload checks, for every list attribute in the
// workflow_template_revision schema (top-level and inside nested objects), that `attr = []` produces a different create and
// update request body than leaving the attribute null. If the two bodies are identical,
// the explicit empty list was dropped (typically by `omitempty` on a []T SDK field), so
// core never stores it, GET omits it, and it reads back as null against the [] plan —
// "Provider produced inconsistent result after apply".
func TestEmptyListAttributesReachThePayload(t *testing.T) {
	ctx := context.Background()

	var schemaResp resource.SchemaResponse
	(&workflowTemplateRevisionResource{}).Schema(ctx, resource.SchemaRequest{}, &schemaResp)
	sch := schemaResp.Schema

	for _, p := range listAttributePaths(sch.Attributes, path.Empty()) {
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

// userScheduleWithInputs builds a user_schedules list with one schedule whose inputs match
// the API example: TerraformAction and VCSConfig.iacInputData. enableChaining is the planned
// value of the Computed-only enable_chaining: unknown on create, the state value on update.
func userScheduleWithInputs(t *testing.T, data string, enableChaining types.Bool) types.List {
	t.Helper()

	iacInputData := types.ObjectValueMust(UserScheduleIacInputDataModel{}.AttributeTypes(), map[string]attr.Value{
		"schema_type": types.StringValue("RAW_JSON"),
		"data":        types.StringValue(data),
	})
	vcsConfig := types.ObjectValueMust(UserScheduleVcsConfigModel{}.AttributeTypes(), map[string]attr.Value{
		"iac_input_data": iacInputData,
	})
	terraformAction := types.ObjectValueMust(UserScheduleTerraformActionModel{}.AttributeTypes(), map[string]attr.Value{
		"action": types.StringValue("apply"),
	})
	inputs := types.ObjectValueMust(UserScheduleInputsModel{}.AttributeTypes(), map[string]attr.Value{
		"terraform_action": terraformAction,
		"enable_chaining":  enableChaining,
		"vcs_config":       vcsConfig,
	})
	schedule := types.ObjectValueMust(UserSchedulesModel{}.AttributeTypes(), map[string]attr.Value{
		"cron":   types.StringValue("0 8 ? * MON *"),
		"state":  types.StringValue("ENABLED"),
		"desc":   types.StringNull(),
		"name":   types.StringNull(),
		"inputs": inputs,
	})
	return types.ListValueMust(types.ObjectType{AttrTypes: UserSchedulesModel{}.AttributeTypes()}, []attr.Value{schedule})
}

// TestUserScheduleInputsRoundTrip checks the update path: user_schedules[*].inputs is sent as
// the API expects, including the enable_chaining value carried over from state, with no
// empty MiniSteps (left out by `omitzero`), and reading it back yields the same object.
func TestUserScheduleInputsRoundTrip(t *testing.T) {
	ctx := context.Background()
	configured := userScheduleWithInputs(t, `{"test":"value"}`, types.BoolValue(true))

	schedules, diags := ConvertUserSchedulesToAPIModel(ctx, configured)
	if diags.HasError() {
		t.Fatalf("ConvertUserSchedulesToAPIModel: %v", diags)
	}

	body, err := json.Marshal(schedules)
	if err != nil {
		t.Fatal(err)
	}
	var sent []map[string]any
	if err := json.Unmarshal(body, &sent); err != nil {
		t.Fatal(err)
	}

	wantInputs := `{"EnableChaining":true,"TerraformAction":{"action":"apply"},"VCSConfig":{"iacInputData":{"data":{"test":"value"},"schemaType":"RAW_JSON"}}}`
	gotInputs, err := json.Marshal(sent[0]["inputs"])
	if err != nil {
		t.Fatal(err)
	}
	if string(gotInputs) != wantInputs {
		t.Errorf("inputs sent:\n got  %s\n want %s", gotInputs, wantInputs)
	}

	readBack, diags := ConvertUserSchedulesFromAPI(ctx, schedules)
	if diags.HasError() {
		t.Fatalf("ConvertUserSchedulesFromAPI: %v", diags)
	}
	if !readBack.Equal(configured) {
		t.Errorf("read back value differs from configured:\n got  %s\n want %s", readBack, configured)
	}
}

// TestUserScheduleInputsInvalidData checks that non-JSON inputs data is reported as an error
// instead of being sent to the API.
func TestUserScheduleInputsInvalidData(t *testing.T) {
	_, diags := ConvertUserSchedulesToAPIModel(context.Background(), userScheduleWithInputs(t, "not json", types.BoolUnknown()))
	if !diags.HasError() {
		t.Fatal("expected an error diagnostic for invalid inputs data")
	}
}

// TestUserScheduleInputsEnableChainingNotReturned checks a schedule whose EnableChaining the
// API doesn't return (for example an imported revision): it reads back as null, not an
// invented value, and a null or unknown value is never sent on the next update.
func TestUserScheduleInputsEnableChainingNotReturned(t *testing.T) {
	ctx := context.Background()

	schedules, diags := ConvertUserSchedulesToAPIModel(ctx, userScheduleWithInputs(t, `{"test":"value"}`, types.BoolUnknown()))
	if diags.HasError() {
		t.Fatalf("ConvertUserSchedulesToAPIModel: %v", diags)
	}
	if schedules[0].Inputs.EnableChaining != nil {
		t.Fatalf("EnableChaining sent on create: got %v, want omitted", *schedules[0].Inputs.EnableChaining)
	}

	readBack, diags := ConvertUserSchedulesFromAPI(ctx, schedules)
	if diags.HasError() {
		t.Fatalf("ConvertUserSchedulesFromAPI: %v", diags)
	}
	var models []UserSchedulesModel
	if diags := readBack.ElementsAs(ctx, &models, false); diags.HasError() {
		t.Fatal(diags)
	}
	var inputs UserScheduleInputsModel
	if diags := models[0].Inputs.As(ctx, &inputs, basetypes.ObjectAsOptions{}); diags.HasError() {
		t.Fatal(diags)
	}
	if !inputs.EnableChaining.IsNull() {
		t.Errorf("enable_chaining read back as %s, want null", inputs.EnableChaining)
	}

	// Next update with the null read back from the API: still not sent.
	schedules, diags = ConvertUserSchedulesToAPIModel(ctx, userScheduleWithInputs(t, `{"test":"value"}`, types.BoolNull()))
	if diags.HasError() {
		t.Fatalf("ConvertUserSchedulesToAPIModel: %v", diags)
	}
	if schedules[0].Inputs.EnableChaining != nil {
		t.Errorf("EnableChaining sent for a null value: got %v, want omitted", *schedules[0].Inputs.EnableChaining)
	}
}

// TestUserSchedulesFromAPIResponse decodes a UserSchedules entry exactly as the revision GET
// endpoint returns it and checks the read path keeps every value, including the
// Computed-only enable_chaining.
func TestUserSchedulesFromAPIResponse(t *testing.T) {
	ctx := context.Background()
	response := `[
	  {
	    "name": "",
	    "state": "ENABLED",
	    "cron": "0 12 ? * 2 *",
	    "desc": "",
	    "inputs": {
	      "VCSConfig": {
	        "iacInputData": {
	          "data": {"test": "value"},
	          "schemaType": "RAW_JSON"
	        }
	      },
	      "TerraformAction": {"action": "apply"},
	      "EnableChaining": true
	    }
	  }
	]`

	var schedules []workflowtemplaterevisions.UserSchedules
	if err := json.Unmarshal([]byte(response), &schedules); err != nil {
		t.Fatal(err)
	}

	list, diags := ConvertUserSchedulesFromAPI(ctx, schedules)
	if diags.HasError() {
		t.Fatalf("ConvertUserSchedulesFromAPI: %v", diags)
	}
	var models []UserSchedulesModel
	if diags := list.ElementsAs(ctx, &models, false); diags.HasError() {
		t.Fatal(diags)
	}
	var inputs UserScheduleInputsModel
	if diags := models[0].Inputs.As(ctx, &inputs, basetypes.ObjectAsOptions{}); diags.HasError() {
		t.Fatal(diags)
	}

	if !inputs.EnableChaining.Equal(types.BoolValue(true)) {
		t.Errorf("enable_chaining = %s, want true", inputs.EnableChaining)
	}
	// The UI stores "" for an empty name/desc; the read path keeps it as "" (not null) so the
	// Optional+Computed attributes carry it forward and send it back unchanged.
	if !models[0].Name.Equal(types.StringValue("")) {
		t.Errorf("name = %s, want \"\"", models[0].Name)
	}
	if !models[0].Desc.Equal(types.StringValue("")) {
		t.Errorf("desc = %s, want \"\"", models[0].Desc)
	}
}
