package constants_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/StackGuardian/terraform-provider-stackguardian/internal/constants"
	"github.com/hashicorp/go-version"
)

// matrixEntry matches one hardcoded Terraform version in a workflow's
// `terraform:` matrix list.
var matrixEntry = regexp.MustCompile(`^\s*- "(\d+\.\d+\.\d+)"`)

// TestSupportedTerraformVersions checks the shape of the matrix itself: it runs
// from the support floor to the newest tested release, hitting the latest patch
// of every minor series in between with no gaps.
func TestSupportedTerraformVersions(t *testing.T) {
	t.Parallel()

	versions := constants.SupportedTerraformVersions
	if len(versions) < 2 {
		t.Fatalf("expected at least a floor and a ceiling, got %v", versions)
	}

	if got, want := versions[0], constants.MinTerraformVersion.String(); got != want {
		t.Errorf("first matrix entry is %s, expected MinTerraformVersion %s", got, want)
	}
	if got, want := versions[len(versions)-1], constants.MaxTerraformVersion.String(); got != want {
		t.Errorf("last matrix entry is %s, expected MaxTerraformVersion %s", got, want)
	}

	var previous *version.Version
	for i, raw := range versions {
		current, err := version.NewVersion(raw)
		if err != nil {
			t.Fatalf("matrix entry %d (%q) is not a valid version: %v", i, raw, err)
		}

		if previous == nil {
			previous = current
			continue
		}

		if !current.GreaterThan(previous) {
			t.Errorf("matrix is not ascending: %s follows %s", current, previous)
		}

		// Each entry must open a new minor series, and the series must be
		// contiguous — a gap means a whole Terraform minor goes untested.
		previousSegments, currentSegments := previous.Segments(), current.Segments()
		if currentSegments[0] != previousSegments[0] || currentSegments[1] != previousSegments[1]+1 {
			t.Errorf("matrix skips a minor series between %s and %s", previous, current)
		}

		previous = current
	}
}

// TestTerraformMatrixMatchesWorkflows fails if a workflow's hardcoded matrix
// drifts from constants.SupportedTerraformVersions, so the list in Go stays the
// one place to edit when the supported range changes.
func TestTerraformMatrixMatchesWorkflows(t *testing.T) {
	t.Parallel()

	workflows := []string{
		filepath.Join("..", "..", ".github", "workflows", "compat.yaml"),
		filepath.Join("..", "..", ".github", "workflows", "test.yaml"),
	}

	for _, workflow := range workflows {
		t.Run(filepath.Base(workflow), func(t *testing.T) {
			t.Parallel()

			content, err := os.ReadFile(workflow)
			if err != nil {
				t.Fatalf("reading %s: %v", workflow, err)
			}

			found := terraformMatrix(string(content))
			if len(found) == 0 {
				t.Fatalf("no terraform matrix found in %s", workflow)
			}

			if strings.Join(found, ",") != strings.Join(constants.SupportedTerraformVersions, ",") {
				t.Errorf("%s matrix is\n  %v\nbut constants.SupportedTerraformVersions is\n  %v\n"+
					"Update the workflow to match the Go list.",
					workflow, found, constants.SupportedTerraformVersions)
			}
		})
	}
}

// TestMakefileUsesMinTerraformVersion keeps the test-acc-min-terraform target
// pinned to the real support floor.
func TestMakefileUsesMinTerraformVersion(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile(filepath.Join("..", "..", "Makefile"))
	if err != nil {
		t.Fatalf("reading Makefile: %v", err)
	}

	want := "TF_ACC_TERRAFORM_VERSION=" + constants.MinTerraformVersion.String()
	if !strings.Contains(string(content), want) {
		t.Errorf("Makefile target test-acc-min-terraform does not pin %q", want)
	}
}

// terraformMatrix pulls the versions out of the `terraform:` matrix list,
// stopping at the first line that is no longer a list entry.
func terraformMatrix(content string) []string {
	lines := strings.Split(content, "\n")

	var versions []string
	collecting := false

	for _, line := range lines {
		if strings.TrimSpace(line) == "terraform:" {
			collecting = true
			continue
		}
		if !collecting {
			continue
		}

		match := matrixEntry.FindStringSubmatch(line)
		if match == nil {
			// Comments and blank lines sit between entries; anything else ends
			// the list.
			trimmed := strings.TrimSpace(line)
			if trimmed == "" || strings.HasPrefix(trimmed, "#") {
				continue
			}
			break
		}

		versions = append(versions, match[1])
	}

	return versions
}
