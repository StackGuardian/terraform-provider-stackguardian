package workflowusingtemplate_test

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"

	sgsdkgo "github.com/StackGuardian/sg-sdk-go"
	sgclient "github.com/StackGuardian/sg-sdk-go/client"
	sgoption "github.com/StackGuardian/sg-sdk-go/option"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

var org = os.Getenv("STACKGUARDIAN_ORG_NAME")

func getClient() *sgclient.Client {
	customHeader := http.Header{}
	customHeader.Set("x-sg-internal-auth-orgid", "sg-provider-test")
	return sgclient.NewClient(
		sgoption.WithApiKey(fmt.Sprintf("apikey %s", os.Getenv("STACKGUARDIAN_API_KEY"))),
		sgoption.WithBaseURL(os.Getenv("STACKGUARDIAN_API_URI")),
		sgoption.WithHTTPHeader(customHeader),
	)
}

func createWorkflowGroupFixture(wfGrpName string) error {
	client := getClient()
	parts := strings.Split(wfGrpName, "/")
	leafName := parts[len(parts)-1]
	payload := &sgsdkgo.WorkflowGroup{
		ResourceName: &leafName,
	}
	if len(parts) > 1 {
		parent := strings.Join(parts[:len(parts)-1], "/")
		_, err := client.WorkflowGroups.CreateChildWorkflowGroup(context.TODO(), org, parent, payload)
		return err
	}
	_, err := client.WorkflowGroups.CreateWorkflowGroup(context.TODO(), org, payload)
	return err
}

func deleteWorkflowGroupFixture(wfGrpName string) {
	client := getClient()
	client.WorkflowGroups.DeleteWorkflowGroup(context.TODO(), org, wfGrpName)
}

func deleteWorkflowUsingTemplateFixture(wfGrpName, workflowName string) {
	client := getClient()
	client.Workflows.DeleteWorkflow(context.TODO(), org, workflowName, wfGrpName)
}

// wfTemplateID returns the workflow template ID from the environment or skips the test.
func wfTemplateID(t *testing.T) string {
	t.Helper()
	id := os.Getenv("STACKGUARDIAN_TEST_WF_TEMPLATE_ID")
	if id == "" {
		t.Skip("STACKGUARDIAN_TEST_WF_TEMPLATE_ID must be set for workflow_using_template acceptance tests")
	}
	return id
}

func customHeader() http.Header {
	h := http.Header{}
	h.Set("x-sg-internal-auth-orgid", "sg-provider-test")
	return h
}

// testAccWorkflowUsingTemplate builds a Terraform config for the resource.
func testAccWorkflowUsingTemplate(wfGrpName, id, wfType, templateID, additionalConfig string) string {
	return fmt.Sprintf(`
resource "stackguardian_workflow_using_template" "test" {
  workflow_group_id = %q
  id                = %q
  wf_type           = %q

  vcs_config = {
    iac_vcs_config = {
      iac_template_id = %q
    }
  }

  %s
}
`, wfGrpName, id, wfType, templateID, additionalConfig)
}

func TestAccWorkflowUsingTemplate_Basic(t *testing.T) {
	templateID := wfTemplateID(t)
	wfGrpName := "tf-provider-wf-template-basic-wfgrp"
	id := "tf-provider-wf-template-basic"

	if err := createWorkflowGroupFixture(wfGrpName); err != nil {
		t.Errorf("failed to create workflow group fixture: %s", err.Error())
	}
	defer deleteWorkflowGroupFixture(wfGrpName)
	defer deleteWorkflowUsingTemplateFixture(wfGrpName, id)

	config := fmt.Sprintf(`
  terraform_config = {
    terraform_version = %q
  }
`, "1.5.0")

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader()),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowUsingTemplate(wfGrpName, id, "TERRAFORM", templateID, config),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_using_template.test", "workflow_group_id", wfGrpName),
					resource.TestCheckResourceAttr("stackguardian_workflow_using_template.test", "id", id),
					resource.TestCheckResourceAttr("stackguardian_workflow_using_template.test", "wf_type", "TERRAFORM"),
					resource.TestCheckResourceAttr("stackguardian_workflow_using_template.test", "vcs_config.iac_vcs_config.iac_template_id", templateID),
					resource.TestCheckResourceAttrSet("stackguardian_workflow_using_template.test", "resolved_schema.resource_name"),
				),
			},
		},
	})
}

func TestAccWorkflowUsingTemplate_WithDescription(t *testing.T) {
	templateID := wfTemplateID(t)
	wfGrpName := "tf-provider-wf-template-desc-wfgrp"
	id := "tf-provider-wf-template-desc"

	if err := createWorkflowGroupFixture(wfGrpName); err != nil {
		t.Errorf("failed to create workflow group fixture: %s", err.Error())
	}
	defer deleteWorkflowGroupFixture(wfGrpName)
	defer deleteWorkflowUsingTemplateFixture(wfGrpName, id)

	config := func(description string) string {
		return fmt.Sprintf(`
  description = %q

  terraform_config = {
    terraform_version = %q
  }
`, description, "1.5.0")
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader()),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowUsingTemplate(wfGrpName, id, "TERRAFORM", templateID, config("initial description")),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_using_template.test", "id", id),
					resource.TestCheckResourceAttr("stackguardian_workflow_using_template.test", "description", "initial description"),
					resource.TestCheckResourceAttr("stackguardian_workflow_using_template.test", "resolved_schema.description", "initial description"),
				),
			},
			{
				Config: testAccWorkflowUsingTemplate(wfGrpName, id, "TERRAFORM", templateID, config("updated description")),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_using_template.test", "id", id),
					resource.TestCheckResourceAttr("stackguardian_workflow_using_template.test", "description", "updated description"),
					resource.TestCheckResourceAttr("stackguardian_workflow_using_template.test", "resolved_schema.description", "updated description"),
				),
			},
		},
	})
}

func TestAccWorkflowUsingTemplate_WithTerraformConfig(t *testing.T) {
	templateID := wfTemplateID(t)
	wfGrpName := "tf-provider-wf-template-tfcfg-wfgrp"
	id := "tf-provider-wf-template-tfcfg"

	if err := createWorkflowGroupFixture(wfGrpName); err != nil {
		t.Errorf("failed to create workflow group fixture: %s", err.Error())
	}
	defer deleteWorkflowGroupFixture(wfGrpName)
	defer deleteWorkflowUsingTemplateFixture(wfGrpName, id)

	config := func(tfVersion string) string {
		return fmt.Sprintf(`
  terraform_config = {
    terraform_version = %q
  }
`, tfVersion)
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader()),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowUsingTemplate(wfGrpName, id, "TERRAFORM", templateID, config("1.5.0")),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_using_template.test", "id", id),
					resource.TestCheckResourceAttr("stackguardian_workflow_using_template.test", "terraform_config.terraform_version", "1.5.0"),
					resource.TestCheckResourceAttr("stackguardian_workflow_using_template.test", "resolved_schema.terraform_config.terraform_version", "1.5.0"),
				),
			},
			{
				Config: testAccWorkflowUsingTemplate(wfGrpName, id, "TERRAFORM", templateID, config("1.6.0")),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_using_template.test", "id", id),
					resource.TestCheckResourceAttr("stackguardian_workflow_using_template.test", "terraform_config.terraform_version", "1.6.0"),
					resource.TestCheckResourceAttr("stackguardian_workflow_using_template.test", "resolved_schema.terraform_config.terraform_version", "1.6.0"),
				),
			},
		},
	})
}

func TestAccWorkflowUsingTemplate_WithEnvironmentVariables(t *testing.T) {
	templateID := wfTemplateID(t)
	wfGrpName := "tf-provider-wf-template-envvars-wfgrp"
	id := "tf-provider-wf-template-envvars"

	if err := createWorkflowGroupFixture(wfGrpName); err != nil {
		t.Errorf("failed to create workflow group fixture: %s", err.Error())
	}
	defer deleteWorkflowGroupFixture(wfGrpName)
	defer deleteWorkflowUsingTemplateFixture(wfGrpName, id)

	config := func(textValue string) string {
		return fmt.Sprintf(`
  terraform_config = {
    terraform_version = %q
  }

  environment_variables = [
    {
      kind = "PLAIN_TEXT"
      config = {
        var_name   = "MY_VAR"
        text_value = %q
      }
    }
  ]
`, "1.5.0", textValue)
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader()),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowUsingTemplate(wfGrpName, id, "TERRAFORM", templateID, config("initial-value")),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_using_template.test", "id", id),
					resource.TestCheckResourceAttr("stackguardian_workflow_using_template.test", "environment_variables.0.kind", "PLAIN_TEXT"),
					resource.TestCheckResourceAttr("stackguardian_workflow_using_template.test", "environment_variables.0.config.var_name", "MY_VAR"),
					resource.TestCheckResourceAttr("stackguardian_workflow_using_template.test", "environment_variables.0.config.text_value", "initial-value"),
				),
			},
			{
				Config: testAccWorkflowUsingTemplate(wfGrpName, id, "TERRAFORM", templateID, config("updated-value")),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_using_template.test", "id", id),
					resource.TestCheckResourceAttr("stackguardian_workflow_using_template.test", "environment_variables.0.config.text_value", "updated-value"),
				),
			},
		},
	})
}

func TestAccWorkflowUsingTemplate_WithUserSchedules(t *testing.T) {
	templateID := wfTemplateID(t)
	wfGrpName := "tf-provider-wf-template-schedules-wfgrp"
	id := "tf-provider-wf-template-schedules"

	if err := createWorkflowGroupFixture(wfGrpName); err != nil {
		t.Errorf("failed to create workflow group fixture: %s", err.Error())
	}
	defer deleteWorkflowGroupFixture(wfGrpName)
	defer deleteWorkflowUsingTemplateFixture(wfGrpName, id)

	config := func(cron string) string {
		return fmt.Sprintf(`
  terraform_config = {
    terraform_version = %q
  }

  user_schedules = [
    {
      cron  = %q
      state = "ENABLED"
      desc  = "Runs on schedule"
	  name  = "schedule_1"
    }
  ]
`, "1.5.0", cron)
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader()),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowUsingTemplate(wfGrpName, id, "TERRAFORM", templateID, config("0 8 ? * MON *")),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_using_template.test", "id", id),
					resource.TestCheckResourceAttr("stackguardian_workflow_using_template.test", "user_schedules.0.cron", "0 8 ? * MON *"),
					resource.TestCheckResourceAttr("stackguardian_workflow_using_template.test", "user_schedules.0.state", "ENABLED"),
				),
			},
			{
				Config: testAccWorkflowUsingTemplate(wfGrpName, id, "TERRAFORM", templateID, config("0 9 ? * MON *")),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_using_template.test", "id", id),
					resource.TestCheckResourceAttr("stackguardian_workflow_using_template.test", "user_schedules.0.cron", "0 9 ? * MON *"),
				),
			},
		},
	})
}

func TestAccWorkflowUsingTemplate_WithTagsAndContextTags(t *testing.T) {
	templateID := wfTemplateID(t)
	wfGrpName := "tf-provider-wf-template-tags-wfgrp"
	id := "tf-provider-wf-template-tags"

	if err := createWorkflowGroupFixture(wfGrpName); err != nil {
		t.Errorf("failed to create workflow group fixture: %s", err.Error())
	}
	defer deleteWorkflowGroupFixture(wfGrpName)
	defer deleteWorkflowUsingTemplateFixture(wfGrpName, id)

	config := func(tag, ctxVal string) string {
		return fmt.Sprintf(`
  terraform_config = {
    terraform_version = %q
  }

  tags = [%q]

  context_tags = {
    env = %q
  }
`, "1.5.0", tag, ctxVal)
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader()),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowUsingTemplate(wfGrpName, id, "TERRAFORM", templateID, config("v1", "staging")),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_using_template.test", "id", id),
					resource.TestCheckResourceAttr("stackguardian_workflow_using_template.test", "tags.0", "v1"),
					resource.TestCheckResourceAttr("stackguardian_workflow_using_template.test", "context_tags.env", "staging"),
				),
			},
			{
				Config: testAccWorkflowUsingTemplate(wfGrpName, id, "TERRAFORM", templateID, config("v2", "production")),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_using_template.test", "id", id),
					resource.TestCheckResourceAttr("stackguardian_workflow_using_template.test", "tags.0", "v2"),
					resource.TestCheckResourceAttr("stackguardian_workflow_using_template.test", "context_tags.env", "production"),
				),
			},
		},
	})
}

func TestAccWorkflowUsingTemplate_WithApprovers(t *testing.T) {
	templateID := wfTemplateID(t)
	wfGrpName := "tf-provider-wf-template-approvers-wfgrp"
	id := "tf-provider-wf-template-approvers"

	if err := createWorkflowGroupFixture(wfGrpName); err != nil {
		t.Errorf("failed to create workflow group fixture: %s", err.Error())
	}
	defer deleteWorkflowGroupFixture(wfGrpName)
	defer deleteWorkflowUsingTemplateFixture(wfGrpName, id)

	config := func(numApprovals int) string {
		return fmt.Sprintf(`
  terraform_config = {
    terraform_version = %q
  }

  approvers                    = ["approver@example.com"]
  number_of_approvals_required = %d
`, "1.5.0", numApprovals)
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader()),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowUsingTemplate(wfGrpName, id, "TERRAFORM", templateID, config(1)),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_using_template.test", "id", id),
					resource.TestCheckResourceAttr("stackguardian_workflow_using_template.test", "approvers.0", "approver@example.com"),
					resource.TestCheckResourceAttr("stackguardian_workflow_using_template.test", "number_of_approvals_required", "1"),
				),
			},
			{
				Config: testAccWorkflowUsingTemplate(wfGrpName, id, "TERRAFORM", templateID, config(2)),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_using_template.test", "id", id),
					resource.TestCheckResourceAttr("stackguardian_workflow_using_template.test", "number_of_approvals_required", "2"),
				),
			},
		},
	})
}

func TestAccWorkflowUsingTemplate_WithRunnerConstraints(t *testing.T) {
	templateID := wfTemplateID(t)
	wfGrpName := "tf-provider-wf-template-runner-wfgrp"
	id := "tf-provider-wf-template-runner"

	if err := createWorkflowGroupFixture(wfGrpName); err != nil {
		t.Errorf("failed to create workflow group fixture: %s", err.Error())
	}
	defer deleteWorkflowGroupFixture(wfGrpName)
	defer deleteWorkflowUsingTemplateFixture(wfGrpName, id)

	configShared := fmt.Sprintf(`
  terraform_config = {
    terraform_version = %q
  }

  runner_constraints = {
    type = "shared"
  }
`, "1.5.0")

	configPrivate := fmt.Sprintf(`
  terraform_config = {
    terraform_version = %q
  }

  runner_constraints = {
    type  = "private"
    names = ["runner-1"]
  }
`, "1.5.0")

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader()),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowUsingTemplate(wfGrpName, id, "TERRAFORM", templateID, configShared),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_using_template.test", "id", id),
					resource.TestCheckResourceAttr("stackguardian_workflow_using_template.test", "runner_constraints.type", "shared"),
				),
			},
			{
				Config: testAccWorkflowUsingTemplate(wfGrpName, id, "TERRAFORM", templateID, configPrivate),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_using_template.test", "id", id),
					resource.TestCheckResourceAttr("stackguardian_workflow_using_template.test", "runner_constraints.type", "private"),
					resource.TestCheckResourceAttr("stackguardian_workflow_using_template.test", "runner_constraints.names.0", "runner-1"),
				),
			},
		},
	})
}

func TestAccWorkflowUsingTemplate_WithIacInputData(t *testing.T) {
	templateID := wfTemplateID(t)
	wfGrpName := "tf-provider-wf-template-iac-input-wfgrp"
	id := "tf-provider-wf-template-iac-input"

	if err := createWorkflowGroupFixture(wfGrpName); err != nil {
		t.Errorf("failed to create workflow group fixture: %s", err.Error())
	}
	defer deleteWorkflowGroupFixture(wfGrpName)
	defer deleteWorkflowUsingTemplateFixture(wfGrpName, id)

	configWithIacInput := func(templateID, dataExpr string) string {
		return fmt.Sprintf(`
resource "stackguardian_workflow_using_template" "test" {
  workflow_group_id = %q
  id                = %q
  wf_type           = "TERRAFORM"

  vcs_config = {
    iac_vcs_config = {
      iac_template_id = %q
    }
    iac_input_data = {
      schema_type = "RAW_JSON"
      data        = %s
    }
  }

  terraform_config = {
    terraform_version = %q
  }
}
`, wfGrpName, id, templateID, dataExpr, "1.5.0")
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader()),
		Steps: []resource.TestStep{
			{
				Config: configWithIacInput(templateID, `jsonencode({"env" = "staging"})`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_using_template.test", "id", id),
					resource.TestCheckResourceAttr("stackguardian_workflow_using_template.test", "vcs_config.iac_input_data.schema_type", "RAW_JSON"),
					resource.TestCheckResourceAttr("stackguardian_workflow_using_template.test", "vcs_config.iac_input_data.data", `{"env":"staging"}`),
				),
			},
			{
				Config: configWithIacInput(templateID, `jsonencode({"env" = "production"})`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_using_template.test", "id", id),
					resource.TestCheckResourceAttr("stackguardian_workflow_using_template.test", "vcs_config.iac_input_data.data", `{"env":"production"}`),
				),
			},
		},
	})
}

func TestAccWorkflowUsingTemplate_InNestedWorkflowGroup(t *testing.T) {
	templateID := wfTemplateID(t)
	parentWfGrpName := "tf-provider-wf-template-nested-parent"
	childWfGrpName := parentWfGrpName + "/tf-provider-wf-template-nested-child"
	id := "tf-provider-wf-template-nested"

	if err := createWorkflowGroupFixture(parentWfGrpName); err != nil {
		t.Errorf("failed to create parent workflow group fixture: %s", err.Error())
	}
	defer deleteWorkflowGroupFixture(parentWfGrpName)

	if err := createWorkflowGroupFixture(childWfGrpName); err != nil {
		t.Errorf("failed to create child workflow group fixture: %s", err.Error())
	}
	defer deleteWorkflowGroupFixture(childWfGrpName)
	defer deleteWorkflowUsingTemplateFixture(childWfGrpName, id)

	config := fmt.Sprintf(`
  terraform_config = {
    terraform_version = %q
  }
`, "1.5.0")

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader()),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowUsingTemplate(childWfGrpName, id, "TERRAFORM", templateID, config),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_using_template.test", "workflow_group_id", childWfGrpName),
					resource.TestCheckResourceAttr("stackguardian_workflow_using_template.test", "id", id),
					resource.TestCheckResourceAttr("stackguardian_workflow_using_template.test", "vcs_config.iac_vcs_config.iac_template_id", templateID),
				),
			},
		},
	})
}

func TestAccWorkflowUsingTemplate_ResolvedSchemaPopulated(t *testing.T) {
	templateID := wfTemplateID(t)
	wfGrpName := "tf-provider-wf-template-resolved-wfgrp"
	id := "tf-provider-wf-template-resolved"

	if err := createWorkflowGroupFixture(wfGrpName); err != nil {
		t.Errorf("failed to create workflow group fixture: %s", err.Error())
	}
	defer deleteWorkflowGroupFixture(wfGrpName)
	defer deleteWorkflowUsingTemplateFixture(wfGrpName, id)

	config := `
  description = "test resolved schema"

  terraform_config = {
    terraform_version = "1.5.0"
  }

  environment_variables = [
    {
      kind = "PLAIN_TEXT"
      config = {
        var_name   = "RESOLVED_VAR"
        text_value = "resolved-value"
      }
    }
  ]
`

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader()),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowUsingTemplate(wfGrpName, id, "TERRAFORM", templateID, config),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_using_template.test", "id", id),
					// Verify resolved_schema reflects what the API returned
					resource.TestCheckResourceAttrSet("stackguardian_workflow_using_template.test", "resolved_schema.resource_name"),
					resource.TestCheckResourceAttr("stackguardian_workflow_using_template.test", "resolved_schema.description", "test resolved schema"),
					resource.TestCheckResourceAttr("stackguardian_workflow_using_template.test", "resolved_schema.terraform_config.terraform_version", "1.5.0"),
					resource.TestCheckResourceAttr("stackguardian_workflow_using_template.test", "resolved_schema.environment_variables.0.kind", "PLAIN_TEXT"),
					resource.TestCheckResourceAttr("stackguardian_workflow_using_template.test", "resolved_schema.environment_variables.0.config.var_name", "RESOLVED_VAR"),
				),
			},
		},
	})
}
