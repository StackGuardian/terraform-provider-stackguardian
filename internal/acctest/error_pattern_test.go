package acctest

import "testing"

func TestErrorPattern(t *testing.T) {
	cases := []struct {
		name string
		msg  string
		text string
		want bool
	}{
		{
			name: "exact match, no wrapping",
			msg:  "auth is required for this runtime_source",
			text: "auth is required for this runtime_source",
			want: true,
		},
		{
			name: "tolerates a line break where the CLI wrapped the text",
			msg:  "wf_steps_config is not allowed when source_config_kind is TERRAFORM or OPENTOFU; those workflow types use fixed, built-in run steps instead.",
			text: "wf_steps_config is not allowed when source_config_kind is TERRAFORM or\nOPENTOFU; those workflow types use fixed, built-in run steps instead.",
			want: true,
		},
		{
			name: "does not match unrelated text",
			msg:  "auth is required for this runtime_source",
			text: "some other error entirely",
			want: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ErrorPattern(tc.msg).MatchString(tc.text)
			if got != tc.want {
				t.Fatalf("ErrorPattern(%q).MatchString(%q) = %v, want %v", tc.msg, tc.text, got, tc.want)
			}
		})
	}
}
