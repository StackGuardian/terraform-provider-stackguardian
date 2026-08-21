package stack_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
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

// setupStackTemplateChain creates and publishes a stack template + revision
// :1 via the SDK, with workflows_config wiring testWfSlotId to
// workflowTemplateID. Registers cleanup. Returns the bare revision id
// ("<name>:1") for use as a stack's template_group_id, which is stored bare
// in state and only gets the "/<org>/" wire prefix on send (see ToAPIModel).
// Publishes directly with no staged unpublished step — a stack can't
// reference an unpublished revision, but publishing itself has no such
// restriction.
func setupStackTemplateChain(t *testing.T, stackTemplateID, workflowTemplateID string) string {
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
		t.Fatalf("setupStackTemplateChain: create stack template %q: %s", stackTemplateID, err)
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
		t.Fatalf("setupStackTemplateChain: create revision for %q: %s", stackTemplateID, err)
	}

	_, err = client.StackTemplateRevisions.UpdateStackTemplateRevision(
		context.TODO(), org, revisionID,
		&stacktemplaterevisions.UpdateStackTemplateRevisionRequest{
			IsPublic: sgsdkgo.Optional(sgsdkgo.IsPublicEnumOne),
		},
	)
	if err != nil {
		t.Fatalf("setupStackTemplateChain: publish revision %q: %s", revisionID, err)
	}

	_, err = client.StackTemplates.UpdateStackTemplate(
		context.TODO(), org, stackTemplateID,
		&stacktemplates.UpdateStackTemplateRequest{
			IsPublic: sgsdkgo.Optional(sgsdkgo.IsPublicEnumOne),
		},
	)
	if err != nil {
		t.Fatalf("setupStackTemplateChain: publish template %q: %s", stackTemplateID, err)
	}

	return revisionID
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
func testAccStackConfig(wfGrpName, stackTemplateRevisionID, id, additionalConfig string) string {
	return fmt.Sprintf(`
resource "stackguardian_stack" "test" {
  workflow_group_id = %q
  id                 = %q
  template_group_id = %q

  %s
}
`, wfGrpName, id, stackTemplateRevisionID, additionalConfig)
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

// --- Remaining cases (not yet implemented) ---
//
// Covered, split across resource_root_test.go /
// resource_actions_test.go / resource_workflows_test.go:
//   - id RequiresReplace; template_group_id round trip + re-resolution on
//     revision change; Read removes state on 404; Delete treats 404 as
//     success (resource.go's isStackNotFound, added alongside its test).
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
