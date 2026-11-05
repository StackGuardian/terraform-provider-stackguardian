package apierrors

import (
	"errors"
	"fmt"
	"testing"

	"github.com/StackGuardian/sg-sdk-go/core"
)

func TestIsNotFound(t *testing.T) {
	apiErr := func(code int, body string) error { return core.NewAPIError(code, errors.New(body)) }

	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"plain error", errors.New("does not exist"), false},
		{"404", apiErr(404, `{"msg":"Workflow does not exist"}`), true},
		{"400 missing template", apiErr(400, `{"msg":"Template /org/x of type IAC does not exist"}`), true},
		{"400 validation", apiErr(400, `{"msg":"field may not be blank"}`), false},
		{"403", apiErr(403, `{"msg":"Unauthorized"}`), false},
		{"wrapped 404", fmt.Errorf("reading: %w", apiErr(404, "gone")), true},
	}
	for _, c := range cases {
		if got := IsNotFound(c.err); got != c.want {
			t.Errorf("%s: IsNotFound = %v, want %v", c.name, got, c.want)
		}
	}
}
