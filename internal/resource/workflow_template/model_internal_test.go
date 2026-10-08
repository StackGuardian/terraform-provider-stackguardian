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
// `shared_orgs_list = []` reach the API on create. ToAPIModel must produce an empty,
// non-nil slice for both, and CreateWorkflowTemplateRequest tags them `omitzero`, which
// only drops a nil slice, so the empty list is encoded as []. With `omitempty`,
// encoding/json would leave both keys out and core would never store the empty value.
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
		VCSTriggers:      types.ObjectNull(VCSTriggersModel{}.AttributeTypes()),
	}

	req, diags := m.ToAPIModel(ctx)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if req.Tags == nil || len(req.Tags) != 0 {
		t.Fatalf("ToAPIModel Tags: got %#v, want []string{}", req.Tags)
	}
	if req.SharedOrgsList == nil || len(req.SharedOrgsList) != 0 {
		t.Fatalf("ToAPIModel SharedOrgsList: got %#v, want []string{}", req.SharedOrgsList)
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

// testRuntimeSource builds a runtime_source object with the given dest kind and repo; repo ""
// means config.repo is null.
func testRuntimeSource(destKind, repo string) types.Object {
	repoValue := types.StringNull()
	if repo != "" {
		repoValue = types.StringValue(repo)
	}
	config := types.ObjectValueMust(RuntimeSourceConfigModel{}.AttributeTypes(), map[string]attr.Value{
		"is_private":                 types.BoolValue(true),
		"auth":                       types.StringValue("/integrations/test-connector"),
		"git_core_auto_crlf":         types.BoolNull(),
		"git_sparse_checkout_config": types.StringNull(),
		"include_sub_module":         types.BoolNull(),
		"ref":                        types.StringNull(),
		"repo":                       repoValue,
		"working_dir":                types.StringNull(),
	})
	return types.ObjectValueMust(RuntimeSourceModel{}.AttributeTypes(), map[string]attr.Value{
		"source_config_dest_kind": types.StringValue(destKind),
		"config":                  config,
	})
}

// testVCSTriggers builds a vcs_triggers object with the given type and create_revision.enabled.
func testVCSTriggers(triggerType string, enabled bool) types.Object {
	createRevision := types.ObjectValueMust(VCSTriggersCreateRevisionModel{}.AttributeTypes(), map[string]attr.Value{
		"enabled": types.BoolValue(enabled),
	})
	createTag := types.ObjectValueMust(VCSTriggersCreateTagModel{}.AttributeTypes(), map[string]attr.Value{
		"create_revision": createRevision,
	})
	return types.ObjectValueMust(VCSTriggersModel{}.AttributeTypes(), map[string]attr.Value{
		"type":       types.StringValue(triggerType),
		"create_tag": createTag,
	})
}

// TestBuildCreateVcsTriggersRequest checks the body sent to the template's
// webhooks/vcs_triggers endpoint: useMarketplaceTemplate false (required by the API), the
// repository from runtime_source under VCSConfig.iacVCSConfig.customSource, and the triggers
// with create_tag.createRevision.
func TestBuildCreateVcsTriggersRequest(t *testing.T) {
	ctx := context.Background()
	m := WorkflowTemplateResourceModel{
		RuntimeSource: testRuntimeSource("GITHUB_COM", "https://github.com/StackGuardian/tf-null-resource.git"),
		VCSTriggers:   testVCSTriggers("GITHUB_COM", true),
	}

	req, diags := m.BuildCreateVcsTriggersRequest(ctx)
	if diags.HasError() {
		t.Fatalf("BuildCreateVcsTriggersRequest: %v", diags)
	}

	body, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	want := `{"VCSConfig":{"iacVCSConfig":{"useMarketplaceTemplate":false,"customSource":{"sourceConfigDestKind":"GITHUB_COM","config":{"auth":"/integrations/test-connector","isPrivate":true,"repo":"https://github.com/StackGuardian/tf-null-resource.git"}}}},"VCSTriggers":{"create_tag":{"createRevision":{"enabled":true}},"type":"GITHUB_COM"}}`
	if string(body) != want {
		t.Errorf("request body:\n got  %s\n want %s", body, want)
	}

	// No vcs_triggers: nothing to register.
	m.VCSTriggers = types.ObjectNull(VCSTriggersModel{}.AttributeTypes())
	req, diags = m.BuildCreateVcsTriggersRequest(ctx)
	if diags.HasError() || req != nil {
		t.Errorf("with vcs_triggers null: got request %v, diags %v; want nil", req, diags)
	}
}

// TestToUpdateAPIModel_VCSTriggers checks the template PATCH: vcs_triggers removed from
// config sends VCSTriggers: null to clear the stored triggers; vcs_triggers set leaves it out,
// since the webhooks/vcs_triggers call registers and stores them.
func TestToUpdateAPIModel_VCSTriggers(t *testing.T) {
	ctx := context.Background()
	base := WorkflowTemplateResourceModel{
		Id:               types.StringValue("tf-provider-test"),
		TemplateName:     types.StringValue("tf-provider-test"),
		OwnerOrg:         types.StringNull(),
		SourceConfigKind: types.StringValue("TERRAFORM"),
		IsPublic:         types.StringNull(),
		ShortDescription: types.StringNull(),
		RuntimeSource:    testRuntimeSource("GITHUB_COM", "https://github.com/StackGuardian/tf-null-resource.git"),
		SharedOrgsList:   types.ListNull(types.StringType),
		Tags:             types.ListNull(types.StringType),
		ContextTags:      types.MapNull(types.StringType),
	}

	for _, tc := range []struct {
		name        string
		vcsTriggers types.Object
		wantKey     bool
		wantNull    bool
	}{
		{name: "removed sends null", vcsTriggers: types.ObjectNull(VCSTriggersModel{}.AttributeTypes()), wantKey: true, wantNull: true},
		{name: "set is left out", vcsTriggers: testVCSTriggers("GITHUB_COM", true), wantKey: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := base
			m.VCSTriggers = tc.vcsTriggers
			req, diags := m.ToUpdateAPIModel(ctx)
			if diags.HasError() {
				t.Fatalf("ToUpdateAPIModel: %v", diags)
			}
			body, err := json.Marshal(req)
			if err != nil {
				t.Fatal(err)
			}
			var sent map[string]any
			if err := json.Unmarshal(body, &sent); err != nil {
				t.Fatal(err)
			}
			v, ok := sent["VCSTriggers"]
			if ok != tc.wantKey {
				t.Fatalf("VCSTriggers present = %v, want %v (body %s)", ok, tc.wantKey, body)
			}
			if tc.wantNull && v != nil {
				t.Errorf("VCSTriggers = %v, want null", v)
			}
		})
	}
}

// TestConvertVCSTriggersFromAPI checks the read path keeps type and create_revision.enabled.
func TestConvertVCSTriggersFromAPI(t *testing.T) {
	ctx := context.Background()
	var stored workflowtemplates.VCSTriggers
	// Stored triggers carry extra keys the API adds; they're not part of the SDK type.
	if err := json.Unmarshal([]byte(`{"type":"GITLAB_COM","create_tag":{"createRevision":{"enabled":true}},"post_comments":true,"gl_hook_id":"42"}`), &stored); err != nil {
		t.Fatal(err)
	}
	got, diags := convertVCSTriggersFromAPI(ctx, &stored)
	if diags.HasError() {
		t.Fatalf("convertVCSTriggersFromAPI: %v", diags)
	}
	if want := testVCSTriggers("GITLAB_COM", true); !got.Equal(want) {
		t.Errorf("got %s, want %s", got, want)
	}

	null, diags := convertVCSTriggersFromAPI(ctx, nil)
	if diags.HasError() || !null.IsNull() {
		t.Errorf("nil triggers: got %s, want null", null)
	}
}
