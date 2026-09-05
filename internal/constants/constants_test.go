package constants

import (
	"strings"
	"testing"

	sgsdkgo "github.com/StackGuardian/sg-sdk-go"
)

// Several attributes document a fixed set of accepted values. Those sets come from
// the SDK, and nothing links the prose to the SDK at compile time, so a value added
// or renamed upstream leaves the documentation quietly wrong. Worse, three of these
// enums overlap without being interchangeable -- RAW_HCL is an IaC-input value but
// not a policy one, TIRITH_JSON the reverse -- so documenting a neighbouring enum's
// value points readers at something the attribute rejects.
//
// Referencing the SDK constants rather than string literals means an upstream rename
// breaks compilation here, which is a louder signal than a failing assertion.
func TestDocumentedEnumsMatchSDK(t *testing.T) {
	cases := []struct {
		name string
		doc  string
		// Every value the SDK accepts must appear in the prose.
		mustHave []string
		// Values from a neighbouring enum that this attribute does not accept.
		mustNotHave []string
	}{
		{
			name: "WorkflowIacInputDataSchemaType",
			doc:  WorkflowIacInputDataSchemaType,
			mustHave: []string{
				string(sgsdkgo.IacInputDataSchemaTypeEnumFormJsonschema),
				string(sgsdkgo.IacInputDataSchemaTypeEnumRawHcl),
				string(sgsdkgo.IacInputDataSchemaTypeEnumRawJson),
				string(sgsdkgo.IacInputDataSchemaTypeEnumNone),
			},
			mustNotHave: []string{
				string(sgsdkgo.InputDataSchemaTypeEnumTirithJson),
			},
		},
		{
			name: "PolicyConfigInputDataSchemaType",
			doc:  PolicyConfigInputDataSchemaType,
			mustHave: []string{
				string(sgsdkgo.InputDataSchemaTypeEnumFormJsonschema),
				string(sgsdkgo.InputDataSchemaTypeEnumRawJson),
				string(sgsdkgo.InputDataSchemaTypeEnumTirithJson),
				string(sgsdkgo.InputDataSchemaTypeEnumNone),
			},
			mustNotHave: []string{
				string(sgsdkgo.IacInputDataSchemaTypeEnumRawHcl),
			},
		},
		{
			name: "EnvVarKind",
			doc:  EnvVarKind,
			mustHave: []string{
				string(sgsdkgo.EnvVarsKindEnumPlainText),
				string(sgsdkgo.EnvVarsKindEnumVaultSecret),
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			for _, v := range tc.mustHave {
				if !strings.Contains(tc.doc, "`"+v+"`") {
					t.Errorf("%s does not document %q, which the SDK accepts", tc.name, v)
				}
			}
			for _, v := range tc.mustNotHave {
				if strings.Contains(tc.doc, "`"+v+"`") {
					t.Errorf("%s documents %q, which this attribute does not accept", tc.name, v)
				}
			}
		})
	}
}

// NO_CODE_JSON is absent from IacInputDataSchemaTypeEnum, yet the platform returns it
// on workflows that already exist. The provider passes it through, so the value has
// to be described without being offered as one to write.
func TestIacInputDataSchemaTypeDocumentsPlatformOnlyValue(t *testing.T) {
	const platformOnly = "NO_CODE_JSON"

	if !strings.Contains(WorkflowIacInputDataSchemaType, "`"+platformOnly+"`") {
		t.Errorf("WorkflowIacInputDataSchemaType no longer mentions %s; a workflow that "+
			"reports it would look undocumented", platformOnly)
	}

	// When the SDK starts accepting it, this note should become a documented value
	// alongside the others, and the caveat around it should go.
	if _, err := sgsdkgo.NewIacInputDataSchemaTypeEnumFromString(platformOnly); err == nil {
		t.Errorf("the SDK now accepts %s: promote it to a normal entry in "+
			"WorkflowIacInputDataSchemaType and drop the platform-only caveat", platformOnly)
	}
}

// The separator is a double colon. A single-dot spelling appears in some tooling, and
// it matches nothing on the platform, so a well-meaning correction to `${secret.name}`
// would silently give every reader a value that never resolves.
func TestSecretReferenceSyntaxUsesDoubleColon(t *testing.T) {
	if !strings.Contains(SecretReferenceSyntax, "${secret::") {
		t.Fatalf("SecretReferenceSyntax must document the ${secret::<name>} form, got: %s",
			SecretReferenceSyntax)
	}

	if strings.Contains(SecretReferenceSyntax, "${secret.") {
		t.Error("SecretReferenceSyntax documents the ${secret.<name>} form, which the " +
			"platform does not resolve; the separator is a double colon")
	}

	// Terraform reads a bare ${ as its own interpolation, so the escaped spelling is
	// the part a reader actually has to copy.
	if !strings.Contains(SecretReferenceSyntax, "$${secret::") {
		t.Error("SecretReferenceSyntax should show the escaped $${secret::<name>} form, " +
			"otherwise a copied example is interpolated by Terraform instead of sent verbatim")
	}
}

// Both attributes that accept a secret reference should describe it identically.
func TestSecretReferenceDocumentedWhereItIsAccepted(t *testing.T) {
	for name, doc := range map[string]string{
		"WorkflowIacInputDataData": WorkflowIacInputDataData,
		"EnvVarConfigTextValue":    EnvVarConfigTextValue,
	} {
		if !strings.Contains(doc, SecretReferenceSyntax) {
			t.Errorf("%s does not carry SecretReferenceSyntax, so the two descriptions "+
				"of the same syntax can drift", name)
		}
	}
}

// PLAIN_TEXT values land in configuration and state, so the description has to say so
// rather than leaving a reader to discover it after committing a credential.
func TestPlainTextVariableWarnsAboutExposure(t *testing.T) {
	for name, doc := range map[string]string{
		"EnvVarKind":            EnvVarKind,
		"EnvVarConfigTextValue": EnvVarConfigTextValue,
	} {
		if !strings.Contains(doc, "state") {
			t.Errorf("%s does not mention that the value is visible in state", name)
		}
		if !strings.Contains(doc, string(sgsdkgo.EnvVarsKindEnumVaultSecret)) {
			t.Errorf("%s does not point at %s as the alternative for credentials",
				name, sgsdkgo.EnvVarsKindEnumVaultSecret)
		}
	}
}
