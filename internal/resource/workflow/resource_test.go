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
  resource_name     = "%s"
  
  wf_type = "%s"
  
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
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "resource_name", resourceName),
					resource.TestCheckResourceAttrSet("stackguardian_workflow.test", "id"),
				),
			},
			// Update and Read
			{
				Config: testAccWorkflow(wfGrpName, resourceName, "CUSTOM", update_terraform_config),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "workflow_group_id", wfGrpName),
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "resource_name", resourceName),
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "description", "updated desc"),
				),
			},
		},
	})
}

func TestAccWorkflow_Basic_With_VCS_Config(t *testing.T) {
	wfGrpName := "tf-provider-workflow-test-wfgrp"
	resourceName := "tf-provider-workflow-basic"

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
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "resource_name", resourceName),
					resource.TestCheckResourceAttrSet("stackguardian_workflow.test", "id"),
				),
			},
			// Update and Read
			{
				Config: testAccWorkflow(wfGrpName, resourceName, "CUSTOM", fmt.Sprintf(terraform_config, "updated desc")),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "workflow_group_id", wfGrpName),
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "resource_name", resourceName),
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "description", "updated desc"),
				),
			},
		},
	})
}

func TestAccWorkflow_TerraformConfig(t *testing.T) {
	wfGrpName := "tf-provider-workflow-test-wfgrp-terraform-config"
	resourceName := "tf-provider-workflow-basic-terraform-config"
	updatedDescription := "Updated workflow description"

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
				Config: testAccWorkflow(wfGrpName, resourceName, "CUSTOM", ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "workflow_group_id", wfGrpName),
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "resource_name", resourceName),
					resource.TestCheckResourceAttrSet("stackguardian_workflow.test", "id"),
				),
			},
			// Update and Read
			{
				Config: testAccWorkflow(wfGrpName, resourceName, "CUSTOM", updatedDescription),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "workflow_group_id", wfGrpName),
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "resource_name", resourceName),
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "description", updatedDescription),
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
	//templateId := "/sg-provider-test/testing-provider:3"
	templateId := "/sg-provider-test/testing-provider-postman:1"

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader),
		Steps: []resource.TestStep{
			// Create and Read
			{
				Config: testAccWorkflowConfigWithMarketplaceTemplate(wfGrpName, resourceName, templateId),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow.marketplace", "workflow_group_id", wfGrpName),
					resource.TestCheckResourceAttr("stackguardian_workflow.marketplace", "resource_name", resourceName),
					resource.TestCheckResourceAttr("stackguardian_workflow.marketplace", "vcs_config.iac_vcs_config.use_marketplace_template", "true"),
					resource.TestCheckResourceAttr("stackguardian_workflow.marketplace", "vcs_config.iac_vcs_config.iac_template_id", templateId),
					resource.TestCheckResourceAttrSet("stackguardian_workflow.marketplace", "id"),
				),
			},
			// Update and Read
			{
				Config: testAccWorkflowConfigWithMarketplaceTemplateUpdated(wfGrpName, resourceName, templateId),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow.marketplace", "workflow_group_id", wfGrpName),
					resource.TestCheckResourceAttr("stackguardian_workflow.marketplace", "resource_name", resourceName),
					resource.TestCheckResourceAttr("stackguardian_workflow.marketplace", "description", "Updated marketplace workflow"),
					resource.TestCheckResourceAttr("stackguardian_workflow.marketplace", "vcs_config.iac_vcs_config.iac_template_id", templateId),
				),
			},
		},
	})
}

func testAccWorkflowConfigWithMarketplaceTemplate(wfGrpName, resourceName, templateId string) string {
	return fmt.Sprintf(`
resource "stackguardian_workflow" "marketplace" {
  workflow_group_id = "%s"
  resource_name     = "%s"
  tags              = ["test", "terraform", "marketplace"]
  
  wf_type = "TERRAFORM"
  
  terraform_config = {
	terraform_version = "v1.5.7"
  }

  mini_steps = {
	notifications = {
	  email = {
	    approval_required = [
		  {recepients = ["taher.kathanawala@gmail.com"]}
		]
	  }
	}
  }

  vcs_config = {
    iac_vcs_config = {
      use_marketplace_template = true
      iac_template_id          = "%s"
    }
  }
}
`, wfGrpName, resourceName, templateId)
}

func testAccWorkflowConfigWithMarketplaceTemplateUpdated(wfGrpName, resourceName, templateId string) string {
	return fmt.Sprintf(`
resource "stackguardian_workflow" "marketplace" {
  workflow_group_id = "%s"
  resource_name     = "%s"
  description       = "Updated marketplace workflow"
  tags              = ["test", "terraform", "marketplace", "updated"]

  wf_type = "TERRAFORM"
  
  terraform_config = {
	terraform_version = "v1.5.7"
  }

  mini_steps = {
	notifications = {
	  email = {
	    approval_required = [
		  {recepients = ["taher.k@gmail.com"]}
		]
	  }
	}
  }
  
  vcs_config = {
    iac_vcs_config = {
      use_marketplace_template = true
      iac_template_id          = "%s"
    }
  }
}
`, wfGrpName, resourceName, templateId)
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
	//templateId := "/sg-provider-test/testing-provider:1"
	//updatedTemplateId := "/sg-provider-test/testing-provider:3"
	templateId := "/sg-provider-test/testing-provider-postman:1"
	updatedTemplateId := "/sg-provider-test/testing-provider-postman:2"

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader),
		Steps: []resource.TestStep{
			// Create workflow pinned to revision 1
			{
				Config: testAccWorkflowConfigWithTemplateRevision(wfGrpName, resourceName, templateId),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow.upgrade", "workflow_group_id", wfGrpName),
					resource.TestCheckResourceAttr("stackguardian_workflow.upgrade", "resource_name", resourceName),
					resource.TestCheckResourceAttr("stackguardian_workflow.upgrade", "vcs_config.iac_vcs_config.iac_template_id", templateId),
					resource.TestCheckResourceAttrSet("stackguardian_workflow.upgrade", "id"),
				),
			},
			// Upgrade to revision 3
			{
				Config: testAccWorkflowConfigWithTemplateRevision(wfGrpName, resourceName, updatedTemplateId),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow.upgrade", "workflow_group_id", wfGrpName),
					resource.TestCheckResourceAttr("stackguardian_workflow.upgrade", "resource_name", resourceName),
					resource.TestCheckResourceAttr("stackguardian_workflow.upgrade", "vcs_config.iac_vcs_config.iac_template_id", updatedTemplateId),
				),
			},
		},
	})
}

func testAccWorkflowConfigWithTemplateRevision(wfGrpName, resourceName string, templateId string) string {
	return fmt.Sprintf(`
resource "stackguardian_workflow" "upgrade" {
  workflow_group_id = "%s"
  resource_name     = "%s"
  tags              = ["test", "terraform", "upgrade"]

  vcs_config = {
    iac_vcs_config = {
      use_marketplace_template = true
      iac_template_id          = "%s"
    }
  }
  
  user_schedules = [{
	  cron = "0 12 ? * 2 *"
	  state = "ENABLED"
  }]
}
`, wfGrpName, resourceName, templateId)
}
