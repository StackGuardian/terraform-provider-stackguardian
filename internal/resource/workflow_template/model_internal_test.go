package workflowtemplate

import (
	"context"
	"encoding/json"
	"testing"

	sgsdkgo "github.com/StackGuardian/sg-sdk-go"
	"github.com/StackGuardian/sg-sdk-go/workflowtemplates"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TestBuildAPIModelToWorkflowTemplateModel_StringLists verifies that tags and
// shared_orgs_list keep the difference between an absent field (nil → null) and an
// empty one ([] → known empty list). Collapsing [] to null makes `tags = []` or
// `shared_orgs_list = []` fail Terraform's post-apply consistency check, since the
// plan holds a known empty list.
func TestBuildAPIModelToWorkflowTemplateModel_StringLists(t *testing.T) {
	cases := []struct {
		name     string
		apiValue []string
		want     types.List
	}{
		{name: "absent field reads as null", apiValue: nil, want: types.ListNull(types.StringType)},
		{name: "empty field reads as empty list", apiValue: []string{}, want: types.ListValueMust(types.StringType, nil)},
		{name: "populated field reads as list", apiValue: []string{"a", "b"}, want: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("a"), types.StringValue("b")})},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			sourceConfigKind := workflowtemplates.WorkflowTemplateSourceConfigKindTerraform
			isPublic := sgsdkgo.IsPublicEnumZero

			model, diags := BuildAPIModelToWorkflowTemplateModel(&workflowtemplates.ReadWorkflowTemplateResponse{
				SourceConfigKind: &sourceConfigKind,
				IsPublic:         &isPublic,
				Tags:             tc.apiValue,
				SharedOrgsList:   tc.apiValue,
			})
			if diags.HasError() {
				t.Fatalf("unexpected diagnostics: %v", diags)
			}

			if !model.Tags.Equal(tc.want) {
				t.Errorf("tags: got %s, want %s", model.Tags, tc.want)
			}
			if !model.SharedOrgsList.Equal(tc.want) {
				t.Errorf("shared_orgs_list: got %s, want %s", model.SharedOrgsList, tc.want)
			}
		})
	}
}

// TestToAPIModel_EmptyStringListsSentInCreateRequest verifies that `tags = []` and
// `shared_orgs_list = []` reach the API on create. CreateWorkflowTemplateRequest holds
// both as *[]string, so `omitempty` only drops a nil pointer and an empty list is
// encoded as []. With plain []string, encoding/json would leave both keys out and core
// would never store the empty value.
func TestToAPIModel_EmptyStringListsSentInCreateRequest(t *testing.T) {
	ctx := context.Background()
	empty := types.ListValueMust(types.StringType, []attr.Value{})

	m := WorkflowTemplateResourceModel{
		Id:               types.StringNull(),
		TemplateName:     types.StringValue("tf-provider-test"),
		OwnerOrg:         types.StringNull(),
		SourceConfigKind: types.StringValue("TERRAFORM"),
		IsPublic:         types.StringNull(),
		ShortDescription: types.StringNull(),
		RuntimeSource:    types.ObjectNull(RuntimeSourceModel{}.AttributeTypes()),
		SharedOrgsList:   empty,
		Tags:             empty,
		ContextTags:      types.MapNull(types.StringType),
	}

	req, diags := m.ToAPIModel(ctx)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if req.Tags == nil || len(*req.Tags) != 0 {
		t.Fatalf("ToAPIModel Tags: got %#v, want &[]string{}", req.Tags)
	}
	if req.SharedOrgsList == nil || len(*req.SharedOrgsList) != 0 {
		t.Fatalf("ToAPIModel SharedOrgsList: got %#v, want &[]string{}", req.SharedOrgsList)
	}

	body, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	var sent map[string]any
	if err := json.Unmarshal(body, &sent); err != nil {
		t.Fatal(err)
	}

	for _, key := range []string{"Tags", "SharedOrgsList"} {
		v, ok := sent[key]
		if !ok {
			t.Errorf("%s = [] was dropped from the create request body: %s", key, body)
			continue
		}
		if list, isList := v.([]any); !isList || len(list) != 0 {
			t.Errorf("%s in create request body: got %#v, want []", key, v)
		}
	}
}
