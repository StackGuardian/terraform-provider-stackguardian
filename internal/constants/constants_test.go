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
		// The SDK still carries VAULT_SECRET, but the platform does not accept it,
		// so the description must not offer it as a value to write.
		{
			name: "EnvVarKind",
			doc:  EnvVarKind,
			mustHave: []string{
				string(sgsdkgo.EnvVarsKindEnumPlainText),
			},
			mustNotHave: []string{
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

// Every attribute whose value the platform resolves at run time has to point at the
// guide. Without this, a new reference-accepting attribute ships with no way for a
// reader to learn the syntax, which is how ${secret::...} came to be the only form
// documented anywhere in the provider.
func TestRuntimeReferenceDocumentedWhereItIsAccepted(t *testing.T) {
	for name, doc := range map[string]string{
		"WorkflowIacInputDataData": WorkflowIacInputDataData,
		"EnvVarConfigTextValue":    EnvVarConfigTextValue,
		"WfStepInputDataData":      WfStepInputDataData,
	} {
		if !strings.Contains(doc, RuntimeReferenceNote) {
			t.Errorf("%s does not carry RuntimeReferenceNote, so a reader has no way to "+
				"find the reference syntax from this attribute", name)
		}
	}
}

// The note is a pointer, not a specification. Spelling a form out in an attribute
// description puts a second copy of the syntax where nothing checks it against the
// guide, which is exactly the drift this arrangement exists to prevent.
func TestRuntimeReferenceNoteNamesNoForms(t *testing.T) {
	if !strings.Contains(RuntimeReferenceNote, RuntimeReferencesGuide) {
		t.Fatal("RuntimeReferenceNote does not link the guide, so it points nowhere")
	}

	for _, form := range referenceForms {
		if strings.Contains(RuntimeReferenceNote, form) {
			t.Errorf("RuntimeReferenceNote spells out %q; the forms belong in the guide "+
				"so there is only one copy to keep correct", form)
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
		if !strings.Contains(doc, RuntimeReferencesGuide) {
			t.Errorf("%s does not link the Runtime References guide, so a reader is told "+
				"not to write a credential without being told what to write instead", name)
		}
	}
}

// terraform_version is asked about often enough that the description has to answer the
// two questions readers actually arrive with: which versions exist for each engine, and
// what supplies them. It previously explained the "TERRAFORM-"/"OPENTOFU-" prefix the API
// stores instead -- provider internals a reader can neither see nor act on.
func TestTerraformVersionDocumentsWhatIsAvailable(t *testing.T) {
	for _, engine := range []string{"OpenTofu", "Terraform"} {
		if !strings.Contains(TerraformVersion, engine) {
			t.Errorf("TerraformVersion does not say which versions %s supports", engine)
		}
	}

	// The stock container carries the open-source releases, so a BUSL-licensed
	// Terraform needs a runtime image of your own. Omitting that leaves a reader
	// guessing why a recent version is missing.
	if !strings.Contains(TerraformVersion, "runtime image") {
		t.Error("TerraformVersion does not mention that a custom runtime image is what " +
			"carries versions the stock workflow step does not")
	}

	for _, prefix := range []string{"TERRAFORM-", "OPENTOFU-"} {
		if strings.Contains(TerraformVersion, prefix) {
			t.Errorf("TerraformVersion documents the %s prefix, which is how the API "+
				"stores the value rather than anything a reader writes", prefix)
		}
	}
}
