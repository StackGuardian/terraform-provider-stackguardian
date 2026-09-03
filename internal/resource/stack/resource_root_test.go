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

// defaultSecondRevisionActions returns the fixed apply/plan Actions map used
// by setupSecondStackTemplateRevision's default revision :2, and reused
// directly by the per-attribute "retained when template has none" tests
// (resource_tags_test.go, resource_context_tags_test.go) that need a
// revision :2 with actions of its own alongside the one field under test.
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

// TestAccStack_TemplateGroupIdReResolution — REVISION SWITCH test.
// Purpose: description and actions are left unset in config, so both start resolved from
// revision1; template_group_id then moves to revision2, which defines its OWN description and
// actions. Both fields must re-resolve to revision2's values, not stay stuck on revision1's (see
// ModifyPlan/reResolveOnRevisionChange).
//
// Mechanism this guards, for actions specifically: revision1 and revision2 both define their
// own apply/plan Actions verbatim (see setupStackTemplateChain/setupSecondStackTemplateRevision),
// so reResolveOnRevisionChange's actions block (which only re-resolves when the NEW revision has
// its own Actions — see expandActionsMap for why the provider never synthesizes a fresh action
// set) picks up revision2's set directly. Step 2's actions assertions confirm that. There is no
// "switching to a revision with no actions of its own" case to test separately — the API rejects
// publishing a stack template revision with an empty Actions map, so every revision a stack can
// reference always has at least one action (see resource_actions_test.go's "actions resolution
// rules" comment).
func TestAccStack_TemplateGroupIdReResolution(t *testing.T) {
	wfGrpName := "tf-provider-stack-tmplswitch-wfgrp"
	wfTemplateName := "tf-provider-stack-tmplswitch-wftmpl"
	stackTemplateName := "tf-provider-stack-tmplswitch-stmpl"
	id := "tf-provider-stack-tmplswitch"

	revision1 := setupStackDependencyChain(t, wfGrpName, wfTemplateName, stackTemplateName, id)
	revision2 := setupSecondStackTemplateRevision(t, stackTemplateName, wfTemplateName, "revision two description", nil)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader()),
		Steps: []resource.TestStep{
			{
				// description left unset — revision1 has none, so it stays empty.
				// actions left unset too — revision1's own apply/plan are inherited
				// verbatim at CREATE time (rule 2 in resource_actions_test.go's
				// "actions resolution rules" comment; Step 2 below re-proves it
				// across a revision switch instead).
				Config: testAccStackConfig(wfGrpName, revision1, id, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_stack.test", "template_group_id", revision1),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "actions.apply.name", "apply"),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "actions.plan.name", "plan"),
				),
			},
			{
				// Switch to revision2 — description must re-resolve to revision2's
				// value, not stay stuck at revision1's (empty).
				Config: testAccStackConfig(wfGrpName, revision2, id, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_stack.test", "template_group_id", revision2),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "description", "revision two description"),
					// revision2's own Actions, copied verbatim (no generation).
					resource.TestCheckResourceAttr("stackguardian_stack.test", "actions.apply.name", "apply"),
					resource.TestCheckResourceAttr("stackguardian_stack.test", "actions.plan.name", "plan"),
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
