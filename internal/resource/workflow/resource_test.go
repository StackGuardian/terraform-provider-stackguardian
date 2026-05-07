package workflow_test

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"

	sgsdkgo "github.com/StackGuardian/sg-sdk-go"
	sgclient "github.com/StackGuardian/sg-sdk-go/client"
	sgoption "github.com/StackGuardian/sg-sdk-go/option"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

var org = os.Getenv("STACKGUARDIAN_ORG_NAME")

func GetClient() *sgclient.Client {
	customHeader := http.Header{}
	customHeader.Set("x-sg-internal-auth-orgid", "sg-provider-test")

	client := sgclient.NewClient(
		sgoption.WithApiKey(fmt.Sprintf("apikey %s", os.Getenv("STACKGUARDIAN_API_KEY"))),
		sgoption.WithBaseURL(os.Getenv("STACKGUARDIAN_API_URI")),
		sgoption.WithHTTPHeader(customHeader),
	)
	return client
}

func createWorkflowGroupFixture(wfGrpName string) error {
	client := GetClient()
	resourceName := wfGrpName
	payload := &sgsdkgo.WorkflowGroup{
		ResourceName: &resourceName,
	}
	_, err := client.WorkflowGroups.CreateWorkflowGroup(context.TODO(), org, payload)
	return err
}

func deleteWorkflowGroupFixture(wfGrpName string) {
	client := GetClient()
	client.WorkflowGroups.DeleteWorkflowGroup(context.TODO(), org, wfGrpName)
}

func deleteWorkflowFixture(wfGrpName, workflowName string) {
	client := GetClient()
	client.Workflows.DeleteWorkflow(context.TODO(), org, workflowName, wfGrpName)
}

func testAccWorkflow(wfGrpName, resourceName, wf_type, additionalConfig string) string {
	return fmt.Sprintf(`
resource "stackguardian_workflow" "test" {
  workflow_group_id = "%s"
  id                = "%s"
  resource_name     = "test-workflow"
  wf_type           = "%s"

  %s
}
`, wfGrpName, resourceName, wf_type, additionalConfig)
}

func TestAccWorkflow_Basic(t *testing.T) {
	wfGrpName := "tf-provider-workflow-test-wfgrp"
	resourceName := "tf-provider-workflow-basic"

	terraform_config := `
description = "desc"
	`

	update_terraform_config := `
description = "updated desc"
	`

	err := createWorkflowGroupFixture(wfGrpName)
	if err != nil {
		t.Errorf("failed to create workflow group fixture: %s", err.Error())
	}
	defer deleteWorkflowGroupFixture(wfGrpName)          // runs last (registered first)
	defer deleteWorkflowFixture(wfGrpName, resourceName) // runs first (registered second)

	customHeader := http.Header{}
	customHeader.Set("x-sg-internal-auth-orgid", "sg-provider-test")

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader),
		Steps: []resource.TestStep{
			// Create and Read
			{
				Config: testAccWorkflow(wfGrpName, resourceName, "CUSTOM", terraform_config),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "workflow_group_id", wfGrpName),
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "id", resourceName),
				),
			},
			// Update and Read
			{
				Config: testAccWorkflow(wfGrpName, resourceName, "CUSTOM", update_terraform_config),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "workflow_group_id", wfGrpName),
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "id", resourceName),
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "description", "updated desc"),
				),
			},
		},
	})
}

func TestAccWorkflow_Basic_With_VCS_Config(t *testing.T) {
	wfGrpName := "tf-provider-workflow-test-vcs-config"
	resourceName := "tf-provider-workflow"

	err := createWorkflowGroupFixture(wfGrpName)
	if err != nil {
		t.Errorf("failed to create workflow group fixture: %s", err.Error())
	}
	defer deleteWorkflowGroupFixture(wfGrpName)          // runs last (registered first)
	defer deleteWorkflowFixture(wfGrpName, resourceName) // runs first (registered second)

	terraform_config := `
  description = "%s"
  vcs_config = {
    iac_vcs_config = {
      use_marketplace_template = false
      custom_source = {
    	config = {
    	  is_private = false
    	  repo = "https://github.com/dummy/test-repo.git"
    	} 
    	source_config_dest_kind = "GIT_OTHER"
      }
    }
  }
	`

	customHeader := http.Header{}
	customHeader.Set("x-sg-internal-auth-orgid", "sg-provider-test")

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader),
		Steps: []resource.TestStep{
			// Create and Read
			{
				Config: testAccWorkflow(wfGrpName, resourceName, "CUSTOM", fmt.Sprintf(terraform_config, "test")),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "workflow_group_id", wfGrpName),
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "id", resourceName),
				),
			},
			// Update and Read
			{
				Config: testAccWorkflow(wfGrpName, resourceName, "CUSTOM", fmt.Sprintf(terraform_config, "updated desc")),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "workflow_group_id", wfGrpName),
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "id", resourceName),
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "description", "updated desc"),
				),
			},
		},
	})
}

func TestAccWorkflow_TerraformConfig(t *testing.T) {
	wfGrpName := "tf-provider-workflow-test-wfgrp-terraform-config"
	resourceName := "tf-provider-workflow-basic-terraform-config"

	err := createWorkflowGroupFixture(wfGrpName)
	if err != nil {
		t.Errorf("failed to create workflow group fixture: %s", err.Error())
	}
	defer deleteWorkflowGroupFixture(wfGrpName)          // runs last (registered first)
	defer deleteWorkflowFixture(wfGrpName, resourceName) // runs first (registered second)

	terraform_config := `
	terraform_config = {
		drift_cron = "0 */6 * * ? *"
		pre_plan_wf_steps_config = [{
			name = "test-step"
			wf_step_template_id = "/sg-provider-test/taher-null-resource:%s"
		}]

		pre_plan_hooks = [
			"%s"
		]
		
		mount_points = [{
			"source": "/source_dir"
			"target": "%s"
		}]		

		managed_terraform_state = false
		terraform_version = "v1.5.7"
	}
`

	customHeader := http.Header{}
	customHeader.Set("x-sg-internal-auth-orgid", "sg-provider-test")

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader),
		Steps: []resource.TestStep{
			// Create and Read
			{
				Config: testAccWorkflow(wfGrpName, resourceName, "TERRAFORM", fmt.Sprintf(terraform_config, "1", "sleep 3", "target_dir")),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "workflow_group_id", wfGrpName),
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "id", resourceName),
				),
			},
			// Update and Read
			{
				Config: testAccWorkflow(wfGrpName, resourceName, "TERRAFORM", fmt.Sprintf(terraform_config, "2", "echo sleep", "updated_dir")),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "workflow_group_id", wfGrpName),
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "id", resourceName),
				),
			},
		},
	})
}

func TestAccWorkflow_WfStepsConfig(t *testing.T) {
	wfGrpName := "tf-provider-workflow-test-wfsteps-wfgrp"
	resourceName := "tf-provider-workflow-wfsteps"

	err := createWorkflowGroupFixture(wfGrpName)
	if err != nil {
		t.Errorf("failed to create workflow group fixture: %s", err.Error())
	}
	defer deleteWorkflowGroupFixture(wfGrpName)
	defer deleteWorkflowFixture(wfGrpName, resourceName)

	wf_steps_config := `
	wf_steps_config = [
		{
			name                = "test-step"
			wf_step_template_id = "/stackguardian/terraform:11"
			wf_step_input_data  = {
				schema_type = "FORM_JSONSCHEMA"
				data        = jsonencode({
					terraformVersion      = "1.3.5"
					managedTerraformState = false
					terraformAction       = "plan-destroy"
				})
			}
			approval = false
			timeout  = 2100
		},
		{
			name                = "calls"
			wf_step_template_id = "/demo-org/ansible:6"
			wf_step_input_data  = {
				schema_type = "FORM_JSONSCHEMA"
				data        = jsonencode({
					playbookPath     = "playbook.yaml"
					captureOutputs   = true
					ansibleLocalhost = false
					ansibleAction    = "run"
				})
			}
			approval = false
			timeout  = 2100
		}
	]
`

	wf_steps_config_updated := `
	wf_steps_config = [
		{
			name                = "calls"
			wf_step_template_id = "/demo-org/ansible:6"
			wf_step_input_data  = {
				schema_type = "FORM_JSONSCHEMA"
				data        = jsonencode({
					playbookPath     = "updated-playbook.yaml"
					captureOutputs   = true
					ansibleLocalhost = false
					ansibleAction    = "run"
				})
			}
			approval = false
			timeout  = 3600
		}
	]
`

	customHeader := http.Header{}
	customHeader.Set("x-sg-internal-auth-orgid", "sg-provider-test")

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader),
		Steps: []resource.TestStep{
			// Create and Read
			{
				Config: testAccWorkflow(wfGrpName, resourceName, "CUSTOM", wf_steps_config),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "workflow_group_id", wfGrpName),
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "id", resourceName),
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "wf_steps_config.#", "2"),
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "wf_steps_config.0.name", "test-step"),
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "wf_steps_config.0.wf_step_template_id", "/stackguardian/terraform:11"),
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "wf_steps_config.1.name", "calls"),
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "wf_steps_config.1.wf_step_template_id", "/demo-org/ansible:6"),
				),
			},
			// Update and Read
			{
				Config: testAccWorkflow(wfGrpName, resourceName, "CUSTOM", wf_steps_config_updated),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "workflow_group_id", wfGrpName),
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "id", resourceName),
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "wf_steps_config.#", "1"),
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "wf_steps_config.0.timeout", "3600"),
				),
			},
		},
	})
}

func TestAccWorkflow_MiniSteps(t *testing.T) {
	wfGrpName := "tf-provider-workflow-test-ministeps-wfgrp"
	resourceName := "tf-provider-workflow-ministeps"

	err := createWorkflowGroupFixture(wfGrpName)
	if err != nil {
		t.Errorf("failed to create workflow group fixture: %s", err.Error())
	}
	defer deleteWorkflowGroupFixture(wfGrpName)
	defer deleteWorkflowFixture(wfGrpName, resourceName)

	mini_steps_config := `
	mini_steps = {
		webhooks = {
			completed = [
				{
					webhook_name = "test_webhook"
					webhook_url  = "https://www.myservice.com/ping/"
				}
			]
			drift_detected = []
			errored        = []
		}
		notifications = {
			email = {
				approval_required = []
				cancelled         = []
				completed         = []
				errored = [
					{
						recipients = ["taherkathanawala@stackguardian.io"]
					}
				]
			}
		}
		wf_chaining = {
			completed = [
				{
					workflow_group_id = "ansible"
					workflow_id       = "ansible-workflow"
					stack_id          = ""
				}
			]
			errored = [
				{
					workflow_group_id    = "test-cli-driven-workflow"
					workflow_id          = "test-cli-driven-workflow"
					workflow_run_payload = jsonencode({
						TerraformAction = {
							action = "apply"
						}
					})
					stack_id = ""
				}
			]
		}
	}
`

	mini_steps_config_updated := `
	mini_steps = {
		webhooks = {
			completed = [
				{
					webhook_name = "test_webhook"
					webhook_url  = "https://www.myservice.com/ping/updated/"
				}
			]
			drift_detected = []
			errored        = []
		}
		notifications = {
			email = {
				approval_required = []
				cancelled         = []
				completed         = []
				errored = [
					{
						recipients = ["taherkathanawala@stackguardian.io", "second@stackguardian.io"]
					}
				]
			}
		}
		wf_chaining = {
			completed = [
				{
					workflow_group_id = "ansible"
					workflow_id       = "ansible-workflow"
					stack_id          = ""
				}
			]
			errored = [
				{
					workflow_group_id    = "test-cli-driven-workflow"
					workflow_id          = "test-cli-driven-workflow"
					workflow_run_payload = jsonencode({
						TerraformAction = {
							action = "apply"
						}
					})
					stack_id = ""
				}
			]
		}
	}
`

	customHeader := http.Header{}
	customHeader.Set("x-sg-internal-auth-orgid", "sg-provider-test")

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader),
		Steps: []resource.TestStep{
			// Create and Read
			{
				Config: testAccWorkflow(wfGrpName, resourceName, "CUSTOM", mini_steps_config),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "workflow_group_id", wfGrpName),
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "id", resourceName),
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "mini_steps.webhooks.completed.#", "1"),
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "mini_steps.webhooks.completed.0.webhook_name", "test_webhook"),
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "mini_steps.webhooks.completed.0.webhook_url", "https://www.myservice.com/ping/"),
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "mini_steps.notifications.email.errored.#", "1"),
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "mini_steps.notifications.email.errored.0.recipients.#", "1"),
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "mini_steps.notifications.email.errored.0.recipients.0", "taherkathanawala@stackguardian.io"),
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "mini_steps.wf_chaining.completed.#", "1"),
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "mini_steps.wf_chaining.completed.0.workflow_group_id", "ansible"),
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "mini_steps.wf_chaining.completed.0.workflow_id", "ansible-workflow"),
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "mini_steps.wf_chaining.errored.#", "1"),
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "mini_steps.wf_chaining.errored.0.workflow_group_id", "test-cli-driven-workflow"),
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "mini_steps.wf_chaining.errored.0.workflow_id", "test-cli-driven-workflow"),
				),
			},
			// Update and Read
			{
				Config: testAccWorkflow(wfGrpName, resourceName, "CUSTOM", mini_steps_config_updated),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "workflow_group_id", wfGrpName),
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "id", resourceName),
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "mini_steps.webhooks.completed.0.webhook_url", "https://www.myservice.com/ping/updated/"),
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "mini_steps.notifications.email.errored.0.recipients.#", "2"),
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "mini_steps.notifications.email.errored.0.recipients.1", "second@stackguardian.io"),
				),
			},
		},
	})
}

func TestAccWorkflow_WithMarketplaceTemplate(t *testing.T) {
	wfGrpName := "tf-provider-workflow-test-marketplace-wfgrp"
	resourceName := "tf-provider-workflow-marketplace"

	err := createWorkflowGroupFixture(wfGrpName)
	if err != nil {
		t.Errorf("failed to create workflow group fixture: %s", err.Error())
	}
	defer deleteWorkflowGroupFixture(wfGrpName)
	defer deleteWorkflowFixture(wfGrpName, resourceName)

	customHeader := http.Header{}
	customHeader.Set("x-sg-internal-auth-orgid", "sg-provider-test")
	templateId := "/sg-provider-test/testing-provider:3"

	baseConfig := func(extraAttrs string) string {
		return fmt.Sprintf(`
%s

terraform_config = {
  drift_cron = "0 */6 * * ? *"
  terraform_version = "v1.5.7"
}

mini_steps = {
  notifications = {
    email = {
      approval_required = [
        { recipients = ["taher.kathanawala@gmail.com"] }
      ]
    }
  }
  webhooks = {
    errored = [
      {
        webhook_name = "on_error"
        webhook_url  = "https://www.myservice.com/error/"
      }
    ]
  }
}

vcs_config = {
  iac_vcs_config = {
    use_marketplace_template = true
    iac_template_id          = "%s"
  }
}
`, extraAttrs, templateId)
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader),
		Steps: []resource.TestStep{
			// Create and Read
			{
				Config: testAccWorkflow(wfGrpName, resourceName, "TERRAFORM", baseConfig(`tags = ["test", "terraform", "marketplace"]`)),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "workflow_group_id", wfGrpName),
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "id", resourceName),
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "vcs_config.iac_vcs_config.use_marketplace_template", "true"),
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "vcs_config.iac_vcs_config.iac_template_id", templateId),
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "mini_steps.webhooks.errored.#", "1"),
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "mini_steps.webhooks.errored.0.webhook_name", "on_error"),
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "mini_steps.webhooks.errored.0.webhook_url", "https://www.myservice.com/error/"),
				),
			},
			// Update and Read
			{
				Config: testAccWorkflow(wfGrpName, resourceName, "TERRAFORM", baseConfig(`
description = "Updated marketplace workflow"
tags        = ["test", "terraform", "marketplace", "updated"]
`)),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "workflow_group_id", wfGrpName),
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "id", resourceName),
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "description", "Updated marketplace workflow"),
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "vcs_config.iac_vcs_config.iac_template_id", templateId),
				),
			},
		},
	})
}

func TestAccWorkflow_TemplateRevisionUpgrade(t *testing.T) {
	wfGrpName := "tf-provider-workflow-test-upgrade-wfgrp"
	resourceName := "tf-provider-workflow-upgrade"

	err := createWorkflowGroupFixture(wfGrpName)
	if err != nil {
		t.Errorf("failed to create workflow group fixture: %s", err.Error())
	}
	defer deleteWorkflowGroupFixture(wfGrpName)
	defer deleteWorkflowFixture(wfGrpName, resourceName)

	customHeader := http.Header{}
	customHeader.Set("x-sg-internal-auth-orgid", "sg-provider-test")
	templateId := "/sg-provider-test/testing-provider:1"
	updatedTemplateId := "/sg-provider-test/testing-provider:4"
	config := func(templateID, description, envValue, webhookURL string) string {
		return fmt.Sprintf(`
description = "%s"

tags = ["test", "terraform", "upgrade"]

context_tags = {
  environment = "test"
  team        = "sg-provider"
}

environment_variables = [
  {
    config = {
      var_name   = "MY_ENV_VAR"
      text_value = "%s"
    }
    kind = "PLAIN_TEXT"
  },
  {
    config = {
      var_name   = "ANOTHER_VAR"
      text_value = "constant"
    }
    kind = "PLAIN_TEXT"
  }
]

runner_constraints = {
  type = "shared"
}

approvers                    = ["taher.kathanawala@stackguardian.io"]
number_of_approvals_required = 1

mini_steps = {
  webhooks = {
    approval_required = []
    cancelled         = []
    completed = [
      {
        webhook_name = "on_complete"
        webhook_url  = "%s"
      }
    ]
    drift_detected = []
    errored = [
      {
        webhook_name = "on_error"
        webhook_url  = "https://www.myservice.com/error/"
      }
    ]
  }
  notifications = {
    email = {
      approval_required = []
      cancelled         = []
      completed = [
        {
          recipients = ["taher.kathanawala@stackguardian.io"]
        }
      ]
      drift_detected = []
      errored = [
        {
          recipients = ["taher.kathanawala@stackguardian.io", "team@stackguardian.io"]
        }
      ]
    }
  }
  wf_chaining = {
    completed = [
      {
        workflow_group_id = "ansible"
        workflow_id       = "ansible-workflow"
        stack_id          = ""
      }
    ]
    errored = [
      {
        workflow_group_id    = "test-cli-driven-workflow"
        workflow_id          = "test-cli-driven-workflow"
        workflow_run_payload = jsonencode({
          TerraformAction = {
            action = "apply"
          }
        })
        stack_id = ""
      }
    ]
  }
}

vcs_config = {
  iac_vcs_config = {
    use_marketplace_template = true
    iac_template_id          = "%s"
  }
}
`, description, envValue, webhookURL, templateID)
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader),
		Steps: []resource.TestStep{
			// Create with full config pinned to revision 1
			{
				Config: testAccWorkflow(wfGrpName, resourceName, "CUSTOM",
					config(templateId, "workflow for template revision upgrade testing", "initial_value", "https://www.myservice.com/ping/")),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "workflow_group_id", wfGrpName),
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "id", resourceName),
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "description", "workflow for template revision upgrade testing"),
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "tags.#", "3"),
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "context_tags.environment", "test"),
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "context_tags.team", "sg-provider"),
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "environment_variables.#", "2"),
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "environment_variables.0.config.var_name", "MY_ENV_VAR"),
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "environment_variables.0.config.text_value", "initial_value"),
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "environment_variables.0.kind", "PLAIN_TEXT"),
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "environment_variables.1.config.var_name", "ANOTHER_VAR"),
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "runner_constraints.type", "shared"),
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "approvers.#", "1"),
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "approvers.0", "taher.kathanawala@stackguardian.io"),
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "number_of_approvals_required", "1"),
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "mini_steps.webhooks.completed.#", "1"),
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "mini_steps.webhooks.completed.0.webhook_name", "on_complete"),
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "mini_steps.webhooks.completed.0.webhook_url", "https://www.myservice.com/ping/"),
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "mini_steps.webhooks.errored.#", "1"),
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "mini_steps.webhooks.errored.0.webhook_name", "on_error"),
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "mini_steps.notifications.email.completed.#", "1"),
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "mini_steps.notifications.email.completed.0.recipients.0", "taher.kathanawala@stackguardian.io"),
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "mini_steps.notifications.email.errored.0.recipients.#", "2"),
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "mini_steps.wf_chaining.completed.#", "1"),
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "mini_steps.wf_chaining.completed.0.workflow_group_id", "ansible"),
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "mini_steps.wf_chaining.completed.0.workflow_id", "ansible-workflow"),
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "mini_steps.wf_chaining.errored.#", "1"),
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "mini_steps.wf_chaining.errored.0.workflow_group_id", "test-cli-driven-workflow"),
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "vcs_config.iac_vcs_config.iac_template_id", templateId),
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "vcs_config.iac_vcs_config.use_marketplace_template", "true"),
				),
			},
			// Upgrade to revision 2 — mutate several fields alongside the template bump
			{
				Config: testAccWorkflow(wfGrpName, resourceName, "CUSTOM",
					config(updatedTemplateId, "updated after template upgrade", "updated_value", "https://www.myservice.com/ping/updated/")),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "workflow_group_id", wfGrpName),
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "id", resourceName),
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "description", "updated after template upgrade"),
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "environment_variables.0.config.text_value", "updated_value"),
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "mini_steps.webhooks.completed.0.webhook_url", "https://www.myservice.com/ping/updated/"),
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "vcs_config.iac_vcs_config.iac_template_id", updatedTemplateId),
					// unchanged fields persist correctly after upgrade
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "environment_variables.1.config.var_name", "ANOTHER_VAR"),
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "runner_constraints.type", "shared"),
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "approvers.0", "taher.kathanawala@stackguardian.io"),
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "mini_steps.wf_chaining.completed.0.workflow_group_id", "ansible"),
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "mini_steps.notifications.email.errored.0.recipients.#", "2"),
				),
			},
		},
	})
}
