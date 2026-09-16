package workflowgit_test

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"

	sgsdkgo "github.com/StackGuardian/sg-sdk-go"
	sgclient "github.com/StackGuardian/sg-sdk-go/client"
	sgoption "github.com/StackGuardian/sg-sdk-go/option"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/acctest"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

func getClient() *sgclient.Client {
	customHeader := http.Header{}
	customHeader.Set("x-sg-internal-auth-orgid", "sg-provider-test")

	return sgclient.NewClient(
		sgoption.WithApiKey(config.Get().FormatApiKey()),
		sgoption.WithBaseURL(config.Get().ApiUri),
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
		_, err := client.WorkflowGroups.CreateChildWorkflowGroup(context.TODO(), config.Get().OrgName, parent, payload)
		return err
	}
	_, err := client.WorkflowGroups.CreateWorkflowGroup(context.TODO(), config.Get().OrgName, payload)
	return err
}

func deleteWorkflowGroupFixture(wfGrpName string) {
	client := getClient()
	_, _ = client.WorkflowGroups.DeleteWorkflowGroup(context.TODO(), config.Get().OrgName, wfGrpName)
}

func deleteWorkflowGitFixture(wfGrpName, workflowName string) {
	client := getClient()
	_, _ = client.Workflows.DeleteWorkflow(context.TODO(), config.Get().OrgName, workflowName, wfGrpName)
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
	wfGrpName := acctest.ResourceName("tf-provider-workflow-git-vcs-wfgrp")
	id := acctest.ResourceName("tf-provider-workflow-git-vcs")

	t.Cleanup(func() {
		deleteWorkflowGitFixture(wfGrpName, id)
		deleteWorkflowGroupFixture(wfGrpName)
	})

	err := createWorkflowGroupFixture(wfGrpName)
	if err != nil {
		t.Errorf("failed to create workflow group fixture: %s", err.Error())
	}

	customHeader := http.Header{}
	customHeader.Set("x-sg-internal-auth-orgid", "sg-provider-test")

	gitCallback := func(description string) string {
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
				Config: testAccWorkflowGit(wfGrpName, id, "CUSTOM", gitCallback("initial description")),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "workflow_group_id", wfGrpName),
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "id", id),
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "description", "initial description"),
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "vcs_config.iac_vcs_config.custom_source.config.repo", "https://github.com/dummy/test-repo.git"),
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "vcs_config.iac_vcs_config.custom_source.source_config_dest_kind", "GIT_OTHER"),
				),
			},
			{
				Config: testAccWorkflowGit(wfGrpName, id, "CUSTOM", gitCallback("updated description")),
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
	wfGrpName := acctest.ResourceName("tf-provider-workflow-git-tfconfig-wfgrp")
	id := acctest.ResourceName("tf-provider-workflow-git-tfconfig")

	t.Cleanup(func() {
		deleteWorkflowGitFixture(wfGrpName, id)
		deleteWorkflowGroupFixture(wfGrpName)
	})

	err := createWorkflowGroupFixture(wfGrpName)
	if err != nil {
		t.Errorf("failed to create workflow group fixture: %s", err.Error())
	}

	customHeader := http.Header{}
	customHeader.Set("x-sg-internal-auth-orgid", "sg-provider-test")

	gitCallback := func(tfVersion string) string {
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
				Config: testAccWorkflowGit(wfGrpName, id, "TERRAFORM", gitCallback("1.5.0")),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "workflow_group_id", wfGrpName),
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "id", id),
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "wf_type", "TERRAFORM"),
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "terraform_config.terraform_version", "1.5.0"),
				),
			},
			{
				Config: testAccWorkflowGit(wfGrpName, id, "TERRAFORM", gitCallback("1.5.7")),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "workflow_group_id", wfGrpName),
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "id", id),
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "terraform_config.terraform_version", "1.5.7"),
				),
			},
		},
	})
}

func TestAccWorkflowGit_WithEnvironmentVariables(t *testing.T) {
	wfGrpName := acctest.ResourceName("tf-provider-workflow-git-envvars-wfgrp")
	id := acctest.ResourceName("tf-provider-workflow-git-envvars")

	t.Cleanup(func() {
		deleteWorkflowGitFixture(wfGrpName, id)
		deleteWorkflowGroupFixture(wfGrpName)
	})

	err := createWorkflowGroupFixture(wfGrpName)
	if err != nil {
		t.Errorf("failed to create workflow group fixture: %s", err.Error())
	}

	customHeader := http.Header{}
	customHeader.Set("x-sg-internal-auth-orgid", "sg-provider-test")

	gitCallback := func(textValue string) string {
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
				Config: testAccWorkflowGit(wfGrpName, id, "CUSTOM", gitCallback("initial-value")),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "id", id),
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "environment_variables.0.kind", "PLAIN_TEXT"),
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "environment_variables.0.config.var_name", "MY_VAR"),
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "environment_variables.0.config.text_value", "initial-value"),
				),
			},
			{
				Config: testAccWorkflowGit(wfGrpName, id, "CUSTOM", gitCallback("updated-value")),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "id", id),
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "environment_variables.0.config.text_value", "updated-value"),
				),
			},
		},
	})
}

func TestAccWorkflowGit_WithTagsAndContextTags(t *testing.T) {
	wfGrpName := acctest.ResourceName("tf-provider-workflow-git-tags-wfgrp")
	id := acctest.ResourceName("tf-provider-workflow-git-tags")

	t.Cleanup(func() {
		deleteWorkflowGitFixture(wfGrpName, id)
		deleteWorkflowGroupFixture(wfGrpName)
	})

	err := createWorkflowGroupFixture(wfGrpName)
	if err != nil {
		t.Errorf("failed to create workflow group fixture: %s", err.Error())
	}

	customHeader := http.Header{}
	customHeader.Set("x-sg-internal-auth-orgid", "sg-provider-test")

	gitCallback := func(tag, ctxVal string) string {
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
				Config: testAccWorkflowGit(wfGrpName, id, "CUSTOM", gitCallback("v1", "staging")),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "id", id),
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "tags.0", "v1"),
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "context_tags.env", "staging"),
				),
			},
			{
				Config: testAccWorkflowGit(wfGrpName, id, "CUSTOM", gitCallback("v2", "production")),
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
	wfGrpName := acctest.ResourceName("tf-provider-workflow-git-approvers-wfgrp")
	id := acctest.ResourceName("tf-provider-workflow-git-approvers")

	t.Cleanup(func() {
		deleteWorkflowGitFixture(wfGrpName, id)
		deleteWorkflowGroupFixture(wfGrpName)
	})

	err := createWorkflowGroupFixture(wfGrpName)
	if err != nil {
		t.Errorf("failed to create workflow group fixture: %s", err.Error())
	}

	customHeader := http.Header{}
	customHeader.Set("x-sg-internal-auth-orgid", "sg-provider-test")

	gitCallback := func(numApprovals int) string {
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
				Config: testAccWorkflowGit(wfGrpName, id, "CUSTOM", gitCallback(1)),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "id", id),
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "approvers.0", "approver@example.com"),
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "number_of_approvals_required", "1"),
				),
			},
			{
				Config: testAccWorkflowGit(wfGrpName, id, "CUSTOM", gitCallback(2)),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "id", id),
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "number_of_approvals_required", "2"),
				),
			},
		},
	})
}

func TestAccWorkflowGit_WithUserSchedules(t *testing.T) {
	wfGrpName := acctest.ResourceName("tf-provider-workflow-git-schedules-wfgrp")
	id := acctest.ResourceName("tf-provider-workflow-git-schedules")

	t.Cleanup(func() {
		deleteWorkflowGitFixture(wfGrpName, id)
		deleteWorkflowGroupFixture(wfGrpName)
	})

	err := createWorkflowGroupFixture(wfGrpName)
	if err != nil {
		t.Errorf("failed to create workflow group fixture: %s", err.Error())
	}

	customHeader := http.Header{}
	customHeader.Set("x-sg-internal-auth-orgid", "sg-provider-test")

	gitCallback := func(cron string) string {
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
				Config: testAccWorkflowGit(wfGrpName, id, "CUSTOM", gitCallback("0 8 ? * MON *")),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "id", id),
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "user_schedules.0.cron", "0 8 ? * MON *"),
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "user_schedules.0.state", "ENABLED"),
				),
			},
			{
				Config: testAccWorkflowGit(wfGrpName, id, "CUSTOM", gitCallback("0 9 ? * MON *")),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "id", id),
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "user_schedules.0.cron", "0 9 ? * MON *"),
				),
			},
		},
	})
}

func TestAccWorkflowGit_WithVcsTriggers_Push(t *testing.T) {
	wfGrpName := acctest.ResourceName("tf-provider-workflow-git-vcs-triggers-push-wfgrp")
	id := acctest.ResourceName("tf-provider-workflow-git-vcs-triggers-push")

	t.Cleanup(func() {
		deleteWorkflowGitFixture(wfGrpName, id)
		deleteWorkflowGroupFixture(wfGrpName)
	})

	err := createWorkflowGroupFixture(wfGrpName)
	if err != nil {
		t.Errorf("failed to create workflow group fixture: %s", err.Error())
	}

	customHeader := http.Header{}
	customHeader.Set("x-sg-internal-auth-orgid", "sg-provider-test")

	configPush := func(planOnly bool) string {
		return fmt.Sprintf(`
		  vcs_config = {
		    iac_vcs_config = {
		      custom_source = {
		        config = {
		          is_private = false
		          repo       = "https://github.com/dummy/test-repo.git"
		        }
		        source_config_dest_kind = "GITHUB_COM"
		      }
		    }
		  }
		
		  vcs_triggers = {
		    tracked_branch = "main"
		    plan_only      = %t
		
		    push = {
		      createWfRun = { enabled = true }
		    }
		  }
		`, planOnly)
	}

	configPushWithFilters := func(patterns string) string {
		return fmt.Sprintf(`
		  vcs_config = {
		    iac_vcs_config = {
		      custom_source = {
		        config = {
		          is_private = false
		          repo       = "https://github.com/dummy/test-repo.git"
		        }
		        source_config_dest_kind = "GITHUB_COM"
		      }
		    }
		  }
		
		  vcs_triggers = {
		    tracked_branch        = "main"
		    file_triggers_enabled = true
		    file_trigger_patterns = %s
		
		    push = {
		      createWfRun = { enabled = true }
		    }
		  }
		`, patterns)
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowGit(wfGrpName, id, "CUSTOM", configPush(false)),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "id", id),
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "vcs_triggers.tracked_branch", "main"),
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "vcs_triggers.plan_only", "false"),
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "vcs_triggers.push.createWfRun.enabled", "true"),
				),
			},
			{
				Config: testAccWorkflowGit(wfGrpName, id, "CUSTOM", configPush(true)),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "id", id),
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "vcs_triggers.plan_only", "true"),
				),
			},
			{
				Config: testAccWorkflowGit(wfGrpName, id, "CUSTOM", configPushWithFilters(`["*.tf"]`)),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "id", id),
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "vcs_triggers.file_triggers_enabled", "true"),
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "vcs_triggers.file_trigger_patterns.0", "*.tf"),
				),
			},
			{
				Config: testAccWorkflowGit(wfGrpName, id, "CUSTOM", configPushWithFilters(`["*.tf", "infra/**/*.json"]`)),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "id", id),
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "vcs_triggers.file_trigger_patterns.0", "*.tf"),
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "vcs_triggers.file_trigger_patterns.1", "infra/**/*.json"),
				),
			},
		},
	})
}

func TestAccWorkflowGit_WithVcsTriggers_PullRequest(t *testing.T) {
	wfGrpName := acctest.ResourceName("tf-provider-workflow-git-vcs-triggers-pr-wfgrp")
	id := acctest.ResourceName("tf-provider-workflow-git-vcs-triggers-pr")

	t.Cleanup(func() {
		deleteWorkflowGitFixture(wfGrpName, id)
		deleteWorkflowGroupFixture(wfGrpName)
	})

	err := createWorkflowGroupFixture(wfGrpName)
	if err != nil {
		t.Errorf("failed to create workflow group fixture: %s", err.Error())
	}

	customHeader := http.Header{}
	customHeader.Set("x-sg-internal-auth-orgid", "sg-provider-test")

	configWithBoth := `
		  vcs_config = {
		    iac_vcs_config = {
		      custom_source = {
		        config = {
		          is_private = false
		          repo       = "https://github.com/dummy/test-repo.git"
		        }
		        source_config_dest_kind = "GITHUB_COM"
		      }
		    }
		  }
		
		  vcs_triggers = {
		    tracked_branch = "main"
		
		    pull_request_opened = {
		      createWfRun = { enabled = true }
		    }
		    pull_request_modified = {
		      createWfRun = { enabled = true }
		    }
		  }
		`

	configOpenedOnly := `
		  vcs_config = {
		    iac_vcs_config = {
		      custom_source = {
		        config = {
		          is_private = false
		          repo       = "https://github.com/dummy/test-repo.git"
		        }
		        source_config_dest_kind = "GITHUB_COM"
		      }
		    }
		  }
		
		  vcs_triggers = {
		    tracked_branch = "main"
		
		    pull_request_opened = {
		      createWfRun = { enabled = true }
		    }
		  }
		`

	configAllPR := func(enabled bool) string {
		return fmt.Sprintf(`
		  vcs_config = {
		    iac_vcs_config = {
		      custom_source = {
		        config = {
		          is_private = false
		          repo       = "https://github.com/dummy/test-repo.git"
		        }
		        source_config_dest_kind = "GITHUB_COM"
		      }
		    }
		  }
		
		  vcs_triggers = {
		    all_pull_requests = {
		      createWfRun = { enabled = %t }
		    }
		  }
		`, enabled)
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowGit(wfGrpName, id, "CUSTOM", configWithBoth),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "id", id),
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "vcs_triggers.tracked_branch", "main"),
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "vcs_triggers.pull_request_opened.createWfRun.enabled", "true"),
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "vcs_triggers.pull_request_modified.createWfRun.enabled", "true"),
				),
			},
			{
				Config: testAccWorkflowGit(wfGrpName, id, "CUSTOM", configOpenedOnly),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "id", id),
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "vcs_triggers.pull_request_opened.createWfRun.enabled", "true"),
					resource.TestCheckNoResourceAttr("stackguardian_workflow_git.test", "vcs_triggers.pull_request_modified"),
				),
			},
			{
				Config: testAccWorkflowGit(wfGrpName, id, "CUSTOM", configAllPR(true)),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "id", id),
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "vcs_triggers.all_pull_requests.createWfRun.enabled", "true"),
				),
			},
			{
				Config: testAccWorkflowGit(wfGrpName, id, "CUSTOM", configAllPR(false)),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "id", id),
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "vcs_triggers.all_pull_requests.createWfRun.enabled", "false"),
				),
			},
		},
	})
}

func TestAccWorkflowGit_WithVcsTriggers_CreateTag(t *testing.T) {
	wfGrpName := acctest.ResourceName("tf-provider-workflow-git-vcs-triggers-tag-wfgrp")
	id := acctest.ResourceName("tf-provider-workflow-git-vcs-triggers-tag")

	t.Cleanup(func() {
		deleteWorkflowGitFixture(wfGrpName, id)
		deleteWorkflowGroupFixture(wfGrpName)
	})

	err := createWorkflowGroupFixture(wfGrpName)
	if err != nil {
		t.Errorf("failed to create workflow group fixture: %s", err.Error())
	}

	customHeader := http.Header{}
	customHeader.Set("x-sg-internal-auth-orgid", "sg-provider-test")

	gitCallback := func(enabled bool) string {
		return fmt.Sprintf(`
		  vcs_config = {
		    iac_vcs_config = {
		      custom_source = {
		        config = {
		          is_private = false
		          repo       = "https://github.com/dummy/test-repo.git"
		        }
		        source_config_dest_kind = "GITHUB_COM"
		      }
		    }
		  }
		
		  vcs_triggers = {
		    create_tag = {
		      createWfRun = { enabled = %t }
		    }
		  }
		`, enabled)
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowGit(wfGrpName, id, "CUSTOM", gitCallback(true)),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "id", id),
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "vcs_triggers.create_tag.createWfRun.enabled", "true"),
				),
			},
			{
				Config: testAccWorkflowGit(wfGrpName, id, "CUSTOM", gitCallback(false)),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "id", id),
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "vcs_triggers.create_tag.createWfRun.enabled", "false"),
				),
			},
		},
	})
}

func TestAccWorkflowGit_InNestedWorkflowGroup(t *testing.T) {
	parentWfGrpName := acctest.ResourceName("tf-provider-wfgit-nested-parent")
	childWfGrpName := parentWfGrpName + "/tf-provider-wfgit-nested-child"
	id := acctest.ResourceName("tf-provider-wfgit-nested")

	t.Cleanup(func() {
		deleteWorkflowGitFixture(childWfGrpName, id)
		deleteWorkflowGroupFixture(childWfGrpName)
		deleteWorkflowGroupFixture(parentWfGrpName)
	})

	if err := createWorkflowGroupFixture(parentWfGrpName); err != nil {
		t.Errorf("failed to create parent workflow group fixture: %s", err.Error())
	}

	if err := createWorkflowGroupFixture(childWfGrpName); err != nil {
		t.Errorf("failed to create child workflow group fixture: %s", err.Error())
	}

	customHeader := http.Header{}
	customHeader.Set("x-sg-internal-auth-orgid", "sg-provider-test")

	gitCallback := `
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
		`

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowGit(childWfGrpName, id, "CUSTOM", gitCallback),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "workflow_group_id", childWfGrpName),
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "id", id),
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "vcs_config.iac_vcs_config.custom_source.config.repo", "https://github.com/dummy/test-repo.git"),
				),
			},
		},
	})
}

func TestAccWorkflowGit_WithIacInputData(t *testing.T) {
	wfGrpName := acctest.ResourceName("tf-provider-workflow-git-iac-input-wfgrp")
	id := acctest.ResourceName("tf-provider-workflow-git-iac-input")

	t.Cleanup(func() {
		deleteWorkflowGitFixture(wfGrpName, id)
		deleteWorkflowGroupFixture(wfGrpName)
	})

	err := createWorkflowGroupFixture(wfGrpName)
	if err != nil {
		t.Errorf("failed to create workflow group fixture: %s", err.Error())
	}

	customHeader := http.Header{}
	customHeader.Set("x-sg-internal-auth-orgid", "sg-provider-test")

	gitCallback := func(schemaType, dataExpr string) string {
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
		    iac_input_data = {
		      schema_type = %q
		      data        = %s
		    }
		  }
		
		  terraform_config = {
		    terraform_version = "1.5.0"
		  }
		`, schemaType, dataExpr)
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowGit(wfGrpName, id, "TERRAFORM", gitCallback("RAW_JSON", `jsonencode({"env" = "staging"})`)),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "id", id),
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "vcs_config.iac_input_data.schema_type", "RAW_JSON"),
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "vcs_config.iac_input_data.data", `{"env":"staging"}`),
				),
			},
			{
				Config: testAccWorkflowGit(wfGrpName, id, "TERRAFORM", gitCallback("RAW_JSON", `jsonencode({"env" = "production"})`)),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "id", id),
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "vcs_config.iac_input_data.data", `{"env":"production"}`),
				),
			},
		},
	})
}

func TestAccWorkflowGit_WithRunnerConstraints(t *testing.T) {
	wfGrpName := acctest.ResourceName("tf-provider-workflow-git-runner-wfgrp")
	id := acctest.ResourceName("tf-provider-workflow-git-runner")

	t.Cleanup(func() {
		deleteWorkflowGitFixture(wfGrpName, id)
		deleteWorkflowGroupFixture(wfGrpName)
	})

	err := createWorkflowGroupFixture(wfGrpName)
	if err != nil {
		t.Errorf("failed to create workflow group fixture: %s", err.Error())
	}

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

func TestAccWorkflowGit_WithMiniSteps_WfChaining(t *testing.T) {
	wfGrpName := acctest.ResourceName("tf-provider-workflow-git-chaining-wfgrp")
	id := acctest.ResourceName("tf-provider-workflow-git-chaining")

	t.Cleanup(func() {
		deleteWorkflowGitFixture(wfGrpName, id)
		deleteWorkflowGroupFixture(wfGrpName)
	})

	err := createWorkflowGroupFixture(wfGrpName)
	if err != nil {
		t.Errorf("failed to create workflow group fixture: %s", err.Error())
	}

	customHeader := http.Header{}
	customHeader.Set("x-sg-internal-auth-orgid", "sg-provider-test")

	gitCallback := func(payloadExpr string) string {
		return fmt.Sprintf(`
		  vcs_config = {
		    iac_vcs_config = {
		      custom_source = {
		        config = {
		          is_private = false
		          repo       = "https://github.com/dummy/test-repo.git"
		        }
		        source_config_dest_kind = "GITHUB_COM"
		      }
		    }
		  }
		
		  mini_steps = {
		    wf_chaining = {
		      errored = [{
		        workflow_group_id    = "kk"
		        workflow_id          = "retest-of-bug-cewgh6dt-i7vp-0vo474rl"
		        workflow_run_payload = %s
		      }]
		    }
		  }
		`, payloadExpr)
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowGit(wfGrpName, id, "CUSTOM", gitCallback(`jsonencode({"test" = "value"})`)),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "id", id),
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "mini_steps.wf_chaining.errored.0.workflow_group_id", "kk"),
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "mini_steps.wf_chaining.errored.0.workflow_id", "retest-of-bug-cewgh6dt-i7vp-0vo474rl"),
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "mini_steps.wf_chaining.errored.0.workflow_run_payload", `{"test":"value"}`),
				),
			},
			{
				Config: testAccWorkflowGit(wfGrpName, id, "CUSTOM", gitCallback(`jsonencode({"test" = "updated"})`)),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "id", id),
					resource.TestCheckResourceAttr("stackguardian_workflow_git.test", "mini_steps.wf_chaining.errored.0.workflow_run_payload", `{"test":"updated"}`),
				),
			},
		},
	})
}
