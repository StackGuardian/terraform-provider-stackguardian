package workflowgit_test

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
	resourceName := wfGrpName
	payload := &sgsdkgo.WorkflowGroup{
		ResourceName: &resourceName,
	}
	_, err := client.WorkflowGroups.CreateWorkflowGroup(context.TODO(), org, payload)
	return err
}

func deleteWorkflowGroupFixture(wfGrpName string) {
	client := getClient()
	client.WorkflowGroups.DeleteWorkflowGroup(context.TODO(), org, wfGrpName)
}

func deleteWorkflowGitFixture(wfGrpName, workflowName string) {
	client := getClient()
	client.Workflows.DeleteWorkflow(context.TODO(), org, workflowName, wfGrpName)
}

func testAccWorkflowGit(wfGrpName, resourceName, wfType, additionalConfig string) string {
	return fmt.Sprintf(`
resource "stackguardian_workflow_git" "test" {
  workflow_group_id = %q
  id			    = %q
  wf_type           = %q

  %s
}
`, wfGrpName, resourceName, wfType, additionalConfig)
}

func TestAccWorkflowGit_WithVcsConfig(t *testing.T) {
	wfGrpName := "tf-provider-workflow-git-vcs-wfgrp"
	id := "tf-provider-workflow-git-vcs"

	err := createWorkflowGroupFixture(wfGrpName)
	if err != nil {
		t.Errorf("failed to create workflow group fixture: %s", err.Error())
	}
	defer deleteWorkflowGroupFixture(wfGrpName)
	defer deleteWorkflowGitFixture(wfGrpName, id)

	customHeader := http.Header{}
	customHeader.Set("x-sg-internal-auth-orgid", "sg-provider-test")

	config := func(description string) string {
		return fmt.Sprintf(`
  description = %q

  vcs_config = {
    iac_vcs_config = {
      custom_source = {
        config = {
          is_private = false
          repo       = "https://github.com/dummy/test-repo.git"
        }
        source_config_dest_kind = "GIT_OTHER"
      }
    }
  }
`, description)
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowGit(wfGrpName, id, "CUSTOM", config("initial description")),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "workflow_group_id", wfGrpName),
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "id", id),
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "description", "initial description"),
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "vcs_config.iac_vcs_config.custom_source.config.repo", "https://github.com/dummy/test-repo.git"),
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "vcs_config.iac_vcs_config.custom_source.source_config_dest_kind", "GIT_OTHER"),
				),
			},
			{
				Config: testAccWorkflowGit(wfGrpName, id, "CUSTOM", config("updated description")),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "workflow_group_id", wfGrpName),
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "id", id),
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "description", "updated description"),
				),
			},
		},
	})
}

func TestAccWorkflowGit_WithTerraformConfig(t *testing.T) {
	wfGrpName := "tf-provider-workflow-git-tfconfig-wfgrp"
	id := "tf-provider-workflow-git-tfconfig"

	err := createWorkflowGroupFixture(wfGrpName)
	if err != nil {
		t.Errorf("failed to create workflow group fixture: %s", err.Error())
	}
	defer deleteWorkflowGroupFixture(wfGrpName)
	defer deleteWorkflowGitFixture(wfGrpName, id)

	customHeader := http.Header{}
	customHeader.Set("x-sg-internal-auth-orgid", "sg-provider-test")

	config := func(tfVersion string) string {
		return fmt.Sprintf(`
  vcs_config = {
    iac_vcs_config = {
      use_marketplace_template = false
      custom_source = {
        config = {
          is_private = false
          repo       = "https://github.com/dummy/test-repo.git"
        }
        source_config_dest_kind = "GIT_OTHER"
      }
    }
  }

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
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowGit(wfGrpName, id, "TERRAFORM", config("1.5.0")),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "workflow_group_id", wfGrpName),
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "id", id),
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "wf_type", "TERRAFORM"),
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "terraform_config.terraform_version", "1.5.0"),
				),
			},
			{
				Config: testAccWorkflowGit(wfGrpName, id, "TERRAFORM", config("1.6.0")),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "workflow_group_id", wfGrpName),
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "id", id),
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "terraform_config.terraform_version", "1.6.0"),
				),
			},
		},
	})
}

func TestAccWorkflowGit_WithEnvironmentVariables(t *testing.T) {
	wfGrpName := "tf-provider-workflow-git-envvars-wfgrp"
	id := "tf-provider-workflow-git-envvars"

	err := createWorkflowGroupFixture(wfGrpName)
	if err != nil {
		t.Errorf("failed to create workflow group fixture: %s", err.Error())
	}
	defer deleteWorkflowGroupFixture(wfGrpName)
	defer deleteWorkflowGitFixture(wfGrpName, id)

	customHeader := http.Header{}
	customHeader.Set("x-sg-internal-auth-orgid", "sg-provider-test")

	config := func(textValue string) string {
		return fmt.Sprintf(`
  vcs_config = {
    iac_vcs_config = {
      custom_source = {
        config = {
          is_private = false
          repo       = "https://github.com/dummy/test-repo.git"
        }
        source_config_dest_kind = "GIT_OTHER"
      }
    }
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
`, textValue)
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowGit(wfGrpName, id, "CUSTOM", config("initial-value")),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "id", id),
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "environment_variables.0.kind", "PLAIN_TEXT"),
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "environment_variables.0.config.var_name", "MY_VAR"),
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "environment_variables.0.config.text_value", "initial-value"),
				),
			},
			{
				Config: testAccWorkflowGit(wfGrpName, id, "CUSTOM", config("updated-value")),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "id", id),
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "environment_variables.0.config.text_value", "updated-value"),
				),
			},
		},
	})
}

func TestAccWorkflowGit_WithTagsAndContextTags(t *testing.T) {
	wfGrpName := "tf-provider-workflow-git-tags-wfgrp"
	id := "tf-provider-workflow-git-tags"

	err := createWorkflowGroupFixture(wfGrpName)
	if err != nil {
		t.Errorf("failed to create workflow group fixture: %s", err.Error())
	}
	defer deleteWorkflowGroupFixture(wfGrpName)
	defer deleteWorkflowGitFixture(wfGrpName, id)

	customHeader := http.Header{}
	customHeader.Set("x-sg-internal-auth-orgid", "sg-provider-test")

	config := func(tag, ctxVal string) string {
		return fmt.Sprintf(`
  vcs_config = {
    iac_vcs_config = {
      custom_source = {
        config = {
          is_private = false
          repo       = "https://github.com/dummy/test-repo.git"
        }
        source_config_dest_kind = "GIT_OTHER"
      }
    }
  }

  tags = [%q]

  context_tags = {
    env = %q
  }
`, tag, ctxVal)
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowGit(wfGrpName, id, "CUSTOM", config("v1", "staging")),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "id", id),
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "tags.0", "v1"),
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "context_tags.env", "staging"),
				),
			},
			{
				Config: testAccWorkflowGit(wfGrpName, id, "CUSTOM", config("v2", "production")),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "id", id),
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "tags.0", "v2"),
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "context_tags.env", "production"),
				),
			},
		},
	})
}

func TestAccWorkflowGit_WithApprovers(t *testing.T) {
	wfGrpName := "tf-provider-workflow-git-approvers-wfgrp"
	id := "tf-provider-workflow-git-approvers"

	err := createWorkflowGroupFixture(wfGrpName)
	if err != nil {
		t.Errorf("failed to create workflow group fixture: %s", err.Error())
	}
	defer deleteWorkflowGroupFixture(wfGrpName)
	defer deleteWorkflowGitFixture(wfGrpName, id)

	customHeader := http.Header{}
	customHeader.Set("x-sg-internal-auth-orgid", "sg-provider-test")

	config := func(numApprovals int) string {
		return fmt.Sprintf(`
  vcs_config = {
    iac_vcs_config = {
      custom_source = {
        config = {
          is_private = false
          repo       = "https://github.com/dummy/test-repo.git"
        }
        source_config_dest_kind = "GIT_OTHER"
      }
    }
  }

  approvers                    = ["approver@example.com"]
  number_of_approvals_required = %d
`, numApprovals)
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowGit(wfGrpName, id, "CUSTOM", config(1)),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "id", id),
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "approvers.0", "approver@example.com"),
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "number_of_approvals_required", "1"),
				),
			},
			{
				Config: testAccWorkflowGit(wfGrpName, id, "CUSTOM", config(2)),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "id", id),
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "number_of_approvals_required", "2"),
				),
			},
		},
	})
}

func TestAccWorkflowGit_WithUserSchedules(t *testing.T) {
	wfGrpName := "tf-provider-workflow-git-schedules-wfgrp"
	id := "tf-provider-workflow-git-schedules"

	err := createWorkflowGroupFixture(wfGrpName)
	if err != nil {
		t.Errorf("failed to create workflow group fixture: %s", err.Error())
	}
	defer deleteWorkflowGroupFixture(wfGrpName)
	defer deleteWorkflowGitFixture(wfGrpName, id)

	customHeader := http.Header{}
	customHeader.Set("x-sg-internal-auth-orgid", "sg-provider-test")

	config := func(cron string) string {
		return fmt.Sprintf(`
  vcs_config = {
    iac_vcs_config = {
      custom_source = {
        config = {
          is_private = false
          repo       = "https://github.com/dummy/test-repo.git"
        }
        source_config_dest_kind = "GIT_OTHER"
      }
    }
  }

  user_schedules = [
    {
      cron  = %q
      state = "ENABLED"
      desc  = "Runs on schedule"
    }
  ]
`, cron)
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowGit(wfGrpName, id, "CUSTOM", config("0 8 ? * MON *")),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "id", id),
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "user_schedules.0.cron", "0 8 ? * MON *"),
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "user_schedules.0.state", "ENABLED"),
				),
			},
			{
				Config: testAccWorkflowGit(wfGrpName, id, "CUSTOM", config("0 9 ? * MON *")),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "id", id),
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "user_schedules.0.cron", "0 9 ? * MON *"),
				),
			},
		},
	})
}

func TestAccWorkflowGit_WithRunnerConstraints(t *testing.T) {
	wfGrpName := "tf-provider-workflow-git-runner-wfgrp"
	id := "tf-provider-workflow-git-runner"

	err := createWorkflowGroupFixture(wfGrpName)
	if err != nil {
		t.Errorf("failed to create workflow group fixture: %s", err.Error())
	}
	defer deleteWorkflowGroupFixture(wfGrpName)
	defer deleteWorkflowGitFixture(wfGrpName, id)

	customHeader := http.Header{}
	customHeader.Set("x-sg-internal-auth-orgid", "sg-provider-test")

	configShared := `
  vcs_config = {
    iac_vcs_config = {
      custom_source = {
        config = {
          is_private = false
          repo       = "https://github.com/dummy/test-repo.git"
        }
        source_config_dest_kind = "GIT_OTHER"
      }
    }
  }

  runner_constraints = {
    type = "shared"
  }
`

	configWithNames := `
  vcs_config = {
    iac_vcs_config = {
      custom_source = {
        config = {
          is_private = false
          repo       = "https://github.com/dummy/test-repo.git"
        }
        source_config_dest_kind = "GIT_OTHER"
      }
    }
  }

  runner_constraints = {
    type  = "private"
    names = ["runner-1"]
  }
`

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowGit(wfGrpName, id, "CUSTOM", configShared),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "id", id),
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "runner_constraints.type", "shared"),
				),
			},
			{
				Config: testAccWorkflowGit(wfGrpName, id, "CUSTOM", configWithNames),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "id", id),
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "runner_constraints.type", "private"),
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "runner_constraints.names.0", "runner-1"),
				),
			},
		},
	})
}
