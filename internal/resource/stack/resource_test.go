package stack_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	sgsdkgo "github.com/StackGuardian/sg-sdk-go"
	sgclient "github.com/StackGuardian/sg-sdk-go/client"
	"github.com/StackGuardian/sg-sdk-go/core"
	sgoption "github.com/StackGuardian/sg-sdk-go/option"
	"github.com/StackGuardian/sg-sdk-go/stacktemplaterevisions"
	"github.com/StackGuardian/sg-sdk-go/stacktemplates"
	"github.com/StackGuardian/sg-sdk-go/workflowtemplaterevisions"
	"github.com/StackGuardian/sg-sdk-go/workflowtemplates"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

var org = os.Getenv("STACKGUARDIAN_ORG_NAME")

// testWfSlotId is the shared workflow slot id used in the stack template
// revision's workflows_config and Actions (see setupStackTemplateChain).
const testWfSlotId = "d8dfaf15-2ad9-da29-8af0-c6b288b12089"

// secondWfSlotId is a second workflow slot UUID, distinct from testWfSlotId,
// shared package-wide by every test that needs a multi-slot stack template
// revision (setupStackTemplateChainNoActions / setupSecondStackTemplateRevisionTwoSlots).
const secondWfSlotId = "3f7c9e2a-5b1d-4e6f-8a2c-9d4b6e1f0a3c"

func getClient() *sgclient.Client {
	return sgclient.NewClient(
		sgoption.WithApiKey(fmt.Sprintf("apikey %s", os.Getenv("STACKGUARDIAN_API_KEY"))),
		sgoption.WithBaseURL(os.Getenv("STACKGUARDIAN_API_URI")),
		sgoption.WithHTTPHeader(customHeader()),
	)
}

func customHeader() http.Header {
	h := http.Header{}
	h.Set("x-sg-internal-auth-orgid", "sg-provider-test")
	return h
}

// --- Fixture setup & cleanup (via SDK) ---
//
// A stack's prerequisites (workflow group, workflow template + revision,
// stack template + revision) are created directly via the SDK rather than as
// Terraform resources: a published template revision must be deprecated
// before it can be deleted, and Terraform's destroy has no way to do that.
// Only the stack itself is Terraform-managed. Cleanup mirrors
// workflow_from_template's setupWorkflowTemplate: deprecate -> delete
// revision -> delete template, best-effort throughout.

func is409(err error) bool {
	var apiErr *core.APIError
	return errors.As(err, &apiErr) && apiErr.StatusCode == 409
}

// logCleanupErr logs a fixture cleanup failure instead of discarding it — a
// swallowed failure here only ever surfaced later, as a confusing 409 on some
// unrelated run reusing the same deterministic fixture name.
func logCleanupErr(t *testing.T, action string, err error) {
	if err != nil {
		t.Logf("cleanup: %s: %s", action, err)
	}
}

func createWorkflowGroupFixture(wfGrpName string) error {
	client := getClient()
	_, err := client.WorkflowGroups.CreateWorkflowGroup(context.TODO(), org, &sgsdkgo.WorkflowGroup{
		ResourceName: &wfGrpName,
	})
	return err
}

func deleteWorkflowGroupFixture(wfGrpName string) error {
	client := getClient()
	_, err := client.WorkflowGroups.DeleteWorkflowGroup(context.TODO(), org, wfGrpName)
	return err
}

func deprecateWorkflowTemplateRevisionFixture(revisionId string) error {
	client := getClient()
	// Must already be in the past, or the revision is only scheduled to
	// deprecate — still actively published (and thus undeletable) until then.
	// This fixture always runs well after the revision was created, so a
	// second back is enough margin without risking landing before creation.
	effectiveDate := fmt.Sprintf("%d", time.Now().Add(-1*time.Second).Unix())
	message := "Test cleanup"
	_, err := client.WorkflowTemplatesRevisions.UpdateWorkflowTemplateRevision(
		context.TODO(), org, revisionId,
		&workflowtemplaterevisions.UpdateWorkflowTemplateRevisionRequest{
			Deprecation: sgsdkgo.Optional(workflowtemplaterevisions.Deprecation{
				EffectiveDate: &effectiveDate,
				Message:       &message,
			}),
		},
	)
	return err
}

func deleteWorkflowTemplateRevisionFixture(revisionId string) error {
	client := getClient()
	return client.WorkflowTemplatesRevisions.DeleteWorkflowTemplateRevision(context.TODO(), org, revisionId, true)
}

func deleteWorkflowTemplateFixture(templateId string) error {
	client := getClient()
	return client.WorkflowTemplates.DeleteWorkflowTemplate(context.TODO(), org, templateId)
}

func deprecateStackTemplateRevisionFixture(revisionId string) error {
	client := getClient()
	// Must already be in the past — see deprecateWorkflowTemplateRevisionFixture.
	effectiveDate := fmt.Sprintf("%d", time.Now().Add(-1*time.Second).Unix())
	message := "Test cleanup"
	_, err := client.StackTemplateRevisions.UpdateStackTemplateRevision(
		context.TODO(), org, revisionId,
		&stacktemplaterevisions.UpdateStackTemplateRevisionRequest{
			Deprecation: sgsdkgo.Optional(stacktemplaterevisions.Deprecation{
				EffectiveDate: &effectiveDate,
				Message:       &message,
			}),
		},
	)
	return err
}

func deleteStackTemplateRevisionFixture(revisionId string) error {
	client := getClient()
	return client.StackTemplateRevisions.DeleteStackTemplateRevision(context.TODO(), org, revisionId, true)
}

func deleteStackTemplateFixture(templateId string) error {
	client := getClient()
	return client.StackTemplates.DeleteStackTemplate(context.TODO(), org, templateId)
}

func deleteStackFixture(wfGrpName, id string) error {
	client := getClient()
	// ForceDelete also removes the workflows inside the stack, so the
	// workflow group is actually empty by the time its own delete runs.
	_, err := client.Stacks.DeleteStack(context.TODO(), org, id, wfGrpName, &sgsdkgo.DeleteStackRequest{
		ForceDelete: sgsdkgo.Bool(true),
	})
	return err
}

// setupStackWorkflowTemplate creates and publishes a workflow template +
// revision :1 via the SDK (mirrors workflow_from_template's own
// setupWorkflowTemplate). Registers cleanup. Returns the bare template id.
func setupStackWorkflowTemplate(t *testing.T, templateID string) string {
	t.Helper()
	client := getClient()
	revisionID := fmt.Sprintf("%s:1", templateID)
	sourceConfigKind := workflowtemplates.WorkflowTemplateSourceConfigKindTerraform

	// Registered before any create/publish call below, so a t.Fatalf or panic
	// partway through still leaves cleanup registered for whatever did make
	// it to the server (registering only after everything succeeds would
	// leak on any failure in between).
	t.Cleanup(func() {
		logCleanupErr(t, fmt.Sprintf("deprecate workflow template revision %q", revisionID), deprecateWorkflowTemplateRevisionFixture(revisionID))
		logCleanupErr(t, fmt.Sprintf("delete workflow template revision %q", revisionID), deleteWorkflowTemplateRevisionFixture(revisionID))
		logCleanupErr(t, fmt.Sprintf("delete workflow template %q", templateID), deleteWorkflowTemplateFixture(templateID))
	})

	_, err := client.WorkflowTemplates.CreateWorkflowTemplate(
		context.TODO(), org, false,
		&workflowtemplates.CreateWorkflowTemplateRequest{
			Id:               &templateID,
			TemplateName:     templateID,
			SourceConfigKind: &sourceConfigKind,
			TemplateType:     sgsdkgo.TemplateTypeEnum("IAC"),
			IsPublic:         sgsdkgo.IsPublicEnumZero.Ptr(),
			OwnerOrg:         fmt.Sprintf("/orgs/%s", org),
		},
	)
	if err != nil && !is409(err) {
		t.Fatalf("setupStackWorkflowTemplate: create template %q: %s", templateID, err)
	}

	alias := "v1"
	tfVersion := "1.5.0"
	_, err = client.WorkflowTemplatesRevisions.CreateWorkflowTemplateRevision(
		context.TODO(), org, templateID,
		&workflowtemplaterevisions.CreateWorkflowTemplateRevisionsRequest{
			Alias:            alias,
			SourceConfigKind: &sourceConfigKind,
			IsPublic:         sgsdkgo.IsPublicEnumZero.Ptr(),
			OwnerOrg:         fmt.Sprintf("/orgs/%s", org),
			TerraformConfig: &sgsdkgo.TerraformConfig{
				TerraformVersion: &tfVersion,
			},
		},
	)
	if err != nil && !is409(err) {
		t.Fatalf("setupStackWorkflowTemplate: create revision for %q: %s", templateID, err)
	}

	_, err = client.WorkflowTemplatesRevisions.UpdateWorkflowTemplateRevision(
		context.TODO(), org, revisionID,
		&workflowtemplaterevisions.UpdateWorkflowTemplateRevisionRequest{
			IsPublic: sgsdkgo.Optional(sgsdkgo.IsPublicEnumOne),
		},
	)
	if err != nil {
		t.Fatalf("setupStackWorkflowTemplate: publish revision %q: %s", revisionID, err)
	}

	_, err = client.WorkflowTemplates.UpdateWorkflowTemplate(
		context.TODO(), org, templateID,
		&workflowtemplates.UpdateWorkflowTemplateRequest{
			IsPublic: sgsdkgo.Optional(sgsdkgo.IsPublicEnumOne),
		},
	)
	if err != nil {
		t.Fatalf("setupStackWorkflowTemplate: publish template %q: %s", templateID, err)
	}

	return templateID
}

// setupStackTemplateChainWithFields is setupStackTemplateChain's general
// form: it additionally sets description/tags/contextTags on the revision
// when the corresponding argument is non-nil/non-empty. Used by the
// per-attribute resolution-scenario tests (resource_attributes_test.go) to produce a
// template revision that DOES supply a value for the one attribute under
// test. setupStackTemplateChain itself is unchanged (still the zero-value
// case) and just delegates here. Publishes directly with no staged
// unpublished step — a stack can't reference an unpublished revision, but
// publishing itself has no such restriction.
func setupStackTemplateChainWithFields(t *testing.T, stackTemplateID, workflowTemplateID string, description *string, tags []string, contextTags map[string]string) string {
	t.Helper()
	client := getClient()
	revisionID := fmt.Sprintf("%s:1", stackTemplateID)
	sourceConfigKind := stacktemplates.StackTemplateSourceConfigKindTerraform

	// Registered before any create/publish call below — see
	// setupStackWorkflowTemplate for why.
	t.Cleanup(func() {
		logCleanupErr(t, fmt.Sprintf("deprecate stack template revision %q", revisionID), deprecateStackTemplateRevisionFixture(revisionID))
		logCleanupErr(t, fmt.Sprintf("delete stack template revision %q", revisionID), deleteStackTemplateRevisionFixture(revisionID))
		logCleanupErr(t, fmt.Sprintf("delete stack template %q", stackTemplateID), deleteStackTemplateFixture(stackTemplateID))
	})

	_, err := client.StackTemplates.CreateStackTemplate(
		context.TODO(), org, false,
		&stacktemplates.CreateStackTemplateRequest{
			Id:               &stackTemplateID,
			TemplateName:     stackTemplateID,
			SourceConfigKind: &sourceConfigKind,
			OwnerOrg:         fmt.Sprintf("/orgs/%s", org),
		},
	)
	if err != nil && !is409(err) {
		t.Fatalf("setupStackTemplateChainWithFields: create stack template %q: %s", stackTemplateID, err)
	}

	// template_id/iac_template_id must be fully org-qualified on the wire;
	// replicate the prefixing stack_template_revision's model.go does
	// internally, since we're bypassing that model here.
	alias := "v1"
	useMarketplace := true
	managedState := true
	tfVersion := "1.5.7"
	prefixedWorkflowTemplateID := fmt.Sprintf("/%s/%s", org, workflowTemplateID)
	prefixedWorkflowRevisionID := fmt.Sprintf("/%s/%s:1", org, workflowTemplateID)

	// apply/plan actions so a plain stack (no actions of its own — that has
	// its own dedicated test) can inherit them via the template fallback;
	// CreateStack requires both.
	applyAction := sgsdkgo.ActionEnumApply
	planAction := sgsdkgo.ActionEnumPlan

	_, err = client.StackTemplateRevisions.CreateStackTemplateRevision(
		context.TODO(), org, stackTemplateID,
		&stacktemplaterevisions.CreateStackTemplateRevisionRequest{
			Alias:            alias,
			SourceConfigKind: &sourceConfigKind,
			IsPublic:         sgsdkgo.IsPublicEnumZero.Ptr(),
			OwnerOrg:         fmt.Sprintf("/orgs/%s", org),
			LongDescription:  description,
			Tags:             tags,
			ContextTags:      contextTags,
			WorkflowsConfig: &stacktemplaterevisions.StackTemplateRevisionWorkflowsConfig{
				Workflows: []*stacktemplaterevisions.StackTemplateRevisionWorkflow{
					{
						Id:           sgsdkgo.String(testWfSlotId),
						TemplateId:   &prefixedWorkflowTemplateID,
						ResourceName: sgsdkgo.String("wf-1"),
						VcsConfig: &sgsdkgo.VcsConfig{
							IacVcsConfig: &sgsdkgo.IacvcsConfig{
								UseMarketplaceTemplate: &useMarketplace,
								IacTemplateId:          &prefixedWorkflowRevisionID,
							},
						},
						TerraformConfig: &sgsdkgo.TerraformConfig{
							ManagedTerraformState: &managedState,
							TerraformVersion:      &tfVersion,
						},
					},
				},
			},
			Actions: map[string]*sgsdkgo.Actions{
				"apply": {
					Name: "apply",
					Order: map[string]*sgsdkgo.ActionOrder{
						testWfSlotId: {
							Parameters: &sgsdkgo.StackActionParameters{
								TerraformAction: &sgsdkgo.TerraformAction{Action: &applyAction},
							},
						},
					},
				},
				"plan": {
					Name: "plan",
					Order: map[string]*sgsdkgo.ActionOrder{
						testWfSlotId: {
							Parameters: &sgsdkgo.StackActionParameters{
								TerraformAction: &sgsdkgo.TerraformAction{Action: &planAction},
							},
						},
					},
				},
			},
		},
	)
	if err != nil && !is409(err) {
		t.Fatalf("setupStackTemplateChainWithFields: create revision for %q: %s", stackTemplateID, err)
	}

	_, err = client.StackTemplateRevisions.UpdateStackTemplateRevision(
		context.TODO(), org, revisionID,
		&stacktemplaterevisions.UpdateStackTemplateRevisionRequest{
			IsPublic: sgsdkgo.Optional(sgsdkgo.IsPublicEnumOne),
		},
	)
	if err != nil {
		t.Fatalf("setupStackTemplateChainWithFields: publish revision %q: %s", revisionID, err)
	}

	_, err = client.StackTemplates.UpdateStackTemplate(
		context.TODO(), org, stackTemplateID,
		&stacktemplates.UpdateStackTemplateRequest{
			IsPublic: sgsdkgo.Optional(sgsdkgo.IsPublicEnumOne),
		},
	)
	if err != nil {
		t.Fatalf("setupStackTemplateChainWithFields: publish template %q: %s", stackTemplateID, err)
	}

	return revisionID
}

// setupStackTemplateChain creates and publishes a stack template + revision
// :1 via the SDK, with workflows_config wiring testWfSlotId to
// workflowTemplateID, and none of description/tags/contextTags set. Returns
// the bare revision id ("<name>:1") for use as a stack's template_group_id,
// which is stored bare in state and only gets the "/<org>/" wire prefix on
// send (see ToAPIModel).
func setupStackTemplateChain(t *testing.T, stackTemplateID, workflowTemplateID string) string {
	t.Helper()
	return setupStackTemplateChainWithFields(t, stackTemplateID, workflowTemplateID, nil, nil, nil)
}

// setupStackDependencyChain creates the full SDK-fixture prerequisite chain
// for a stack (workflow group, workflow template + revision, stack template +
// revision) and registers cleanup in dependency order (t.Cleanup is LIFO):
// stack, then stack template chain, then workflow template chain, then
// workflow group last. Only the stack itself is left for the caller to
// create via Terraform. Returns the stack template revision id for
// template_group_id.
func setupStackDependencyChain(t *testing.T, wfGrpName, wfTemplateName, stackTemplateName, stackId string) string {
	t.Helper()

	// Registered before the create call — see setupStackWorkflowTemplate.
	t.Cleanup(func() {
		logCleanupErr(t, fmt.Sprintf("delete workflow group %q", wfGrpName), deleteWorkflowGroupFixture(wfGrpName))
	})
	// 409-tolerant: a leftover from an interrupted prior run bypasses
	// t.Cleanup entirely, so reuse it rather than blocking every future run
	// under this deterministic name.
	if err := createWorkflowGroupFixture(wfGrpName); err != nil && !is409(err) {
		t.Fatalf("setupStackDependencyChain: create workflow group %q: %s", wfGrpName, err)
	}

	workflowTemplateID := setupStackWorkflowTemplate(t, wfTemplateName)
	stackTemplateRevisionID := setupStackTemplateChain(t, stackTemplateName, workflowTemplateID)

	// Registered last (after the chain, so t.Cleanup's LIFO order runs this
	// FIRST — deleting the stack before the template it references, and
	// before the workflow group it lives in). The stack itself isn't created
	// by this function at all (the caller creates it later via Terraform), so
	// there's no fallible operation here to register ahead of. This is purely
	// a safety net — Terraform's own destroy is what's supposed to delete the
	// stack, so an error here (typically just a 404 for the normal case where
	// destroy already succeeded) isn't logged like the other cleanups.
	t.Cleanup(func() {
		deleteStackFixture(wfGrpName, stackId)
	})

	return stackTemplateRevisionID
}

// --- Terraform config generator ---

// testAccStackConfig returns config for the stack resource alone; its
// prerequisites are SDK fixtures (setupStackDependencyChain), not Terraform
// resources. No actions here — that has its own test; apply/plan come from
// the stack template revision instead. additionalConfig is inserted verbatim
// into the resource body.
//
// workflows_config is Required and must declare exactly the workflow slots
// the referenced revision defines, in order (validateWorkflowsConfigMatchesRevision).
// Every fixture used by additionalConfig-only callers here
// (setupStackDependencyChain/setupStackTemplateChain and
// setupSecondStackTemplateRevision) wires exactly one slot, testWfSlotId, so
// that's injected automatically unless additionalConfig already declares its
// own workflows_config — needed by callers against a multi-slot fixture
// (setupStackTemplateChainNoActions) or a workflows_config value under test.
func testAccStackConfig(wfGrpName, stackTemplateRevisionID, id, additionalConfig string) string {
	if !strings.Contains(additionalConfig, "workflows_config") {
		additionalConfig = fmt.Sprintf(`
  workflows_config = {
    workflows = [
      { id = %q }
    ]
  }

  %s`, testWfSlotId, additionalConfig)
	}
	return fmt.Sprintf(`
resource "stackguardian_stack" "test" {
  workflow_group_id = %q
  id                 = %q
  template_group_id = %q

  %s
}
`, wfGrpName, id, stackTemplateRevisionID, additionalConfig)
}

// setupStackTemplateChainNoActions creates and publishes a stack template +
// revision :1 via the SDK, like setupStackTemplateChain, but wires TWO
// workflow slots (testWfSlotId, secondWfSlotId — both pointing at the same
// workflow template) instead of one, and defines no Actions of its own at
// all. Used by tests that need a template supplying neither apply/plan/destroy
// NOR a dependency chain to inherit, so the only source for them is the API's
// own create-time default (the provider no longer synthesizes one itself —
// see expandActionsMap), and by tests that need a second workflow slot to
// exercise workflows_config against a multi-slot revision.
// Registers cleanup. Returns the bare revision id ("<name>:1").
func setupStackTemplateChainNoActions(t *testing.T, stackTemplateID, workflowTemplateID string) string {
	t.Helper()
	client := getClient()
	revisionID := fmt.Sprintf("%s:1", stackTemplateID)
	sourceConfigKind := stacktemplates.StackTemplateSourceConfigKindTerraform

	t.Cleanup(func() {
		logCleanupErr(t, fmt.Sprintf("deprecate stack template revision %q", revisionID), deprecateStackTemplateRevisionFixture(revisionID))
		logCleanupErr(t, fmt.Sprintf("delete stack template revision %q", revisionID), deleteStackTemplateRevisionFixture(revisionID))
		logCleanupErr(t, fmt.Sprintf("delete stack template %q", stackTemplateID), deleteStackTemplateFixture(stackTemplateID))
	})

	_, err := client.StackTemplates.CreateStackTemplate(
		context.TODO(), org, false,
		&stacktemplates.CreateStackTemplateRequest{
			Id:               &stackTemplateID,
			TemplateName:     stackTemplateID,
			SourceConfigKind: &sourceConfigKind,
			OwnerOrg:         fmt.Sprintf("/orgs/%s", org),
		},
	)
	if err != nil && !is409(err) {
		t.Fatalf("setupStackTemplateChainNoActions: create stack template %q: %s", stackTemplateID, err)
	}

	prefixedWorkflowTemplateID := fmt.Sprintf("/%s/%s", org, workflowTemplateID)
	prefixedWorkflowRevisionID := fmt.Sprintf("/%s/%s:1", org, workflowTemplateID)
	useMarketplace := true
	managedState := true
	tfVersion := "1.5.7"

	makeSlot := func(slotId, resourceName string) *stacktemplaterevisions.StackTemplateRevisionWorkflow {
		return &stacktemplaterevisions.StackTemplateRevisionWorkflow{
			Id:           sgsdkgo.String(slotId),
			TemplateId:   &prefixedWorkflowTemplateID,
			ResourceName: sgsdkgo.String(resourceName),
			VcsConfig: &sgsdkgo.VcsConfig{
				IacVcsConfig: &sgsdkgo.IacvcsConfig{
					UseMarketplaceTemplate: &useMarketplace,
					IacTemplateId:          &prefixedWorkflowRevisionID,
				},
			},
			TerraformConfig: &sgsdkgo.TerraformConfig{
				ManagedTerraformState: &managedState,
				TerraformVersion:      &tfVersion,
			},
		}
	}

	_, err = client.StackTemplateRevisions.CreateStackTemplateRevision(
		context.TODO(), org, stackTemplateID,
		&stacktemplaterevisions.CreateStackTemplateRevisionRequest{
			Alias:            "v1",
			SourceConfigKind: &sourceConfigKind,
			IsPublic:         sgsdkgo.IsPublicEnumZero.Ptr(),
			OwnerOrg:         fmt.Sprintf("/orgs/%s", org),
			WorkflowsConfig: &stacktemplaterevisions.StackTemplateRevisionWorkflowsConfig{
				Workflows: []*stacktemplaterevisions.StackTemplateRevisionWorkflow{
					makeSlot(testWfSlotId, "wf-1"),
					makeSlot(secondWfSlotId, "wf-2"),
				},
			},
			// Deliberately no Actions — the template supplies none of its own, so
			// apply/plan/destroy can only come from the API's own create-time default.
		},
	)
	if err != nil && !is409(err) {
		t.Fatalf("setupStackTemplateChainNoActions: create revision for %q: %s", stackTemplateID, err)
	}

	_, err = client.StackTemplateRevisions.UpdateStackTemplateRevision(
		context.TODO(), org, revisionID,
		&stacktemplaterevisions.UpdateStackTemplateRevisionRequest{
			IsPublic: sgsdkgo.Optional(sgsdkgo.IsPublicEnumOne),
		},
	)
	if err != nil {
		t.Fatalf("setupStackTemplateChainNoActions: publish revision %q: %s", revisionID, err)
	}

	_, err = client.StackTemplates.UpdateStackTemplate(
		context.TODO(), org, stackTemplateID,
		&stacktemplates.UpdateStackTemplateRequest{
			IsPublic: sgsdkgo.Optional(sgsdkgo.IsPublicEnumOne),
		},
	)
	if err != nil {
		t.Fatalf("setupStackTemplateChainNoActions: publish template %q: %s", stackTemplateID, err)
	}

	return revisionID
}

// setupSecondStackTemplateRevisionWithFields is setupSecondStackTemplateRevision's general
// form: additionally sets tags/contextTags/actions on revision :2. actions must be non-empty —
// the API rejects publishing a stack template revision with an empty Actions map ("Stack actions
// are empty"). setupSecondStackTemplateRevision delegates here with tags/contextTags nil and its
// own fixed apply/plan Actions map.
func setupSecondStackTemplateRevisionWithFields(t *testing.T, stackTemplateID, workflowTemplateID, description string, numberOfApprovalsRequired *int, tags []string, contextTags map[string]string, actions map[string]*sgsdkgo.Actions) string {
	t.Helper()
	client := getClient()
	revisionID := fmt.Sprintf("%s:2", stackTemplateID)
	sourceConfigKind := stacktemplates.StackTemplateSourceConfigKindTerraform

	t.Cleanup(func() {
		logCleanupErr(t, fmt.Sprintf("deprecate stack template revision %q", revisionID), deprecateStackTemplateRevisionFixture(revisionID))
		logCleanupErr(t, fmt.Sprintf("delete stack template revision %q", revisionID), deleteStackTemplateRevisionFixture(revisionID))
	})

	prefixedWorkflowTemplateID := fmt.Sprintf("/%s/%s", org, workflowTemplateID)
	prefixedWorkflowRevisionID := fmt.Sprintf("/%s/%s:1", org, workflowTemplateID)
	useMarketplace := true
	managedState := true
	tfVersion := "1.5.7"

	_, err := client.StackTemplateRevisions.CreateStackTemplateRevision(
		context.TODO(), org, stackTemplateID,
		&stacktemplaterevisions.CreateStackTemplateRevisionRequest{
			Alias:            "v2",
			SourceConfigKind: &sourceConfigKind,
			IsPublic:         sgsdkgo.IsPublicEnumZero.Ptr(),
			OwnerOrg:         fmt.Sprintf("/orgs/%s", org),
			LongDescription:  &description,
			Tags:             tags,
			ContextTags:      contextTags,
			WorkflowsConfig: &stacktemplaterevisions.StackTemplateRevisionWorkflowsConfig{
				Workflows: []*stacktemplaterevisions.StackTemplateRevisionWorkflow{
					{
						Id:                        sgsdkgo.String(testWfSlotId),
						TemplateId:                &prefixedWorkflowTemplateID,
						ResourceName:              sgsdkgo.String("wf-1"),
						NumberOfApprovalsRequired: numberOfApprovalsRequired,
						VcsConfig: &sgsdkgo.VcsConfig{
							IacVcsConfig: &sgsdkgo.IacvcsConfig{
								UseMarketplaceTemplate: &useMarketplace,
								IacTemplateId:          &prefixedWorkflowRevisionID,
							},
						},
						TerraformConfig: &sgsdkgo.TerraformConfig{
							ManagedTerraformState: &managedState,
							TerraformVersion:      &tfVersion,
						},
					},
				},
			},
			Actions: actions,
		},
	)
	if err != nil && !is409(err) {
		t.Fatalf("setupSecondStackTemplateRevisionWithFields: create revision for %q: %s", stackTemplateID, err)
	}

	_, err = client.StackTemplateRevisions.UpdateStackTemplateRevision(
		context.TODO(), org, revisionID,
		&stacktemplaterevisions.UpdateStackTemplateRevisionRequest{
			IsPublic: sgsdkgo.Optional(sgsdkgo.IsPublicEnumOne),
		},
	)
	if err != nil {
		t.Fatalf("setupSecondStackTemplateRevisionWithFields: publish revision %q: %s", revisionID, err)
	}

	return revisionID
}

// setupSecondStackTemplateRevision creates and publishes revision :2 of an
// existing stack template (already created by setupStackTemplateChain), with
// the given description and wired to the same workflow slot/template as
// revision :1, and its own fixed apply/plan Actions verbatim.
// numberOfApprovalsRequired, if non-nil, is set on that workflow slot — used
// to test workflows_config's revision-based re-resolution
// (reResolveWorkflowsConfigOnRevisionChange), since revision :1 never sets it.
// Registers cleanup. Returns the bare revision id ("<name>:2").
func setupSecondStackTemplateRevision(t *testing.T, stackTemplateID, workflowTemplateID, description string, numberOfApprovalsRequired *int) string {
	t.Helper()
	return setupSecondStackTemplateRevisionWithFields(t, stackTemplateID, workflowTemplateID, description, numberOfApprovalsRequired, nil, nil, defaultSecondRevisionActions())
}

// setupSecondStackTemplateRevisionTwoSlots creates and publishes revision :2
// of an existing stack template (already created by setupStackTemplateChain /
// setupStackDependencyChain, whose revision :1 wires only testWfSlotId),
// adding a second workflow slot (secondWfSlotId) alongside it — both pointing
// at the same workflow template. Used by
// TestAccStack_WorkflowsConfigAddSecondWorkflowOverride to exercise
// workflows_config growing across a genuine revision change: declaring an
// extra slot workflows_config against a STATIC revision is no longer
// possible (see validateWorkflowsConfigMatchesRevision), so growth can now
// only come from the referenced revision itself defining more slots.
// Registers cleanup. Returns the bare revision id ("<name>:2").
func setupSecondStackTemplateRevisionTwoSlots(t *testing.T, stackTemplateID, workflowTemplateID string) string {
	t.Helper()
	client := getClient()
	revisionID := fmt.Sprintf("%s:2", stackTemplateID)
	sourceConfigKind := stacktemplates.StackTemplateSourceConfigKindTerraform

	t.Cleanup(func() {
		logCleanupErr(t, fmt.Sprintf("deprecate stack template revision %q", revisionID), deprecateStackTemplateRevisionFixture(revisionID))
		logCleanupErr(t, fmt.Sprintf("delete stack template revision %q", revisionID), deleteStackTemplateRevisionFixture(revisionID))
	})

	prefixedWorkflowTemplateID := fmt.Sprintf("/%s/%s", org, workflowTemplateID)
	prefixedWorkflowRevisionID := fmt.Sprintf("/%s/%s:1", org, workflowTemplateID)
	useMarketplace := true
	managedState := true
	tfVersion := "1.5.7"

	makeSlot := func(slotId, resourceName string) *stacktemplaterevisions.StackTemplateRevisionWorkflow {
		return &stacktemplaterevisions.StackTemplateRevisionWorkflow{
			Id:           sgsdkgo.String(slotId),
			TemplateId:   &prefixedWorkflowTemplateID,
			ResourceName: sgsdkgo.String(resourceName),
			VcsConfig: &sgsdkgo.VcsConfig{
				IacVcsConfig: &sgsdkgo.IacvcsConfig{
					UseMarketplaceTemplate: &useMarketplace,
					IacTemplateId:          &prefixedWorkflowRevisionID,
				},
			},
			TerraformConfig: &sgsdkgo.TerraformConfig{
				ManagedTerraformState: &managedState,
				TerraformVersion:      &tfVersion,
			},
		}
	}

	_, err := client.StackTemplateRevisions.CreateStackTemplateRevision(
		context.TODO(), org, stackTemplateID,
		&stacktemplaterevisions.CreateStackTemplateRevisionRequest{
			Alias:            "v2",
			SourceConfigKind: &sourceConfigKind,
			IsPublic:         sgsdkgo.IsPublicEnumZero.Ptr(),
			OwnerOrg:         fmt.Sprintf("/orgs/%s", org),
			WorkflowsConfig: &stacktemplaterevisions.StackTemplateRevisionWorkflowsConfig{
				Workflows: []*stacktemplaterevisions.StackTemplateRevisionWorkflow{
					makeSlot(testWfSlotId, "wf-1"),
					makeSlot(secondWfSlotId, "wf-2"),
				},
			},
			Actions: defaultSecondRevisionActions(),
		},
	)
	if err != nil && !is409(err) {
		t.Fatalf("setupSecondStackTemplateRevisionTwoSlots: create revision for %q: %s", stackTemplateID, err)
	}

	_, err = client.StackTemplateRevisions.UpdateStackTemplateRevision(
		context.TODO(), org, revisionID,
		&stacktemplaterevisions.UpdateStackTemplateRevisionRequest{
			IsPublic: sgsdkgo.Optional(sgsdkgo.IsPublicEnumOne),
		},
	)
	if err != nil {
		t.Fatalf("setupSecondStackTemplateRevisionTwoSlots: publish revision %q: %s", revisionID, err)
	}

	return revisionID
}

// defaultSecondRevisionActions returns the fixed apply/plan Actions map used
// by setupSecondStackTemplateRevision's default revision :2, and reused
// directly by the per-attribute "retained when template has none" tests
// (resource_attributes_stack_upgrade_test.go) that need a revision :2 with
// actions of its own alongside the one field under test.
func defaultSecondRevisionActions() map[string]*sgsdkgo.Actions {
	applyAction := sgsdkgo.ActionEnumApply
	planAction := sgsdkgo.ActionEnumPlan
	return map[string]*sgsdkgo.Actions{
		"apply": {
			Name: "apply",
			Order: map[string]*sgsdkgo.ActionOrder{
				testWfSlotId: {Parameters: &sgsdkgo.StackActionParameters{TerraformAction: &sgsdkgo.TerraformAction{Action: &applyAction}}},
			},
		},
		"plan": {
			Name: "plan",
			Order: map[string]*sgsdkgo.ActionOrder{
				testWfSlotId: {Parameters: &sgsdkgo.StackActionParameters{TerraformAction: &sgsdkgo.TerraformAction{Action: &planAction}}},
			},
		},
	}
}

// --- Tests ---

// TestAccStack_Basic covers Create/Read/Update of a stack whose dependency
// chain is set up via SDK fixtures. Step 2 updates the stack only.
func TestAccStack_Basic(t *testing.T) {
	wfGrpName := "tf-provider-stack-basic-wfgrp"
	wfTemplateName := "tf-provider-stack-basic-wftmpl"
	stackTemplateName := "tf-provider-stack-basic-stmpl"
	id := "tf-provider-stack-basic"

	stackTemplateRevisionID := setupStackDependencyChain(t, wfGrpName, wfTemplateName, stackTemplateName, id)

	config := func(desc, tagVal, ctxVal string) string {
		return fmt.Sprintf(`
  description = %q
  tags        = [%q]

  context_tags = {
    env = %q
  }
`, desc, tagVal, ctxVal)
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader()),
		Steps: []resource.TestStep{
			{
				Config: testAccStackConfig(wfGrpName, stackTemplateRevisionID, id, config("first", "tag-a", "dev")),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_stack.test", "workflow_group_id", wfGrpName),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "template_group_id", stackTemplateRevisionID),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "id", id),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "description", "first"),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "tags.0", "tag-a"),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "context_tags.env", "dev"),
					// Optional+Computed and unset; server must have assigned one.
					resource.TestCheckResourceAttrSet("stackguardian_stack.test", "resource_name"),
				),
			},
			{
				// Update the stack only.
				Config: testAccStackConfig(wfGrpName, stackTemplateRevisionID, id, config("second", "tag-b", "prod")),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_stack.test", "description", "second"),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "tags.0", "tag-b"),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "context_tags.env", "prod"),
				),
			},
		},
	})
}

// TestAccStack_Import covers importing a stack via "workflow_group_id/id"
// and verifies the imported state matches what was just applied.
func TestAccStack_Import(t *testing.T) {
	wfGrpName := "tf-provider-stack-import-wfgrp"
	wfTemplateName := "tf-provider-stack-import-wftmpl"
	stackTemplateName := "tf-provider-stack-import-stmpl"
	id := "tf-provider-stack-import"

	stackTemplateRevisionID := setupStackDependencyChain(t, wfGrpName, wfTemplateName, stackTemplateName, id)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader()),
		Steps: []resource.TestStep{
			{
				Config: testAccStackConfig(wfGrpName, stackTemplateRevisionID, id, `description = "importable"`),
			},
			{
				ResourceName:      "stackguardian_stack.test",
				ImportState:       true,
				ImportStateId:     fmt.Sprintf("%s/%s", wfGrpName, id),
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccStack_IdRequiresReplace verifies that changing id forces a
// destroy-and-recreate (the SDK's PatchedStack has no Id field, so there's no
// other way to apply a change) rather than an in-place update, and that the
// new id is correctly applied afterward.
func TestAccStack_IdRequiresReplace(t *testing.T) {
	wfGrpName := "tf-provider-stack-idreplace-wfgrp"
	wfTemplateName := "tf-provider-stack-idreplace-wftmpl"
	stackTemplateName := "tf-provider-stack-idreplace-stmpl"
	id1 := "tf-provider-stack-idreplace-a"
	id2 := "tf-provider-stack-idreplace-b"

	revision := setupStackDependencyChain(t, wfGrpName, wfTemplateName, stackTemplateName, id1)
	// Safety net for the post-replace id too — the chain's own registration
	// only knows about id1.
	t.Cleanup(func() { deleteStackFixture(wfGrpName, id2) })

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader()),
		Steps: []resource.TestStep{
			{
				Config: testAccStackConfig(wfGrpName, revision, id1, ""),
				Check:  resource.TestCheckResourceAttr("stackguardian_stack.test", "id", id1),
			},
			{
				Config: testAccStackConfig(wfGrpName, revision, id2, ""),
				Check:  resource.TestCheckResourceAttr("stackguardian_stack.test", "id", id2),
			},
		},
	})
}

// TestAccStack_ReadRemovesOnNotFound verifies that a stack deleted
// out-of-band (API 404 on the next Read) is removed from Terraform state
// instead of erroring, leaving a non-empty plan (a pending create) since the
// config still declares the resource.
func TestAccStack_ReadRemovesOnNotFound(t *testing.T) {
	wfGrpName := "tf-provider-stack-read404-wfgrp"
	wfTemplateName := "tf-provider-stack-read404-wftmpl"
	stackTemplateName := "tf-provider-stack-read404-stmpl"
	id := "tf-provider-stack-read404"

	revision := setupStackDependencyChain(t, wfGrpName, wfTemplateName, stackTemplateName, id)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader()),
		Steps: []resource.TestStep{
			{
				Config: testAccStackConfig(wfGrpName, revision, id, ""),
			},
			{
				PreConfig: func() {
					if err := deleteStackFixture(wfGrpName, id); err != nil {
						t.Fatalf("failed to delete stack out-of-band: %s", err)
					}
				},
				RefreshState:       true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// TestAccStack_DeleteAlreadyGone verifies that deleting a stack that's
// already gone (API 404) is treated as a successful delete instead of
// erroring (isStackNotFound in resource.go), by deleting it out-of-band and
// then dropping it from Terraform config entirely so a real Delete() call is
// issued against an already-gone stack.
func TestAccStack_DeleteAlreadyGone(t *testing.T) {
	wfGrpName := "tf-provider-stack-del404-wfgrp"
	wfTemplateName := "tf-provider-stack-del404-wftmpl"
	stackTemplateName := "tf-provider-stack-del404-stmpl"
	id := "tf-provider-stack-del404"

	revision := setupStackDependencyChain(t, wfGrpName, wfTemplateName, stackTemplateName, id)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader()),
		Steps: []resource.TestStep{
			{
				Config: testAccStackConfig(wfGrpName, revision, id, ""),
			},
			{
				PreConfig: func() {
					if err := deleteStackFixture(wfGrpName, id); err != nil {
						t.Fatalf("failed to delete stack out-of-band: %s", err)
					}
				},
				// Resource removed from config entirely: Terraform issues a real
				// Delete() call against a stack that's already gone.
				Config: `# stack intentionally removed from config`,
			},
		},
	})
}

// --- Remaining cases (not yet implemented) ---
//
// Covered, split across resource_attributes_test.go /
// resource_attributes_stack_upgrade_test.go / resource_workflows_test.go:
//   - id RequiresReplace; template_group_id round trip + re-resolution on
//     revision change; Read removes state on 404; Delete treats 404 as
//     success (resource.go's isStackNotFound, added alongside its test —
//     all four above, in this file).
//   - workflows_config.workflows[]: minimal entry + Optional+Computed guard
//     regression; invalid wf_type/parallel_execution diagnostics;
//     three-way terraform_config precedence merge; vcs_config.iac_vcs_config
//     Computed-only rejection; approvers, user_schedules, context_tags,
//     runner_constraints, mini_steps round trip.
//   - actions: template fallback when unset (generated apply/plan/destroy,
//     and verbatim template Actions), wholesale override once the user
//     declares actions (see TestAccStack_ActionsGeneratedFromTemplate);
//     nested round trip (terraform_action, environment_variables,
//     dependencies) + removal on update (TestAccStack_ActionsRoundTrip);
//     dangling workflow reference rejected on a template_group_id change
//     (TestAccStack_ActionsRevisionRemovedWorkflow).
//
// Deferred — each needs infrastructure or live-API knowledge this session
// doesn't have:
//   - wf_steps_config round trip (top-level, per-workflow, and inside
//     actions[].order[].parameters), and actions'
//     deployment_platform_config: both need wf_step_template_id /
//     integration_id fixtures respectively (workflow_from_template's tests
//     have a setupWorkflowStepTemplate fixture that could be ported over).
// - vcs_config.iac_input_data: needs a schema_type value confirmed valid
//     against the live API to write a meaningful round trip. (The root-level
//     iac_input_data on workflows_config.workflows[] — TemplatesIacInputData
//     — was removed: it doesn't apply to stacks.)
// - Multiple workflows in workflows_config.workflows[]: needs a second
//     workflow slot registered on the stack template revision fixture (a
//     second workflow template + revision, or a second slot on the same
//     one) — setupStackTemplateChain only wires one slot currently.
// - updateWorkflowsFromConfig query param: not observable through
//     resource.Test's black-box testing (would need an HTTP-level
//     interceptor to inspect the actual request query string).
// - actions read/refresh reflecting an out-of-band API-side Actions change:
//     doable, but needs the raw PATCH payload to exactly reproduce the
//     existing "apply" action's shape alongside the injected one to avoid an
//     unrelated diff — deferred as the highest-effort-for-value item in this
//     group.
