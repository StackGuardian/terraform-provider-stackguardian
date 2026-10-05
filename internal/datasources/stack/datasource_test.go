package stackdatasource_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"regexp"
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

// workflowUUID is the id of the single workflow the stack template revision fixture declares.
const workflowUUID = "d8dfaf15-2ad9-da29-8af0-c6b288b12089"

func customHeader() http.Header {
	h := http.Header{}
	h.Set("x-sg-internal-auth-orgid", "sg-provider-test")
	return h
}

func getClient() *sgclient.Client {
	return sgclient.NewClient(
		sgoption.WithApiKey(fmt.Sprintf("apikey %s", os.Getenv("STACKGUARDIAN_API_KEY"))),
		sgoption.WithBaseURL(os.Getenv("STACKGUARDIAN_API_URI")),
		sgoption.WithHTTPHeader(customHeader()),
	)
}

func is409(err error) bool {
	var apiErr *core.APIError
	return errors.As(err, &apiErr) && apiErr.StatusCode == 409
}

func logCleanupErr(t *testing.T, action string, err error) {
	if err != nil {
		t.Logf("cleanup: %s: %s", action, err)
	}
}

// deprecationNow is an effective date already in the past, so a published revision is deprecated
// immediately (and so becomes deletable) rather than only scheduled to be.
func deprecationNow() *string {
	d := fmt.Sprintf("%d", time.Now().Add(-1*time.Second).Unix())
	return &d
}

// setupStackPrerequisites creates, via the SDK, everything a stack needs to exist: a workflow
// group, a published workflow template revision, and a published stack template revision that
// declares one workflow (workflowUUID) built from that workflow template. Cleanup is registered
// before each create, so a failure partway through still tears down what was made; t.Cleanup is
// LIFO, so the stack goes first and the workflow group last. Returns the stack template revision
// id in the bare "<name>:1" form a stack's template_group_id takes.
func setupStackPrerequisites(t *testing.T, wfGrpName, wfTemplateName, stackTemplateName, stackId string) string {
	t.Helper()
	ctx := context.TODO()
	client := getClient()
	ownerOrg := fmt.Sprintf("/orgs/%s", org)

	// Workflow group.
	t.Cleanup(func() {
		_, err := client.WorkflowGroups.DeleteWorkflowGroup(ctx, org, wfGrpName)
		logCleanupErr(t, fmt.Sprintf("delete workflow group %q", wfGrpName), err)
	})
	if _, err := client.WorkflowGroups.CreateWorkflowGroup(ctx, org, &sgsdkgo.WorkflowGroup{ResourceName: &wfGrpName}); err != nil && !is409(err) {
		t.Fatalf("create workflow group %q: %s", wfGrpName, err)
	}

	// Workflow template + revision :1, published.
	wfRevisionId := wfTemplateName + ":1"
	t.Cleanup(func() {
		msg := "Test cleanup"
		_, err := client.WorkflowTemplatesRevisions.UpdateWorkflowTemplateRevision(ctx, org, wfRevisionId,
			&workflowtemplaterevisions.UpdateWorkflowTemplateRevisionRequest{
				Deprecation: sgsdkgo.Optional(workflowtemplaterevisions.Deprecation{EffectiveDate: deprecationNow(), Message: &msg}),
			})
		logCleanupErr(t, fmt.Sprintf("deprecate workflow template revision %q", wfRevisionId), err)
		logCleanupErr(t, fmt.Sprintf("delete workflow template revision %q", wfRevisionId),
			client.WorkflowTemplatesRevisions.DeleteWorkflowTemplateRevision(ctx, org, wfRevisionId, true))
		logCleanupErr(t, fmt.Sprintf("delete workflow template %q", wfTemplateName),
			client.WorkflowTemplates.DeleteWorkflowTemplate(ctx, org, wfTemplateName))
	})
	wfSourceKind := workflowtemplates.WorkflowTemplateSourceConfigKindTerraform
	if _, err := client.WorkflowTemplates.CreateWorkflowTemplate(ctx, org, false, &workflowtemplates.CreateWorkflowTemplateRequest{
		Id:               &wfTemplateName,
		TemplateName:     wfTemplateName,
		SourceConfigKind: &wfSourceKind,
		TemplateType:     sgsdkgo.TemplateTypeEnum("IAC"),
		IsPublic:         sgsdkgo.IsPublicEnumZero.Ptr(),
		OwnerOrg:         ownerOrg,
	}); err != nil && !is409(err) {
		t.Fatalf("create workflow template %q: %s", wfTemplateName, err)
	}
	wfTfVersion := "1.5.0"
	if _, err := client.WorkflowTemplatesRevisions.CreateWorkflowTemplateRevision(ctx, org, wfTemplateName, &workflowtemplaterevisions.CreateWorkflowTemplateRevisionsRequest{
		Alias:            "v1",
		SourceConfigKind: &wfSourceKind,
		IsPublic:         sgsdkgo.IsPublicEnumZero.Ptr(),
		OwnerOrg:         ownerOrg,
		TerraformConfig:  &sgsdkgo.TerraformConfig{TerraformVersion: &wfTfVersion},
	}); err != nil && !is409(err) {
		t.Fatalf("create workflow template revision %q: %s", wfRevisionId, err)
	}
	if _, err := client.WorkflowTemplatesRevisions.UpdateWorkflowTemplateRevision(ctx, org, wfRevisionId,
		&workflowtemplaterevisions.UpdateWorkflowTemplateRevisionRequest{IsPublic: sgsdkgo.Optional(sgsdkgo.IsPublicEnumOne)}); err != nil {
		t.Fatalf("publish workflow template revision %q: %s", wfRevisionId, err)
	}
	if _, err := client.WorkflowTemplates.UpdateWorkflowTemplate(ctx, org, wfTemplateName,
		&workflowtemplates.UpdateWorkflowTemplateRequest{IsPublic: sgsdkgo.Optional(sgsdkgo.IsPublicEnumOne)}); err != nil {
		t.Fatalf("publish workflow template %q: %s", wfTemplateName, err)
	}

	// Stack template + revision :1 declaring one workflow, published.
	stackRevisionId := stackTemplateName + ":1"
	t.Cleanup(func() {
		msg := "Test cleanup"
		_, err := client.StackTemplateRevisions.UpdateStackTemplateRevision(ctx, org, stackRevisionId,
			&stacktemplaterevisions.UpdateStackTemplateRevisionRequest{
				Deprecation: sgsdkgo.Optional(stacktemplaterevisions.Deprecation{EffectiveDate: deprecationNow(), Message: &msg}),
			})
		logCleanupErr(t, fmt.Sprintf("deprecate stack template revision %q", stackRevisionId), err)
		logCleanupErr(t, fmt.Sprintf("delete stack template revision %q", stackRevisionId),
			client.StackTemplateRevisions.DeleteStackTemplateRevision(ctx, org, stackRevisionId, true))
		logCleanupErr(t, fmt.Sprintf("delete stack template %q", stackTemplateName),
			client.StackTemplates.DeleteStackTemplate(ctx, org, stackTemplateName))
	})
	stackSourceKind := stacktemplates.StackTemplateSourceConfigKindTerraform
	if _, err := client.StackTemplates.CreateStackTemplate(ctx, org, false, &stacktemplates.CreateStackTemplateRequest{
		Id:               &stackTemplateName,
		TemplateName:     stackTemplateName,
		SourceConfigKind: &stackSourceKind,
		OwnerOrg:         ownerOrg,
	}); err != nil && !is409(err) {
		t.Fatalf("create stack template %q: %s", stackTemplateName, err)
	}
	prefixedWfTemplateId := fmt.Sprintf("/%s/%s", org, wfTemplateName)
	prefixedWfRevisionId := fmt.Sprintf("/%s/%s", org, wfRevisionId)
	useMarketplace := true
	managedState := true
	stackTfVersion := "1.5.7"
	applyAction := sgsdkgo.ActionEnumApply
	planAction := sgsdkgo.ActionEnumPlan
	if _, err := client.StackTemplateRevisions.CreateStackTemplateRevision(ctx, org, stackTemplateName, &stacktemplaterevisions.CreateStackTemplateRevisionRequest{
		Alias:            "v1",
		SourceConfigKind: &stackSourceKind,
		IsPublic:         sgsdkgo.IsPublicEnumZero.Ptr(),
		OwnerOrg:         ownerOrg,
		WorkflowsConfig: &stacktemplaterevisions.StackTemplateRevisionWorkflowsConfig{
			Workflows: []*stacktemplaterevisions.StackTemplateRevisionWorkflow{
				{
					Id:           sgsdkgo.String(workflowUUID),
					TemplateId:   &prefixedWfTemplateId,
					ResourceName: sgsdkgo.String("wf-1"),
					VcsConfig: &sgsdkgo.VcsConfig{
						IacVcsConfig: &sgsdkgo.IacvcsConfig{
							UseMarketplaceTemplate: &useMarketplace,
							IacTemplateId:          &prefixedWfRevisionId,
						},
					},
					TerraformConfig: &sgsdkgo.TerraformConfig{
						ManagedTerraformState: &managedState,
						TerraformVersion:      &stackTfVersion,
					},
				},
			},
		},
		// A stack template revision can't be published without actions.
		Actions: map[string]*sgsdkgo.Actions{
			"apply": {Name: "apply", Order: map[string]*sgsdkgo.ActionOrder{
				workflowUUID: {Parameters: &sgsdkgo.StackActionParameters{TerraformAction: &sgsdkgo.TerraformAction{Action: &applyAction}}},
			}},
			"plan": {Name: "plan", Order: map[string]*sgsdkgo.ActionOrder{
				workflowUUID: {Parameters: &sgsdkgo.StackActionParameters{TerraformAction: &sgsdkgo.TerraformAction{Action: &planAction}}},
			}},
		},
	}); err != nil && !is409(err) {
		t.Fatalf("create stack template revision %q: %s", stackRevisionId, err)
	}
	if _, err := client.StackTemplateRevisions.UpdateStackTemplateRevision(ctx, org, stackRevisionId,
		&stacktemplaterevisions.UpdateStackTemplateRevisionRequest{IsPublic: sgsdkgo.Optional(sgsdkgo.IsPublicEnumOne)}); err != nil {
		t.Fatalf("publish stack template revision %q: %s", stackRevisionId, err)
	}
	if _, err := client.StackTemplates.UpdateStackTemplate(ctx, org, stackTemplateName,
		&stacktemplates.UpdateStackTemplateRequest{IsPublic: sgsdkgo.Optional(sgsdkgo.IsPublicEnumOne)}); err != nil {
		t.Fatalf("publish stack template %q: %s", stackTemplateName, err)
	}

	// Safety net only — Terraform's own destroy normally deletes the stack first.
	t.Cleanup(func() {
		_, _ = client.Stacks.DeleteStack(ctx, org, stackId, wfGrpName, &sgsdkgo.DeleteStackRequest{ForceDelete: sgsdkgo.Bool(true)})
	})

	return stackRevisionId
}

// TestAccStackDataSource_Basic is an integration test: it creates a stack with a broad set of
// attributes, reads it back through the data source, and checks every attribute the data source
// returns against the value the stack was created with.
func TestAccStackDataSource_Basic(t *testing.T) {
	wfGrpName := acctest.ResourceName("ds-stack-wfgrp")
	wfTemplateName := acctest.ResourceName("ds-stack-wftmpl")
	stackTemplateName := acctest.ResourceName("ds-stack-stmpl")
	id := acctest.ResourceName("ds-stack")

	revision := setupStackPrerequisites(t, wfGrpName, wfTemplateName, stackTemplateName, id)

	config := fmt.Sprintf(`
resource "stackguardian_stack" "test" {
  id                = %[1]q
  workflow_group_id = %[2]q
  template_group_id = %[3]q

  description = "stack read by the data source"
  tags        = ["ds-tag"]

  context_tags = {
    team = "platform"
  }

  workflows_config = {
    workflows = [
      {
        id        = %[4]q
        tags      = ["ds-workflow-tag"]
        approvers = ["alice@example.com"]

        number_of_approvals_required = 1

        terraform_config = {
          terraform_version = "1.5.5"
        }

        environment_variables = [
          {
            kind = "PLAIN_TEXT"
            config = {
              var_name   = "DS_VAR"
              text_value = "ds-value"
            }
          }
        ]

        user_schedules = [
          {
            cron  = "0 8 ? * MON *"
            state = "ENABLED"
          }
        ]

        context_tags = {
          env = "dev"
        }

        runner_constraints = {
          type  = "private"
          names = ["runner-1"]
        }

        mini_steps = {
          webhooks = {
            completed = [
              {
                webhook_name = "on-completed"
                webhook_url  = "https://example.com/hook"
              }
            ]
          }
        }
      }
    ]
  }

  actions = {
    apply = {
      name        = "apply"
      description = "Custom apply action"
      order = {
        %[4]q = {
          parameters = {
            terraform_action = {
              action = "apply"
            }
          }
        }
      }
    }
    destroy = {
      name = "destroy"
      order = {
        %[4]q = {
          parameters = {
            terraform_action = {
              action = "destroy"
            }
          }
        }
      }
    }
  }
}

data "stackguardian_stack" "test" {
  id                = stackguardian_stack.test.id
  workflow_group_id = stackguardian_stack.test.workflow_group_id
}
`, id, wfGrpName, revision, workflowUUID)

	const res, ds = "stackguardian_stack.test", "data.stackguardian_stack.test"
	wf := "workflows_config.workflows.0."
	applyOrder := fmt.Sprintf("actions.apply.order.%s.", workflowUUID)
	destroyOrder := fmt.Sprintf("actions.destroy.order.%s.", workflowUUID)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader()),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Stack-level attributes.
					resource.TestCheckResourceAttr(ds, "id", id),
					resource.TestCheckResourceAttr(ds, "workflow_group_id", wfGrpName),
					resource.TestCheckResourceAttr(ds, "template_group_id", revision),
					resource.TestCheckResourceAttr(ds, "description", "stack read by the data source"),
					resource.TestCheckResourceAttr(ds, "tags.#", "1"),
					resource.TestCheckResourceAttr(ds, "tags.0", "ds-tag"),
					resource.TestCheckResourceAttr(ds, "context_tags.%", "1"),
					resource.TestCheckResourceAttr(ds, "context_tags.team", "platform"),
					// Not set in config — computed, so compared against the created stack.
					resource.TestCheckResourceAttrPair(ds, "resource_name", res, "resource_name"),

					// The workflow.
					resource.TestCheckResourceAttr(ds, "workflows_config.workflows.#", "1"),
					resource.TestCheckResourceAttr(ds, wf+"id", workflowUUID),
					resource.TestCheckResourceAttr(ds, wf+"tags.#", "1"),
					resource.TestCheckResourceAttr(ds, wf+"tags.0", "ds-workflow-tag"),
					resource.TestCheckResourceAttr(ds, wf+"approvers.#", "1"),
					resource.TestCheckResourceAttr(ds, wf+"approvers.0", "alice@example.com"),
					resource.TestCheckResourceAttr(ds, wf+"number_of_approvals_required", "1"),
					resource.TestCheckResourceAttr(ds, wf+"terraform_config.terraform_version", "1.5.5"),
					resource.TestCheckResourceAttr(ds, wf+"environment_variables.#", "1"),
					resource.TestCheckResourceAttr(ds, wf+"environment_variables.0.kind", "PLAIN_TEXT"),
					resource.TestCheckResourceAttr(ds, wf+"environment_variables.0.config.var_name", "DS_VAR"),
					resource.TestCheckResourceAttr(ds, wf+"environment_variables.0.config.text_value", "ds-value"),
					resource.TestCheckResourceAttr(ds, wf+"user_schedules.#", "1"),
					resource.TestCheckResourceAttr(ds, wf+"user_schedules.0.cron", "0 8 ? * MON *"),
					resource.TestCheckResourceAttr(ds, wf+"user_schedules.0.state", "ENABLED"),
					resource.TestCheckResourceAttr(ds, wf+"context_tags.env", "dev"),
					resource.TestCheckResourceAttr(ds, wf+"runner_constraints.type", "private"),
					resource.TestCheckResourceAttr(ds, wf+"runner_constraints.names.#", "1"),
					resource.TestCheckResourceAttr(ds, wf+"runner_constraints.names.0", "runner-1"),
					resource.TestCheckResourceAttr(ds, wf+"mini_steps.webhooks.completed.#", "1"),
					resource.TestCheckResourceAttr(ds, wf+"mini_steps.webhooks.completed.0.webhook_name", "on-completed"),
					resource.TestCheckResourceAttr(ds, wf+"mini_steps.webhooks.completed.0.webhook_url", "https://example.com/hook"),
					// Computed by the provider, so compared against the created stack.
					resource.TestCheckResourceAttrPair(ds, wf+"workflow_id", res, wf+"workflow_id"),

					// Actions, keyed by the workflow's UUID exactly as declared.
					resource.TestCheckResourceAttr(ds, "actions.%", "2"),
					resource.TestCheckResourceAttr(ds, "actions.apply.name", "apply"),
					resource.TestCheckResourceAttr(ds, "actions.apply.description", "Custom apply action"),
					resource.TestCheckResourceAttr(ds, "actions.apply.order.%", "1"),
					resource.TestCheckResourceAttr(ds, applyOrder+"parameters.terraform_action.action", "apply"),
					resource.TestCheckResourceAttr(ds, "actions.destroy.name", "destroy"),
					resource.TestCheckResourceAttr(ds, "actions.destroy.order.%", "1"),
					resource.TestCheckResourceAttr(ds, destroyOrder+"parameters.terraform_action.action", "destroy"),
				),
			},
		},
	})
}

// TestAccStackDataSource_NotFound verifies that looking up a stack that doesn't exist fails
// instead of returning an empty result.
func TestAccStackDataSource_NotFound(t *testing.T) {
	wfGrpName := acctest.ResourceName("ds-stack-missing-wfgrp")
	client := getClient()

	t.Cleanup(func() {
		_, err := client.WorkflowGroups.DeleteWorkflowGroup(context.TODO(), org, wfGrpName)
		logCleanupErr(t, fmt.Sprintf("delete workflow group %q", wfGrpName), err)
	})
	if _, err := client.WorkflowGroups.CreateWorkflowGroup(context.TODO(), org, &sgsdkgo.WorkflowGroup{ResourceName: &wfGrpName}); err != nil && !is409(err) {
		t.Fatalf("create workflow group %q: %s", wfGrpName, err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader()),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
data "stackguardian_stack" "missing" {
  id                = %q
  workflow_group_id = %q
}
`, acctest.ResourceName("ds-stack-does-not-exist"), wfGrpName),
				ExpectError: regexp.MustCompile("Unable to read stack"),
			},
		},
	})
}
