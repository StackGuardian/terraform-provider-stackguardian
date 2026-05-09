# Example 1: Basic workflow with VCS config (minimal required fields)
resource "stackguardian_workflow_git" "basic" {
  workflow_group_id = "my-workflow-group"
  id                = "my-workflow"
  wf_type           = "CUSTOM"

  description = "A basic git-backed workflow"

  vcs_config = {
    iac_vcs_config = {
      custom_source = {
        source_config_dest_kind = "GIT_OTHER"
        config = {
          is_private = false
          repo       = "https://github.com/my-org/my-repo.git"
        }
      }
    }
  }
}

# Example 2: Terraform workflow with terraform_config
resource "stackguardian_workflow_git" "terraform" {
  workflow_group_id = "my-workflow-group"
  id                = "my-terraform-workflow"
  wf_type           = "TERRAFORM"

  vcs_config = {
    iac_vcs_config = {
      custom_source = {
        source_config_dest_kind = "GIT_OTHER"
        config = {
          is_private = false
          repo       = "https://github.com/my-org/my-repo.git"
        }
      }
    }
  }

  terraform_config = {
    terraform_version = "1.5.0"
  }
}

# Example 3: Workflow with environment variables
resource "stackguardian_workflow_git" "with_env_vars" {
  workflow_group_id = "my-workflow-group"
  id                = "my-workflow-with-env"
  wf_type           = "CUSTOM"

  vcs_config = {
    iac_vcs_config = {
      custom_source = {
        source_config_dest_kind = "GIT_OTHER"
        config = {
          is_private = false
          repo       = "https://github.com/my-org/my-repo.git"
        }
      }
    }
  }

  environment_variables = [
    {
      kind = "PLAIN_TEXT"
      config = {
        var_name   = "MY_VAR"
        text_value = "my-value"
      }
    }
  ]
}

# Example 4: Workflow with tags and context_tags
resource "stackguardian_workflow_git" "with_tags" {
  workflow_group_id = "my-workflow-group"
  id                = "my-workflow-with-tags"
  wf_type           = "CUSTOM"

  vcs_config = {
    iac_vcs_config = {
      custom_source = {
        source_config_dest_kind = "GIT_OTHER"
        config = {
          is_private = false
          repo       = "https://github.com/my-org/my-repo.git"
        }
      }
    }
  }

  tags = ["v1"]

  context_tags = {
    env = "production"
  }
}

# Example 5: Workflow with approvers
resource "stackguardian_workflow_git" "with_approvers" {
  workflow_group_id = "my-workflow-group"
  id                = "my-workflow-with-approvers"
  wf_type           = "CUSTOM"

  vcs_config = {
    iac_vcs_config = {
      custom_source = {
        source_config_dest_kind = "GIT_OTHER"
        config = {
          is_private = false
          repo       = "https://github.com/my-org/my-repo.git"
        }
      }
    }
  }

  approvers                    = ["approver@example.com"]
  number_of_approvals_required = 1
}

# Example 6: Workflow with user schedules
resource "stackguardian_workflow_git" "with_schedule" {
  workflow_group_id = "my-workflow-group"
  id                = "my-scheduled-workflow"
  wf_type           = "CUSTOM"

  vcs_config = {
    iac_vcs_config = {
      custom_source = {
        source_config_dest_kind = "GIT_OTHER"
        config = {
          is_private = false
          repo       = "https://github.com/my-org/my-repo.git"
        }
      }
    }
  }

  user_schedules = [
    {
      cron  = "0 8 ? * MON *"
      state = "ENABLED"
      desc  = "Runs on schedule"
    }
  ]
}

# Example 7: Workflow with runner constraints
resource "stackguardian_workflow_git" "shared_runner" {
  workflow_group_id = "my-workflow-group"
  id                = "my-workflow-shared-runner"
  wf_type           = "CUSTOM"

  vcs_config = {
    iac_vcs_config = {
      custom_source = {
        source_config_dest_kind = "GIT_OTHER"
        config = {
          is_private = false
          repo       = "https://github.com/my-org/my-repo.git"
        }
      }
    }
  }

  runner_constraints = {
    type = "shared"
  }
}

resource "stackguardian_workflow_git" "private_runner" {
  workflow_group_id = "my-workflow-group"
  id                = "my-workflow-private-runner"
  wf_type           = "CUSTOM"

  vcs_config = {
    iac_vcs_config = {
      custom_source = {
        source_config_dest_kind = "GIT_OTHER"
        config = {
          is_private = false
          repo       = "https://github.com/my-org/my-repo.git"
        }
      }
    }
  }

  runner_constraints = {
    type  = "private"
    names = ["runner-1"]
  }
}
