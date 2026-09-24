package expanders

import (
	"encoding/json"
	"testing"
)

func TestPointer(t *testing.T) {
	if got := Pointer([]string(nil)); got != nil {
		t.Errorf("nil slice: got %v, want nil", got)
	}
	if got := Pointer(map[string]string(nil)); got != nil {
		t.Errorf("nil map: got %v, want nil", got)
	}
	if got := Pointer((*int)(nil)); got != nil {
		t.Errorf("nil pointer: got %v, want nil", got)
	}
	if got := Pointer([]string{}); got == nil || *got == nil || len(*got) != 0 {
		t.Errorf("empty slice: got %v, want pointer to non-nil []", got)
	}
	if got := Pointer("x"); got == nil || *got != "x" {
		t.Errorf("string: got %v, want pointer to \"x\"", got)
	}
	if got := Pointer(0); got == nil || *got != 0 {
		t.Errorf("zero int: got %v, want pointer to 0", got)
	}

	// How the result encodes in a request body with omitempty.
	type request struct {
		List *[]string `json:"List,omitempty"`
	}
	for _, tc := range []struct {
		in   []string
		want string
	}{
		{in: nil, want: `{}`},
		{in: []string{}, want: `{"List":[]}`},
		{in: []string{"a"}, want: `{"List":["a"]}`},
	} {
		body, err := json.Marshal(request{List: Pointer(tc.in)})
		if err != nil {
			t.Fatal(err)
		}
		if string(body) != tc.want {
			t.Errorf("Pointer(%#v) encodes as %s, want %s", tc.in, body, tc.want)
		}
	}
}
