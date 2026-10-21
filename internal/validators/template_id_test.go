package validators_test

import (
	"context"
	"testing"

	"github.com/StackGuardian/terraform-provider-stackguardian/internal/validators"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestQualifiedTemplateRevisionId(t *testing.T) {
	cases := []struct {
		value string
		valid bool
	}{
		{"/my-org/my-template:1", true},
		{"/stackguardian/aws-s3-demo-website:latest", true},
		{"/my-org/my-template", false}, // revision must be given
		{"my-template:1", false},       // org must be given, even your own
		{"my-template", false},
		{"/my-org/my-template:1:2", false},
		{"/my-org/nested/my-template:1", false},
		{"//my-template:1", false},
		{"/my-org/:1", false},
		{"/my-org/my-template:", false},
	}

	for _, c := range cases {
		resp := &validator.StringResponse{}
		validators.QualifiedTemplateRevisionId().ValidateString(context.Background(), validator.StringRequest{
			Path:        path.Root("id"),
			ConfigValue: types.StringValue(c.value),
		}, resp)
		if got := !resp.Diagnostics.HasError(); got != c.valid {
			t.Errorf("QualifiedTemplateRevisionId(%q) valid = %v, want %v", c.value, got, c.valid)
		}
	}
}
