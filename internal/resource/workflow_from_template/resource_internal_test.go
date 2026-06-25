package workflowfromtemplate

import "testing"

func TestSplitTemplateOrg(t *testing.T) {
	cases := []struct {
		name        string
		id          string
		callerOrg   string
		wantOrg     string
		wantRevisID string
	}{
		{
			name:        "own template fully-qualified",
			id:          "/demo-org/my-template:3",
			callerOrg:   "demo-org",
			wantOrg:     "demo-org",
			wantRevisID: "my-template:3",
		},
		{
			name:        "shared template from another org",
			id:          "/other-org/their-template:2",
			callerOrg:   "demo-org",
			wantOrg:     "other-org",
			wantRevisID: "their-template:2",
		},
		{
			name:        "bare id without org prefix falls back to caller org",
			id:          "my-template:1",
			callerOrg:   "demo-org",
			wantOrg:     "demo-org",
			wantRevisID: "my-template:1",
		},
		{
			name:        "empty id falls back to caller org",
			id:          "",
			callerOrg:   "demo-org",
			wantOrg:     "demo-org",
			wantRevisID: "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gotOrg, gotRevID := splitTemplateOrg(tc.id, tc.callerOrg)
			if gotOrg != tc.wantOrg || gotRevID != tc.wantRevisID {
				t.Errorf("splitTemplateOrg(%q, %q) = (%q, %q), want (%q, %q)",
					tc.id, tc.callerOrg, gotOrg, gotRevID, tc.wantOrg, tc.wantRevisID)
			}
		})
	}
}
