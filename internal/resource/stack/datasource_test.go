package stack_test

import (
	"fmt"
	"strings"
	"testing"

	sgsdkgo "github.com/StackGuardian/sg-sdk-go"
	"github.com/StackGuardian/sg-sdk-go/stacktemplaterevisions"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

// Data source tests live in the resource package to reuse its fixtures. The
// data source's own package covers a stack with every attribute set explicitly
// (TestAccStackDataSource_Basic); these cover values the stack doesn't declare.

// testAccStackDataSourceConfig appends a stackguardian_stack data source that
// reads the stack declared by stackConfig.
func testAccStackDataSourceConfig(stackConfig string) string {
	return stackConfig + `
data "stackguardian_stack" "test" {
  id                = stackguardian_stack.test.id
  workflow_group_id = stackguardian_stack.test.workflow_group_id
}
`
}

// checkDataSourceMatchesResource checks the data source holds exactly the
// attributes, with the same values, as the resource's state.
func checkDataSourceMatchesResource(ds, res string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		dsState, ok := s.RootModule().Resources[ds]
		if !ok {
			return fmt.Errorf("%s not found in state", ds)
		}
		resState, ok := s.RootModule().Resources[res]
		if !ok {
			return fmt.Errorf("%s not found in state", res)
		}
		want, got := resState.Primary.Attributes, dsState.Primary.Attributes

		var diffs []string
		for k, v := range want {
			if gv, ok := got[k]; !ok {
				diffs = append(diffs, fmt.Sprintf("%s: missing, want %q", k, v))
			} else if gv != v {
				diffs = append(diffs, fmt.Sprintf("%s: got %q, want %q", k, gv, v))
			}
		}
		for k, v := range got {
			if _, ok := want[k]; !ok {
				diffs = append(diffs, fmt.Sprintf("%s: unexpected %q", k, v))
			}
		}
		if len(diffs) > 0 {
			return fmt.Errorf("%s differs from %s:\n%s", ds, res, strings.Join(diffs, "\n"))
		}
		return nil
	}
}

// TestAccStackDataSource_InheritedValues verifies the data source returns the values a
// stack inherits from its stack template revision, not only the ones it declares. The
// stack declares only its workflow id; description, tags, context_tags, the workflow's
// terraform_config and vcs_config, and the apply/plan actions all come from the revision,
// whose workflow also sets an explicitly empty iac_input_data ("No Inputs").
//
// The data source must match the resource's state attribute for attribute, and the
// inherited values are also checked explicitly so a regression in both shows up.
func TestAccStackDataSource_InheritedValues(t *testing.T) {
	acctest.SkipUnlessAcceptance(t)
	wfGrpName := acctest.ResourceName("ds-stack-inherit-wfgrp")
	wfTemplateName := acctest.ResourceName("ds-stack-inherit-wftmpl")
	stackTemplateName := acctest.ResourceName("ds-stack-inherit-stmpl")
	id := acctest.ResourceName("ds-stack-inherit")

	t.Cleanup(func() {
		logCleanupErr(t, fmt.Sprintf("delete workflow group %q", wfGrpName), deleteWorkflowGroupFixture(wfGrpName))
	})
	if err := createWorkflowGroupFixture(wfGrpName); err != nil && !is409(err) {
		t.Fatalf("TestAccStackDataSource_InheritedValues: create workflow group %q: %s", wfGrpName, err)
	}
	workflowTemplateID := setupStackWorkflowTemplate(t, wfTemplateName)
	description := "inherited description"
	revision := setupStackTemplateChainWithWorkflow(t, stackTemplateName, workflowTemplateID,
		&description, []string{"inherited-tag"}, map[string]string{"team": "inherited"},
		func(wf *stacktemplaterevisions.StackTemplateRevisionWorkflow) {
			// The fixture's VcsConfig is non-nil (it carries IacVcsConfig).
			wf.VcsConfig.IacInputData = &sgsdkgo.IacInputData{}
		})
	t.Cleanup(func() { deleteStackFixture(wfGrpName, id) })

	const res, ds = "stackguardian_stack.test", "data.stackguardian_stack.test"
	wf := "workflows_config.workflows.0."

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader()),
		Steps: []resource.TestStep{
			{
				Config: testAccStackDataSourceConfig(testAccStackConfig(wfGrpName, revision, id, "")),
				Check: resource.ComposeAggregateTestCheckFunc(
					checkDataSourceMatchesResource(ds, res),

					resource.TestCheckResourceAttr(ds, "description", description),
					resource.TestCheckResourceAttr(ds, "tags.#", "1"),
					resource.TestCheckResourceAttr(ds, "tags.0", "inherited-tag"),
					resource.TestCheckResourceAttr(ds, "context_tags.team", "inherited"),

					resource.TestCheckResourceAttr(ds, wf+"terraform_config.terraform_version", "1.5.7"),
					resource.TestCheckResourceAttr(ds, wf+"terraform_config.managed_terraform_state", "true"),
					resource.TestCheckResourceAttr(ds, wf+"vcs_config.iac_vcs_config.iac_template_id",
						fmt.Sprintf("/%s/%s:1", org, workflowTemplateID)),
					resource.TestCheckNoResourceAttr(ds, wf+"vcs_config.iac_input_data.schema_type"),
					resource.TestCheckNoResourceAttr(ds, wf+"vcs_config.iac_input_data.data"),
					resource.TestCheckResourceAttrSet(ds, wf+"workflow_id"),

					resource.TestCheckResourceAttr(ds, "actions.%", "2"),
					resource.TestCheckResourceAttr(ds, "actions.apply.default", "true"),
					resource.TestCheckResourceAttr(ds, "actions.plan.default", "true"),
				),
			},
		},
	})
}

// TestAccStackDataSource_ReferenceValues pins how the data source returns ${reference::}
// values. The platform stores them rewritten to ${workflow::...} when it creates the
// workflows. The resource keeps the declared ${reference::} form because it compares
// against its prior state (preserveReferenceForms). The data source has no prior state,
// so it returns the platform's ${workflow::} form.
//
// The second workflow of the two-workflow revision references the first one's output.
func TestAccStackDataSource_ReferenceValues(t *testing.T) {
	acctest.SkipUnlessAcceptance(t)
	wfGrpName := acctest.ResourceName("ds-stack-ref-wfgrp")
	wfTemplateName := acctest.ResourceName("ds-stack-ref-wftmpl")
	stackTemplateName := acctest.ResourceName("ds-stack-ref-stmpl")
	id := acctest.ResourceName("ds-stack-ref")

	t.Cleanup(func() {
		logCleanupErr(t, fmt.Sprintf("delete workflow group %q", wfGrpName), deleteWorkflowGroupFixture(wfGrpName))
	})
	if err := createWorkflowGroupFixture(wfGrpName); err != nil && !is409(err) {
		t.Fatalf("TestAccStackDataSource_ReferenceValues: create workflow group %q: %s", wfGrpName, err)
	}
	workflowTemplateID := setupStackWorkflowTemplate(t, wfTemplateName)
	revision := setupStackTemplateChainNoActions(t, stackTemplateName, workflowTemplateID)
	t.Cleanup(func() { deleteStackFixture(wfGrpName, id) })

	const reference = "${reference::template-group.0.outputs.bucket_region}"
	// $${ escapes HCL interpolation, so Terraform sees a literal ${reference::...}.
	stackConfig := testAccStackConfig(wfGrpName, revision, id,
		referenceWorkflowsConfig(`"$${reference::template-group.0.outputs.bucket_region}"`))

	const res, ds = "stackguardian_stack.test", "data.stackguardian_stack.test"
	wf := "workflows_config.workflows.1."

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader()),
		Steps: []resource.TestStep{
			{
				Config: testAccStackDataSourceConfig(stackConfig),
				Check: resource.ComposeAggregateTestCheckFunc(
					// The resource keeps the declared form.
					resource.TestCheckResourceAttr(res, wf+"vcs_config.iac_input_data.data",
						fmt.Sprintf(`{"bucket_region":%q}`, reference)),
					resource.TestCheckResourceAttr(res, wf+"environment_variables.0.config.text_value", reference),

					// The data source returns the platform's rewritten form.
					resource.TestCheckResourceAttrWith(ds, wf+"vcs_config.iac_input_data.data", func(v string) error {
						if !strings.Contains(v, `"bucket_region":"${workflow::`) {
							return fmt.Errorf(`got %q, want bucket_region in the ${workflow::} form`, v)
						}
						return nil
					}),
					resource.TestCheckResourceAttrWith(ds, wf+"environment_variables.0.config.text_value", func(v string) error {
						if !strings.HasPrefix(v, "${workflow::") {
							return fmt.Errorf(`got %q, want the ${workflow::} form`, v)
						}
						return nil
					}),
					resource.TestCheckResourceAttr(ds, wf+"vcs_config.iac_input_data.schema_type", "RAW_JSON"),
					resource.TestCheckResourceAttr(ds, wf+"environment_variables.0.config.var_name", "BUCKET_REGION"),
				),
			},
		},
	})
}
