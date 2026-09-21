package flatteners

import "testing"

func TestListOfStringToTerraformList(t *testing.T) {
	cases := []struct {
		name     string
		in       []string
		wantNull bool
		wantLen  int
	}{
		{name: "nil slice becomes null", in: nil, wantNull: true},
		{name: "empty non-nil slice stays an empty list, not null", in: []string{}, wantNull: false, wantLen: 0},
		{name: "non-empty slice converts normally", in: []string{"a", "b"}, wantNull: false, wantLen: 2},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, diags := ListOfStringToTerraformList(tc.in)
			if diags.HasError() {
				t.Fatalf("unexpected error: %v", diags)
			}
			if got.IsNull() != tc.wantNull {
				t.Fatalf("IsNull() = %v, want %v", got.IsNull(), tc.wantNull)
			}
			if !tc.wantNull && len(got.Elements()) != tc.wantLen {
				t.Fatalf("len(Elements()) = %d, want %d", len(got.Elements()), tc.wantLen)
			}
		})
	}
}
