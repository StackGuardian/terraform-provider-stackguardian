package stacktemplate

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestToUpdateAPIModel_IsPublic(t *testing.T) {
	ctx := context.Background()
	base := func(isPublic types.String) StackTemplateResourceModel {
		return StackTemplateResourceModel{
			TemplateName:     types.StringValue("tpl"),
			SourceConfigKind: types.StringValue("TERRAFORM"),
			IsPublic:         isPublic,
		}
	}

	tests := []struct {
		name    string
		planned types.String
		prior   *StackTemplateResourceModel
		want    *string
	}{
		{name: "unchanged is omitted", planned: types.StringValue("0"), prior: ptr(base(types.StringValue("0"))), want: nil},
		{name: "changed is sent", planned: types.StringValue("1"), prior: ptr(base(types.StringValue("0"))), want: strPtr("1")},
		{name: "set for first time is sent", planned: types.StringValue("0"), prior: ptr(base(types.StringNull())), want: strPtr("0")},
		{name: "no prior state is sent", planned: types.StringValue("0"), prior: nil, want: strPtr("0")},
		{name: "null plan is omitted", planned: types.StringNull(), prior: ptr(base(types.StringValue("0"))), want: nil},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := base(tc.planned)
			got, diags := m.ToUpdateAPIModel(ctx, tc.prior)
			if diags.HasError() {
				t.Fatalf("unexpected diagnostics: %v", diags)
			}
			switch {
			case tc.want == nil && got.IsPublic != nil:
				t.Fatalf("expected IsPublic to be omitted, got %q", string(got.IsPublic.Value))
			case tc.want != nil && got.IsPublic == nil:
				t.Fatalf("expected IsPublic %q, got nil", *tc.want)
			case tc.want != nil && string(got.IsPublic.Value) != *tc.want:
				t.Fatalf("expected IsPublic %q, got %q", *tc.want, string(got.IsPublic.Value))
			}
		})
	}
}

func ptr(m StackTemplateResourceModel) *StackTemplateResourceModel { return &m }
func strPtr(s string) *string                                      { return &s }
