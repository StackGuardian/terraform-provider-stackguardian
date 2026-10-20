package stacktemplaterevisiondatasource

import "testing"

func TestParentTemplateId(t *testing.T) {
	for _, tc := range []struct{ id, want string }{
		{"my-stack-template:1", "my-stack-template"},
		{"my-stack-template:latest", "my-stack-template"},
		{"/my-org/my-stack-template:1", "my-stack-template"},
		{"my-stack-template", "my-stack-template"},
	} {
		if got := parentTemplateId(tc.id); got != tc.want {
			t.Errorf("parentTemplateId(%q) = %q, want %q", tc.id, got, tc.want)
		}
	}
}
