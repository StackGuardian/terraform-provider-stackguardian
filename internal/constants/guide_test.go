package constants

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// runtimeReferencesGuide is the one place the reference syntax is written down. Attribute
// descriptions only link to it, so nothing else in the repo can be checked against the
// platform -- which makes this file the thing worth pinning.
const runtimeReferencesGuide = "../../docs-templates/guides/RuntimeReferences.md"

// referenceForms are the prefixes the platform resolves. Each separator is a double
// colon: a single dot matches nothing and fails silently at run time, so a plausible
// "correction" to ${secret.name} would hand every reader a value that never resolves.
var referenceForms = []string{
	"${secret::",
	"${ext-secret::azure-kv::",
	"${workflow::",
	"${reference::",
}

func readGuide(t *testing.T) string {
	t.Helper()

	body, err := os.ReadFile(filepath.Clean(runtimeReferencesGuide))
	if err != nil {
		t.Fatalf("cannot read the Runtime References guide, which every attribute "+
			"description links to: %v", err)
	}
	return string(body)
}

func TestGuideDocumentsEveryReferenceForm(t *testing.T) {
	guide := readGuide(t)

	for _, form := range referenceForms {
		if !strings.Contains(guide, form) {
			t.Errorf("the guide does not document %q, so an attribute linking to it "+
				"leaves that form undocumented", form)
		}
	}
}

// A single dot is the spelling that appears in some tooling. It resolves to nothing.
func TestGuideUsesDoubleColonSeparator(t *testing.T) {
	guide := readGuide(t)

	for _, wrong := range []string{"${secret.", "${workflow.", "${reference."} {
		if strings.Contains(guide, wrong) {
			t.Errorf("the guide contains %q; the separator is a double colon and the "+
				"single-dot form does not resolve", wrong)
		}
	}
}

// Stacks are addressed through ${workflow::<group>.<stack>.<workflow>.<key>}. A
// ${stack::...} form is a natural guess and does not exist.
func TestGuideDoesNotInventAStackForm(t *testing.T) {
	if guide := readGuide(t); strings.Contains(guide, "${stack::") {
		t.Error("the guide documents a ${stack::...} form, which the platform has no " +
			"such thing as; a stack's workflows are addressed through ${workflow::")
	}
}

// Terraform reads a bare ${ as its own interpolation, so a reader copying an example
// without the doubled dollar gets an interpolation error or a silently wrong value.
func TestGuideShowsTheTerraformEscape(t *testing.T) {
	if guide := readGuide(t); !strings.Contains(guide, "$${") {
		t.Error("the guide never shows the escaped $${ form, so a copied example is " +
			"interpolated by Terraform instead of being sent verbatim")
	}
}
