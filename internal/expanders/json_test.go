package expanders_test

import (
	"testing"

	"github.com/StackGuardian/terraform-provider-stackguardian/internal/expanders"
)

func TestJSONStringToObject(t *testing.T) {
	cases := []struct {
		in      string
		wantNil bool
		wantErr bool
	}{
		{"", true, false},
		{`{"bucket_region":"eu-central-1"}`, false, false},
		{`{}`, false, false},
		{`{"a":`, true, true}, // invalid JSON
		{`[1,2]`, true, true}, // not an object
		{`"text"`, true, true},
		{`null`, true, true},
	}
	for _, c := range cases {
		got, err := expanders.JSONStringToObject(c.in)
		if (err != nil) != c.wantErr {
			t.Errorf("JSONStringToObject(%q) error = %v, wantErr %v", c.in, err, c.wantErr)
		}
		if (got == nil) != c.wantNil {
			t.Errorf("JSONStringToObject(%q) = %v, wantNil %v", c.in, got, c.wantNil)
		}
	}
}
