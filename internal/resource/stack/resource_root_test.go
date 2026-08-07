package stack_test

import (
	"context"
	"fmt"
	"testing"

	sgsdkgo "github.com/StackGuardian/sg-sdk-go"
	"github.com/StackGuardian/sg-sdk-go/stacktemplaterevisions"
	"github.com/StackGuardian/sg-sdk-go/stacktemplates"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

// setupSecondStackTemplateRevision creates and publishes revision :2 of an
// existing stack template (already created by setupStackTemplateChain), with
// the given description and wired to the same workflow slot/template as
// revision :1. Registers cleanup. Returns the bare revision id ("<name>:2").
func setupSecondStackTemplateRevision(t *testing.T, stackTemplateID, workflowTemplateID, description string) string {
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
	applyAction := sgsdkgo.ActionEnumApply
	planAction := sgsdkgo.ActionEnumPlan

	_, err := client.StackTemplateRevisions.CreateStackTemplateRevision(
		context.TODO(), org, stackTemplateID,
		&stacktemplaterevisions.CreateStackTemplateRevisionRequest{
			Alias:            "v2",
			SourceConfigKind: &sourceConfigKind,
			IsPublic:         sgsdkgo.IsPublicEnumZero.Ptr(),
			OwnerOrg:         fmt.Sprintf("/orgs/%s", org),
			LongDescription:  &description,
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
						testWfSlotId: {Parameters: &sgsdkgo.StackActionParameters{TerraformAction: &sgsdkgo.TerraformAction{Action: &applyAction}}},
					},
				},
				"plan": {
					Name: "plan",
					Order: map[string]*sgsdkgo.ActionOrder{
						testWfSlotId: {Parameters: &sgsdkgo.StackActionParameters{TerraformAction: &sgsdkgo.TerraformAction{Action: &planAction}}},
					},
				},
			},
		},
	)
	if err != nil && !is409(err) {
		t.Fatalf("setupSecondStackTemplateRevision: create revision for %q: %s", stackTemplateID, err)
	}

	_, err = client.StackTemplateRevisions.UpdateStackTemplateRevision(
		context.TODO(), org, revisionID,
		&stacktemplaterevisions.UpdateStackTemplateRevisionRequest{
			IsPublic: sgsdkgo.Optional(sgsdkgo.IsPublicEnumOne),
		},
	)
	if err != nil {
		t.Fatalf("setupSecondStackTemplateRevision: publish revision %q: %s", revisionID, err)
	}

	return revisionID
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

// TestAccStack_TemplateGroupIdReResolution verifies that changing
// template_group_id to a different stack template revision re-resolves the
// Optional+Computed fields the user left unset (description here) against
// the NEW revision, rather than carrying the old revision's value forward
// via UseStateForUnknown (see ModifyPlan/reResolveOnRevisionChange).
//
// It also regression-tests a real "Provider produced inconsistent result
// after apply" risk in that same code path: revision2 (like revision1) only
// defines apply/plan, not destroy, so ToUpdateAPIModel's expandActionsMap
// generates "destroy" at apply time (fillGeneratedDefaultActions) — if
// reResolveOnRevisionChange naively set plan.DefaultActions to a KNOWN value
// containing only apply/plan, apply's actual result (including the generated
// destroy) wouldn't match what the plan promised, and Terraform would fail
// the whole step with that harness-level error rather than any assertion
// below even running. Step 2 succeeding at all, plus destroy actually
// appearing afterward, is the proof.
func TestAccStack_TemplateGroupIdReResolution(t *testing.T) {
	wfGrpName := "tf-provider-stack-tmplswitch-wfgrp"
	wfTemplateName := "tf-provider-stack-tmplswitch-wftmpl"
	stackTemplateName := "tf-provider-stack-tmplswitch-stmpl"
	id := "tf-provider-stack-tmplswitch"

	revision1 := setupStackDependencyChain(t, wfGrpName, wfTemplateName, stackTemplateName, id)
	revision2 := setupSecondStackTemplateRevision(t, stackTemplateName, wfTemplateName, "revision two description")

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader()),
		Steps: []resource.TestStep{
			{
				// description left unset — revision1 has none, so it stays empty.
				Config: testAccStackConfig(wfGrpName, revision1, id, ""),
				Check:  resource.TestCheckResourceAttr("stackguardian_stack.test", "template_group_id", revision1),
			},
			{
				// Switch to revision2 — description must re-resolve to revision2's
				// value, not stay stuck at revision1's (empty).
				Config: testAccStackConfig(wfGrpName, revision2, id, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_stack.test", "template_group_id", revision2),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "description", "revision two description"),
					// Neither revision defines "destroy" — generated at apply time.
					resource.TestCheckResourceAttr("stackguardian_stack.test", "default_actions.destroy.name", "Destroy"),
				),
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
